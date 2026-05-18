package nlbv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapFull(t *testing.T) {
	assert.Equal(t, mapfull(), true)

	initial.Store("default/test-svc", 0)
	assert.Equal(t, mapfull(), false)

	initial.Store("default/test-svc", 1)
	assert.Equal(t, mapfull(), true)
}

//func TestInitMap(t *testing.T) {
//	initMap(getFakeKubeClient())
//	_, ok := initial.Load(fmt.Sprintf("%s/%s", NS, SvcName))
//	assert.Equal(t, ok, true)
//}
