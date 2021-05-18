package testcase

import (
	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	req "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
)

var _ = ginkgo.Describe("TestSpecType", func() {

	f := framework.NewNamedFrameWork("TestSpecTypeChange")
	// setup nginx once for all test
	//f.SetUp()
	//defer f.CleanUp()

	// uncomment if u want to run initialize at each iteration
	ginkgo.AfterSuite(f.AfterEach)
	ginkgo.BeforeSuite(f.BeforeEach)

	ginkgo.It("test change loadbalancer spec type", func() {
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.Spec): "slb.s1.small",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "init default address type svc", ExpectOK: ExpectSpecTypeOK,
		}

		spec := &framework.TestUnit{
			Service: base,
			Mutator: func(svc *v1.Service) error {
				svc.SetAnnotations(map[string]string{
					req.Annotation(req.Spec): "slb.s2.small",
				})
				return nil
			},
			ExpectOK:    ExpectSpecTypeOK,
			Description: "change spec type from slb.s1.small to slb.s2.small",
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
		)
		framework.ExpectNoError(err)
		framework.Logf("test address change finished")
	})

	ginkgo.It("test2 create service Creating pay per bandwidth load balancing", func() {
		// TODO:
		// some other address test
	})

})

func ExpectSpecTypeOK(m *framework.Expectation) (bool, error) {
	// TODO:
	// implement me
	// compare svc and slb configuration difference
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(m.Case.Service),
	}

	err := m.E2E.ModelBuilder.LoadBalancerMgr.Find(m.Case.ReqCtx, lbMdl)
	if err != nil {
		// TODO if err, need retry
		return false, framework.NewErrorRetry(err)
	}
	if lbMdl.LoadBalancerAttribute.LoadBalancerId == "" {
		return false, framework.NewErrorRetry(err)
	}

	spec := m.Case.ReqCtx.Anno.Get(req.Spec)

	klog.Infof("expect spec type ok: %s", spec)
	if string(lbMdl.LoadBalancerAttribute.LoadBalancerSpec) != spec {
		klog.Info("expected: waiting slb spec change")
		return false, framework.NewErrorRetry(err)
	}
	return true, nil
}
