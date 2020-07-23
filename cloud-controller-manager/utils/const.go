package utils

import (
	"k8s.io/api/core/v1"
	"os"
)

type contextKey string

const (
	BACKEND_TYPE_LABEL                                    = "service.beta.kubernetes.io/backend-type"
	BACKEND_TYPE_ENI                                      = "eni"
	BACKEND_TYPE_ECS                                      = "ecs"
	ServiceAnnotationLoadBalancerRemoveUnscheduledBackend = "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend"
	// LabelNodeRoleExcludeNode specifies that the node should be exclude from CCM
	LabelNodeRoleExcludeNode = "service.beta.kubernetes.io/exclude-node"
	// LabelNodeRoleExcludeBalancer specifies that the node should be
	// exclude from loadbalancers created by a cloud provider.
	LabelNodeRoleExcludeBalancer            = "alpha.service-controller.kubernetes.io/exclude-balancer"
	ECINodeLabel                            = "virtual-kubelet"
	ContextService               contextKey = "request.service"
)

func IsENIBackendType(svc *v1.Service) bool {
	if svc.Annotations[BACKEND_TYPE_LABEL] != "" {
		return svc.Annotations[BACKEND_TYPE_LABEL] == BACKEND_TYPE_ENI
	}

	if os.Getenv("SERVICE_FORCE_BACKEND_ENI") == "true" {
		return true
	}

	return false
}
