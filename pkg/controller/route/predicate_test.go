package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestPredicateForNodeEvent_Create(t *testing.T) {
	predicate := &predicateForNodeEvent{}

	t.Run("node with empty PodCIDR", func(t *testing.T) {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
			Spec: corev1.NodeSpec{
				PodCIDR: "", // Empty PodCIDR
			},
		}

		createEvent := event.CreateEvent{
			Object: node,
		}

		result := predicate.Create(createEvent)
		assert.False(t, result, "Expected Create to return false for node with empty PodCIDR")
	})

	t.Run("node with valid PodCIDR", func(t *testing.T) {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
			Spec: corev1.NodeSpec{
				PodCIDR: "10.0.0.0/24",
			},
		}

		createEvent := event.CreateEvent{
			Object: node,
		}

		result := predicate.Create(createEvent)
		assert.True(t, result, "Expected Create to return true for node with valid PodCIDR")
	})

	t.Run("non-node object", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
		}

		createEvent := event.CreateEvent{
			Object: pod,
		}

		result := predicate.Create(createEvent)
		assert.True(t, result, "Expected Create to return true for non-node objects")
	})
}

func TestPredicateForNodeEvent_Update(t *testing.T) {
	predicate := &predicateForNodeEvent{}

	t.Run("node UID changed", func(t *testing.T) {
		oldNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				UID:  "uid-1",
			},
			Spec: corev1.NodeSpec{
				PodCIDR: "10.0.0.0/24",
			},
		}

		newNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				UID:  "uid-2", // Changed UID
			},
			Spec: corev1.NodeSpec{
				PodCIDR: "10.0.0.0/24",
			},
		}

		updateEvent := event.UpdateEvent{
			ObjectOld: oldNode,
			ObjectNew: newNode,
		}

		result := predicate.Update(updateEvent)
		assert.True(t, result, "Expected Update to return true when node UID changes")
	})

	t.Run("node PodCIDR changed", func(t *testing.T) {
		oldNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				UID:  "uid-1",
			},
			Spec: corev1.NodeSpec{
				PodCIDR: "10.0.0.0/24",
			},
		}

		newNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				UID:  "uid-1",
			},
			Spec: corev1.NodeSpec{
				PodCIDR: "10.0.1.0/24", // Changed PodCIDR
			},
		}

		updateEvent := event.UpdateEvent{
			ObjectOld: oldNode,
			ObjectNew: newNode,
		}

		result := predicate.Update(updateEvent)
		assert.True(t, result, "Expected Update to return true when node PodCIDR changes")
	})

	t.Run("node PodCIDRs changed", func(t *testing.T) {
		oldNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				UID:  "uid-1",
			},
			Spec: corev1.NodeSpec{
				PodCIDR:  "10.0.0.0/24",
				PodCIDRs: []string{"10.0.0.0/24"},
			},
		}

		newNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				UID:  "uid-1",
			},
			Spec: corev1.NodeSpec{
				PodCIDR:  "10.0.0.0/24",
				PodCIDRs: []string{"10.0.1.0/24"}, // Changed PodCIDRs
			},
		}

		updateEvent := event.UpdateEvent{
			ObjectOld: oldNode,
			ObjectNew: newNode,
		}

		result := predicate.Update(updateEvent)
		assert.True(t, result, "Expected Update to return true when node PodCIDRs changes")
	})

	t.Run("node unchanged", func(t *testing.T) {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				UID:  "uid-1",
			},
			Spec: corev1.NodeSpec{
				PodCIDR:  "10.0.0.0/24",
				PodCIDRs: []string{"10.0.0.0/24"},
			},
		}

		updateEvent := event.UpdateEvent{
			ObjectOld: node,
			ObjectNew: node,
		}

		result := predicate.Update(updateEvent)
		assert.False(t, result, "Expected Update to return false when node is unchanged")
	})

	t.Run("non-node objects", func(t *testing.T) {
		oldPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "old-pod",
				Namespace: "default",
			},
		}

		newPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "new-pod",
				Namespace: "default",
			},
		}

		updateEvent := event.UpdateEvent{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}

		result := predicate.Update(updateEvent)
		assert.True(t, result, "Expected Update to return true for non-node objects")
	})
}