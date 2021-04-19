package pvtz

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func serviceRr(svc *corev1.Service) string {
	return fmt.Sprintf("%s.%s.svc", svc.Name, svc.Namespace)
}

func serviceRrByName(svcName types.NamespacedName) string {
	return fmt.Sprintf("%s.%s.svc", svcName.Name, svcName.Namespace)
}