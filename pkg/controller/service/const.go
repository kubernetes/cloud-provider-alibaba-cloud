package service

const (
	serviceFinalizer             = "service.k8s.alibaba/resources"
	LabelServiceHash             = "service.beta.kubernetes.io/hash"
	CCM_CLASS                    = "service.beta.kubernetes.io/class"
	DEFAULT_SERVER_WEIGHT        = 100
	LabelNodeRoleExcludeBalancer = "alpha.service-controller.kubernetes.io/exclude-balancer"
	ModificationProtectionReason = "managed.by.ack"
)

const (
	BACKEND_TYPE_ENI = "eni"
	BACKEND_TYPE_ECS = "ecs"

)
