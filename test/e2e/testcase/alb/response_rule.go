package alb

import (
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/alb/common"
)

var (
	responseServiceName = "response-service"
	httpStatusCode      = "{\n    \"type\": \"ResponseStatusCode\",\n    \"responseStatusCodeConfig\": {\n      \"values\": [\n        \"400\"\n      ]\n    }\n  }"
	responseHeader      = "{\n    \"type\": \"ResponseHeader\",\n    \"responseHeaderConfig\": {\n      \"key\": \"headername\",\n      \"values\": [\n        \"headervalue1\"\n      ]\n   }\n  }"
)

var (
	removeHeader = "{\"type\":\"RemoveHeader\",\"RemoveHeaderConfig\":{\"key\":\"key\"}}"
)

var (
	httpStatusCodeSingleTest = "[" + httpStatusCode + "]"
	responseHeaderSingleTest = "[" + responseHeader + "]"

	removeHeaderSingleTest = "[" + removeHeader + "]"
	insertHeaderSingleTest = "[" + insertHeader + "]"
)

var (
	responseAllConditionsWithoutPath        = "[" + host + "," + cookie + "," + httpStatusCode + "," + header + "," + sourceIp + "," + querystring + "," + method + "]"
	responseAllConditionsWithoutHost        = "[" + path + "," + cookie + "," + header + "," + httpStatusCode + "," + sourceIp + "," + querystring + "," + method + "]"
	responseAllConditionsWithoutHeader      = "[" + path + "," + host + "," + cookie + "," + httpStatusCode + "," + sourceIp + "," + querystring + "," + method + "]"
	responseAllConditionsWithoutCookie      = "[" + path + "," + host + "," + header + "," + httpStatusCode + "," + sourceIp + "," + querystring + "," + method + "]"
	responseAllConditionsWithoutQuerystring = "[" + path + "," + host + "," + cookie + "," + httpStatusCode + "," + header + "," + sourceIp + "," + method + "]"
	responseAllConditionsWithoutMethod      = "[" + path + "," + host + "," + cookie + "," + httpStatusCode + "," + header + "," + sourceIp + "," + querystring + "]"
	responseAllConditionsWithoutSourceIp    = "[" + path + "," + host + "," + cookie + "," + httpStatusCode + "," + header + "," + querystring + "," + method + "]"
)

func RunResponseRuleTestCases(f *framework.Framework) {
	rule := common.Rule{}
	ingress := common.Ingress{}
	service := common.Service{}
	ginkgo.BeforeEach(func() {
		service.CreateDefaultService(f)
	})

	ginkgo.AfterEach(func() {
		ingress.DeleteIngress(f, ingress.DefaultIngress(f))
	})
	ginkgo.Describe("alb-ingress-controller: ingress", func() {
		ginkgo.Context("ingress create with response rule", func() {
			ginkgo.It("[alb][p0] ingress with httpStatusCode condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): httpStatusCodeSingleTest,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with httpStatusCode condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): httpStatusCodeSingleTest,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseHeader condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseHeaderSingleTest,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseHeader condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseHeaderSingleTest,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with allConditionsWithoutPath condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): allConditionsWithoutPath,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with allConditionsWithoutHost condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): allConditionsWithoutHost,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with allConditionsWithoutHeader condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): allConditionsWithoutHeader,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with allConditionsWithoutCookie condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): allConditionsWithoutCookie,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with allConditionsWithoutQuerystring condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): allConditionsWithoutQuerystring,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with allConditionsWithoutMethod condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): allConditionsWithoutMethod,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with allConditionsWithoutSourceIp condition and removeHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): allConditionsWithoutSourceIp,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    removeHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseAllConditionsWithoutPath condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseAllConditionsWithoutPath,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseAllConditionsWithoutHost condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseAllConditionsWithoutHost,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseAllConditionsWithoutHeader condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseAllConditionsWithoutHeader,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseAllConditionsWithoutCookie condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseAllConditionsWithoutCookie,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseAllConditionsWithoutQuerystring condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseAllConditionsWithoutQuerystring,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseAllConditionsWithoutMethod condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseAllConditionsWithoutMethod,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with responseAllConditionsWithoutSourceIp condition and insertHeader action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, responseServiceName): responseAllConditionsWithoutSourceIp,
					fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, responseServiceName):         "Response",
					fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, responseServiceName):    insertHeaderSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultUseAnnotationRule("", responseServiceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

		})
	})
}
