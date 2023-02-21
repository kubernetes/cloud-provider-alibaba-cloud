package common

import networkingv1 "k8s.io/api/networking/v1"

// Rule func meaning
// DefaultRule--创建一条可以指定path或host的rule，path为空时，默认为/default-rule
type Rule struct {
}

func (*Rule) DefaultRule(path, host, serviceName string) networkingv1.IngressRule {
	if path == "" {
		path = "/default-rule"
	}
	exact := networkingv1.PathTypeExact
	ret := networkingv1.IngressRule{
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     path,
						PathType: &exact,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: serviceName,
								Port: networkingv1.ServiceBackendPort{
									Number: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	if host != "" {
		ret.Host = host
	}
	return ret
}

func (*Rule) DefaultUseAnnotationRule(path, serviceName string) networkingv1.IngressRule {
	if path == "" {
		path = "/default-rule"
	}
	exact := networkingv1.PathTypeExact
	return networkingv1.IngressRule{
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     path,
						PathType: &exact,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: serviceName,
								Port: networkingv1.ServiceBackendPort{
									Name: "use-annotation",
								},
							},
						},
					},
				},
			},
		},
	}
}
