package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/cloud-provider/config"
	"k8s.io/klog/v2"
	sigConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	flagCloudProvider                   = "cloud-provider"
	flagClusterName                     = "cluster-name"
	flagCloudConfig                     = "cloud-config"
	flagControllers                     = "controllers"
	flagConfigureCloudRoutes            = "configure-cloud-routes"
	flagClusterCidr                     = "cluster-cidr"
	flagAllocateNodeCIDRs               = "allocate-node-cidrs"
	flagFeatureGates                    = "feature-gates"
	flagUseServiceAccountCredentials    = "use-service-account-credentials"
	flagSkipDisableSourceDestCheck      = "skip-disable-source-dest-check"
	flagNodeEventAggregationWaitSeconds = "node-event-aggregation-wait-seconds"

	flagDryRun                         = "dry-run"
	flagServiceMaxConcurrentReconciles = "concurrent-service-syncs"
	flagRouteReconciliationPeriod      = "route-reconciliation-period"
	flagNodeMonitorPeriod              = "node-monitor-period"
	flagServerGroupBatchSize           = "sg-batch-size"
	flagNetwork                        = "network"

	defaultCloudProvider                  = "alibabacloud"
	defaultClusterName                    = "kubernetes"
	defaultConfigureCloudRoutes           = true
	defaultServiceMaxConcurrentReconciles = 3
	defaultCloudConfig                    = ""
	defaultRouteReconciliationPeriod      = 5 * time.Minute
	defaultNodeMonitorPeriod              = 5 * time.Minute
	defaultServerGroupBatchSize           = 40
	defaultNetwork                        = "vpc"

	defaultMaxConcurrentActions = 10
)

var ControllerCFG = &ControllerConfig{
	CloudConfig: CloudCFG,
}

// Flag stores the configuration for global usage
type ControllerConfig struct {
	config.KubeCloudSharedConfiguration
	CloudConfigPath                 string
	Controllers                     []string
	FeatureGates                    string
	ServerGroupBatchSize            int
	MaxConcurrentActions            int
	LogLevel                        int
	DryRun                          bool
	NetWork                         string
	NodeReconcileBatchSize          int
	RouteReconcileBatchSize         int
	SkipDisableSourceDestCheck      bool
	NodeEventAggregationWaitSeconds int

	RuntimeConfig RuntimeConfig
	CloudConfig   *CloudConfig
}

func (cfg *ControllerConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cfg.CloudProvider.Name, flagCloudProvider, defaultCloudProvider, "The provider for cloud services. Empty string for no provider.")
	fs.StringVar(&cfg.ClusterName, flagClusterName, defaultClusterName, "The instance prefix for the cluster.")
	fs.StringVar(&cfg.CloudConfigPath, flagCloudConfig, defaultCloudConfig,
		"The path to the cloud provider configuration file. Empty string for no configuration file.")
	fs.StringSliceVar(&cfg.Controllers, flagControllers, []string{"node", "route", "service", "nlb"}, "A list of controllers to enable.")
	fs.BoolVar(&cfg.UseServiceAccountCredentials, flagUseServiceAccountCredentials, false, "If true, use individual service account credentials for each controller.")
	fs.BoolVar(&cfg.ConfigureCloudRoutes, flagConfigureCloudRoutes, defaultConfigureCloudRoutes, "Should CIDRs allocated by allocate-node-cidrs be configured on the cloud provider.")
	fs.StringVar(&cfg.ClusterCIDR, flagClusterCidr, "", "CIDR Range for Pods in cluster. Requires --allocate-node-cidrs to be true.")
	fs.BoolVar(&cfg.AllocateNodeCIDRs, flagAllocateNodeCIDRs, false, "Should CIDRs for Pods be allocated and set on the cloud provider.")
	fs.IntVar(&cfg.CloudConfig.Global.ServiceMaxConcurrentReconciles, flagServiceMaxConcurrentReconciles, defaultServiceMaxConcurrentReconciles,
		"[Deprecated, please use cloud-config config file instead] Maximum number of concurrently running reconcile loops for service")
	fs.BoolVar(&cfg.DryRun, flagDryRun, false, "whether to perform a dry run")
	fs.StringVar(&cfg.NetWork, flagNetwork, defaultNetwork, "Set network type for controller.")
	fs.DurationVar(&cfg.RouteReconciliationPeriod.Duration, flagRouteReconciliationPeriod, defaultRouteReconciliationPeriod,
		"The period for reconciling routes created for nodes by cloud provider. The minimum value is 1 minute")
	fs.DurationVar(&cfg.NodeMonitorPeriod.Duration, flagNodeMonitorPeriod, defaultNodeMonitorPeriod, "The period for syncing NodeStatus in NodeController.")
	fs.IntVar(&cfg.ServerGroupBatchSize, flagServerGroupBatchSize, defaultServerGroupBatchSize, "The batch size for syncing server group. The value range is 1-40")
	fs.BoolVar(&cfg.AllowUntaggedCloud, "allow-untagged-cloud", false, "Allow the cluster to run without the cluster-id on cloud instances. This is a legacy mode of operation and a cluster-id will be required in the future.")
	fs.IntVar(&cfg.MaxConcurrentActions, "max-concurrent-actions", defaultMaxConcurrentActions, "The max concurrent number of actions for listener and server group updates")
	_ = fs.MarkDeprecated("allow-untagged-cloud", "This flag is deprecated and will be removed in a future release. A cluster-id will be required on cloud instances.")

	fs.IntVar(&cfg.NodeReconcileBatchSize, "node-reconcile-batch-size", 100, "The batch size for syncing node status. The value range is 1-100")
	fs.IntVar(&cfg.RouteReconcileBatchSize, "route-reconcile-batch-size", 50, "The batch size for syncing route status. The value range is 1-50")
	fs.BoolVar(&cfg.SkipDisableSourceDestCheck, flagSkipDisableSourceDestCheck, false, "Skip disable source dest check for nodes")
	fs.IntVar(&cfg.NodeEventAggregationWaitSeconds, flagNodeEventAggregationWaitSeconds, 1, "The wait second for aggregating node events in node & route controller")

	cfg.RuntimeConfig.BindFlags(fs)
}

// Validate the controller configuration
func (cfg *ControllerConfig) Validate() error {
	if cfg.CloudConfigPath == "" {
		return fmt.Errorf("cloud config cannot be empty")
	}

	if cfg.ConfigureCloudRoutes && cfg.ClusterCIDR == "" {
		return fmt.Errorf("--cluster-cidr must be set when --configure-cloud-routes=true")
	}

	if cfg.RouteReconciliationPeriod.Duration < 1*time.Minute {
		cfg.RouteReconciliationPeriod.Duration = 1 * time.Minute
	}

	if cfg.NodeReconcileBatchSize == 0 {
		cfg.NodeReconcileBatchSize = 100
	}

	if cfg.NodeEventAggregationWaitSeconds < 0 {
		cfg.NodeEventAggregationWaitSeconds = 0
	}

	if cfg.MaxConcurrentActions <= 0 {
		return fmt.Errorf("--max-concurrent-actions must be set to a positive integer")
	}
	return nil
}

func (cfg *ControllerConfig) LoadControllerConfig() error {
	klog.InitFlags(nil)

	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	fs.AddGoFlagSet(flag.CommandLine)
	cfg.BindFlags(fs)

	if err := fs.Parse(os.Args); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	if err := cfg.CloudConfig.LoadCloudCFG(); err != nil {
		return fmt.Errorf("load cloud config error: %s", err.Error())
	}
	cfg.CloudConfig.PrintInfo()
	klog.Infof("NodeReconcileBatchSize: %d, RouteReconcileBatchSize: %d", cfg.NodeReconcileBatchSize, cfg.RouteReconcileBatchSize)

	if cfg.CloudConfig.Global.FeatureGates != "" {
		apiClient := apiext.NewForConfigOrDie(sigConfig.GetConfigOrDie())
		return BindFeatureGates(apiClient, cfg.CloudConfig.Global.FeatureGates)
	}
	return nil
}
