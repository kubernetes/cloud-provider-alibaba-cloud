package app

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ccfg "k8s.io/kubernetes/pkg/apis/componentconfig"
	"k8s.io/kubernetes/pkg/master/ports"
	"runtime"
	"time"

	"fmt"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1c "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/route"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/service"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/client/leaderelectionconfig"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	ccontroller "k8s.io/kubernetes/pkg/controller/cloud"
	"k8s.io/kubernetes/pkg/util/configz"
	"math/rand"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
)

const (
	// ControllerStartJitter is the jitter value used when starting controller managers.
	ControllerStartJitter = 1.0
)

// ServerCCM is the main context object for the controller manager.
type ServerCCM struct {
	ccfg.KubeControllerManagerConfiguration
	restConfig *rest.Config
	election   *kubernetes.Clientset
	client     *kubernetes.Clientset
	recorder   record.EventRecorder
	Master     string
	Kubeconfig string
	cloud      cloudprovider.Interface

	// NodeStatusUpdateFrequency is the frequency at which the controller
	// updates nodes' status
	NodeStatusUpdateFrequency metav1.Duration
}

// NewServerCCM creates a new ExternalCMServer with a default config.
func NewServerCCM() *ServerCCM {
	ccm := ServerCCM{
		// Part of these default values also present in 'cmd/kube-controller-manager/app/options/options.go'.
		// Please keep them in sync when doing update.
		KubeControllerManagerConfiguration: ccfg.KubeControllerManagerConfiguration{
			Port:                      ports.CloudControllerManagerPort,
			Address:                   "0.0.0.0",
			ConcurrentServiceSyncs:    1,
			MinResyncPeriod:           metav1.Duration{Duration: 12 * time.Hour},
			NodeMonitorPeriod:         metav1.Duration{Duration: 5 * time.Second},
			ClusterName:               "kubernetes",
			ConfigureCloudRoutes:      true,
			ContentType:               "application/vnd.kubernetes.protobuf",
			KubeAPIQPS:                20.0,
			KubeAPIBurst:              30,
			LeaderElection:            leaderelectionconfig.DefaultLeaderElectionConfiguration(),
			ControllerStartInterval:   metav1.Duration{Duration: 0 * time.Second},
			RouteReconciliationPeriod: metav1.Duration{Duration: 10 * time.Second},
		},
		NodeStatusUpdateFrequency: metav1.Duration{Duration: 5 * time.Minute},
	}
	ccm.LeaderElection.LeaderElect = true
	return &ccm
}

func createRecorder(client *kubernetes.Clientset) record.EventRecorder {
	cast := record.NewBroadcaster()
	cast.StartLogging(glog.Infof)
	cast.StartRecordingToSink(
		&v1c.EventSinkImpl{
			Interface: v1c.New(client.CoreV1().RESTClient()).Events(""),
		},
	)
	return cast.NewRecorder(
		legacyscheme.Scheme,
		v1.EventSource{Component: "cloud-controller-manager"},
	)
}

func client(ccm *ServerCCM) (*kubernetes.Clientset, *kubernetes.Clientset, *rest.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(ccm.Master, ccm.Kubeconfig)
	if err != nil {
		return nil, nil, kubeconfig, err
	}

	kubeconfig.QPS = ccm.KubeAPIQPS
	kubeconfig.Burst = int(ccm.KubeAPIBurst)
	kubeconfig.ContentConfig.ContentType = ccm.ContentType

	client, err := kubernetes.NewForConfig(
		rest.AddUserAgent(
			kubeconfig,
			"cloud-controller-manager",
		),
	)
	if err != nil {
		glog.Fatalf("invalid API configuration: %v", err)
	}
	leader := kubernetes.NewForConfigOrDie(
		rest.AddUserAgent(
			kubeconfig,
			"leader-election",
		),
	)
	return client, leader, kubeconfig, nil
}

func (ccm *ServerCCM) initialization() error {
	if ccm.CloudProvider == "" {
		return fmt.Errorf("--cloud-provider cannot be empty")
	}

	cloud, err := cloudprovider.InitCloudProvider(ccm.CloudProvider, ccm.CloudConfigFile)
	if err != nil {
		return fmt.Errorf("cloud provider could not be initialized: %v", err)
	}
	ccm.cloud = cloud
	if cloud.HasClusterID() == false {
		if ccm.AllowUntaggedCloud == true {
			glog.Warning("detected a cluster without a ClusterID.  A ClusterID will " +
				"be required in the future.  Please tag your cluster to avoid any future issues")
		} else {
			return fmt.Errorf("no ClusterID found.  A ClusterID is required for the " +
				"cloud provider to function properly.  This check can be bypassed by " +
				"setting the allow-untagged-cloud option")
		}
	}

	cfg, err := configz.New("componentconfig")
	if err != nil {
		return fmt.Errorf("unable to register configz: %v", err)
	}
	cfg.Set(ccm.KubeControllerManagerConfiguration)

	ccm.client, ccm.election, ccm.restConfig, err = client(ccm)
	if err != nil {
		return fmt.Errorf("create client error: %s", err.Error())
	}
	ccm.recorder = createRecorder(ccm.client)
	return err
}

func (ccm *ServerCCM) Start() error {
	if err := ccm.initialization(); err != nil {
		return fmt.Errorf("verify ccm config: %s", err.Error())
	}

	// Start the external controller manager server
	go func() {
		mux := http.NewServeMux()
		healthz.InstallHandler(mux)
		if ccm.EnableProfiling {
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			if ccm.EnableContentionProfiling {
				runtime.SetBlockProfileRate(1)
			}
		}
		configz.InstallHandler(mux)
		mux.Handle("/metrics", prometheus.Handler())

		server := &http.Server{
			Addr:    net.JoinHostPort(ccm.Address, strconv.Itoa(int(ccm.Port))),
			Handler: mux,
		}
		glog.Fatal(server.ListenAndServe())
	}()
	return nil
}

func (ccm *ServerCCM) MainLoop(stop <-chan struct{}) {
	var (
		builder controller.ControllerClientBuilder
	)
	builder = controller.SimpleControllerClientBuilder{
		ClientConfig: ccm.restConfig,
	}
	if ccm.UseServiceAccountCredentials {
		builder = controller.SAControllerClientBuilder{
			ClientConfig:         rest.AnonymousClientConfig(ccm.restConfig),
			CoreClient:           ccm.client.CoreV1(),
			AuthenticationClient: ccm.client.Authentication(),
			Namespace:            "kube-system",
		}
	}

	panic(fmt.Sprintf("unreachable: %v", RunControllers(ccm, builder, stop)))
}

// Run runs the ExternalCMServer.  This should never exit.
func Run(ccm *ServerCCM) error {
	if err := ccm.Start(); err != nil {
		return err
	}

	route.Options = route.RoutesOptions{
		ClusterCIDR:               ccm.ClusterCIDR,
		AllocateNodeCIDRs:         ccm.AllocateNodeCIDRs,
		ConfigCloudRoutes:         ccm.ConfigureCloudRoutes,
		RouteReconciliationPeriod: ccm.RouteReconciliationPeriod,
		ControllerStartInterval:   ccm.ControllerStartInterval,
	}

	if !ccm.LeaderElection.LeaderElect {
		ccm.MainLoop(nil)
	}

	// Identity used to distinguish between multiple cloud controller manager instances
	id, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("hostname: %s", err.Error())
	}

	// Lock required for leader election
	rl, err := resourcelock.New(
		ccm.LeaderElection.ResourceLock,
		"kube-system",
		"ccm",
		ccm.election.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      "ccm-" + id,
			EventRecorder: ccm.recorder,
		})
	if err != nil {
		return fmt.Errorf("creating etcd resource lock: %v", err)
	}

	// Try and become the leader and start cloud controller manager loops
	leaderelection.RunOrDie(
		leaderelection.LeaderElectionConfig{
			Lock: rl,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: ccm.MainLoop,
				OnStoppedLeading: func() {
					glog.Fatalf("leaderelection lost")
				},
			},
			LeaseDuration: ccm.LeaderElection.LeaseDuration.Duration,
			RenewDeadline: ccm.LeaderElection.RenewDeadline.Duration,
			RetryPeriod:   ccm.LeaderElection.RetryPeriod.Duration,
		},
	)
	panic("unreachable")
}

// RunControllers starts the cloud specific controller loops.
func RunControllers(
	ccm *ServerCCM,
	clientBuilder controller.ControllerClientBuilder,
	stop <-chan struct{},
) error {

	if ccm.cloud != nil {
		// Initialize the cloud provider with a reference to the clientBuilder
		ccm.cloud.Initialize(clientBuilder)
	}
	client := clientBuilder.ClientOrDie("shared-informers")

	ifactory := informers.NewSharedInformerFactory(client, resyncPeriod(ccm)())

	if err := runControllerPV(ccm, clientBuilder, stop); err != nil {
		return fmt.Errorf("run pvcontroller: %s", err.Error())
	}
	time.Sleep(wait.Jitter(ccm.ControllerStartInterval.Duration, ControllerStartJitter))

	if err := runControllerService(ccm, clientBuilder, ifactory, stop); err != nil {
		return fmt.Errorf("run service controller: %s", err.Error())
	}

	time.Sleep(wait.Jitter(ccm.ControllerStartInterval.Duration, ControllerStartJitter))

	// If apiserver is not running we should wait for some time and fail
	// only then. This is particularly important when we start apiserver
	// and controller manager at the same time.
	err := wait.PollImmediate(
		time.Second,
		10*time.Second,
		func() (bool, error) {
			_, err := rest.ServerAPIVersions(ccm.restConfig)
			if err == nil {
				return true, nil
			}
			glog.Errorf("failed to get api versions from server: %v", err)
			return false, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get api versions from server: %v", err)
	}

	ifactory.Start(stop)
	glog.Infof("informer started")

	select {}
}

func runControllerPV(
	ccm *ServerCCM,
	builder controller.ControllerClientBuilder,
	stop <-chan struct{},
) error {

	con := ccontroller.NewPersistentVolumeLabelController(
		builder.ClientOrDie("pvl-controller"),
		ccm.cloud,
	)
	go con.Run(5, stop)
	return nil
}

func runControllerService(
	ccm *ServerCCM,
	builder controller.ControllerClientBuilder,
	informer informers.SharedInformerFactory,
	stop <-chan struct{},
) error {
	cloudslb, ok := ccm.cloud.LoadBalancer()
	if !ok {
		return fmt.Errorf("loadbalancer interface must be implemented")
	}

	scon, err := service.NewController(
		cloudslb,
		builder.ClientOrDie("service-controller"),
		informer,
		ccm.ClusterName,
	)
	if err != nil {
		return fmt.Errorf("failed to start service controller: %v", err)
	}
	go scon.Run(stop, int(ccm.ConcurrentServiceSyncs))
	return nil
}

func resyncPeriod(ccm *ServerCCM) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(ccm.MinResyncPeriod.Nanoseconds()) * factor)
	}
}
