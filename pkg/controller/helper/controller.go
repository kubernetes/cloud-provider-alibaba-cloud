package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func FindCondition(
	conds []corev1.NodeCondition,
	typ corev1.NodeConditionType,
) (*corev1.NodeCondition, bool) {
	for i := range conds {
		if conds[i].Type == typ {
			return &conds[i], true
		}
	}
	// condition not found, do not trigger repair
	return &corev1.NodeCondition{}, false
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

// Patch used to patch an object. get the latest object,
// make a copy and do some modification.
// finally apply the diff.
// eg.
// 	otask := &acv1.Task{}
//	err := mclient.Get(
//		context.TODO(),
//		client.ObjectKey{
//			Name:      task.Name,
//			Namespace: task.Namespace,
//		}, otask,
//	)
//	if err != nil {
//		return fmt.Errorf("get task: %s", err.Error())
//	}
// diff := func() (o, n runtime.Object, e error) {
//		ntask := otask.DeepCopy()
//		ntask.Status.Hash = value
//		ntask.Status.Phase = phase
//		ntask.Status.Reason = reason
//		ntask.Status.LastOperateTime = metav1.Now()
//		return otask,ntask,nil
//	}
//	return tools.Patch(mclient, diff, tools.PatchStatus)
func Patch(mclient client.Client, getter func() (o, n client.Object, e error), resource string) error {
	otarget, ntarget, err := getter()
	if err != nil {
		return fmt.Errorf("get object diff patch: %s", err.Error())
	}
	oldData, err := json.Marshal(otarget)
	if err != nil {
		return fmt.Errorf("ensure marshal: %s", err.Error())
	}
	newData, err := json.Marshal(ntarget)
	if err != nil {
		return fmt.Errorf("ensure marshal: %s", err.Error())
	}
	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, otarget)
	if patchErr != nil {
		return fmt.Errorf("create merge patch: %s", patchErr.Error())
	}

	if string(patchBytes) == "{}" {
		return nil
	}

	if resource == PatchSpec || resource == PatchAll {
		err := mclient.Patch(
			context.TODO(), ntarget,
			client.RawPatch(types.MergePatchType, patchBytes),
		)
		if err != nil {
			return fmt.Errorf("patch spec: %s", err.Error())
		}
	}

	if resource == PatchStatus || resource == PatchAll {
		return mclient.Status().Patch(
			context.TODO(), ntarget,
			client.RawPatch(types.MergePatchType, patchBytes),
		)
	}
	return nil
}

// PatchM patch object
// diff := func(copy runtime.Object) (client.Object,error) {
//      nins := copy.(*v1.Node)
//		nins.Status.Hash = value
//		nins.Status.Phase = phase
//		nins.Status.Reason = reason
//		nins.Status.LastOperateTime = metav1.Now()
//		return nins,nil
//	}
//	return tools.PatchM(mclient,yourObject, diff, tools.PatchStatus)
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
	if resource == PatchSpec || resource == PatchAll {
		err := mclient.Patch(
			context.TODO(), ntarget,
			client.RawPatch(types.MergePatchType, patchBytes),
		)
		if err != nil {
			return fmt.Errorf("patch spec: %s", err.Error())
		}
	}

	if resource == PatchStatus || resource == PatchAll {
		return mclient.Status().Patch(
			context.TODO(), ntarget,
			client.RawPatch(types.MergePatchType, patchBytes),
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
