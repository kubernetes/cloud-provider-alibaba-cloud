/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/apis/componentconfig"
	"k8s.io/kubernetes/pkg/client/leaderelectionconfig"
	// add the kubernetes feature gates
	_ "k8s.io/kubernetes/pkg/features"

	"github.com/spf13/pflag"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/cloudprovider/app"
)



// AddFlags adds flags for a specific ExternalCMServer to the specified FlagSet
func AddFlags(ccm *app.ServerCCM, fs *pflag.FlagSet) {
	fs.Int32Var(&ccm.Port, "port", ccm.Port, "The port that the cloud-controller-manager'ccm http service runs on.")
	fs.Var(componentconfig.IPVar{Val: &ccm.Address}, "address", "The IP address to serve on (set to 0.0.0.0 for all interfaces).")
	fs.StringVar(&ccm.CloudProvider, "cloud-provider", ccm.CloudProvider, "The provider of cloud services. Cannot be empty.")
	fs.StringVar(&ccm.CloudConfigFile, "cloud-config", ccm.CloudConfigFile, "The path to the cloud provider configuration file. Empty string for no configuration file.")
	fs.BoolVar(&ccm.AllowUntaggedCloud, "allow-untagged-cloud", false, "Allow the cluster to run without the cluster-id on cloud instances. This is a legacy mode of operation and a cluster-id will be required in the future.")
	fs.MarkDeprecated("allow-untagged-cloud", "This flag is deprecated and will be removed in a future release. A cluster-id will be required on cloud instances.")
	fs.DurationVar(&ccm.MinResyncPeriod.Duration, "min-resync-period", ccm.MinResyncPeriod.Duration, "The resync period in reflectors will be random between MinResyncPeriod and 2*MinResyncPeriod.")
	fs.DurationVar(&ccm.NodeMonitorPeriod.Duration, "node-monitor-period", ccm.NodeMonitorPeriod.Duration,
		"The period for syncing NodeStatus in NodeController.")
	fs.DurationVar(&ccm.NodeStatusUpdateFrequency.Duration, "node-status-update-frequency", ccm.NodeStatusUpdateFrequency.Duration, "Specifies how often the controller updates nodes' status.")
	// TODO: remove --service-account-private-key-file 6 months after 1.8 is released (~1.10)
	fs.StringVar(&ccm.ServiceAccountKeyFile, "service-account-private-key-file", ccm.ServiceAccountKeyFile, "Filename containing a PEM-encoded private RSA or ECDSA key used to sign service account tokens.")
	fs.MarkDeprecated("service-account-private-key-file", "This flag is currently no-op and will be deleted.")
	fs.BoolVar(&ccm.UseServiceAccountCredentials, "use-service-account-credentials", ccm.UseServiceAccountCredentials, "If true, use individual service account credentials for each controller.")
	fs.DurationVar(&ccm.RouteReconciliationPeriod.Duration, "route-reconciliation-period", ccm.RouteReconciliationPeriod.Duration, "The period for reconciling routes created for Nodes by cloud provider.")
	fs.BoolVar(&ccm.ConfigureCloudRoutes, "configure-cloud-routes", true, "Should CIDRs allocated by allocate-node-cidrs be configured on the cloud provider.")
	fs.BoolVar(&ccm.EnableProfiling, "profiling", true, "Enable profiling via web interface host:port/debug/pprof/.")
	fs.BoolVar(&ccm.EnableContentionProfiling, "contention-profiling", false, "Enable lock contention profiling, if profiling is enabled.")
	fs.StringVar(&ccm.ClusterCIDR, "cluster-cidr", ccm.ClusterCIDR, "CIDR Range for Pods in cluster.")
	fs.StringVar(&ccm.ClusterName, "cluster-name", ccm.ClusterName, "The instance prefix for the cluster.")
	fs.BoolVar(&ccm.AllocateNodeCIDRs, "allocate-node-cidrs", false, "Should CIDRs for Pods be allocated and set on the cloud provider.")
	fs.StringVar(&ccm.Master, "master", ccm.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&ccm.Kubeconfig, "kubeconfig", ccm.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.StringVar(&ccm.ContentType, "kube-api-content-type", ccm.ContentType, "Content type of requests sent to apiserver.")
	fs.Float32Var(&ccm.KubeAPIQPS, "kube-api-qps", ccm.KubeAPIQPS, "QPS to use while talking with kubernetes apiserver.")
	fs.Int32Var(&ccm.KubeAPIBurst, "kube-api-burst", ccm.KubeAPIBurst, "Burst to use while talking with kubernetes apiserver.")
	fs.DurationVar(&ccm.ControllerStartInterval.Duration, "controller-start-interval", ccm.ControllerStartInterval.Duration, "Interval between starting controller managers.")
	fs.Int32Var(&ccm.ConcurrentServiceSyncs, "concurrent-service-syncs", ccm.ConcurrentServiceSyncs, "The number of services that are allowed to sync concurrently. Larger number = more responsive service management, but more CPU (and network) load")
	leaderelectionconfig.BindFlags(&ccm.LeaderElection, fs)

	utilfeature.DefaultFeatureGate.AddFlag(fs)
}
