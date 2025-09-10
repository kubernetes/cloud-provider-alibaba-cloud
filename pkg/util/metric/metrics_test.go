package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMsSince(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     float64
	}{
		{
			name:     "1 second",
			duration: 1 * time.Second,
			want:     1000,
		},
		{
			name:     "500 milliseconds",
			duration: 500 * time.Millisecond,
			want:     500,
		},
		{
			name:     "100 milliseconds",
			duration: 100 * time.Millisecond,
			want:     100,
		},
		{
			name:     "10 milliseconds",
			duration: 10 * time.Millisecond,
			want:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now().Add(-tt.duration)
			result := MsSince(start)
			// Allow small tolerance for timing variations
			assert.InDelta(t, tt.want, result, 50.0)
		})
	}
}

func TestMsSince_Zero(t *testing.T) {
	now := time.Now()
	result := MsSince(now)
	// Should be very close to 0
	assert.InDelta(t, 0.0, result, 10.0)
}

func TestMsSince_Future(t *testing.T) {
	future := time.Now().Add(1 * time.Second)
	result := MsSince(future)
	// Should be negative
	assert.Less(t, result, 0.0)
}

func TestUniqueServiceCnt(t *testing.T) {
	// Understanding the function logic:
	// - If uid NOT in map: returns 0 (line 86)
	// - If uid IS in map: stores it and returns 1 (lines 88-89)
	
	t.Run("new UID not in map", func(t *testing.T) {
		uid := "test-unique-new"
		result := UniqueServiceCnt(uid)
		assert.Equal(t, 0.0, result, "first call with new UID should return 0")
	})
	
	t.Run("existing UID in map", func(t *testing.T) {
		uid := "test-unique-existing"
		// Manually add to map first to test the other branch
		serviceUIDMap.Store(uid, true)
		
		result := UniqueServiceCnt(uid)
		assert.Equal(t, 1.0, result, "call with existing UID should return 1")
	})
}

func TestRegisterPrometheus(t *testing.T) {
	// This test just ensures the function doesn't panic
	// We can't easily test the actual registration without setting up a full prometheus registry
	assert.NotPanics(t, func() {
		RegisterPrometheus()
	})
}

func TestMetricVariables(t *testing.T) {
	// Test that metric variables are not nil
	assert.NotNil(t, NodeLatency)
	assert.NotNil(t, RouteLatency)
	assert.NotNil(t, SLBLatency)
	assert.NotNil(t, SLBOperationStatus)
}

func TestSLBTypeConstants(t *testing.T) {
	assert.Equal(t, "CLBType", CLBType)
	assert.Equal(t, "NLBType", NLBType)
	assert.Equal(t, "ALBType", ALBType)
}

func TestVerbConstants(t *testing.T) {
	assert.Equal(t, "Creation", string(VerbCreation))
	assert.Equal(t, "Deletion", string(VerbDeletion))
	assert.Equal(t, "Update", string(VerbUpdate))
}

func TestOperationResultConstants(t *testing.T) {
	assert.Equal(t, "Fail", string(ResultFail))
	assert.Equal(t, "Success", string(ResultSuccess))
}
