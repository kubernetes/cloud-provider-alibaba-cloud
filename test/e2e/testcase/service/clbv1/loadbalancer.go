package clbv1

import (
	"context"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
)

func RunLoadBalancerTestCases(f *framework.Framework) {

	ginkgo.Describe("clb service controller: loadbalancer", func() {

		ginkgo.AfterEach(func() {
			ginkgo.By("delete service")
			err := f.AfterEach()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.Context("address-type", func() {
			ginkgo.It("address-type=internet", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(
					map[string]string{
						annotation.Annotation(annotation.AddressType): string(model.InternetAddressType),
					})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("address-type=intranet", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(
					map[string]string{
						annotation.Annotation(annotation.AddressType): string(model.IntranetAddressType),
					})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())

			})

			ginkgo.It("address-type: intranet->internet", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(
					map[string]string{
						annotation.Annotation(annotation.AddressType): string(model.IntranetAddressType),
					})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[annotation.Annotation(annotation.AddressType)] = string(model.InternetAddressType)
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		if options.TestConfig.VSwitchID != "" {
			ginkgo.Context("virtual switches", func() {

				ginkgo.It("create slb with vsw annotation", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AddressType): string(model.IntranetAddressType),
						annotation.Annotation(annotation.VswitchId):   options.TestConfig.VSwitchID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("vswid is not exist", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AddressType): string(model.IntranetAddressType),
						annotation.Annotation(annotation.VswitchId):   "vsw-xxxx",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				if options.TestConfig.VSwitchID2 != "" {
					ginkgo.It("update vsw id", func() {
						oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							annotation.Annotation(annotation.AddressType): string(model.IntranetAddressType),
							annotation.Annotation(annotation.VswitchId):   options.TestConfig.VSwitchID,
						})

						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())

						newSvc := oldSvc.DeepCopy()
						newSvc.Annotations[annotation.Annotation(annotation.VswitchId)] = options.TestConfig.VSwitchID2
						newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
						gomega.Expect(err).To(gomega.BeNil())

						err = f.ExpectLoadBalancerEqual(newSvc)
						gomega.Expect(err).NotTo(gomega.BeNil())
					})
				}
			})
		}

		ginkgo.Context("lb spec", func() {
			ginkgo.It("spec: s1.slb.small", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.Spec): model.S1Small,
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("spec: s1->s2", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.Spec): model.S1Small,
				})

				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s2.small"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("reuse lb", func() {
			if options.TestConfig.InternetLoadBalancerID != "" {
				ginkgo.It("reuse internet lb", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AddressType):    string(model.InternetAddressType),
						annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetLoadBalancerID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("reuse internet lb with override-listener=false", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AddressType):      string(model.InternetAddressType),
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "false",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("reuse internet lb with override-listener=true", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AddressType):      string(model.InternetAddressType),
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("ccm created slb -> reused slb", func() {
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
					gomega.Expect(err).To(gomega.BeNil())
					_, oldlb, err := f.FindLoadBalancer()
					gomega.Expect(err).To(gomega.BeNil())

					lbid := oldlb.LoadBalancerAttribute.LoadBalancerId
					defer func(id string) {
						err := f.DeleteLoadBalancer(id)
						gomega.Expect(err).To(gomega.BeNil())
					}(lbid)

					newsvc := oldSvc.DeepCopy()
					newsvc.Annotations = map[string]string{
						annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetLoadBalancerID,
					}
					newsvc, err = f.Client.KubeClient.PatchService(oldSvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("reuse intranet lb with override-listener=true", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AddressType):      string(model.IntranetAddressType),
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.IntranetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}

			ginkgo.It("reuse ccm created slb", func() {
				_, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				svc, remote, err := f.FindLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())
				defer func(id string) {
					err := f.DeleteLoadBalancer(id)
					gomega.Expect(err).To(gomega.BeNil())
				}(remote.LoadBalancerAttribute.LoadBalancerId)

				newSvc := svc.DeepCopy()
				newSvc.Annotations = map[string]string{
					annotation.Annotation(annotation.AddressType):      string(model.InternetAddressType),
					annotation.Annotation(annotation.LoadBalancerId):   remote.LoadBalancerAttribute.LoadBalancerId,
					annotation.Annotation(annotation.OverrideListener): "true",
					annotation.Annotation(annotation.Spec):             "slb.s2.small",
				}
				_, err = f.Client.KubeClient.PatchService(svc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("reuse not exist lb", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.LoadBalancerId):   "lb-xxxxxxxxxx",
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			if options.TestConfig.VPCLoadBalancerID != "" {
				ginkgo.It("reuse intranet lb in other VPC", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.VPCLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			}

			if options.TestConfig.EipLoadBalancerID != "" {
				ginkgo.It("reuse intranet lb with eip", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.EipLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
						annotation.Annotation(annotation.ExternalIPType):   "eip",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}

		})

		ginkgo.Context("hostname", func() {
			ginkgo.It("add hostname", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.HostName): "www.test.com",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())

			})
			ginkgo.It("remove hostname", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.HostName): "www.test.com",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				delete(newsvc.Annotations, annotation.HostName)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("update hostname", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.HostName): "www.test.com",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				updateSvc := oldSvc.DeepCopy()
				updateSvc.Annotations[annotation.Annotation(annotation.HostName)] = "www.update.com"
				updateSvc, err = f.Client.KubeClient.PatchService(oldSvc, updateSvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(updateSvc)
				gomega.Expect(err).To(gomega.BeNil())

			})
			if options.TestConfig.EipLoadBalancerID != "" {
				ginkgo.It("reuse intranet lb with eip && add hostname", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.EipLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
						annotation.Annotation(annotation.ExternalIPType):   "eip",
						annotation.Annotation(annotation.HostName):         "www.test.com",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := svc.DeepCopy()
					delete(newsvc.Annotations, annotation.Annotation(annotation.HostName))
					newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("charge type", func() {
			ginkgo.It("charge-type: paybybandwidth", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.AddressType): string(model.InternetAddressType),
					annotation.Annotation(annotation.ChargeType):  string(model.PayByBandwidth),
					annotation.Annotation(annotation.Bandwidth):   "5",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[annotation.Annotation(annotation.Bandwidth)] = "20"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("charge-type: paybytraffic", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.AddressType): string(model.InternetAddressType),
					annotation.Annotation(annotation.ChargeType):  "paybytraffic",
				})

				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("charge-type: paybybandwidth -> paybytraffic", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.AddressType): string(model.InternetAddressType),
					annotation.Annotation(annotation.ChargeType):  string(model.PayByBandwidth),
					annotation.Annotation(annotation.Bandwidth):   "5",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[annotation.Annotation(annotation.ChargeType)] = "paybytraffic"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				// paybytraffic are effective the next day
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("charge-type: paybybandwidth; address-type: intranet", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.AddressType): string(model.IntranetAddressType),
					annotation.Annotation(annotation.ChargeType):  string(model.PayByBandwidth),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("slb protection", func() {
			ginkgo.It("delete-protection: on -> off", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.DeleteProtection): string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[annotation.Annotation(annotation.DeleteProtection)] = string(model.OffFlag)
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("modification-protection: console -> none", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ModificationProtection): string(model.ConsoleProtection),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[annotation.Annotation(annotation.ModificationProtection)] = "NonProtection"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("delete-protection & modification-protection ", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ModificationProtection): string(model.ConsoleProtection),
					annotation.Annotation(annotation.DeleteProtection):       string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[annotation.Annotation(annotation.ModificationProtection)] = "NonProtection"
				newSvc.Annotations[annotation.Annotation(annotation.DeleteProtection)] = string(model.OffFlag)
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("lb name", func() {
			ginkgo.It("lb-name: test-lb-name", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.LoadBalancerName): "test-lb-name",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("lb-name: lb-name-1 -> lb-name-2", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.LoadBalancerName): "test-lb-name",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[annotation.Annotation(annotation.LoadBalancerName)] = "modify-lb-name"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		if options.TestConfig.ResourceGroupID != "" {
			ginkgo.Context("resource group", func() {
				ginkgo.It("create lb with resource-group-id", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ResourceGroupId): options.TestConfig.ResourceGroupID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("change resource-group-id", func() {
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Annotations = make(map[string]string)
					newSvc.Annotations[annotation.Annotation(annotation.ResourceGroupId)] = options.TestConfig.ResourceGroupID
					_, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("reuse lb and set resource-group-id is inconsistent with lb", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ResourceGroupId): options.TestConfig.ResourceGroupID,
						annotation.Annotation(annotation.LoadBalancerId):  options.TestConfig.InternetLoadBalancerID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			})
		}

		ginkgo.Context("IP version", func() {
			ginkgo.It("ip-version: ipv4", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.IPVersion): string(model.IPv4),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("ip-version: ipv6", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.IPVersion): string(model.IPv6),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("lb-version: ipv4 -> ipv6", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations = make(map[string]string)
				newSvc.Annotations[annotation.Annotation(annotation.IPVersion)] = string(model.IPv6)
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("add tags", func() {
			ginkgo.It("add tag for lb", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.AdditionalTags): "Key1=Value1,Key2=Value2",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("update tags", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations = make(map[string]string)
				newSvc.Annotations[annotation.Annotation(annotation.AdditionalTags)] = "Key1=Value1,Key2=Value2"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("add tag for reused lb", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.AdditionalTags): "Key1=Value1,Key2=Value2",
					annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetLoadBalancerID,
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		if options.TestConfig.MasterZoneID != "" && options.TestConfig.SlaveZoneID != "" {
			ginkgo.Context("zone", func() {
				ginkgo.It("create lb with master-zone and slave-zone", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.MasterZoneID): options.TestConfig.MasterZoneID,
						annotation.Annotation(annotation.SlaveZoneID):  options.TestConfig.SlaveZoneID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("create lb with not exist master-zone", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.MasterZoneID): "cn-xxxx-x",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("change master zone & slave zone", func() {
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Annotations = make(map[string]string)
					newSvc.Annotations[annotation.Annotation(annotation.MasterZoneID)] = options.TestConfig.SlaveZoneID
					newSvc.Annotations[annotation.Annotation(annotation.SlaveZoneID)] = options.TestConfig.MasterZoneID
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			})
		}

		ginkgo.Context("instance-charge-type", func() {
			ginkgo.It("instance-charge-type: PayBySpec -> PayByCLCU", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.InstanceChargeType): "PayBySpec",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.InstanceChargeType)] = "PayByCLCU"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("instance-charge-type: PayByCLCU -> PayBySpec without loadbalancer-spec", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.InstanceChargeType): "PayByCLCU",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				delete(newsvc.Annotations, annotation.Annotation(annotation.InstanceChargeType))
				newsvc.Annotations[annotation.Annotation(annotation.InstanceChargeType)] = "PayBySpec"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("instance-charge-type: PayByCLCU -> PayBySpec with loadbalancer-spec", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.InstanceChargeType): "PayByCLCU",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				delete(newsvc.Annotations, annotation.Annotation(annotation.InstanceChargeType))
				newsvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s2.small"
				newsvc.Annotations[annotation.Annotation(annotation.InstanceChargeType)] = "PayBySpec"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("instance-charge-type: PayByCLCU & spec annotation", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.InstanceChargeType): "PayByCLCU",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s1.small"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("instance-charge-type: PayBySpec & spec annotation", func() {
				oldsvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.Spec): "slb.s2.small",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s1.small"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("annotation prefix", func() {
			ginkgo.It("annotation prefix", func() {
				_, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.Spec):                  model.S1Small,
					"service.beta.kubernetes.io/alicloud-loadbalancer-spec": "slb.s2.small",
				})
				gomega.Expect(err).To(gomega.BeNil())
				_, slb, err := f.FindLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(string(slb.LoadBalancerAttribute.LoadBalancerSpec)).To(gomega.Equal(model.S1Small))
			})
		})

		ginkgo.Context("service label", func() {
			ginkgo.It("find slb", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())

				svc, err = f.Client.KubeClient.GetService()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(svc.Labels).Should(gomega.HaveKey(helper.LabelLoadBalancerId))
			})
		})

		ginkgo.Context("service type", func() {
			ginkgo.It("type: NodePort -> LoadBalancer", func() {
				svc := f.Client.KubeClient.DefaultService()
				svc.Spec.Type = v1.ServiceTypeNodePort
				_, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())

				svc, _, err = f.FindLoadBalancer()
				// should not create slb
				gomega.Expect(err).NotTo(gomega.BeNil())

				lbSvc := svc.DeepCopy()
				lbSvc.Spec.Type = v1.ServiceTypeLoadBalancer
				lbSvc, err = f.Client.KubeClient.PatchService(svc, lbSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(lbSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("type: LoadBalancer->NodePort", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				npSvc := oldSvc.DeepCopy()
				npSvc.Spec.Type = v1.ServiceTypeNodePort
				_, err = f.Client.KubeClient.PatchService(oldSvc, npSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerDeleted(npSvc)
				gomega.Expect(err).To(gomega.BeNil())
				svc, err := f.Client.KubeClient.GetService()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(svc.Labels).ShouldNot(gomega.HaveKey(helper.LabelServiceHash))
				gomega.Expect(svc.Labels).ShouldNot(gomega.HaveKey(helper.LabelLoadBalancerId))
			})
			ginkgo.It("type: ClusterIP -> LoadBalancer", func() {
				svc := f.Client.KubeClient.DefaultService()
				svc.Spec.Type = v1.ServiceTypeClusterIP
				_, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())

				_, _, err = f.FindLoadBalancer()
				// should not create slb
				gomega.Expect(err).NotTo(gomega.BeNil())

				lbSvc := svc.DeepCopy()
				lbSvc.Spec.Type = v1.ServiceTypeLoadBalancer
				lbSvc, err = f.Client.KubeClient.PatchService(svc, lbSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(lbSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("type: LoadBalancer -> ClusterIP", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				npSvc := oldSvc.DeepCopy()
				npSvc.Spec.Type = v1.ServiceTypeClusterIP
				_, err = f.Client.KubeClient.PatchService(oldSvc, npSvc)
				gomega.Expect(err).To(gomega.BeNil())

				gomega.Eventually(func(g gomega.Gomega) {
					_, lb, err := f.FindLoadBalancer()
					g.Expect(lb).To(gomega.BeNil())
					g.Expect(err).NotTo(gomega.BeNil())
				}).Should(gomega.Succeed())
			})
		})

		ginkgo.Context("can not find slb", func() {
			ginkgo.It("auto-created slb & delete slb", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				_, slb, err := f.FindLoadBalancer()
				err = f.DeleteLoadBalancer(slb.LoadBalancerAttribute.LoadBalancerId)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.Client.KubeClient.ScaleDeployment(0)
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					err = f.Client.KubeClient.ScaleDeployment(3)
					gomega.Expect(err).To(gomega.BeNil())
				}()
				// can not find slb
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("auto-created slb & delete tag", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				_, slb, err := f.FindLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())
				err = f.Client.CloudClient.UntagResources(context.TODO(),
					slb.LoadBalancerAttribute.LoadBalancerId, &[]string{helper.TAGKEY})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.Client.KubeClient.ScaleDeployment(0)
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					err = f.Client.KubeClient.ScaleDeployment(3)
					gomega.Expect(err).To(gomega.BeNil())
				}()

				// find slb by name
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("auto-created slb & delete tag & change name", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(
					map[string]string{
						annotation.Annotation(annotation.LoadBalancerName): "test-lb-name",
					})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				_, slb, err := f.FindLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())
				err = f.Client.CloudClient.UntagResources(context.TODO(),
					slb.LoadBalancerAttribute.LoadBalancerId, &[]string{helper.TAGKEY})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.Client.KubeClient.ScaleDeployment(0)
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					err = f.Client.KubeClient.ScaleDeployment(3)
					gomega.Expect(err).To(gomega.BeNil())
				}()

				// can not find slb
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		if options.TestConfig.Address != "" {
			ginkgo.Context("loadbalancer address", func() {
				ginkgo.It("loadbalancer address", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.AddressType): string(model.IntranetAddressType),
						annotation.Annotation(annotation.VswitchId):   options.TestConfig.VSwitchID,
						annotation.Annotation(annotation.IP):          options.TestConfig.Address,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})
		}
	})

}
