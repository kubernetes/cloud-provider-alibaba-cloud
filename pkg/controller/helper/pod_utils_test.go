package helper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuildReadinessGatePodConditionTypeWithPrefix(t *testing.T) {
	prefix := "service.readiness.alibabacloud.com"
	id := "test-svc"
	expected := corev1.PodConditionType("service.readiness.alibabacloud.com/test-svc")
	result := BuildReadinessGatePodConditionTypeWithPrefix(prefix, id)
	assert.Equal(t, expected, result)
}

func TestIsPodHasReadinessGate(t *testing.T) {
	conditionType := "service.readiness.alibabacloud.com/test-svc"
	t.Run("Pod has readiness gate", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				ReadinessGates: []corev1.PodReadinessGate{
					{ConditionType: corev1.PodConditionType(conditionType)},
				},
			},
		}
		result := IsPodHasReadinessGate(pod, conditionType)
		assert.True(t, result)
	})

	t.Run("Pod does not have readiness gate", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				ReadinessGates: []corev1.PodReadinessGate{},
			},
		}
		result := IsPodHasReadinessGate(pod, conditionType)
		assert.False(t, result)
	})

	t.Run("Pod has multiple readiness gates", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				ReadinessGates: []corev1.PodReadinessGate{
					{ConditionType: "other-condition"},
					{ConditionType: corev1.PodConditionType(conditionType)},
					{ConditionType: "another-condition"},
				},
			},
		}
		result := IsPodHasReadinessGate(pod, conditionType)
		assert.True(t, result)
	})
}

func TestIsPodContainersReady(t *testing.T) {
	t.Run("Pod containers are ready", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.ContainersReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
		result := IsPodContainersReady(pod)
		assert.True(t, result)
	})

	t.Run("Pod containers are not ready", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.ContainersReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		}
		result := IsPodContainersReady(pod)
		assert.False(t, result)
	})

	t.Run("Pod has no ContainersReady condition", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
		result := IsPodContainersReady(pod)
		assert.False(t, result)
	})
}

func TestGetPodCondition(t *testing.T) {
	conditionType := corev1.PodReady
	condition := corev1.PodCondition{
		Type:   conditionType,
		Status: corev1.ConditionTrue,
	}

	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{condition},
		},
	}

	t.Run("Condition exists", func(t *testing.T) {
		result := GetPodCondition(pod, conditionType)
		assert.NotNil(t, result)
		assert.Equal(t, condition, *result)
	})

	t.Run("Condition does not exist", func(t *testing.T) {
		result := GetPodCondition(pod, corev1.ContainersReady)
		assert.Nil(t, result)
	})
}

func TestUpdatePodCondition(t *testing.T) {
	conditionType := corev1.PodConditionType("service.readiness.alibabacloud.com/test-svc")
	initialCondition := corev1.PodCondition{
		Type:   conditionType,
		Status: corev1.ConditionFalse,
	}
	newCondition := corev1.PodCondition{
		Type:    conditionType,
		Status:  corev1.ConditionTrue,
		Reason:  ConditionReasonServerRegistered,
		Message: ConditionMessageServerRegistered,
	}

	t.Run("Add new condition", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{},
			},
		}
		UpdatePodCondition(pod, newCondition)
		assert.Len(t, pod.Status.Conditions, 1)
		assert.Equal(t, newCondition, pod.Status.Conditions[0])
	})

	t.Run("Update existing condition", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{initialCondition},
			},
		}
		UpdatePodCondition(pod, newCondition)
		assert.Len(t, pod.Status.Conditions, 1)
		assert.Equal(t, newCondition, pod.Status.Conditions[0])
	})
}

func TestUpdateReadinessConditionForPod(t *testing.T) {
	conditionType := corev1.PodConditionType("service.readiness.alibabacloud.com/test-svc")
	reason := ConditionReasonServerRegistered
	message := ConditionMessageServerRegistered

	t.Run("Pod without readiness gate", func(t *testing.T) {
		pod := &corev1.Pod{}
		client := fake.NewClientBuilder().Build()
		err := UpdateReadinessConditionForPod(context.TODO(), client, pod, conditionType, corev1.ConditionTrue, reason, message)
		assert.NoError(t, err)
	})

	t.Run("Pod with readiness gate but condition already exists with same values", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				ReadinessGates: []corev1.PodReadinessGate{
					{ConditionType: conditionType},
				},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:    conditionType,
						Status:  corev1.ConditionTrue,
						Reason:  reason,
						Message: message,
					},
				},
			},
		}

		client := fake.NewClientBuilder().WithObjects(pod).Build()
		err := UpdateReadinessConditionForPod(context.TODO(), client, pod, conditionType, corev1.ConditionTrue, reason, message)
		assert.NoError(t, err)
	})

	t.Run("Pod with readiness gate and condition needs update", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				UID:       types.UID("test-uid"),
			},
			Spec: corev1.PodSpec{
				ReadinessGates: []corev1.PodReadinessGate{
					{ConditionType: conditionType},
				},
			},
		}

		client := fake.NewClientBuilder().WithObjects(pod).Build()
		err := UpdateReadinessConditionForPod(context.TODO(), client, pod, conditionType, corev1.ConditionTrue, reason, message)
		assert.NoError(t, err)
	})

	t.Run("Not found pod", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				UID:       types.UID("test-uid"),
			},
			Spec: corev1.PodSpec{
				ReadinessGates: []corev1.PodReadinessGate{
					{ConditionType: conditionType},
				},
			},
		}

		c := fake.NewClientBuilder().Build()
		err := c.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, &corev1.Pod{})
		assert.Error(t, err)
		assert.True(t, errors.IsNotFound(err))

		err = UpdateReadinessConditionForPod(context.TODO(), c, pod, conditionType, corev1.ConditionTrue, reason, message)
		assert.NoError(t, err)
	})

	t.Run("Pod with same status but different reason", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-reason",
				Namespace: "default",
				UID:       types.UID("test-uid-reason"),
			},
			Spec: corev1.PodSpec{
				ReadinessGates: []corev1.PodReadinessGate{
					{ConditionType: conditionType},
				},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:    conditionType,
						Status:  corev1.ConditionTrue,
						Reason:  "OldReason",
						Message: "old message",
					},
				},
			},
		}
		client := fake.NewClientBuilder().WithObjects(pod).Build()
		err := UpdateReadinessConditionForPod(context.TODO(), client, pod, conditionType, corev1.ConditionTrue, reason, message)
		assert.NoError(t, err)
	})
}

func TestBuildPodConditionPatch(t *testing.T) {
	t.Run("successfully create patch", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				UID:       types.UID("test-uid-123"),
			},
		}
		condition := corev1.PodCondition{
			Type:    corev1.PodConditionType("service.readiness.alibabacloud.com/test-svc"),
			Status:  corev1.ConditionTrue,
			Reason:  "TestReason",
			Message: "TestMessage",
		}

		patch, err := buildPodConditionPatch(pod, condition)
		assert.NoError(t, err)
		assert.NotNil(t, patch)
	})

	t.Run("create patch with different condition types", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-2",
				Namespace: "default",
				UID:       types.UID("test-uid-456"),
			},
		}
		condition := corev1.PodCondition{
			Type:    corev1.ContainersReady,
			Status:  corev1.ConditionFalse,
			Reason:  "ContainersNotReady",
			Message: "Some containers are not ready",
		}

		patch, err := buildPodConditionPatch(pod, condition)
		assert.NoError(t, err)
		assert.NotNil(t, patch)
	})

	t.Run("create patch with empty UID", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-3",
				Namespace: "default",
				UID:       "",
			},
		}
		condition := corev1.PodCondition{
			Type:   corev1.PodReady,
			Status: corev1.ConditionTrue,
		}

		patch, err := buildPodConditionPatch(pod, condition)
		assert.NoError(t, err)
		assert.NotNil(t, patch)
	})

	t.Run("create patch with all condition fields", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-4",
				Namespace: "default",
				UID:       types.UID("test-uid-789"),
			},
		}
		condition := corev1.PodCondition{
			Type:               corev1.PodConditionType("custom.condition"),
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "CustomReason",
			Message:            "Custom message with details",
		}

		patch, err := buildPodConditionPatch(pod, condition)
		assert.NoError(t, err)
		assert.NotNil(t, patch)
	})
}
