package service

import (
	"context"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/options"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
)

func RunLoadBalancerTestCases(f *framework.Framework) {

	ginkgo.Describe("service controller: loadbalancer", func() {

		ginkgo.By("delete service")
		ginkgo.AfterEach(func() {
			err := f.AfterEach()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.Context("address-type", func() {
			ginkgo.It("address-type=internet", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(
					map[string]string{
						service.Annotation(service.AddressType): string(model.InternetAddressType),
					})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("address-type=intranet", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(
					map[string]string{
						service.Annotation(service.AddressType): string(model.IntranetAddressType),
					})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())

			})

			ginkgo.It("address-type: intranet->internet", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(
					map[string]string{
						service.Annotation(service.AddressType): string(model.IntranetAddressType),
					})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[service.Annotation(service.AddressType)] = string(model.InternetAddressType)
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
						service.Annotation(service.AddressType): string(model.IntranetAddressType),
						service.Annotation(service.VswitchId):   options.TestConfig.VSwitchID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("vswid is not exist", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.AddressType): string(model.IntranetAddressType),
						service.Annotation(service.VswitchId):   "vsw-xxxx",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				if options.TestConfig.VSwitchID2 != "" {
					ginkgo.It("update vsw id", func() {
						oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							service.Annotation(service.AddressType): string(model.IntranetAddressType),
							service.Annotation(service.VswitchId):   options.TestConfig.VSwitchID,
						})

						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())

						newSvc := oldSvc.DeepCopy()
						newSvc.Annotations[service.Annotation(service.VswitchId)] = options.TestConfig.VSwitchID2
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
					service.Annotation(service.Spec): model.S1Small,
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("spec: s1->s2", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.Spec): model.S1Small,
				})

				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[service.Annotation(service.Spec)] = "slb.s2.small"
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
						service.Annotation(service.AddressType):    string(model.InternetAddressType),
						service.Annotation(service.LoadBalancerId): options.TestConfig.InternetLoadBalancerID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("reuse internet lb with override-listener=false", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.AddressType):      string(model.InternetAddressType),
						service.Annotation(service.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						service.Annotation(service.OverrideListener): "false",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("reuse internet lb with override-listener=true", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.AddressType):      string(model.InternetAddressType),
						service.Annotation(service.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						service.Annotation(service.OverrideListener): "true",
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
						service.Annotation(service.LoadBalancerId): options.TestConfig.InternetLoadBalancerID,
					}
					newsvc, err = f.Client.KubeClient.PatchService(oldSvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("reuse intranet lb with override-listener=true", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.AddressType):      string(model.IntranetAddressType),
						service.Annotation(service.LoadBalancerId):   options.TestConfig.IntranetLoadBalancerID,
						service.Annotation(service.OverrideListener): "true",
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
					service.Annotation(service.AddressType):      string(model.InternetAddressType),
					service.Annotation(service.LoadBalancerId):   remote.LoadBalancerAttribute.LoadBalancerId,
					service.Annotation(service.OverrideListener): "true",
					service.Annotation(service.Spec):             "slb.s2.small",
				}
				_, err = f.Client.KubeClient.PatchService(svc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("reuse not exist lb", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.LoadBalancerId):   "lb-xxxxxxxxxx",
					service.Annotation(service.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			if options.TestConfig.VPCLoadBalancerID != "" {
				ginkgo.It("reuse intranet lb in other VPC", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.LoadBalancerId):   options.TestConfig.VPCLoadBalancerID,
						service.Annotation(service.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			}

			if options.TestConfig.EipLoadBalancerID != "" {
				ginkgo.It("reuse intranet lb with eip", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.LoadBalancerId):   options.TestConfig.EipLoadBalancerID,
						service.Annotation(service.OverrideListener): "true",
						service.Annotation(service.ExternalIPType):   "eip",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}

		})

		ginkgo.Context("charge type", func() {
			ginkgo.It("charge-type: paybybandwidth", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.AddressType): string(model.InternetAddressType),
					service.Annotation(service.ChargeType):  string(model.PayByBandwidth),
					service.Annotation(service.Bandwidth):   "5",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[service.Annotation(service.Bandwidth)] = "100"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("charge-type: paybytraffic", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.AddressType): string(model.InternetAddressType),
					service.Annotation(service.ChargeType):  "paybytraffic",
				})

				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("charge-type: paybybandwidth -> paybytraffic", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.AddressType): string(model.InternetAddressType),
					service.Annotation(service.ChargeType):  string(model.PayByBandwidth),
					service.Annotation(service.Bandwidth):   "5",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[service.Annotation(service.ChargeType)] = "paybytraffic"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("charge-type: paybybandwidth; address-type: intranet", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.AddressType): string(model.IntranetAddressType),
					service.Annotation(service.ChargeType):  string(model.PayByBandwidth),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("slb protection", func() {
			ginkgo.It("delete-protection: on -> off", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.DeleteProtection): string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[service.Annotation(service.DeleteProtection)] = string(model.OffFlag)
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("modification-protection: console -> none", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ModificationProtection): string(model.ConsoleProtection),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[service.Annotation(service.ModificationProtection)] = "NonProtection"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("delete-protection & modification-protection ", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.ModificationProtection): string(model.ConsoleProtection),
					service.Annotation(service.DeleteProtection):       string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[service.Annotation(service.ModificationProtection)] = "NonProtection"
				newSvc.Annotations[service.Annotation(service.DeleteProtection)] = string(model.OffFlag)
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("lb name", func() {
			ginkgo.It("lb-name: test-lb-name", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.LoadBalancerName): "test-lb-name",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("lb-name: lb-name-1 -> lb-name-2", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.LoadBalancerName): "test-lb-name",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				newSvc := oldSvc.DeepCopy()
				newSvc.Annotations[service.Annotation(service.LoadBalancerName)] = "modify-lb-name"
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
						service.Annotation(service.ResourceGroupId): options.TestConfig.ResourceGroupID,
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
					newSvc.Annotations[service.Annotation(service.ResourceGroupId)] = options.TestConfig.ResourceGroupID
					_, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("reuse lb and set resource-group-id is inconsistent with lb", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.ResourceGroupId): options.TestConfig.ResourceGroupID,
						service.Annotation(service.LoadBalancerId):  options.TestConfig.InternetLoadBalancerID,
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
					service.Annotation(service.IPVersion): string(model.IPv4),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("ip-version: ipv6", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.IPVersion): string(model.IPv6),
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
				newSvc.Annotations[service.Annotation(service.IPVersion)] = string(model.IPv6)
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("add tags", func() {
			ginkgo.It("add tag for lb", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.AdditionalTags): "Key1=Value1,Key2=Value2",
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
				newSvc.Annotations[service.Annotation(service.AdditionalTags)] = "Key1=Value1,Key2=Value2"
				newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("add tag for reused lb", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.AdditionalTags): "Key1=Value1,Key2=Value2",
					service.Annotation(service.LoadBalancerId): options.TestConfig.InternetLoadBalancerID,
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
						service.Annotation(service.MasterZoneID): options.TestConfig.MasterZoneID,
						service.Annotation(service.SlaveZoneID):  options.TestConfig.SlaveZoneID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("create lb with not exist master-zone", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						service.Annotation(service.MasterZoneID): "cn-xxxx-x",
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
					newSvc.Annotations[service.Annotation(service.MasterZoneID)] = options.TestConfig.SlaveZoneID
					newSvc.Annotations[service.Annotation(service.SlaveZoneID)] = options.TestConfig.MasterZoneID
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			})
		}

		ginkgo.Context("annotation prefix", func() {
			ginkgo.It("annotation prefix", func() {
				_, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					service.Annotation(service.Spec):                        model.S1Small,
					"service.beta.kubernetes.io/alicloud-loadbalancer-spec": "slb.s2.small",
				})
				gomega.Expect(err).To(gomega.BeNil())
				_, slb, err := f.FindLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(string(slb.LoadBalancerAttribute.LoadBalancerSpec)).To(gomega.Equal(model.S1Small))
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

				gomega.Eventually(func(g gomega.Gomega) {
					_, lb, err := f.FindLoadBalancer()
					g.Expect(lb).To(gomega.BeNil())
					g.Expect(err).NotTo(gomega.BeNil())
				}).Should(gomega.Succeed())
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
					slb.LoadBalancerAttribute.LoadBalancerId, &[]string{service.TAGKEY})
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
						service.Annotation(service.LoadBalancerName): "test-lb-name",
					})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				_, slb, err := f.FindLoadBalancer()
				gomega.Expect(err).To(gomega.BeNil())
				err = f.Client.CloudClient.UntagResources(context.TODO(),
					slb.LoadBalancerAttribute.LoadBalancerId, &[]string{service.TAGKEY})
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
	})

}
