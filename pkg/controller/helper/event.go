package helper

const (
	// Service events
	ServiceEventReasonFailedAddFinalizer    = "FailedAddFinalizer"
	ServiceEventReasonFailedRemoveFinalizer = "FailedRemoveFinalizer"
	ServiceEventReasonFailedUpdateStatus    = "FailedUpdateStatus"
	ServiceEventReasonFailedDeployModel      = "FailedDeployModel"
	ServiceEventReasonFailedReconciled       = "FailedReconciled"
	ServiceEventReasonSuccessfullyReconciled = "SuccessfullyReconciled"
)
