package common

import (
	"context"
	"github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
)

type IngressClass struct {
}

func (*IngressClass) CreateIngressClass(f *framework.Framework) {
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
}
