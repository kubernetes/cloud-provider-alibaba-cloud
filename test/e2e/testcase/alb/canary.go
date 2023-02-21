package alb

import (
	"github.com/onsi/ginkgo/v2"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/alb/common"
)

func RunCanaryTestCases(f *framework.Framework) {
	ingress := common.Ingress{}
	service := common.Service{}
	ginkgo.BeforeEach(func() {
		service.CreateDefaultService(f)
	})

	ginkgo.AfterEach(func() {
		ingress.DeleteIngress(f, ingress.DefaultIngress(f))
	})

	ginkgo.It("[alb][p0] ingress with customize condition querystring and forward action", func() {
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/canary":        "true",
			"alb.ingress.kubernetes.io/canary-weight": "50",
		}
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-weight"))
		ingress.WaitCreateIngress(f, ing, true)
	})

	ginkgo.It("[alb][p0] ingress with customize condition querystring and forward action", func() {
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/canary":                 "true",
			"alb.ingress.kubernetes.io/canary-by-header":       "location",
			"alb.ingress.kubernetes.io/canary-by-header-value": "hz",
		}
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header"))
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header-value"))
		ingress.WaitCreateIngress(f, ing, true)
	})

	ginkgo.It("[alb][p0] ingress with customize condition querystring and forward action", func() {
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/canary":                 "true",
			"alb.ingress.kubernetes.io/canary-by-header":       "location",
			"alb.ingress.kubernetes.io/canary-by-header-value": "hz",
			"alb.ingress.kubernetes.io/canary-weight":          "50",
		}
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header"))
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header-value"))
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-weight"))
		ingress.WaitCreateIngress(f, ing, true)
	})
}
