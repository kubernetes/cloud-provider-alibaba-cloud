package context

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
)

type RequestContext struct {
	Ctx      context.Context
	Service  *v1.Service
	Anno     *annotation.AnnotationRequest
	Log      logr.Logger
	Recorder record.EventRecorder
}
