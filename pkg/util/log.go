package util

import (
	"github.com/go-logr/logr"
	"k8s.io/klog/v2/klogr"
)

var (
	ServiceLog logr.Logger
	NLBLog     logr.Logger
	ELBLog     logr.Logger
)

func init() {
	ServiceLog = klogr.New().WithName("service-controller")
	NLBLog = klogr.New().WithName("nlb-controller")
	ELBLog = klogr.New().WithName("elb-controller")
}
