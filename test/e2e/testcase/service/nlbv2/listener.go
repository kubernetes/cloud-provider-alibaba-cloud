package nlbv2

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
)

func RunListenerTestCases(f *framework.Framework) {
	ginkgo.Describe("nlb service controller: listener", func() {

		ginkgo.AfterEach(func() {
			ginkgo.By("delete service")
			err := f.AfterEach()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.Context("service port", func() {
			ginkgo.It("tcp port", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("udp port", func() {
				svc := f.Client.KubeClient.DefaultService()
				svc.Annotations = map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				}
				svc.Spec.Ports = []v1.ServicePort{
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
				}
				lbClass := helper.NLBClass
				svc.Spec.LoadBalancerClass = &lbClass

				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("port: 80 -> 81; protocol: tcp", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				//update listener port
				newsvc := oldsvc.DeepCopy()
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       81,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolTCP,
					},
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("port: 80; protocol: tcp -> udp", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				//update listener port
				newsvc := oldsvc.DeepCopy()
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolUDP,
					},
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("port: 80 -> 81; protocol: udp", func() {
				oldsvc := f.Client.KubeClient.DefaultService()
				oldsvc.Annotations = map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				}
				oldsvc.Spec.Ports = []v1.ServicePort{
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
				}
				lbClass := helper.NLBClass
				oldsvc.Spec.LoadBalancerClass = &lbClass

				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       81,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolUDP,
					},
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("port: 80; protocol: udp -> tcp", func() {
				oldsvc := f.Client.KubeClient.DefaultService()
				oldsvc.Annotations = map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				}
				oldsvc.Spec.Ports = []v1.ServicePort{
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
				}
				lbClass := helper.NLBClass
				oldsvc.Spec.LoadBalancerClass = &lbClass

				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolTCP,
					},
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("port: 80; mixed protocol: udp & tcp", func() {
				oldsvc := f.Client.KubeClient.DefaultService()
				oldsvc.Annotations = map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				}
				oldsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "tcp",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolUDP,
					},
					{
						Name:       "udp",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolTCP,
					},
				}
				lbClass := helper.NLBClass
				oldsvc.Spec.LoadBalancerClass = &lbClass

				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Spec.Ports = []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolTCP,
					},
				}
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			if options.TestConfig.NLBCertID != "" {
				ginkgo.It("tcpssl port", func() {
					svc := f.Client.KubeClient.DefaultService()
					svc.Annotations = map[string]string{
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					}
					lbClass := helper.NLBClass
					svc.Spec.LoadBalancerClass = &lbClass

					svc, err := f.Client.KubeClient.CreateService(svc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("port: 80; protocol: tcp -> tcpssl", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					//update listener port
					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "tcpssl:443"
					newsvc.Annotations[annotation.Annotation(annotation.CertID)] = options.TestConfig.NLBCertID
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("port: 80; protocol: tcpssl -> tcp", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					//update listener port
					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations = map[string]string{
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					}

					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("port: 443 -> 6443; protocol: tcpssl", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					//update listener port
					newsvc := oldsvc.DeepCopy()
					newsvc.Spec.Ports = []v1.ServicePort{
						{
							Name:       "https",
							Port:       6443,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
					}
					newsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "tcpssl:444"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			}
		})

		if options.TestConfig.NLBCertID != "" {
			ginkgo.Context("tcpssl cert-id", func() {
				ginkgo.It("cert-id: not exist cert-id", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.CertID):           "000000-cn-hangzhou",
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.OverrideListener): "true",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})
				if options.TestConfig.NLBCertID2 != "" {
					ginkgo.It("cert-id: NLBCertID -> NLBCertID2", func() {
						oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
							annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
							annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.OverrideListener): "true",
						})
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
						gomega.Expect(err).To(gomega.BeNil())

						//update certificate
						newsvc := oldsvc.DeepCopy()
						newsvc.Annotations[annotation.Annotation(annotation.CertID)] = options.TestConfig.NLBCertID2
						newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(newsvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
			})

			ginkgo.Context("tcpssl alpn", func() {
				ginkgo.It("alpn: on", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
						annotation.Annotation(annotation.OverrideListener): "true",
						annotation.Annotation(annotation.AlpnEnabled):      string(model.OnFlag),
						annotation.Annotation(annotation.AlpnPolicy):       "HTTP1Only",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("alpn: off -> on with alpn-policy", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
						annotation.Annotation(annotation.OverrideListener): "true",
						annotation.Annotation(annotation.AlpnEnabled):      string(model.OffFlag),
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.AlpnEnabled)] = string(model.OnFlag)
					newsvc.Annotations[annotation.Annotation(annotation.AlpnPolicy)] = "HTTP1Only"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("alpn: off -> on without alpn-policy", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
						annotation.Annotation(annotation.OverrideListener): "true",
						annotation.Annotation(annotation.AlpnEnabled):      string(model.OffFlag),
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.AlpnEnabled)] = string(model.OnFlag)
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("alpn: on -> off", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
						annotation.Annotation(annotation.OverrideListener): "true",
						annotation.Annotation(annotation.AlpnEnabled):      string(model.OnFlag),
						annotation.Annotation(annotation.AlpnPolicy):       "HTTP1Only",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.AlpnEnabled)] = string(model.OffFlag)
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("alpn: HTTP1Only -> HTTP2Only", func() {
					oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
						annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
						annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
						annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
						annotation.Annotation(annotation.OverrideListener): "true",
						annotation.Annotation(annotation.AlpnEnabled):      string(model.OnFlag),
						annotation.Annotation(annotation.AlpnPolicy):       "HTTP1Only",
					})
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
					gomega.Expect(err).To(gomega.BeNil())

					newsvc := oldsvc.DeepCopy()
					newsvc.Annotations[annotation.Annotation(annotation.AlpnPolicy)] = "HTTP2Only"
					newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
					gomega.Expect(err).To(gomega.BeNil())
					err = f.ExpectNetworkLoadBalancerEqual(newsvc)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})

			if options.TestConfig.NLBCACertID != "" {
				ginkgo.Context("tcpssl cacert", func() {
					ginkgo.It("enable cacert", func() {
						svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
							annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
							annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
							annotation.Annotation(annotation.CaCert):           string(model.OnFlag),
							annotation.Annotation(annotation.CaCertID):         options.TestConfig.NLBCACertID,
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.OverrideListener): "true",
						})

						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(svc)
						gomega.Expect(err).To(gomega.BeNil())
					})

					ginkgo.It("cacert off -> on", func() {
						oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
							annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
							annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.OverrideListener): "true",
						})

						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
						gomega.Expect(err).To(gomega.BeNil())

						newsvc := oldsvc.DeepCopy()
						newsvc.Annotations[annotation.Annotation(annotation.CaCert)] = string(model.OnFlag)
						newsvc.Annotations[annotation.Annotation(annotation.CaCertID)] = options.TestConfig.NLBCACertID
						newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(newsvc)
						gomega.Expect(err).To(gomega.BeNil())
					})

					ginkgo.It("cacert on -> off", func() {
						oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
							annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
							annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
							annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
							annotation.Annotation(annotation.CaCert):           string(model.OnFlag),
							annotation.Annotation(annotation.CaCertID):         options.TestConfig.NLBCACertID,
							annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
							annotation.Annotation(annotation.OverrideListener): "true",
						})

						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
						gomega.Expect(err).To(gomega.BeNil())

						newsvc := oldsvc.DeepCopy()
						delete(newsvc.Annotations, annotation.Annotation(annotation.CaCert))
						delete(newsvc.Annotations, annotation.Annotation(annotation.CaCertID))
						newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
						gomega.Expect(err).To(gomega.BeNil())
						err = f.ExpectNetworkLoadBalancerEqual(newsvc)
						gomega.Expect(err).To(gomega.BeNil())
					})
				})
			}
		}

		ginkgo.Context("tls cipher policy", func() {
			ginkgo.It("set another tls cipher policy", func() {
				svc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
					annotation.Annotation(annotation.TLSCipherPolicy):  "tls_cipher_policy_1_2_strict_with_1_3",
					annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("change tls cipher policy", func() {
				oldsvc, err := f.Client.KubeClient.CreateNLBServiceByAnno(map[string]string{
					annotation.Annotation(annotation.ZoneMaps):         options.TestConfig.NLBZoneMaps,
					annotation.Annotation(annotation.ProtocolPort):     "tcpssl:443",
					annotation.Annotation(annotation.CertID):           options.TestConfig.NLBCertID,
					annotation.Annotation(annotation.LoadBalancerId):   options.TestConfig.InternetNetworkLoadBalancerID,
					annotation.Annotation(annotation.OverrideListener): "true",
				})
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.TLSCipherPolicy] = "tls_cipher_policy_1_2_strict_with_1_3"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
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
				Protocol:   v1.ProtocolUDP,
			},
		}

		if options.TestConfig.NLBCertID != "" {
			testsvc.Spec.Ports = append(testsvc.Spec.Ports, v1.ServicePort{
				Name:       "tcpssl",
				Port:       443,
				TargetPort: intstr.FromInt(80),
				Protocol:   v1.ProtocolTCP,
			})
			testsvc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "tcpssl:443"
			testsvc.Annotations[annotation.Annotation(annotation.CertID)] = options.TestConfig.NLBCertID
		}
		lbClass := helper.NLBClass
		testsvc.Spec.LoadBalancerClass = &lbClass

		ginkgo.Context("proxy protocol", func() {
			ginkgo.It("proxy-protocol on", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("proxy-protocol off -> on", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("proxy-protocol on -> off", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("cps", func() {
			ginkgo.It("cps on", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.Cps)] = "100"
				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("cps off -> 100", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.Cps)] = "100"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("idle timeout", func() {
			ginkgo.It("idle-timeout 60", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.IdleTimeout)] = "60"
				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("idle timeout default -> 60", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.IdleTimeout)] = "60"
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})

		ginkgo.Context("ppv2 privatelink", func() {
			ginkgo.It("all features on", func() {
				svc := testsvc.DeepCopy()
				svc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpIdEnabled)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpsIdEnabled)] = string(model.OnFlag)
				svc.Annotations[annotation.Annotation(annotation.Ppv2VpcIdEnabled)] = string(model.OnFlag)
				svc, err := f.Client.KubeClient.CreateService(svc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(svc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("all features off -> on", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpIdEnabled)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpsIdEnabled)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2VpcIdEnabled)] = string(model.OnFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("only proxy protocol on -> all features on", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpIdEnabled)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpsIdEnabled)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2VpcIdEnabled)] = string(model.OnFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("all features on -> only proxy protocol on ", func() {
				oldsvc := testsvc.DeepCopy()
				oldsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				oldsvc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpIdEnabled)] = string(model.OnFlag)
				oldsvc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpsIdEnabled)] = string(model.OnFlag)
				oldsvc.Annotations[annotation.Annotation(annotation.Ppv2VpcIdEnabled)] = string(model.OnFlag)
				oldsvc, err := f.Client.KubeClient.CreateService(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(oldsvc)
				gomega.Expect(err).To(gomega.BeNil())

				newsvc := oldsvc.DeepCopy()
				newsvc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = string(model.OnFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpIdEnabled)] = string(model.OffFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2PrivateLinkEpsIdEnabled)] = string(model.OffFlag)
				newsvc.Annotations[annotation.Annotation(annotation.Ppv2VpcIdEnabled)] = string(model.OffFlag)
				newsvc, err = f.Client.KubeClient.PatchService(oldsvc, newsvc)
				gomega.Expect(err).To(gomega.BeNil())
				err = f.ExpectNetworkLoadBalancerEqual(newsvc)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
}
