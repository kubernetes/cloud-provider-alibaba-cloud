package util

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestNewReconcileNeedRequeue(t *testing.T) {
	reason := "test reason"
	err := NewReconcileNeedRequeue(reason)
	assert.Equal(t, err.reason, reason)
	assert.Equal(t, err.Error(), reason)
}

func TestHandleReconcileResult(t *testing.T) {
	t.Run("need requeue", func(t *testing.T) {
		originalErr := NewReconcileNeedRequeue("test reason")
		r, err := HandleReconcileResult(reconcile.Request{}, originalErr)
		assert.NoError(t, err)
		assert.Zero(t, r.Requeue)
		assert.Equal(t, r.RequeueAfter, defaultRequeueAfter)
	})

	t.Run("need requeue with retry after", func(t *testing.T) {
		after := time.Second * 200
		assert.NotEqual(t, after, defaultRequeueAfter)
		originalErr := NewReconcileNeedRequeue("test reason")
		originalErr.after = after
		r, err := HandleReconcileResult(reconcile.Request{}, originalErr)
		assert.NoError(t, err)
		assert.Zero(t, r.Requeue)
		assert.Equal(t, r.RequeueAfter, after)
	})

	t.Run("nil error", func(t *testing.T) {
		r, err := HandleReconcileResult(reconcile.Request{}, nil)
		assert.NoError(t, err)
		assert.Zero(t, r.Requeue)
		assert.Zero(t, r.RequeueAfter)
	})

	t.Run("normal error", func(t *testing.T) {
		originalErr := errors.New("test error")
		r, err := HandleReconcileResult(reconcile.Request{}, originalErr)
		assert.Error(t, err)
		assert.Zero(t, r.Requeue)
		assert.Zero(t, r.RequeueAfter)
	})
}
