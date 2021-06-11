package context

import (
	"github.com/spf13/pflag"
)

const (
	flagLogLevel                       = "log-level"
	flagDryRun                         = "dry-run"
	flagCloudConfig                    = "cloud-config"
	flagServiceMaxConcurrentReconciles = "concurrent-service-syncs"
	flagEnableControllers              = "enable-controllers"
	defaultLogLevel                    = "info"
	defaultMaxConcurrentReconciles     = 3
	defaultCloudConfig                 = ""
)

var ControllerCFG = &ControllerConfig{}

// Flag stores the configuration for global usage
type ControllerConfig struct {
	LogLevel                       string
	ServiceMaxConcurrentReconciles int
	EnableControllers              []string
	DryRun                         bool
	CloudConfig                    string

	RuntimeConfig RuntimeConfig
}

func (cfg *ControllerConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cfg.LogLevel, flagLogLevel, defaultLogLevel,
		"Set the controller log level - info(default), debug")
	fs.IntVar(&cfg.ServiceMaxConcurrentReconciles, flagServiceMaxConcurrentReconciles, defaultMaxConcurrentReconciles,
		"Maximum number of concurrently running reconcile loops for service")
	fs.StringSliceVar(&cfg.EnableControllers, flagEnableControllers, []string{"node", "route", "service"},
		"Enable controllers, default enable node controller, route controller and service controller")
	fs.BoolVar(&cfg.DryRun, flagDryRun, false, "whether to perform a dry run")
	fs.StringVar(&cfg.CloudConfig, flagCloudConfig, defaultCloudConfig,
		"Path to the cloud provider configuration file. Empty string for no configuration file.")

	cfg.RuntimeConfig.BindFlags(fs)
}

// Validate the controller configuration
func (cfg *ControllerConfig) Validate() error {

	return nil
}
