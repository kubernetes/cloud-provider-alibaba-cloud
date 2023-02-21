package alb

import (
	"context"
	"fmt"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/alb/common"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/klog/v2"
)

func RunIngressTestCases(f *framework.Framework) {

	ginkgo.BeforeEach(func() {
		svc := f.Client.KubeClient.DefaultService()
		svc.Spec.Type = v1.ServiceTypeNodePort
		_, err := f.Client.KubeClient.CoreV1().Services(svc.Namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
		if err != nil {
			_, err = f.Client.KubeClient.CreateService(svc)
		}
		gomega.Expect(err).To(gomega.BeNil())
		klog.Infof("node port service created :%s/%s", svc.Namespace, svc.Name)
	})

	ginkgo.AfterEach(func() {
		ing := f.Client.KubeClient.DefaultIngress()
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
	})

	ginkgo.Describe("alb-ingress-controller: ingress", func() {

		ginkgo.Context("ingress create", func() {
			ginkgo.It("[alb][p0] ingress with NodePort Service", func() {
				ing := defaultIngress(f)
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p0] ingress delete not remove listener", func() {
				ing := defaultIngress(f)
				waitCreateIngress(f, ing)
				waitDeleteIngress(f, ing)
				crd := getAlbconfigCrd(f)
				listeners := getAlbconfigListeners(crd)
				gomega.Expect(len(listeners) > 0).To(gomega.BeTrue())
			})
			ginkgo.It("[alb][p0] ingress with rewrite-target annotations", func() {
				ing := defaultIngress(f)
				anno := map[string]string{}
				anno["alb.ingress.kubernetes.io/rewrite-target"] = "/path/${2}"
				ing.Annotations = anno
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p0] ingress with TCP healthcheck Service", func() {
				ing := defaultIngress(f)
				anno := map[string]string{}
				anno["alb.ingress.kubernetes.io/healthcheck-enabled"] = "true"
				anno["alb.ingress.kubernetes.io/healthcheck-protocol"] = "TCP"
				anno["alb.ingress.kubernetes.io/healthcheck-timeout-seconds"] = "5"
				anno["alb.ingress.kubernetes.io/healthcheck-interval-seconds"] = "2"
				anno["alb.ingress.kubernetes.io/healthy-threshold-count"] = "3"
				anno["alb.ingress.kubernetes.io/unhealthy-threshold-count"] = "3"
				ing.Annotations = anno
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p1] ingress with TLS secret", func() {
				ing := defaultTLSIngress(f)
				ing.Annotations = map[string]string{
					"alb.ingress.kubernetes.io/listen-ports": "[{\"HTTPS\": 3443}]",
				}
				secret := defaultSecret(f)
				waitCreateSecret(f, secret)
				ing.Spec.TLS[0].SecretName = secret.Name
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p1] ingress with TLS Certificate", func() {
				ing := defaultTLSIngress(f)
				ing.Annotations = map[string]string{
					"alb.ingress.kubernetes.io/listen-ports": "[{\"HTTPS\": 3443}]",
				}
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p1] ingress with backend protocol", func() {
				ing := defaultTLSIngress(f)
				ing.Annotations = map[string]string{
					"alb.ingress.kubernetes.io/backend-protocol": "GRPC",
				}
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p0] ingress with backend scheduler", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					"alb.ingress.kubernetes.io/backend-scheduler": "sch",
				}
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p0] ingress with action-name", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					"alb.ingress.kubernetes.io/actions.response-503":  "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":      "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header": "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p1] ingress with traffic-limit only", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					"alb.ingress.kubernetes.io/traffic-limit-qps": "50",
				}
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p0] ingress with rewrite-target and traffic-limit", func() {
				ing := defaultIngress(f)
				anno := map[string]string{}
				anno["alb.ingress.kubernetes.io/rewrite-target"] = "/path/${2}"
				anno["alb.ingress.kubernetes.io/traffic-limit-qps"] = "50"
				ing.Annotations = anno
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p0] ingress with cors only", func() {
				ing := defaultIngress(f)
				anno := map[string]string{}
				ing.Annotations = anno
				waitCreateIngress(f, ing)
			})
			ginkgo.It("[alb][p0] ingress with rewrite-target and cors", func() {
				ing := defaultIngress(f)
				anno := map[string]string{}
				anno["alb.ingress.kubernetes.io/rewrite-target"] = "/path/${2}"
				anno["alb.ingress.kubernetes.io/cors-enable"] = "true"
				ing.Annotations = anno
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				waitCreateIngress(f, ing)
			})
		})
	})

	ginkgo.Describe("alb-ingress-controller: albconfig", func() {

		ginkgo.Context("albconfig update", func() {
			ginkgo.It("[alb][p1] albconfig https listens securityPolicyId", func() {
				ing := defaultTLSIngress(f)
				ing.Annotations = map[string]string{
					"alb.ingress.kubernetes.io/listen-ports": "[{\"HTTPS\": 3443}]",
				}
				waitCreateIngress(f, ing)
				crd := getAlbconfigCrd(f)
				listeners := getAlbconfigListeners(crd)
				lbId := getAlbconfigLbId(crd)

				for _, listener := range listeners {
					ls := listener.(map[string]interface{})
					port, _, _ := unstructured.NestedInt64(ls, "port")
					if port == 3443 {
						unstructured.SetNestedField(ls, "tls_cipher_policy_1_1", "securityPolicyId")
					}
				}
				err := unstructured.SetNestedSlice(crd.Object, listeners, "spec", "listeners")
				gomega.Expect(err).To(gomega.BeNil())
				updateAlbConfig(f, crd)

				albSdkProvider := f.Client.CloudClient.ALBProvider
				policySet := false
				// 调谐成功后云上资源对应的监听有SecurityPolicyId设置即可
				wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
					cloudListeners, err := albSdkProvider.ListALBListeners(context.TODO(), lbId)
					if err != nil {
						klog.Infof("list alb listerners: %s:%s", lbId, err.Error())
						return false, nil
					}
					for _, ls := range cloudListeners {
						if ls.ListenerPort == 3443 && ls.SecurityPolicyId == "tls_cipher_policy_1_1" {
							policySet = true
							return true, nil
						}
					}
					klog.Infof("wait for alb listeners security policy id: %s", lbId)
					return false, nil
				})
				gomega.Expect(policySet).To(gomega.BeTrue())
			})
			ginkgo.It("[alb][p0] albconfig listener aclconfig", func() {
				ing := defaultIngress(f)
				waitCreateIngress(f, ing)
				crd := getAlbconfigCrd(f)
				listeners := getAlbconfigListeners(crd)
				lbId := getAlbconfigLbId(crd)

				for _, listener := range listeners {
					ls := listener.(map[string]interface{})
					aclConfigMap := map[string]interface{}{
						"aclType": "White",
						"aclName": "e2e-test-acl",
					}
					unstructured.SetNestedMap(ls, aclConfigMap, "aclConfig")
					unstructured.SetNestedStringSlice(ls, []string{"127.0.0.1/32"}, "aclConfig", "aclEntries")
				}
				err := unstructured.SetNestedSlice(crd.Object, listeners, "spec", "listeners")
				gomega.Expect(err).To(gomega.BeNil())
				updateAlbConfig(f, crd)

				albSdkProvider := f.Client.CloudClient.ALBProvider
				cloudAclType := ""
				// 调谐成功后云上资源对应的监听有SecurityPolicyId设置即可
				wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
					cloudListeners, err := albSdkProvider.ListALBListeners(context.TODO(), lbId)
					if err != nil {
						klog.Infof("list alb listerners: %s:%s", lbId, err.Error())
						return false, nil
					}
					for _, ls := range cloudListeners {
						cloudLs, err := albSdkProvider.GetALBListenerAttribute(context.TODO(), ls.ListenerId)
						if err != nil {
							klog.Infof("get listener attribute failed: %s, reason: %s", lbId, err.Error())
							return false, nil
						}
						if cloudLs.AclConfig.AclType != "" {
							cloudAclType = cloudLs.AclConfig.AclType
							return true, nil
						}
					}
					klog.Infof("wait for alb listeners : %s", lbId)
					return false, nil
				})
				gomega.Expect(cloudAclType).To(gomega.Equal("White"))
			})
		})
	})
}

func defaultIngress(f *framework.Framework) *networkingv1.Ingress {
	return f.Client.KubeClient.DefaultIngress()
}
func defaultIngressWithSvcName(f *framework.Framework, svcName string) *networkingv1.Ingress {
	return f.Client.KubeClient.DefaultIngressWithSvcName(svcName)
}

func defaultTLSIngress(f *framework.Framework) *networkingv1.Ingress {
	ingTmp := f.Client.KubeClient.DefaultIngress()
	return f.Client.KubeClient.IngressWithTLS(ingTmp, []string{ingTmp.Spec.Rules[0].Host})
}
func defaultRule(path, serviceName string) networkingv1.IngressRule {
	exact := networkingv1.PathTypeExact
	return networkingv1.IngressRule{
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     path,
						PathType: &exact,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: serviceName,
								Port: networkingv1.ServiceBackendPort{
									Name: "use-annotation",
								},
							},
						},
					},
				},
			},
		},
	}
}

func defaultSecret(f *framework.Framework) *v1.Secret {
	return f.Client.KubeClient.DefaultSecret()
}

func waitCreateSecret(f *framework.Framework, secret *v1.Secret) {
	_, err := f.Client.KubeClient.CoreV1().Secrets(secret.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	klog.Info("apply  to apiserver: secret=", secret, ", result=", err)
	gomega.Expect(err).To(gomega.BeNil())
}

func waitDeleteIngress(f *framework.Framework, ing *networkingv1.Ingress) {
	err := f.Client.KubeClient.NetworkingV1().Ingresses(ing.Namespace).Delete(context.TODO(), ing.Name, metav1.DeleteOptions{})
	wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, waitErr error) {
		ingT, err := f.Client.KubeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, metav1.GetOptions{})
		if err == nil {
			klog.Infof("ingress has not been deleted: %s/%s", ingT.Namespace, ingT.Name)
			return false, nil
		}
		klog.Infof("ingress had been deleted: %s/%s", ing.Namespace, ing.Name)
		return true, nil
	})
	gomega.Expect(err).To(gomega.BeNil())
	klog.Infof("ingress delete :%s/%s", ing.Namespace, ing.Name)
}

func waitCreateIngress(f *framework.Framework, ing *networkingv1.Ingress) {
	_, err := f.Client.KubeClient.CreateIngress(ing)
	klog.Info("apply ingress to apiserver: ingress=", ing, ", result=", err)
	gomega.Expect(err).To(gomega.BeNil())

	wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		ingT, err := f.Client.KubeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, metav1.GetOptions{})
		if err != nil {
			for _, rule := range ing.Spec.Rules {
				for _, path := range rule.HTTP.Paths {
					_, conditionExist := ing.Annotations[fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, path.Backend.Service.Name)]
					_, canaryExist := ing.Annotations[fmt.Sprintf(annotations.AlbCanary, path.Backend.Service.Name)]
					if conditionExist && canaryExist {
						klog.Errorf(" can't exist Canary and customize condition at the same time")
						return true, nil
					}
				}
			}
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
	klog.Info("ingress reconcile by alb-ingress-controller, ingress=", ing, ", result=", err)
	gomega.Expect(err).To(gomega.BeNil())
	lb := ingN.Status.LoadBalancer
	dnsName := ""
	if len(lb.Ingress) > 0 && lb.Ingress[0].Hostname != "" {
		dnsName = lb.Ingress[0].Hostname
	}
	klog.Info("Expect ingress.Status.loadBalance.Ingress.Hostname != nil")
	gomega.Expect(dnsName).NotTo(gomega.BeEmpty(), common.PrintEventsWhenError(f))
}

func getAlbconfigCrd(f *framework.Framework) *unstructured.Unstructured {
	albResource := schema.GroupVersionResource{Group: "alibabacloud.com", Version: "v1", Resource: "albconfigs"}
	crd, err := f.Client.DynamicClient.Resource(albResource).Get(context.TODO(), "default", metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), common.PrintEventsWhenError(f))
	return crd
}

func getAlbconfigListeners(albconfig *unstructured.Unstructured) []interface{} {
	listeners, found, err := unstructured.NestedSlice(albconfig.UnstructuredContent(), "spec", "listeners")
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(found).To(gomega.BeTrue())
	return listeners
}

func getAlbconfigLbId(albconfig *unstructured.Unstructured) string {
	lbId, found, err := unstructured.NestedString(albconfig.UnstructuredContent(), "status", "loadBalancer", "id")
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(found).To(gomega.BeTrue())
	return lbId
}

func updateAlbConfig(f *framework.Framework, albconfig *unstructured.Unstructured) {
	albResource := schema.GroupVersionResource{Group: "alibabacloud.com", Version: "v1", Resource: "albconfigs"}
	crd, err := f.Client.DynamicClient.Resource(albResource).Update(context.TODO(), albconfig, metav1.UpdateOptions{})
	gomega.Expect(err).To(gomega.BeNil(), common.PrintEventsWhenError(f))
	klog.Infof("update albconfig with: %s", crd)
}
