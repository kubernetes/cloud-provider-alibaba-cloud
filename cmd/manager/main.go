package main

import (
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/version"
	"k8s.io/klog/klogr"
	"net/http"
	"os"
	"runtime"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/apis"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	ctrl "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/health"
	ctrlCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller"
)

var log = klogr.New()

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	err := loadControllerConfig()
	if err != nil {
		log.Error(err, "unable to load controller config")
		os.Exit(1)
	}
	ctrl.SetLogger(klogr.New())

	printVersion()

	// Get a config to talk to the api-server
	cfg := config.GetConfigOrDie()

	// Create a new manager to provide shared dependencies and start components
	mgr, err := manager.New(cfg, ctrlCtx.BuildRuntimeOptions(ctrlCtx.ControllerCFG.RuntimeConfig))
	if err != nil {
		log.Error(err, "fail to create manager")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "add apis to schema: %s", err.Error())
		os.Exit(1)
	}

	var cloud prvd.Provider
	if ctrlCtx.ControllerCFG.DryRun {
		log.Info("using DryRun Mode")
		cloud = dryrun.NewDryRunCloud()
	} else {
		cloud = alibaba.NewAlibabaCloud()
	}
	ctx := shared.NewSharedContext(cloud)

	log.Info("start to register crds")
	err = controller.RegisterCRD(cfg)
	if err != nil {
		log.Error(err, "register crd: %s", err.Error())
		os.Exit(1)
	}

	log.Info("Registering Components.")
	if err := controller.AddToManager(mgr, ctx, ctrlCtx.ControllerCFG.EnableControllers); err != nil {
		log.Error(err, "add controller: %s", err.Error())
		os.Exit(1)
	} else {
		log.Info(fmt.Sprintf("Loaded controllers: %v", ctrlCtx.ControllerCFG.EnableControllers))
	}

	// Start the Cmd
	log.Info("Starting the Cmd.")
	if err := mgr.AddHealthzCheck("default", func(req *http.Request) error {
		errs := make([]error, 0)
		for _, fun := range health.CheckFuncList {
			if err := fun.Check(); err != nil {
				errs = append(errs, err)
			}
		}
		return utilerrors.NewAggregate(errs)
	}); err != nil {
		log.Error(err, "failed to add default health check: %w", err.Error())
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero: %s", err.Error())
		os.Exit(1)
	}

}

func loadControllerConfig() error {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	fs.AddGoFlagSet(flag.CommandLine)
	ctrlCtx.ControllerCFG.BindFlags(fs)

	if err := fs.Parse(os.Args); err != nil {
		return err
	}

	if err := ctrlCtx.ControllerCFG.Validate(); err != nil {
		return err
	}
	return nil
}
