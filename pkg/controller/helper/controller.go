package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog/v2"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func FindCondition(
	conds []corev1.NodeCondition,
	conditionType corev1.NodeConditionType,
) (*corev1.NodeCondition, bool) {
	var retCon *corev1.NodeCondition
	for i := range conds {
		if conds[i].Type == conditionType {
			if retCon == nil || retCon.LastHeartbeatTime.Before(&conds[i].LastHeartbeatTime) {
				retCon = &conds[i]
			}
		}
	}

	if retCon == nil {
		return &corev1.NodeCondition{}, false
	} else {
		return retCon, true
	}
}

func NewDelay(d int) reconcile.Result {
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: time.Duration(d) * time.Second,
	}
}

const (
	PatchAll    = "all"
	PatchSpec   = "spec"
	PatchStatus = "status"
)

func PatchM(
	mclient client.Client,
	target client.Object,
	getter func(runtime.Object) (client.Object, error),
	resource string,
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

const (
	// LabelNodeExcludeNodeDeprecated specifies that the node should be exclude from CCM
	LabelNodeExcludeNodeDeprecated = "service.beta.kubernetes.io/exclude-node"
	LabelNodeExcludeNode           = "service.alibabacloud.com/exclude-node"
)

func HasExcludeLabel(node *corev1.Node) bool {
	if _, exclude := node.Labels[LabelNodeExcludeNodeDeprecated]; exclude {
		return true
	}
	if _, exclude := node.Labels[LabelNodeExcludeNode]; exclude {
		return true
	}
	return false
}
