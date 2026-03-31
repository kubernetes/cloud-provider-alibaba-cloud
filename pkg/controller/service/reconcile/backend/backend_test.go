package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	sum := 0
	addFunc := func(a []int) error {
		for _, num := range a {
			sum += num
		}
		return nil
	}
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if err := Batch(nums, 3, addFunc); err != nil {
		t.Fatalf("Batch error: %s", err.Error())
	}
	assert.Equal(t, sum, 55)
}
