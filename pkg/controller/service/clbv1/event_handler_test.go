package clbv1

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	helper "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"testing"
	"time"
)

func TestEnqueueRequestForServiceEvent(t *testing.T) {
	h := NewEnqueueRequestForServiceEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	svc := getDefaultService()
	svc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s1.small"
	// create event
	h.Create(event.CreateEvent{Object: svc}, queue)
	// delete event
	h.Delete(event.DeleteEvent{Object: svc}, queue)
	// update event
	newSvc := svc.DeepCopy()
	newSvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s2.small"
	h.Update(event.UpdateEvent{ObjectOld: svc, ObjectNew: newSvc}, queue)

}

func TestNeedUpdate(t *testing.T) {
	fakeRecord := record.NewFakeRecorder(100)

	oldSvc := getDefaultService()
	oldSvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s1.small"

	newSvc := oldSvc.DeepCopy()
	newSvc.Spec.Type = v1.ServiceTypeClusterIP
	assert.Equal(t, needUpdate(oldSvc, newSvc, fakeRecord), true)

	newSvc = oldSvc.DeepCopy()
	newSvc.ObjectMeta.UID = types.UID(SvcUID)
	assert.Equal(t, needUpdate(oldSvc, newSvc, fakeRecord), true)

	newSvc = oldSvc.DeepCopy()
	oldSvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s2.small"
	assert.Equal(t, needUpdate(oldSvc, newSvc, fakeRecord), true)

	newSvc = oldSvc.DeepCopy()
	newSvc.Spec.Ports = []v1.ServicePort{
		{
			Protocol:   "HTTPS",
			Port:       443,
			TargetPort: intstr.FromInt(443),
		},
	}
	assert.Equal(t, needUpdate(oldSvc, newSvc, fakeRecord), true)

	newSvc = oldSvc.DeepCopy()
	newSvc.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	assert.Equal(t, needUpdate(oldSvc, newSvc, fakeRecord), true)
}

func TestNeedAdd(t *testing.T) {
	svc := getDefaultService()
	assert.Equal(t, needAdd(svc), true)

	newSvc := svc.DeepCopy()
	newSvc.Finalizers = []string{helper.ServiceFinalizer}
	newSvc.Spec.Type = v1.ServiceTypeClusterIP
	assert.Equal(t, needAdd(newSvc), true)

	clusterIPSvc := svc.DeepCopy()
	clusterIPSvc.Spec.Type = v1.ServiceTypeClusterIP
	assert.Equal(t, needAdd(clusterIPSvc), false)
}

func TestEnqueueRequestForEndpointEvent(t *testing.T) {
	h := NewEnqueueRequestForEndpointEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	cl := getFakeKubeClient()
	_ = h.InjectClient(cl)

	ep := &v1.Endpoints{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Namespace: NS,
		Name:      SvcName,
	}, ep); err != nil {
		t.Error(err)
	}

	h.Create(event.CreateEvent{Object: ep}, queue)
	h.Delete(event.DeleteEvent{Object: ep}, queue)
	h.Update(event.UpdateEvent{ObjectOld: ep, ObjectNew: ep}, queue)
}

func TestIsEndpointProcessNeeded(t *testing.T) {
	cl := getFakeKubeClient()
	ep := &v1.Endpoints{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Namespace: NS,
		Name:      SvcName,
	}, ep); err != nil {
		t.Error(err)
	}
	assert.Equal(t, isEndpointProcessNeeded(ep, cl), true)
}

func TestEnqueueRequestForNodeEvent(t *testing.T) {
	h := NewEnqueueRequestForNodeEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	cl := getFakeKubeClient()
	_ = h.InjectClient(cl)

	node := &v1.Node{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Name: NodeName,
	}, node); err != nil {
		t.Error(err)
	}

	h.Create(event.CreateEvent{Object: node}, queue)
	h.Delete(event.DeleteEvent{Object: node}, queue)

	newN := node.DeepCopy()
	newN.Spec.Unschedulable = true
	h.Update(event.UpdateEvent{ObjectOld: node, ObjectNew: newN}, queue)
}

func TestNodeSpecChanged(t *testing.T) {
	cl := getFakeKubeClient()

	oldN := &v1.Node{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Name: NodeName,
	}, oldN); err != nil {
		t.Error(err)
	}

	newN := oldN.DeepCopy()
	assert.Equal(t, nodeSpecChanged(oldN, newN), false)

	newN.Spec.Unschedulable = true
	assert.Equal(t, nodeSpecChanged(oldN, newN), true)

	newN = oldN.DeepCopy()
	newN.Status.Conditions = []v1.NodeCondition{
		{
			Reason: "KubeletReady",
			Status: "True",
			Type:   v1.NodeReady,
		},
		{
			Reason: "KubeletHasSufficientPID",
			Status: "False",
			Type:   v1.NodePIDPressure,
		},
	}
	assert.Equal(t, nodeSpecChanged(oldN, newN), true)

	newN = oldN.DeepCopy()
	newN.Status.Conditions = []v1.NodeCondition{
		{
			Reason: "KubeletReady",
			Status: "False",
			Type:   v1.NodeReady,
		},
	}
	assert.Equal(t, nodeSpecChanged(oldN, newN), true)

	newN = oldN.DeepCopy()
	newN.Labels["ack.aliyun.com"] = "new"
	assert.Equal(t, nodeSpecChanged(oldN, newN), true)
}

func TestEnqueueRequestForEndpointSliceEvent(t *testing.T) {
	h := NewEnqueueRequestForEndpointSliceEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	cl := getFakeKubeClient()
	_ = h.InjectClient(cl)

	es := &discovery.EndpointSlice{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Name:      SvcName,
		Namespace: NS,
	}, es); err != nil {
		t.Error(err)
	}

	h.Create(event.CreateEvent{Object: es}, queue)
	h.Delete(event.DeleteEvent{Object: es}, queue)
	h.Update(event.UpdateEvent{ObjectOld: es, ObjectNew: es}, queue)
}

func TestIsEndpointSliceProcessNeeded(t *testing.T) {
	cl := getFakeKubeClient()
	es := &discovery.EndpointSlice{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Name:      SvcName,
		Namespace: NS,
	}, es); err != nil {
		t.Error(err)
	}
	assert.Equal(t, isEndpointSliceProcessNeeded(es, cl), true)
}

func TestIsEndpointSliceUpdateNeeded(t *testing.T) {
	cl := getFakeKubeClient()
	oldES := &discovery.EndpointSlice{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Name:      SvcName,
		Namespace: NS,
	}, oldES); err != nil {
		t.Error(err)
	}

	var newPort int32 = 81
	newES := oldES.DeepCopy()
	newES.Ports[0].Port = &newPort
	assert.Equal(t, isEndpointSliceUpdateNeeded(oldES, newES), true)
}
