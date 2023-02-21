package common

import (
	"context"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/klog/v2"
	"time"
)

// Service
// createDefaultService
type Service struct {
}

func (*Service) CreateDefaultService(f *framework.Framework) {
	svc := f.Client.KubeClient.DefaultService()
	svc.Spec.Type = v1.ServiceTypeNodePort
	_, err := f.Client.KubeClient.CoreV1().Services(svc.Namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
	if err != nil {
		_, err = f.Client.KubeClient.CreateService(svc)
	}
	gomega.Expect(err).To(gomega.BeNil())
	klog.Infof("node port service created :%s/%s", svc.Namespace, svc.Name)
}

func (*Service) WaitCreateDefaultServiceWithSvcName(f *framework.Framework, svcName string) {
	svc := f.Client.KubeClient.DefaultService()
	svc.Spec.Type = v1.ServiceTypeNodePort
	svc.Name = svcName
	_, err := f.Client.KubeClient.CoreV1().Services(svc.Namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
	if err != nil {
		_, err = f.Client.KubeClient.CreateService(svc)
	}
	timeout := 30 * time.Second
	err = wait.Poll(7*time.Second, timeout, func() (done bool, err error) {
		svc, err = f.Client.KubeClient.CoreV1().Services(svc.Namespace).Get(context.TODO(), svcName, metav1.GetOptions{})
		if err != nil {
			klog.Infof("wait for service ready: %s", err.Error())
			return false, nil
		}
		if svc.Spec.ClusterIP != "" {
			return true, nil
		}
		klog.Infof("wait for service not ready: %s", svcName)
		return false, nil
	})
	gomega.Expect(err).To(gomega.BeNil())
	klog.Infof("node port service created :%s/%s", svc.Namespace, svc.Name)
}
