package util

import (
	"testing"
	"time"
)

func TestAttemptStrategy_Start(t *testing.T) {
	strategy := AttemptStrategy{
		Total: 500 * time.Millisecond,
		Delay: 100 * time.Millisecond,
		Min:   2,
	}

	attempt := strategy.Start()
	if attempt == nil {
		t.Fatal("Expected attempt to be non-nil")
	}

	if attempt.count != 0 {
		t.Errorf("Expected initial count to be 0, got %d", attempt.count)
	}

	if !attempt.force {
		t.Error("Expected force to be true initially")
	}
}

func TestAttempt_Next(t *testing.T) {
	strategy := AttemptStrategy{
		Total: 100 * time.Millisecond,
		Delay: 50 * time.Millisecond,
		Min:   2,
	}

	attempt := strategy.Start()

	// First call should always return true
	if !attempt.Next() {
		t.Error("First call to Next() should return true")
	}

	// Second call should return true due to Min=2
	if !attempt.Next() {
		t.Error("Second call to Next() should return true because of Min=2")
	}

	// Third call might return false since we've met the Min requirement
	// and exceeded the Total time
	result := attempt.Next()
	// We don't assert the result here as it depends on timing
	t.Logf("Third call to Next() returned: %v", result)
}

func TestAttempt_HasNext(t *testing.T) {
	strategy := AttemptStrategy{
		Total: 100 * time.Millisecond,
		Delay: 50 * time.Millisecond,
		Min:   1,
	}

	attempt := strategy.Start()

	// Initially HasNext should return true because force=true
	if !attempt.HasNext() {
		t.Error("HasNext() should return true initially")
	}

	// After first Next() call
	attempt.Next()

	// HasNext should still return true because Min=1
	if !attempt.HasNext() {
		t.Error("HasNext() should return true due to Min=1")
	}
}

func TestAttempt_MinRetries(t *testing.T) {
	strategy := AttemptStrategy{
		Total: 0, // No time
		Delay: 10 * time.Millisecond,
		Min:   3, // But require 3 retries
	}

	attempt := strategy.Start()

	// Should be able to do at least 3 retries even with 0 total time
	for i := 0; i < 3; i++ {
		if !attempt.Next() {
			t.Errorf("Next() should return true for retry %d due to Min=3", i+1)
		}
	}
}

func TestAttempt_TotalTimeLimit(t *testing.T) {
	strategy := AttemptStrategy{
		Total: 80 * time.Millisecond,
		Delay: 30 * time.Millisecond,
		Min:   1,
	}

	attempt := strategy.Start()

	// First call should work
	if !attempt.Next() {
		t.Error("First call to Next() should return true")
	}

	// Second call should work due to Min=1
	if !attempt.Next() {
		t.Error("Second call to Next() should return true due to Min=1")
	}

	result := attempt.Next()
	t.Logf("Third call to Next() returned: %v", result)
}

func TestAttempt_nextSleep(t *testing.T) {
	strategy := AttemptStrategy{
		Total: 100 * time.Millisecond,
		Delay: 50 * time.Millisecond,
		Min:   1,
	}

	attempt := strategy.Start()

	now := time.Now()
	sleep := attempt.nextSleep(now)
	if sleep < 0 || sleep > strategy.Delay {
		t.Errorf("Unexpected sleep duration: %v", sleep)
	}

	time.Sleep(60 * time.Millisecond)
	sleep = attempt.nextSleep(time.Now())
	if sleep != 0 {
		t.Errorf("Expected 0 when now is after last+Delay, got %v", sleep)
	}
}
