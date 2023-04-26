package backend

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Edge Backend
const (
	EdgeNodeLabel = "alibabacloud.com/is-edge-worker"
	EdgeNodeValue = "true"
	ENSLabel      = "alibabacloud.com/ens-instance-id"
)

// NewEdgeEndpoints Collect nodes that meet the conditions as the back end of the ELB Service
func NewEdgeEndpoints(reqCtx *svcCtx.RequestContext, kubeClient client.Client) (edgeBackend *EndpointWithENI, err error) {
	edgeBackend = new(EndpointWithENI)
	edgeBackend.setTrafficPolicy(reqCtx)
	if edgeBackend.TrafficPolicy == helper.ENITrafficPolicy {
		return edgeBackend, fmt.Errorf("the backend of elb can not adopt ENI traffic policy")
	}
	edgeBackend.setAddressIpVersion(reqCtx)
	if edgeBackend.AddressIPVersion == model.IPv6 {
		return edgeBackend, fmt.Errorf("the backend of elb can not adopt IPv6 address")
	}

	edgeBackend.Nodes, err = getEdgeENSNodes(reqCtx, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("get nodes error: %s", err.Error())
	}
	if utilfeature.DefaultMutableFeatureGate.Enabled(ctrlCfg.EndpointSlice) {
		edgeBackend.EndpointSlices, err = getEndpointByEndpointSlice(reqCtx, kubeClient, edgeBackend.AddressIPVersion)
		if err != nil {
			return nil, fmt.Errorf("get endpointslice error: %s", err.Error())
		}
		reqCtx.Log.Info("backend details", "endpointslices", helper.LogEndpointSliceList(edgeBackend.EndpointSlices))
	} else {
		edgeBackend.Endpoints, err = getEndpoints(reqCtx, kubeClient)
		if err != nil {
			return nil, fmt.Errorf("get endpoints error: %s", err.Error())
		}
		reqCtx.Log.Info("backend details", "endpoints", helper.LogEndpoints(edgeBackend.Endpoints))
	}
	return edgeBackend, nil
}

func getEdgeENSNodes(reqCtx *svcCtx.RequestContext, kubeClient client.Client) ([]v1.Node, error) {
	matchLabels := make(client.MatchingLabels)
	if reqCtx.Anno.Get(annotation.BackendLabel) != "" {
		var err error
		matchLabels, err = splitBackendLabel(reqCtx.Anno.Get(annotation.BackendLabel))
		if err != nil {
			return nil, fmt.Errorf("filter nodes by label %s error: %s", reqCtx.Anno.Get(annotation.BackendLabel), err.Error())
		}
	}
	matchLabels[EdgeNodeLabel] = EdgeNodeValue
	nodeList := v1.NodeList{}
	err := kubeClient.List(reqCtx.Ctx, &nodeList, client.HasLabels{ENSLabel}, matchLabels)
	if err != nil {
		return nil, fmt.Errorf("get edge ens node error: %s", err.Error())
	}
	return filterOutNodes(reqCtx, nodeList.Items), nil
}

func splitBackendLabel(labels string) (map[string]string, error) {
	if labels == "" {
		return nil, nil
	}
	ret := make(map[string]string)
	labelSlice := strings.Split(labels, ",")
	for _, v := range labelSlice {
		lb := strings.Split(v, "=")
		if len(lb) != 2 {
			return ret, fmt.Errorf("parse backend label: %s, [k1=v1,k2=v2]", v)
		}
		ret[lb[0]] = lb[1]
	}
	return ret, nil
}

func filterOutNodes(reqCtx *svcCtx.RequestContext, nodes []v1.Node) []v1.Node {
	condidateNodes := make([]v1.Node, 0)
	for _, node := range nodes {
		if shouldSkipNode(reqCtx, &node) {
			continue
		}
		condidateNodes = append(condidateNodes, node)
	}
	return condidateNodes
}

// skip cloud, master, vk, unscheduled(when enabled), notReady nodes
func shouldSkipNode(reqCtx *svcCtx.RequestContext, node *v1.Node) bool {
	// need to keep the node who has exclude label in order to be compatible with vk node
	// It's safe because these nodes will be filtered in build backends func

	if helper.IsMasterNode(node) {
		klog.Info("[%s] node %s is master node, skip adding it to lb", util.Key(reqCtx.Service), node.Name)
		return true
	}

	if helper.IsNodeExcludeFromEdgeLoadBalancer(node) {
		klog.Info("[%s] node %s is exclude node, skip adding it to lb", util.Key(reqCtx.Service), node.Name)
		return true
	}

	// filter unscheduled node
	if node.Spec.Unschedulable && reqCtx.Anno.Get(annotation.RemoveUnscheduled) != "" {
		if reqCtx.Anno.Get(annotation.RemoveUnscheduled) == string(model.OnFlag) {
			klog.Infof("[%s] node %s is unschedulable, skip adding to lb", util.Key(reqCtx.Service), node.Name)
			return true
		}
	}

	// ignore vk node condition check.
	// Even if the vk node is NotReady, it still can be added to lb. Because the eci pod that actually joins the lb, not a vk node
	if label, ok := node.Labels["type"]; ok && label == helper.LabelNodeTypeVK {
		return true
	}

	// If we have no info, don't accept
	if len(node.Status.Conditions) == 0 {
		return true
	}

	for _, cond := range node.Status.Conditions {
		// We consider the node for load balancing only when its NodeReady
		// condition status is ConditionTrue
		if cond.Type == v1.NodeReady &&
			cond.Status != v1.ConditionTrue {
			klog.Infof("[%s] node %v with %v condition "+
				"status %v", util.Key(reqCtx.Service), node.Name, cond.Type, cond.Status)
			return true
		}
	}

	return false
}
