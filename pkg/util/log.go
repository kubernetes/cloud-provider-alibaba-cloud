package util

import (
	"k8s.io/klog"
)

func NewReqLog(prefix string) *Log {
	return &Log{
		prefix: prefix,
	}
}

type Log struct {
	prefix string
}

func (log *Log) Infof(format string, args ...interface{}) {
	klog.Infof(log.prefix+format, args...)
}

func (log *Log) Warningf(format string, args ...interface{}) {
	klog.Warningf(log.prefix+format, args...)
}
