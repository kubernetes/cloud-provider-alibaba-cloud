package config

import (
	"time"

	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	flagMetricsBindAddr              = "metrics-bind-addr"
	flagHealthProbeBindAddr          = "health-probe-bind-addr"
	flagPprofBindAddr                = "pprof-bind-addr"
	flagQPS                          = "kube-api-qps"
	flagBurst                        = "kube-api-burst"
	flagLeaderElect                  = "leader-elect"
	flagLeaderElectLeaseDuration     = "leader-elect-lease-duration"
	flagLeaderElectRenewDeadline     = "leader-elect-renew-deadline"
	flagLeaderElectResourceLock      = "leader-elect-resource-lock"
	flagLeaderElectResourceName      = "leader-elect-resource-name"
	flagLeaderElectResourceNamespace = "leader-elect-resource-namespace"
	flagLeaderElectRetryPeriod       = "leader-elect-retry-period"

	defaultMetricsAddr                  = ":8080"
	defaultHealthProbeBindAddr          = ":10258"
	defaultPprofBindAddr                = ":6060"
	defaultLeaderElect                  = true
	defaultLeaderElectLeaseDuration     = 15 * time.Second
	defaultElectRenewDeadline           = 10 * time.Second
	defaultLeaderElectRetryPeriod       = 2 * time.Second
	defaultLeaderElectResourceLock      = "leases"
	defaultLeaderElectResourceName      = "ccm"
	defaultLeaderElectResourceNamespace = "kube-system"
	defaultQPS                          = 20.0
	defaultBurst                        = 30
)

// RuntimeConfig stores the configuration for controller-runtime
type RuntimeConfig struct {
	MetricsBindAddress           string
	HealthProbeBindAddress       string
	PprofBindAddress             string
	LeaderElect                  bool
	LeaderElectLeaseDuration     time.Duration
	LeaderElectRenewDeadline     time.Duration
	LeaderElectRetryPeriod       time.Duration
	LeaderElectResourceLock      string
	LeaderElectResourceName      string
	LeaderElectResourceNamespace string
	QPS                          float32
	Burst                        int
}

func (c *RuntimeConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.MetricsBindAddress, flagMetricsBindAddr, defaultMetricsAddr, "The address the metric endpoint binds to.")
	fs.StringVar(&c.HealthProbeBindAddress, flagHealthProbeBindAddr, defaultHealthProbeBindAddr, "The address the health probes binds to.")
	fs.StringVar(&c.PprofBindAddress, flagPprofBindAddr, defaultPprofBindAddr,
		"The address that the controller should bind to for serving pprof. It can be set to '' or 0 to disable the pprof serving.")
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
}

func BuildRuntimeOptions(rtCfg RuntimeConfig) manager.Options {
	return manager.Options{
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&v1.Node{}: {
					Transform: stripNodeFields(),
				},
				&v1.Pod{}: {
					Transform: stripPodFields(),
				},
			},
		},
		ClientDisableCacheFor: []client.Object{
			&v1.Node{},
			&v1.Service{},
			&v1.Endpoints{},
			&discovery.EndpointSlice{},
		},
		MetricsBindAddress:         rtCfg.MetricsBindAddress,
		HealthProbeBindAddress:     rtCfg.HealthProbeBindAddress,
		PprofBindAddress:           rtCfg.PprofBindAddress,
		LeaderElection:             rtCfg.LeaderElect,
		LeaderElectionID:           rtCfg.LeaderElectResourceName,
		LeaderElectionResourceLock: rtCfg.LeaderElectResourceLock,
		LeaderElectionNamespace:    rtCfg.LeaderElectResourceNamespace,
		LeaseDuration:              &rtCfg.LeaderElectLeaseDuration,
		RenewDeadline:              &rtCfg.LeaderElectRenewDeadline,
		RetryPeriod:                &rtCfg.LeaderElectRetryPeriod,
	}
}

func stripNodeFields() toolscache.TransformFunc {
	return func(obj interface{}) (interface{}, error) {
		node, ok := obj.(*v1.Node)
		if !ok {
			return obj, nil
		}
		node.Status.Images = nil
		node.ManagedFields = nil
		return node, nil
	}
}

func stripPodFields() toolscache.TransformFunc {
	return func(obj interface{}) (interface{}, error) {
		pod, ok := obj.(*v1.Pod)
		if !ok {
			return obj, nil
		}
		pod.ManagedFields = nil
		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].Env = nil
			pod.Spec.Containers[i].VolumeMounts = nil
		}
		for i := range pod.Spec.InitContainers {
			pod.Spec.InitContainers[i].Env = nil
			pod.Spec.InitContainers[i].VolumeMounts = nil
		}
		pod.Spec.Volumes = nil
		return pod, nil
	}
}
