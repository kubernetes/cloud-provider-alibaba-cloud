package service

import (
	"context"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/options"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
)

func RunListenerTestCases(f *framework.Framework) {
	ginkgo.Describe("service controller: listener", func() {

		ginkgo.By("delete service")
		ginkgo.AfterEach(func() {
			err := f.AfterEach()
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.Context("scheduler", func() {
			ginkgo.It("scheduler: rr", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.Scheduler): "rr",
				})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("scheduler: wrr", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.Scheduler): "wrr",
				})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("scheduler: wlc", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.Scheduler): "wlc",
				})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("scheduler: rr -> wrr", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.Scheduler): "rr",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.ObjectMeta.Annotations[service.Annotation(service.Scheduler)] = "wrr"
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
						service.Annotation(service.AclID):     options.TestConfig.AclID,
						service.Annotation(service.AclStatus): string(model.OnFlag),
						service.Annotation(service.AclType):   "white",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.AclType)] = "black"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())

				})
				ginkgo.It("acl-status: on -> off", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.AclID):     options.TestConfig.AclID,
						service.Annotation(service.AclStatus): string(model.OnFlag),
						service.Annotation(service.AclType):   "white",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.AclStatus)] = string(model.OffFlag)
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				if options.TestConfig.AclID2 != "" {
					ginkgo.It("update acl-id", func() {
						oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							service.Annotation(service.AclID):     options.TestConfig.AclID,
							service.Annotation(service.AclStatus): string(model.OnFlag),
							service.Annotation(service.AclType):   "white",
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldsvc)
						gomega.Expect(err).To(gomega.BeNil())

						newsvc := oldsvc.DeepCopy()
						newsvc.Annotations[service.Annotation(service.AclID)] = options.TestConfig.AclID2
						newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
						gomega.Expect(err).To(gomega.BeNil())

						err = f.ExpectLoadBalancerEqual(newsvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
				ginkgo.It("acl-id: not exist", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.AclID):     "acl-xxxxx",
						service.Annotation(service.AclStatus): string(model.OnFlag),
						service.Annotation(service.AclType):   "white",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("acl-id: exist -> not exist", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.AclID):     options.TestConfig.AclID,
						service.Annotation(service.AclStatus): string(model.OnFlag),
						service.Annotation(service.AclType):   "white",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.AclID)] = "acl-xxxxx"
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
					service.Annotation(service.ProtocolPort):       "http:80",
					service.Annotation(service.XForwardedForProto): string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.XForwardedForProto)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("xforwardedfor-proto; https", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort):       "http:443",
						service.Annotation(service.CertID):             options.TestConfig.CertID,
						service.Annotation(service.XForwardedForProto): string(model.OnFlag),
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.XForwardedForProto)] = string(model.OffFlag)
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
					service.Annotation(service.ProtocolPort): "http:80",
					service.Annotation(service.IdleTimeout):  "10",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.IdleTimeout)] = "40"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("idle-timeout: 10 -> 40", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.IdleTimeout):  "10",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					// modify
					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.IdleTimeout)] = "40"
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
					service.Annotation(service.ProtocolPort):   "https:443",
					service.Annotation(service.CertID):         options.TestConfig.CertID,
					service.Annotation(service.RequestTimeout): "1",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.RequestTimeout)] = "40"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("request-timeout: 100 -> 180", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort):   "https:443",
						service.Annotation(service.CertID):         options.TestConfig.CertID,
						service.Annotation(service.RequestTimeout): "100",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					// modify
					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.RequestTimeout)] = "180"
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
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.EnableHttp2):  string(model.OnFlag),
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					// modify
					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.EnableHttp2)] = string(model.OffFlag)
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
					service.Annotation(service.HealthCheckType):           model.TCP,
					service.Annotation(service.HealthCheckConnectTimeout): "8",
					service.Annotation(service.HealthyThreshold):          "5",
					service.Annotation(service.UnhealthyThreshold):        "5",
					service.Annotation(service.HealthCheckInterval):       "3",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.HealthCheckConnectTimeout)] = "10"
				newsvc.Annotations[service.Annotation(service.HealthyThreshold)] = "4"
				newsvc.Annotations[service.Annotation(service.UnhealthyThreshold)] = "4"
				newsvc.Annotations[service.Annotation(service.HealthCheckInterval)] = "2"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("health-check: http", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort):        "http:80",
					service.Annotation(service.HealthCheckType):     model.HTTP,
					service.Annotation(service.HealthCheckFlag):     string(model.OnFlag),
					service.Annotation(service.HealthCheckTimeout):  "8",
					service.Annotation(service.HealthyThreshold):    "5",
					service.Annotation(service.UnhealthyThreshold):  "5",
					service.Annotation(service.HealthCheckInterval): "3",
					service.Annotation(service.HealthCheckMethod):   "head",
					service.Annotation(service.HealthCheckHTTPCode): "http_3xx",
					service.Annotation(service.HealthCheckDomain):   "192.168.0.3",
					service.Annotation(service.HealthCheckURI):      "/test/index.html",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.HealthCheckTimeout)] = "10"
				newsvc.Annotations[service.Annotation(service.HealthyThreshold)] = "4"
				newsvc.Annotations[service.Annotation(service.UnhealthyThreshold)] = "4"
				newsvc.Annotations[service.Annotation(service.HealthCheckInterval)] = "2"
				newsvc.Annotations[service.Annotation(service.HealthCheckMethod)] = "get"
				newsvc.Annotations[service.Annotation(service.HealthCheckHTTPCode)] = "http_2xx"
				newsvc.Annotations[service.Annotation(service.HealthCheckDomain)] = "192.168.0.2"
				newsvc.Annotations[service.Annotation(service.HealthCheckURI)] = "/test/index1.html"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("health-check-flag: on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort):        "http:80",
					service.Annotation(service.HealthCheckType):     model.HTTP,
					service.Annotation(service.HealthCheckFlag):     string(model.OnFlag),
					service.Annotation(service.HealthCheckTimeout):  "8",
					service.Annotation(service.HealthyThreshold):    "5",
					service.Annotation(service.UnhealthyThreshold):  "5",
					service.Annotation(service.HealthCheckInterval): "3",
					service.Annotation(service.HealthCheckHTTPCode): "http_3xx",
					service.Annotation(service.HealthCheckMethod):   "get",
					service.Annotation(service.HealthCheckDomain):   "192.168.0.3",
					service.Annotation(service.HealthCheckURI):      "/test/index.html",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// close health check
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.HealthCheckFlag)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("health-check: https", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort):        "https:443",
					service.Annotation(service.CertID):              options.TestConfig.CertID,
					service.Annotation(service.HealthCheckType):     model.HTTP,
					service.Annotation(service.HealthCheckFlag):     string(model.OnFlag),
					service.Annotation(service.HealthCheckMethod):   "head",
					service.Annotation(service.HealthCheckTimeout):  "8",
					service.Annotation(service.HealthyThreshold):    "5",
					service.Annotation(service.UnhealthyThreshold):  "5",
					service.Annotation(service.HealthCheckInterval): "3",
					service.Annotation(service.HealthCheckHTTPCode): "http_3xx",
					service.Annotation(service.HealthCheckDomain):   "192.168.0.3",
					service.Annotation(service.HealthCheckURI):      "/test/index.html",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.HealthCheckMethod)] = "get"
				newsvc.Annotations[service.Annotation(service.HealthCheckTimeout)] = "10"
				newsvc.Annotations[service.Annotation(service.HealthyThreshold)] = "4"
				newsvc.Annotations[service.Annotation(service.UnhealthyThreshold)] = "4"
				newsvc.Annotations[service.Annotation(service.HealthCheckInterval)] = "2"
				newsvc.Annotations[service.Annotation(service.HealthCheckHTTPCode)] = "http_2xx"
				newsvc.Annotations[service.Annotation(service.HealthCheckDomain)] = "192.168.0.2"
				newsvc.Annotations[service.Annotation(service.HealthCheckURI)] = "/test/index1.html"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("health-check: http -> tcp", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort):        "http:80",
					service.Annotation(service.HealthCheckType):     model.HTTP,
					service.Annotation(service.HealthCheckFlag):     string(model.OnFlag),
					service.Annotation(service.HealthCheckTimeout):  "8",
					service.Annotation(service.HealthyThreshold):    "5",
					service.Annotation(service.UnhealthyThreshold):  "5",
					service.Annotation(service.HealthCheckInterval): "3",
					service.Annotation(service.HealthCheckHTTPCode): "http_3xx",
					service.Annotation(service.HealthCheckDomain):   "192.168.0.3",
					service.Annotation(service.HealthCheckURI):      "/test/index.html",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations = map[string]string{
					service.Annotation(service.HealthCheckType):           model.TCP,
					service.Annotation(service.HealthCheckConnectTimeout): "8",
					service.Annotation(service.HealthyThreshold):          "4",
					service.Annotation(service.UnhealthyThreshold):        "4",
					service.Annotation(service.HealthCheckInterval):       "3",
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
					service.Annotation(service.PersistenceTimeout): "0",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				//from close to open
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.PersistenceTimeout)] = "1000"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("established timeout", func() {
			ginkgo.It("established timeout: 20->10", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.EstablishedTimeout): "20",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.EstablishedTimeout)] = "10"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("session & cookie", func() {
			ginkgo.It("cookie-timeout: 1800 -> 100", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort):     "http:80",
					service.Annotation(service.SessionStick):     string(model.OnFlag),
					service.Annotation(service.SessionStickType): "insert",
					service.Annotation(service.CookieTimeout):    "1800",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify cookie timeout
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.CookieTimeout)] = "100"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("cookie: cookie-AAAA -> cookie-BBBB", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort):     "http:80",
					service.Annotation(service.SessionStick):     string(model.OnFlag),
					service.Annotation(service.SessionStickType): "server",
					service.Annotation(service.Cookie):           "cookie-AAAA",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.Cookie)] = "cookie-BBBB"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("sticky-session: on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort):     "http:80",
					service.Annotation(service.SessionStick):     string(model.OnFlag),
					service.Annotation(service.SessionStickType): "insert",
					service.Annotation(service.CookieTimeout):    "1800",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.SessionStick)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("session-stick-type: insert -> server", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort):     "http:80",
					service.Annotation(service.SessionStick):     string(model.OnFlag),
					service.Annotation(service.SessionStickType): "insert",
					service.Annotation(service.CookieTimeout):    "1800",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify session stick type
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations = map[string]string{
					service.Annotation(service.SessionStickType): "server",
					service.Annotation(service.Cookie):           "B490B52BCDA1598",
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
					service.Annotation(service.ProtocolPort):           "tcp:80,udp:443",
					service.Annotation(service.ConnectionDrain):        string(model.OnFlag),
					service.Annotation(service.ConnectionDrainTimeout): "10",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// modify connection drain timeout
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.ConnectionDrainTimeout)] = "60"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("connection-drain: on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ConnectionDrain):        string(model.OnFlag),
					service.Annotation(service.ConnectionDrainTimeout): "10",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				// close connection drain
				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.Annotation(service.ConnectionDrain)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("protocol-port", func() {
			ginkgo.It("protocol-port: udp:80 -> http:80", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ProtocolPort): "udp:80",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[service.ProtocolPort] = "http:80"
				newsvc, err = f.Client.KubeClient.PatchService(newsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("protocol-port: http:80,https:443", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "http:80,https:443",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("protocol-port: https:444; service without 444 port", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:444",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
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
						service.Annotation(service.ProtocolPort): "https:443",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				if options.TestConfig.CertID2 != "" {
					ginkgo.It("cert-id: certID -> certID2", func() {
						oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							service.Annotation(service.ProtocolPort): "https:443",
							service.Annotation(service.CertID):       options.TestConfig.CertID,
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldsvc)
						gomega.Expect(err).To(gomega.BeNil())

						//update certificate
						newsvc := oldsvc.DeepCopy()
						newsvc.Annotations[service.Annotation(service.CertID)] = options.TestConfig.CertID2
						newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(newsvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
				ginkgo.It("cert-id: no certID", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("cert-id: not exist", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443",
						service.Annotation(service.CertID):       "cert-xxxxx",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			})

			ginkgo.Context("forward-port", func() {
				ginkgo.It("create forward port : http 80 -> https 443", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.ForwardPort):  "80:443",
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
					newsvc.Annotations[service.Annotation(service.ProtocolPort)] = "https:443,http:80,https:444,http:81"
					newsvc.Annotations[service.Annotation(service.CertID)] = options.TestConfig.CertID
					newsvc.Annotations[service.Annotation(service.ForwardPort)] = "80:443,81:444"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("forward-port: format error", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						// right 80:443
						service.Annotation(service.ForwardPort): "80,443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())

					// correct
					newsvc := svc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.ForwardPort)] = "80:443"
					newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())

				})
				ginkgo.It("modify forward port: http 80 -> 81", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.ForwardPort):  "80:443",
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
					newsvc.Annotations[service.Annotation(service.ForwardPort)] = "81:443"
					newsvc.Annotations[service.Annotation(service.ProtocolPort)] = "https:443,http:81"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("modify forward port: https 443 -> 444 ", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.ForwardPort):  "80:443",
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
					newsvc.Annotations[service.Annotation(service.ForwardPort)] = "80:444"
					newsvc.Annotations[service.Annotation(service.ProtocolPort)] = "https:444,http:80"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					// The listener forward relationship was not deleted, expected failure
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("delete listener forward", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					delete(newsvc.Annotations, service.Annotation(service.ForwardPort))
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					// assert no forwardport and get tcp 80 and 443 listener
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("delete listener forward and modify forward port", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.ForwardPort):  "80:443",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					forwardOffSvc := oldsvc.DeepCopy()
					delete(forwardOffSvc.Annotations, service.Annotation(service.ForwardPort))
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
					updatesvc.Annotations[service.Annotation(service.ForwardPort)] = "80:444"
					updatesvc.Annotations[service.Annotation(service.ProtocolPort)] = "https:444,http:80"
					updatesvc, err = f.Client.KubeClient.PatchService(forwardOffSvc, updatesvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(updatesvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("forward port: delete http 80 listener", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.ForwardPort):  "80:443",
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
					newsvc.Annotations[service.Annotation(service.ProtocolPort)] = "https:443"
					delete(newsvc.Annotations, newsvc.Annotations[service.ForwardPort])
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("forward port: delete https 443 listener", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443,http:80",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
						service.Annotation(service.ForwardPort):  "80:443",
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
					newsvc.Annotations[service.Annotation(service.ProtocolPort)] = "http:80"
					delete(newsvc.Annotations, newsvc.Annotations[service.ForwardPort])
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})

			ginkgo.Context("tls-cipher-policy", func() {
				ginkgo.It("tls-cipher-policy: tls_cipher_policy_1_1", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort):    "https:443",
						service.Annotation(service.CertID):          options.TestConfig.CertID,
						service.Annotation(service.TLSCipherPolicy): "tls_cipher_policy_1_1",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("tls-cipher-policy: tls_cipher_policy_1_1 -> tls_cipher_policy_1_2", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort):    "https:443",
						service.Annotation(service.CertID):          options.TestConfig.CertID,
						service.Annotation(service.TLSCipherPolicy): "tls_cipher_policy_1_1",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[service.Annotation(service.ProtocolPort)] = "https:443"
					newsvc.Annotations[service.Annotation(service.CertID)] = options.TestConfig.CertID
					newsvc.Annotations[service.Annotation(service.TLSCipherPolicy)] = "tls_cipher_policy_1_2"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("tls-cipher-policy: not exist policy", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort):    "https:443",
						service.Annotation(service.CertID):          options.TestConfig.CertID,
						service.Annotation(service.TLSCipherPolicy): "tls_cipher_policy_no_exist",
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
				newsvc.Annotations[service.ProtocolPort] = "udp:53"
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
					service.Annotation(service.ProtocolPort): "http:80",
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
				newsvc.Annotations[service.Annotation(service.ProtocolPort)] = "http:81"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.CertID != "" {
				ginkgo.It("port: 443 -> 444; protocol: https", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ProtocolPort): "https:443",
						service.Annotation(service.CertID):       options.TestConfig.CertID,
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
					newsvc.Annotations[service.Annotation(service.ProtocolPort)] = "https:444"
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
					context.TODO(), lb.LoadBalancerAttribute.LoadBalancerId, 80)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldsvc.DeepCopy()
				newSvc.Annotations = make(map[string]string)
				newSvc.Annotations[service.Annotation(service.Spec)] = "slb.s2.small"
				newSvc, err = f.Client.KubeClient.PatchService(oldsvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.InternetLoadBalancerID != "" {
				ginkgo.It("delete listener of reused slb", func() {
					oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.Spec):             model.S1Small,
						service.Annotation(service.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						service.Annotation(service.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					_, lb, err := f.FindLoadBalancer()
					gomega.Expect(err).To(gomega.BeNil())

					err = f.Client.CloudClient.SLBProvider.DeleteLoadBalancerListener(
						context.TODO(), lb.LoadBalancerAttribute.LoadBalancerId, 80)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldsvc.DeepCopy()
					newSvc.Annotations[service.Annotation(service.Spec)] = "slb.s2.small"
					newSvc, err = f.Client.KubeClient.PatchService(oldsvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

	})
}
