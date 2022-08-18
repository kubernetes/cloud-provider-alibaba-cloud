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

// NodeEventReason
const (
	FailedDeleteNode  = "DeleteNodeFailed"
	FailedAddNode     = "AddNodeFailed"
	FailedSyncNode    = "SyncNodeFailed"
	SucceedDeleteNode = "DeletedNode"
	InitializedNode   = "InitializedNode"
)

// RouteEventReason
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
	sub := re.FindStringSubmatch(err.Error())
	if len(sub) > 1 {
		message = sub[1]
	} else {
		message = err.Error()
	}
	return message
}

const (
	// Ingress events
	IngressEventReasonFailedAddFinalizer     = "FailedAddFinalizer"
	IngressEventReasonFailedRemoveFinalizer  = "FailedRemoveFinalizer"
	IngressEventReasonFailedUpdateStatus     = "FailedUpdateStatus"
	IngressEventReasonFailedBuildModel       = "FailedBuildModel"
	IngressEventReasonFailedApplyModel       = "FailedApplyModel"
	IngressEventReasonSuccessfullyReconciled = "SuccessfullyReconciled"
)

// EventType type of event associated with an informer
type EventType string

const (
	// CreateEvent event associated with new objects in an informer
	CreateEvent EventType = "CREATE"
	// UpdateEvent event associated with an object update in an informer
	UpdateEvent EventType = "UPDATE"
	// DeleteEvent event associated when an object is removed from an informer
	IngressDeleteEvent EventType = "DELETE"
	// ConfigurationEvent event associated when a controller configuration object is created or updated
	ConfigurationEvent EventType = "CONFIGURATION"

	// NodeEvent event associated when a controller configuration object is created or updated
	NodeEvent EventType = "NODE"

	// ServiceEvent event associated when a controller configuration object is created or updated
	ServiceEvent EventType = "SERVICE"

	// EndPointEvent event associated when a controller configuration object is created or updated
	EndPointEvent EventType = "ENDPOINT"

	// IngressEvent event associated when a controller configuration object is created or updated
	IngressEvent EventType = "Ingress"
)

// Event holds the context of an event.
type Event struct {
	Type EventType
	Obj  interface{}
}
