package context

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

const (
	flagMetricsBindAddr            = "metrics-bind-addr"
	flagHealthProbeBindAddr        = "health-probe-bind-addr"
	flagEnableLeaderElection       = "enable-leader-election"
	flagLeaderElectionID           = "leader-election-id"
	flagLeaderElectionNamespace    = "leader-election-namespace"
	flagLeaderElectionResourceLock = "leader-election-resource-lock"
	flagSyncPeriod                 = "sync-period"
	flagQPS                        = "kube-api-qps"
	flagBurst                      = "kube-api-burst"

	defaultMetricsAddr                = ":8080"
	defaultHealthProbeBindAddress     = ":10258"
	defaultLeaderElectionID           = "ccm"
	defaultLeaderElectionNamespace    = "kube-system"
	defaultLeaderElectionResourceLock = "endpointsleases"
	defaultSyncPeriod                 = 60 * time.Minute
	defaultQPS                        = 20.0
	defaultBurst                      = 30
)

// RuntimeConfig stores the configuration for controller-runtime
type RuntimeConfig struct {
	MetricsBindAddress         string
	HealthProbeBindAddress     string
	EnableLeaderElection       bool
	LeaderElectionID           string
	LeaderElectionNamespace    string
	LeaderElectionResourceLock string
	SyncPeriod                 time.Duration
	QPS                        float32
	Burst                      int
}

func (c *RuntimeConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.MetricsBindAddress, flagMetricsBindAddr, defaultMetricsAddr,
		"The address the metric endpoint binds to.")
	fs.StringVar(&c.HealthProbeBindAddress, flagHealthProbeBindAddr, defaultHealthProbeBindAddress,
		"The address the health probes binds to.")
	fs.BoolVar(&c.EnableLeaderElection, flagEnableLeaderElection, true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&c.LeaderElectionID, flagLeaderElectionID, defaultLeaderElectionID,
		"Name of the leader election ID to use for this controller")
	fs.StringVar(&c.LeaderElectionNamespace, flagLeaderElectionNamespace, defaultLeaderElectionNamespace,
		"Name of the leader election ID to use for this controller")
	fs.StringVar(&c.LeaderElectionResourceLock, flagLeaderElectionResourceLock, defaultLeaderElectionResourceLock,
		"Resource lock to use for leader election")
	fs.DurationVar(&c.SyncPeriod, flagSyncPeriod, defaultSyncPeriod,
		"Period at which the controller forces the repopulation of its local object stores.")
	fs.Float32Var(&c.QPS, flagQPS, defaultQPS, "QPS to use while talking with kubernetes apiserver.")
	fs.IntVar(&c.Burst, flagBurst, defaultBurst, "Burst to use while talking with kubernetes apiserver.")
}

func BuildRuntimeOptions(rtCfg RuntimeConfig) manager.Options {
	return manager.Options{
		MetricsBindAddress:         rtCfg.MetricsBindAddress,
		HealthProbeBindAddress:     rtCfg.HealthProbeBindAddress,
		LeaderElection:             rtCfg.EnableLeaderElection,
		LeaderElectionID:           rtCfg.LeaderElectionID,
		LeaderElectionResourceLock: rtCfg.LeaderElectionResourceLock,
		LeaderElectionNamespace:    rtCfg.LeaderElectionNamespace,
		SyncPeriod:                 &rtCfg.SyncPeriod,
	}
}
