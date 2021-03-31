package pvtz

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func fullServiceName(svc *corev1.Service) string {
	return fmt.Sprintf("%s/%ss", svc.Namespace, svc.Name)
}

func serviceRr(svc *corev1.Service) string {
	return fmt.Sprintf("%s.%s.svc", svc.Name, svc.Namespace)
}
