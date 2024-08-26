package helper

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Prefix for TargetHealth pod condition type.
const (
	TargetHealthPodConditionALBTypePrefix     = "target-health.alb.k8s.alicloud"
	TargetHealthPodConditionServiceTypePrefix = "service.readiness.alibabacloud.com"
)

const (
	ConditionReasonServerRegistered  = "ServerRegistered"
	ConditionMessageServerRegistered = "The backend has been added to the server group"
)

// BuildTargetHealthPodConditionType constructs the condition type for TargetHealth pod condition.
func BuildReadinessGatePodConditionType(cond string) corev1.PodConditionType {
	return corev1.PodConditionType(cond)
}

func BuildReadinessGatePodConditionTypeWithPrefix(prefix, id string) corev1.PodConditionType {
	return corev1.PodConditionType(fmt.Sprintf("%s/%s", prefix, id))
}

func IsPodHasReadinessGate(pod *corev1.Pod, cond string) bool {
	conditionType := BuildReadinessGatePodConditionType(cond)
	for _, rg := range pod.Spec.ReadinessGates {
		if rg.ConditionType == conditionType {
			return true
		}
	}
	return false
}

// IsPodContainersReady returns whether pod is containersReady.
func IsPodContainersReady(pod *corev1.Pod) bool {
	containersReadyCond := GetPodCondition(pod, corev1.ContainersReady)
	return containersReadyCond != nil && containersReadyCond.Status == corev1.ConditionTrue
}

// GetPodCondition will get pointer to Pod's existing condition.
// returns nil if no matching condition found.
func GetPodCondition(pod *corev1.Pod, conditionType corev1.PodConditionType) *corev1.PodCondition {
	for i := range pod.Status.Conditions {
		if pod.Status.Conditions[i].Type == conditionType {
			return &pod.Status.Conditions[i]
		}
	}
	return nil
}

// UpdatePodCondition will update Pod to contain specified condition.
func UpdatePodCondition(pod *corev1.Pod, condition corev1.PodCondition) {
	existingCond := GetPodCondition(pod, condition.Type)
	if existingCond != nil {
		*existingCond = condition
		return
	}
	pod.Status.Conditions = append(pod.Status.Conditions, condition)
}

func UpdateReadinessConditionForPod(ctx context.Context, kubeClient client.Client, pod *corev1.Pod, cond corev1.PodConditionType, reason, message string) error {
	if !IsPodHasReadinessGate(pod, string(cond)) {
		return nil
	}

	targetHealthCondStatus := corev1.ConditionTrue

	existedCond := GetPodCondition(pod, cond)
	// we skip patch pod if it matches current computed status/reason/message.
	if existedCond != nil &&
		existedCond.Status == targetHealthCondStatus &&
		existedCond.Reason == reason &&
		existedCond.Message == message {
		return nil
	}

	newTargetHealthCond := corev1.PodCondition{
		Type:    cond,
		Status:  targetHealthCondStatus,
		Reason:  reason,
		Message: message,
	}
	if existedCond == nil || existedCond.Status != targetHealthCondStatus {
		newTargetHealthCond.LastTransitionTime = metav1.Now()
	}

	patch, err := buildPodConditionPatch(pod, newTargetHealthCond)
	if err != nil {
		return err
	}
	k8sPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			UID:       pod.UID,
		},
	}
	if err := kubeClient.Status().Patch(ctx, k8sPod, patch); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func buildPodConditionPatch(pod *corev1.Pod, condition corev1.PodCondition) (client.Patch, error) {
	oldData, err := json.Marshal(corev1.Pod{
		Status: corev1.PodStatus{
			Conditions: nil,
		},
	})
	if err != nil {
		return nil, err
	}
	newData, err := json.Marshal(corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{UID: pod.UID}, // only put the uid in the new object to ensure it appears in the patch as a precondition
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{condition},
		},
	})
	if err != nil {
		return nil, err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Pod{})
	if err != nil {
		return nil, err
	}
	return client.RawPatch(types.StrategicMergePatchType, patchBytes), nil
}
