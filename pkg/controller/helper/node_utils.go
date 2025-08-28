package helper

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
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

	ProviderIdPrefixHybridNode = "ack-hybrid://"
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

func PatchNodeStatus(mclient client.Client, target *v1.Node, getter func(*v1.Node) (*v1.Node, error)) error {
	err := mclient.Get(
		context.TODO(),
		client.ObjectKey{
			Name: target.GetName(),
		}, target,
	)
	if err != nil {
		return fmt.Errorf("get origin object: %s", err.Error())
	}

	ntarget, err := getter(target.DeepCopy())
	if err != nil {
		return fmt.Errorf("get object diff patch: %s", err.Error())
	}

	diffTarget := ntarget
	manuallyPatchAddresses := (len(target.Status.Addresses) > 0) &&
		!equality.Semantic.DeepEqual(target.Status.Addresses, ntarget.Status.Addresses)
	if manuallyPatchAddresses {
		diffTarget = diffTarget.DeepCopy()
		diffTarget.Status.Addresses = target.Status.Addresses
	}

	oldData, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("ensure marshal: %s", err.Error())
	}
	newData, err := json.Marshal(diffTarget)
	if err != nil {
		return fmt.Errorf("ensure marshal: %s", err.Error())
	}
	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, target)
	if patchErr != nil {
		return fmt.Errorf("create merge patch: %s", patchErr.Error())
	}

	if string(patchBytes) == "{}" && !manuallyPatchAddresses {
		return nil
	}

	if manuallyPatchAddresses {
		patchBytes, err = fixupPatchForNodeStatusAddresses(patchBytes, ntarget.Status.Addresses)
		if err != nil {
			return fmt.Errorf("fixup patch for node status addresses: %s", err.Error())
		}
	}

	klog.Infof("try to patch node status %s/%s, %s ", target.GetNamespace(), target.GetName(), string(patchBytes))

	return mclient.Status().Patch(
		context.TODO(), ntarget,
		client.RawPatch(types.StrategicMergePatchType, patchBytes))
}

// fixupPatchForNodeStatusAddresses adds a replace-strategy patch for Status.Addresses to
// the existing patch
func fixupPatchForNodeStatusAddresses(patchBytes []byte, addresses []v1.NodeAddress) ([]byte, error) {
	// Given patchBytes='{"status": {"conditions": [ ... ], "phase": ...}}' and
	// addresses=[{"type": "InternalIP", "address": "10.0.0.1"}], we need to generate:
	//
	//   {
	//     "status": {
	//       "conditions": [ ... ],
	//       "phase": ...,
	//       "addresses": [
	//         {
	//           "type": "InternalIP",
	//           "address": "10.0.0.1"
	//         },
	//         {
	//           "$patch": "replace"
	//         }
	//       ]
	//     }
	//   }

	var patchMap map[string]interface{}
	if err := json.Unmarshal(patchBytes, &patchMap); err != nil {
		return nil, err
	}

	addrBytes, err := json.Marshal(addresses)
	if err != nil {
		return nil, err
	}
	var addrArray []interface{}
	if err := json.Unmarshal(addrBytes, &addrArray); err != nil {
		return nil, err
	}
	addrArray = append(addrArray, map[string]interface{}{"$patch": "replace"})

	status := patchMap["status"]
	if status == nil {
		status = map[string]interface{}{}
		patchMap["status"] = status
	}
	statusMap, ok := status.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data in patch")
	}
	statusMap["addresses"] = addrArray

	return json.Marshal(patchMap)
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

func IsExcludedNode(node *v1.Node) bool {
	if node == nil {
		return false
	}
	if _, exclude := node.Labels[LabelNodeExcludeNodeDeprecated]; exclude {
		return true
	}
	if _, exclude := node.Labels[LabelNodeExcludeNode]; exclude {
		return true
	}
	if strings.HasPrefix(node.Spec.ProviderID, ProviderIdPrefixHybridNode) {
		klog.V(5).Infof("node %s is hybrid node type which should be excluded", node.Name)
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

	if IsExcludedNode(node) {
		return true
	}
	return false
}

func GetNodeInternalIP(node *v1.Node) (string, error) {
	if len(node.Status.Addresses) == 0 {
		return "", fmt.Errorf("node %s do not contains addresses", node.Name)
	}
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP && IsIPv4(addr.Address) {
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
