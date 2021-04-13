package service

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"strings"
)

const (
	// AnnotationPrefix prefix of service annotation
	AnnotationPrefix = "service.beta.kubernetes.io/alibaba-cloud"

	// AnnotationLegacyPrefix legacy prefix of service annotation
	AnnotationLegacyPrefix = "service.beta.kubernetes.io/alicloud"

	// parameters
	AddressType  = "address-type"
	BackendLabel = "backend-label"
)

var DefaultValue = map[string]string{
	composite(AnnotationPrefix, AddressType): string("internet"),
}

type AnnotationRequest struct{ svc *v1.Service }

// Get(AddressType)
func (n *AnnotationRequest) Get(k string) string {
	if n.svc == nil {
		klog.Infof("extract annotation %s from empty service", k)
		return ""
	}
	if n.svc.Annotations == nil {
		return ""
	}
	key := composite(AnnotationPrefix, k)
	v, ok := n.svc.Annotations[key]
	if ok {
		//
		return v
	}
	// todo: fix legacy key.  address-type to AddressType
	lkey := composite(AnnotationLegacyPrefix, k)
	v, ok = n.svc.Annotations[lkey]
	if ok {
		//
		return v
	}
	// return default
	return DefaultValue[key]
}

/*
	some help functions
*/
func composite(p, k string) string {
	return fmt.Sprintf("%s-%s", p, k)
}

func filterOutByLabel(nodes []v1.Node, labels string) ([]v1.Node, error) {
	if labels == "" {
		return nodes, nil
	}
	var result []v1.Node
	lbl := strings.Split(labels, ",")
	var records []string
	for _, node := range nodes {
		found := true
		for _, v := range lbl {
			l := strings.Split(v, "=")
			if len(l) < 2 {
				return []v1.Node{}, fmt.Errorf("parse backend label: %s, [k1=v1,k2=v2]", v)
			}
			if nv, exist := node.Labels[l[0]]; !exist || nv != l[1] {
				found = false
				break
			}
		}
		if found {
			result = append(result, node)
			records = append(records, node.Name)
		}
	}
	klog.V(4).Infof("accept nodes backend labels[%s], %v", labels, records)
	return result, nil
}

func key(svc *v1.Service) string {
	return fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)
}

func isLocalModeService(svc *v1.Service) bool {
	return svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeLocal
}

func IsENIBackendType(svc *v1.Service) bool {
	//if svc.Annotations[util.BACKEND_TYPE_LABEL] != "" {
	//	return svc.Annotations[util.BACKEND_TYPE_LABEL] == util.BACKEND_TYPE_ENI
	//}
	//
	//if os.Getenv("SERVICE_FORCE_BACKEND_ENI") != "" {
	//	return os.Getenv("SERVICE_FORCE_BACKEND_ENI") == "true"
	//}
	//
	//return cfg.Global.ServiceBackendType == util.BACKEND_TYPE_ENI
	return false
}

func isSLBNeeded(svc *v1.Service) bool {
	// finalizer must be supported
	return svc.DeletionTimestamp == nil && svc.Spec.Type == v1.ServiceTypeLoadBalancer
}
