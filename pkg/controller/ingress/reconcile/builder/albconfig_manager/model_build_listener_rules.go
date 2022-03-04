package albconfigmanager

import (
	"context"
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
)

func (t *defaultModelBuildTask) buildListenerRules(ctx context.Context, lsID core.StringToken, port int32, ingList []networking.Ingress) error {
	var rules []alb.ListenerRule
	carryWeight := make(map[string][]alb.ServerGroupTuple)
	for _, ing := range ingList {
		if v := annotations.GetStringAnnotationMutil(annotations.NginxCanary, annotations.AlbCanary, &ing); v == "true" {
			weight := 0
			w := annotations.GetStringAnnotationMutil(annotations.NginxCanaryWeight, annotations.AlbCanaryWeight, &ing)
			wi, _ := strconv.Atoi(w)
			weight = wi
			for _, rule := range ing.Spec.Rules {
				if rule.HTTP == nil {
					continue
				}
				for _, path := range rule.HTTP.Paths {
					if _, ok := carryWeight[rule.Host+"-"+path.Path]; ok {
						carryWeight[rule.Host+"-"+path.Path] = append(carryWeight[rule.Host+"-"+path.Path], alb.ServerGroupTuple{
							ServiceName: path.Backend.Service.Name,
							ServicePort: int(path.Backend.Service.Port.Number),
							Weight:      weight,
						})
					} else {
						carryWeight[rule.Host+"-"+path.Path] = []alb.ServerGroupTuple{{
							ServiceName: path.Backend.Service.Name,
							ServicePort: int(path.Backend.Service.Port.Number),
							Weight:      weight,
						},
						}
					}

					break
				}
			}
			break
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

			for _, path := range rule.HTTP.Paths {
				var action alb.Action
				if v := annotations.GetStringAnnotationMutil(annotations.NginxSslRedirect, annotations.AlbSslRedirect, &ing); v == "true" && port != 443 {
					action = buildActionViaHostAndPath(ctx, rule.Host, path.Path)
				} else {
					action = buildActionViaServiceAndServicePort(ctx, path.Backend.Service.Name, int(path.Backend.Service.Port.Number), 100)
					if v := annotations.GetStringAnnotationMutil(annotations.NginxCanary, annotations.AlbCanary, &ing); v != "true" {
						if canaryWeights, ok := carryWeight[rule.Host+"-"+path.Path]; ok {
							canaryWeight := 0
							for _, cw := range canaryWeights {
								canaryWeight += cw.Weight
							}
							for i := range action.ForwardConfig.ServerGroups {
								action.ForwardConfig.ServerGroups[i].Weight = 100 - canaryWeight
							}
							action.ForwardConfig.ServerGroups = append(action.ForwardConfig.ServerGroups, canaryWeights...)
						}
					}
				}

				conditions, err := t.buildRuleConditions(ctx, rule, path, ing)
				if err != nil {
					return errors.Wrapf(err, "ingress: %v", util.NamespacedName(&ing))
				}
				action2, err := t.buildAction(ctx, ing, action)
				if err != nil {
					return errors.Wrapf(err, "ingress: %v", util.NamespacedName(&ing))
				}
				lrs := alb.ListenerRuleSpec{
					ListenerID: lsID,
				}
				lrs.RuleActions = []alb.Action{action2}
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

func buildActionViaHostAndPath(_ context.Context, host, path string) alb.Action {
	action := alb.Action{
		Type: util.RuleActionTypeRedirect,
		RedirectConfig: &alb.RedirectConfig{
			Host:     host,
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

func (t *defaultModelBuildTask) buildPathPatternsForPrefixPathType(path string) ([]string, error) {
	if path == "/" {
		return []string{"/*"}, nil
	}
	if strings.ContainsAny(path, "*?") {
		return nil, errors.Errorf("prefix path shouldn't contain wildcards: %v", path)
	}

	normalizedPath := strings.TrimSuffix(path, "/")
	return []string{normalizedPath, normalizedPath + "/*"}, nil
}

func (t *defaultModelBuildTask) buildPathPatterns(path string, pathType *networking.PathType) ([]string, error) {
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
		return t.buildPathPatternsForPrefixPathType(path)
	default:
		return nil, errors.Errorf("unsupported pathType: %v", normalizedPathType)
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

func (t *defaultModelBuildTask) buildRuleConditions(ctx context.Context, rule networking.IngressRule,
	path networking.HTTPIngressPath, ing networking.Ingress) ([]alb.Condition, error) {
	var hosts []string
	if rule.Host != "" {
		hosts = append(hosts, rule.Host)
	}
	var paths []string
	if path.Path != "" {
		pathPatterns, err := t.buildPathPatterns(path.Path, path.PathType)
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
