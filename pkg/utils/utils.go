package utils

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
)

const (
	NODE_MASTER_LABEL = "node-role.kubernetes.io/master"
)

func NodeIsMaster(node *corev1.Node) bool {
	labels := node.Labels
	if labels == nil {
		return false
	}
	_, ok := labels[NODE_MASTER_LABEL]
	return ok
}

func GetNamePrefix(p string) string {
	// use the dash (if the name isn't too long) to make the pod name a bit prettier
	prefix := fmt.Sprintf("%s-", p)
	if len(apimachineryvalidation.NameIsDNSSubdomain(prefix, true)) != 0 {
		prefix = prefix
	}
	return prefix
}
