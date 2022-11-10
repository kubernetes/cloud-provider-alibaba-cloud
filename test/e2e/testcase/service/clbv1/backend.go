package clbv1

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"time"

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

	ginkgo.Describe("clb service controller: backend", func() {
		ginkgo.AfterEach(func() {
			ginkgo.By("delete service")
			err := f.AfterEach()
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.Context("backend-label", func() {
			ginkgo.It("backend-label", func() {
				// label node
				node, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(node).NotTo(gomega.BeNil())
				err = f.Client.KubeClient.LabelNode(node.Name, client.NodeLabel, client.NodeLabel)
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					_ = f.Client.KubeClient.UnLabelNode(node.Name, client.NodeLabel)
				}()

				svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.BackendLabel): fmt.Sprintf("%s=%s", client.NodeLabel, client.NodeLabel),
				})
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("externalTrafficPolicy", func() {
			ginkgo.It("cluster -> local", func() {
				svc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := svc.DeepCopy()
				newsvc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
				newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.Network == options.Terway {
				ginkgo.It("local -> eni", func() {
					svc := f.Client.KubeClient.DefaultService()
					svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
					svc, err := f.Client.KubeClient.CreateService(svc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := svc.DeepCopy()
					newsvc.Annotations = make(map[string]string)
					newsvc.Annotations[annotation.BackendType] = model.ENIBackendType
					newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("cluster -> eni", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := svc.DeepCopy()
					newsvc.Annotations = make(map[string]string)
					newsvc.Annotations[annotation.BackendType] = model.ENIBackendType
					newsvc, err = f.Client.KubeClient.PatchService(svc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		ginkgo.Context("remove-unscheduled-backend", func() {
			ginkgo.It("remove-unscheduled-backend: off; node: unschedulable -> schedulable", func() {
				// unscheduled node
				node, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(node).NotTo(gomega.BeNil())
				err = f.Client.KubeClient.UnscheduledNode(node.Name)
				gomega.Expect(err).To(gomega.BeNil())

				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.RemoveUnscheduled): string(model.OffFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.Client.KubeClient.ScheduledNode(node.Name)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("remove-unscheduled-backend: on;  node: schedulable -> unschedulable", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
					annotation.Annotation(annotation.RemoveUnscheduled): string(model.OnFlag),
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				// unscheduled node
				node, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(node).NotTo(gomega.BeNil())
				err = f.Client.KubeClient.UnscheduledNode(node.Name)
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					_ = f.Client.KubeClient.ScheduledNode(node.Name)
				}()

				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
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

				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.Client.KubeClient.RemoveTaint(node.Name, taint)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})

		if options.TestConfig.Network == options.Terway {
			ginkgo.Context("backend-type", func() {
				ginkgo.It("backend-type: eni", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.BackendType: model.ENIBackendType,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("backend-type: eni -> ecs", func() {
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.BackendType: model.ENIBackendType,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Annotations[annotation.BackendType] = model.ECSBackendType
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})
		}

		if options.TestConfig.InternetLoadBalancerID != "" && options.TestConfig.VServerGroupID != "" {
			ginkgo.Context("vgroup-port", func() {
				ginkgo.It("vgroup-port: rsp-id:80", func() {
					vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID, 80)
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.VGroupPort):     vGroupPort,
						annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetLoadBalancerID,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				if options.TestConfig.VServerGroupID2 != "" {
					ginkgo.It("vgroup-port: rsp-id-1:80,rsp-id-2:443", func() {
						vGroupPort := fmt.Sprintf("%s:%d,%s:%d",
							options.TestConfig.VServerGroupID, 80, options.TestConfig.VServerGroupID2, 443)
						oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							annotation.Annotation(annotation.VGroupPort):     vGroupPort,
							annotation.Annotation(annotation.LoadBalancerId): options.TestConfig.InternetLoadBalancerID,
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
					ginkgo.It("vgroup-port: rsp-id-1:80 -> rsp-id-2:80", func() {
						vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID, 80)
						oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							annotation.Annotation(annotation.VGroupPort):       vGroupPort,
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
							annotation.Annotation(annotation.OverrideListener): "false",
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())

						newSvc := oldSvc.DeepCopy()
						newVGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID2, 80)
						newSvc.Annotations[annotation.Annotation(annotation.VGroupPort)] = newVGroupPort
						newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
						gomega.Expect(err).To(gomega.BeNil())

						err = f.ExpectLoadBalancerEqual(newSvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
				ginkgo.It("vgroup-port: not exist rsp-id", func() {
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.VGroupPort):       "rsp-xxx:80",
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "false",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				ginkgo.It("vgroup-port: rsp-id belongs to other slb", func() {
					vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID, 80)
					svc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.VGroupPort):       vGroupPort,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.IntranetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "false",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
			})

			ginkgo.Context("weight", func() {
				ginkgo.It("cluster mode: weight: 60 -> 80", func() {
					vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID, 80)
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.VGroupPort):       vGroupPort,
						annotation.Annotation(annotation.VGroupWeight):     "60",
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "false",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("local mode: weight: 60 -> 80", func() {
					svc := f.Client.KubeClient.DefaultService()
					vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID, 80)
					svc.Annotations = map[string]string{
						annotation.Annotation(annotation.VGroupPort):       vGroupPort,
						annotation.Annotation(annotation.VGroupWeight):     "60",
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "false",
					}
					svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
					oldSvc, err := f.Client.KubeClient.CreateService(svc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					newSvc := oldSvc.DeepCopy()
					newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
					newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				if options.TestConfig.Network == options.Terway {
					ginkgo.It("eni mode: weight: 60 -> 80", func() {
						vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID, 80)
						oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
							annotation.Annotation(annotation.VGroupPort):       vGroupPort,
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
							annotation.Annotation(annotation.OverrideListener): "false",
							annotation.Annotation(annotation.VGroupWeight):     "60",
							annotation.BackendType:                             model.ENIBackendType,
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())

						newSvc := oldSvc.DeepCopy()
						newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
						newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectLoadBalancerEqual(newSvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
					ginkgo.It("ecs mode; weight: nil -> 80", func() {
						vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID, 80)
						svc := f.Client.KubeClient.DefaultService()
						svc.Annotations = map[string]string{
							annotation.Annotation(annotation.VGroupPort):       vGroupPort,
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
							annotation.Annotation(annotation.OverrideListener): "false",
							annotation.BackendType:                             model.ECSBackendType,
						}
						svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
						oldSvc, err := f.Client.KubeClient.CreateService(svc)
						gomega.Expect(err).To(gomega.BeNil())

						err = f.ExpectLoadBalancerEqual(oldSvc)
						gomega.Expect(err).To(gomega.BeNil())

						newSvc := oldSvc.DeepCopy()
						newSvc.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
						newSvc, err = f.Client.KubeClient.PatchService(oldSvc, newSvc)
						gomega.Expect(err).To(gomega.BeNil())

						err = f.ExpectLoadBalancerEqual(newSvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
				ginkgo.It("weight: 0 -> 100", func() {
					vGroupPort := fmt.Sprintf("%s:%d", options.TestConfig.VServerGroupID, 80)
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.Annotation(annotation.VGroupPort):       vGroupPort,
						annotation.Annotation(annotation.VGroupWeight):     "0",
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "false",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
					gomega.Expect(err).To(gomega.BeNil())

					hundred := oldSvc.DeepCopy()
					hundred.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "100"
					hundred, err = f.Client.KubeClient.PatchService(oldSvc, hundred)
					gomega.Expect(err).To(gomega.BeNil())

					err = f.ExpectLoadBalancerEqual(hundred)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})
		}

		ginkgo.Context("special endpoints", func() {
			ginkgo.It("service with no selector", func() {
				svc, err := f.Client.KubeClient.CreateServiceWithoutSelector(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("endpoint with not exist node", func() {
				// only for ecs mode
				svc, err := f.Client.KubeClient.CreateServiceWithoutSelector(map[string]string{
					annotation.BackendType: model.ECSBackendType,
				})
				gomega.Expect(err).To(gomega.BeNil())
				_, err = f.Client.KubeClient.CreateEndpointsWithNotExistNode()
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("endpoint without node name", func() {
				// only for ecs mode
				svc, err := f.Client.KubeClient.CreateServiceWithoutSelector(map[string]string{
					annotation.BackendType: model.ECSBackendType,
				})
				gomega.Expect(err).To(gomega.BeNil())
				_, err = f.Client.KubeClient.CreateEndpointsWithoutNodeName()
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(svc)
				// if a endpoint does not have node name, fail
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("target port", func() {
			ginkgo.It("targetPort: 80 -> 81; ecs mode", func() {
				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
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

				err = f.ExpectLoadBalancerEqual(newSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
			if options.TestConfig.Network == options.Terway {
				ginkgo.It("targetPort: 80 -> 81; eni mode", func() {
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.BackendType: model.ENIBackendType,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
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

					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("targetPort: http; eni mode", func() {
					svc, err := f.Client.KubeClient.CreateServiceWithStringTargetPort(map[string]string{
						annotation.BackendType: model.ENIBackendType,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("targetPort: 80 -> http; eni mode", func() {
					oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(map[string]string{
						annotation.BackendType: model.ENIBackendType,
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectLoadBalancerEqual(oldSvc)
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

					err = f.ExpectLoadBalancerEqual(newSvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
				ginkgo.It("targetPort: http -> 80; eni mode", func() {
					svc := f.Client.KubeClient.DefaultService()
					svc.Annotations = map[string]string{
						annotation.BackendType: model.ENIBackendType,
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
					err = f.ExpectLoadBalancerEqual(svc)
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

					err = f.ExpectLoadBalancerEqual(newSvc)
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

				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
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

				oldSvc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectLoadBalancerEqual(oldSvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
}
