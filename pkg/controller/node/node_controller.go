package node

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/metric"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"
)

var log = klogr.New().WithName("node-controller")

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	return add(mgr, newReconciler(mgr, ctx))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, ctx *shared.SharedContext) *ReconcileNode {
	recon := &ReconcileNode{
		monitorPeriod:   5 * time.Minute,
		statusFrequency: 5 * time.Minute,
		// provider
		cloud:  ctx.Provider(),
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		record: mgr.GetEventRecorderFor("node-controller"),
	}
	return recon
}

type nodeController struct {
	c     controller.Controller
	recon *ReconcileNode
}

// Start this function will not be called until the resource lock is acquired
func (controller nodeController) Start(ctx context.Context) error {
	controller.recon.PeriodicalSync()
	return controller.c.Start(ctx)
}

// add a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileNode) error {
	// Create a new controller
	c, err := controller.NewUnmanaged(
		"node-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 1,
		},
	)
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AutoRepair
	if err := c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	return mgr.Add(&nodeController{c: c, recon: r})
}

// ReconcileNode implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNode{}

// ReconcileNode reconciles a AutoRepair object
type ReconcileNode struct {
	cloud  prvd.Provider
	client client.Client
	scheme *runtime.Scheme

	// monitorPeriod controlling monitoring period,
	// i.e. how often does NodeController check node status
	// posted from kubelet. This value should be lower than
	// nodeMonitorGracePeriod set in controller-manager
	monitorPeriod    time.Duration
	statusFrequency  time.Duration
	configCloudRoute bool

	//record event recorder
	record record.EventRecorder
}

func (m *ReconcileNode) Reconcile(
	ctx context.Context, request reconcile.Request,
) (reconcile.Result, error) {
	klog.V(5).Infof("reconcile node: %s", request.NamespacedName)
	node := &corev1.Node{}
	err := m.client.Get(context.TODO(), request.NamespacedName, node)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("node not found, skip")
			// Request object not found, could have been deleted
			// after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, m.syncCloudNode(node)
}

func (m *ReconcileNode) syncCloudNode(node *corev1.Node) error {
	cloudTaint := findCloudTaint(node.Spec.Taints)
	if cloudTaint == nil {
		klog.V(5).Infof("node %s is registered without cloud taint. return ok", node.Name)
		return nil
	}
	return m.doAddCloudNode(node)
}

// This processes nodes that were added into the cluster, and cloud initialize them if appropriate
func (m *ReconcileNode) doAddCloudNode(node *corev1.Node) error {
	start := time.Now()
	prvdId := node.Spec.ProviderID
	if prvdId == "" {
		log.Info(fmt.Sprintf("warning: provider id not exist, skip %s initialize", node.Name))
		return nil
	}

	instance, err := findCloudECS(m.cloud, prvdId)
	if err != nil {
		if err == ErrNotFound {
			log.Info("cloud instance %s not found", node.Name)
			return nil
		}
		log.Error(err, "fail to find ecs", "providerId", prvdId)
		return fmt.Errorf("find ecs: %s", err.Error())
	}

	// If user provided an IP address, ensure that IP address is found
	// in the cloud provider before removing the taint on the node
	if nodeIP, ok := isProvidedAddrExist(node, instance.Addresses); ok && nodeIP == nil {
		return fmt.Errorf("failed to get specified nodeIP in cloud provider")
	}

	initializer := func() (done bool, err error) {
		log.Info("try remove cloud taints", "node", node.Name)

		diff := func(copy runtime.Object) (client.Object, error) {
			nins := copy.(*corev1.Node)
			setFields(nins, instance, m.configCloudRoute)
			return nins, nil
		}
		err = helper.PatchM(m.client, node, diff, helper.PatchAll)
		if err != nil {
			log.Error(err, "fail to patch node", "node", node.Name)
			return false, nil
		}
		tags := map[string]string{
			"k8s.aliyun.com": "true",
			"kubernetes.ccm": "true",
		}
		err = m.cloud.SetInstanceTags(context.TODO(), instance.InstanceID, tags)
		if err != nil {
			if !strings.Contains(err.Error(), "Forbidden.RAM") {
				log.Error(err, "fail to tag instance", "node", instance.InstanceID)
				//retry
				return false, nil
			}
		}

		log.Info("finished remove uninitialized cloud taints", "node", node.Name)
		// After adding, call UpdateNodeAddress to set the CloudProvider provided IPAddresses
		// So that users do not see any significant delay in IP addresses being filled into the node
		_ = m.syncNode([]corev1.Node{*node})
		return true, nil
	}

	nodeRef := &corev1.ObjectReference{
		Kind:      "Node",
		Name:      node.Name,
		UID:       types.UID(node.Name),
		Namespace: "",
	}

	err = wait.PollImmediate(2*time.Second, 20*time.Second, initializer)
	if err != nil {
		m.record.Event(
			nodeRef,
			corev1.EventTypeWarning,
			helper.FailedAddNode,
			fmt.Sprintf("Error adding node: %s", helper.GetLogMessage(err)),
		)
		return fmt.Errorf("doAddCloudNode %s error: %s", node.Name, err.Error())
	}

	m.record.Event(nodeRef, corev1.EventTypeNormal, helper.InitializedNode, "Initialize node successfully")
	metric.NodeLatency.WithLabelValues("remove_taint").Observe(metric.MsSince(start))
	log.Info("Successfully initialized node", "node", node.Name)

	return nil
}

// syncNode sync the nodeAddress & cloud node existence
func (m *ReconcileNode) syncNode(nodes []corev1.Node) error {

	instances, err := m.cloud.ListInstances(context.TODO(), nodeids(nodes))
	if err != nil {
		return fmt.Errorf("[NodeAddress] list instances from api: %s", err.Error())
	}

	for i := range nodes {
		node := &nodes[i]
		cloudNode := instances[node.Spec.ProviderID]
		nodeRef := &corev1.ObjectReference{
			Kind:      "Node",
			Name:      node.Name,
			UID:       types.UID(node.Name),
			Namespace: "",
		}

		if cloudNode == nil {
			// if cloud node has been deleted, try to delete node from cluster
			condition := nodeConditionReady(m.client, node)
			if condition != nil && condition.Status == corev1.ConditionUnknown {
				log.Info("node is NotReady and cloud node can not found by prvdId, try to delete node from cluster ", "node", node.Name, "prvdId", node.Spec.ProviderID)
				// ignore error, retry next loop
				deleteNode(m, node)
			}

			log.Info("cloud node not found by prvdId, skip update node address", "node", node.Name, "prvdId", node.Spec.ProviderID)
			continue
		}

		cloudNode.Addresses = setHostnameAddress(node, cloudNode.Addresses)
		// If nodeIP was suggested by user, ensure that
		// it can be found in the cloud as well (consistent with the behaviour in kubelet)
		nodeIP, ok := isProvidedAddrExist(node, cloudNode.Addresses)
		if ok {
			if nodeIP == nil {
				log.Error(fmt.Errorf("user specified ip is not found in cloudprovider: %v", node.Status.Addresses),
					"", "node", node.Name)
				continue
			}
			// override addresses
			cloudNode.Addresses = []corev1.NodeAddress{*nodeIP}
		}

		diff := func(copy runtime.Object) (client.Object, error) {
			nins := copy.(*corev1.Node)
			nins.Status.Addresses = cloudNode.Addresses
			return nins, nil
		}
		err := helper.PatchM(m.client, node, diff, helper.PatchStatus)
		if err != nil {
			log.Error(err, "patch node address error, wait for next retry", "node", node.Name)
			m.record.Event(
				nodeRef, corev1.EventTypeWarning, helper.FailedSyncNode, err.Error(),
			)
		}

		diff = func(copy runtime.Object) (client.Object, error) {
			nins := copy.(*corev1.Node)
			setFields(nins, cloudNode, false)
			return nins, nil
		}

		err = helper.PatchM(m.client, node, diff, helper.PatchAll)
		if err != nil {
			log.Error(err, "patch node label error, wait for next retry", "node", node.Name)
			m.record.Event(
				nodeRef, corev1.EventTypeWarning, helper.FailedSyncNode, err.Error(),
			)
		}
	}
	return nil
}

func (m *ReconcileNode) PeriodicalSync() {
	// Start a loop to periodically update the node addresses obtained from the cloud
	syncNode := func() {

		nodes, err := NodeList(m.client)
		if err != nil {
			log.Error(err, "address sync error")
			return
		}

		// ignore return value, retry on error
		err = batchOperate(
			nodes.Items,
			m.syncNode,
		)
		if err != nil {
			log.Error(err, "periodically sync node error")
		}
		log.Info("sync node successfully", "length", len(nodes.Items))
	}

	go wait.Until(syncNode, m.statusFrequency, wait.NeverStop)
}
