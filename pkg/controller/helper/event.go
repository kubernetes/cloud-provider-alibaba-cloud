package helper

// ServiceEventReason
const (
	FailedAddFinalizer    = "FailedAddFinalizer"
	FailedRemoveFinalizer = "FailedRemoveFinalizer"
	FailedAddHash         = "FailedAddHash"
	FailedRemoveHash      = "FailedRemoveHash"
	FailedUpdateStatus    = "FailedUpdateStatus"
	UnAvailableBackends   = "UnAvailableLoadBalancer"
	FailedSyncLB          = "SyncLoadBalancerFailed"
	SucceedDeleteLB       = "DeletedLoadBalancer"
	FailedDeleteLB        = "DeleteLoadBalancerFailed"
	SucceedSyncLB         = "EnsuredLoadBalancer"
	AnnoChanged           = "AnnotationChanged"
	SpecChanged           = "ServiceSpecChanged"
)

//NodeEventReason
const (
	FailedDeleteNode  = "DeleteNodeFailed"
	FailedAddonNode   = "AddNodeFailed"
	FailedSyncNode    = "SyncNodeFailed"
	SucceedDeleteNode = "DeletedNode"
	InitializedNode   = "InitializedNode"
)

//RouteEventReason
const (
	FailedCreateRoute  = "CreateRouteFailed"
	FailedSyncRoute    = "SyncRouteFailed"
	SucceedCreateRoute = "CreatedRoute"
)
