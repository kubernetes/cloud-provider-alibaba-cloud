package alb

import (
	"context"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
	"time"

	"github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/alb/common"
	"k8s.io/klog/v2"
)

var (
	redirect      = "{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}"
	insertHeader  = "{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"123\",\"valueType\":\"UserDefined\"}}"
	fixedResponse = "{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}"
)

func findTrafficMirrorServerGroupId(f *framework.Framework, tags map[string]string) (string, error) {
	var trafficMirrorServerGroupId string
	timeout := 60 * time.Second
	err := wait.Poll(5*time.Second, timeout, func() (done bool, err error) {
		sgps, _ := f.Client.CloudClient.ALBProvider.ListALBServerGroupsByTag(context.TODO(), tags)
		for _, sgp := range sgps {
			trafficMirrorServerGroupId = sgp.ServerGroupId
			klog.Infof("%s", trafficMirrorServerGroupId)
		}
		if trafficMirrorServerGroupId != "" {
			return true, nil
		}
		return false, nil
	})
	return trafficMirrorServerGroupId, err
}

func buildActionViaServiceAndServicePort(_ context.Context, svcName string, svcPort int, weight int) alb.Action {
	action := alb.Action{
		Type: util.RuleActionTypeForward,
		ForwardConfig: &alb.ForwardActionConfig{
			ServerGroups: []alb.ServerGroupTuple{
				{
					ServiceName: svcName,
					ServicePort: svcPort,
					Weight:      weight,
				},
			},
		},
	}
	return action
}

func buildServerGroupTags(ing *networkingv1.Ingress) (map[string]string, error) {
	Tags := make(map[string]string)
	timeout := 60 * time.Second
	err := wait.Poll(5*time.Second, timeout, func() (done bool, err error) {
		for _, rule := range ing.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				port := int(path.Backend.Service.Port.Number)
				action := buildActionViaServiceAndServicePort(context.TODO(), path.Backend.Service.Name, port, 100)
				for _, sgp := range action.ForwardConfig.ServerGroups {
					svc := new(corev1.Service)
					svc.Namespace = ing.Namespace
					svc.Name = sgp.ServiceName
					tags := make([]alb.ALBTag, 0)
					tags = append(tags, []alb.ALBTag{
						{
							Key:   util.ClusterNameTagKey,
							Value: options.TestConfig.ClusterId,
						},
						{
							Key:   util.ServiceNamespaceTagKey,
							Value: ing.Namespace,
						},
						{
							Key:   util.IngressNameTagKey,
							Value: ing.Name,
						},
						{
							Key:   util.ServiceNameTagKey,
							Value: svc.Name,
						},
						{
							Key:   util.ServicePortTagKey,
							Value: fmt.Sprintf("%v", port),
						},
					}...)
					for _, tag := range tags {
						Tags[tag.Key] = tag.Value
					}
				}
			}
		}
		if len(Tags) != 0 {
			return true, nil
		}
		return false, nil
	})
	return Tags, err
}

func RunTrafficMirrorTestCases(f *framework.Framework) {
	rule := common.Rule{}
	ingress := common.Ingress{}
	service := common.Service{}
	var trafficMirror string
	ginkgo.BeforeEach(func() {
		service.CreateDefaultService(f)
	})

	ginkgo.AfterEach(func() {
		ingress.DeleteIngress(f, ingress.DefaultIngress(f))
		ingress.DeleteIngress(f, defaultIngressWithSvcName(f, "traffic-mirror-svc"))
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorSingleTest := "[" + trafficMirror + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, serviceName): trafficMirrorSingleTest,
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
		ingress.WaitCreateIngress(f, ing, true)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and FixedResponse final action", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorFixedResponse := "[" + trafficMirror + "," + fixedResponse + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, serviceName): trafficMirrorFixedResponse,
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
		ingress.WaitCreateIngress(f, ing, false)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and redirect final action", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorRedirect := "[" + trafficMirror + "," + redirect + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, serviceName): trafficMirrorRedirect,
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
		ingress.WaitCreateIngress(f, ing, false)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and insert-header action", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorInsertHeader := "[" + trafficMirror + "," + insertHeader + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, serviceName): trafficMirrorInsertHeader,
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
		ingress.WaitCreateIngress(f, ing, false)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and rewrite-target", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorSingleTest := "[" + trafficMirror + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/actions.traffic-mirror": trafficMirrorSingleTest,
			"alb.ingress.kubernetes.io/rewrite-target":         "/path/${2}",
		}
		ing.Spec.Rules[0].HTTP.Paths[0].Path = "/something(/|$)(.*)"
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "traffic-mirror"))
		ingress.WaitCreateIngress(f, ing, false)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and FixedResponse/Redirect/InsertHeader action", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorSingleTest := "[" + trafficMirror + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/actions.response-503":   "[{\"type\":\"FixedResponse\",\"FixedResponseConfig\":{\"contentType\":\"text/plain\",\"httpCode\":\"503\",\"content\":\"503 error text\"}}]",
			"alb.ingress.kubernetes.io/actions.redirect":       "[{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
			"alb.ingress.kubernetes.io/actions.insert-header":  "[{\"type\":\"InsertHeader\",\"InsertHeaderConfig\":{\"key\":\"key\",\"value\":\"value\",\"valueType\":\"UserDefined\"}},{\"type\":\"Redirect\",\"RedirectConfig\":{\"httpCode\":\"307\",\"port\":\"443\"}}]",
			"alb.ingress.kubernetes.io/actions.traffic-mirror": trafficMirrorSingleTest,
		}
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path1", "response-503"))
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path2", "redirect"))
		ing.Spec.Rules = append(ing.Spec.Rules, defaultRule("/path3", "insert-header"))
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "traffic-mirror"))
		ingress.WaitCreateIngress(f, ing, true)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and all customize conditions", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorSingleTest := "[" + trafficMirror + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, serviceName): allConditions,
			fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, serviceName):    trafficMirrorSingleTest,
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("", "", serviceName))
		ingress.WaitCreateIngress(f, ing, true)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and TCP healthcheck Service", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorSingleTest := "[" + trafficMirror + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/actions.traffic-mirror":       trafficMirrorSingleTest,
			"alb.ingress.kubernetes.io/healthcheck-enabled":          "true",
			"alb.ingress.kubernetes.io/healthcheck-protocol":         "TCP",
			"alb.ingress.kubernetes.io/healthcheck-timeout-seconds":  "5",
			"alb.ingress.kubernetes.io/healthcheck-interval-seconds": "2",
			"alb.ingress.kubernetes.io/healthy-threshold-count":      "3",
			"alb.ingress.kubernetes.io/unhealthy-threshold-count":    "3",
			"alb.ingress.kubernetes.io/healthcheck-connect-port":     "81",
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "traffic-mirror"))
		ingress.WaitCreateIngress(f, ing, true)
	})
	ginkgo.It("[alb][p0] ingress with traffic-mirror action and canary-weight", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorSingleTest := "[" + trafficMirror + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/actions.traffic-mirror": trafficMirrorSingleTest,
			"alb.ingress.kubernetes.io/canary":                 "true",
			"alb.ingress.kubernetes.io/canary-weight":          "50",
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "canary"))
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "canary-weight"))
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "traffic-mirror"))
		ingress.WaitCreateIngress(f, ing, true)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and canary-by-header", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorSingleTest := "[" + trafficMirror + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/actions.traffic-mirror": trafficMirrorSingleTest,
			"alb.ingress.kubernetes.io/canary":                 "true",
			"alb.ingress.kubernetes.io/canary-by-header":       "location",
			"alb.ingress.kubernetes.io/canary-by-header-value": "hz",
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "canary"))
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "canary-by-header"))
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "canary-by-header-value"))
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "traffic-mirror"))
		ingress.WaitCreateIngress(f, ing, true)
	})

	ginkgo.It("[alb][p0] ingress with traffic-mirror action and traffic-limit", func() {
		service.WaitCreateDefaultServiceWithSvcName(f, "traffic-mirror-svc")
		sgpIng := defaultIngressWithSvcName(f, "traffic-mirror-svc")
		ingress.WaitCreateIngress(f, sgpIng, true)
		tags, _ := buildServerGroupTags(sgpIng)
		klog.Infof("%s %s %s %s", tags[util.ServiceNamespaceTagKey], tags[util.IngressNameTagKey], tags[util.ServiceNameTagKey], tags[util.ServicePortTagKey])
		trafficMirrorServerGroupId, _ := findTrafficMirrorServerGroupId(f, tags)
		trafficMirror = "{\"type\": \"TrafficMirror\",\"TrafficMirrorConfig\": {\"TargetType\" : \"ForwardGroupMirror\",\"MirrorGroupConfig\": {\"ServerGroupTuples\" : [{\"ServerGroupID\": \"" + trafficMirrorServerGroupId + "\" }] } } }"
		trafficMirrorSingleTest := "[" + trafficMirror + "]"
		ing := defaultIngress(f)
		ing.Annotations = map[string]string{
			"alb.ingress.kubernetes.io/actions.traffic-mirror": trafficMirrorSingleTest,
			"alb.ingress.kubernetes.io/traffic-limit-qps":      "50",
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule.DefaultRule("/path1", "", "traffic-mirror"))
		ingress.WaitCreateIngress(f, ing, true)
	})
}
