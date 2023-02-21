package alb

import (
	"github.com/onsi/ginkgo/v2"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/alb/common"
)

func RunHealthCheckTestCases(f *framework.Framework) {
	ingress := common.Ingress{}
	service := common.Service{}
	ginkgo.BeforeEach(func() {
		service.CreateDefaultService(f)
	})

	ginkgo.AfterEach(func() {
		ingress.DeleteIngress(f, ingress.DefaultIngress(f))
	})

	ginkgo.It("[alb][p0] ingress with TCP healthcheck Service", func() {
		ing := defaultIngress(f)
		anno := map[string]string{}
		anno["alb.ingress.kubernetes.io/healthcheck-enabled"] = "true"
		anno["alb.ingress.kubernetes.io/healthcheck-protocol"] = "TCP"
		anno["alb.ingress.kubernetes.io/healthcheck-timeout-seconds"] = "5"
		anno["alb.ingress.kubernetes.io/healthcheck-interval-seconds"] = "2"
		anno["alb.ingress.kubernetes.io/healthy-threshold-count"] = "3"
		anno["alb.ingress.kubernetes.io/unhealthy-threshold-count"] = "3"
		anno["alb.ingress.kubernetes.io/healthcheck-connect-port"] = "81"
		ing.Annotations = anno
		waitCreateIngress(f, ing)
	})
	ginkgo.It("[alb][p0] ingress with HTTP healthcheck Service", func() {
		ing := defaultIngress(f)
		anno := map[string]string{}
		anno["alb.ingress.kubernetes.io/healthcheck-enabled"] = "true"
		anno["alb.ingress.kubernetes.io/healthcheck-protocol"] = "HTTP"
		anno["alb.ingress.kubernetes.io/healthcheck-timeout-seconds"] = "5"
		anno["alb.ingress.kubernetes.io/healthcheck-interval-seconds"] = "2"
		anno["alb.ingress.kubernetes.io/healthy-threshold-count"] = "3"
		anno["alb.ingress.kubernetes.io/unhealthy-threshold-count"] = "3"
		anno["alb.ingress.kubernetes.io/healthcheck-connect-port"] = "82"
		ing.Annotations = anno
		waitCreateIngress(f, ing)
	})
}
