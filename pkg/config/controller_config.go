package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"time"
)

const (
	flagCloudProvider                = "cloud-provider"
	flagClusterName                  = "cluster-name"
	flagCloudConfig                  = "cloud-config"
	flagControllers                  = "controllers"
	flagConfigureCloudRoutes         = "configure-cloud-routes"
	flagClusterCidr                  = "cluster-cidr"
	flagFeatureGates                 = "feature-gates"
	flagUseServiceAccountCredentials = "use-service-account-credentials"

	flagDryRun                         = "dry-run"
	flagServiceMaxConcurrentReconciles = "concurrent-service-syncs"
	flagRouteReconciliationPeriod      = "route-reconciliation-period"
	flagNodeMonitorPeriod              = "node-monitor-period"
	flagNetwork                        = "network"

	defaultCloudProvider             = "alibabacloud"
	defaultClusterName               = "kubernetes"
	defaultConfigureCloudRoutes      = true
	defaultMaxConcurrentReconciles   = 3
	defaultCloudConfig               = ""
	defaultRouteReconciliationPeriod = 5 * time.Minute
	defaultNodeMonitorPeriod         = 5 * time.Minute
	defaultNetwork                   = "vpc"
)

var ControllerCFG = &ControllerConfig{}

// Flag stores the configuration for global usage
type ControllerConfig struct {
	CloudProvider                string
	ClusterName                  string
	CloudConfig                  string
	Controllers                  []string
	UseServiceAccountCredentials bool
	FeatureGates                 string
	//For Flannel Network
	ConfigureCloudRoutes           bool
	ClusterCidr                    string
	AllocateNodeCIDRs              bool
	ServiceMaxConcurrentReconciles int
	RouteReconciliationPeriod      time.Duration
	NodeMonitorPeriod              time.Duration
	LogLevel                       int
	DryRun                         bool
	NetWork                        string

	RuntimeConfig RuntimeConfig
}

func (cfg *ControllerConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cfg.CloudProvider, flagCloudProvider, defaultCloudProvider, "The provider for cloud services. Empty string for no provider.")
	fs.StringVar(&cfg.ClusterName, flagClusterName, defaultClusterName, "The instance prefix for the cluster.")
	fs.StringVar(&cfg.CloudConfig, flagCloudConfig, defaultCloudConfig,
		"The path to the cloud provider configuration file. Empty string for no configuration file.")
	fs.StringSliceVar(&cfg.Controllers, flagControllers, []string{"node", "route", "service"}, "A list of controllers to enable.")
	fs.BoolVar(&cfg.UseServiceAccountCredentials, flagUseServiceAccountCredentials, false, "If true, use individual service account credentials for each controller.")
	fs.BoolVar(&cfg.ConfigureCloudRoutes, flagConfigureCloudRoutes, defaultConfigureCloudRoutes, "Should CIDRs allocated by allocate-node-cidrs be configured on the cloud provider.")
	fs.StringVar(&cfg.ClusterCidr, flagClusterCidr, "", "CIDR Range for Pods in cluster. Requires --allocate-node-cidrs to be true.")
	fs.IntVar(&cfg.ServiceMaxConcurrentReconciles, flagServiceMaxConcurrentReconciles, defaultMaxConcurrentReconciles,
		"Maximum number of concurrently running reconcile loops for service")
	fs.BoolVar(&cfg.DryRun, flagDryRun, false, "whether to perform a dry run")
	fs.StringVar(&cfg.NetWork, flagNetwork, defaultNetwork, "Set network type for controller.")
	fs.DurationVar(&cfg.RouteReconciliationPeriod, flagRouteReconciliationPeriod, defaultRouteReconciliationPeriod,
		"The period for reconciling routes created for nodes by cloud provider. The minimum value is 1 minute")
	fs.DurationVar(&cfg.NodeMonitorPeriod, flagNodeMonitorPeriod, defaultNodeMonitorPeriod, "The period for syncing NodeStatus in NodeController.")
	fs.StringVar(&cfg.FeatureGates, flagFeatureGates, "", "A set of key=value pairs that describe feature gates for alpha/experimental features.")

	cfg.RuntimeConfig.BindFlags(fs)
}

// Validate the controller configuration
func (cfg *ControllerConfig) Validate() error {
	if cfg.CloudConfig == "" {
		return fmt.Errorf("cloud config cannot be empty")
	}

	if cfg.ConfigureCloudRoutes == true && cfg.ClusterCidr == "" {
		return fmt.Errorf("--cluster-cidr must be set when --configure-cloud-routes=true")
	}

	if cfg.RouteReconciliationPeriod < 1*time.Minute {
		cfg.RouteReconciliationPeriod = 1 * time.Minute
	}
	return nil
}
