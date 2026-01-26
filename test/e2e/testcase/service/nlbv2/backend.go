package nlbv2

import (
	"context"
	"fmt"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/client"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
)

func RunBackendTestCases(f *framework.Framework) {

	ginkgo.Describe("nlb service controller: backend", func() {
		ginkgo.AfterEach(func() {
			ginkgo.By("delete service")
			err := f.AfterEach()
			gomega.Expect(err).To(gomega.BeNil())
		})

		testsvc := f.Client.KubeClient.DefaultService()
		testsvc.Annotations = map[string]string{
			annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
			annotation.Annotation(annotation.ProtocolPort):     "udp:53",
			annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
			annotation.Annotation(annotation.OverrideListener): "true",
		}
		testsvc.Spec.Ports = []v1.ServicePort{
			{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(80),
				Protocol:   v1.ProtocolTCP,
			},
			{
				Name:       "udp",
				Port:       53,
				TargetPort: intstr.FromInt(80),
				Protocol:   v1.ProtocolTCP,
			},
		}

		if options.TestConfig.NLBCertID != "" {
			testsvc.Spec.Ports = append(testsvc.Spec.Ports, v1.ServicePort{
				Name:       "tcpssl",
				Port:       443,
				TargetPort: intstr.FromInt(80),
				Protocol:   v1.ProtocolTCP,
			})
			testsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "udp:53,tcpssl:443"
			testsvc.Annotations[annotation.Annotation(annotation.CertID)] = options.TestConfig.NLBCertID
		}
		lbClass := helper.NLBClass
		testsvc.Spec.LoadBalancerClass = &lbClass

		ginkgo.Describe("health check", func() {
			ginkgo.It("tcp health check", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = "tcp"
				svc.Annotations[annotation.Annotation(annotation.HealthCheckConnectTimeout)] = "12"
				svc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				svc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				svc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "5"

				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("http", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = "http"
				svc.Annotations[annotation.Annotation(annotation.HealthCheckURI)] = "/"
				svc.Annotations[annotation.Annotation(annotation.HealthCheckDomain)] = "example.com"
				svc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				svc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				svc.Annotations[annotation.Annotation(annotation.HealthCheckConnectTimeout)] = "12"
				svc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "5"
				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("udp", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = "udp"
				svc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				svc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				// udp health check interval should be greater than or equal to health check connect timeout
				svc.Annotations[annotation.Annotation(annotation.HealthCheckConnectTimeout)] = "5"
				svc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "5"
				for i := range svc.Spec.Ports {
					svc.Spec.Ports[i].Protocol = v1.ProtocolUDP
				}
				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())

			})

			ginkgo.It("udp with tcp port", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = "udp"
				svc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				svc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				// udp health check interval should be greater than or equal to health check connect timeout
				svc.Annotations[annotation.Annotation(annotation.HealthCheckConnectTimeout)] = "12"
				svc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "5"
				svc.Spec.Ports[0].Protocol = v1.ProtocolUDP
				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).NotTo(gomega.BeNil())

			})

			ginkgo.It("health check: none -> tcp", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = "tcp"
				newsvc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "5"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("health check: none -> http", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = "http"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckURI)] = "/"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckDomain)] = "example.com"
				newsvc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "4"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckConnectTimeout)] = "12"
				newsvc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "5"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("flag on with listener port range", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.ListenerPortRange)] = "40-53:53,60-80:80"

				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("udp type with listener port range", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = string(model.UDP)
				svc.Annotations[annotation.Annotation(annotation.ListenerPortRange)] = "40-53:53,60-80:80"
				for i := range svc.Spec.Ports {
					svc.Spec.Ports[i].Protocol = v1.ProtocolUDP
				}

				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Describe("scheduler", func() {
			ginkgo.It("sch", func() {
				svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.Scheduler):        "sch",
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("sch -> rr", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.Scheduler):        "sch",
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.Scheduler)] = "rr"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Describe("connection drain", func() {
			ginkgo.It("on", func() {
				svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):               options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.ConnectionDrain):        string(model.OnFlag),
					annotation.Annotation(annotation.ConnectionDrainTimeout): "12",
					annotation.Annotation(annotation.LoadBalancerId):         options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener):       "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("off -> on", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ConnectionDrain)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.ConnectionDrainTimeout)] = "12"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):               options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.ConnectionDrain):        string(model.OnFlag),
					annotation.Annotation(annotation.ConnectionDrainTimeout): "12",
					annotation.Annotation(annotation.LoadBalancerId):         options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener):       "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.ConnectionDrain] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Describe("preserve client ip", func() {
			ginkgo.It("on", func() {
				svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.PreserveClientIp): string(model.OnFlag),
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("off -> on", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.PreserveClientIp)] = string(model.OnFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("on -> off", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.PreserveClientIp): string(model.OnFlag),
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.PreserveClientIp] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("externalTrafficPolicy", func() {
			ginkgo.It("cluster -> local", func() {
				svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := svc.DeepCopy()
				newsvc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
				newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.Network == options.Terway {
				ginkgo.It("local -> eni", func() {
					lbClass := helper.NLBClass
					svc := f.Client.KubeClient.DefaultService()
					svc.Spec.LoadBalancerClass = &lbClass
					svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
					svc.Annotations = map[string]string{
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					}

					svc, err := f.Client.KubeClient.CreateService(svc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := svc.DeepCopy()
					newsvc.Annotations[annotation.BackendType] = model.ENIBackendType
					newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("cluster -> eni", func() {
					svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})

					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := svc.DeepCopy()
					newsvc.Annotations[annotation.BackendType] = model.ENIBackendType
					newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("to-be-deleted-taint", func() {
			ginkgo.It("node: to-be-deleted-taint", func() {
				taint := v1.Taint{
					Key:    helper.ToBeDeletedTaint,
					Value:  fmt.Sprint(time.Now().Unix()),
					Effect: v1.TaintEffectNoSchedule,
				}
				// add ToBeDeletedTaint
				node, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(node).NotTo(gomega.BeNil())
				err = f.Client.KubeClient.AddTaint(node.Name, taint)
				gomega.Expect(err).To(gomega.BeNil())

				oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.Client.KubeClient.RemoveTaint(node.Name, taint)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})

		if options.TestConfig.Network == options.Terway {
			ginkgo.Context("backend-type", func() {
				ginkgo.It("backend-type: eni", func() {
					svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.BackendType:                             model.ENIBackendType,
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("backend-type: eni -> ecs", func() {
					oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.BackendType:                             model.ENIBackendType,
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Annotations[annotation.BackendType] = model.ECSBackendType
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectNetworkLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})
		}

		ginkgo.Context("special endpoints", func() {
			ginkgo.It("service with no selector", func() {
				svc, err := f.Client.KubeClient.CreateNLBServiceWithoutSelector(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("endpoint with not exist node", func() {
				// only for ecs mode
				svc, err := f.Client.KubeClient.CreateNLBServiceWithoutSelector(map[string]string{
					annotation.BackendType:                             model.ECSBackendType,
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				_, err = f.Client.KubeClient.CreateEndpointsWithNotExistNode()
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("endpoint without node name", func() {
				// only for ecs mode
				svc, err := f.Client.KubeClient.CreateNLBServiceWithoutSelector(map[string]string{
					annotation.BackendType:                             model.ECSBackendType,
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				_, err = f.Client.KubeClient.CreateEndpointsWithoutNodeName()
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				// if a endpoint does not have node name, fail
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("target port", func() {
			ginkgo.It("targetPort: 80 -> 81; ecs mode", func() {
				oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(81),
						Protocol:   v1.ProtocolTCP,
					},
				}
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectNetworkLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.Network == options.Terway {
				ginkgo.It("targetPort: 80 -> 81; eni mode", func() {
					oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.BackendType:                             model.ENIBackendType,
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(81),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectNetworkLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("targetPort: http; eni mode", func() {
					svc, err := f.Client.KubeClient.CreateNLBServiceWithStringTargetPort(map[string]string{
						annotation.BackendType:                             model.ENIBackendType,
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("targetPort: 80 -> http; eni mode", func() {
					oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.BackendType:                             model.ENIBackendType,
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					// update listener port
					newSvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromString("http"),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectNetworkLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("targetPort: http -> 80; eni mode", func() {
					lbClass := helper.NLBClass
					svc := f.Client.KubeClient.DefaultService()
					svc.Spec.LoadBalancerClass = &lbClass
					svc.Annotations = map[string]string{
						annotation.BackendType:                             model.ENIBackendType,
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					}
					svc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromString("http"),
							Protocol:   v1.ProtocolTCP,
						},
					}
					svc, err := f.Client.KubeClient.CreateService(svc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := svc.DeepCopy()
					newSvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newSvc, err = f.Client.KubeClient.PatchService(svc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectNetworkLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("exclude-balancer", func() {
			ginkgo.It("exclude-balancer", func() {
				// label node
				node, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(node).NotTo(gomega.BeNil())
				err = f.Client.KubeClient.LabelNode(node.Name, helper.LabelNodeExcludeBalancer, "true")
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					_ = f.Client.KubeClient.UnLabelNode(node.Name, helper.LabelNodeExcludeBalancer)
				}()

				oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("exclude-node", func() {
			ginkgo.It("exclude-node", func() {
				// label node
				node, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(node).NotTo(gomega.BeNil())
				err = f.Client.KubeClient.LabelNode(node.Name, client.ExcludeNodeLabel, "true")
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					_ = f.Client.KubeClient.UnLabelNode(node.Name, client.ExcludeNodeLabel)
				}()

				oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("update-backend", func() {
			ginkgo.It("scale deploy", func() {
				rawsvc := f.Client.KubeClient.DefaultNLBService()
				rawsvc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
				rawsvc.Annotations = map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				}
				oldSvc, err := f.Client.KubeClient.CreateService(rawsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				// scale deploy
				err = f.Client.KubeClient.ScaleDeployment(1)
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					err = f.Client.KubeClient.ScaleDeployment(3)
					gomega.Expect(err).To(gomega.BeNil())
				}()

				err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("server-group-type", func() {
			ginkgo.It("ip", func() {
				rawsvc := f.Client.KubeClient.DefaultNLBService()
				rawsvc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
				rawsvc.Annotations = map[string]string{
					annotation.Annotation(annotation.ZoneMaps):        options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.ServerGroupType): string(nlb.IpServerGroupType),
				}
				svc, err := f.Client.KubeClient.CreateService(rawsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("ip -> instance", func() {
				rawsvc := f.Client.KubeClient.DefaultNLBService()
				rawsvc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
				rawsvc.Annotations = map[string]string{
					annotation.Annotation(annotation.ZoneMaps):        options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.ServerGroupType): string(nlb.IpServerGroupType),
				}
				oldsvc, err := f.Client.KubeClient.CreateService(rawsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ServerGroupType)] = string(nlb.InstanceServerGroupType)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		if options.TestConfig.InternetNetworkLoadBalancerID != "" && options.TestConfig.NLBServerGroupID != "" {
			ginkgo.Context("vgroup-port", func() {
				ginkgo.It("vgroup-port: sg-id-1:80", func() {
					vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.NLBServerGroupID, 80)
					svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.VGroupPort):       vGroupPort,
						annotation.Annotation(annotation.OverrideListener): "false",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				if options.TestConfig.NLBServerGroupID2 != "" {
					ginkgo.It("vgroup-port: sg-id-1:80, sg-id-2:443", func() {
						vGroupPort := fmt.Sprintf("%s:%d,%s:%d", options.TestConfig.NLBServerGroupID, 80, options.TestConfig.NLBServerGroupID2, 443)
						svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.VGroupPort):       vGroupPort,
							annotation.Annotation(annotation.OverrideListener): "false",
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(svc)
						gomega.Expect(err).To(gomega.BeNil())
					})
					ginkgo.It("vgroup-port: sg-id-1:80 -> sg-id-2:80", func() {
						vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.NLBServerGroupID, 80)
						oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.VGroupPort):       vGroupPort,
							annotation.Annotation(annotation.OverrideListener): "false",
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())
						newSvc := oldSvc.DeepCopy()
						newVGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.NLBServerGroupID2, 80)
						newSvc.Annotations[annotation.Annotation(annotation.VGroupPort)] = newVGroupPort
						newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(newSvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
					ginkgo.It("vgroup-port: not exist sgp-id", func() {
						svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.VGroupPort):       "sgp-id-not-exist:80",
							annotation.Annotation(annotation.OverrideListener): "false",
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(svc)
						gomega.Expect(err).NotTo(gomega.BeNil())
					})
				}
			})

			ginkgo.Context("weight", func() {
				ginkgo.It("cluster mode: weight: 60 -> 80", func() {
					vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.NLBServerGroupID, 80)
					oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.VGroupPort):       vGroupPort,
						annotation.Annotation(annotation.VGroupWeight):     "60",
						annotation.Annotation(annotation.OverrideListener): "false",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("local mode: weight: 60 -> 80", func() {
					vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.NLBServerGroupID, 80)
					svc := f.Client.KubeClient.DefaultNLBService()
					svc.Annotations = map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.VGroupPort):       vGroupPort,
						annotation.Annotation(annotation.VGroupWeight):     "60",
						annotation.Annotation(annotation.OverrideListener): "false",
					}
					svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
					oldSvc, err := f.Client.KubeClient.CreateService(svc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				if options.TestConfig.Network == options.Terway {
					ginkgo.It("eni mode: weight: 60 -> 80", func() {
						vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.NLBServerGroupID, 80)
						oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.VGroupPort):       vGroupPort,
							annotation.Annotation(annotation.VGroupWeight):     "60",
							annotation.Annotation(annotation.OverrideListener): "false",
							annotation.BackendType:                             model.ENIBackendType,
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())

						newSvc := oldSvc.DeepCopy()
						newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
						newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(newSvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
					ginkgo.It("ecs mode; weight: nil -> 80", func() {
						vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.NLBServerGroupID, 80)
						svc := f.Client.KubeClient.DefaultNLBService()
						svc.Annotations = map[string]string{
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.VGroupPort):       vGroupPort,
							annotation.Annotation(annotation.OverrideListener): "false",
							annotation.BackendType:                             model.ECSBackendType,
						}
						svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
						oldSvc, err := f.Client.KubeClient.CreateService(svc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())

						newSvc := oldSvc.DeepCopy()
						newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
						newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(newSvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
					ginkgo.It("weight: 0 -> 100", func() {
						vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.NLBServerGroupID, 80)
						oldSvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.VGroupPort):       vGroupPort,
							annotation.Annotation(annotation.VGroupWeight):     "0",
							annotation.Annotation(annotation.OverrideListener): "false",
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())
						newSvc := oldSvc.DeepCopy()
						newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "100"
						newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(newSvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
			})
		}

		ginkgo.Context("ignore-weight-update", func() {
			ginkgo.It("should update weight without annotation", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetNetworkLoadBalancerID,
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				_, lb, err := f.FindNetworkLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(lb.ServerGroups).NotTo(gomega.BeEmpty())

				b := lb.ServerGroups[0].Servers
				gomega.Expect(b).NotTo(gomega.BeEmpty())
				if b[0].Weight == 100 {
					b[0].Weight = 1
				} else {
					b[0].Weight += 1
				}

				err = f.Client.CloudClient.UpdateNLBServers(context.TODO(), lb.ServerGroups[0].ServerGroupId, b)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldsvc.DeepCopy()
				newSvc.Annotations["reconcile-timestamp"] = time.Now().String()
				newSvc, err = f.Client.KubeClient.PatchService(oldsvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not update weight", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.LoadBalancerId):     options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.IgnoreWeightUpdate): string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				_, lb, err := f.FindNetworkLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(lb.ServerGroups).NotTo(gomega.BeEmpty())

				b := lb.ServerGroups[0].Servers
				gomega.Expect(b).NotTo(gomega.BeEmpty())
				if b[0].Weight == 100 {
					b[0].Weight = 1
				} else {
					b[0].Weight += 1
				}

				err = f.Client.CloudClient.UpdateNLBServers(context.TODO(), lb.ServerGroups[0].ServerGroupId, b)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldsvc.DeepCopy()
				newSvc.Annotations["reconcile-timestamp"] = time.Now().String()
				newSvc, err = f.Client.KubeClient.PatchService(oldsvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("named port", func() {
			ginkgo.It("named http port", func() {
				svc := f.Client.KubeClient.DefaultNLBService()
				svc.Annotations = map[string]string{
					annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetNetworkLoadBalancerID,
				}
				svc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromString("http"),
						Protocol:   v1.ProtocolTCP,
					},
					{
						Name:       "https",
						Port:       433,
						TargetPort: intstr.FromString("https"),
						Protocol:   v1.ProtocolTCP,
					},
				}

				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("nonexistent named port", func() {
				svc := f.Client.KubeClient.DefaultNLBService()
				svc.Annotations = map[string]string{
					annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetNetworkLoadBalancerID,
				}
				svc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromString("http"),
						Protocol:   v1.ProtocolTCP,
					},
					{
						Name:       "https",
						Port:       433,
						TargetPort: intstr.FromString("nonexistent"),
						Protocol:   v1.ProtocolTCP,
					},
				}

				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("named port select partial pods", func() {
				err := f.Client.KubeClient.CreateSecondaryDeployment()
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					err = f.Client.KubeClient.DeleteSecondaryDeployment()
					gomega.Expect(err).To(gomega.BeNil())
				}()

				svc := f.Client.KubeClient.DefaultNLBService()
				svc.Annotations = map[string]string{
					annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetNetworkLoadBalancerID,
				}
				svc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromString("http"),
						Protocol:   v1.ProtocolTCP,
					},
					{
						Name:       "https",
						Port:       433,
						TargetPort: intstr.FromString("https"),
						Protocol:   v1.ProtocolTCP,
					},
					{
						Name:       "https-1",
						Port:       444,
						TargetPort: intstr.FromString("https-1"),
						Protocol:   v1.ProtocolTCP,
					},
				}

				svc, err = f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid backends", func() {
			if options.TestConfig.Network == options.Terway {
				ginkgo.It("not found eni id", func() {
					pods, err := f.Client.KubeClient.GetDeploymentPods()
					gomega.Expect(err).To(gomega.BeNil())
					gomega.Expect(pods).NotTo(gomega.BeEmpty())
					var ips []string
					for _, p := range pods {
						ips = append(ips, p.Status.PodIP)
					}
					vsw, err := f.Client.CloudClient.VswitchID()
					gomega.Expect(err).To(gomega.BeNil())
					ip, err := f.FindFreeIPv4AddressFromVSwitch(context.TODO(), vsw)
					gomega.Expect(err).To(gomega.BeNil())
					ips = append(ips, ip.String())

					svc, err := f.Client.KubeClient.CreateNLBServiceWithoutSelector(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					_, err = f.Client.KubeClient.CreateEndpointsWithIPs(svc, ips)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectNetworkLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEvent(svc, helper.SkipSyncBackends, ip.String())
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})
	})
}
