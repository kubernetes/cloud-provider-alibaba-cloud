package helper

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/types"

	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
)

type TrafficPolicy string

const (
	LocalTrafficPolicy   = TrafficPolicy("Local")
	ClusterTrafficPolicy = TrafficPolicy("Cluster")
	ENITrafficPolicy     = TrafficPolicy("ENI")
)

func GetService(k8sClient client.Client, svcKey types.NamespacedName) (*v1.Service, error) {
	svc := &v1.Service{}
	err := k8sClient.Get(context.Background(), svcKey, svc)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func GetServiceTrafficPolicy(svc *v1.Service) (TrafficPolicy, error) {
	if isENIBackendType(svc) {
		return ENITrafficPolicy, nil
	}
	if isClusterIPService(svc) {
		return "", fmt.Errorf("cluster service type just support eni mode for alb ingress")
	}
	if isLocalModeService(svc) {
		return LocalTrafficPolicy, nil
	}
	return ClusterTrafficPolicy, nil
}

func isENIBackendType(svc *v1.Service) bool {
	if svc.Annotations[annotations.BackendType] != "" {
		return svc.Annotations[annotations.BackendType] == alb.ENIBackendType
	}

	if os.Getenv("SERVICE_FORCE_BACKEND_ENI") != "" {
		return os.Getenv("SERVICE_FORCE_BACKEND_ENI") == "true"
	}

	return strings.EqualFold(ctrlCfg.CloudCFG.Global.ServiceBackendType, alb.ENIBackendType)
}

type Func func([]interface{}) error

func isLocalModeService(svc *v1.Service) bool {
	return svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeLocal
}

func isClusterIPService(svc *v1.Service) bool {
	return svc.Spec.Type == v1.ServiceTypeClusterIP
}

func NodeFromProviderID(providerID string) (string, string, error) {
	if strings.HasPrefix(providerID, "alicloud://") {
		k8sName := strings.Split(providerID, "://")
		if len(k8sName) < 2 {
			return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
		} else {
			providerID = k8sName[1]
		}
	}

	name := strings.Split(providerID, ".")
	if len(name) < 2 {
		return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
	}
	return name[0], name[1], nil
}

func GetNodes(svc *v1.Service, client client.Client) ([]v1.Node, error) {
	nodeList := v1.NodeList{}
	err := client.List(context.Background(), &nodeList)
	if err != nil {
		return nil, fmt.Errorf("get nodes error: %v", err)
	}

	// 1. filter by label
	items := nodeList.Items
	if a, ok := svc.Annotations[annotations.BackendLabel]; ok {
		items, err = filterOutByLabel(nodeList.Items, a)
		if err != nil {
			return nil, fmt.Errorf("filter nodes by label error: %s", err.Error())
		}
	}

	var nodes []v1.Node
	for _, n := range items {
		if needExcludeFromLB(svc, &n) {
			continue
		}
		nodes = append(nodes, n)
	}

	return nodes, nil
}

func needExcludeFromLB(svc *v1.Service, node *v1.Node) bool {
	// need to keep the node who has exclude label in order to be compatible with vk node
	// It's safe because these nodes will be filtered in build backends func

	// exclude node which has exclude balancer label
	if _, exclude := node.Labels[util.LabelNodeExcludeApplicationLoadBalancer]; exclude {
		return true
	}

	if isMasterNode(node) {
		klog.Infof("[%s] node %s is master node, skip adding it to lb", util.Key(svc), node.Name)
		return true
	}

	// filter unscheduled node
	if node.Spec.Unschedulable {
		if v, ok := svc.Annotations[annotations.RemoveUnscheduled]; ok {
			if v == string(alb.OnFlag) {
				klog.Infof("[%s] node %s is unscheduled, skip adding to lb", util.Key(svc), node.Name)
				return true
			}
		}
	}

	// ignore vk node condition check.
	// Even if the vk node is NotReady, it still can be added to lb. Because the eci pod that actually joins the lb, not a vk node
	if label, ok := node.Labels["type"]; ok && label == util.LabelNodeTypeVK {
		return false
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
				"status %v", util.Key(svc), node.Name, cond.Type, cond.Status)
			return true
		}
	}

	return false
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

func isMasterNode(node *v1.Node) bool {
	if _, isMaster := node.Labels[util.LabelNodeRoleMaster]; isMaster {
		return true
	}
	return false
}
