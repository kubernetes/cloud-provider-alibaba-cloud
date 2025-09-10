package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildRuntimeOptions(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = BuildRuntimeOptions(RuntimeConfig{})
	})
}
