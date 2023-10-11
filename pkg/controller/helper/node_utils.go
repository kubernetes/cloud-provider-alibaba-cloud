package helper

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	PatchAll    = "all"
	PatchSpec   = "spec"
	PatchStatus = "status"
)

const (
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"
	LabelNodeTypeVK     = "virtual-kubelet"
	// LabelNodeExcludeBalancer specifies that the node should be
	// exclude from loadbalancers created by a cloud provider.
	LabelNodeExcludeBalancerDeprecated = "alpha.service-controller.kubernetes.io/exclude-balancer"
	LabelNodeExcludeBalancer           = v1.LabelNodeExcludeBalancers
	// ToBeDeletedTaint is a taint used by the CLuster Autoscaler before marking a node for deletion.
	// Details in https://github.com/kubernetes/cloud-provider/blob/5bb9b27442bcb2613a9ca4046c89109de4435824/controllers/service/controller.go#L58
	ToBeDeletedTaint = "ToBeDeletedByClusterAutoscaler"

	// LabelNodeExcludeNodeDeprecated specifies that the node should be exclude from CCM
	LabelNodeExcludeNodeDeprecated = "service.beta.kubernetes.io/exclude-node"
	LabelNodeExcludeNode           = "service.alibabacloud.com/exclude-node"
)

func PatchM(mclient client.Client, target client.Object, getter func(runtime.Object) (client.Object, error), resource string,
) error {
	err := mclient.Get(
		context.TODO(),
		client.ObjectKey{
			Name:      target.GetName(),
			Namespace: target.GetNamespace(),
		}, target,
	)
	if err != nil {
		return fmt.Errorf("get origin object: %s", err.Error())
	}

	ntarget, err := getter(target.DeepCopyObject())
	if err != nil {
		return fmt.Errorf("get object diff patch: %s", err.Error())
	}
	oldData, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("ensure marshal: %s", err.Error())
	}
	newData, err := json.Marshal(ntarget)
	if err != nil {
		return fmt.Errorf("ensure marshal: %s", err.Error())
	}
	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, target)
	if patchErr != nil {
		return fmt.Errorf("create merge patch: %s", patchErr.Error())
	}

	if string(patchBytes) == "{}" {
		return nil
	}

	klog.Infof("try to patch %s/%s, %s ", target.GetNamespace(), target.GetName(), string(patchBytes))
	if resource == PatchSpec || resource == PatchAll {
		err := mclient.Patch(
			context.TODO(), ntarget,
			client.RawPatch(types.StrategicMergePatchType, patchBytes),
		)
		if err != nil {
			return fmt.Errorf("patch spec: %s", err.Error())
		}
	}

	if resource == PatchStatus || resource == PatchAll {
		return mclient.Status().Patch(
			context.TODO(), ntarget,
			client.RawPatch(types.StrategicMergePatchType, patchBytes),
		)
	}
	return nil
}

func FindCondition(conds []v1.NodeCondition, conditionType v1.NodeConditionType) (*v1.NodeCondition, bool) {
	var retCon *v1.NodeCondition
	for i := range conds {
		if conds[i].Type == conditionType {
			if retCon == nil || retCon.LastHeartbeatTime.Before(&conds[i].LastHeartbeatTime) {
				retCon = &conds[i]
			}
		}
	}

	if retCon == nil {
		return &v1.NodeCondition{}, false
	} else {
		return retCon, true
	}
}

// GetNodeCondition will get pointer to Node's existing condition.
// returns nil if no matching condition found.
func GetNodeCondition(node *v1.Node, conditionType v1.NodeConditionType) *v1.NodeCondition {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == conditionType {
			return &node.Status.Conditions[i]
		}
	}
	return nil
}

func HasExcludeLabel(node *v1.Node) bool {
	if _, exclude := node.Labels[LabelNodeExcludeNodeDeprecated]; exclude {
		return true
	}
	if _, exclude := node.Labels[LabelNodeExcludeNode]; exclude {
		return true
	}
	return false
}

func FindNodeByNodeName(nodes []v1.Node, nodeName string) *v1.Node {
	for _, n := range nodes {
		if n.Name == nodeName {
			return &n
		}
	}
	return nil
}

// providerID
// 1) the id of the instance in the alicloud API. Use '.' to separate providerID which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7'. The format of "REGION.NODEID"
// 2) the id for an instance in the kubernetes API, which has 'alicloud://' prefix. e.g. alicloud://cn-hangzhou.i-v98dklsmnxkkgiiil7
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

func IsMasterNode(node *v1.Node) bool {
	if _, isMaster := node.Labels[LabelNodeRoleMaster]; isMaster {
		return true
	}
	return false
}

func IsNodeExcludeFromLoadBalancer(node *v1.Node) bool {
	if _, exclude := node.Labels[LabelNodeExcludeBalancer]; exclude {
		return true
	}

	if _, exclude := node.Labels[LabelNodeExcludeBalancerDeprecated]; exclude {
		return true
	}

	if HasExcludeLabel(node) {
		return true
	}
	return false
}

func IsNodeExcludeFromEdgeLoadBalancer(node *v1.Node) bool {
	if _, exclude := node.Labels[LabelNodeExcludeBalancer]; exclude {
		return true
	}

	if _, exclude := node.Labels[LabelNodeExcludeBalancerDeprecated]; exclude {
		return true
	}
	return false
}

func GetNodeInternalIP(node *v1.Node) (string, error) {
	if len(node.Status.Addresses) == 0 {
		return "", fmt.Errorf("node %s do not contains addresses", node.Name)
	}
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address, nil
		}
	}
	return "", fmt.Errorf("node %s can not find InternalIP in node addresses", node.Name)
}

func NodeInfo(node *v1.Node) string {
	if node == nil {
		return ""
	}
	pNode := node.DeepCopy()
	pNode.ManagedFields = nil
	pNode.Status.Images = nil
	jsonByte, _ := json.Marshal(pNode)
	return string(jsonByte)
}
