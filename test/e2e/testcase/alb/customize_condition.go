package alb

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/client"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/alb/common"
)

var (
	serviceName = client.Service
	path        = "{\n    \"type\": \"Path\",\n    \"pathConfig\": {\n      \"values\": [\n        \"/anno/path\"\n      ]\n    }\n  }"
	host        = "{\n      \"type\": \"Host\",\n      \"hostConfig\": {\n        \"values\": [\n          \"anno.example.com\"\n        ]\n      }\n  }"
	sourceIp    = "{\n    \"type\": \"SourceIp\",\n    \"sourceIpConfig\": {\n      \"values\": [\n        \"172.16.0.0/16\"\n      ]\n    }\n  }"
	cookie      = "{\n    \"type\": \"Cookie\",\n    \"cookieConfig\": {\n      \"values\": [\n        {\n           \"key\":\"cookiekey1\",\n        \t\t\"value\":\"cookievalue2\"\n        }\n      ]\n  \t}\n  }"
	header      = "{\n    \"type\": \"Header\",\n    \"headerConfig\": {\n      \"key\": \"headername\",\n      \"values\": [\n        \"headervalue1\"\n      ]\n   }\n  }"
	method      = "{\n    \"type\": \"Method\",\n    \"methodConfig\": {\n      \"values\": [\n        \"GET\",\n        \"HEAD\"\n      ]\n    }\n  }"
	querystring = "{\n    \"type\": \"QueryString\",\n    \"queryStringConfig\": {\n      \"values\": [\n        {\n           \"key\":\"querystringkey1\",\n        \t\t\"value\":\"querystringvalue2\"\n        }\n      ]\n  \t}\n  }"
)

var (
	pathSingleTest        = "[" + path + "]"
	hostSingleTest        = "[" + host + "]"
	headerSingleTest      = "[" + header + "]"
	cookieSingleTest      = "[" + cookie + "]"
	sourceIpSingleTest    = "[" + sourceIp + "]"
	querystringSingleTest = "[" + querystring + "]"
	methodSingleTest      = "[" + method + "]"
	pathHost              = "[" + path + "," + host + "]"
	pathSourceIp          = "[" + path + "," + sourceIp + "]"
	pathCookie            = "[" + path + "," + cookie + "]"
	pathHeader            = "[" + path + "," + header + "]"
	pathMethod            = "[" + path + "," + method + "]"
	pathQuerystring       = "[" + path + "," + querystring + "]"

	allConditionsWithoutPath        = "[" + host + "," + cookie + "," + header + "," + sourceIp + "," + querystring + "," + method + "]"
	allConditionsWithoutHost        = "[" + path + "," + cookie + "," + header + "," + sourceIp + "," + querystring + "," + method + "]"
	allConditionsWithoutHeader      = "[" + path + "," + host + "," + cookie + "," + sourceIp + "," + querystring + "," + method + "]"
	allConditionsWithoutCookie      = "[" + path + "," + host + "," + header + "," + sourceIp + "," + querystring + "," + method + "]"
	allConditionsWithoutQuerystring = "[" + path + "," + host + "," + cookie + "," + header + "," + sourceIp + "," + method + "]"
	allConditionsWithoutMethod      = "[" + path + "," + host + "," + cookie + "," + header + "," + sourceIp + "," + querystring + "]"
	allConditionsWithoutSourceIp    = "[" + path + "," + host + "," + cookie + "," + header + "," + querystring + "," + method + "]"

	allConditions = "[" + path + "," + host + "," + cookie + "," + header + "," + sourceIp + "," + querystring + "," + method + "]"
)

func RunCustomizeConditionTestCases(f *framework.Framework) {
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
		ginkgo.Context("ingress create with customize condition", func() {
			ginkgo.It("[alb][p0] ingress with customize condition path and forward action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition host and forward action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): hostSingleTest,
				}

				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition sourceIp and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): sourceIpSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition cookie and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): cookieSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition header and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): headerSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition method and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): methodSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition querystring and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): querystringSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition path and forward action with ingress-host", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "www.test-host.com", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition host and forward action with ingress-host", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): hostSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "www.test-host.com", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition sourceIp and forward action with ingress-host", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): sourceIpSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "www.test-host.com", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition cookie and forward action with ingress-host", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): cookieSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "www.test-host.com", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition header and forward action with ingress-host", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): headerSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "www.test-host.com", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition method and forward action with ingress-host", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): methodSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "www.test-host.com", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition querystring and forward action with ingress-host", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): querystringSingleTest,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "www.test-host.com", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition querystring and fixed-response/redirect/insert-header action ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): querystringSingleTest,
					"alb.ingress.kubernetes.io/actions.response-503":                         "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":                             "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header":                        "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition method and fixed-response/redirect/insert-header action ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): methodSingleTest,
					"alb.ingress.kubernetes.io/actions.response-503":                         "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":                             "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header":                        "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition cookie and fixed-response/redirect/insert-header action ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): cookieSingleTest,
					"alb.ingress.kubernetes.io/actions.response-503":                         "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":                             "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header":                        "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition path and fixed-response/redirect/insert-header action ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathSingleTest,
					"alb.ingress.kubernetes.io/actions.response-503":                         "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":                             "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header":                        "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition host and fixed-response/redirect/insert-header action ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): hostSingleTest,
					"alb.ingress.kubernetes.io/actions.response-503":                         "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":                             "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header":                        "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition header and fixed-response/redirect/insert-header action ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): headerSingleTest,
					"alb.ingress.kubernetes.io/actions.response-503":                         "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":                             "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header":                        "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition sourceIp and fixed-response/redirect/insert-header action ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): sourceIpSingleTest,
					"alb.ingress.kubernetes.io/actions.response-503":                         "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":                             "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header":                        "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition path and forward/rewrite-target/traffic-limit action ", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathSingleTest,
					"alb.ingress.kubernetes.io/rewrite-target":                               "/path/${2}",
					"alb.ingress.kubernetes.io/traffic-limit-qps":                            "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition host and forward/rewrite-target/traffic-limit action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): hostSingleTest,
					"alb.ingress.kubernetes.io/rewrite-target":                               "/path/${2}",
					"alb.ingress.kubernetes.io/traffic-limit-qps":                            "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition sourceIp and forward/rewrite-target/traffic-limit action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): sourceIpSingleTest,
					"alb.ingress.kubernetes.io/rewrite-target":                               "/path/${2}",
					"alb.ingress.kubernetes.io/traffic-limit-qps":                            "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition cookie and forward/rewrite-target/traffic-limit action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): cookieSingleTest,
					"alb.ingress.kubernetes.io/rewrite-target":                               "/path/${2}",
					"alb.ingress.kubernetes.io/traffic-limit-qps":                            "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition header and forward/rewrite-target/traffic-limit action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): headerSingleTest,
					"alb.ingress.kubernetes.io/rewrite-target":                               "/path/${2}",
					"alb.ingress.kubernetes.io/traffic-limit-qps":                            "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition method and forward/rewrite-target/traffic-limit action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): methodSingleTest,
					"alb.ingress.kubernetes.io/rewrite-target":                               "/path/${2}",
					"alb.ingress.kubernetes.io/traffic-limit-qps":                            "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition querystring and forward/rewrite-target/traffic-limit action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): querystringSingleTest,
					"alb.ingress.kubernetes.io/rewrite-target":                               "/path/${2}",
					"alb.ingress.kubernetes.io/traffic-limit-qps":                            "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition path and forward action with canary header", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathSingleTest,
					"alb.ingress.kubernetes.io/canary":                                       "true",
					"alb.ingress.kubernetes.io/canary-by-header":                             "location",
					"alb.ingress.kubernetes.io/canary-by-header-value":                       "hz",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header-value"))
				ingress.WaitCreateIngress(f, ing, false)

			})

			ginkgo.It("[alb][p0] ingress with customize condition host and forward action with canary header", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): hostSingleTest,
					"alb.ingress.kubernetes.io/canary":                                       "true",
					"alb.ingress.kubernetes.io/canary-by-header":                             "location",
					"alb.ingress.kubernetes.io/canary-by-header-value":                       "hz",
					"alb.ingress.kubernetes.io/canary-weight":                                "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header-value"))
				ingress.WaitCreateIngress(f, ing, false)
			})

			ginkgo.It("[alb][p0] ingress with customize condition sourceIp and forward action with canary header", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): sourceIpSingleTest,
					"alb.ingress.kubernetes.io/canary":                                       "true",
					"alb.ingress.kubernetes.io/canary-by-header":                             "location",
					"alb.ingress.kubernetes.io/canary-by-header-value":                       "hz",
					"alb.ingress.kubernetes.io/canary-weight":                                "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-by-header-value"))
				ingress.WaitCreateIngress(f, ing, false)
			})

			ginkgo.It("[alb][p0] ingress with customize condition cookie and forward action with canary weight", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): cookieSingleTest,
					"alb.ingress.kubernetes.io/canary":                                       "true",
					"alb.ingress.kubernetes.io/canary-weight":                                "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-weight"))
				ingress.WaitCreateIngress(f, ing, false)
			})

			ginkgo.It("[alb][p0] ingress with customize condition header and forward action with canary weight", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): headerSingleTest,
					"alb.ingress.kubernetes.io/canary":                                       "true",
					"alb.ingress.kubernetes.io/canary-weight":                                "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-weight"))
				ingress.WaitCreateIngress(f, ing, false)
			})

			ginkgo.It("[alb][p0] ingress with customize condition method and forward action with canary weight", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): methodSingleTest,
					"alb.ingress.kubernetes.io/canary":                                       "true",
					"alb.ingress.kubernetes.io/canary-weight":                                "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-weight"))
				ingress.WaitCreateIngress(f, ing, false)
			})

			ginkgo.It("[alb][p0] ingress with customize condition querystring and forward action with canary weight", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): querystringSingleTest,
					"alb.ingress.kubernetes.io/canary":                                       "true",
					"alb.ingress.kubernetes.io/canary-weight":                                "50",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "canary-weight"))
				ingress.WaitCreateIngress(f, ing, false)
			})

			ginkgo.It("[alb][p0] ingress with customize condition pathHost and forward action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathHost,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition pathCookie and forward action", func() {
				ing := ingress.DefaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathCookie,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition pathSourceIp and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathSourceIp,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition pathHeader and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathHeader,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition pathMethod and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathMethod,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with customize condition pathQuerystring and forward action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): pathQuerystring,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with forwardGroup final action and all customize conditions without cookie ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditionsWithoutCookie,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with forwardGroup final action and all customize conditions without path", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditionsWithoutPath,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with forwardGroup final action and all customize conditions without host", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditionsWithoutHost,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with forwardGroup final action and all customize conditions without header", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditionsWithoutHeader,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] iexitngress with forwardGroup final action and all customize conditions without querystring", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditionsWithoutQuerystring,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with forwardGroup final action and all customize conditions without method", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditionsWithoutMethod,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with forwardGroup final action and all customize conditions without sourceIp ", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditionsWithoutSourceIp,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with all customize conditions and forwardGroup final action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditions,
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ingress.WaitCreateIngress(f, ing, true)
			})

			ginkgo.It("[alb][p0] ingress with all customize conditions and FixedResponse/Redirect/InsertHeader final action", func() {
				ing := defaultIngress(f)
				ing.Annotations = map[string]string{
					fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditions,
					"alb.ingress.kubernetes.io/actions.response-503":                         "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
					"alb.ingress.kubernetes.io/actions.redirect":                             "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
					"alb.ingress.kubernetes.io/actions.insert-header":                        "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
				}
				ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
				ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
				ingress.WaitCreateIngress(f, ing, true)
			})
		})
	})
}
