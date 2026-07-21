package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildRuntimeOptions(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = BuildRuntimeOptions(RuntimeConfig{})
	})
}

func TestStripNodeFields(t *testing.T) {
	transform := stripNodeFields()

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			ManagedFields: []metav1.ManagedFieldsEntry{
				{Manager: "kubelet"},
			},
		},
		Status: v1.NodeStatus{
			Images: []v1.ContainerImage{
				{Names: []string{"nginx:latest"}, SizeBytes: 100},
				{Names: []string{"busybox:latest"}, SizeBytes: 50},
			},
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}

	result, err := transform(node)
	assert.NoError(t, err)

	stripped := result.(*v1.Node)
	assert.Nil(t, stripped.Status.Images)
	assert.Nil(t, stripped.ManagedFields)
	// preserved fields
	assert.Equal(t, "test-node", stripped.Name)
	assert.Len(t, stripped.Status.Conditions, 1)
}

func TestStripNodeFields_NonNodeObject(t *testing.T) {
	transform := stripNodeFields()

	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod"}}
	result, err := transform(pod)
	assert.NoError(t, err)
	assert.Equal(t, pod, result)
}

func TestStripPodFields(t *testing.T) {
	transform := stripPodFields()

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels:    map[string]string{"app": "nginx"},
			ManagedFields: []metav1.ManagedFieldsEntry{
				{Manager: "kubectl"},
			},
		},
		Spec: v1.PodSpec{
			NodeName: "node-1",
			Containers: []v1.Container{
				{
					Name:  "main",
					Image: "nginx",
					Env: []v1.EnvVar{
						{Name: "FOO", Value: "bar"},
					},
					VolumeMounts: []v1.VolumeMount{
						{Name: "data", MountPath: "/data"},
					},
				},
			},
			InitContainers: []v1.Container{
				{
					Name:  "init",
					Image: "busybox",
					Env: []v1.EnvVar{
						{Name: "INIT_VAR", Value: "val"},
					},
					VolumeMounts: []v1.VolumeMount{
						{Name: "config", MountPath: "/config"},
					},
				},
			},
			Volumes: []v1.Volume{
				{Name: "data"},
				{Name: "config"},
			},
		},
		Status: v1.PodStatus{
			PodIP: "10.0.0.1",
			Conditions: []v1.PodCondition{
				{Type: v1.PodReady, Status: v1.ConditionTrue},
			},
		},
	}

	result, err := transform(pod)
	assert.NoError(t, err)

	stripped := result.(*v1.Pod)
	assert.Nil(t, stripped.ManagedFields)
	assert.Nil(t, stripped.Spec.Containers[0].Env)
	assert.Nil(t, stripped.Spec.Containers[0].VolumeMounts)
	assert.Nil(t, stripped.Spec.InitContainers[0].Env)
	assert.Nil(t, stripped.Spec.InitContainers[0].VolumeMounts)
	assert.Nil(t, stripped.Spec.Volumes)
	// preserved fields
	assert.Equal(t, "test-pod", stripped.Name)
	assert.Equal(t, "node-1", stripped.Spec.NodeName)
	assert.Equal(t, "nginx", stripped.Spec.Containers[0].Image)
	assert.Equal(t, "10.0.0.1", stripped.Status.PodIP)
	assert.Equal(t, map[string]string{"app": "nginx"}, stripped.Labels)
}

func TestStripPodFields_NonPodObject(t *testing.T) {
	transform := stripPodFields()

	node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
	result, err := transform(node)
	assert.NoError(t, err)
	assert.Equal(t, node, result)
}
