package backend

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func NewEndpointWithENI(reqCtx *svcCtx.RequestContext, kubeClient client.Client) (*EndpointWithENI, error) {
	endpointWithENI := &EndpointWithENI{}
	endpointWithENI.setTrafficPolicy(reqCtx)
	endpointWithENI.setAddressIpVersion(reqCtx)

	nodes, err := GetNodes(reqCtx, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("get nodes error: %s", err.Error())
	}
	endpointWithENI.Nodes = nodes

	if utilfeature.DefaultMutableFeatureGate.Enabled(ctrlCfg.EndpointSlice) {
		esList, err := getEndpointByEndpointSlice(reqCtx, kubeClient, endpointWithENI.AddressIPVersion)
		if err != nil {
			return nil, fmt.Errorf("get endpointslice error: %s", err.Error())
		}
		endpointWithENI.EndpointSlices = esList
		reqCtx.Log.Info("backend details", "endpointslices", helper.LogEndpointSliceList(esList))
	} else {
		eps, err := getEndpoints(reqCtx, kubeClient)
		if err != nil {
			return nil, fmt.Errorf("get endpoints error: %s", err.Error())
		}
		endpointWithENI.Endpoints = eps
		reqCtx.Log.Info("backend details", "endpoints", helper.LogEndpoints(eps))
	}

	return endpointWithENI, nil
}

// EndpointWithENI
// Currently, EndpointWithENI accept two kind of backend
// normal nodes of type []*v1.Node, and endpoints of type *v1.Endpoints
type EndpointWithENI struct {
	// TrafficPolicy
	// external traffic policy.
	TrafficPolicy helper.TrafficPolicy
	// AddressIPVersion
	// it indicates the address ip version type of the backends attached to the LoadBalancer
	AddressIPVersion model.AddressIPVersionType
	// Nodes
	// contains all the candidate nodes consider of LoadBalance Backends.
	// Cloud implementation has the right to make any filter on it.
	Nodes []v1.Node
	// Endpoints
	// It is the direct pod location information which cloud implementation
	// may needed for some kind of filtering. eg. direct ENI attach.
	Endpoints *v1.Endpoints
	// EndpointSlices
	// contains all the endpointslices of a service
	EndpointSlices []discovery.EndpointSlice
}

func (e *EndpointWithENI) setTrafficPolicy(reqCtx *svcCtx.RequestContext) {
	if helper.IsENIBackendType(reqCtx.Service) {
		e.TrafficPolicy = helper.ENITrafficPolicy
		return
	}
	if helper.IsLocalModeService(reqCtx.Service) {
		e.TrafficPolicy = helper.LocalTrafficPolicy
		return
	}
	e.TrafficPolicy = helper.ClusterTrafficPolicy
}

func (e *EndpointWithENI) setAddressIpVersion(reqCtx *svcCtx.RequestContext) {
	// Only EndpointSlice support dual stack.
	// Enable IPv6DualStack and EndpointSlice feature gates if you want to use ipv6 backends
	if utilfeature.DefaultMutableFeatureGate.Enabled(ctrlCfg.IPv6DualStack) &&
		utilfeature.DefaultMutableFeatureGate.Enabled(ctrlCfg.EndpointSlice) &&
		reqCtx.Anno.Get(annotation.IPVersion) == string(model.IPv6) &&
		reqCtx.Anno.Get(annotation.BackendIPVersion) == string(model.IPv6) {
		e.AddressIPVersion = model.IPv6
		reqCtx.Log.Info("backend address ip version is ipv6")
		return
	}
	e.AddressIPVersion = model.IPv4
}

func GetNodes(reqCtx *svcCtx.RequestContext, client client.Client) ([]v1.Node, error) {
	nodeList := v1.NodeList{}
	err := client.List(reqCtx.Ctx, &nodeList)
	if err != nil {
		return nil, fmt.Errorf("get nodes error: %s", err.Error())
	}

	// 1. filter by label
	items := nodeList.Items
	if reqCtx.Anno.Get(annotation.BackendLabel) != "" {
		items, err = filterOutByLabel(nodeList.Items, reqCtx.Anno.Get(annotation.BackendLabel))
		if err != nil {
			return nil, fmt.Errorf("filter nodes by label error: %s", err.Error())
		}
	}

	var nodes []v1.Node
	for _, n := range items {
		if needExcludeFromLB(reqCtx, &n) {
			continue
		}
		nodes = append(nodes, n)
	}

	return nodes, nil
}

func filterOutByLabel(nodes []v1.Node, labels string) ([]v1.Node, error) {
	if labels == "" {
		return nodes, nil
	}
	var result []v1.Node
	lbl := strings.Split(labels, ",")
	var records []string
	for _, node := range nodes {
		found := true
		for _, v := range lbl {
			l := strings.Split(v, "=")
			if len(l) < 2 {
				return []v1.Node{}, fmt.Errorf("parse backend label: %s, [k1=v1,k2=v2]", v)
			}
			if nv, exist := node.Labels[l[0]]; !exist || nv != l[1] {
				found = false
				break
			}
		}
		if found {
			result = append(result, node)
			records = append(records, node.Name)
		}
	}
	klog.V(4).Infof("accept nodes backend labels[%s], %v", labels, records)
	return result, nil
}

func needExcludeFromLB(reqCtx *svcCtx.RequestContext, node *v1.Node) bool {
	// need to keep the node who has exclude label in order to be compatible with vk node
	// It's safe because these nodes will be filtered in build backends func

	if helper.IsMasterNode(node) {
		klog.V(5).Infof("[%s] node %s is master node, skip adding it to lb", util.Key(reqCtx.Service), node.Name)
		return true
	}

	// Remove nodes that are about to be deleted by the cluster autoscaler.
	for _, taint := range node.Spec.Taints {
		if taint.Key == helper.ToBeDeletedTaint {
			klog.Infof("Ignoring node %v with autoscaler taint %+v", node.Name, taint)
			return true
		}
	}

	// filter unscheduled node
	if node.Spec.Unschedulable && reqCtx.Anno.Get(annotation.RemoveUnscheduled) != "" {
		if reqCtx.Anno.Get(annotation.RemoveUnscheduled) == string(model.OnFlag) {
			reqCtx.Log.Info("node is unschedulable, skip add to lb", "node", node.Name)
			return true
		}
	}

	// ignore vk node condition check.
	// Even if the vk node is NotReady, it still can be added to lb. Because the eci pod that actually joins the lb, not a vk node
	if label, ok := node.Labels["type"]; ok && label == helper.LabelNodeTypeVK {
		return false
	}

	// If we have no info, don't accept
	if len(node.Status.Conditions) == 0 {
		reqCtx.Log.Info("node condition is nil, skip add to lb", "node", node.Name)
		return true
	}

	for _, cond := range node.Status.Conditions {
		// We consider the node for load balancing only when its NodeReady
		// condition status is ConditionTrue
		if cond.Type == v1.NodeReady &&
			cond.Status != v1.ConditionTrue {
			reqCtx.Log.Info(fmt.Sprintf("node not ready with %v condition, status %v", cond.Type, cond.Status),
				"node", node.Name)
			return true
		}
	}

	return false
}

func getEndpoints(reqCtx *svcCtx.RequestContext, client client.Client) (*v1.Endpoints, error) {
	eps := &v1.Endpoints{}
	err := client.Get(reqCtx.Ctx, util.NamespacedName(reqCtx.Service), eps)
	if err != nil && apierrors.IsNotFound(err) {
		reqCtx.Log.Info("warning: endpoint not found")
		return eps, nil
	}
	return eps, err
}

func getEndpointByEndpointSlice(reqCtx *svcCtx.RequestContext, kubeClient client.Client, ipVersion model.AddressIPVersionType) ([]discovery.EndpointSlice, error) {
	epsList := &discovery.EndpointSliceList{}
	err := kubeClient.List(reqCtx.Ctx, epsList, client.MatchingLabels{
		discovery.LabelServiceName: reqCtx.Service.Name,
	}, client.InNamespace(reqCtx.Service.Namespace))
	if err != nil {
		return nil, err
	}

	addressType := discovery.AddressTypeIPv4
	if ipVersion == model.IPv6 {
		addressType = discovery.AddressTypeIPv6
	}

	var ret []discovery.EndpointSlice
	for _, es := range epsList.Items {
		if es.AddressType == addressType {
			ret = append(ret, es)
		}
	}

	return ret, nil
}

// MAX_BACKEND_NUM max batch backend num
const (
	MaxBackendNum = 39
)

type Func func([]interface{}) error

// Batch batch process `object` m with func `func`
// for general purpose
func Batch(m interface{}, cnt int, batch Func) error {
	if cnt <= 0 {
		cnt = MaxBackendNum
	}
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("non-slice type for %v", m)
	}

	// need to convert interface to []interface
	// see https://github.com/golang/go/wiki/InterfaceSlice
	target := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		target[i] = v.Index(i).Interface()
	}
	klog.Infof("batch process ,total length %d", len(target))
	for len(target) > cnt {
		if err := batch(target[0:cnt]); err != nil {

			return err
		}
		target = target[cnt:]
	}
	if len(target) <= 0 {
		return nil
	}

	klog.Infof("batch process ,total length %d last section", len(target))
	return batch(target)
}
