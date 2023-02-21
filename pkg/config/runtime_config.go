package config

import (
	"time"

	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	flagMetricsBindAddr              = "metrics-bind-addr"
	flagHealthProbeBindAddr          = "health-probe-bind-addr"
	flagQPS                          = "kube-api-qps"
	flagBurst                        = "kube-api-burst"
	flagLeaderElect                  = "leader-elect"
	flagLeaderElectLeaseDuration     = "leader-elect-lease-duration"
	flagLeaderElectRenewDeadline     = "leader-elect-renew-deadline"
	flagLeaderElectResourceLock      = "leader-elect-resource-lock"
	flagLeaderElectResourceName      = "leader-elect-resource-name"
	flagLeaderElectResourceNamespace = "leader-elect-resource-namespace"
	flagLeaderElectRetryPeriod       = "leader-elect-retry-period"
	flagSyncPeriod                   = "sync-period"

	defaultMetricsAddr                  = ":8080"
	defaultHealthProbeBindAddress       = ":10258"
	defaultLeaderElect                  = true
	defaultLeaderElectLeaseDuration     = 15 * time.Second
	defaultElectRenewDeadline           = 10 * time.Second
	defaultLeaderElectRetryPeriod       = 2 * time.Second
	defaultLeaderElectResourceLock      = "endpointsleases"
	defaultLeaderElectResourceName      = "ccm"
	defaultLeaderElectResourceNamespace = "kube-system"
	defaultSyncPeriod                   = 60 * time.Minute
	defaultQPS                          = 20.0
	defaultBurst                        = 30
)

// RuntimeConfig stores the configuration for controller-runtime
type RuntimeConfig struct {
	MetricsBindAddress           string
	HealthProbeBindAddress       string
	LeaderElect                  bool
	LeaderElectLeaseDuration     time.Duration
	LeaderElectRenewDeadline     time.Duration
	LeaderElectRetryPeriod       time.Duration
	LeaderElectResourceLock      string
	LeaderElectResourceName      string
	LeaderElectResourceNamespace string
	SyncPeriod                   time.Duration
	QPS                          float32
	Burst                        int
}

func (c *RuntimeConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.MetricsBindAddress, flagMetricsBindAddr, defaultMetricsAddr, "The address the metric endpoint binds to.")
	fs.StringVar(&c.HealthProbeBindAddress, flagHealthProbeBindAddr, defaultHealthProbeBindAddress, "The address the health probes binds to.")
	fs.Float32Var(&c.QPS, flagQPS, defaultQPS, "QPS to use while talking with kubernetes apiserver.")
	fs.IntVar(&c.Burst, flagBurst, defaultBurst, "Burst to use while talking with kubernetes apiserver.")
	fs.BoolVar(&c.LeaderElect, flagLeaderElect, defaultLeaderElect,
		"Start a leader election client and gain leadership before executing the main loop. Enable this when running replicated components for high availability.")
	fs.DurationVar(&c.LeaderElectLeaseDuration, flagLeaderElectLeaseDuration, defaultLeaderElectLeaseDuration,
		"he duration that non-leader candidates will wait after observing a leadership renewal until attempting to acquire "+
			"leadership of a led but unrenewed leader slot. This is effectively the maximum duration that a leader can be stopped"+
			" before it is replaced by another candidate. This is only applicable if leader election is enabled.")
	fs.DurationVar(&c.LeaderElectRenewDeadline, flagLeaderElectRenewDeadline, defaultElectRenewDeadline,
		"The interval between attempts by the acting master to renew a leadership slot before it stops leading. This"+
			"must be less than or equal to the lease duration. This is only applicable if leader election is enabled.")
	fs.DurationVar(&c.LeaderElectRetryPeriod, flagLeaderElectRetryPeriod, defaultLeaderElectRetryPeriod,
		"The duration the clients should wait between attempting acquisition and renewal of a leadership. This is only"+
			"applicable if leader election is enabled.")
	fs.StringVar(&c.LeaderElectResourceLock, flagLeaderElectResourceLock, defaultLeaderElectResourceLock,
		"The type of resource object that is used for locking during leader election. Supported options are "+
			"'endpoints', 'configmaps', 'leases', 'endpointsleases' and 'configmapsleases'")
	fs.StringVar(&c.LeaderElectResourceName, flagLeaderElectResourceName, defaultLeaderElectResourceName,
		"The name of resource object that is used for locking during leader election. ")
	fs.StringVar(&c.LeaderElectResourceNamespace, flagLeaderElectResourceNamespace, defaultLeaderElectResourceNamespace,
		"The namespace of resource object that is used for locking during leader election.")
	fs.DurationVar(&c.SyncPeriod, flagSyncPeriod, defaultSyncPeriod,
		"Period at which the controller forces the repopulation of its local object stores.")

}

func BuildRuntimeOptions(rtCfg RuntimeConfig) manager.Options {
	return manager.Options{
		MetricsBindAddress:         rtCfg.MetricsBindAddress,
		HealthProbeBindAddress:     rtCfg.HealthProbeBindAddress,
		LeaderElection:             rtCfg.LeaderElect,
		LeaderElectionID:           rtCfg.LeaderElectResourceName,
		LeaderElectionResourceLock: rtCfg.LeaderElectResourceLock,
		LeaderElectionNamespace:    rtCfg.LeaderElectResourceNamespace,
		LeaseDuration:              &rtCfg.LeaderElectLeaseDuration,
		RenewDeadline:              &rtCfg.LeaderElectRenewDeadline,
		RetryPeriod:                &rtCfg.LeaderElectRetryPeriod,
		SyncPeriod:                 &rtCfg.SyncPeriod,
	}
}
