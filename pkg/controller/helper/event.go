package helper

import "regexp"

// ServiceEventReason
const (
	FailedAddFinalizer     = "FailedAddFinalizer"
	FailedRemoveFinalizer  = "FailedRemoveFinalizer"
	FailedAddHash          = "FailedAddHash"
	FailedRemoveHash       = "FailedRemoveHash"
	FailedUpdateStatus     = "FailedUpdateStatus"
	UnAvailableBackends    = "UnAvailableLoadBalancer"
	FailedSyncLB           = "SyncLoadBalancerFailed"
	SucceedCleanLB         = "CleanLoadBalancer"
	FailedCleanLB          = "CleanLoadBalancerFailed"
	SucceedSyncLB          = "EnsuredLoadBalancer"
	AnnoChanged            = "AnnotationChanged"
	TypeChanged            = "TypeChanged"
	SpecChanged            = "ServiceSpecChanged"
	DeleteTimestampChanged = "DeleteTimestampChanged"
)

//NodeEventReason
const (
	FailedDeleteNode  = "DeleteNodeFailed"
	FailedAddNode     = "AddNodeFailed"
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

var re = regexp.MustCompile(".*(Message:.*)")

func GetLogMessage(err error) string {
	if err == nil {
		return ""
	}
	var message string
	sub := re.FindSubmatch([]byte(err.Error()))
	if len(sub) > 1 {
		message = string(sub[1])
	} else {
		message = err.Error()
	}
	return message
}
