package clbv1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	helper "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestEnqueueRequestForServiceEvent(t *testing.T) {
	h := NewEnqueueRequestForServiceEvent(record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	svc := getDefaultService()
	svc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s1.small"
	ctx := context.TODO()
	// create event
	h.Create(ctx, event.CreateEvent{Object: svc}, queue)
	// delete event
	h.Delete(ctx, event.DeleteEvent{Object: svc}, queue)
	// update event
	newSvc := svc.DeepCopy()
	newSvc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s2.small"
	h.Update(ctx, event.UpdateEvent{ObjectOld: svc, ObjectNew: newSvc}, queue)

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

	tunnelSvc := svc.DeepCopy()
	tunnelSvc.Spec.Type = v1.ServiceTypeClusterIP
	tunnelSvc.Annotations = map[string]string{helper.TunnelType: "tunnel"}
	assert.Equal(t, needAdd(tunnelSvc), false)
}

func TestEnqueueRequestForEndpointEvent(t *testing.T) {
	cl := getFakeKubeClient()
	h := NewEnqueueRequestForEndpointEvent(cl, record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	ep := &v1.Endpoints{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Namespace: NS,
		Name:      SvcName,
	}, ep); err != nil {
		t.Error(err)
	}
	ctx := context.TODO()

	h.Create(ctx, event.CreateEvent{Object: ep}, queue)
	h.Delete(ctx, event.DeleteEvent{Object: ep}, queue)
	h.Update(ctx, event.UpdateEvent{ObjectOld: ep, ObjectNew: ep}, queue)
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
	assert.True(t, isEndpointProcessNeeded(ep, cl))

	assert.False(t, isEndpointProcessNeeded(nil, cl))

	epLeader := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NS,
			Name:      SvcName,
			Annotations: map[string]string{
				resourcelock.LeaderElectionRecordAnnotationKey: "{}",
			},
		},
	}
	assert.False(t, isEndpointProcessNeeded(epLeader, cl))

	epNoSvc := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: NS, Name: "no-such-svc"},
	}
	clNoSvc := fake.NewClientBuilder().WithRuntimeObjects(epNoSvc).Build()
	assert.False(t, isEndpointProcessNeeded(epNoSvc, clNoSvc))

	clusterIPSvc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "clusterip-svc", Namespace: NS},
		Spec:       v1.ServiceSpec{Type: v1.ServiceTypeClusterIP},
	}
	epClusterIP := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Name: "clusterip-svc", Namespace: NS},
	}
	clClusterIP := fake.NewClientBuilder().WithRuntimeObjects(clusterIPSvc, epClusterIP).Build()
	assert.False(t, isEndpointProcessNeeded(epClusterIP, clClusterIP))
}

func TestEnqueueRequestForNodeEvent(t *testing.T) {
	cl := getFakeKubeClient()
	h := NewEnqueueRequestForNodeEvent(cl, record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	node := &v1.Node{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Name: NodeName,
	}, node); err != nil {
		t.Error(err)
	}
	ctx := context.TODO()

	h.Create(ctx, event.CreateEvent{Object: node}, queue)
	h.Delete(ctx, event.DeleteEvent{Object: node}, queue)

	newN := node.DeepCopy()
	newN.Spec.Unschedulable = true
	h.Update(ctx, event.UpdateEvent{ObjectOld: node, ObjectNew: newN}, queue)
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
			Status: v1.ConditionTrue,
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
			Status: v1.ConditionFalse,
			Type:   v1.NodeReady,
		},
	}
	assert.Equal(t, nodeSpecChanged(oldN, newN), true)

	newN = oldN.DeepCopy()
	newN.Labels["ack.aliyun.com"] = "new"
	assert.Equal(t, nodeSpecChanged(oldN, newN), true)
}

func TestEnqueueRequestForEndpointSliceEvent(t *testing.T) {
	cl := getFakeKubeClient()
	h := NewEnqueueRequestForEndpointSliceEvent(cl, record.NewFakeRecorder(100))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	es := &discovery.EndpointSlice{}
	if err := cl.Get(context.TODO(), types.NamespacedName{
		Name:      SvcName,
		Namespace: NS,
	}, es); err != nil {
		t.Error(err)
	}
	ctx := context.TODO()

	h.Create(ctx, event.CreateEvent{Object: es}, queue)
	h.Delete(ctx, event.DeleteEvent{Object: es}, queue)
	h.Update(ctx, event.UpdateEvent{ObjectOld: es, ObjectNew: es}, queue)
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
	assert.True(t, isEndpointSliceProcessNeeded(es, cl))

	assert.False(t, isEndpointSliceProcessNeeded(nil, cl))

	esNoLabel := &discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "no-label", Namespace: NS, Labels: map[string]string{}},
	}
	assert.False(t, isEndpointSliceProcessNeeded(esNoLabel, cl))

	esNoSvc := &discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "no-such-svc-es",
			Namespace: NS,
			Labels:    map[string]string{discovery.LabelServiceName: "no-such-svc"},
		},
	}
	clNoSvc := fake.NewClientBuilder().WithRuntimeObjects(esNoSvc).Build()
	assert.False(t, isEndpointSliceProcessNeeded(esNoSvc, clNoSvc))

	clusterIPSvc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "clusterip-svc", Namespace: NS},
		Spec:       v1.ServiceSpec{Type: v1.ServiceTypeClusterIP},
	}
	esClusterIP := &discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clusterip-svc-es",
			Namespace: NS,
			Labels:    map[string]string{discovery.LabelServiceName: "clusterip-svc"},
		},
	}
	clClusterIP := fake.NewClientBuilder().WithRuntimeObjects(clusterIPSvc, esClusterIP).Build()
	assert.False(t, isEndpointSliceProcessNeeded(esClusterIP, clClusterIP))
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

func TestCheckServiceAffected(t *testing.T) {
	cl := getFakeKubeClient()
	h := NewEnqueueRequestForNodeEvent(cl, record.NewFakeRecorder(100))
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: NodeName,
		},
	}

	t.Run("ENI backend type", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[helper.BackendType] = "eni"
		result := h.checkServiceAffected(node, svc)
		assert.False(t, result)
	})

	t.Run("feature gate disabled", func(t *testing.T) {
		defaultEnabled := feature.DefaultMutableFeatureGate.Enabled(config.FilterServiceOnNodeChange)
		err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
			string(config.FilterServiceOnNodeChange): false,
		})
		assert.NoError(t, err)
		defer func() {
			err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
				string(config.FilterServiceOnNodeChange): defaultEnabled,
			})
			assert.NoError(t, err)
		}()

		svc := getDefaultService()
		result := h.checkServiceAffected(node, svc)
		assert.True(t, result)
	})

	t.Run("cluster external traffic policy", func(t *testing.T) {
		defaultEnabled := feature.DefaultMutableFeatureGate.Enabled(config.FilterServiceOnNodeChange)
		err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
			string(config.FilterServiceOnNodeChange): true,
		})
		assert.NoError(t, err)
		defer func() {
			err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
				string(config.FilterServiceOnNodeChange): defaultEnabled,
			})
			assert.NoError(t, err)
		}()

		svc := getDefaultService()
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
		result := h.checkServiceAffected(node, svc)
		assert.True(t, result)
	})

	t.Run("with backend label annotation", func(t *testing.T) {
		defaultEnabled := feature.DefaultMutableFeatureGate.Enabled(config.FilterServiceOnNodeChange)
		err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
			string(config.FilterServiceOnNodeChange): true,
		})
		assert.NoError(t, err)
		defer func() {
			err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
				string(config.FilterServiceOnNodeChange): defaultEnabled,
			})
			assert.NoError(t, err)
		}()

		svc := getDefaultService()
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
		svc.Annotations[annotation.Annotation(annotation.BackendLabel)] = "app=nginx"
		result := h.checkServiceAffected(node, svc)
		assert.True(t, result)
	})

	t.Run("with remove unscheduled annotation", func(t *testing.T) {
		defaultEnabled := feature.DefaultMutableFeatureGate.Enabled(config.FilterServiceOnNodeChange)
		err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
			string(config.FilterServiceOnNodeChange): true,
		})
		assert.NoError(t, err)
		defer func() {
			err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
				string(config.FilterServiceOnNodeChange): defaultEnabled,
			})
			assert.NoError(t, err)
		}()

		svc := getDefaultService()
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
		svc.Annotations[annotation.Annotation(annotation.RemoveUnscheduled)] = "true"
		result := h.checkServiceAffected(node, svc)
		assert.True(t, result)
	})

	t.Run("not affected", func(t *testing.T) {
		defaultEnabled := feature.DefaultMutableFeatureGate.Enabled(config.FilterServiceOnNodeChange)
		err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
			string(config.FilterServiceOnNodeChange): true,
		})
		assert.NoError(t, err)
		defer func() {
			err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
				string(config.FilterServiceOnNodeChange): defaultEnabled,
			})
			assert.NoError(t, err)
		}()

		svc := getDefaultService()
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
		result := h.checkServiceAffected(node, svc)
		assert.False(t, result)
	})
}

func TestCanNodeSkipEventHandler(t *testing.T) {
	assert.False(t, canNodeSkipEventHandler(nil))
	nodeNoLabels := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n"}}
	nodeNoLabels.Labels = nil
	assert.False(t, canNodeSkipEventHandler(nodeNoLabels))
	assert.False(t, canNodeSkipEventHandler(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n", Labels: map[string]string{}}}))

	nodeExcluded := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "excluded",
			Labels: map[string]string{helper.LabelNodeExcludeNode: "true"},
		},
	}
	assert.True(t, canNodeSkipEventHandler(nodeExcluded))

	nodeMaster := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "master",
			Labels: map[string]string{
				"node-role.kubernetes.io/master": "",
			},
		},
	}
	assert.True(t, canNodeSkipEventHandler(nodeMaster))
}
