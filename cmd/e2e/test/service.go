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
							alicloud.ServiceAnnotationLoadBalancerId:               "lb-2ze6jp9vemd1gvj9ku83e",
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

var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				f.Desribe = "TestCombinationCase"
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
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

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
					Description: "mutate lb spec to slb.s2.small. function not ready yet",
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
