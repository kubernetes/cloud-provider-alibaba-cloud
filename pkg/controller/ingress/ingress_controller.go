package ingress

import (
	"context"
	"time"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/controller"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ingressController struct {
	c     controller.Controller
	recon *albconfigReconciler
}

func (ingC ingressController) Start(ctx context.Context) error {
	klog.Infof("ingressController start")
	return ingC.c.Start(ctx)
}
func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	rateLimit := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 300*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)
	r, err := NewAlbConfigReconciler(mgr, ctx)
	if err != nil {
		return err
	}
	// Create a new controller
	c, err := controller.NewUnmanaged(
		albIngressControllerName, mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: r.maxConcurrentReconciles,
			RateLimiter:             rateLimit,
		},
	)
	if err != nil {
		return err
	}
	klog.Infof("setupWatches start")
	r.acEventChan = make(chan event.GenericEvent)
	acEventHandler := NewEnqueueRequestsForAlbconfigEvent(r.k8sClient, r.eventRecorder, r.logger)
	if err := c.Watch(&source.Channel{Source: r.acEventChan}, acEventHandler); err != nil {
		return err
	}

	if err := c.Watch(&source.Kind{Type: &v1.AlbConfig{}}, acEventHandler); err != nil {
		return err
	}

	klog.Infof("Add start")
	return mgr.Add(&ingressController{c: c, recon: r})
}
