package util

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// LabelNodeRoleExcludeNodeDeprecated specifies that the node should be exclude from CCM
	LabelNodeRoleExcludeNodeDeprecated = "service.beta.kubernetes.io/exclude-node"
	LabelNodeRoleExcludeNode           = "service.alibabacloud.com/exclude-node"
)

func GetNamePrefix(p string) string {
	// use the dash (if the name isn't too long) to make the pod name a bit prettier
	prefix := fmt.Sprintf("%s-", p)
	if len(apimachineryvalidation.NameIsDNSSubdomain(prefix, true)) != 0 {
		prefix = prefix
	}
	return prefix
}

func IsExcludedNode(node *v1.Node) bool {
	if node == nil || node.Labels == nil {
		return false
	}
	if _, exclude := node.Labels[LabelNodeRoleExcludeNodeDeprecated]; exclude {
		return true
	}
	if _, exclude := node.Labels[LabelNodeRoleExcludeNode]; exclude {
		return true
	}
	return false
}

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
