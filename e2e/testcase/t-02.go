package testcase

import (
	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
)

var _ = ginkgo.Describe("TestAddressType", func() {

	f := framework.NewFrameWork(
		func(f *framework.FrameWorkE2E) {
			f.Desribe = "TestAddressType"
		},
	)
	// setup nginx once for all test
	f.SetUp()

	defer f.CleanUp()
	// uncomment if u want to run initialize at each iteration
	// ginkgo.BeforeEach(f.BeforeEach)

	ginkgo.It("test change loadbalancer address type", func() {
		base := framework.Defaultsvc
		base.SetAnnotations(
			map[string]string{
				//alicloud.ServiceAnnotationLoadBalancerAddressType: "intranet",
			},
		)
		spec := &framework.TestUnit{
			Mutator: func(service *v1.Service) error {
				service.Annotations = map[string]string{
					//alicloud.ServiceAnnotationLoadBalancerAddressType: "internet",
				}
				return nil
			},
			ExpectOK:    ExpectAddressTypeOK,
			Description: "change address type from internet to intranet",
		}
		err := framework.RunActions(f,
			framework.NewDefaultAction(&framework.TestUnit{Service: base, Description: "default init"}),
			framework.NewDefaultAction(spec),
		)
		framework.ExpectNoError(err)
		framework.Logf("test address change finished%v", spec)
	})

	ginkgo.It("test2 create service Creating pay per bandwidth load balancing", func() {
		// TODO:
		// some other address test
	})

})

func ExpectAddressTypeOK(m *framework.Expectation) error {
	// TODO:
	// implement me
	// compare svc and slb configuration difference
	//m.E2E.Cloud.ListSLB()

	return nil
}
