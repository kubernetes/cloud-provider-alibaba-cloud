package utils

import (
	"fmt"
	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
)

func Logf(svc *v1.Service, format string, args ...interface{}) {
	prefix := ""
	if svc != nil {
		prefix = fmt.Sprintf("[%s/%s]", svc.Namespace, svc.Name)
	}
	format = fmt.Sprintf("%s %s", prefix, format)
	glog.Infof(format, args)
}
