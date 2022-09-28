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
	var err error
	err = addServiceReconciler(mgr, ctx)
	if err != nil {
		return err
	}
	// TODO should turn off by default
	err = addPodReconciler(mgr, ctx)
	if err != nil {
		return err
	}
	return nil
}

func addServiceReconciler(mgr manager.Manager, ctx *shared.SharedContext) error {
	r := &ServiceReconciler{
		cloud:    ctx.Provider(),
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		actuator: NewActuator(mgr.GetClient(), ctx.Provider()),
		record:   mgr.GetEventRecorderFor("Pvtz"),
	}
	c, err := controller.New(
		"pvtz-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 1,
			RecoverPanic:            true,
		},
	)
	if err != nil {
		return err
	}
	eventHandler := NewEventHandlerWithClient()
	kinds := []*source.Kind{
		{
			Type: &corev1.Service{},
		},
		{
			Type: &corev1.Endpoints{},
		},
	}
	sp := &ServicePredicate{}
	for i := range kinds {
		err = c.Watch(kinds[i], eventHandler, sp)
		if err != nil {
			return fmt.Errorf("watch resource: %s, %s", kinds[i].Type, err.Error())
		}
	}
	return nil
}

type ServiceReconciler struct {
	cloud  prvd.Provider
	client client.Client
	scheme *runtime.Scheme

	actuator *Actuator

	//record event recorder
	record record.EventRecorder
}

func (m *ServiceReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	svc := &corev1.Service{}
	err := m.client.Get(context.TODO(), request.NamespacedName, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			err = m.actuator.DeleteService(request.NamespacedName)
			if err != nil {
				m.record.Event(svc, corev1.EventTypeWarning, EventReasonHandleServiceDeletionError, err.Error())
			} else {
				m.record.Event(svc, corev1.EventTypeNormal, EventReasonHandleServiceDeletionSucceed, EventReasonHandleServiceDeletionSucceed)
			}
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}
	err = m.actuator.UpdateService(svc)
	if err != nil {
		m.record.Event(svc, corev1.EventTypeWarning, EventReasonHandleServiceUpdateError, err.Error())
	} else {
		m.record.Event(svc, corev1.EventTypeNormal, EventReasonHandleServiceUpdateSucceed, EventReasonHandleServiceUpdateSucceed)
	}
	return reconcile.Result{}, err
}

func addPodReconciler(mgr manager.Manager, ctx *shared.SharedContext) error {
	r := &PodReconciler{
		cloud:    ctx.Provider(),
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		actuator: NewActuator(mgr.GetClient(), ctx.Provider()),
		record:   mgr.GetEventRecorderFor("Pvtz"),
	}
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
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, eventHandler)
	if err != nil {
		return fmt.Errorf("watch resource: pod, %s", err.Error())
	}
	return nil
}

type PodReconciler struct {
	cloud  prvd.Provider
	client client.Client
	scheme *runtime.Scheme

	actuator *Actuator

	//record event recorder
	record record.EventRecorder
}

func (m *PodReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	pod := &corev1.Pod{}
	err := m.client.Get(context.TODO(), request.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			err = m.actuator.DeletePod(request.NamespacedName)
			if err != nil {
				m.record.Event(pod, corev1.EventTypeWarning, EventReasonHandlePodDeletionError, err.Error())
			} else {
				m.record.Event(pod, corev1.EventTypeNormal, EventReasonHandlePodDeletionSucceed, EventReasonHandlePodDeletionSucceed)
			}
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}
	err = m.actuator.UpdatePod(pod)
	if err != nil {
		m.record.Event(pod, corev1.EventTypeWarning, EventReasonHandlePodUpdateError, err.Error())
	} else {
		m.record.Event(pod, corev1.EventTypeNormal, EventReasonHandlePodUpdateSucceed, EventReasonHandlePodUpdateSucceed)
	}
	return reconcile.Result{}, err
}
