package service

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// EndpointWithENI
// Currently, EndpointWithENI accept two kind of backend
// normal nodes of type []*v1.Node, and endpoints of type *v1.Endpoints
type EndpointWithENI struct {
	// LocalMode externalTraffic=Local
	LocalMode bool
	// BackendTypeENI
	// whether it is an eni backend
	BackendTypeENI bool
	// Nodes
	// contains all the candidate nodes consider of LoadBalance Backends.
	// Cloud implementation has the right to make any filter on it.
	Nodes []v1.Node

	// Endpoints
	// It is the direct pod location information which cloud implementation
	// may needed for some kind of filtering. eg. direct ENI attach.
	Endpoints *v1.Endpoints
}

func buildVGroupsFromService(c localModel) ([]model.VServerGroup, error) {
	var vgs []model.VServerGroup

	nodes, err := getNodes(c.ctx, c.kubeClient, c.anno)
	if err != nil {
		return vgs, fmt.Errorf("build vserver group error: %s", err.Error())
	}

	eps, err := getEndpoints(c.ctx, c.kubeClient, c.svc)

	candidates := EndpointWithENI{
		LocalMode:      isLocalModeService(c.svc),
		BackendTypeENI: isENIBackendType(c.svc),
		Nodes:          nodes,
		Endpoints:      eps,
	}

	for _, port := range c.svc.Spec.Ports {
		vg, err := buildVGroup(c, port, &candidates)
		if err != nil {
			return vgs, err
		}
		vgs = append(vgs, vg)
	}
	return vgs, nil
}

func getNodes(ctx context.Context, client client.Client, anno *AnnotationRequest) ([]v1.Node, error) {
	nodeList := v1.NodeList{}
	err := client.List(ctx, &nodeList)
	if err != nil {
		return nil, fmt.Errorf("get nodes error: %s", err.Error())
	}

	// 1. filter by label
	items := nodeList.Items
	if anno.Get(BackendLabel) != nil {
		items, err = filterOutByLabel(nodeList.Items, *anno.Get(BackendLabel))
		if err != nil {
			return nil, fmt.Errorf("filter nodes by label error: %s", err.Error())
		}
	}

	var nodes []v1.Node
	for _, n := range items {
		// 2. filter unscheduled node
		if n.Spec.Unschedulable && anno.Get(RemoveUnscheduled) != nil {
			if *anno.Get(RemoveUnscheduled) == string(model.OnFlag) {
				klog.Infof("ignore node %s with unschedulable condition", n.Name)
				continue
			}
		}

		// As of 1.6, we will taint the master, but not necessarily mark
		// it unschedulable. Recognize nodes labeled as master, and filter
		// them also, as we were doing previously.
		if _, isMaster := n.Labels[LabelNodeRoleMaster]; isMaster {
			continue
		}

		// ignore eci node condition check
		if label, ok := n.Labels["type"]; ok && label == ECINodeLabel {
			klog.Infof("ignoring eci node %v condition check", n.Name)
			nodes = append(nodes, n)
			continue
		}

		// If we have no info, don't accept
		if len(n.Status.Conditions) == 0 {
			continue
		}

		for _, cond := range n.Status.Conditions {
			// We consider the node for load balancing only when its NodeReady
			// condition status is ConditionTrue
			if cond.Type == v1.NodeReady &&
				cond.Status != v1.ConditionTrue {
				klog.Infof("ignoring node %v with %v condition "+
					"status %v", n.Name, cond.Type, cond.Status)
				continue
			}
		}

		nodes = append(nodes, n)
	}

	return items, nil
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

func getEndpoints(ctx context.Context, client client.Client, svc *v1.Service) (*v1.Endpoints, error) {
	eps := &v1.Endpoints{}
	err := client.Get(ctx, util.NamespacedName(svc), eps)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return eps, fmt.Errorf("get endpoints %s from k8s error: %s", util.Key(svc), err.Error())
		}
		klog.Warningf("endpoint not found: %s", util.Key(svc))
	}
	return eps, nil
}

func buildVGroup(c localModel, port v1.ServicePort, candidates *EndpointWithENI) (model.VServerGroup, error) {
	vg := model.VServerGroup{
		NamedKey: getVGroupNamedKey(c.svc, port),
	}
	vg.VGroupName = vg.NamedKey.Key()

	backends, err := getBackends(c, candidates, port)
	if err != nil {
		return vg, fmt.Errorf("get backends error: %s", err.Error())
	}
	// TODO set weight for backends

	vg.Backends = backends
	return vg, nil
}

func getBackends(c localModel, candidates *EndpointWithENI, port v1.ServicePort) ([]model.BackendAttribute, error) {
	var backends []model.BackendAttribute
	nodes := candidates.Nodes
	eps := candidates.Endpoints

	// ENI mode
	if candidates.BackendTypeENI {
		if len(eps.Subsets) == 0 {
			klog.Warningf("%s endpoint is nil in eni mode", util.Key(c.svc))
			return nil, nil
		}
		klog.Infof("[ENI] mode service: %s", util.Key(c.svc))
		var privateIpAddress []string
		for _, ep := range eps.Subsets {
			for _, addr := range ep.Addresses {
				privateIpAddress = append(privateIpAddress, addr.IP)
			}
		}
		err := Batch(privateIpAddress, 40, buildFunc(c, &backends))
		if err != nil {
			return backends, fmt.Errorf("batch process eni fail: %s", err.Error())
		}
	}
	// Local mode
	if candidates.LocalMode {
		if len(eps.Subsets) == 0 {
			klog.Warningf("%s endpoint is nil in local mode", util.Key(c.svc))
			return nil, nil
		}
		klog.Infof("[Local] mode service: %s", util.Key(c.svc))
		// 1. add duplicate ecs backends
		for _, sub := range eps.Subsets {
			for _, add := range sub.Addresses {
				if add.NodeName == nil {
					return nil, fmt.Errorf("add ecs backends for service[%s] error, NodeName is nil for ip %s ",
						util.Key(c.svc), add.IP)
				}
				node := findNodeByNodeName(nodes, *add.NodeName)
				if node == nil {
					klog.Warningf("can not find correspond node %s for endpoint %s", *add.NodeName, add.IP)
					continue
				}
				if isExcludeNode(node) {
					// filter vk node
					continue
				}
				_, id, err := nodeFromProviderID(node.Spec.ProviderID)
				if err != nil {
					return backends, fmt.Errorf("parse providerid: %s. "+
						"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
				}
				backends = append(
					backends,
					model.BackendAttribute{
						ServerId: id,
						Weight:   DEFAULT_SERVER_WEIGHT,
						Type:     "ecs",
						Port:     int(port.NodePort),
					},
				)
			}
		}
		// 2. add eci backends
		return addECIBackends(backends)
	}

	// Cluster mode
	// When ecs and eci are deployed in a cluster, add ecs first and then add eci
	klog.Infof("[Cluster] mode service: %s", util.Key(c.svc))
	// 1. add ecs backends
	for _, node := range nodes {
		if isExcludeNode(&node) {
			continue
		}
		_, id, err := nodeFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return backends, fmt.Errorf("normal parse providerid: %s. "+
				"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
		}

		backends = append(
			backends,
			model.BackendAttribute{
				ServerId: id,
				Weight:   DEFAULT_SERVER_WEIGHT,
				Type:     "ecs",
			},
		)
	}
	// 2. add eci backends
	return addECIBackends(backends)

}

func addECIBackends(backends []model.BackendAttribute) ([]model.BackendAttribute, error) {
	// TODO
	klog.Infof("implement me!")
	return backends, nil

}

func buildFunc(c localModel, backend *[]model.BackendAttribute) func(o []interface{}) error {

	// backend build function
	return func(o []interface{}) error {
		var ips []string
		for _, i := range o {
			ip, ok := i.(string)
			if !ok {
				return fmt.Errorf("not string: %v", i)
			}
			ips = append(ips, ip)
		}
		// TODO FIX ME
		resp, err := c.cloud.DescribeNetworkInterfaces(ctx2.CFG.Global.VpcID, &ips)
		if err != nil {
			return fmt.Errorf("call DescribeNetworkInterfaces: %s", err.Error())
		}
		for _, ip := range ips {
			eniid, err := findENIbyAddrIP(resp, ip)
			if err != nil {
				return err
			}
			*backend = append(
				*backend,
				model.BackendAttribute{
					ServerId: eniid,
					Weight:   DEFAULT_SERVER_WEIGHT,
					Type:     "eni",
					ServerIp: ip,
				},
			)
		}
		return nil
	}
}
