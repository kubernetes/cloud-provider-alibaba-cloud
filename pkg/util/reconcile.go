package util

import (
	"errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ error = &ReconcileNeedRequeue{}

type ReconcileNeedRequeue struct {
	reason string
}

func (r *ReconcileNeedRequeue) Error() string {
	return r.reason
}

func NewReconcileNeedRequeue(reason string) *ReconcileNeedRequeue {
	return &ReconcileNeedRequeue{
		reason: reason,
	}
}

func HandleReconcileResult(request reconcile.Request, err error) (reconcile.Result, error) {
	if err == nil {
		return reconcile.Result{}, nil
	}

	var needRequeue *ReconcileNeedRequeue
	if errors.As(err, &needRequeue) {
		klog.Infof("[%s] requeue for next reconcile: %s", request, needRequeue.reason)
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, err
}
