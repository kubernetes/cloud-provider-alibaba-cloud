package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"time"
)

const (
	flagDryRun                         = "dry-run"
	flagCloudConfig                    = "cloud-config"
	flagServiceMaxConcurrentReconciles = "concurrent-service-syncs"
	flagEnableControllers              = "enable-controllers"
	flagConfigureCloudRoutes           = "configure-cloud-routes"
	flagClusterCidr                    = "cluster-cidr"
	flagRouteReconciliationPeriod      = "route-reconciliation-period"
	flagNetwork                        = "network"

	defaultMaxConcurrentReconciles   = 3
	defaultCloudConfig               = ""
	defaultRouteReconciliationPeriod = 5 * time.Minute
	defaultNetwork                   = "vpc"
)

var ControllerCFG = &ControllerConfig{}

// Flag stores the configuration for global usage
type ControllerConfig struct {
	LogLevel                       int
	ServiceMaxConcurrentReconciles int
	EnableControllers              []string
	DryRun                         bool
	CloudConfig                    string
	NetWork                        string
	//For Flannel Network
	ConfigureCloudRoutes      bool
	ClusterCidr               string
	RouteReconciliationPeriod time.Duration

	RuntimeConfig RuntimeConfig
}

func (cfg *ControllerConfig) BindFlags(fs *pflag.FlagSet) {
	fs.IntVar(&cfg.ServiceMaxConcurrentReconciles, flagServiceMaxConcurrentReconciles, defaultMaxConcurrentReconciles,
		"Maximum number of concurrently running reconcile loops for service")
	fs.StringSliceVar(&cfg.EnableControllers, flagEnableControllers, []string{"node", "route", "service"},
		"Enable controllers, default enable node controller, route controller and service controller")
	fs.BoolVar(&cfg.DryRun, flagDryRun, false, "whether to perform a dry run")
	fs.StringVar(&cfg.CloudConfig, flagCloudConfig, defaultCloudConfig,
		"Path to the cloud provider configuration file. Empty string for no configuration file.")
	fs.StringVar(&cfg.NetWork, flagNetwork, defaultNetwork, "Set network type for controller.")
	fs.BoolVar(&cfg.ConfigureCloudRoutes, flagConfigureCloudRoutes, false, "Enable configure cloud routes.")
	fs.StringVar(&cfg.ClusterCidr, flagClusterCidr, "", "CIDR Range for Pods in cluster.") // todo: support ipv6 dual stack
	fs.DurationVar(&cfg.RouteReconciliationPeriod,
		flagRouteReconciliationPeriod, defaultRouteReconciliationPeriod,
		"The period for reconciling routes created for nodes by cloud provider. The minimum value is 1 minute")

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
