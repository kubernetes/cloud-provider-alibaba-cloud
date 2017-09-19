package main

import (
	"fmt"
	"os"

	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app/options"
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus"
	"k8s.io/kubernetes/pkg/cloudprovider"
	_ "k8s.io/kubernetes/pkg/cloudprovider/providers"
	_ "k8s.io/kubernetes/pkg/version/prometheus"
	"k8s.io/kubernetes/pkg/version/verflag"

	_ "github.com/AliyunContainerService/alicloud-controller-manager/alicloud"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

func init() {
	healthz.DefaultHealthz()

}

func main() {
	s := options.NewCloudControllerManagerServer()
	s.AddFlags(pflag.CommandLine)

	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	verflag.PrintAndExitIfRequested()

	cloud, err := cloudprovider.InitCloudProvider("alicloud", s.CloudConfigFile)
	if err != nil {
		glog.Fatalf("Alibaba cloud provider could not be initialized: %v", err)

	}

	if err := app.Run(s, cloud); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)

	}
}
