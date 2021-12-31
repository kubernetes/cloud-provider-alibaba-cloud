package ingress

import (
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
)

func NewEnqueueRequestsForAlbconfigEvent(k8sClient client.Client, eventRecorder record.EventRecorder, logger logr.Logger) *enqueueRequestsForAlbconfigEvent {
	return &enqueueRequestsForAlbconfigEvent{
		k8sClient:     k8sClient,
		eventRecorder: eventRecorder,
		logger:        logger,
	}
}

var _ handler.EventHandler = (*enqueueRequestsForAlbconfigEvent)(nil)

type enqueueRequestsForAlbconfigEvent struct {
	ingressEventChan chan<- event.GenericEvent
	k8sClient        client.Client
	eventRecorder    record.EventRecorder
	logger           logr.Logger
}

func (h *enqueueRequestsForAlbconfigEvent) Create(e event.CreateEvent, queue workqueue.RateLimitingInterface) {
	albconfig, ok := e.Object.(*v1.AlbConfig)
	if ok {
		h.logger.Info("controller: albconfig Create event", "albconfig", util.NamespacedName(albconfig).String())
		h.enqueueAlbconfig(queue, albconfig)
	}
}

func (h *enqueueRequestsForAlbconfigEvent) Update(e event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	albconfigOld := e.ObjectOld.(*v1.AlbConfig)
	albconfigNew := e.ObjectNew.(*v1.AlbConfig)

	if equality.Semantic.DeepEqual(albconfigOld.Annotations, albconfigNew.Annotations) &&
		equality.Semantic.DeepEqual(albconfigOld.Spec, albconfigNew.Spec) &&
		equality.Semantic.DeepEqual(albconfigOld.DeletionTimestamp.IsZero(), albconfigNew.DeletionTimestamp.IsZero()) {
		return
	}

	h.logger.Info("controller: albconfig Update event", "albconfig", util.NamespacedName(albconfigNew).String())
	h.enqueueAlbconfig(queue, albconfigNew)
}

func (h *enqueueRequestsForAlbconfigEvent) Delete(e event.DeleteEvent, queue workqueue.RateLimitingInterface) {
}

func (h *enqueueRequestsForAlbconfigEvent) Generic(e event.GenericEvent, queue workqueue.RateLimitingInterface) {
	albconfig, ok := e.Object.(*v1.AlbConfig)
	if ok {
		h.logger.Info("controller: albconfig Generic event", "albconfig", util.NamespacedName(albconfig).String())
		h.enqueueAlbconfig(queue, albconfig)
	}
}

func (h *enqueueRequestsForAlbconfigEvent) enqueueAlbconfig(queue workqueue.RateLimitingInterface, albconfig *v1.AlbConfig) {
	queue.Add(reconcile.Request{
		NamespacedName: util.NamespacedName(albconfig),
	})
}
