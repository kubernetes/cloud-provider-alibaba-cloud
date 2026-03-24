package helper

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

func TestGetLogMessage(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.Equal(t, "", GetLogMessage(nil))
	})

	t.Run("simple error", func(t *testing.T) {
		err := errors.New("simple error message")
		assert.Equal(t, "simple error message", GetLogMessage(err))
	})

	t.Run("aggregate error with multiple errors", func(t *testing.T) {
		err1 := errors.New("first error")
		err2 := errors.New("second error")
		agg := utilerrors.NewAggregate([]error{err1, err2})
		assert.Equal(t, "first error", GetLogMessage(agg))
	})

	t.Run("aggregate error with no errors", func(t *testing.T) {
		agg := utilerrors.NewAggregate([]error{})
		assert.Equal(t, "", GetLogMessage(agg))
	})

	t.Run("sdk error with message", func(t *testing.T) {
		err := errors.New("some error [SDKError] Message: SDK error message content, more details")
		assert.Equal(t, "Message: SDK error message content, more details", GetLogMessage(err))
	})

	t.Run("sdk error without message pattern", func(t *testing.T) {
		err := errors.New("some error [SDKError] other error content")
		assert.Equal(t, "some error [SDKError] other error content", GetLogMessage(err))
	})

	t.Run("error without sdk pattern", func(t *testing.T) {
		err := errors.New("regular error without SDK tag")
		assert.Equal(t, "regular error without SDK tag", GetLogMessage(err))
	})
}