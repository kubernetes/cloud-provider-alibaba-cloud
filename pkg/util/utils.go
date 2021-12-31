package util

import (
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog"
)

func NamespacedName(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func Key(obj metav1.Object) string {
	return fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
}

func PrettyJson(object interface{}) string {
	b, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		fmt.Printf("ERROR: PrettyJson, %v\n %s\n", err, b)
	}
	return string(b)
}

// ClusterVersionAtLeast Check kubernetes version whether higher than the specific version
func ClusterVersionAtLeast(client *apiext.Clientset, min string) (bool, error) {
	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return false, fmt.Errorf("get server version: %s", err.Error())
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		return false, fmt.Errorf("unexpected error parsing running Kubernetes version, %s", err.Error())
	}
	klog.Infof("kubernetes version: %s", serverVersion.String())

	least, err := version.ParseGeneric(min)
	if err != nil {
		klog.Errorf("parse version %s error: %s", min, err.Error())
	}

	return runningVersion.AtLeast(least), nil
}

// MergeStringMap will merge multiple map[string]string into single one.
// The merge is executed for maps argument in sequential order, if a key already exists, the value from previous map is kept.
// e.g. MergeStringMap(map[string]string{"a": "1", "b": "2"}, map[string]string{"a": "3", "d": "4"}) == map[string]string{"a": "1", "b": "2", "d": "4"}
func MergeStringMap(maps ...map[string]string) map[string]string {
	ret := make(map[string]string)
	for _, _map := range maps {
		for k, v := range _map {
			if _, ok := ret[k]; !ok {
				ret[k] = v
			}
		}
	}
	return ret
}
func RetryImmediateOnError(interval time.Duration, timeout time.Duration, retryable func(error) bool, fn func() error) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		err := fn()
		if err != nil {
			if retryable(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
}
