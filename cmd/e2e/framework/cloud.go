package framework

import (
	"fmt"
	"k8s.io/cloud-provider"
	cloud "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
)

func NewAlibabaCloudOrDie(configpath string) *cloud.Cloud {
	iprovider, err := cloudprovider.InitCloudProvider("alicloud", configpath)
	if err != nil {
		panic(fmt.Sprintf("FrameWorkE2E could not be initialized: %v", err))
	}
	provider, ok := iprovider.(*cloud.Cloud)
	if !ok {
		panic("not alibaba cloud provider")
	}
	return provider
}
