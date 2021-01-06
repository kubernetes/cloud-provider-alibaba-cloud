package utils

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
	LabelServiceHash                        = "service.beta.kubernetes.io/hash"
	ECINodeLabel                            = "virtual-kubelet"
	ContextService               contextKey = "request.service"
	ContextRecorder              contextKey = "context.recorder"
)