package parallel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoPiece(t *testing.T) {
	// Test basic functionality
	t.Run("basic functionality", func(t *testing.T) {
		results := make([]int, 10)
		ctx := context.Background()

		DoPiece(ctx, 3, 10, func(i int) {
			results[i] = i * 2
		})

		for i := 0; i < 10; i++ {
			assert.Equal(t, i*2, results[i], "Expected results[%d] to be %d, but got %d", i, i*2, results[i])
		}
	})

	// Test with default worker count
	t.Run("default worker count", func(t *testing.T) {
		results := make([]bool, 5)
		ctx := context.Background()

		DoPiece(ctx, 0, 5, func(i int) {
			results[i] = true
		})

		for i := 0; i < 5; i++ {
			assert.True(t, results[i], "Expected results[%d] to be true, but got false", i)
		}
	})

	// Test with context cancellation
	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		DoPiece(ctx, 2, 1000, func(i int) {
			time.Sleep(1 * time.Millisecond)
		})

		// Should complete quickly due to context cancellation
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Error("Expected context to be cancelled")
		}
	})
}

func TestDo(t *testing.T) {
	// Test successful execution
	t.Run("successful execution", func(t *testing.T) {
		calls := make([]int, 3)

		err := Do(
			func() error { calls[0] = 1; return nil },
			func() error { calls[1] = 2; return nil },
			func() error { calls[2] = 3; return nil },
		)

		assert.NoError(t, err, "Expected no error, but got: %v", err)

		for i := 0; i < 3; i++ {
			assert.Equal(t, i+1, calls[i], "Expected calls[%d] to be %d, but got %d", i, i+1, calls[i])
		}
	})

	// Test error handling
	t.Run("error handling", func(t *testing.T) {
		expectedErr1 := errors.New("error 1")
		expectedErr2 := errors.New("error 2")

		err := Do(
			func() error { return nil },
			func() error { return expectedErr1 },
			func() error { return nil },
			func() error { return expectedErr2 },
		)

		assert.Error(t, err, "Expected error, but got nil")

		errs := err.(interface{ Errors() []error }).Errors()
		assert.Len(t, errs, 2, "Expected 2 errors, but got %d", len(errs))
	})

	// Test empty function list
	t.Run("empty function list", func(t *testing.T) {
		err := Do()
		assert.NoError(t, err, "Expected no error for empty function list, but got: %v", err)
	})

	// Test single function
	t.Run("single function", func(t *testing.T) {
		var called bool
		expectedErr := errors.New("single function error")

		err := Do(func() error { called = true; return expectedErr })

		assert.True(t, called, "Expected function to be called")
		assert.Error(t, err, "Expected error, but got nil")
		assert.Equal(t, expectedErr.Error(), err.Error(), "Expected error %v, but got %v", expectedErr, err)
	})
}
