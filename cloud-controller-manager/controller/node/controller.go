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
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"github.com/golang/glog"

	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/route"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	nodeutilv1 "k8s.io/kubernetes/pkg/api/v1/node"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/pkg/util/metrics"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
	"k8s.io/kubernetes/plugin/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/kubernetes"
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
}

// CloudInstance is an interface to interact with cloud api
type CloudInstance interface {
	// SetInstanceTags set instance tags for instance id
	SetInstanceTags(insid string, tags map[string]string) error

	// ListInstances list instance by given ids.
	ListInstances(ids []string) (map[string]*CloudNodeAttribute, error)
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
		glog.Infof("start sending events to api server.")
	} else {
		glog.Infof("no api server defined - no events will be sent to API server.")
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
				glog.V(4).Infof("receive node add event: %s", node.Name)
				err := cnc.AddCloudNode(node)
				if err != nil {
					glog.Errorf("remove cloud node taints fail: %s", err.Error())
				}
			},
		},
	)
}

// This controller deletes a node if kubelet is not reporting
// and the node is gone from the cloud provider.
func (cnc *CloudNodeController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	glog.Info("starting node controller")

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
				glog.Errorf("Error monitoring node status: %v", err)
				return
			}

			// ignore return value, retry on error
			err = batchAddressUpdate(
				nodes.Items,
				cnc.syncNodeAddress,
			)
			if err != nil {
				glog.Errorf("periodically update address: %s", err.Error())
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
				glog.Errorf("Error monitoring node status: %v", err)
				return
			}
			// ignore return value, retry on error
			err = batchAddressUpdate(
				nodes.Items,
				cnc.syncCloudNodes,
			)
			if err != nil {
				glog.Errorf("periodically try detect node existence: %s", err.Error())
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
				glog.Errorf("Error monitoring node status: %v", err)
				return
			}
			for _, node := range nodes.Items {
				err := cnc.AddCloudNode(&node)
				if err != nil {
					glog.Errorf("periodically remove cloud node %s taints: %s", node.Name, err.Error())
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
		Get(node.Name, metav1.GetOptions{})
	if err != nil {
		//retry
		return fmt.Errorf("retrieve node error: %s", err.Error())
	}
	cloudTaint := findCloudTaint(curNode.Spec.Taints)
	if cloudTaint == nil {
		glog.V(4).Infof("Node %s is registered without cloud taint. Will not process.", node.Name)
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
	instances, err := ins.ListInstances(nodeids(nodes))
	if err != nil {
		return fmt.Errorf("syncNodeAddress, retrieve instances from api error: %s", err.Error())
	}

	for i := range nodes {
		node := &nodes[i]
		cloudNode := instances[node.Spec.ProviderID]
		if cloudNode == nil {
			glog.Infof("node %s not found, skip update node address", node.Spec.ProviderID)
			continue
		}
		cloudNode.Addresses = setHostnameAddress(node, cloudNode.Addresses)
		// If nodeIP was suggested by user, ensure that
		// it can be found in the cloud as well (consistent with the behaviour in kubelet)
		nodeIP, ok := isProvidedAddrExist(node, cloudNode.Addresses)
		if ok {
			if nodeIP == nil {
				glog.Errorf("User has specified Node IP in kubelet but it is not found in cloudprovider")
				continue
			}
			// override addresss
			cloudNode.Addresses = []v1.NodeAddress{*nodeIP}
		}
		err := tryPatchNodeAddress(cnc.kclient, node, cloudNode.Addresses)
		if err != nil {
			glog.Error("Wait for next retry, patch node address error: %s", err.Error())
		}
	}
	return nil
}

func (cnc *CloudNodeController) syncCloudNodes(nodes []v1.Node) error {
	ins, ok := cnc.cloud.(CloudInstance)
	if !ok {
		return fmt.Errorf("cloud instance not implemented")
	}
	instances, err := ins.ListInstances(nodeids(nodes))
	if err != nil {
		return fmt.Errorf("syncCloudNodes, retrieve instances from api error: %s", err.Error())
	}

	for i := range nodes {
		node := &nodes[i]

		condition := nodeConditionReady(cnc.kclient, node)
		if condition == nil {
			glog.Infof("node condition not ready, wait for next retry", node.Spec.ProviderID)
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

		glog.Infof("node %s not found, start to delete from meta", node.Spec.ProviderID)
		// try delete node and ignore error, retry next loop
		deleteNode(cnc, node)
	}
	return nil
}

// This processes nodes that were added into the cluster, and cloud initialize them if appropriate
func (cnc *CloudNodeController) doAddCloudNode(node *v1.Node) error {

	ins, ok := cnc.cloud.(CloudInstance)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("failed to get ins from cloud provider"))
		return fmt.Errorf("cloud instance is not implemented")
	}
	err := wait.PollImmediate(
		2*time.Second,
		20*time.Second,
		func() (done bool, err error) {
			glog.V(5).Infof("try remove cloud taints for %s", node.Name)
			curNode, err := cnc.kclient.CoreV1().Nodes().Get(node.Name, metav1.GetOptions{})
			if err != nil {
				glog.Errorf("retrieve node error: %s", err.Error())
				//retry
				return false, nil
			}
			orignode := curNode.DeepCopy()
			setDefaultProviderID(cnc, curNode)

			nodes, err := ins.ListInstances([]string{curNode.Spec.ProviderID})
			if err != nil {
				glog.Errorf("cloud instance api fail, %s", err.Error())
				//retry
				return false, nil
			}
			cloudins, ok := nodes[curNode.Spec.ProviderID]
			if !ok || cloudins == nil {
				return false, fmt.Errorf("instance not found")
			}

			// If user provided an IP address, ensure that IP address is found
			// in the cloud provider before removing the taint on the node
			nodeIP, ok := isProvidedAddrExist(curNode, cloudins.Addresses)
			if ok && nodeIP == nil {
				glog.Errorf("failed to get specified nodeIP in cloudprovider, fail fast")
				return true, nil
			}

			if cloudins.InstanceType != "" {
				glog.Infof("Adding node label from cloud "+
					"provider: %s=%s", kubeletapis.LabelInstanceType, cloudins.InstanceType)
				curNode.ObjectMeta.Labels[kubeletapis.LabelInstanceType] = cloudins.InstanceType
			}

			// TODO(wlan0): Move this logic to the route controller using the node taint instead of condition
			// Since there are node taints, do we still need this?
			// This condition marks the node as unusable until routes are initialized in the cloud provider
			// Aoxn: Hack for alibaba cloud
			if route.Options.ConfigCloudRoutes &&
				cnc.cloud.ProviderName() == "alicloud" {
				curNode.Status.Conditions = append(
					node.Status.Conditions,
					v1.NodeCondition{
						Type:               v1.NodeNetworkUnavailable,
						Status:             v1.ConditionTrue,
						Reason:             "NoRouteCreated",
						Message:            "Node created without a route",
						LastTransitionTime: metav1.Now(),
					},
				)
			}

			if err = setFailureZones(cnc.cloud, curNode); err != nil {
				glog.Errorf("set failed zone error: %s", err.Error())
				//retry
				return false, nil
			}
			removeCloudTaints(curNode)

			err = ins.SetInstanceTags(
				cloudins.InstanceID,
				map[string]string{
					"k8s.aliyun.com": "true",
					"kubernetes.ccm": "true",
				},
			)
			if err != nil {
				if !strings.Contains(err.Error(), "Forbidden.RAM") {
					glog.Errorf("tag instance %s error: %s", cloudins.InstanceID, err.Error())
					//retry
					return false, nil
				}
				// Old ROS template does not have AddTags Permission.
				// It is ok to skip `Forbidden` error for compatible reason.
			}

			nnode, err := PatchNode(cnc.kclient, orignode, curNode)
			if err != nil {
				glog.Errorf("patch error: %s", err.Error())
				return false, nil
			}
			glog.V(5).Infof("finished remove uninitialized cloud taints for %s", node.Name)
			// After adding, call UpdateNodeAddress to set the CloudProvider provided IPAddresses
			// So that users do not see any significant delay in IP addresses being filled into the node
			_ = cnc.syncNodeAddress([]v1.Node{*nnode})
			return true, nil
		},
	)
	if err != nil {
		glog.Errorf("doAddCloudNode error: %s", err.Error())
		utilruntime.HandleError(err)
		return err
	}
	glog.Infof("Successfully initialized node %s with cloud provider", node.Name)
	return nil
}

func batchAddressUpdate(
	nodes []v1.Node,
	batch func([]v1.Node) error,
) error {

	glog.Infof("batch process update node address, length %d", len(nodes))
	for len(nodes) > MAX_BATCH_NUM {
		if err := batch(nodes[0:MAX_BATCH_NUM]); err != nil {
			glog.Errorf("batch process func error: %s", err.Error())
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
	glog.Infof("try patch node address for %s", node.Spec.ProviderID)
	_, err := nodeutil.PatchNodeStatus(
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
	glog.V(2).Infof("recording %s event message for node %s", "DeletingNode", node.Name)

	msg := fmt.Sprintf("Deleting Node %v because it's not present according to cloud provider", node.Name)

	cnc.recorder.Eventf(
		ref,
		v1.EventTypeNormal,
		msg,
		"Node %s event: %s",
		node.Name,
		"DeletingNode",
	)

	go func(nodeName string) {
		defer utilruntime.HandleCrash()
		if err := cnc.kclient.CoreV1().Nodes().Delete(nodeName, nil); err != nil {
			glog.Errorf("unable to delete node %q: %v", nodeName, err)
		}
	}(node.Name)
}

func nodeConditionReady(kclient kubernetes.Interface, node *v1.Node) *v1.NodeCondition {
	// Try to get the current node status
	// If node status is empty, then kubelet has not posted ready status yet.
	// In this case, process next node
	var err error
	for rep := 0; rep < RETRY_COUNT; rep++ {
		_, ccondition := nodeutilv1.GetNodeCondition(&node.Status, v1.NodeReady)
		if ccondition != nil {
			return ccondition
		}
		name := node.Name
		node, err = kclient.CoreV1().Nodes().Get(name, metav1.GetOptions{})
		if err != nil {
			glog.Errorf("Failed while getting a Node to retry updating "+
				"NodeStatus. Probably Node %s was deleted.", name)
			break
		}
		time.Sleep(retrySleepTime)
	}
	return nil
}

func setDefaultProviderID(cnc *CloudNodeController, node *v1.Node) {

	if node.Spec.ProviderID != "" {
		return
	}
	id, err := cloudprovider.GetInstanceProviderID(cnc.cloud, types.NodeName(node.Name))
	if err == nil {
		node.Spec.ProviderID = id
	} else {
		// we should attempt to set providerID on curNode, but
		// we can continue if we fail since we will attempt to set
		// node addresses given the node name in getNodeAddressesByProviderIDOrName
		glog.Errorf("failed to set node provider id: %v", err)
	}
}

func setFailureZones(cloud cloudprovider.Interface, node *v1.Node) error {
	zones, ok := cloud.Zones()
	if !ok {
		return nil
	}
	zone, err := zones.GetZoneByProviderID(node.Spec.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to get zone from cloud provider: %v", err)
	}
	if zone.FailureDomain != "" {
		glog.Infof("Adding node label from cloud "+
			"provider: %s=%s", kubeletapis.LabelZoneFailureDomain, zone.FailureDomain)
		node.ObjectMeta.Labels[kubeletapis.LabelZoneFailureDomain] = zone.FailureDomain
	}
	if zone.Region != "" {
		glog.Infof("Adding node label from cloud "+
			"provider: %s=%s", kubeletapis.LabelZoneRegion, zone.Region)
		node.ObjectMeta.Labels[kubeletapis.LabelZoneRegion] = zone.Region
	}
	return nil
}

func findCloudTaint(taints []v1.Taint) *v1.Taint {
	for _, taint := range taints {
		if taint.Key == algorithm.TaintExternalCloudProvider {
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
		glog.Infof("RemoveCloudTaints, Node %s is registered without "+
			"cloud taint. Will not process.", node.Name)
		return
	}
	node.Spec.Taints = excludeTaintFromList(node.Spec.Taints, *cloudTaint)
}

func nodeLists(kclient kubernetes.Interface) (*v1.NodeList, error) {
	allNodes, err := kclient.CoreV1().Nodes().List(metav1.ListOptions{ResourceVersion: "0"})
	if allNodes == nil {
		return nil, err
	}
	var nodes []v1.Node
	for _, node := range allNodes.Items {
		if _, exclude := node.Labels[utils.LabelNodeRoleExcludeNode]; exclude {
			continue
		}
		if node.Spec.ProviderID == "" {
			glog.Warningf("ignore node[%s] without providerid", node.Name)
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
	caster.StartLogging(glog.Infof)
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
		Patch(patched.Name, types.MergePatchType, data)
}
