package common

import (
	"context"
	"fmt"
	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type ALB struct {
}

func getAlbResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "alibabacloud.com", Version: "v1", Resource: "albconfigs"}
}

func (*ALB) GetCloudALBByName(name string, f *framework.Framework) ([]albsdk.LoadBalancer, error) {
	clulsterId := options.TestConfig.ClusterId
	albConfigName := "kube-system/" + name
	tags := make(map[string]string)
	tags["ack.aliyun.com"] = clulsterId
	tags["ingress.k8s.alibaba/albconfig"] = albConfigName
	tags["ingress.k8s.alibaba/resource"] = "ApplicationLoadBalancer"
	if albsWithTags, err := f.Client.CloudClient.ALBProvider.ListAlbLoadBalancersByTag(context.TODO(), tags); err == nil {
		return albsWithTags, err
	} else {
		return nil, err
	}
}

func (*ALB) DeleteCloudALBById(Id string, f *framework.Framework) error {
	if err := f.Client.CloudClient.ALBProvider.DeleteALB(context.TODO(), Id); err != nil {
		return err
	} else {
		return nil
	}
}

func (*ALB) CleanIngressClassByName(f *framework.Framework, name string) error {
	if err := f.Client.KubeClient.NetworkingV1().IngressClasses().Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func (*ALB) CleanAlbConfigByName(f *framework.Framework, name string) error {
	if _, err := f.Client.DynamicClient.Resource(getAlbResource()).Get(context.TODO(), name, metav1.GetOptions{}); errors.IsNotFound(err) {
		return nil
	}
	if err := f.Client.DynamicClient.Resource(getAlbResource()).Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
		return err
	} else {
		wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
			_, err = f.Client.DynamicClient.Resource(getAlbResource()).Get(context.TODO(), name, metav1.GetOptions{})
			if err == nil {
				klog.Infof("wait albconfig delete")
				return false, nil
			}
			klog.Infof("exist albconfig deleted")
			return true, nil
		})
	}
	return nil
}

func (*ALB) GetAlbConfigByName(f *framework.Framework, name string, namespace string) (*v1.AlbConfig, error) {
	albconfig := &v1.AlbConfig{}
	if err := f.Client.RuntimeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, albconfig); err != nil {
		return nil, err
	} else {
		return albconfig, nil
	}
}

func (*ALB) GetDefaultAlbConfig() *v1.AlbConfig {
	forceOverride := false
	deletionProtectionEnabled := true

	albConfig := &v1.AlbConfig{}
	albConfig.Name = "default"
	albConfig.Spec.LoadBalancer = &v1.LoadBalancerSpec{
		AddressAllocatedMode: "Dynamic",
		ZoneMappings: []v1.ZoneMapping{
			{VSwitchId: options.TestConfig.VSwitchID},
			{VSwitchId: options.TestConfig.VSwitchID2},
		},
		AddressType:               "Intranet",
		Edition:                   "Standard",
		DeletionProtectionEnabled: &deletionProtectionEnabled,
		Name:                      "alb-config-e2etest",
		ForceOverride:             &forceOverride,
		BillingConfig: v1.BillingConfig{
			PayType: "PostPay",
		},
	}
	return albConfig
}

func (*ALB) CreateAlbConfig(f *framework.Framework, albconfig *v1.AlbConfig) error {
	if err := f.Client.RuntimeClient.Create(context.TODO(), albconfig, &client.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func (a *ALB) WaitCreateALBConfig(f *framework.Framework, albconfig *v1.AlbConfig, expectSuccess bool) {
	err := a.CreateAlbConfig(f, albconfig)
	klog.Info("apply albconfig to apiserver: ingress=", albconfig, ", result=", err)
	gomega.Expect(err).To(gomega.BeNil())
	var timeout time.Duration
	if expectSuccess {
		timeout = 2 * time.Minute
	} else {
		timeout = 30 * time.Second
	}
	var lbId string
	err = wait.Poll(5*time.Second, timeout, func() (done bool, err error) {
		alb, err := a.GetAlbConfigByName(f, "default", "")
		gomega.Expect(err).To(gomega.BeNil())
		if err != nil {
			klog.Infof("albconfig get lbId failed: %s", err.Error())
			return false, nil
		}
		lbId = alb.Status.LoadBalancer.Id
		klog.Infof("alb config get lbId: %s", lbId)
		if lbId == "" {
			return false, nil
		}
		return true, nil
	})
	if expectSuccess {
		gomega.Expect(lbId).NotTo(gomega.BeEmpty())
	} else {
		gomega.Expect(lbId).To(gomega.BeEmpty())
	}
}

func (*ALB) CleanAllIngressClass(f *framework.Framework) error {
	if ingressClass, err := f.Client.KubeClient.NetworkingV1().IngressClasses().List(context.TODO(), metav1.ListOptions{}); err == nil {
		for _, ingressClassItem := range ingressClass.Items {
			err = f.Client.KubeClient.NetworkingV1().IngressClasses().Delete(context.TODO(), ingressClassItem.Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}
	return nil
}

func (*ALB) CleanAllALbConfig(f *framework.Framework) error {
	if albs, err := f.Client.DynamicClient.Resource(getAlbResource()).List(context.TODO(), metav1.ListOptions{}); err == nil {
		for _, albConfig := range albs.Items {
			err = f.Client.DynamicClient.Resource(getAlbResource()).Delete(context.TODO(), albConfig.GetName(), metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}
	return nil
}

func (*ALB) GetALBLoadBalancersWithTags(f *framework.Framework) ([]alb.AlbLoadBalancerWithTags, error) {
	var lbId string
	Tags := make(map[string]string)
	ALBsWithTags, err := f.Client.CloudClient.ALBProvider.ListALBsWithTags(context.TODO(), Tags)
	if err != nil {
		return ALBsWithTags, fmt.Errorf("get ALBLoadBalancersByTags error: %s", err.Error())
	}

	if len(ALBsWithTags) > 1 {
		return ALBsWithTags, fmt.Errorf("invalid sdk loadBalancers: %v, at most one loadBalancer can find ,must delete manually", ALBsWithTags)
	}

	for _, ALBWithTags := range ALBsWithTags {
		if ALBWithTags.LoadBalancerId != lbId {
			return ALBsWithTags, fmt.Errorf("invalid sdk loadBalancerId: %v ", ALBWithTags.LoadBalancerId)
		}
	}
	return ALBsWithTags, nil
}

func (*ALB) CheckVPCIdOfALBLoadBalancersWithTags(f *framework.Framework, lbs []alb.AlbLoadBalancerWithTags) error {
	VpcId := ""
	for _, LoadBalancerWithTags := range lbs {
		if LoadBalancerWithTags.VpcId != VpcId {
			return fmt.Errorf("CheckVPCIdOfALBLoadBalancersWithTags failed")
		}
	}
	return nil
}
