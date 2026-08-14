#!/usr/bin/env bash

set -Eeuo pipefail

usage() {
	cat <<'EOF'
Run ACK CCM E2E tests in three phases:
  1. CLB specs with 2 workers (default)
  2. NLB specs with 6 workers (default), in parallel with CLB
  3. cluster-mutating specs with 1 worker, after CLB and NLB succeed

Required environment variables:
  KUBECONFIG           kubeconfig for the dedicated E2E cluster
  E2E_CLUSTER_ID       ACK cluster ID
  E2E_REGION_ID        ACK region ID, for example cn-hangzhou

Optional environment variables:
  E2E_CLOUD_CONFIG     CCM cloud config used by the E2E client. When omitted,
                       a minimal config is generated and cluster identity is
                       discovered from ACK using E2E_CLUSTER_ID.
  E2E_CLB_PROCS        CLB worker count (default: 2)
  E2E_NLB_PROCS        NLB worker count (default: 6)
  E2E_CONTROLLERS      registered controllers (default: service)
  E2E_OUTPUT_DIR       reports/log directory (default: _output/e2e/<timestamp>)
  E2E_PHASE_TIMEOUT    timeout for each phase (default: 6h)
  E2E_FIXTURE_TIMEOUT  backend fixture readiness timeout (default: 10m)
  E2E_GINKGO_BIN       optional path to a Ginkgo v2 binary

Additional arguments are passed to the E2E test binary. Example:

  KUBECONFIG="$HOME/.kube/e2e-ccm-config" \
  E2E_CLUSTER_ID="<cluster-id>" \
  E2E_REGION_ID="cn-hangzhou" \
  ./hack/run-e2e.sh --ipv6=true

The script does not start or deploy CCM and does not load .asi-env. Ensure the
CCM under test is already running against the same cluster.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
	usage
	exit 0
fi

require_env() {
	local name=$1
	if [[ -z "${!name:-}" ]]; then
		printf 'required environment variable %s is not set\n\n' "$name" >&2
		usage >&2
		exit 2
	fi
}

require_positive_integer() {
	local name=$1
	local value=$2
	if [[ ! "$value" =~ ^[1-9][0-9]*$ ]]; then
		printf '%s must be a positive integer, got %q\n' "$name" "$value" >&2
		exit 2
	fi
}

require_env KUBECONFIG
require_env E2E_CLUSTER_ID
require_env E2E_REGION_ID

if [[ ! -f "$KUBECONFIG" ]]; then
	printf 'KUBECONFIG does not exist: %s\n' "$KUBECONFIG" >&2
	exit 2
fi
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
clb_procs=${E2E_CLB_PROCS:-2}
nlb_procs=${E2E_NLB_PROCS:-6}
controllers=${E2E_CONTROLLERS:-service}
phase_timeout=${E2E_PHASE_TIMEOUT:-6h}
fixture_timeout=${E2E_FIXTURE_TIMEOUT:-10m}
output_dir=${E2E_OUTPUT_DIR:-"$repo_root/_output/e2e/$(date +%Y%m%d-%H%M%S)"}

require_positive_integer E2E_CLB_PROCS "$clb_procs"
require_positive_integer E2E_NLB_PROCS "$nlb_procs"

mkdir -p "$output_dir"
if [[ -n "${E2E_CLOUD_CONFIG:-}" ]]; then
	if [[ ! -f "$E2E_CLOUD_CONFIG" ]]; then
		printf 'E2E_CLOUD_CONFIG does not exist: %s\n' "$E2E_CLOUD_CONFIG" >&2
		exit 2
	fi
	cloud_config_dir=$(cd "$(dirname "$E2E_CLOUD_CONFIG")" && pwd)
	cloud_config="$cloud_config_dir/$(basename "$E2E_CLOUD_CONFIG")"
else
	cloud_config="$output_dir/cloud-config.json"
	printf '{"Global": {}}\n' >"$cloud_config"
fi

if [[ -n "${E2E_GINKGO_BIN:-}" ]]; then
	if [[ ! -x "$E2E_GINKGO_BIN" ]]; then
		printf 'E2E_GINKGO_BIN is not executable: %s\n' "$E2E_GINKGO_BIN" >&2
		exit 2
	fi
	ginkgo_cmd=("$E2E_GINKGO_BIN")
elif command -v ginkgo >/dev/null 2>&1 && ginkgo version >/dev/null 2>&1; then
	ginkgo_cmd=(ginkgo)
else
	# Use the version pinned by go.mod. This also works when an asdf ginkgo shim
	# exists but no matching shim version is currently selected.
	ginkgo_cmd=(go run -mod=mod github.com/onsi/ginkgo/v2/ginkgo)
fi

export KUBECONFIG

common_test_args=(
	"--cloud-config=$cloud_config"
	"--region-id=$E2E_REGION_ID"
	"--cluster-id=$E2E_CLUSTER_ID"
	"--controllers=$controllers"
	"--allow-create-cloud-resources=true"
	"--fixture-ready-timeout=$fixture_timeout"
)
extra_test_args=("$@")

run_phase() {
	local name=$1
	local procs=$2
	local label_filter=$3
	local resource_types=$4
	local phase_dir="$output_dir/$name"

	mkdir -p "$phase_dir"
	printf '[%s] start: procs=%s labels=%s resources=%s\n' \
		"$name" "$procs" "$label_filter" "$resource_types"

	(
		cd "$repo_root"
		"${ginkgo_cmd[@]}" \
			--procs="$procs" \
			--label-filter="$label_filter" \
			--timeout="$phase_timeout" \
			--fail-on-empty \
			--no-color \
			--json-report=report.json \
			--output-dir="$phase_dir" \
			./test/e2e -- \
			"${common_test_args[@]}" \
			"--cloud-resource-types=$resource_types" \
			"${extra_test_args[@]}"
	) 2>&1 | tee "$phase_dir/run.log"
}

printf 'E2E cluster: %s (%s)\n' "$E2E_CLUSTER_ID" "$E2E_REGION_ID"
printf 'Kubeconfig: %s\n' "$KUBECONFIG"
printf 'Cloud config: %s\n' "$cloud_config"
printf 'Output: %s\n' "$output_dir"

run_phase clb "$clb_procs" 'clb && !cluster-serial' clb &
clb_pid=$!
run_phase nlb "$nlb_procs" 'nlb && !cluster-serial' nlb &
nlb_pid=$!

set +e
wait "$clb_pid"
clb_status=$?
wait "$nlb_pid"
nlb_status=$?
set -e

if ((clb_status != 0 || nlb_status != 0)); then
	printf 'parallel phases failed: clb=%d nlb=%d; serial phase was not started\n' \
		"$clb_status" "$nlb_status" >&2
	exit 1
fi

run_phase cluster-serial 1 'cluster-serial' 'clb,nlb'

printf 'all E2E phases passed; reports are in %s\n' "$output_dir"
