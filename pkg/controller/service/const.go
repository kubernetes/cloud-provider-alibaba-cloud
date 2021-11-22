package service

// Service related
const (
	ServiceFinalizer = "service.k8s.alibaba/resources"
	CCMClass         = "service.beta.kubernetes.io/class"
	LabelServiceHash = "service.beta.kubernetes.io/hash"
)

// Node related
const (
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"
	LabelNodeTypeVK     = "virtual-kubelet"
	// LabelNodeExcludeBalancer specifies that the node should be
	// exclude from loadbalancers created by a cloud provider.
	LabelNodeExcludeBalancer = "alpha.service-controller.kubernetes.io/exclude-balancer"
)

// Default value
const (
	DefaultServerWeight      = 100
	DefaultListenerBandwidth = -1
)
