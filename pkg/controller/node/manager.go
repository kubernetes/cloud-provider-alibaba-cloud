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
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider/api"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

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

const (
	ecsTagNodePoolID = "ack.alibabacloud.com/nodepool-id"

	LabelNodePoolID         = "node.alibabacloud.com/nodepool-id"
	LabelInstanceChargeType = "node.alibabacloud.com/instance-charge-type"
	LabelSpotStrategy       = "node.alibabacloud.com/spot-strategy"
)

var ErrNotFound = errors.New("instance not found")

type nodeModifier func(*v1.Node)

func batchOperate(
	nodes []v1.Node,
	batch func([]v1.Node) error,
) error {
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

func deleteNode(cnc *ReconcileNode, node *v1.Node) {

	ref := &v1.ObjectReference{
		Kind:      "Node",
		Name:      node.Name,
		UID:       node.UID,
		Namespace: "",
	}

	deleteOne := func() {
		defer utilruntime.HandleCrash()
		err := cnc.client.Delete(
			context.Background(), node,
		)
		if err != nil {
			log.Error(err, "failed to delete node", "node", node.Name, "prvdId", node.Spec.ProviderID)
			cnc.record.Event(
				node, v1.EventTypeWarning, helper.FailedDeleteNode,
				fmt.Sprintf("Error deleting node: %s", helper.GetLogMessage(err)),
			)
			return
		}
		cnc.record.Event(ref, v1.EventTypeNormal, helper.SucceedDeleteNode, node.Name)
		log.Info("delete node from cluster successfully", "node", node.Name, "prvdId", node.Spec.ProviderID)
	}
	go deleteOne()
}

func nodeConditionReady(kclient client.Client, node *v1.Node) *v1.NodeCondition {
	// Try to get the current node status
	// If node status is empty, then kubelet has not posted ready status yet.
	// In this case, process next node
	var err error
	for rep := 0; rep < RETRY_COUNT; rep++ {
		ccondition, ok := helper.FindCondition(node.Status.Conditions, v1.NodeReady)
		if ok {
			return ccondition
		}
		err = kclient.Get(context.Background(), client.ObjectKey{Name: node.Name}, node)
		if err != nil {
			klog.Errorf("Failed while getting a Node to retry updating "+
				"NodeStatus. Probably Node %s was deleted.", node.Name)
			break
		}
		time.Sleep(retrySleepTime)
	}
	return nil
}

func findCloudECS(ins prvd.IInstance, node *v1.Node) (*prvd.NodeAttribute, error) {
	if node.Spec.ProviderID != "" {
		return findCloudECSById(ins, node)
	} else {
		return findCloudECSByIp(ins, node)
	}
}

func findCloudECSById(ins prvd.IInstance, node *v1.Node) (*prvd.NodeAttribute, error) {
	klog.Infof("try to find ecs for node %s by provider id %s", node.Name, node.Spec.ProviderID)
	nodes, err := ins.ListInstances(context.TODO(), []string{node.Spec.ProviderID})
	if err != nil {
		return nil, fmt.Errorf("cloud instance api fail, %s", err.Error())
	}
	cloudIns, ok := nodes[node.Spec.ProviderID]
	if !ok || cloudIns == nil {
		return nil, ErrNotFound
	}
	return cloudIns, nil
}

func findCloudECSByIp(ins prvd.IInstance, node *v1.Node) (*prvd.NodeAttribute, error) {
	var internalIP []string
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			internalIP = append(internalIP, addr.Address)
		}
	}
	if len(internalIP) != 1 {
		klog.Infof("try to find ecs for node %s by internal ip, error: internal ip [%v] not single",
			internalIP, node.Name)
		return nil, ErrNotFound
	}

	klog.Infof("try to find ecs for node %s by internal ip %v", node.Name, internalIP)
	cloudIns, err := ins.GetInstancesByIP(context.TODO(), internalIP)
	if err != nil {
		return nil, fmt.Errorf("list instances by ip fail: %s", err.Error())
	}
	if cloudIns == nil {
		return nil, ErrNotFound
	}
	return cloudIns, nil
}

func setFields(node *v1.Node, ins *prvd.NodeAttribute, cfgRoute bool) {

	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	var modifiers []nodeModifier
	if ins.InstanceType != "" {
		klog.V(5).Infof(
			"node %s, Adding node label from cloud provider: %s=%s, %s=%s",
			node.Name,
			v1.LabelInstanceType, ins.InstanceType,
			v1.LabelInstanceTypeStable, ins.InstanceType,
		)
		modify := func(n *v1.Node) {
			n.Labels[v1.LabelInstanceType] = ins.InstanceType
			n.Labels[v1.LabelInstanceTypeStable] = ins.InstanceType
		}
		modifiers = append(modifiers, modify)
	}

	if ins.Zone != "" {
		klog.V(5).Infof(
			"node %s, Adding node label from cloud provider: %s=%s, %s=%s",
			node.Name,
			v1.LabelZoneFailureDomain, ins.Zone,
			v1.LabelTopologyZone, ins.Zone,
		)
		modify := func(n *v1.Node) {
			n.Labels[v1.LabelZoneFailureDomain] = ins.Zone
			n.Labels[v1.LabelTopologyZone] = ins.Zone
		}
		modifiers = append(modifiers, modify)
	}

	if ins.Region != "" {
		klog.V(5).Infof(
			"node %s,Adding node label from cloud provider: %s=%s, %s=%s",
			node.Name,
			v1.LabelZoneRegion, ins.Region,
			v1.LabelTopologyRegion, ins.Region,
		)
		modify := func(n *v1.Node) {
			n.Labels[v1.LabelZoneRegion] = ins.Region
			n.Labels[v1.LabelTopologyRegion] = ins.Region
		}
		modifiers = append(modifiers, modify)
	}

	if node.Spec.ProviderID == "" && ins.InstanceID != "" {
		prvdId := fmt.Sprintf("%s.%s", ins.Region, ins.InstanceID)
		klog.V(5).Infof(
			"node %s,Adding provider id from cloud provider: %s",
			node.Name,
			prvdId,
		)
		modify := func(n *v1.Node) {
			n.Spec.ProviderID = prvdId
		}
		modifiers = append(modifiers, modify)
	}

	if ins.InstanceChargeType != "" {
		klog.V(5).Infof(
			"node %s, Adding node label from cloud provider: %s=%s",
			node.Name,
			LabelInstanceChargeType, ins.InstanceChargeType,
		)
		modify := func(n *v1.Node) {
			n.Labels[LabelInstanceChargeType] = ins.InstanceChargeType
		}
		modifiers = append(modifiers, modify)
	}

	if ins.SpotStrategy != "" {
		klog.V(5).Infof(
			"node %s, Adding node label from cloud provider: %s=%s",
			node.Name,
			LabelSpotStrategy, ins.SpotStrategy)
		modify := func(n *v1.Node) {
			n.Labels[LabelSpotStrategy] = ins.SpotStrategy
		}
		modifiers = append(modifiers, modify)
	}

	if nodePoolID, ok := ins.Tags[ecsTagNodePoolID]; ok {
		klog.V(5).Infof(
			"node %s, Adding node label from cloud provider: %s=%s",
			node.Name,
			LabelNodePoolID, nodePoolID,
		)
		modify := func(n *v1.Node) {
			n.Labels[LabelNodePoolID] = nodePoolID
		}
		modifiers = append(modifiers, modify)
	}

	modifiers = append(modifiers, removeCloudTaints)

	if cfgRoute && !helper.HasExcludeLabel(node) {

		modifiers = append(modifiers, setNetworkUnavailable)
	}

	for _, modify := range modifiers {
		modify(node)
	}
}

func setNetworkUnavailable(n *v1.Node) {
	var conditions []v1.NodeCondition

	for _, condition := range n.Status.Conditions {
		if condition.Type == v1.NodeNetworkUnavailable {
			if condition.Status == v1.ConditionFalse {
				klog.Infof("node %s network is available, skip patch node status", n.Name)
				return
			}
			continue
		}
		conditions = append(conditions, condition)
	}

	con := v1.NodeCondition{
		Type:               v1.NodeNetworkUnavailable,
		Status:             v1.ConditionTrue,
		Reason:             "NoRouteCreated",
		Message:            "Node created without a route",
		LastTransitionTime: metav1.Now(),
	}
	conditions = append(conditions, con)
	n.Status.Conditions = conditions
}

func removeCloudTaints(node *v1.Node) {
	// make sure only cloud node is processed
	cloudTaint := findCloudTaint(node.Spec.Taints)
	if cloudTaint == nil {
		klog.V(5).Infof("node %s is registered without "+
			"cloud taint. skip.", node.Name)
		return
	}
	node.Spec.Taints = excludeTaintFromList(node.Spec.Taints, *cloudTaint)
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

func NodeList(kclient client.Client) (*v1.NodeList, error) {
	nodes := &v1.NodeList{}
	err := kclient.List(context.TODO(), nodes, &client.ListOptions{
		Raw: &metav1.ListOptions{
			ResourceVersion: "0",
		},
	})
	if err != nil {
		return nil, err
	}
	var mnodes []v1.Node
	for _, node := range nodes.Items {
		if helper.HasExcludeLabel(&node) {
			continue
		}
		if node.Spec.ProviderID == "" {
			klog.Warningf("ignore node[%s] without providerid", node.Name)
			continue
		}
		mnodes = append(mnodes, node)
	}
	nodes.Items = mnodes
	return nodes, nil
}

func isProvidedAddrExist(node *v1.Node, nodeAddresses []v1.NodeAddress) (*v1.NodeAddress, bool) {
	var nodeIP *v1.NodeAddress
	ipExists := false
	addr, ok := node.ObjectMeta.Annotations[api.AnnotationAlphaProvidedIPAddr]
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
