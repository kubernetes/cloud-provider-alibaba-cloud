package nlbv2

import (
	"context"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"testing"
	"time"
)

func TestEnqueueRequestForServiceEvent(t *testing.T) {
	h := NewEnqueueRequestForServiceEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	services := make(map[string]*v1.Service)
	ctx := context.TODO()
	kubeClient := getFakeKubeClient()
	svcList := &v1.ServiceList{}
	_ = kubeClient.List(context.TODO(), svcList)
	for i, svc := range svcList.Items {
		services[svc.Name] = &svcList.Items[i]
	}

	// create event
	h.Create(ctx, event.CreateEvent{Object: services[ServiceName]}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	h.Create(ctx, event.CreateEvent{Object: services["clb"]}, queue)
	assert.Equal(t, 0, queue.Len())

	h.Create(ctx, event.CreateEvent{Object: services["nodePort"]}, queue)
	assert.Equal(t, 0, queue.Len())

	// update event
	nlbNew := services[ServiceName].DeepCopy()
	nlbNew.Spec.Ports[0].Port = 81
	h.Update(ctx, event.UpdateEvent{ObjectOld: services[ServiceName], ObjectNew: nlbNew}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	clbNew := services["clb"].DeepCopy()
	clbNew.Spec.Ports[0].Port = 81
	h.Update(ctx, event.UpdateEvent{ObjectOld: services["clb"], ObjectNew: clbNew}, queue)
	assert.Equal(t, 0, queue.Len())

	nodePortNew := services["nodePort"].DeepCopy()
	nodePortNew.Spec.Ports[0].Port = 81
	h.Update(ctx, event.UpdateEvent{ObjectOld: services["nodePort"], ObjectNew: nodePortNew}, queue)
	assert.Equal(t, 0, queue.Len())

	h.Update(ctx, event.UpdateEvent{ObjectOld: services["nodePort"], ObjectNew: services[ServiceName]}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	h.Update(ctx, event.UpdateEvent{ObjectOld: services[ServiceName], ObjectNew: services["nodePort"]}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	nlbDel := services[ServiceName].DeepCopy()
	nlbDel.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	h.Update(ctx, event.UpdateEvent{ObjectOld: services[ServiceName], ObjectNew: nlbDel}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}
}

func TestNewEnqueueRequestForEndpointEvent(t *testing.T) {
	h := NewEnqueueRequestForEndpointEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	kubeClient := getFakeKubeClient()
	_ = h.InjectClient(kubeClient)

	ep := &v1.Endpoints{}
	_ = kubeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: v1.NamespaceDefault,
		Name:      ServiceName,
	}, ep)

	ctx := context.TODO()

	h.Create(ctx, event.CreateEvent{Object: ep}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	epNew := ep.DeepCopy()
	epNew.Subsets = []v1.EndpointSubset{
		{
			Addresses: []v1.EndpointAddress{
				{
					IP:       "10.96.0.15",
					NodeName: tea.String(NodeName),
				},
			},
			Ports: []v1.EndpointPort{
				{
					Name:     "tcp",
					Port:     80,
					Protocol: "TCP",
				},
			},
		},
	}
	h.Update(ctx, event.UpdateEvent{ObjectOld: ep, ObjectNew: epNew}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	h.Delete(ctx, event.DeleteEvent{Object: ep}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}
}

func TestNewEnqueueRequestForEndpointSliceEvent(t *testing.T) {
	h := NewEnqueueRequestForEndpointSliceEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	kubeClient := getFakeKubeClient()
	_ = h.InjectClient(kubeClient)

	es := &discovery.EndpointSlice{}
	_ = kubeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: v1.NamespaceDefault,
		Name:      ServiceName,
	}, es)

	ctx := context.TODO()

	h.Create(ctx, event.CreateEvent{Object: es}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	esNew := es.DeepCopy()
	esNew.Endpoints[0].Addresses = []string{"10.96.0.16"}

	h.Update(ctx, event.UpdateEvent{ObjectOld: es, ObjectNew: esNew}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	h.Delete(ctx, event.DeleteEvent{Object: es}, queue)
	assert.Equal(t, 1, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}
}

func TestNewEnqueueRequestForNodeEvent(t *testing.T) {
	h := NewEnqueueRequestForNodeEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	kubeClient := getFakeKubeClient()
	_ = h.InjectClient(kubeClient)

	node := &v1.Node{}
	_ = kubeClient.Get(context.TODO(), types.NamespacedName{
		Name: NodeName,
	}, node)

	ctx := context.TODO()
	h.Create(ctx, event.CreateEvent{Object: node}, queue)
	assert.Equal(t, 2, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	h.Delete(ctx, event.DeleteEvent{Object: node}, queue)
	assert.Equal(t, 2, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}

	newN := node.DeepCopy()
	newN.Spec.Unschedulable = true
	h.Update(ctx, event.UpdateEvent{ObjectOld: node, ObjectNew: newN}, queue)
	assert.Equal(t, 2, queue.Len())
	for queue.Len() > 0 {
		item, _ := queue.Get()
		queue.Done(item)
	}
}

func Test_canNodeSkipEventHandler(t *testing.T) {
	kubeClient := getFakeKubeClient()
	node := &v1.Node{}
	_ = kubeClient.Get(context.TODO(), types.NamespacedName{
		Name: NodeName,
	}, node)
	node.Labels = map[string]string{
		helper.LabelNodeExcludeNode: "true",
	}

	assert.Equal(t, canNodeSkipEventHandler(node), true)
}

func Test_isEndpointProcessNeeded(t *testing.T) {
	kubeClient := getFakeKubeClient()
	ep := &v1.Endpoints{}
	_ = kubeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: v1.NamespaceDefault,
		Name:      ServiceName,
	}, ep)
	assert.Equal(t, isEndpointProcessNeeded(ep, kubeClient), true)
}

func Test_isEndpointSliceProcessNeeded(t *testing.T) {
	kubeClient := getFakeKubeClient()
	es := &discovery.EndpointSlice{}
	_ = kubeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: v1.NamespaceDefault,
		Name:      ServiceName,
	}, es)
	assert.Equal(t, isEndpointSliceProcessNeeded(es, kubeClient), true)
}

func Test_nodeConditionChanged(t *testing.T) {
	cl := getFakeKubeClient()

	oldN := &v1.Node{}
	_ = cl.Get(context.TODO(), types.NamespacedName{
		Name: NodeName,
	}, oldN)

	newN := oldN.DeepCopy()
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
	assert.Equal(t, nodeSpecChanged(oldN, newN), false)

	newN = oldN.DeepCopy()
	newN.Status.Conditions = []v1.NodeCondition{
		{
			Reason: "KubeletReady",
			Status: "False",
			Type:   v1.NodeReady,
		},
	}
	assert.Equal(t, nodeReadyChanged(oldN, newN), true)
}

func Test_nodeLabelsChanged(t *testing.T) {
	cl := getFakeKubeClient()

	oldN := &v1.Node{}
	_ = cl.Get(context.TODO(), types.NamespacedName{
		Name: NodeName,
	}, oldN)

	newN := oldN.DeepCopy()
	newN.Labels["ack.aliyun.com"] = "new"
	assert.Equal(t, nodeLabelsChanged(NodeName, oldN.Labels, newN.Labels), true)
}

func Test_nodeSpecChanged(t *testing.T) {
	cl := getFakeKubeClient()

	oldN := &v1.Node{}
	_ = cl.Get(context.TODO(), types.NamespacedName{
		Name: NodeName,
	}, oldN)

	newN := oldN.DeepCopy()
	assert.Equal(t, nodeSpecChanged(oldN, newN), false)

	newN.Spec.Unschedulable = true
	assert.Equal(t, nodeSpecChanged(oldN, newN), true)
}
