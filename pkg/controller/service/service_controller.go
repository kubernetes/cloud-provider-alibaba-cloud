package service

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
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
	recon := &ReconcileService{
		cloud:  ctx.Provider(),
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		record: mgr.GetEventRecorderFor("NodePool"),
	}
	return recon
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(
		"service-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 1,
		},
	)
	if err != nil {
		return err
	}
	hand := NewMapHandler()
	ki := []*source.Kind{
		{Type: &corev1.Service{}}, {Type: &corev1.Endpoints{}}, {Type: &corev1.Node{}},
	}
	for i := range ki {
		err = c.Watch(ki[i], hand)
		if err != nil {
			return fmt.Errorf("watch resource: %s, %s", ki[i].Type, err.Error())
		}
	}
	return nil
}

// ReconcileService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileService{}

// ReconcileService reconciles a AutoRepair object
type ReconcileService struct {
	cloud  provider.Provider
	client client.Client
	scheme *runtime.Scheme

	//record event recorder
	record record.EventRecorder
}

func (r *ReconcileService) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	rlog := log.WithFields(log.Fields{"Service": request.NamespacedName})

	svc := &corev1.Service{}
	err := r.client.Get(context.TODO(), request.NamespacedName, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			rlog.Infof("service %s not found, skip", request.NamespacedName)
			// Request object not found, could have been deleted
			// after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	log.Infof("do reconcile service %s/%s", svc.Namespace, svc.Name)
	return reconcile.Result{}, nil
}
