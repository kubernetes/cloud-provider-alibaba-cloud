package clbv1

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
)

func RunListenerTestCases(f *framework.Framework) {
	ginkgo.Describe("clb service controller: listener", func() {

		ginkgo.AfterEach(func() {
			ginkgo.By("delete service")
			err := f.AfterEach()
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.Context("scheduler", func() {
			ginkgo.It("scheduler: rr", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.Scheduler): "rr",
				})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("scheduler: wrr", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.Scheduler): "wrr",
				})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("scheduler: wlc", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.Scheduler): "wlc",
				})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("scheduler: rr -> wrr", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.Scheduler): "rr",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.ObjectMeta.Annotations[annotation.Annotation(annotation.Scheduler)] = "wrr"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		if options.TestConfig.AclID != "" {
			ginkgo.Context("acl", func() {
				ginkgo.It("acl-type: white -> black", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AclID):     options.TestConfig.AclID,
						annotation.Annotation(annotation.AclStatus): string(model.OnFlag),
						annotation.Annotation(annotation.AclType):   "white",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.AclType)] = "black"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())

				})
				ginkgo.It("acl-status: on -> off", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AclID):     options.TestConfig.AclID,
						annotation.Annotation(annotation.AclStatus): string(model.OnFlag),
						annotation.Annotation(annotation.AclType):   "white",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.AclStatus)] = string(model.OffFlag)
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				if options.TestConfig.AclID2 != "" {
					ginkgo.It("update acl-id", func() {
						oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							annotation.Annotation(annotation.AclID):     options.TestConfig.AclID,
							annotation.Annotation(annotation.AclStatus): string(model.OnFlag),
							annotation.Annotation(annotation.AclType):   "white",
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldsvc)
						gomega.Expect(err).To(gomega.BeNil())

						newsvc := oldsvc.DeepCopy()
						newsvc.Annotations[annotation.Annotation(annotation.AclID)] = options.TestConfig.AclID2
						newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
						gomega.Expect(err).To(gomega.BeNil())

						err = f.ExpectLoadBalancerEqual(newsvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
				ginkgo.It("acl-id: not exist", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AclID):     "acl-xxxxx",
						annotation.Annotation(annotation.AclStatus): string(model.OnFlag),
						annotation.Annotation(annotation.AclType):   "white",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("acl-id: exist -> not exist", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AclID):     options.TestConfig.AclID,
						annotation.Annotation(annotation.AclStatus): string(model.OnFlag),
						annotation.Annotation(annotation.AclType):   "white",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.AclID)] = "acl-xxxxx"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			})
		}

		ginkgo.Context("XForwardedForProto", func() {
			ginkgo.It("xforwardedfor-proto; http", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):       "http:80",
					annotation.Annotation(annotation.XForwardedForProto): string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.XForwardedForProto)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("xforwardedfor-proto; https", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort):       "http:443",
						annotation.Annotation(annotation.CertID):             options.TestConfig.CertID,
						annotation.Annotation(annotation.XForwardedForProto): string(model.OnFlag),
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.XForwardedForProto)] = string(model.OffFlag)
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("X-Forwarded-SLBPort", func() {
			ginkgo.It("X-Forwarded-SLBPort; http", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):         "http:80",
					annotation.Annotation(annotation.XForwardedForSLBPort): string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.XForwardedForSLBPort)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("X-Forwarded-SLBPort; https", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort):         "http:443",
						annotation.Annotation(annotation.CertID):               options.TestConfig.CertID,
						annotation.Annotation(annotation.XForwardedForSLBPort): string(model.OnFlag),
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.XForwardedForSLBPort)] = string(model.OffFlag)
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("X-Forwarded-Client-srcport", func() {
			ginkgo.It("X-Forwarded-Client-srcport; http", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):               "http:80",
					annotation.Annotation(annotation.XForwardedForClientSrcPort): string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.XForwardedForClientSrcPort)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("X-Forwarded-Client-srcport; https", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort):               "http:443",
						annotation.Annotation(annotation.CertID):                     options.TestConfig.CertID,
						annotation.Annotation(annotation.XForwardedForClientSrcPort): string(model.OnFlag),
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.XForwardedForClientSrcPort)] = string(model.OffFlag)
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("idle timeout", func() {
			ginkgo.It("idle-timeout: 10 -> 40", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort): "http:80",
					annotation.Annotation(annotation.IdleTimeout):  "10",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.IdleTimeout)] = "40"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("idle-timeout: 10 -> 40", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.IdleTimeout):  "10",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					// modify
					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.IdleTimeout)] = "40"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("request timeout", func() {
			ginkgo.It("request-timeout: 1 -> 40", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):   "https:443",
					annotation.Annotation(annotation.CertID):         options.TestConfig.CertID,
					annotation.Annotation(annotation.RequestTimeout): "1",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.RequestTimeout)] = "40"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("request-timeout: 100 -> 180", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort):   "https:443",
						annotation.Annotation(annotation.CertID):         options.TestConfig.CertID,
						annotation.Annotation(annotation.RequestTimeout): "100",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					// modify
					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.RequestTimeout)] = "180"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}

		})

		if options.TestConfig.CertID != "" {
			ginkgo.Context("http2-enabled", func() {
				ginkgo.It("http2-enabled: on -> off", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.EnableHttp2):  string(model.OnFlag),
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					// modify
					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.EnableHttp2)] = string(model.OffFlag)
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})
		}

		ginkgo.Context("health check", func() {
			ginkgo.It("health-check: tcp", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.HealthCheckType):           model.TCP,
					annotation.Annotation(annotation.HealthCheckConnectTimeout): "8",
					annotation.Annotation(annotation.HealthyThreshold):          "5",
					annotation.Annotation(annotation.UnhealthyThreshold):        "5",
					annotation.Annotation(annotation.HealthCheckInterval):       "3",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckConnectTimeout)] = "10"
				newsvc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "2"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("health-check: http", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):        "http:80",
					annotation.Annotation(annotation.HealthCheckType):     model.HTTP,
					annotation.Annotation(annotation.HealthCheckFlag):     string(model.OnFlag),
					annotation.Annotation(annotation.HealthCheckTimeout):  "8",
					annotation.Annotation(annotation.HealthyThreshold):    "5",
					annotation.Annotation(annotation.UnhealthyThreshold):  "5",
					annotation.Annotation(annotation.HealthCheckInterval): "3",
					annotation.Annotation(annotation.HealthCheckMethod):   "head",
					annotation.Annotation(annotation.HealthCheckHTTPCode): "http_3xx",
					annotation.Annotation(annotation.HealthCheckDomain):   "192.168.0.3",
					annotation.Annotation(annotation.HealthCheckURI):      "/test/index.html",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckTimeout)] = "10"
				newsvc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "2"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckMethod)] = "get"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckHTTPCode)] = "http_2xx"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckDomain)] = "192.168.0.2"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckURI)] = "/test/index1.html"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("health-check-swith off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.HealthCheckType):           model.TCP,
					annotation.Annotation(annotation.HealthCheckConnectTimeout): "8",
					annotation.Annotation(annotation.HealthyThreshold):          "5",
					annotation.Annotation(annotation.UnhealthyThreshold):        "5",
					annotation.Annotation(annotation.HealthCheckInterval):       "3",
					annotation.Annotation(annotation.HealthCheckSwitch):         "off",
					annotation.Annotation(annotation.ProtocolPort):              "tcp:80,udp:443",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("health-check-swith: on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.HealthCheckType):           model.TCP,
					annotation.Annotation(annotation.HealthCheckConnectTimeout): "8",
					annotation.Annotation(annotation.HealthyThreshold):          "5",
					annotation.Annotation(annotation.UnhealthyThreshold):        "5",
					annotation.Annotation(annotation.HealthCheckInterval):       "3",
					annotation.Annotation(annotation.HealthCheckSwitch):         "on",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// close health check for tcp & udp
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckSwitch)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("health-check-flag: on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):        "http:80",
					annotation.Annotation(annotation.HealthCheckType):     model.HTTP,
					annotation.Annotation(annotation.HealthCheckFlag):     string(model.OnFlag),
					annotation.Annotation(annotation.HealthCheckTimeout):  "8",
					annotation.Annotation(annotation.HealthyThreshold):    "5",
					annotation.Annotation(annotation.UnhealthyThreshold):  "5",
					annotation.Annotation(annotation.HealthCheckInterval): "3",
					annotation.Annotation(annotation.HealthCheckHTTPCode): "http_3xx",
					annotation.Annotation(annotation.HealthCheckMethod):   "get",
					annotation.Annotation(annotation.HealthCheckDomain):   "192.168.0.3",
					annotation.Annotation(annotation.HealthCheckURI):      "/test/index.html",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// close health check
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("health-check: https", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):        "https:443",
					annotation.Annotation(annotation.CertID):              options.TestConfig.CertID,
					annotation.Annotation(annotation.HealthCheckType):     model.HTTP,
					annotation.Annotation(annotation.HealthCheckFlag):     string(model.OnFlag),
					annotation.Annotation(annotation.HealthCheckMethod):   "head",
					annotation.Annotation(annotation.HealthCheckTimeout):  "8",
					annotation.Annotation(annotation.HealthyThreshold):    "5",
					annotation.Annotation(annotation.UnhealthyThreshold):  "5",
					annotation.Annotation(annotation.HealthCheckInterval): "3",
					annotation.Annotation(annotation.HealthCheckHTTPCode): "http_3xx",
					annotation.Annotation(annotation.HealthCheckDomain):   "192.168.0.3",
					annotation.Annotation(annotation.HealthCheckURI):      "/test/index.html",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckMethod)] = "get"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckTimeout)] = "10"
				newsvc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "2"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckHTTPCode)] = "http_2xx"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckDomain)] = "192.168.0.2"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckURI)] = "/test/index1.html"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("health-check: http -> tcp", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):        "http:80",
					annotation.Annotation(annotation.HealthCheckType):     model.HTTP,
					annotation.Annotation(annotation.HealthCheckFlag):     string(model.OnFlag),
					annotation.Annotation(annotation.HealthCheckTimeout):  "8",
					annotation.Annotation(annotation.HealthyThreshold):    "5",
					annotation.Annotation(annotation.UnhealthyThreshold):  "5",
					annotation.Annotation(annotation.HealthCheckInterval): "3",
					annotation.Annotation(annotation.HealthCheckHTTPCode): "http_3xx",
					annotation.Annotation(annotation.HealthCheckDomain):   "192.168.0.3",
					annotation.Annotation(annotation.HealthCheckURI):      "/test/index.html",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations = map[string]string{
					annotation.Annotation(annotation.HealthCheckType):           model.TCP,
					annotation.Annotation(annotation.HealthCheckConnectTimeout): "8",
					annotation.Annotation(annotation.HealthyThreshold):          "4",
					annotation.Annotation(annotation.UnhealthyThreshold):        "4",
					annotation.Annotation(annotation.HealthCheckInterval):       "3",
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})

		ginkgo.Context("persistence-timeout", func() {
			ginkgo.It("persistence-timeout", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.PersistenceTimeout): "0",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				//from close to open
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.PersistenceTimeout)] = "1000"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("established timeout", func() {
			ginkgo.It("established timeout: 20->10", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.EstablishedTimeout): "20",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.EstablishedTimeout)] = "10"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("session & cookie", func() {
			ginkgo.It("cookie-timeout: 1800 -> 100", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):     "http:80",
					annotation.Annotation(annotation.SessionStick):     string(model.OnFlag),
					annotation.Annotation(annotation.SessionStickType): "insert",
					annotation.Annotation(annotation.CookieTimeout):    "1800",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify cookie timeout
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.CookieTimeout)] = "100"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("cookie: cookie-AAAA -> cookie-BBBB", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):     "http:80",
					annotation.Annotation(annotation.SessionStick):     string(model.OnFlag),
					annotation.Annotation(annotation.SessionStickType): "server",
					annotation.Annotation(annotation.Cookie):           "cookie-AAAA",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.Cookie)] = "cookie-BBBB"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("sticky-session: on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):     "http:80",
					annotation.Annotation(annotation.SessionStick):     string(model.OnFlag),
					annotation.Annotation(annotation.SessionStickType): "insert",
					annotation.Annotation(annotation.CookieTimeout):    "1800",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.SessionStick)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("session-stick-type: insert -> server", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):     "http:80",
					annotation.Annotation(annotation.SessionStick):     string(model.OnFlag),
					annotation.Annotation(annotation.SessionStickType): "insert",
					annotation.Annotation(annotation.CookieTimeout):    "1800",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify session stick type
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations = map[string]string{
					annotation.Annotation(annotation.SessionStickType): "server",
					annotation.Annotation(annotation.Cookie):           "B490B52BCDA1598",
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("connection drain", func() {
			ginkgo.It("connection-drain-timeout: 10 -> 60", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort):           "tcp:80,udp:443",
					annotation.Annotation(annotation.ConnectionDrain):        string(model.OnFlag),
					annotation.Annotation(annotation.ConnectionDrainTimeout): "10",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify connection drain timeout
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ConnectionDrainTimeout)] = "60"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("connection-drain: on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ConnectionDrain):        string(model.OnFlag),
					annotation.Annotation(annotation.ConnectionDrainTimeout): "10",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// close connection drain
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ConnectionDrain)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("protocol-port", func() {
			ginkgo.It("protocol-port: udp:80 -> http:80", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort): "udp:80",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.ProtocolPort] = "http:80"
				newsvc, err = f.Client.KubeClient.PatchService(newsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("protocol-port: http:80,https:443", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "http:80,https:443",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("protocol-port: https:444; service without 444 port", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:444",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		if options.TestConfig.CertID != "" {
			ginkgo.Context("cert-id", func() {
				ginkgo.It("cert-id: certID", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				if options.TestConfig.CertID2 != "" {
					ginkgo.It("cert-id: certID -> certID2", func() {
						oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							annotation.Annotation(annotation.ProtocolPort): "https:443",
							annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldsvc)
						gomega.Expect(err).To(gomega.BeNil())

						//update certificate
						newsvc := oldsvc.DeepCopy()
						newsvc.Annotations[annotation.Annotation(annotation.CertID)] = options.TestConfig.CertID2
						newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(newsvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
				ginkgo.It("cert-id: no certID", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("cert-id: not exist", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443",
						annotation.Annotation(annotation.CertID):       "cert-xxxxx",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			})

			ginkgo.Context("forward-port", func() {
				ginkgo.It("create forward port : http 80 -> https 443", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("forward-port: 80:443,81:444", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations = make(map[string]string)
					newsvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolUDP,
						},
						{
							Name:       "https",
							Port:       443,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolUDP,
						},
						{
							Name:       "http1",
							Port:       81,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolUDP,
						},
						{
							Name:       "https1",
							Port:       444,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolUDP,
						},
					}
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "https:443,http:80,https:444,http:81"
					newsvc.Annotations[annotation.Annotation(annotation.CertID)] = options.TestConfig.CertID
					newsvc.Annotations[annotation.Annotation(annotation.ForwardPort)] = "80:443,81:444"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("forward-port: format error", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						// right 80:443
						annotation.Annotation(annotation.ForwardPort): "80,443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())

					// correct
					newsvc := svc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.ForwardPort)] = "80:443"
					newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())

				})
				ginkgo.It("modify forward port: http 80 -> 81", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					//update port
					newsvc := oldsvc.DeepCopy()
					newsvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Port:       81,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
						{
							Name:       "https",
							Port:       443,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newsvc.Annotations[annotation.Annotation(annotation.ForwardPort)] = "81:443"
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "https:443,http:81"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("modify forward port: https 443 -> 444 ", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					//update port
					newsvc := oldsvc.DeepCopy()
					newsvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
						{
							Name:       "https",
							Port:       444,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newsvc.Annotations[annotation.Annotation(annotation.ForwardPort)] = "80:444"
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "https:444,http:80"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					// The listener forward relationship was not deleted, expected failure
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("delete listener forward", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					delete(newsvc.Annotations, annotation.Annotation(annotation.ForwardPort))
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					// assert no forwardport and get tcp 80 and 443 listener
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("delete listener forward and modify forward port", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					forwardOffSvc := oldsvc.DeepCopy()
					delete(forwardOffSvc.Annotations, annotation.Annotation(annotation.ForwardPort))
					forwardOffSvc, err = f.Client.KubeClient.PatchService(oldsvc, forwardOffSvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(forwardOffSvc)
					gomega.Expect(err).To(gomega.BeNil())

					//update port
					updatesvc := forwardOffSvc.DeepCopy()
					updatesvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
						{
							Name:       "https",
							Port:       444,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
					}
					updatesvc.Annotations[annotation.Annotation(annotation.ForwardPort)] = "80:444"
					updatesvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "https:444,http:80"
					updatesvc, err = f.Client.KubeClient.PatchService(forwardOffSvc, updatesvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(updatesvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("forward port: delete http 80 listener", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "https",
							Port:       443,
							TargetPort: intstr.FromInt(443),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "https:443"
					delete(newsvc.Annotations, newsvc.Annotations[annotation.ForwardPort])
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("forward port: delete https 443 listener", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "https",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "http:80"
					delete(newsvc.Annotations, newsvc.Annotations[annotation.ForwardPort])
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("http:80,https:443,http:8080 and forward port: 80:443", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443,http:80,http:8080",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
						annotation.Annotation(annotation.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})

			ginkgo.Context("tls-cipher-policy", func() {
				ginkgo.It("tls-cipher-policy: tls_cipher_policy_1_1", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort):    "https:443",
						annotation.Annotation(annotation.CertID):          options.TestConfig.CertID,
						annotation.Annotation(annotation.TLSCipherPolicy): "tls_cipher_policy_1_1",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("tls-cipher-policy: tls_cipher_policy_1_1 -> tls_cipher_policy_1_2", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort):    "https:443",
						annotation.Annotation(annotation.CertID):          options.TestConfig.CertID,
						annotation.Annotation(annotation.TLSCipherPolicy): "tls_cipher_policy_1_1",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "https:443"
					newsvc.Annotations[annotation.Annotation(annotation.CertID)] = options.TestConfig.CertID
					newsvc.Annotations[annotation.Annotation(annotation.TLSCipherPolicy)] = "tls_cipher_policy_1_2"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("tls-cipher-policy: not exist policy", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort):    "https:443",
						annotation.Annotation(annotation.CertID):          options.TestConfig.CertID,
						annotation.Annotation(annotation.TLSCipherPolicy): "tls_cipher_policy_no_exist",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			})

		}

		ginkgo.Context("service port", func() {
			ginkgo.It("add & delete port", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations = make(map[string]string)
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       53,
						TargetPort: intstr.FromInt(53),
						Protocol:   v1.ProtocolUDP,
					},
				}
				newsvc.Annotations[annotation.ProtocolPort] = "udp:53"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("port: 80 -> 81; protocol: tcp", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				//update listener port
				newsvc := oldsvc.DeepCopy()
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       81,
						TargetPort: intstr.FromInt(81),
						Protocol:   v1.ProtocolTCP,
					},
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("port: 80 -> 81; protocol: http", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProtocolPort): "http:80",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				//update listener port
				newsvc := oldsvc.DeepCopy()
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       81,
						TargetPort: intstr.FromInt(81),
						Protocol:   v1.ProtocolTCP,
					},
				}
				newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "http:81"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("mixed port protocol", func() {
				rawsvc := f.Client.KubeClient.DefaultService()
				rawsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "tcp",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolTCP,
					},
					{
						Name:       "udp",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolUDP,
					},
				}
				oldsvc, err := f.Client.KubeClient.CreateService(rawsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				//update listener port
				newsvc := oldsvc.DeepCopy()
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "tcp",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolTCP,
					},
					{
						Name:       "udp",
						Port:       53,
						TargetPort: intstr.FromInt(53),
						Protocol:   v1.ProtocolUDP,
					},
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("port: 443 -> 444; protocol: https", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ProtocolPort): "https:443",
						annotation.Annotation(annotation.CertID):       options.TestConfig.CertID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					//update listener port
					newsvc := oldsvc.DeepCopy()
					newsvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "https",
							Port:       444,
							TargetPort: intstr.FromInt(443),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "https:444"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("sync", func() {
			ginkgo.It("delete listener of ccm created slb", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				_, lb, err := f.FindLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())

				err = f.Client.CloudClient.DeleteLoadBalancerListener(
					context.TODO(), lb.LoadBalancerAttribute.LoadBalancerId, 80, model.TCP)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldsvc.DeepCopy()
				newSvc.Annotations = make(map[string]string)
				newSvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s2.small"
				newSvc, err = f.Client.KubeClient.PatchService(oldsvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.InternetLoadBalancerID != "" {
				ginkgo.It("delete listener of reused slb", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.Spec):             model.S1Small,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					_, lb, err := f.FindLoadBalancer()
					gomega.Expect(err).To(gomega.BeNil())

					err = f.Client.CloudClient.SLBProvider.DeleteLoadBalancerListener(
						context.TODO(), lb.LoadBalancerAttribute.LoadBalancerId, 80, model.TCP)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldsvc.DeepCopy()
					newSvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s2.small"
					newSvc, err = f.Client.KubeClient.PatchService(oldsvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("proxy protocol", func() {
			ginkgo.It("proxy-protocol on", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProxyProtocol): string(model.OnFlag),
					annotation.Annotation(annotation.ProtocolPort):  "tcp:80,udp:443",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("proxy-protocol on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProxyProtocol): string(model.OnFlag),
					annotation.Annotation(annotation.ProtocolPort):  "tcp:80,udp:443",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("proxy-protocol off -> on", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ProxyProtocol): string(model.OffFlag),
					annotation.Annotation(annotation.ProtocolPort):  "tcp:80,udp:443",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
}
