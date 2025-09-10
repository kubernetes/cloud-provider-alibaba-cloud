package controller

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func TestAddToManager(t *testing.T) {
	originalMap := controllerMap
	defer func() {
		controllerMap = originalMap
	}()

	t.Run("success with multiple controllers", func(t *testing.T) {
		called := make([]string, 0, 2)
		controllerMap = map[string]func(manager.Manager, *shared.SharedContext) error{
			"node": func(manager.Manager, *shared.SharedContext) error {
				called = append(called, "node")
				return nil
			},
			"route": func(manager.Manager, *shared.SharedContext) error {
				called = append(called, "route")
				return nil
			},
		}

		err := AddToManager(nil, nil, []string{"node", "route"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"node", "route"}, called)
	})

	t.Run("unknown controller returns error", func(t *testing.T) {
		controllerMap = map[string]func(manager.Manager, *shared.SharedContext) error{}

		err := AddToManager(nil, nil, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find controller unknown")
	})

	t.Run("controller add failed returns error", func(t *testing.T) {
		controllerMap = map[string]func(manager.Manager, *shared.SharedContext) error{
			"node": func(manager.Manager, *shared.SharedContext) error {
				return errors.New("add node failed")
			},
		}

		err := AddToManager(nil, nil, []string{"node"})
		assert.EqualError(t, err, "add node failed")
	})
}
