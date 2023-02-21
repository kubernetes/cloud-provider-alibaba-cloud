package alb

import (
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/klog/v2"
)

type runFunc func(f *framework.Framework)

type RegisterCaseE2E struct {
	f      runFunc
	reason string
	author string
	flag   string
}

func InSlice(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

var e2eAlbConfigCases []RegisterCaseE2E

func InitAlbConfigE2ECases() {
	e2eAlbConfigCases = append(e2eAlbConfigCases, RegisterCaseE2E{
		f:      RunTagLoadBalancerTestCases,
		reason: "tag albConfig test",
		author: "yuri",
		flag:   "tag-alb",
	})
}

func ExecuteAlbConfigE2ECases(frame *framework.Framework, albFlags []string) {
	for n, e2eFunc := range e2eAlbConfigCases {
		if len(albFlags) != 0 && !InSlice(albFlags, e2eFunc.flag) {
			continue
		}
		klog.Infof("ExecuteIngressE2ECases %d %s created by %s", n, e2eFunc.reason, e2eFunc.author)
		e2eFunc.f(frame)
	}
}

var e2eIngressCases []RegisterCaseE2E

func InitAlbIngressE2ECases() {
	e2eIngressCases = append(e2eIngressCases, RegisterCaseE2E{
		f:      RunIngressTestCases,
		reason: "basic ingress test",
		author: "yuri",
		flag:   "basic",
	})

	e2eIngressCases = append(e2eIngressCases, RegisterCaseE2E{
		f:      RunCustomizeConditionTestCases,
		reason: "test customize condition",
		author: "yuri",
		flag:   "customize-condition",
	})

	e2eIngressCases = append(e2eIngressCases, RegisterCaseE2E{
		f:      RunCanaryTestCases,
		reason: "test canary condition",
		author: "xiaosha",
		flag:   "canary",
	})

	e2eIngressCases = append(e2eIngressCases, RegisterCaseE2E{
		f:      RunHealthCheckTestCases,
		reason: "test healthCheck",
		author: "xiaosha",
		flag:   "healthcheck",
	})

	e2eIngressCases = append(e2eIngressCases, RegisterCaseE2E{
		f:      RunTrafficMirrorTestCases,
		reason: "test trafficMirror",
		author: "xiaosha",
		flag:   "traffic-mirror",
	})

	e2eIngressCases = append(e2eIngressCases, RegisterCaseE2E{
		f:      RunResponseRuleTestCases,
		reason: "test response rule",
		author: "yuri",
		flag:   "response-rule",
	})
}

func needExecuteIngressE2ECases(albFlags []string) bool {
	need := false
	if len(albFlags) == 0 {
		need = true
	}
	for _, e2eFunc := range e2eIngressCases {
		if InSlice(albFlags, e2eFunc.flag) {
			need = true
			break
		}
	}
	return need
}

func ExecuteIngressE2ECases(frame *framework.Framework, albFlags []string) {
	if !needExecuteIngressE2ECases(albFlags) {
		klog.Info("No Need Execute Ingress E2ECases")
		return
	}
	RunAlbConfigTestCases(frame)
	for n, e2eFunc := range e2eIngressCases {
		if len(albFlags) != 0 && !InSlice(albFlags, e2eFunc.flag) {
			continue
		}
		klog.Infof("ExecuteIngressE2ECases %d %s created by %s", n, e2eFunc.reason, e2eFunc.author)
		e2eFunc.f(frame)
	}
	CleanAlbconfigTestCases(frame)
}
