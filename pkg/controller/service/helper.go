package service

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	v1 "k8s.io/api/core/v1"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	"os"
	"reflect"
	"strings"
)

func filterNodes(nodes []v1.Node, labels string) ([]v1.Node, error) {
	nodes, err := filterOutByLabel(nodes, labels)
	return nodes, err
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

func isLocalModeService(svc *v1.Service) bool {
	return svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeLocal
}

func isENIBackendType(svc *v1.Service) bool {
	if svc.Annotations[util.BACKEND_TYPE_LABEL] != "" {
		return svc.Annotations[util.BACKEND_TYPE_LABEL] == util.BACKEND_TYPE_ENI
	}

	if os.Getenv("SERVICE_FORCE_BACKEND_ENI") != "" {
		return os.Getenv("SERVICE_FORCE_BACKEND_ENI") == "true"
	}

	return ctx2.CFG.Global.ServiceBackendType == util.BACKEND_TYPE_ENI
}

func isSLBNeeded(svc *v1.Service) bool {
	// finalizer must be supported
	return svc.DeletionTimestamp == nil && svc.Spec.Type == v1.ServiceTypeLoadBalancer
}

func findENIbyAddrIP(resp *ecs.DescribeNetworkInterfacesResponse, addrIP string) (string, error) {
	for _, eni := range resp.NetworkInterfaceSets.NetworkInterfaceSet {
		for _, privateIpType := range eni.PrivateIpSets.PrivateIpSet {
			if addrIP == privateIpType.PrivateIpAddress {
				return eni.NetworkInterfaceId, nil
			}
		}
	}
	return "", fmt.Errorf("private ip address not found in openapi %s", addrIP)
}

func findNodeByNodeName(nodes []v1.Node, nodeName string) *v1.Node {
	for _, n := range nodes {
		if n.Name == nodeName {
			return &n
		}
	}
	klog.Infof("node %s not found ", nodeName)
	return nil
}

// TODO fix me
// providerID
// 1) the id of the instance in the alicloud API. Use '.' to separate providerID which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7'. The format of "REGION.NODEID"
// 2) the id for an instance in the kubernetes API, which has 'alicloud://' prefix. e.g. alicloud://cn-hangzhou.i-v98dklsmnxkkgiiil7
func nodeFromProviderID(providerID string) (string, string, error) {
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

func isExcludeNode(node *v1.Node) bool {
	if util.IsExcludedNode(node) {
		klog.Infof("ignore node with exclude node label %s", node.Name)
		return true
	}
	if _, exclude := node.Labels[LabelNodeRoleExcludeBalancer]; exclude {
		klog.Infof("ignore node with exclude balancer label %s", node.Name)
		return true
	}
	return false
}

func needLoadBalancer(service *v1.Service) bool {
	return service.Spec.Type == v1.ServiceTypeLoadBalancer
}

// MAX_BACKEND_NUM max batch backend num
const MAX_BACKEND_NUM = 39

type Func func([]interface{}) error

// Batch batch process `object` m with func `func`
// for general purpose
func Batch(m interface{}, cnt int, batch Func) error {
	if cnt <= 0 {
		cnt = MAX_BACKEND_NUM
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
