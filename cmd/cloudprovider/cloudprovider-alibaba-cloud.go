/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	f "flag"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog"
	"os"

	"github.com/spf13/pflag"
	_ "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/cloudprovider/app"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/cloudprovider/app/options"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
)

func main() {
	err := f.CommandLine.Parse([]string{})
	if err != nil {
		klog.Warningf("parse command line error: %s", err.Error())
	}

	ccm := app.NewServerCCM()
	options.AddFlags(ccm, pflag.CommandLine)

	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()
	verflag.PrintAndExitIfRequested()

	if err := app.Run(ccm); err != nil {
		klog.Errorf("Run CCM error: %s", err.Error())
		os.Exit(1)
	}
}
