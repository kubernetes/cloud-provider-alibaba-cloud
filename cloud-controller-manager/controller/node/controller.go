/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package node

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils/metric"
	"k8s.io/klog"
	"time"

	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/route"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	//nodeutilv1 "k8s.io/kubernetes/pkg/api/v1/node"
	"k8s.io/cloud-provider"
	"k8s.io/cloud-provider/api"
	"k8s.io/cloud-provider/node/helpers"
	metrics "k8s.io/component-base/metrics/prometheus/ratelimiter"
	controller "k8s.io/kube-aggregator/pkg/controllers"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
)

type CloudNodeController struct {
	informer    coreinformers.NodeInformer
	kclient     clientset.Interface
	recorder    record.EventRecorder
	cloud       cloudprovider.Interface
	broadcaster record.EventBroadcaster

	// monitorPeriod controlling CloudNodeController monitoring period,
	// i.e. how often does CloudNodeController check node status posted from kubelet.
	// This value should be lower than nodeMonitorGracePeriod set in controller-manager
	monitorPeriod time.Duration

	statusFrequency  time.Duration
	nodeListerSynced cache.InformerSynced
}

const (
	// RETRY_COUNT controls the number of retries of writing NodeStatus update.
	RETRY_COUNT = 5

	// The amount of time the nodecontroller should sleep between retrying NodeStatus updates
	retrySleepTime = 20 * time.Millisecond

	// NODE_CONTROLLER name of node controller
	NODE_CONTROLLER = "cloud-node-controller"

	// MAX_BATCH_NUM batch process per loop.
	MAX_BATCH_NUM = 50
)

// CloudNodeAttribute node attribute from cloud instance
type CloudNodeAttribute struct {
	InstanceID   string
	Addresses    []v1.NodeAddress
	InstanceType string
	Zone         string
	Region       string
}

// CloudInstance is an interface to interact with cloud api
type CloudInstance interface {
	// SetInstanceTags set instance tags for instance id
	SetInstanceTags(ctx context.Context, insid string, tags map[string]string) error

	// ListInstances list instance by given ids.
	ListInstances(ctx context.Context, ids []string) (map[string]*CloudNodeAttribute, error)
}

// NewCloudNodeController creates a CloudNodeController object
func NewCloudNodeController(
	ninformer coreinformers.NodeInformer,
	kubeClient clientset.Interface,
	cloud cloudprovider.Interface,
	nodeMonitorPeriod time.Duration,
	nodeStatusUpdateFrequency time.Duration,
) *CloudNodeController {

	eventer, caster := broadcaster()

	if kubeClient != nil && kubeClient.CoreV1().RESTClient().GetRateLimiter() != nil {
		limitor := kubeClient.CoreV1().RESTClient().GetRateLimiter()
		_ = metrics.RegisterMetricAndTrackRateLimiterUsage(NODE_CONTROLLER, limitor)
		klog.Infof("start sending events to api server.")
	} else {
		klog.Infof("no api server defined - no events will be sent to API server.")
	}

	cnc := &CloudNodeController{
		informer:         ninformer,
		kclient:          kubeClient,
		recorder:         eventer,
		broadcaster:      caster,
		cloud:            cloud,
		monitorPeriod:    nodeMonitorPeriod,
		statusFrequency:  nodeStatusUpdateFrequency,
		nodeListerSynced: ninformer.Informer().HasSynced,
	}

	HandlerForNode(cnc, ninformer)

	return cnc
}

func HandlerForNode(
	cnc *CloudNodeController,
	ninformer coreinformers.NodeInformer,
) {
	ninformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				node := obj.(*v1.Node)
				klog.V(4).Infof("receive node add event: %s", node.Name)
				start := time.Now()
				err := cnc.AddCloudNode(node)
				if err != nil {
					klog.Errorf("remove cloud node taints fail: %s", err.Error())
				}
				metric.NodeLatency.WithLabelValues("remove_taint").Observe(metric.MsSince(start))
			},
		},
	)
}

// This controller deletes a node if kubelet is not reporting
// and the node is gone from the cloud provider.
func (cnc *CloudNodeController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	klog.Info("starting node controller")

	if !controller.WaitForCacheSync(
		NODE_CONTROLLER,
		stopCh,
		cnc.nodeListerSynced) {
		return
	}

	if cnc.broadcaster != nil {
		sink := &v1core.EventSinkImpl{
			Interface: v1core.New(cnc.kclient.CoreV1().RESTClient()).Events(""),
		}
		cnc.broadcaster.StartRecordingToSink(sink)
	}

	// The following loops run communicate with the APIServer with a worst case complexity
	// of O(num_nodes) per cycle. These functions are justified here because these events fire
	// very infrequently. DO NOT MODIFY this to perform frequent operations.

	// Start a loop to periodically update the node addresses obtained from the cloud
	go wait.Until(
		func() {
			nodes, err := nodeLists(cnc.kclient)
			if err != nil {
				klog.Errorf("Error monitoring node status: %v", err)
				return
			}

			// ignore return value, retry on error
			err = batchAddressUpdate(
				nodes.Items,
				cnc.syncNodeAddress,
			)
			if err != nil {
				klog.Errorf("periodically update address: %s", err.Error())
			}
		},
		cnc.statusFrequency,
		wait.NeverStop,
	)

	// Start a loop to periodically check if any nodes have been deleted from cloudprovider
	go wait.Until(
		func() {
			nodes, err := nodeLists(cnc.kclient)
			if err != nil {
				klog.Errorf("Error monitoring node status: %v", err)
				return
			}
			// ignore return value, retry on error
			err = batchAddressUpdate(
				nodes.Items,
				cnc.syncCloudNodes,
			)
			if err != nil {
				klog.Errorf("periodically try detect node existence: %s", err.Error())
			}
		},
		cnc.monitorPeriod,
		wait.NeverStop,
	)

	// Start a loop to periodically check if uninitialized taints has been remove from node
	go wait.Until(
		func() {
			nodes, err := nodeLists(cnc.kclient)
			if err != nil {
				klog.Errorf("Error monitoring node status: %v", err)
				return
			}
			for _, node := range nodes.Items {
				err := cnc.AddCloudNode(&node)
				if err != nil {
					klog.Errorf("periodically remove cloud node %s taints: %s", node.Name, err.Error())
				}
			}
		},
		3*time.Minute,
		wait.NeverStop,
	)
}

func (cnc *CloudNodeController) AddCloudNode(node *v1.Node) error {
	curNode, err := cnc.kclient.
		CoreV1().
		Nodes().
		Get(context.Background(), node.Name, metav1.GetOptions{})
	if err != nil {
		//retry
		return fmt.Errorf("retrieve node error: %s", err.Error())
	}
	cloudTaint := findCloudTaint(curNode.Spec.Taints)
	if cloudTaint == nil {
		klog.V(4).Infof("Node %s is registered without cloud taint. Will not process.", node.Name)
		return nil
	}
	return cnc.doAddCloudNode(curNode)
}

// syncNodeAddress updates the nodeAddress
func (cnc *CloudNodeController) syncNodeAddress(nodes []v1.Node) error {

	ins, ok := cnc.cloud.(CloudInstance)
	if !ok {
		return fmt.Errorf("cloud instance not implemented")
	}
	instances, err := ins.ListInstances(context.Background(), nodeids(nodes))
	if err != nil {
		return fmt.Errorf("syncNodeAddress, retrieve instances from api error: %s", err.Error())
	}

	for i := range nodes {
		node := &nodes[i]
		cloudNode := instances[node.Spec.ProviderID]
		if cloudNode == nil {
			klog.Infof("node %s not found, skip update node address", node.Spec.ProviderID)
			continue
		}
		cloudNode.Addresses = setHostnameAddress(node, cloudNode.Addresses)
		// If nodeIP was suggested by user, ensure that
		// it can be found in the cloud as well (consistent with the behaviour in kubelet)
		nodeIP, ok := isProvidedAddrExist(node, cloudNode.Addresses)
		if ok {
			if nodeIP == nil {
				klog.Errorf("User has specified Node IP in kubelet but it is not found in cloudprovider")
				continue
			}
			// override addresses
			cloudNode.Addresses = []v1.NodeAddress{*nodeIP}
		}
		err := tryPatchNodeAddress(cnc.kclient, node, cloudNode.Addresses)
		if err != nil {
			klog.Errorf("Wait for next retry, patch node address error: %s", err.Error())
			cnc.recorder.Eventf(
				node,
				v1.EventTypeWarning,
				"SyncNodeFailed",
				"Error patching node address: %s", err.Error(),
			)
		}
	}
	return nil
}

func (cnc *CloudNodeController) syncCloudNodes(nodes []v1.Node) error {
	ins, ok := cnc.cloud.(CloudInstance)
	if !ok {
		return fmt.Errorf("cloud instance not implemented")
	}
	instances, err := ins.ListInstances(context.Background(), nodeids(nodes))
	if err != nil {
		return fmt.Errorf("syncCloudNodes, retrieve instances from api error: %s", err.Error())
	}

	for i := range nodes {
		node := &nodes[i]

		condition := nodeConditionReady(cnc.kclient, node)
		if condition == nil {
			klog.Infof("node %s condition not ready, wait for next retry", node.Spec.ProviderID)
			continue
		}

		if condition.Status == v1.ConditionTrue {
			// skip ready nodes
			continue
		}

		cloudNode := instances[node.Spec.ProviderID]
		if cloudNode != nil {
			continue
		}

		klog.Infof("node %s not found, start to delete from meta", node.Spec.ProviderID)
		// try delete node and ignore error, retry next loop
		deleteNode(cnc, node)
	}
	return nil
}

type nodeModifier func(*v1.Node)

// This processes nodes that were added into the cluster, and cloud initialize them if appropriate
func (cnc *CloudNodeController) doAddCloudNode(node *v1.Node) error {
	ctx := context.Background()
	ins, ok := cnc.cloud.(CloudInstance)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("failed to get ins from cloud provider"))
		return fmt.Errorf("cloud instance is not implemented")
	}
	err := wait.PollImmediate(
		2*time.Second,
		20*time.Second,
		func() (done bool, err error) {
			klog.V(5).Infof("try remove cloud taints for %s", node.Name)
			curNode, err := cnc.kclient.CoreV1().Nodes().Get(context.Background(), node.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("retrieve node error: %s", err.Error())
				//retry
				return false, nil
			}
			copyNode := curNode.DeepCopy()
			providerID, err := cnc.getProviderID(copyNode)
			if err != nil {
				klog.Errorf("failed to get provider ID for node %s at cloudprovider: %v", node.Name, err)
				return false, nil
			}

			cloudIns, err := cnc.getCloudInstance(ctx, ins, providerID)
			if err != nil {
				klog.Errorf("failed to get cloud instance data for node %s: %v", node.Name, err)
				return false, nil
			}

			nodeModifiers, err := cnc.getNodeModifiersFromCloudProvider(providerID, copyNode, cloudIns)
			if err != nil {
				klog.Errorf("failed to get node modifiers for node %s: %v", node.Name, err)
				// fast fail
				return true, nil
			}

			// remove cloud taint
			nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
				removeCloudTaints(n)
			})

			// TODO(wlan0): Move this logic to the route controller using the node taint instead of condition
			// Since there are node taints, do we still need this?
			// This condition marks the node as unusable until routes are initialized in the cloud provider
			// Aoxn: Hack for alibaba cloud
			if route.Options.ConfigCloudRoutes &&
				cnc.cloud.ProviderName() == "alicloud" {
				if err := cnc.setNodeCondition(node); err != nil {
					klog.Errorf("set node %s condition error: %s", node.Name, err.Error())
					return false, nil
				}

				// fetch latest node from API server since alicloud-specific condition was set and informer cache may be stale
				curNode, err = cnc.kclient.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
				if err != nil {
					klog.Errorf("get node %s error: %s", node.Name, err.Error())
					return false, nil
				}
			}

			newNode := curNode.DeepCopy()
			for _, modify := range nodeModifiers {
				modify(newNode)
			}

			err = ins.SetInstanceTags(
				ctx,
				cloudIns.InstanceID,
				map[string]string{
					"k8s.aliyun.com": "true",
					"kubernetes.ccm": "true",
				},
			)
			if err != nil {
				if !strings.Contains(err.Error(), "Forbidden.RAM") {
					klog.Errorf("tag instance %s error: %s", cloudIns.InstanceID, err.Error())
					//retry
					return false, nil
				}
				// Old ROS template does not have AddTags Permission.
				// It is ok to skip `Forbidden` error for compatible reason.
			}

			nnode, err := PatchNode(cnc.kclient, curNode, newNode)
			if err != nil {
				klog.Errorf("patch error: %s", err.Error())
				return false, nil
			}
			klog.V(5).Infof("finished remove uninitialized cloud taints for %s", node.Name)
			// After adding, call UpdateNodeAddress to set the CloudProvider provided IPAddresses
			// So that users do not see any significant delay in IP addresses being filled into the node
			_ = cnc.syncNodeAddress([]v1.Node{*nnode})
			return true, nil
		},
	)

	ref := &v1.ObjectReference{
		Kind:      "Node",
		Name:      node.Name,
		UID:       types.UID(node.UID),
		Namespace: "",
	}

	if err != nil {
		klog.Errorf("doAddCloudNode %s error: %s", node.Name, err.Error())
		cnc.recorder.Eventf(
			ref,
			v1.EventTypeWarning,
			"AddNodeFailed",
			"Error add node: %s",
			err.Error(),
		)
		utilruntime.HandleError(err)
		return err
	}

	klog.Infof("Successfully initialized node %s with cloud provider", node.Name)

	cnc.recorder.Eventf(
		ref,
		v1.EventTypeNormal,
		"InitializedNode",
		"Initialize node successfully",
	)
	return nil
}

func batchAddressUpdate(
	nodes []v1.Node,
	batch func([]v1.Node) error,
) error {

	klog.Infof("batch process update node address, length %d", len(nodes))
	for len(nodes) > MAX_BATCH_NUM {
		if err := batch(nodes[0:MAX_BATCH_NUM]); err != nil {
			klog.Errorf("batch process func error: %s", err.Error())
			return err
		}
		nodes = nodes[MAX_BATCH_NUM:]
	}
	if len(nodes) <= 0 {
		return nil
	}
	return batch(nodes)
}

func nodeids(nodes []v1.Node) []string {
	var ids []string
	for _, node := range nodes {
		ids = append(ids, node.Spec.ProviderID)
	}
	return ids
}

func setHostnameAddress(node *v1.Node, addrs []v1.NodeAddress) []v1.NodeAddress {
	// Check if a hostname address exists in the cloud provided addresses
	hostnameExists := false
	for i := range addrs {
		if addrs[i].Type == v1.NodeHostName {
			hostnameExists = true
		}
	}
	// If hostname was not present in cloud provided addresses, use the hostname
	// from the existing node (populated by kubelet)
	if !hostnameExists {
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeHostName {
				addrs = append(addrs, addr)
			}
		}
	}
	return addrs
}

func tryPatchNodeAddress(
	kclient kubernetes.Interface,
	node *v1.Node,
	address []v1.NodeAddress,
) error {

	clone := node.DeepCopy()
	clone.Status.Addresses = address

	if !isNodeAddressChanged(
		node.Status.Addresses,
		clone.Status.Addresses,
	) {
		return nil
	}
	klog.Infof("try patch node address for %s", node.Spec.ProviderID)
	_, _, err := nodeutil.PatchNodeStatus(
		kclient.CoreV1(),
		types.NodeName(node.Name),
		node,
		clone,
	)
	return err
}

func deleteNode(cnc *CloudNodeController, node *v1.Node) {

	ref := &v1.ObjectReference{
		Kind:      "Node",
		Name:      node.Name,
		UID:       types.UID(node.UID),
		Namespace: "",
	}
	klog.V(2).Infof("recording %s event message for node %s", "DeletingNode", node.Name)

	go func(nodeName string) {
		defer utilruntime.HandleCrash()
		if err := cnc.kclient.CoreV1().
			Nodes().Delete(
			context.Background(), nodeName, metav1.DeleteOptions{},
		); err != nil {
			klog.Errorf("unable to delete node %q: %v", nodeName, err)
			cnc.recorder.Eventf(
				ref,
				v1.EventTypeWarning,
				"DeleteNodeFailed",
				"Error deleting node: %s",
				err.Error(),
			)
		} else {
			cnc.recorder.Eventf(
				ref,
				v1.EventTypeNormal,
				"DeletedNode",
				"Deleted node",
			)
		}
	}(node.Name)
}

func nodeConditionReady(kclient kubernetes.Interface, node *v1.Node) *v1.NodeCondition {
	// Try to get the current node status
	// If node status is empty, then kubelet has not posted ready status yet.
	// In this case, process next node
	var err error
	for rep := 0; rep < RETRY_COUNT; rep++ {
		_, ccondition := helpers.GetNodeCondition(&node.Status, v1.NodeReady)
		if ccondition != nil {
			return ccondition
		}
		name := node.Name
		node, err = kclient.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed while getting a Node to retry updating "+
				"NodeStatus. Probably Node %s was deleted.", name)
			break
		}
		time.Sleep(retrySleepTime)
	}
	return nil
}

func (cnc *CloudNodeController) getProviderID(node *v1.Node) (string, error) {
	if node.Spec.ProviderID != "" {
		return node.Spec.ProviderID, nil
	}

	providerID, err := cloudprovider.GetInstanceProviderID(context.Background(), cnc.cloud, types.NodeName(node.Name))
	if err == cloudprovider.NotImplemented {
		// we should attempt to set providerID on curNode, but
		// we can continue if we fail since we will attempt to set
		// node addresses given the node name in getNodeAddressesByProviderIDOrName
		klog.Warningf("cloud provider does not set node provider ID, using node name to discover node %s", node.Name)
		return "", nil
	}

	return providerID, err
}

func (cnc *CloudNodeController) getCloudInstance(
	ctx context.Context, ins CloudInstance,
	providerID string,
) (*CloudNodeAttribute, error) {
	nodes, err := ins.ListInstances(ctx, []string{providerID})
	if err != nil {
		return nil, fmt.Errorf("cloud instance api fail, %s", err.Error())
	}
	cloudIns, ok := nodes[providerID]
	if !ok || cloudIns == nil {
		return nil, fmt.Errorf("instance not found")
	}
	return cloudIns, nil
}

func (cnc *CloudNodeController) getNodeModifiersFromCloudProvider(
	providerID string,
	node *v1.Node,
	cloudIns *CloudNodeAttribute,
) ([]nodeModifier, error) {
	var nodeModifiers []nodeModifier
	if node.Spec.ProviderID == "" {
		if providerID != "" {
			nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
				n.Spec.ProviderID = providerID
			})
		}
	}

	// If user provided an IP address, ensure that IP address is found
	// in the cloud provider before removing the taint on the node
	if nodeIP, ok := isProvidedAddrExist(node, cloudIns.Addresses); ok && nodeIP == nil {
		return nil, fmt.Errorf("failed to get specified nodeIP in cloud provider")
	}

	if cloudIns.InstanceType != "" {
		klog.Infof(
			"Adding node label from cloud provider: %s=%s, %s=%s",
			v1.LabelInstanceType, cloudIns.InstanceType,
			v1.LabelInstanceTypeStable, cloudIns.InstanceType,
		)
		nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
			if n.Labels == nil {
				n.Labels = map[string]string{}
			}
			n.Labels[v1.LabelInstanceType] = cloudIns.InstanceType
			n.Labels[v1.LabelInstanceTypeStable] = cloudIns.InstanceType

		})
	}

	if cloudIns.Zone != "" {
		klog.Infof(
			"Adding node label from cloud provider: %s=%s, %s=%s",
			v1.LabelZoneFailureDomain, cloudIns.Zone,
			v1.LabelZoneFailureDomainStable, cloudIns.Zone,
		)
		nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
			if n.Labels == nil {
				n.Labels = map[string]string{}
			}
			n.Labels[v1.LabelZoneFailureDomain] = cloudIns.Zone
			n.Labels[v1.LabelZoneFailureDomainStable] = cloudIns.Zone
		})
	}

	if cloudIns.Region != "" {
		klog.Infof(
			"Adding node label from cloud provider: %s=%s, %s=%s",
			v1.LabelZoneRegion, cloudIns.Region,
			v1.LabelZoneRegionStable, cloudIns.Region,
		)
		nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
			if n.Labels == nil {
				n.Labels = map[string]string{}
			}
			n.Labels[v1.LabelZoneRegion] = cloudIns.Region
			n.Labels[v1.LabelZoneRegionStable] = cloudIns.Region
		})
	}

	return nodeModifiers, nil
}

func (cnc *CloudNodeController) setNodeCondition(n *v1.Node) error {
	for _, condition := range n.Status.Conditions {
		if condition.Type == v1.NodeNetworkUnavailable &&
			condition.Status == v1.ConditionFalse {
			klog.Infof("node %s network is available", n.Name)
			return nil
		}
	}
	return nodeutil.SetNodeCondition(cnc.kclient, types.NodeName(n.Name), v1.NodeCondition{
		Type:               v1.NodeNetworkUnavailable,
		Status:             v1.ConditionTrue,
		Reason:             "NoRouteCreated",
		Message:            "Node created without a route",
		LastTransitionTime: metav1.Now(),
	})
}

func findCloudTaint(taints []v1.Taint) *v1.Taint {
	for _, taint := range taints {
		if taint.Key == api.TaintExternalCloudProvider {
			return &taint
		}
	}
	return nil
}

func excludeTaintFromList(taints []v1.Taint, toExclude v1.Taint) []v1.Taint {
	var excluded []v1.Taint
	for _, taint := range taints {
		if toExclude.MatchTaint(&taint) {
			continue
		}
		excluded = append(excluded, taint)
	}
	return excluded
}

func removeCloudTaints(node *v1.Node) {
	// make sure only cloud node is processed
	cloudTaint := findCloudTaint(node.Spec.Taints)
	if cloudTaint == nil {
		klog.Infof("RemoveCloudTaints, Node %s is registered without "+
			"cloud taint. Will not process.", node.Name)
		return
	}
	node.Spec.Taints = excludeTaintFromList(node.Spec.Taints, *cloudTaint)
}

func nodeLists(kclient kubernetes.Interface) (*v1.NodeList, error) {
	allNodes, err := kclient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{ResourceVersion: "0"})
	if allNodes == nil {
		return nil, err
	}
	var nodes []v1.Node
	for _, node := range allNodes.Items {
		if utils.IsExcludedNode(&node) {
			continue
		}
		if node.Spec.ProviderID == "" {
			klog.Warningf("ignore node[%s] without providerid", node.Name)
			continue
		}
		nodes = append(nodes, node)
	}
	allNodes.Items = nodes
	return allNodes, err
}

func isNodeAddressChanged(addr1, addr2 []v1.NodeAddress) bool {
	if len(addr1) != len(addr2) {
		return true
	}
	addressMap1 := map[v1.NodeAddressType]string{}
	addressMap2 := map[v1.NodeAddressType]string{}

	for i := range addr1 {
		addressMap1[addr1[i].Type] = addr1[i].Address
		addressMap2[addr2[i].Type] = addr2[i].Address
	}

	for k, v := range addressMap1 {
		if addressMap2[k] != v {
			return true
		}
	}
	return false
}

func isProvidedAddrExist(node *v1.Node, nodeAddresses []v1.NodeAddress) (*v1.NodeAddress, bool) {
	var nodeIP *v1.NodeAddress
	ipExists := false
	addr, ok := node.ObjectMeta.Annotations[kubeletapis.AnnotationProvidedIPAddr]
	if ok {
		ipExists = true
		for i := range nodeAddresses {
			if nodeAddresses[i].Address == addr {
				nodeIP = &nodeAddresses[i]
				break
			}
		}
	}
	return nodeIP, ipExists
}

func broadcaster() (record.EventRecorder, record.EventBroadcaster) {
	caster := record.NewBroadcaster()
	caster.StartLogging(klog.Infof)
	source := v1.EventSource{Component: "node-controller"}
	return caster.NewRecorder(scheme.Scheme, source), caster
}

func PatchNode(
	kdm kubernetes.Interface,
	origined, patched *v1.Node,
) (*v1.Node, error) {

	originedData, err := json.Marshal(origined)
	if err != nil {
		return nil, err
	}
	patchedData, err := json.Marshal(patched)
	if err != nil {
		return nil, err
	}
	data, err := strategicpatch.CreateTwoWayMergePatch(originedData, patchedData, &v1.Node{})
	if err != nil {
		return nil, err
	}
	return kdm.
		CoreV1().
		Nodes().
		Patch(context.Background(), patched.Name, types.MergePatchType, data, metav1.PatchOptions{})
}
