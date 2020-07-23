package test

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	alicloud "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/e2e/framework"
	"k8s.io/klog"
	"testing"
)

var _ = framework.Mark(
	"quick",
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestRandomCombination"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				// reset f.InitService if needed
			},
		)

		defer func() {
			err := f.Destroy()
			if err != nil {
				klog.Errorf("destroy error: %s", err.Error())
			}
		}()

		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}

		rand := []framework.Action{
			// set spec
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerSpec: "slb.s1.small",
						}
						return nil
					},
					Description: "mutate lb spec to slb.s1.small. function not ready yet",
				},
			),
			// set health check
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerHealthCheckInterval:       "10",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectTimeout: "3",
						}
						return nil
					},
					Description: "mutate health check",
				},
			),
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerProtocolPort: "http:80",
						}
						return nil
					},
					Description: "mutate lb 80 listener to http",
				},
			),
		}
		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewRandomAction(rand),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

var _ = framework.Mark(
	"quick",
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestIntranetCombination"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerAddressType:                   "intranet",
							alicloud.ServiceAnnotationLoadBalancerSLBNetworkType:                "vpc",
							alicloud.ServiceAnnotationLoadBalancerVswitch:                       framework.TestContext.VSwitchID,
							alicloud.ServiceAnnotationLoadBalancerSpec:                          "slb.s1.small",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckType:               "tcp",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectPort:        "80",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold:   "4",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold: "4",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckInterval:           "3",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectTimeout:     "100",
						},
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Port:       80,
								TargetPort: intstr.FromInt(80),
								Protocol:   v1.ProtocolTCP,
							},
						},
						Type:            v1.ServiceTypeLoadBalancer,
						SessionAffinity: v1.ServiceAffinityNone,
						Selector: map[string]string{
							"run": "nginx",
						},
					},
				}
			},
		)

		defer func() {
			err := f.Destroy()
			if err != nil {
				klog.Errorf("destroy error: %s", err.Error())
			}
		}()

		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}

		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerAddressType:                   "intranet",
					alicloud.ServiceAnnotationLoadBalancerSLBNetworkType:                "vpc",
					alicloud.ServiceAnnotationLoadBalancerVswitch:                       framework.TestContext.VSwitchID,
					alicloud.ServiceAnnotationLoadBalancerSpec:                          "slb.s2.small",
					alicloud.ServiceAnnotationLoadBalancerProtocolPort:                  "http:80",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckFlag:               "on",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckType:               "http",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckURI:                "/index.html",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectPort:        "80",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold:   "4",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold: "4",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckInterval:           "3",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckTimeout:            "10",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckDomain:             "192.168.0.3",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckHTTPCode:           "http_2xx",
				}
				return nil
			},
			Description: "mutate spec type to slb.s2.small and health check type to http type. add additional tags.",
		}
		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

var _ = framework.Mark(
	"quick",
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestClassicCombination"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerMasterZoneID:                  framework.TestContext.MasterZoneID,
							alicloud.ServiceAnnotationLoadBalancerSlaveZoneID:                   framework.TestContext.SlaveZoneID,
							alicloud.ServiceAnnotationLoadBalancerBandwidth:                     "45",
							alicloud.ServiceAnnotationLoadBalancerChargeType:                    "paybybandwidth",
							alicloud.ServiceAnnotationLoadBalancerProtocolPort:                  "http:80,https:443",
							alicloud.ServiceAnnotationLoadBalancerCertID:                        framework.TestContext.CertID,
							alicloud.ServiceAnnotationLoadBalancerPersistenceTimeout:            "1800",
							alicloud.ServiceAnnotationLoadBalancerScheduler:                     "wlc",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckFlag:               "on",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckType:               "http",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckURI:                "/index.html",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectPort:        "80",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold:   "4",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold: "4",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckInterval:           "3",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckTimeout:            "10",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckDomain:             "192.168.0.3",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckHTTPCode:           "http_2xx",
							alicloud.ServiceAnnotationLoadBalancerAdditionalTags:                "k1=v1,k2=v2",
						},
					},
					Spec: v1.ServiceSpec{
						ExternalTrafficPolicy: "Local",
						Ports: []v1.ServicePort{
							{
								Name:       "http",
								Port:       80,
								TargetPort: intstr.FromInt(80),
								Protocol:   v1.ProtocolTCP,
							},
							{
								Name:       "https",
								Port:       443,
								TargetPort: intstr.FromInt(443),
								Protocol:   v1.ProtocolTCP,
							},
						},
						Type:            v1.ServiceTypeLoadBalancer,
						SessionAffinity: v1.ServiceAffinityNone,
						Selector: map[string]string{
							"run": "nginx",
						},
					},
				}
			},
		)
		defer func() {
			if err := f.Destroy(); err != nil {
				klog.Errorf("destroy error: %s", err.Error())
			}
		}()
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}

		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerMasterZoneID:                  framework.TestContext.MasterZoneID,
					alicloud.ServiceAnnotationLoadBalancerSlaveZoneID:                   framework.TestContext.SlaveZoneID,
					alicloud.ServiceAnnotationLoadBalancerBandwidth:                     "88",
					alicloud.ServiceAnnotationLoadBalancerChargeType:                    "paybybandwidth",
					alicloud.ServiceAnnotationLoadBalancerProtocolPort:                  "http:80",
					alicloud.ServiceAnnotationLoadBalancerPersistenceTimeout:            "2000",
					alicloud.ServiceAnnotationLoadBalancerScheduler:                     "rr",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckType:               "tcp",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectPort:        "443",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold:   "5",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold: "5",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckInterval:           "5",
					alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectTimeout:     "8",
				}
				return nil
			},

			Description: "mutate bandwidth to 88, persistence timeout to 2000, scheduler to rr," +
				" health check type to tcp type and remove http:443 protocol port.",
		}

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

var _ = framework.Mark(
	"quick",
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestUserDefinedLBCombination"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerId:               framework.TestContext.LoadBalancerID,
							alicloud.ServiceAnnotationLoadBalancerOverrideListener: "true",
							alicloud.ServiceAnnotationLoadBalancerBackendLabel:     framework.TestContext.BackendLabel,
							alicloud.ServiceAnnotationLoadBalancerAclStatus:        "on",
							alicloud.ServiceAnnotationLoadBalancerAclID:            framework.TestContext.AclID,
							alicloud.ServiceAnnotationLoadBalancerAclType:          "white",
							alicloud.ServiceAnnotationLoadBalancerSessionStick:     "on",
							alicloud.ServiceAnnotationLoadBalancerSessionStickType: "insert",
							alicloud.ServiceAnnotationLoadBalancerCookieTimeout:    "1800",
							alicloud.ServiceAnnotationLoadBalancerProtocolPort:     "http:80",
						},
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:       "http",
								Port:       80,
								TargetPort: intstr.FromInt(80),
								Protocol:   v1.ProtocolTCP,
							},
							{
								Name:       "https",
								Port:       443,
								TargetPort: intstr.FromInt(443),
								Protocol:   v1.ProtocolTCP,
							},
						},
						Type:            v1.ServiceTypeLoadBalancer,
						SessionAffinity: v1.ServiceAffinityNone,
						Selector: map[string]string{
							"run": "nginx",
						},
					},
				}
			},
		)
		defer func() {
			if err := f.Destroy(); err != nil {
				klog.Errorf("destroy error: %s", err.Error())
			}
		}()
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}

		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerId:               framework.TestContext.LoadBalancerID,
					alicloud.ServiceAnnotationLoadBalancerOverrideListener: "true",
					alicloud.ServiceAnnotationLoadBalancerBackendLabel:     framework.TestContext.BackendLabel,
					alicloud.ServiceAnnotationLoadBalancerSessionStick:     "on",
					alicloud.ServiceAnnotationLoadBalancerSessionStickType: "server",
					alicloud.ServiceAnnotationLoadBalancerCookie:           "your_cookie-1",
					alicloud.ServiceAnnotationLoadBalancerProtocolPort:     "http:80,https:443",
					alicloud.ServiceAnnotationLoadBalancerCertID:           framework.TestContext.CertID,
					alicloud.ServiceAnnotationLoadBalancerForwardPort:      "80:443",
				}

				return nil
			},
			Description: "remove acl; mutate session type to server cookie;  add forward port http:80 https:443",
		}

		del := &framework.TestUnit{
			Description: "user defined lb deletion. expect exist",
			ExpectOK:    alicloud.ExpectExist,
		}

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(del),
		)
	},
)

var _ = framework.Mark(
	"quick",
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestUserDefinedLBCombinationWithHttpForward"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerId:               framework.TestContext.LoadBalancerID,
							alicloud.ServiceAnnotationLoadBalancerOverrideListener: "true",
							alicloud.ServiceAnnotationLoadBalancerBackendLabel:     framework.TestContext.BackendLabel,
							alicloud.ServiceAnnotationLoadBalancerAclStatus:        "on",
							alicloud.ServiceAnnotationLoadBalancerAclID:            framework.TestContext.AclID,
							alicloud.ServiceAnnotationLoadBalancerAclType:          "white",
							alicloud.ServiceAnnotationLoadBalancerSessionStick:     "on",
							alicloud.ServiceAnnotationLoadBalancerSessionStickType: "insert",
							alicloud.ServiceAnnotationLoadBalancerCookieTimeout:    "1800",
							alicloud.ServiceAnnotationLoadBalancerProtocolPort:     "http:80",
						},
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:       "https",
								Port:       443,
								TargetPort: intstr.FromInt(443),
								Protocol:   v1.ProtocolTCP,
							},
							{
								Name:       "http",
								Port:       80,
								TargetPort: intstr.FromInt(80),
								Protocol:   v1.ProtocolTCP,
							},
						},
						Type:            v1.ServiceTypeLoadBalancer,
						SessionAffinity: v1.ServiceAffinityNone,
						Selector: map[string]string{
							"run": "nginx",
						},
					},
				}
			},
		)
		defer func() {
			if err := f.Destroy(); err != nil {
				klog.Errorf("destroy error: %s", err.Error())
			}
		}()
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}

		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerId:               framework.TestContext.LoadBalancerID,
					alicloud.ServiceAnnotationLoadBalancerOverrideListener: "true",
					alicloud.ServiceAnnotationLoadBalancerBackendLabel:     framework.TestContext.BackendLabel,
					alicloud.ServiceAnnotationLoadBalancerSessionStick:     "on",
					alicloud.ServiceAnnotationLoadBalancerSessionStickType: "server",
					alicloud.ServiceAnnotationLoadBalancerCookie:           "your_cookie-1",
					alicloud.ServiceAnnotationLoadBalancerProtocolPort:     "http:80,https:443",
					alicloud.ServiceAnnotationLoadBalancerCertID:           framework.TestContext.CertID,
					alicloud.ServiceAnnotationLoadBalancerForwardPort:      "80:443",
				}

				return nil
			},
			Description: "remove acl; mutate session type to server cookie;  add forward port http:80 https:443",
		}

		del := &framework.TestUnit{
			Description: "user defined lb deletion. expect exist",
			ExpectOK:    alicloud.ExpectExist,
		}

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(del),
		)
	},
)
