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
	"k8s.io/kubernetes/pkg/util/flag"
	"k8s.io/kubernetes/pkg/client/leaderelectionconfig"
	// add the kubernetes feature gates
	_ "k8s.io/kubernetes/pkg/features"

	"github.com/spf13/pflag"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/cloudprovider/app"
)

// AddFlags adds flags for a specific ExternalCMServer to the specified FlagSet
func AddFlags(ccm *app.ServerCCM, fs *pflag.FlagSet) {
	fs.Int32Var(&ccm.Generic.Port, "port", ccm.Generic.Port, "The port that the cloud-controller-manager'ccm http service runs on.")
	fs.Var(flag.IPVar{Val: &ccm.Generic.Address}, "address", "The IP address to serve on (set to 0.0.0.0 for all interfaces).")
	fs.StringVar(&ccm.KubeCloudShared.CloudProvider.Name, "cloud-provider", ccm.KubeCloudShared.CloudProvider.Name, "The provider of cloud services. Cannot be empty.")
	fs.StringVar(&ccm.KubeCloudShared.CloudProvider.CloudConfigFile, "cloud-config", ccm.KubeCloudShared.CloudProvider.CloudConfigFile, "The path to the cloud provider configuration file. Empty string for no configuration file.")
	fs.BoolVar(&ccm.KubeCloudShared.AllowUntaggedCloud, "allow-untagged-cloud", false, "Allow the cluster to run without the cluster-id on cloud instances. This is a legacy mode of operation and a cluster-id will be required in the future.")
	fs.MarkDeprecated("allow-untagged-cloud", "This flag is deprecated and will be removed in a future release. A cluster-id will be required on cloud instances.")
	fs.DurationVar(&ccm.Generic.MinResyncPeriod.Duration, "min-resync-period", ccm.Generic.MinResyncPeriod.Duration, "The resync period in reflectors will be random between MinResyncPeriod and 2*MinResyncPeriod.")
	fs.DurationVar(&ccm.KubeCloudShared.NodeMonitorPeriod.Duration, "node-monitor-period", ccm.KubeCloudShared.NodeMonitorPeriod.Duration,
		"The period for syncing NodeStatus in NodeController.")
	fs.DurationVar(&ccm.NodeStatusUpdateFrequency.Duration, "node-status-update-frequency", ccm.NodeStatusUpdateFrequency.Duration, "Specifies how often the controller updates nodes' status.")
	// TODO: remove --service-account-private-key-file 6 months after 1.8 is released (~1.10)
	//fs.StringVar(&ccm.ServiceAccountKeyFile, "service-account-private-key-file", ccm.ServiceAccountKeyFile, "Filename containing a PEM-encoded private RSA or ECDSA key used to sign service account tokens.")
	fs.MarkDeprecated("service-account-private-key-file", "This flag is currently no-op and will be deleted.")
	fs.BoolVar(&ccm.KubeCloudShared.UseServiceAccountCredentials, "use-service-account-credentials", ccm.KubeCloudShared.UseServiceAccountCredentials, "If true, use individual service account credentials for each controller.")
	fs.DurationVar(&ccm.KubeCloudShared.RouteReconciliationPeriod.Duration, "route-reconciliation-period", ccm.KubeCloudShared.RouteReconciliationPeriod.Duration, "The period for reconciling routes created for nodes by cloud provider.")
	fs.BoolVar(&ccm.KubeCloudShared.ConfigureCloudRoutes, "configure-cloud-routes", true, "Should CIDRs allocated by allocate-node-cidrs be configured on the cloud provider.")
	fs.BoolVar(&ccm.Generic.Debugging.EnableProfiling, "profiling", true, "Enable profiling via web interface host:port/debug/pprof/.")
	fs.BoolVar(&ccm.Generic.Debugging.EnableContentionProfiling, "contention-profiling", false, "Enable lock contention profiling, if profiling is enabled.")
	fs.StringVar(&ccm.KubeCloudShared.ClusterCIDR, "cluster-cidr", ccm.KubeCloudShared.ClusterCIDR, "CIDR Range for Pods in cluster.")
	fs.StringVar(&ccm.KubeCloudShared.ClusterName, "cluster-name", ccm.KubeCloudShared.ClusterName, "The instance prefix for the cluster.")
	fs.BoolVar(&ccm.KubeCloudShared.AllocateNodeCIDRs, "allocate-node-cidrs", false, "Should CIDRs for Pods be allocated and set on the cloud provider.")
	fs.StringVar(&ccm.Master, "master", ccm.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&ccm.Kubeconfig, "kubeconfig", ccm.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.StringVar(&ccm.Generic.ClientConnection.ContentType, "kube-api-content-type", ccm.Generic.ClientConnection.ContentType, "Content type of requests sent to apiserver.")
	fs.Float32Var(&ccm.Generic.ClientConnection.QPS, "kube-api-qps", ccm.Generic.ClientConnection.QPS, "QPS to use while talking with kubernetes apiserver.")
	fs.Int32Var(&ccm.Generic.ClientConnection.Burst, "kube-api-burst", ccm.Generic.ClientConnection.Burst, "Burst to use while talking with kubernetes apiserver.")
	fs.DurationVar(&ccm.Generic.ControllerStartInterval.Duration, "controller-start-interval", ccm.Generic.ControllerStartInterval.Duration, "Interval between starting controller managers.")
	fs.Int32Var(&ccm.ServiceController.ConcurrentServiceSyncs, "concurrent-service-syncs", ccm.ServiceController.ConcurrentServiceSyncs, "The number of services that are allowed to sync concurrently. Larger number = more responsive service management, but more CPU (and network) load")
	leaderelectionconfig.BindFlags(&ccm.Generic.LeaderElection, fs)

	utilfeature.DefaultMutableFeatureGate.AddFlag(fs)
}
