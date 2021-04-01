package pvtz

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	return add(mgr, newReconciler(mgr, ctx))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, ctx *shared.SharedContext) reconcile.Reconciler {
	recon := &ReconcileDNS{
		cloud:    ctx.Provider(),
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		actuator: NewActuator(mgr.GetClient(), ctx.Provider()),
		record:   mgr.GetEventRecorderFor("Pvtz"),
	}
	return recon
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(
		"pvtz-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 1,
		},
	)
	if err != nil {
		return err
	}
	eventHandler := NewEventHandlerWithClient()
	ki := []*source.Kind{
		{
			Type: &corev1.Service{},
		},
		{
			Type: &corev1.Endpoints{},
		},
	}
	pred := &PredicateDNS{}
	for i := range ki {
		err = c.Watch(ki[i], eventHandler, pred)
		if err != nil {
			return fmt.Errorf("watch resource: %s, %s", ki[i].Type, err.Error())
		}
	}
	return nil
}

// ReconcileService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileDNS{}

// ReconcileService reconciles a AutoRepair object
type ReconcileDNS struct {
	cloud  provider.Provider
	client client.Client
	scheme *runtime.Scheme

	actuator *Actuator

	//record event recorder
	record record.EventRecorder
}

func (m *ReconcileDNS) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	svc := &corev1.Service{}
	err := m.client.Get(context.TODO(), request.NamespacedName, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			err = m.actuator.Delete(request.NamespacedName)
			if err != nil {
				m.record.Event(svc, corev1.EventTypeWarning, EventReasonHandleServiceDeletionError, err.Error())
			} else {
				m.record.Event(svc, corev1.EventTypeNormal, EventReasonHandleServiceDeletionSucceed, EventReasonHandleServiceDeletionSucceed)
			}
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}
	err = m.actuator.Update(svc)
	if err != nil {
		m.record.Event(svc, corev1.EventTypeWarning, EventReasonHandleServiceUpdateError, err.Error())
	} else {
		m.record.Event(svc, corev1.EventTypeNormal, EventReasonHandleServiceUpdateSucceed, EventReasonHandleServiceUpdateSucceed)
	}
	return reconcile.Result{}, err
}
