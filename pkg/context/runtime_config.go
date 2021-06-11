package context

import (
	"fmt"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

const (
	flagKubeconfig              = "kubeconfig"
	flagAddress                 = "address"
	flagMetricsPort             = "metrics-port"
	flagHealthProbePort         = "health-probe-port"
	flagEnableLeaderElection    = "enable-leader-election"
	flagLeaderElectionID        = "leader-election-id"
	flagLeaderElectionNamespace = "leader-election-namespace"
	flagSyncPeriod              = "sync-period"
	flagQPS                     = "kube-api-qps"
	flagBurst                   = "kube-api-burst"

	defaultKubeConfig              = ""
	defaultAddress                 = "127.0.0.1"
	defaultMetricsPort             = 10259
	defaultHealthProbePort         = 10258
	defaultLeaderElectionID        = "ccm"
	defaultLeaderElectionNamespace = "kube-system"
	defaultSyncPeriod              = 60 * time.Minute
	defaultQPS                     = 20
	defaultBurst                   = 30
)

// RuntimeConfig stores the configuration for controller-runtime
type RuntimeConfig struct {
	KubeConfig              string
	Address                 string
	MetricsPort             int32
	HealthProbePort         int
	EnableLeaderElection    bool
	LeaderElectionID        string
	LeaderElectionNamespace string
	SyncPeriod              time.Duration
	QPS                     int
	Burst                   int
}

func (c *RuntimeConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.KubeConfig, flagKubeconfig, defaultKubeConfig,
		"Path to the kubeconfig file containing authorization and API server information.")
	fs.StringVar(&c.Address, flagAddress, defaultAddress,
		"The IP address to serve on (set to 0.0.0.0 for all interfaces).")
	fs.Int32Var(&c.MetricsPort, flagMetricsPort, defaultMetricsPort,
		"The port the metric endpoints binds to.")
	fs.IntVar(&c.HealthProbePort, flagHealthProbePort, defaultHealthProbePort,
		"The port the health probe endpoints binds to.")
	fs.BoolVar(&c.EnableLeaderElection, flagEnableLeaderElection, true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&c.LeaderElectionID, flagLeaderElectionID, defaultLeaderElectionID,
		"Name of the leader election ID to use for this controller")
	fs.StringVar(&c.LeaderElectionNamespace, flagLeaderElectionNamespace, defaultLeaderElectionNamespace,
		"Name of the leader election ID to use for this controller")
	fs.DurationVar(&c.SyncPeriod, flagSyncPeriod, defaultSyncPeriod,
		"Period at which the controller forces the repopulation of its local object stores.")
	fs.IntVar(&c.QPS, flagQPS, defaultQPS, "QPS to use while talking with kubernetes apiserver.")
	fs.IntVar(&c.Burst, flagBurst, defaultBurst, "Burst to use while talking with kubernetes apiserver.")
}

func BuildRuntimeOptions(rtCfg RuntimeConfig) manager.Options {
	return manager.Options{
		MetricsBindAddress:      fmt.Sprintf("%s:%d", rtCfg.Address, rtCfg.MetricsPort),
		HealthProbeBindAddress:  fmt.Sprintf(":%d", rtCfg.HealthProbePort),
		LeaderElection:          rtCfg.EnableLeaderElection,
		LeaderElectionID:        rtCfg.LeaderElectionID,
		LeaderElectionNamespace: rtCfg.LeaderElectionNamespace,
		SyncPeriod:              &rtCfg.SyncPeriod,
	}
}
