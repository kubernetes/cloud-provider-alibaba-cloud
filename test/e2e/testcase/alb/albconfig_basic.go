package alb

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
	"k8s.io/klog/v2"
)

func RunAlbConfigTestCases(f *framework.Framework) {

	ginkgo.Describe("alb-ingress-controller testcase: albconfig", func() {
		ginkgo.Context("albconfig create", func() {
			ginkgo.It("[alb][p0] albconfig base create", func() {
				beforeItem(f)
				albconfig := getAlbConfig("Standard")
				waitCreateAlb(f, albconfig)
			})
			ginkgo.It("[alb][p1] albconfig waf create", func() {
				beforeItem(f)
				albconfig := getAlbConfig("StandardWithWaf")
				waitCreateAlb(f, albconfig)
			})
		})
	})

}

func beforeItem(f *framework.Framework) {
	// 如果存在ingressClass和albconfig先删除，创建e2e自己的
	_, err := f.Client.KubeClient.NetworkingV1().IngressClasses().Get(context.TODO(), "alb", metav1.GetOptions{})
	if err != nil {
		apiGroup := "alibabacloud.com"
		ingressClass := networkingv1.IngressClass{}
		ingressClass.Name = "alb"
		ingressClass.Spec.Controller = "ingress.k8s.alibabacloud/alb"
		ingressClass.Spec.Parameters = &networkingv1.IngressClassParametersReference{
			APIGroup: &apiGroup,
			Kind:     "AlbConfig",
			Name:     "default",
		}
		_, ingressClassCreateErr := f.Client.KubeClient.NetworkingV1().IngressClasses().Create(context.TODO(), &ingressClass, metav1.CreateOptions{})
		gomega.Expect(ingressClassCreateErr).To(gomega.BeNil())
	}
	albResource := schema.GroupVersionResource{Group: "alibabacloud.com", Version: "v1", Resource: "albconfigs"}
	_, err = f.Client.DynamicClient.Resource(albResource).Get(context.TODO(), "default", metav1.GetOptions{})
	if err == nil {
		f.Client.DynamicClient.Resource(albResource).Delete(context.TODO(), "default", metav1.DeleteOptions{})
		wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
			_, err = f.Client.DynamicClient.Resource(albResource).Get(context.TODO(), "default", metav1.GetOptions{})
			if err == nil {
				klog.Infof("wait albconfig delete")
				return false, nil
			}
			klog.Infof("exist albconfig deleted")
			return true, nil
		})
	}
}

func getAlbResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "alibabacloud.com", Version: "v1", Resource: "albconfigs"}
}

func getAlbConfig(edition string) *unstructured.Unstructured {
	albconfig := &unstructured.Unstructured{}
	albconfig.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "alibabacloud.com/v1",
		"kind":       "AlbConfig",
		"metadata": map[string]interface{}{
			"name": "default",
		},
		"spec": map[string]interface{}{
			"config": map[string]interface{}{
				"name": "alb-config-e2etest",
				"accessLogConfig": map[string]interface{}{
					"logProject": "",
					"logStore":   "",
				},
				"addressAllocatedMode": "Dynamic",
				"billingConfig": map[string]interface{}{
					"internetBandwidth":  0,
					"internetChargeType": "",
					"payType":            "PostPay",
				},
				"deletionProtectionEnabled": true,
				"edition":                   edition,
				"forceOverride":             false,
				"addressType":               "Intranet",
				"zoneMappings": []map[string]interface{}{
					{"vSwitchId": options.TestConfig.VSwitchID},
					{"vSwitchId": options.TestConfig.VSwitchID2},
				},
			},
		},
	})
	return albconfig
}

func waitCreateAlb(f *framework.Framework, albconfig *unstructured.Unstructured) {
	albResource := getAlbResource()
	_, err := f.Client.DynamicClient.Resource(albResource).Create(context.TODO(), albconfig, metav1.CreateOptions{})
	gomega.Expect(err).To(gomega.BeNil())
	var lbId string
	var found bool
	// 调谐成功后云上资源对应的监听有LoadBalancerId设置即可
	wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		crd, err := f.Client.DynamicClient.Resource(albResource).Get(context.TODO(), "default", metav1.GetOptions{})
		gomega.Expect(err).To(gomega.BeNil())
		lbId, found, err = unstructured.NestedString(crd.UnstructuredContent(), "status", "loadBalancer", "id")
		if err != nil {
			klog.Infof("albconfig get lbId failed: %s", err.Error())
			return false, nil
		}
		if !found {
			klog.Infof("albconfig loadBalancer Id not exist")
			return false, nil
		}
		klog.Infof("alb config get lbId: %s", lbId)
		return true, nil
	})
	gomega.Expect(lbId).NotTo(gomega.BeEmpty())
}

func CleanAlbconfigTestCases(f *framework.Framework) {
	// 如果存在ingressClass和albconfig删除它

	ginkgo.Describe("alb-ingress-controller: albconfig", func() {
		ginkgo.Context("albconfig delete", func() {
			ginkgo.It("[alb][p0] alb-ingress resource after check", func() {
				inc, err := f.Client.KubeClient.NetworkingV1().IngressClasses().Get(context.TODO(), "alb", metav1.GetOptions{})
				klog.Infof("deleteing ingressclasses ", inc)
				if err == nil {
					f.Client.KubeClient.NetworkingV1().IngressClasses().Delete(context.TODO(), inc.Name, metav1.DeleteOptions{})
				}
				albResource := schema.GroupVersionResource{Group: "alibabacloud.com", Version: "v1", Resource: "albconfigs"}
				alb, err := f.Client.DynamicClient.Resource(albResource).Get(context.TODO(), "default", metav1.GetOptions{})
				klog.Infof("deleteing albconfig ", alb)
				if err == nil {
					f.Client.DynamicClient.Resource(albResource).Delete(context.TODO(), "default", metav1.DeleteOptions{})
					wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
						_, err = f.Client.DynamicClient.Resource(albResource).Get(context.TODO(), "default", metav1.GetOptions{})
						if err == nil {
							klog.Infof("wait albconfig delete")
							return false, nil
						}
						klog.Infof("exist albconfig deleted")
						return true, nil
					})
				}
				_, err = f.Client.DynamicClient.Resource(albResource).Get(context.TODO(), "default", metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})

}
