package test

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	alicloud "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/e2e/framework"
	"testing"
)

var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestRandomCombinationCase"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				// reset f.InitService if needed
			},
		)
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}
		defer f.Destroy()
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
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestRandomCombinationVPCCase"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerAddressType:    "intranet",
							alicloud.ServiceAnnotationLoadBalancerSLBNetworkType: "vpc",
							alicloud.ServiceAnnotationLoadBalancerVswitch:        framework.TestContext.VSwitchID,
							alicloud.ServiceAnnotationLoadBalancerProtocolPort:   "http:80",
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
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}
		defer f.Destroy()
		rand := []framework.Action{
			// set Access control
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerAclStatus: "on",
							alicloud.ServiceAnnotationLoadBalancerAclID:     framework.TestContext.AclID,
							alicloud.ServiceAnnotationLoadBalancerAclType:   "white",
						}
						return nil
					},
					Description: "set access control",
				},
			),
			// set forward port
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerForwardPort: "81",
						}
						return nil
					},
					Description: "set lb scheduler",
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
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestRandomCombinationReusingLBCase"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerMasterZoneID:       framework.TestContext.MasterZoneID,
							alicloud.ServiceAnnotationLoadBalancerSlaveZoneID:        framework.TestContext.SlaveZoneID,
							alicloud.ServiceAnnotationLoadBalancerRegion:             framework.TestContext.MasterZoneID,
							alicloud.ServiceAnnotationLoadBalancerBackendLabel:       framework.TestContext.BackendLabel,
							alicloud.ServiceAnnotationLoadBalancerIPVersion:          "ipv6",
							alicloud.ServiceAnnotationLoadBalancerBandwidth:          "45",
							alicloud.ServiceAnnotationLoadBalancerChargeType:         "paybybandwidth",
							alicloud.ServiceAnnotationLoadBalancerPersistenceTimeout: "1200",
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
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}
		defer f.Destroy()
		rand := []framework.Action{
			// set protocol port
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerBackendLabel: framework.TestContext.BackendLabel,
							alicloud.ServiceAnnotationLoadBalancerProtocolPort: "https:443",
							alicloud.ServiceAnnotationLoadBalancerCertID:       framework.TestContext.CertID,
						}
						return nil
					},
					Description: "add protocol port https:443",
				},
			),
			//set band width
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerMasterZoneID:       framework.TestContext.MasterZoneID,
							alicloud.ServiceAnnotationLoadBalancerSlaveZoneID:        framework.TestContext.SlaveZoneID,
							alicloud.ServiceAnnotationLoadBalancerRegion:             framework.TestContext.MasterZoneID,
							alicloud.ServiceAnnotationLoadBalancerBackendLabel:       framework.TestContext.BackendLabel,
							alicloud.ServiceAnnotationLoadBalancerIPVersion:          "ipv6",
							alicloud.ServiceAnnotationLoadBalancerBandwidth:          "88",
							alicloud.ServiceAnnotationLoadBalancerChargeType:         "paybybandwidth",
							alicloud.ServiceAnnotationLoadBalancerPersistenceTimeout: "1200",
						}
						return nil
					},
					Description: "mutate charge type: paybybandwidth",
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
