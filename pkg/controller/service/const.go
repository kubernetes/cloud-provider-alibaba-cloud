package service

// Service related
const (
	ServiceFinalizer = "service.k8s.alibaba/resources"
	CCMClass         = "service.beta.kubernetes.io/class"
	LabelServiceHash = "service.beta.kubernetes.io/hash"
)

// Node related
const (
	LabelNodeTypeVK = "virtual-kubelet"
	// LabelNodeExcludeBalancer specifies that the node should be
	// exclude from loadbalancers created by a cloud provider.
	LabelNodeExcludeBalancer = "alpha.service-controller.kubernetes.io/exclude-balancer"
	// ToBeDeletedTaint is a taint used by the CLuster Autoscaler before marking a node for deletion.
	// Details in https://github.com/kubernetes/cloud-provider/blob/5bb9b27442bcb2613a9ca4046c89109de4435824/controllers/service/controller.go#L58
	ToBeDeletedTaint = "ToBeDeletedByClusterAutoscaler"
)

// Default value
const (
	DefaultServerWeight      = 100
	DefaultListenerBandwidth = -1
)
