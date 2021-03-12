package tools

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
) (corev1.NodeCondition, bool) {
	for i := range conds {
		if conds[i].Type == typ {
			return conds[i], true
		}
	}
	// condition not found, do not trigger repair
	return corev1.NodeCondition{}, false
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

func Patch(rclient client.Client, getter func() (o, n runtime.Object, e error), resource string) error {

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

	if resource == PatchSpec || resource == PatchAll {
		err := rclient.Patch(
			context.TODO(), ntarget,
			client.RawPatch(types.MergePatchType, patchBytes),
		)
		if err != nil {
			return fmt.Errorf("patch spec: %s", err.Error())
		}
	}

	if resource == PatchStatus || resource == PatchAll {
		return rclient.Status().Patch(
			context.TODO(), ntarget,
			client.RawPatch(types.MergePatchType, patchBytes),
		)
	}
	return nil
}
