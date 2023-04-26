package elb

// Attributes
const (
	ELBClass          = "alibabacloud.com/elb"
	ELBFinalizer      = "service.k8s.alibaba/elb"
	BaseBackendWeight = "base"
	EipPrefix         = "acs"
)

// Status
const (
	//ELB
	ELBActive   = "Active"
	ELBInActive = "InActive"

	//EIP
	EipAvailable = "Available"
	EipInUse     = "InUse"

	//ServerGroups
	ENSRunning = "Running"
	ENSStopped = "Stopped"
	ENSExpired = "Expired"

	//Listener
	ListenerRunning     = "Running"
	ListenerStopped     = "Stopped"
	ListenerStarting    = "Starting"
	ListenerConfiguring = "Configuring"
	ListenerStopping    = "Stopping"
)

const (
	InstanceNotFound     = "find no"
	StatusAberrant       = "aberrant"
	ENSBatchAddMaxNumber = 19
)
