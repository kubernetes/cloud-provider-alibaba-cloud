package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

// Prefix for TargetHealth pod condition type.
const TargetHealthPodConditionTypePrefix = "target-health.alb.k8s.alicloud"

// BuildTargetHealthPodConditionType constructs the condition type for TargetHealth pod condition.
func BuildReadinessGatePodConditionType() corev1.PodConditionType {
	return corev1.PodConditionType(TargetHealthPodConditionTypePrefix)
}

func IsPodHasReadinessGate(pod *corev1.Pod) bool {
	conditionType := BuildReadinessGatePodConditionType()
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
