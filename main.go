package main

import (
	"fmt"
	"os"

	_ "github.com/AliyunContainerService/alicloud-controller-manager/cloudprovider/alicloud"
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app/options"
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus"
	_ "k8s.io/kubernetes/pkg/version/prometheus"
	"k8s.io/kubernetes/pkg/version/verflag"
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

	//cloud, err := cloudprovider.InitCloudProvider("alicloud", s.CloudConfigFile)
	//if err != nil {
	//	glog.Fatalf("Alibaba cloud provider could not be initialized: %v", err)
	//}

	if err := app.Run(s); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)

	}
}
