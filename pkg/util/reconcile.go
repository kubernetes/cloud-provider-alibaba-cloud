package util

import (
	"errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var _ error = &ReconcileNeedRequeue{}

const (
	defaultRequeueAfter = time.Second * 10
)

type ReconcileNeedRequeue struct {
	reason string
	after  time.Duration
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
		after := defaultRequeueAfter
		if needRequeue.after > 0 {
			after = needRequeue.after
		}
		return reconcile.Result{RequeueAfter: after}, nil
	}

	return reconcile.Result{}, err
}
