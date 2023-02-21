package common

import (
	"context"
	"time"

	"github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/klog/v2"
)

// Ingress func meaning
// waitCreateIngress--等待指定的ingress创建成功
// deleteIngress--删除指定的ingress
// defaultIngress--返回默认的ingress
// defaultTLSIngress--返回默认的TLS类型的ingress
type Ingress struct {
}

func (*Ingress) DefaultIngress(f *framework.Framework) *networkingv1.Ingress {
	return f.Client.KubeClient.DefaultIngress()
}

func (*Ingress) defaultTLSIngress(f *framework.Framework) *networkingv1.Ingress {
	ingTmp := f.Client.KubeClient.DefaultIngress()
	return f.Client.KubeClient.IngressWithTLS(ingTmp, []string{ingTmp.Spec.Rules[0].Host})
}

func (*Ingress) WaitCreateIngress(f *framework.Framework, ing *networkingv1.Ingress, expectSuccess bool) {
	_, err := f.Client.KubeClient.CreateIngress(ing)
	klog.Info("apply ingress to apiserver: ingress=", ing, ", result=", err)
	gomega.Expect(err).To(gomega.BeNil())
	var timeout time.Duration
	if expectSuccess {
		timeout = 2 * time.Minute
	} else {
		timeout = 30 * time.Second
	}
	err = wait.Poll(5*time.Second, timeout, func() (done bool, err error) {
		ingT, err := f.Client.KubeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, metav1.GetOptions{})
		if err != nil {
			klog.Infof("wait for ingress alb status ready: %s", err.Error())
			return false, nil
		}
		lb := ingT.Status.LoadBalancer

		if len(lb.Ingress) > 0 && lb.Ingress[0].Hostname != "" {
			return true, nil
		}
		klog.Infof("wait for ingress alb status not ready: %s", ingT.Name)
		return false, nil
	})
	ingN, err := f.Client.KubeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, metav1.GetOptions{})
	if expectSuccess {
		klog.Info("Expect ingress can create successful, ingN=", ing, ", result=", err)
		gomega.Expect(err).To(gomega.BeNil())
		lb := ingN.Status.LoadBalancer
		dnsName := ""
		if len(lb.Ingress) > 0 && lb.Ingress[0].Hostname != "" {
			dnsName = lb.Ingress[0].Hostname
		}
		klog.Info("Expect ingress.Status.loadBalance.Ingress.Hostname != nil")
		gomega.Expect(dnsName).NotTo(gomega.BeEmpty(), PrintEventsWhenError(f))
	} else {
		klog.Info("Expect ingress can not create successful ingN=", ingN, ", result=", err)
		gomega.Expect(err).To(gomega.BeNil())
		lb := ingN.Status.LoadBalancer
		dnsName := ""
		if len(lb.Ingress) > 0 && lb.Ingress[0].Hostname != "" {
			dnsName = lb.Ingress[0].Hostname
		}
		klog.Info("Expect ingress.Status.loadBalance.Ingress.Hostname == nil")
		gomega.Expect(dnsName).To(gomega.BeEmpty(), PrintEventsWhenError(f))
	}
}

func (*Ingress) DeleteIngress(f *framework.Framework, ing *networkingv1.Ingress) {
	ingN, err := f.Client.KubeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, metav1.GetOptions{})
	if err == nil {
		err = f.Client.KubeClient.NetworkingV1().Ingresses(ingN.Namespace).Delete(context.TODO(), ingN.Name, metav1.DeleteOptions{})
		var ingT *networkingv1.Ingress
		wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, waitErr error) {
			ingT, err = f.Client.KubeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, metav1.GetOptions{})
			if err == nil {
				klog.Infof("ingress has not been deleted: %s/%s", ingT.Namespace, ingT.Name)
				return false, nil
			}
			klog.Infof("ingress had been deleted: %s/%s", ing.Namespace, ing.Name)
			return true, nil
		})
	}
	gomega.Expect(err).NotTo(gomega.BeNil())
	klog.Infof("ingress delete :%s/%s", ingN.Namespace, ingN.Name)
}
