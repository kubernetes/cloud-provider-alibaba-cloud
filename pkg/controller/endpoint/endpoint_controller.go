package endpoint

import (
	"context"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	recon := &ReconcileEndpoint{
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

	// Watch for changes to primary resource
	return c.Watch(
		&source.Kind{
			Type: &corev1.Node{},
		},
		&handler.EnqueueRequestForObject{},
	)
}

// ReconcileEndpoint implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileEndpoint{}

// ReconcileEndpoint reconciles a AutoRepair object
type ReconcileEndpoint struct {
	client client.Client
	scheme *runtime.Scheme

	//record event recorder
	record record.EventRecorder
}

func (r *ReconcileEndpoint) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	rlog := log.WithFields(log.Fields{"NodePool": request.NamespacedName})

	nodepool := &corev1.Node{}
	err := r.client.Get(context.TODO(), request.NamespacedName, nodepool)
	if err != nil {
		if errors.IsNotFound(err) {
			rlog.Infof("service not found, skip")
			// Request object not found, could have been deleted
			// after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
