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

// 0:test basic ensure LB
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestBasicEnsureLoadBalancer"
				f.Test = t
				f.Client = framework.NewClientOrDie()
			},
		)
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}
		defer f.Destroy()

		return f.RunDefaultTest(
			framework.NewDefaultAction(
				&framework.TestUnit{Description: "default init action"},
			),
			framework.NewDeleteAction(
				&framework.TestUnit{Description: "default delete action"},
			),
		)
	},
)

// 1:test LB with vpc
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestAddressType"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerAddressType: "intranet",
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

		spec := &framework.TestUnit{
			ExpectOK: alicloud.ExpectAddressTypeNotEqual,
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerAddressType: "internet",
				}
				return nil
			},
			Description: "change address type from internet to intranet",
		}
		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 2:test LB protocol type
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestProtocolPort"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerProtocolPort: "http:80",
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
		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerProtocolPort: "https:80",
					alicloud.ServiceAnnotationLoadBalancerCertID:       framework.TestContext.CertID,
				}
				return nil
			},
			Description: "change protocol from http to https. Note that the ports need to be same.",
		}
		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 3:test mutate LB Spec
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestMutateSpec"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				// set f.InitService if needed
			},
		)
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}
		defer f.Destroy()
		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerSpec: "slb.s1.small",
				}
				return nil
			},
			Description: "mutate loadbalancer spec to slb.s1.small",
		}

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete action"}),
		)
	},
)

// 4:test for reusing user defined LB
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestUserDefinedLoadBalancer"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerId:               framework.TestContext.LoadBalancerID,
							alicloud.ServiceAnnotationLoadBalancerOverrideListener: "true",
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

		del := &framework.TestUnit{
			Description: "user defined lb deletion. expect exist",
			ExpectOK:    alicloud.ExpectExist,
		}
		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init action"}),
			framework.NewDeleteAction(del),
		)
	},
)

// 5:test TCP session sticky
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestTCPSessionSticky"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerPersistenceTimeout: "1800",
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

		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerPersistenceTimeout: "2400",
				}
				return nil
			},
			Description: "mutate session sticky to 2400",
		}

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 6:test for changing LB scheduler
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestSchedulerCase"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				// set f.InitService if needed
			},
		)
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}
		defer f.Destroy()
		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerScheduler: "wlc",
				}
				return nil
			},
			Description: "mutate loadbalancer scheduler to wlc",
		}

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete action"}),
		)
	},
)

// 7:test master&slave zone
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestMasterSlaveZone"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerMasterZoneID: framework.TestContext.MasterZoneID,
							alicloud.ServiceAnnotationLoadBalancerSlaveZoneID:  framework.TestContext.SlaveZoneID,
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

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 8:test LB region
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestLoadBalancerRegion"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerRegion: framework.TestContext.MasterZoneID,
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

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 9:test health check (TCP type)
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestHealthCheck"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerHealthCheckFlag:               "on",
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
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}
		defer f.Destroy()

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 10:test health check (HTTP type)
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestHTTPHealthCheck"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerHealthCheckFlag:               "on",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckType:               "http",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckURI:                "/index.html",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectPort:        "80",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold:   "4",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold: "4",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckInterval:           "3",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckTimeout:            "10",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckDomain:             "192.168.0.85",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckHTTPCode:           "http_2xx",
							alicloud.ServiceAnnotationLoadBalancerProtocolPort:                  "http:80",
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

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 11:test backend label
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestBackendLabel"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerBackendLabel: framework.TestContext.BackendLabel,
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

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 12:test network type
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestNetworkType"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerSLBNetworkType: "vpc",
							alicloud.ServiceAnnotationLoadBalancerAddressType:    "intranet",
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

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 13:test Access control
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestACL"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerAclStatus: "on",
							alicloud.ServiceAnnotationLoadBalancerAclID:     framework.TestContext.AclID,
							alicloud.ServiceAnnotationLoadBalancerAclType:   "white",
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

		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerAclStatus: "off",
				}
				return nil
			},
			Description: "disable acl",
		}

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 14:test forward port
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestForwardPort"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerForwardPort: "81",
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

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 15:test IP version
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestForwardPort"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerIPVersion: "ipv6",
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

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 16:test VSwtich
// Only the SLB of the intranet needs vswitchid
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestVSwitch"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerVswitch:     framework.TestContext.VSwitchID,
							alicloud.ServiceAnnotationLoadBalancerAddressType: "intranet",
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

		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

// 17:test PayByBandWidth
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestPayByBandWidth"
				f.Test = t
				f.Client = framework.NewClientOrDie()
				f.InitService = &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "basic-service",
						Namespace: framework.NameSpace,
						Annotations: map[string]string{
							alicloud.ServiceAnnotationLoadBalancerBandwidth:  "45",
							alicloud.ServiceAnnotationLoadBalancerChargeType: "paybybandwidth",
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
		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					alicloud.ServiceAnnotationLoadBalancerBandwidth:  "88",
					alicloud.ServiceAnnotationLoadBalancerChargeType: "paybybandwidth",
				}
				return nil
			},
			Description: "mutate loadbalancer bandwidth to 88",
		}
		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)
