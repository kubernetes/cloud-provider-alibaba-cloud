package utils

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func Logf(svc *v1.Service, format string, args ...interface{}) {
	prefix := ""
	if svc != nil {
		prefix = fmt.Sprintf("[%s/%s]", svc.Namespace, svc.Name)
	}
	klog.Infof(prefix+format, args...)
}
