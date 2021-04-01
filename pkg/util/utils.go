package util

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
)

const (
	NODE_MASTER_LABEL  = "node-role.kubernetes.io/master"
	BACKEND_TYPE_LABEL = "service.beta.kubernetes.io/backend-type"
	BACKEND_TYPE_ENI   = "eni"
	BACKEND_TYPE_ECS   = "ecs"

	ServiceAnnotationLoadBalancerRemoveUnscheduledBackend = "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend"
	// LabelNodeRoleExcludeNodeDeprecated specifies that the node should be exclude from CCM
	LabelNodeRoleExcludeNodeDeprecated = "service.beta.kubernetes.io/exclude-node"
	LabelNodeRoleExcludeNode           = "service.alibabacloud.com/exclude-node"
	// LabelNodeRoleExcludeBalancer specifies that the node should be
	// exclude from loadbalancers created by a cloud provider.
	LabelNodeRoleExcludeBalancer = "alpha.service-controller.kubernetes.io/exclude-balancer"
	LabelServiceHash             = "service.beta.kubernetes.io/hash"
	ECINodeLabel                 = "virtual-kubelet"
	ContextService               = "request.service"
	ContextRecorder              = "context.recorder"
)

func NodeIsMaster(node *v1.Node) bool {
	labels := node.Labels
	if labels == nil {
		return false
	}
	_, ok := labels[NODE_MASTER_LABEL]
	return ok
}

func GetNamePrefix(p string) string {
	// use the dash (if the name isn't too long) to make the pod name a bit prettier
	prefix := fmt.Sprintf("%s-", p)
	if len(apimachineryvalidation.NameIsDNSSubdomain(prefix, true)) != 0 {
		prefix = prefix
	}
	return prefix
}

func IsExcludedNode(node *v1.Node) bool {
	if node == nil || node.Labels == nil {
		return false
	}
	if _, exclude := node.Labels[LabelNodeRoleExcludeNodeDeprecated]; exclude {
		return true
	}
	if _, exclude := node.Labels[LabelNodeRoleExcludeNode]; exclude {
		return true
	}
	return false
}
