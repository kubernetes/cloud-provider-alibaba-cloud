package app

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	alicloud "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils/metric"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/klog"
	ccfg "k8s.io/kubernetes/cmd/cloud-controller-manager/app/apis/config"
	genericcontrollermanager "k8s.io/kubernetes/cmd/controller-manager/app"
	kubectrlmgrconfig "k8s.io/kubernetes/pkg/controller/apis/config"
	serviceconfig "k8s.io/kubernetes/pkg/controller/service/config"
	"k8s.io/kubernetes/pkg/master/ports"
	"runtime"
	"time"

	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	"k8s.io/cloud-provider"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/route"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/service"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/controller"
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
	ccfg.CloudControllerManagerConfiguration
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
		CloudControllerManagerConfiguration: ccfg.CloudControllerManagerConfiguration{
			Generic: kubectrlmgrconfig.GenericControllerManagerConfiguration{
				Port:            ports.CloudControllerManagerPort,
				Address:         "0.0.0.0",
				MinResyncPeriod: metav1.Duration{Duration: 5 * time.Minute},
				ClientConnection: componentbaseconfig.ClientConnectionConfiguration{
					ContentType: "application/vnd.kubernetes.protobuf",
					QPS:         20.0,
					Burst:       30,
				},
				LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
					LeaderElect:   false,
					LeaseDuration: metav1.Duration{Duration: 15 * time.Second},
					RenewDeadline: metav1.Duration{Duration: 10 * time.Second},
					RetryPeriod:   metav1.Duration{Duration: 2 * time.Second},
				},
				ControllerStartInterval: metav1.Duration{Duration: 0 * time.Second},
			},
			KubeCloudShared: kubectrlmgrconfig.KubeCloudSharedConfiguration{
				NodeMonitorPeriod:         metav1.Duration{Duration: 5 * time.Second},
				ClusterName:               "kubernetes",
				ConfigureCloudRoutes:      true,
				RouteReconciliationPeriod: metav1.Duration{Duration: 10 * time.Second},
			},
			ServiceController: serviceconfig.ServiceControllerConfiguration{
				ConcurrentServiceSyncs: 3,
			},
		},
		NodeStatusUpdateFrequency: metav1.Duration{Duration: 5 * time.Minute},
	}
	ccm.Generic.LeaderElection.LeaderElect = true
	return &ccm
}

func createRecorder(client *kubernetes.Clientset) record.EventRecorder {
	cast := record.NewBroadcaster()
	cast.StartLogging(klog.Infof)
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

	kubeconfig.QPS = ccm.Generic.ClientConnection.QPS
	kubeconfig.Burst = int(ccm.Generic.ClientConnection.Burst)
	kubeconfig.ContentConfig.ContentType = ccm.Generic.ClientConnection.ContentType

	client, err := kubernetes.NewForConfig(
		rest.AddUserAgent(
			kubeconfig,
			"cloud-controller-manager",
		),
	)
	if err != nil {
		klog.Fatalf("invalid API configuration: %v", err)
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
	if ccm.KubeCloudShared.CloudProvider.Name == "" {
		return fmt.Errorf("--cloud-provider cannot be empty")
	}

	alicloud.CloudConfigFile = ccm.KubeCloudShared.CloudProvider.CloudConfigFile
	cloud, err := cloudprovider.InitCloudProvider(
		ccm.KubeCloudShared.CloudProvider.Name,
		ccm.KubeCloudShared.CloudProvider.CloudConfigFile,
	)
	if err != nil {
		return fmt.Errorf("cloud provider could not be initialized: %v", err)
	}

	ccm.cloud = cloud
	if !cloud.HasClusterID() {
		if ccm.KubeCloudShared.AllowUntaggedCloud {
			klog.Warning("detected a cluster without a ClusterID.  A ClusterID will " +
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
	cfg.Set(ccm.CloudControllerManagerConfiguration)

	ccm.client, ccm.election, ccm.restConfig, err = client(ccm)
	if err != nil {
		return fmt.Errorf("create client error: %s", err.Error())
	}

	if err := setResourceLock(ccm); err != nil {
		return fmt.Errorf("set resource lock error: %s", err.Error())
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
		if ccm.Generic.Debugging.EnableProfiling {
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			if ccm.Generic.Debugging.EnableContentionProfiling {
				runtime.SetBlockProfileRate(1)
			}
		}
		configz.InstallHandler(mux)
		metric.RegisterPrometheus()
		mux.Handle("/metrics", promhttp.Handler())
		server := &http.Server{
			Addr:    net.JoinHostPort(ccm.Generic.Address, strconv.Itoa(int(ccm.Generic.Port))),
			Handler: mux,
		}
		klog.Fatal(server.ListenAndServe())
	}()
	return nil
}

func (ccm *ServerCCM) MainLoop(ctx context.Context) {
	var (
		builder controller.ControllerClientBuilder
	)
	builder = controller.SimpleControllerClientBuilder{
		ClientConfig: ccm.restConfig,
	}
	if ccm.KubeCloudShared.UseServiceAccountCredentials {
		builder = controller.SAControllerClientBuilder{
			ClientConfig:         rest.AnonymousClientConfig(ccm.restConfig),
			CoreClient:           ccm.client.CoreV1(),
			AuthenticationClient: ccm.client.AuthenticationV1(),
			Namespace:            "kube-system",
		}
	}
	panic(fmt.Sprintf("unreachable: %v", RunControllers(ccm, builder, nil)))
}

// Run runs the ExternalCMServer.  This should never exit.
func Run(ccm *ServerCCM) error {
	if err := ccm.Start(); err != nil {
		return err
	}

	route.Options = route.RoutesOptions{
		ClusterCIDR:               ccm.KubeCloudShared.ClusterCIDR,
		AllocateNodeCIDRs:         ccm.KubeCloudShared.AllocateNodeCIDRs,
		ConfigCloudRoutes:         ccm.KubeCloudShared.ConfigureCloudRoutes,
		RouteReconciliationPeriod: ccm.KubeCloudShared.RouteReconciliationPeriod,
		ControllerStartInterval:   ccm.Generic.ControllerStartInterval,
	}

	if !ccm.Generic.LeaderElection.LeaderElect {
		ccm.MainLoop(context.TODO())
	}

	// Identity used to distinguish between multiple cloud controller manager instances
	id, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("hostname: %s", err.Error())
	}

	// Lock required for leader election
	rl, err := resourcelock.New(
		ccm.Generic.LeaderElection.ResourceLock,
		"kube-system",
		"ccm",
		ccm.election.CoreV1(),
		ccm.election.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      "ccm-" + id,
			EventRecorder: ccm.recorder,
		})
	if err != nil {
		return fmt.Errorf("creating etcd resource lock: %v", err)
	}

	// Try and become the leader and start cloud controller manager loops
	leaderelection.RunOrDie(
		context.Background(),
		leaderelection.LeaderElectionConfig{
			Lock: rl,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: ccm.MainLoop,
				OnStoppedLeading: func() {
					klog.Fatalf("leaderelection lost")
				},
			},
			LeaseDuration: ccm.Generic.LeaderElection.LeaseDuration.Duration,
			RenewDeadline: ccm.Generic.LeaderElection.RenewDeadline.Duration,
			RetryPeriod:   ccm.Generic.LeaderElection.RetryPeriod.Duration,
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
		ccm.cloud.Initialize(clientBuilder, stop)
	}
	client := clientBuilder.ClientOrDie("shared-informers")

	ifactory := informers.NewSharedInformerFactory(client, resyncPeriod(ccm)())

	//if err := runControllerPV(ccm, clientBuilder, stop); err != nil {
	//	return fmt.Errorf("run pvcontroller: %s", err.Error())
	//}
	time.Sleep(wait.Jitter(ccm.Generic.ControllerStartInterval.Duration, ControllerStartJitter))

	if err := runControllerService(ccm, clientBuilder, ifactory, stop); err != nil {
		return fmt.Errorf("run service controller: %s", err.Error())
	}

	time.Sleep(wait.Jitter(ccm.Generic.ControllerStartInterval.Duration, ControllerStartJitter))

	// If apiserver is not running we should wait for some time and fail
	// only then. This is particularly important when we start apiserver
	// and controller manager at the same time.
	err := genericcontrollermanager.WaitForAPIServer(ccm.client, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get api versions from server: %v", err)
	}

	ifactory.Start(stop)
	klog.Infof("informer started")

	select {}
}

//func runControllerPV(
//	ccm *ServerCCM,
//	builder controller.ControllerClientBuilder,
//	stop <-chan struct{},
//) error {
//
//	con := ccontroller.NewPersistentVolumeLabelController(
//		builder.ClientOrDie("pvl-controller"),
//		ccm.cloud,
//	)
//	go con.Run(5, stop)
//	return nil
//}

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
		builder.ClientOrDie("cloud-controller-manager"),
		informer,
		ccm.KubeCloudShared.ClusterName,
	)
	if err != nil {
		return fmt.Errorf("failed to start service controller: %v", err)
	}
	go scon.Run(stop, int(ccm.ServiceController.ConcurrentServiceSyncs))
	return nil
}

func resyncPeriod(ccm *ServerCCM) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(ccm.Generic.MinResyncPeriod.Nanoseconds()) * factor)
	}
}

func setResourceLock(ccm *ServerCCM) error {
	// check kubernetes version to use lease or not
	version114, _ := version.ParseGeneric("v1.14.0")
	serverVersion, err := ccm.client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("get kubernetes version error: %s", err.Error())
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		return fmt.Errorf("unexpected error parsing running Kubernetes version, %s", err.Error())
	}

	if runningVersion.AtLeast(version114) {
		ccm.CloudControllerManagerConfiguration.Generic.LeaderElection.ResourceLock = resourcelock.LeasesResourceLock
	} else {
		klog.V(5).Infof("kubernetes version: %s, resource lock use EndpointsResourceLock", runningVersion)
		ccm.CloudControllerManagerConfiguration.Generic.LeaderElection.ResourceLock = resourcelock.EndpointsResourceLock
	}

	return nil
}
