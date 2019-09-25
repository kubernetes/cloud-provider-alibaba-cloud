package utils

import (
	"k8s.io/api/core/v1"
	"os"
)

const (
	BACKEND_TYPE_LABEL                                    = "service.beta.kubernetes.io/backend-type"
	BACKEND_TYPE_ENI                                      = "eni"
	BACKEND_TYPE_ECS                                      = "ecs"
	ServiceAnnotationLoadBalancerRemoveUnscheduledBackend = "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend"
	// LabelNodeRoleExcludeNode specifies that the node should be exclude from CCM
	LabelNodeRoleExcludeNode = "service.beta.kubernetes.io/exclude-node"
)

func IsENIBackendType(svc *v1.Service) bool {
	if os.Getenv("SERVICE_FORCE_BACKEND_ENI") == "true" {
		return true
	}
	return svc.Annotations[BACKEND_TYPE_LABEL] == BACKEND_TYPE_ENI
}
