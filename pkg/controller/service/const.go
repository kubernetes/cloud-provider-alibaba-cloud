package service

const (
	serviceFinalizer = "service.k8s.alibaba/resources"

	CCM_CLASS = "service.beta.kubernetes.io/class"


	LabelNodeRoleMaster = "node-role.kubernetes.io/master"

	DEFAULT_SERVER_WEIGHT      = 100
	DEFAULT_LISTENER_BANDWIDTH = -1

	BACKEND_TYPE_LABEL                                    = "service.beta.kubernetes.io/backend-type"
	BACKEND_TYPE_ENI                                      = "eni"
	BACKEND_TYPE_ECS                                      = "ecs"
	ServiceAnnotationLoadBalancerRemoveUnscheduledBackend = "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend"

	// LabelNodeRoleExcludeBalancer specifies that the node should be
	// exclude from loadbalancers created by a cloud provider.
	LabelNodeRoleExcludeBalancer = "alpha.service-controller.kubernetes.io/exclude-balancer"
	LabelServiceHash             = "service.beta.kubernetes.io/hash"
	ECINodeLabel                 = "virtual-kubelet"
)
