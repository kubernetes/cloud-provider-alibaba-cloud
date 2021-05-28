package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/apis"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	ctrl "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"k8s.io/cloud-provider-alibaba-cloud/cmd/health"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller"
	"k8s.io/cloud-provider-alibaba-cloud/version"
)

// Change below variables to serve metric on different host or port.
var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8089
	operatorMetricsPort int32 = 8088
	healthAddr                = ":8087"
)

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.StringVar(
		&ctx2.GlobalFlag.LogLevel,
		"loglevel",
		log.InfoLevel.String(),
		"logrus log level ",
	)
	pflag.CommandLine.StringVar(
		&ctx2.GlobalFlag.CloudConfig,
		"cloud-config",
		"",
		"cloud config file for ncl",
	)

	pflag.CommandLine.BoolVar(
		&ctx2.GlobalFlag.EnableLeaderSelect,
		"enable-leader-select",
		true,
		"enable leader select or not",
	)
	pflag.CommandLine.StringSliceVar(
		&ctx2.GlobalFlag.EnableControllers,
		"enable-controllers",
		[]string{},
		"controllers to enable, e.g., node, route, service, ingress, pvtz, default ['node','route','service']",
	)
	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	ctrl.SetLogger(zap.New())

	pflag.Parse()

	// let's explicitly set stdout
	log.SetOutput(os.Stdout)
	// this formatter is the default, but the timestamps output aren't
	// particularly useful, they're relative to the command start
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
		// we force colors because this only forces over the isTerminal check
		// and this will not be accurately checkable later on when we wrap
		// the logger output with our logutil.StatusFriendlyWriter
		//ForceColors: logutil.IsTerminal(log.StandardLogger().Out),
	})

	log.SetLevel(log.InfoLevel)

	printVersion()

	if ctx2.GlobalFlag.CloudConfig == "" {
		log.Errorf("config file must be provided for ak. --config")
		os.Exit(1)
	}
	// Get a config to talk to the api-server
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Become the leader before proceeding
	if ctx2.GlobalFlag.EnableLeaderSelect {
		err = leader.Become(context.TODO(), "ccm-lock")
		if err != nil {
			log.Errorf("leader error: %s", err.Error())
			os.Exit(1)
		}
	}
	log.Info("start to register crds")
	err = controller.RegisterCRD(cfg)
	if err != nil {
		log.Errorf("register crd: %s", err.Error())
		os.Exit(1)
	}

	// Set default manager options
	options := manager.Options{
		MetricsBindAddress:     fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		HealthProbeBindAddress: healthAddr,
	}

	// Set ReSync period
	syncPeriod := 3 * time.Minute
	options.SyncPeriod = &syncPeriod

	// Create a new manager to provide shared dependencies and start components
	mgr, err := manager.New(cfg, options)
	if err != nil {
		log.Errorf("create manager: %s", err.Error())
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "add apis to schema: %s", err.Error())
		os.Exit(1)
	}

	ctx := shared.NewSharedContext(alibaba.NewAlibabaCloud())
	// Setup all Controllers
	if err := controller.AddToManager(mgr, ctx, ctx2.GlobalFlag.EnableControllers); err != nil {
		log.Errorf("add controller: %s", err.Error())
		os.Exit(1)
	} else {
		log.Infof("Loaded controllers: %s\n", ctx2.GlobalFlag.EnableControllers)
	}

	// Add the Metrics Service
	addMetrics(context.TODO(), cfg)

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.AddHealthzCheck("default", func(req *http.Request) error {
		errs := make([]error, 0)
		for _, fun := range health.CheckFuncList {
			if err := fun.Check(); err != nil {
				errs = append(errs, err)
			}
		}
		return utilerrors.NewAggregate(errs)
	}); err != nil {
		log.Errorf("failed to add default health check: %w", err.Error())
		os.Exit(1)
	}
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Errorf("Manager exited non-zero: %s", err.Error())
		os.Exit(1)
	}
}

// addMetrics will create the Services and Service Monitors to allow the operator export the metric by using
// the Prometheus operator
func addMetrics(ctx context.Context, cfg *rest.Config) {
	// Get the namespace the operator is currently deployed in.
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		if errors.Is(err, k8sutil.ErrRunLocal) {
			log.Info("Skipping CR metric server creation; not running in a cluster.")
			return
		}
	}

	// Add to the below struct any other metric ports you want to expose.
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	}

	// Create Service object to expose the metric port(s).
	service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		log.Infof("Could not create metric Service: %s", err.Error())
	}

	// CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
	// necessary to configure Prometheus to scrape metric from this operator.
	services := []*v1.Service{service}

	// The ServiceMonitor is created in the same namespace where the operator is deployed
	_, err = metrics.CreateServiceMonitors(cfg, operatorNs, services)
	if err != nil {
		log.Infof("Could not create ServiceMonitor object: %s", err.Error())
		// If this operator is deployed to a cluster without the prometheus-operator running, it will return
		// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
		if err == metrics.ErrServiceMonitorNotPresent {
			log.Infof("Install prometheus-operator in your cluster to create ServiceMonitor objects: %s", err.Error())
		}
	}
}
