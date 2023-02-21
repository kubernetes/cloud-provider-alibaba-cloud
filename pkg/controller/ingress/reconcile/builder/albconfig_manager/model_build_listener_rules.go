package albconfigmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
	"k8s.io/klog/v2"

	"github.com/pkg/errors"
)

const (
	ListenerRuleNamePrefix = "rule"
	HTTPRedirectCode       = "308"
	CookieAlways           = "always"
	HTTPS443               = "443"
	KnativeIngress         = "knative.aliyun.com/ingress"
)

var (
	actionTypeMap = map[string]string{
		util.RuleActionTypeFixedResponse: "FixedResponseConfig",
		util.RuleActionTypeRedirect:      "RedirectConfig",
		util.RuleActionTypeInsertHeader:  "InsertHeaderConfig",
		util.RuleActionTypeRemoveHeader:  "RemoveHeaderConfig",
		util.RuleActionTypeTrafficMirror: "TrafficMirrorConfig",
	}
)

// use lower case
var (
	conditionTypeMap = map[string]string{
		util.RuleConditionFieldHost:        "hostConfig",
		util.RuleConditionFieldPath:        "pathConfig",
		util.RuleConditionFieldHeader:      "headerConfig",
		util.RuleConditionFieldQueryString: "queryStringConfig",
		util.RuleConditionFieldMethod:      "methodConfig",
		util.RuleConditionFieldCookie:      "cookieConfig",
		util.RuleConditionFieldSourceIp:    "sourceIpConfig",

		util.RuleConditionResponseHeader:     "responseHeaderConfig",
		util.RuleConditionResponseStatusCode: "responseStatusCodeConfig",
	}
)

func (t *defaultModelBuildTask) checkCanaryAndCustomizeCondition(ing networking.Ingress) (bool, error) {
	for _, rule := range ing.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			_, exist := ing.Annotations[fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, path.Backend.Service.Name)]
			if exist {
				klog.Errorf("%v can't exist Canary and customize condition at the same time", util.NamespacedName(&ing))
				return false, fmt.Errorf("%v can't exist Canary and customize condition at the same time", util.NamespacedName(&ing))
			}
		}
	}
	return true, nil
}

type canarySGPWithIngress struct {
	canaryServerGroupTuple alb.ServerGroupTuple
	canaryIngress          networking.Ingress
}

func (t *defaultModelBuildTask) buildListenerRules(ctx context.Context, lsID core.StringToken, port int32, protocol Protocol, ingList []networking.Ingress) error {
	if len(ingList) > 0 {
		ing := ingList[0]
		if _, ok := ing.Labels[KnativeIngress]; ok {
			return t.buildListenerRulesCommon(ctx, lsID, port, ingList)
		}
	}
	var rules []alb.ListenerRule
	canaryServerGroupWithIngress := make(map[string][]canarySGPWithIngress, 0)
	nonCanaryPath := make(map[string]bool, 0)
	for _, ing := range ingList {
		if v := annotations.GetStringAnnotationMutil(annotations.NginxCanary, annotations.AlbCanary, &ing); v == "true" {
			//canary and customize condition can't simultaneously exist
			if ok, err := t.checkCanaryAndCustomizeCondition(ing); !ok {
				t.errResultWithIngress[&ing] = err
				return err
			}
			weight := getIfOnlyWeight(&ing)
			if weight == 0 {
				continue
			}
			for _, rule := range ing.Spec.Rules {
				if rule.HTTP == nil {
					continue
				}
				for _, path := range rule.HTTP.Paths {
					if _, ok := canaryServerGroupWithIngress[rule.Host+"-"+path.Path]; !ok {
						canaryServerGroupWithIngress[rule.Host+"-"+path.Path] = make([]canarySGPWithIngress, 0)
					}
					canaryServerGroupWithIngress[rule.Host+"-"+path.Path] = append(canaryServerGroupWithIngress[rule.Host+"-"+path.Path], canarySGPWithIngress{
						canaryServerGroupTuple: alb.ServerGroupTuple{
							ServiceName: path.Backend.Service.Name,
							ServicePort: int(path.Backend.Service.Port.Number),
							Weight:      weight,
						},
						canaryIngress: ing,
					})

				}
			}
		} else {
			for _, rule := range ing.Spec.Rules {
				if rule.HTTP == nil {
					continue
				}
				for _, path := range rule.HTTP.Paths {
					nonCanaryPath[rule.Host+"-"+path.Path] = true
				}
			}
		}
	}
	for _, ing := range ingList {
		if v := annotations.GetStringAnnotationMutil(annotations.NginxCanary, annotations.AlbCanary, &ing); v == "true" {
			we := annotations.GetStringAnnotationMutil(annotations.NginxCanaryWeight, annotations.AlbCanaryWeight, &ing)
			if we != "" {
				continue
			}
		}
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			if rule.Host != "" && !forceListenIngress(ing) {
				isTlsRule := withTlsRule(protocol, rule.Host, ing)
				if protocol == ProtocolHTTPS && !isTlsRule {
					continue
				}
				if protocol == ProtocolHTTP && isTlsRule {
					continue
				}
			}
			for _, path := range rule.HTTP.Paths {
				if _, ok := nonCanaryPath[rule.Host+"-"+path.Path]; !ok {
					continue
				}
				actions, err := t.buildRuleActions(ctx, &ing, &path, canaryServerGroupWithIngress[rule.Host+"-"+path.Path], port == 443)
				if err != nil {
					t.errResultWithIngress[&ing] = err
					return errors.Wrapf(err, "buildListenerRules-Actions(ingress: %v)", util.NamespacedName(&ing))
				}
				conditions, err := t.buildRuleConditions(ctx, rule, path, ing)
				if err != nil {
					t.errResultWithIngress[&ing] = err
					return errors.Wrapf(err, "buildListenerRules-Conditions(ingress: %v)", util.NamespacedName(&ing))
				}
				direction, err := t.buildRuleDirection(ctx, rule, path, ing)
				if err != nil {
					t.errResultWithIngress[&ing] = err
					return errors.Wrapf(err, "buildListenerRules-Direction(ingress: %v)", util.NamespacedName(&ing))
				}
				rules = append(rules, alb.ListenerRule{
					Spec: alb.ListenerRuleSpec{
						ListenerID: lsID,
						ALBListenerRuleSpec: alb.ALBListenerRuleSpec{
							RuleActions:    actions,
							RuleConditions: conditions,
							RuleDirection:  direction,
						},
					},
				})
			}
		}
	}

	priority := 1
	for _, rule := range rules {
		ruleResID := fmt.Sprintf("%v:%v", port, priority)
		klog.Infof("ruleResID: %s", ruleResID)
		lrs := alb.ListenerRuleSpec{
			ListenerID: lsID,
			ALBListenerRuleSpec: alb.ALBListenerRuleSpec{
				Priority:       priority,
				RuleConditions: rule.Spec.RuleConditions,
				RuleActions:    rule.Spec.RuleActions,
				RuleName:       fmt.Sprintf("%v-%v-%v", ListenerRuleNamePrefix, port, priority),
				RuleDirection:  rule.Spec.RuleDirection,
			},
		}
		_ = alb.NewListenerRule(t.stack, ruleResID, lrs)
		priority += 1
	}

	return nil
}

/*
 * true if rule need config on https listener
 */
func withTlsRule(protocol Protocol, host string, ing networking.Ingress) bool {
	// only https listen && len(tls) == 0: use annotation listen-port config https
	if protocol == ProtocolHTTPS && len(ing.Spec.TLS) == 0 {
		return true
	}
	// host == "": the rule use to default request
	if host == "" {
		return true
	}
	for _, tls := range ing.Spec.TLS {
		if contains(tls.Hosts, host) {
			return true
		}
	}
	return false
}

func forceListenIngress(ing networking.Ingress) bool {
	if v := annotations.GetStringAnnotationMutil(annotations.NginxSslRedirect, annotations.AlbSslRedirect, &ing); v == "true" {
		return true
	}
	_, err := annotations.GetStringAnnotation(annotations.ListenPorts, &ing)
	return err == nil
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func (t *defaultModelBuildTask) buildListenerRulesCommon(ctx context.Context, lsID core.StringToken, port int32, ingList []networking.Ingress) error {
	var rules []alb.ListenerRule
	for _, ing := range ingList {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				var action alb.Action
				actionStr := ing.Annotations[fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, path.Backend.Service.Name)]
				if actionStr != "" {
					err := json.Unmarshal([]byte(actionStr), &action)
					if err != nil {
						klog.Errorf("buildListenerRulesCommon: %s Unmarshal: %s", actionStr, err.Error())
						continue
					}
				}
				klog.Infof("INGRESS_ALB_ACTIONS_ANNOTATIONS: %s", actionStr)
				conditions, err := t.buildRuleConditionsCommon(ctx, rule, path, ing)
				if err != nil {
					klog.Errorf("buildListenerRulesCommon error: %s", err.Error())
					continue
				}
				action2, err := t.buildAction(ctx, ing, action)
				if err != nil {
					klog.Errorf("buildListenerRulesCommon error: %s", err.Error())
					continue
				}
				actions := []alb.Action{action2}
				if v, _ := annotations.GetStringAnnotation(annotations.AlbSslRedirect, &ing); v == "true" && !(port == 443) {
					actions = []alb.Action{buildActionViaHostAndPath(ctx, path.Path)}
				}
				lrs := alb.ListenerRuleSpec{
					ListenerID: lsID,
				}
				lrs.RuleActions = actions
				lrs.RuleConditions = conditions
				rules = append(rules, alb.ListenerRule{
					Spec: lrs,
				})
			}
		}
	}

	priority := 1
	for _, rule := range rules {
		ruleResID := fmt.Sprintf("%v:%v", port, priority)
		klog.Infof("ruleResID: %s", ruleResID)
		lrs := alb.ListenerRuleSpec{
			ListenerID: lsID,
		}
		lrs.Priority = priority
		lrs.RuleConditions = rule.Spec.RuleConditions
		lrs.RuleActions = rule.Spec.RuleActions
		lrs.RuleName = fmt.Sprintf("%v-%v-%v", ListenerRuleNamePrefix, port, priority)
		_ = alb.NewListenerRule(t.stack, ruleResID, lrs)
		priority += 1
	}

	return nil
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

func buildActionViaHostAndPath(_ context.Context, path string) alb.Action {
	action := alb.Action{
		Type: util.RuleActionTypeRedirect,
		RedirectConfig: &alb.RedirectConfig{
			Host:     "${host}",
			Path:     "${path}",
			Protocol: util.ListenerProtocolHTTPS,
			Port:     HTTPS443,
			HttpCode: HTTPRedirectCode,
			Query:    "${query}",
		},
	}
	return action
}

func (t *defaultModelBuildTask) buildPathPatternsForImplementationSpecificPathType(path string) ([]string, error) {
	return []string{path}, nil
}

func (t *defaultModelBuildTask) buildPathPatternsForExactPathType(path string) ([]string, error) {
	if strings.ContainsAny(path, "*?") {
		return nil, errors.Errorf("exact path shouldn't contain wildcards: %v", path)
	}
	return []string{path}, nil
}

func (t *defaultModelBuildTask) buildPathPatternsForPrefixPathType(path string, ing networking.Ingress) ([]string, error) {
	if path == "/" {
		return []string{"/*"}, nil
	}
	useRegex := false
	if v, err := annotations.GetStringAnnotation(annotations.AlbRewriteTarget, &ing); err == nil && v != path {
		useRegex = true
	}
	var paths []string
	if useRegex {
		paths = []string{"~*" + path}
	} else {
		if strings.ContainsAny(path, "*?") {
			return nil, errors.Errorf("prefix path shouldn't contain wildcards: %v", path)
		}
		normalizedPath := strings.TrimSuffix(path, "/")
		paths = []string{normalizedPath, normalizedPath + "/*"}
	}
	return paths, nil
}

func (t *defaultModelBuildTask) buildPathPatterns(path string, pathType *networking.PathType, ing networking.Ingress) ([]string, error) {
	normalizedPathType := networking.PathTypeImplementationSpecific
	if pathType != nil {
		normalizedPathType = *pathType
	}
	switch normalizedPathType {
	case networking.PathTypeImplementationSpecific:
		return t.buildPathPatternsForImplementationSpecificPathType(path)
	case networking.PathTypeExact:
		return t.buildPathPatternsForExactPathType(path)
	case networking.PathTypePrefix:
		return t.buildPathPatternsForPrefixPathType(path, ing)
	default:
		return nil, errors.Errorf("unsupported pathType: %v", normalizedPathType)
	}
}

func (t *defaultModelBuildTask) buildSourceIpCondition(_ context.Context, sourceIps []string) alb.Condition {
	return alb.Condition{
		Type: util.RuleConditionFieldSourceIp,
		SourceIpConfig: alb.SourceIpConfig{
			Values: sourceIps,
		},
	}
}

func (t *defaultModelBuildTask) buildHostHeaderCondition(_ context.Context, hosts []string) alb.Condition {
	return alb.Condition{
		Type: util.RuleConditionFieldHost,
		HostConfig: alb.HostConfig{
			Values: hosts,
		},
	}
}

func (t *defaultModelBuildTask) buildMethodCondition(_ context.Context, methods []string) alb.Condition {
	return alb.Condition{
		Type: util.RuleConditionFieldMethod,
		MethodConfig: alb.MethodConfig{
			Values: methods,
		},
	}
}

func (t *defaultModelBuildTask) buildHeaderCondition(_ context.Context, key string, values []string) alb.Condition {
	return alb.Condition{
		Type: util.RuleConditionFieldHeader,
		HeaderConfig: alb.HeaderConfig{
			Key:    key,
			Values: values,
		},
	}
}
func (t *defaultModelBuildTask) buildCookieCondition(_ context.Context, key string, value string) alb.Condition {
	return alb.Condition{
		Type: util.RuleConditionFieldCookie,
		CookieConfig: alb.CookieConfig{
			Values: []alb.Value{{
				Key:   key,
				Value: value,
			},
			},
		},
	}
}

func (t *defaultModelBuildTask) buildPathPatternCondition(_ context.Context, paths []string) alb.Condition {
	return alb.Condition{
		Type: util.RuleConditionFieldPath,
		PathConfig: alb.PathConfig{
			Values: paths,
		},
	}
}

func (t *defaultModelBuildTask) buildResponseStatusCodeCondition(_ context.Context, statusCodes []string) alb.Condition {
	return alb.Condition{
		Type: util.RuleConditionResponseStatusCode,
		ResponseStatusCodeConfig: alb.ResponseStatusCodeConfig{
			Values: statusCodes,
		},
	}
}

func readConditionType(actType interface{}) (string, error) {
	tp := actType.(string)
	switch tp {
	case util.RuleConditionFieldHost, util.RuleConditionFieldPath, util.RuleConditionFieldHeader, util.RuleConditionFieldQueryString,
		util.RuleConditionFieldMethod, util.RuleConditionFieldCookie, util.RuleConditionFieldSourceIp:
		return tp, nil
	case util.RuleConditionResponseHeader, util.RuleConditionResponseStatusCode:
		return tp, nil
	default:
		return tp, fmt.Errorf("readConditionType Failed(unknown condition type): %s", tp)
	}
}

func conditionCfgByType(conType string, conMap map[string]interface{}) interface{} {
	return conMap[conditionTypeMap[conType]]
}

// readNoKeyCondition for host/method/sourceIp/path condition
func readValuesNoKeyCondition(conCfg interface{}, conType string) ([]string, error) {
	bConCfg, err := json.Marshal(conCfg)
	if err != nil {
		return nil, err
	}
	switch conType {
	case util.RuleConditionFieldHost:
		hostCfg := alb.HostConfig{}
		err = json.Unmarshal(bConCfg, &hostCfg)
		if err != nil {
			return nil, err
		}
		return hostCfg.Values, nil
	case util.RuleConditionFieldPath:
		pathCfg := alb.PathConfig{}
		err = json.Unmarshal(bConCfg, &pathCfg)
		if err != nil {
			return nil, err
		}
		return pathCfg.Values, nil
	case util.RuleConditionFieldMethod:
		methodCfg := alb.MethodConfig{}
		err = json.Unmarshal(bConCfg, &methodCfg)
		if err != nil {
			return nil, err
		}
		return methodCfg.Values, nil
	case util.RuleConditionFieldSourceIp:
		sourceIpCfg := alb.SourceIpConfig{}
		err = json.Unmarshal(bConCfg, &sourceIpCfg)
		if err != nil {
			return nil, err
		}
		return sourceIpCfg.Values, nil
	case util.RuleConditionResponseStatusCode:
		responseStatusCodeCfg := alb.ResponseStatusCodeConfig{}
		err = json.Unmarshal(bConCfg, &responseStatusCodeCfg)
		if err != nil {
			return nil, err
		}
		return responseStatusCodeCfg.Values, nil

	default:
		return nil, fmt.Errorf("readCondition Failed(unknown condition type): %s", conType)
	}
}

func (t *defaultModelBuildTask) buildRuleDirection(ctx context.Context, rule networking.IngressRule,
	path networking.HTTPIngressPath, ing networking.Ingress) (string, error) {
	direction, exist := ing.Annotations[fmt.Sprintf(annotations.INGRESS_ALB_RULE_DIRECTION, path.Backend.Service.Name)]
	if exist {
		switch direction {
		case util.RuleRequestDirection:
			return util.RuleRequestDirection, nil
		case util.RuleResponseDirection:
			return util.RuleResponseDirection, nil
		default:
			return "", fmt.Errorf("readDirection Failed(unknown direction type): %s", direction)
		}
	}
	return util.RuleRequestDirection, nil
}

func (t *defaultModelBuildTask) buildRuleConditions(ctx context.Context, rule networking.IngressRule,
	path networking.HTTPIngressPath, ing networking.Ingress) ([]alb.Condition, error) {
	var conditions []alb.Condition
	//first deal with customize condition
	var conditionsMapArray []map[string]interface{}
	//no key condition
	var hosts []string
	var paths []string
	var sourceIps []string
	var methods []string
	var statusCodes []string
	conditionStr, exist := ing.Annotations[fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, path.Backend.Service.Name)]
	if exist {
		err := json.Unmarshal([]byte(conditionStr), &conditionsMapArray)
		if err != nil {
			klog.Errorf("buildRuleConditions: %s Unmarshal: %s", conditionStr, err.Error())
			return nil, err
		}
		for _, conditionMap := range conditionsMapArray {
			conditionType, err := readConditionType(conditionMap["type"])
			if err != nil {
				return nil, err
			}
			conditionCfg := conditionCfgByType(conditionType, conditionMap)
			if conditionCfg == nil {
				return nil, fmt.Errorf("condition Config is nil with type(%s)", conditionType)
			}
			switch conditionType {
			case util.RuleConditionFieldHost:
				values, err := readValuesNoKeyCondition(conditionCfg, conditionType)
				if err != nil {
					return nil, err
				}
				hosts = append(hosts, values...)
			case util.RuleConditionFieldPath:
				values, err := readValuesNoKeyCondition(conditionCfg, conditionType)
				if err != nil {
					return nil, err
				}
				paths = append(paths, values...)
			case util.RuleConditionFieldMethod:
				values, err := readValuesNoKeyCondition(conditionCfg, conditionType)
				if err != nil {
					return nil, err
				}
				methods = append(methods, values...)
			case util.RuleConditionFieldSourceIp:
				values, err := readValuesNoKeyCondition(conditionCfg, conditionType)
				if err != nil {
					return nil, err
				}
				sourceIps = append(sourceIps, values...)
			case util.RuleConditionResponseStatusCode:
				values, err := readValuesNoKeyCondition(conditionCfg, conditionType)
				if err != nil {
					return nil, err
				}
				statusCodes = append(statusCodes, values...)
			case util.RuleConditionFieldHeader:
				headerCondition := alb.Condition{}
				headerCondition.Type = util.RuleConditionFieldHeader
				bConCfg, err := json.Marshal(conditionCfg)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(bConCfg, &headerCondition.HeaderConfig)
				if err != nil {
					return nil, err
				}
				conditions = append(conditions, headerCondition)
			case util.RuleConditionFieldQueryString:
				queryStringCondition := alb.Condition{}
				queryStringCondition.Type = util.RuleConditionFieldQueryString
				bConCfg, err := json.Marshal(conditionCfg)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(bConCfg, &queryStringCondition.QueryStringConfig)
				if err != nil {
					return nil, err
				}
				conditions = append(conditions, queryStringCondition)
			case util.RuleConditionFieldCookie:
				cookieCondition := alb.Condition{}
				cookieCondition.Type = util.RuleConditionFieldCookie
				bConCfg, err := json.Marshal(conditionCfg)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(bConCfg, &cookieCondition.CookieConfig)
				if err != nil {
					return nil, err
				}
				conditions = append(conditions, cookieCondition)
			case util.RuleConditionResponseHeader:
				responseHeaderCondition := alb.Condition{}
				responseHeaderCondition.Type = util.RuleConditionResponseHeader
				bConCfg, err := json.Marshal(conditionCfg)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(bConCfg, &responseHeaderCondition.ResponseHeaderConfig)
				if err != nil {
					return nil, err
				}
				conditions = append(conditions, responseHeaderCondition)
			default:
				return nil, fmt.Errorf("readCondition Failed(unknown condition type): %s", conditionType)
			}
		}
	}

	if rule.Host != "" {
		hosts = append(hosts, rule.Host)
	}
	if path.Path != "" {
		pathPatterns, err := t.buildPathPatterns(path.Path, path.PathType, ing)
		if err != nil {
			return nil, err
		}
		paths = append(paths, pathPatterns...)
	}

	if len(hosts) != 0 {
		conditions = append(conditions, t.buildHostHeaderCondition(ctx, hosts))
	}
	if len(paths) != 0 {
		conditions = append(conditions, t.buildPathPatternCondition(ctx, paths))
	}
	if len(methods) != 0 {
		conditions = append(conditions, t.buildMethodCondition(ctx, methods))
	}
	if len(sourceIps) != 0 {
		conditions = append(conditions, t.buildSourceIpCondition(ctx, sourceIps))
	}
	if len(statusCodes) != 0 {
		conditions = append(conditions, t.buildResponseStatusCodeCondition(ctx, statusCodes))
	}
	if v := annotations.GetStringAnnotationMutil(annotations.NginxCanary, annotations.AlbCanary, &ing); v == "true" {
		header := annotations.GetStringAnnotationMutil(annotations.NginxCanaryByHeader, annotations.AlbCanaryByHeader, &ing)
		if header != "" {
			value := annotations.GetStringAnnotationMutil(annotations.NginxCanaryByHeaderValue, annotations.AlbCanaryByHeaderValue, &ing)
			conditions = append(conditions, t.buildHeaderCondition(ctx, header, []string{value}))
		}
		cookie := annotations.GetStringAnnotationMutil(annotations.NginxCanaryByCookie, annotations.AlbCanaryByCookie, &ing)
		if cookie != "" {
			conditions = append(conditions, t.buildCookieCondition(ctx, cookie, CookieAlways))
		}

	}

	if len(conditions) == 0 {
		conditions = append(conditions, t.buildPathPatternCondition(ctx, []string{"/*"}))
	}

	return conditions, nil
}

func (t *defaultModelBuildTask) buildRuleConditionsCommon(ctx context.Context, rule networking.IngressRule,
	path networking.HTTPIngressPath, ing networking.Ingress) ([]alb.Condition, error) {
	var hosts []string
	if rule.Host != "" {
		hosts = append(hosts, rule.Host)
	}
	var paths []string
	if path.Path != "" {
		pathPatterns, err := t.buildPathPatterns(path.Path, path.PathType, ing)
		if err != nil {
			return nil, err
		}
		paths = append(paths, pathPatterns...)
	}

	var conditions []alb.Condition
	if len(hosts) != 0 {
		conditions = append(conditions, t.buildHostHeaderCondition(ctx, hosts))
	}
	if len(paths) != 0 {
		conditions = append(conditions, t.buildPathPatternCondition(ctx, paths))
	}
	conditionItems := make([]alb.Condition, 0)
	conditionStr := ing.Annotations[fmt.Sprintf(annotations.INGRESS_ALB_CONDITIONS_ANNOTATIONS, path.Backend.Service.Name)]
	if conditionStr != "" {
		klog.Infof("INGRESS_ALB_CONDITIONS_ANNOTATIONS: %s", conditionStr)
		err := json.Unmarshal([]byte(conditionStr), &conditionItems)
		if err != nil {
			return conditions, fmt.Errorf("buildRuleConditionsCommon: %s Unmarshal: %s", conditionStr, err.Error())
		}
		//for _, item := range conditionItems {
		conditions = append(conditions, conditionItems...)
		//}
	}

	if len(conditions) == 0 {
		conditions = append(conditions, t.buildPathPatternCondition(ctx, []string{"/*"}))
	}

	return conditions, nil
}
func (t *defaultModelBuildTask) buildFixedResponseAction(_ context.Context, actionCfg alb.Action) (*alb.Action, error) {
	if len(actionCfg.FixedResponseConfig.ContentType) == 0 {
		return nil, errors.New("missing FixedResponseConfig")
	}
	return &alb.Action{
		Type: util.RuleActionTypeFixedResponse,
		FixedResponseConfig: &alb.FixedResponseConfig{
			ContentType: actionCfg.FixedResponseConfig.ContentType,
			Content:     actionCfg.FixedResponseConfig.Content,
			HttpCode:    actionCfg.FixedResponseConfig.HttpCode,
		},
	}, nil
}

func (t *defaultModelBuildTask) buildRedirectAction(_ context.Context, actionCfg alb.Action) (*alb.Action, error) {
	if actionCfg.RedirectConfig == nil {
		return nil, errors.New("missing RedirectConfig")
	}
	return &alb.Action{
		Type: util.RuleActionTypeRedirect,
		RedirectConfig: &alb.RedirectConfig{
			Host:     actionCfg.RedirectConfig.Host,
			Path:     actionCfg.RedirectConfig.Path,
			Port:     actionCfg.RedirectConfig.Port,
			Protocol: actionCfg.RedirectConfig.Protocol,
			Query:    actionCfg.RedirectConfig.Query,
			HttpCode: actionCfg.RedirectConfig.HttpCode,
		},
	}, nil
}

func (t *defaultModelBuildTask) buildForwardAction(ctx context.Context, ing networking.Ingress, actionCfg alb.Action) (*alb.Action, error) {
	if actionCfg.ForwardConfig == nil {
		return nil, errors.New("missing ForwardConfig")
	}

	var serverGroupTuples []alb.ServerGroupTuple
	for _, sgp := range actionCfg.ForwardConfig.ServerGroups {
		svc := new(corev1.Service)
		svc.Namespace = ing.Namespace
		svc.Name = sgp.ServiceName
		modelSgp, err := t.buildServerGroup(ctx, &ing, svc, sgp.ServicePort)
		if err != nil {
			return nil, err
		}
		serverGroupTuples = append(serverGroupTuples, alb.ServerGroupTuple{
			ServerGroupID: modelSgp.ServerGroupID(),
			Weight:        sgp.Weight,
		})
	}

	return &alb.Action{
		Type: util.RuleActionTypeForward,
		ForwardConfig: &alb.ForwardActionConfig{
			ServerGroups: serverGroupTuples,
		},
	}, nil
}

func (t *defaultModelBuildTask) buildBackendAction(ctx context.Context, ing networking.Ingress, actionCfg alb.Action) (*alb.Action, error) {
	switch actionCfg.Type {
	case util.RuleActionTypeFixedResponse:
		return t.buildFixedResponseAction(ctx, actionCfg)
	case util.RuleActionTypeRedirect:
		return t.buildRedirectAction(ctx, actionCfg)
	case util.RuleActionTypeForward:
		return t.buildForwardAction(ctx, ing, actionCfg)
	}
	return nil, errors.Errorf("unknown action type: %v", actionCfg.Type)
}
func (t *defaultModelBuildTask) buildAction(ctx context.Context, ing networking.Ingress, action alb.Action) (alb.Action, error) {
	backendAction, err := t.buildBackendAction(ctx, ing, action)
	if err != nil {
		return alb.Action{}, err
	}
	return *backendAction, nil
}

func getIfOnlyWeight(ing *networking.Ingress) int {
	weight := 0
	w := annotations.GetStringAnnotationMutil(annotations.NginxCanaryWeight, annotations.AlbCanaryWeight, ing)
	header := annotations.GetStringAnnotationMutil(annotations.NginxCanaryByHeader, annotations.AlbCanaryByHeader, ing)
	headerValue := annotations.GetStringAnnotationMutil(annotations.NginxCanaryByHeaderValue, annotations.AlbCanaryByHeaderValue, ing)
	cookie := annotations.GetStringAnnotationMutil(annotations.NginxCanaryByCookie, annotations.AlbCanaryByCookie, ing)
	if header != "" || headerValue != "" || cookie != "" {
		return 0
	}
	weight, _ = strconv.Atoi(w)
	return weight
}

func (t *defaultModelBuildTask) buildRuleActions(ctx context.Context, ing *networking.Ingress, path *networking.HTTPIngressPath, canaryWithIngress []canarySGPWithIngress, listen443 bool) ([]alb.Action, error) {
	actions := make([]alb.Action, 0)
	rawActions := make([]alb.Action, 0)
	var actionsMapArray []map[string]interface{}
	var extAction alb.Action
	if v, err := annotations.GetStringAnnotation(annotations.AlbTrafficLimitQps, ing); err == nil {
		var qps int
		if qps, err = strconv.Atoi(v); err != nil {
			return actions, err
		}
		if qps < util.ActionTrafficLimitQpsMin || qps > util.ActionTrafficLimitQpsMax {
			return actions, fmt.Errorf("traffic limit action qps out of range: %d", qps)
		}
		extAction = alb.Action{
			Type: util.RuleActionTypeTrafficLimit,
			TrafficLimitConfig: &alb.TrafficLimitConfig{
				QPS: qps,
			},
		}
		rawActions = append(rawActions, extAction)
	}
	// CorsConfig must after TrafficMirror and before other
	if v, _ := annotations.GetStringAnnotation(annotations.AlbEnableCors, ing); v == "true" {
		var corsAllowOrigin []string = splitAndTrim(util.DefaultCorsAllowOrigin)
		if v, err := annotations.GetStringAnnotation(annotations.AlbCorsAllowOrigin, ing); err == nil {
			corsAllowOrigin = splitAndTrim(v)
		}
		var corsAllowMethods []string = splitAndTrim(util.DefaultCorsAllowMethods)
		if v, err := annotations.GetStringAnnotation(annotations.AlbCorsAllowMethods, ing); err == nil {
			corsAllowMethods = splitAndTrim(v)
		}
		var corsAllowHeaders []string = splitAndTrim(util.DefaultCorsAllowHeaders)
		if v, err := annotations.GetStringAnnotation(annotations.AlbCorsAllowHeaders, ing); err == nil {
			corsAllowHeaders = splitAndTrim(v)
		}
		var corsExposeHeaders []string
		if v, err := annotations.GetStringAnnotation(annotations.AlbCorsExposeHeaders, ing); err == nil {
			corsExposeHeaders = splitAndTrim(v)
		}
		var corsAllowCredentials string = util.DefaultCorsAllowCredentials
		if v, err := annotations.GetStringAnnotation(annotations.AlbCorsAllowCredentials, ing); err == nil {
			if v == "true" || v == "false" {
				corsAllowCredentials = map[string]string{
					"true":  "on",
					"false": "off",
				}[v]
			} else {
				klog.Warning("Unexpect AlbCorsAllowCredentials value, expect: true or false, got " + v)
			}
		}
		var corsMaxAge string = util.DefaultCorsMaxAge
		if v, err := annotations.GetStringAnnotation(annotations.AlbCorsMaxAge, ing); err == nil {
			corsMaxAge = v
		}

		corsAction := alb.Action{
			Type: util.RuleActionTypeCors,
			CorsConfig: &alb.CorsConfig{
				AllowCredentials: corsAllowCredentials,
				MaxAge:           corsMaxAge,
				AllowOrigin:      corsAllowOrigin,
				AllowMethods:     corsAllowMethods,
				AllowHeaders:     corsAllowHeaders,
				ExposeHeaders:    corsExposeHeaders,
			},
		}
		rawActions = append(rawActions, corsAction)
	}
	actionStr, exist := ing.Annotations[fmt.Sprintf(annotations.INGRESS_ALB_ACTIONS_ANNOTATIONS, path.Backend.Service.Name)]
	if exist {
		err := json.Unmarshal([]byte(actionStr), &actionsMapArray)
		if err != nil {
			klog.Errorf("buildRuleActions: %s Unmarshal: %s", actionStr, err.Error())
			return nil, err
		}
		for _, actionMap := range actionsMapArray {
			actionType, err := readActionType(actionMap["type"])
			if err != nil {
				return nil, err
			}
			actionCfg := actionCfgByType(actionType, actionMap)
			act, err := readAction(actionType, actionCfg)
			if err != nil {
				return nil, err
			}
			rawActions = append(rawActions, act)
		}
	}
	if v, err := annotations.GetStringAnnotation(annotations.AlbRewriteTarget, ing); err == nil {
		extAction = alb.Action{
			Type: util.RuleActionTypeRewrite,
			RewriteConfig: &alb.RewriteConfig{
				Host:  "${host}",
				Path:  v,
				Query: "${query}",
			},
		}
		rawActions = append(rawActions, extAction)
	}
	var finalAction alb.Action
	var hasFinal bool = false
	var err error
	if v := annotations.GetStringAnnotationMutil(annotations.NginxSslRedirect, annotations.AlbSslRedirect, ing); v == "true" && !listen443 {
		finalAction = buildActionViaHostAndPath(ctx, path.Path)
		hasFinal = true
	} else if path.Backend.Service.Port.Name != "use-annotation" {
		canary := annotations.GetStringAnnotationMutil(annotations.NginxCanary, annotations.AlbCanary, ing)
		finalAction, err = t.buildActionForwardSGP(ctx, ing, path, canary != "true" && len(canaryWithIngress) > 0, canaryWithIngress)
		hasFinal = true
	}
	if err != nil {
		return nil, err
	}
	if hasFinal {
		rawActions = append(rawActions, finalAction)
	}
	// buildExtAction
	for _, act := range rawActions {
		if act.Type == util.RuleActionTypeInsertHeader ||
			act.Type == util.RuleActionTypeRemoveHeader ||
			act.Type == util.RuleActionTypeRewrite ||
			act.Type == util.RuleActionTypeTrafficLimit ||
			act.Type == util.RuleActionTypeCors ||
			act.Type == util.RuleActionTypeTrafficMirror {
			actions = append(actions, act)
		}
	}
	// buildFinalAction
	var finalTypes []string
	for _, act := range rawActions {
		if act.Type == util.RuleActionTypeForward ||
			act.Type == util.RuleActionTypeRedirect ||
			act.Type == util.RuleActionTypeFixedResponse {
			actions = append(actions, act)
			finalTypes = append(finalTypes, act.Type)
		}
	}
	if len(finalTypes) > 1 {
		return actions, fmt.Errorf("multi finalType action find: %v", finalTypes)
	}
	return actions, nil
}

func (t *defaultModelBuildTask) buildForwardCanarySGP(ctx context.Context, canaryServerGroupWithIngress []canarySGPWithIngress) ([]alb.ServerGroupTuple, error) {
	var serverGroupTuples []alb.ServerGroupTuple
	for _, sgpWithIngress := range canaryServerGroupWithIngress {
		svc := new(corev1.Service)
		svc.Namespace = sgpWithIngress.canaryIngress.Namespace
		svc.Name = sgpWithIngress.canaryServerGroupTuple.ServiceName
		modelSgp, err := t.buildServerGroup(ctx, &sgpWithIngress.canaryIngress, svc, sgpWithIngress.canaryServerGroupTuple.ServicePort)
		if err != nil {
			return serverGroupTuples, err
		}
		serverGroupTuples = append(serverGroupTuples, alb.ServerGroupTuple{
			ServerGroupID: modelSgp.ServerGroupID(),
			Weight:        sgpWithIngress.canaryServerGroupTuple.Weight,
		})
	}
	return serverGroupTuples, nil
}

func (t *defaultModelBuildTask) buildActionForwardSGP(ctx context.Context, ing *networking.Ingress, path *networking.HTTPIngressPath, withCanary bool, canaryServerGroupWithIngress []canarySGPWithIngress) (alb.Action, error) {
	action := buildActionViaServiceAndServicePort(ctx, path.Backend.Service.Name, int(path.Backend.Service.Port.Number), 100)
	var canaryServerGroupTuples []alb.ServerGroupTuple
	if withCanary {
		canaryWeight := 0
		for _, canaryWithIngress := range canaryServerGroupWithIngress {
			canaryWeight += canaryWithIngress.canaryServerGroupTuple.Weight
		}
		if canaryWeight == 100 {
			serverGroupTuples, err := t.buildForwardCanarySGP(ctx, canaryServerGroupWithIngress)
			if err != nil {
				return alb.Action{}, err
			}
			action.ForwardConfig.ServerGroups = serverGroupTuples
			return action, nil
		} else {
			for i := range action.ForwardConfig.ServerGroups {
				action.ForwardConfig.ServerGroups[i].Weight = 100 - canaryWeight
			}
			serverGroupTuples, err := t.buildForwardCanarySGP(ctx, canaryServerGroupWithIngress)
			if err != nil {
				return alb.Action{}, err
			}
			canaryServerGroupTuples = serverGroupTuples
		}
	}
	var serverGroupTuples []alb.ServerGroupTuple
	for _, sgp := range action.ForwardConfig.ServerGroups {
		svc := new(corev1.Service)
		svc.Namespace = ing.Namespace
		svc.Name = sgp.ServiceName
		modelSgp, err := t.buildServerGroup(ctx, ing, svc, sgp.ServicePort)
		if err != nil {
			return alb.Action{}, err
		}
		serverGroupTuples = append(serverGroupTuples, alb.ServerGroupTuple{
			ServerGroupID: modelSgp.ServerGroupID(),
			Weight:        sgp.Weight,
		})
	}
	serverGroupTuples = append(serverGroupTuples, canaryServerGroupTuples...)
	action.ForwardConfig.ServerGroups = serverGroupTuples
	return action, nil
}

func readActionType(actType interface{}) (string, error) {
	tp := actType.(string)
	if tp != util.RuleActionTypeFixedResponse &&
		tp != util.RuleActionTypeRedirect &&
		tp != util.RuleActionTypeInsertHeader &&
		tp != util.RuleActionTypeRemoveHeader &&
		tp != util.RuleActionTypeTrafficMirror {
		return tp, fmt.Errorf("readActionType Failed(unknown action type): %s", tp)
	}
	return tp, nil
}

func actionCfgByType(actType string, actMap map[string]interface{}) interface{} {
	return actMap[actionTypeMap[actType]]
}

func readAction(actType string, actCfg interface{}) (alb.Action, error) {
	toAct := alb.Action{
		RedirectConfig: &alb.RedirectConfig{
			Host:     "${host}",
			HttpCode: "301",
			Path:     "${path}",
			Port:     "${port}",
			Protocol: "${protocol}",
			Query:    "${query}",
		},
	}
	if actCfg == nil {
		return toAct, fmt.Errorf("action Config is nil with type(%s)", actType)
	}
	bActCfg, err := json.Marshal(actCfg)
	if err != nil {
		return toAct, err
	}
	switch actType {
	case util.RuleActionTypeFixedResponse:
		toAct.Type = util.RuleActionTypeFixedResponse
		err = json.Unmarshal(bActCfg, &toAct.FixedResponseConfig)
	case util.RuleActionTypeRedirect:
		toAct.Type = util.RuleActionTypeRedirect
		err = json.Unmarshal(bActCfg, &toAct.RedirectConfig)
	case util.RuleActionTypeInsertHeader:
		toAct.Type = util.RuleActionTypeInsertHeader
		err = json.Unmarshal(bActCfg, &toAct.InsertHeaderConfig)
	case util.RuleActionTypeRemoveHeader:
		toAct.Type = util.RuleActionTypeRemoveHeader
		err = json.Unmarshal(bActCfg, &toAct.RemoveHeaderConfig)
	case util.RuleActionTypeTrafficMirror:
		toAct.Type = util.RuleActionTypeTrafficMirror
		err = json.Unmarshal(bActCfg, &toAct.TrafficMirrorConfig)
	default:
		return toAct, fmt.Errorf("readAction Failed(unknown action type): %s", actType)
	}
	return toAct, err
}

func splitAndTrim(values string) []string {
	ret := make([]string, 0)
	for _, value := range strings.Split(values, ",") {
		ret = append(ret, strings.Trim(value, " "))
	}
	return ret
}
