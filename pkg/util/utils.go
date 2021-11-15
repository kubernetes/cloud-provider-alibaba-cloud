package util

import (
	"encoding/json"
	"fmt"
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

	least, _ := version.ParseGeneric(min)

	return runningVersion.AtLeast(least), nil
}
