package clbv1

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"testing"
	"time"
)

func TestCreateLB(t *testing.T) {
	builder := &ModelBuilder{
		LoadBalancerMgr: NewLoadBalancerManager(getMockCloudProvider()),
		ListenerMgr:     NewListenerManager(getMockCloudProvider()),
		VGroupMgr:       getTestVGroupManager(),
	}
	applier := NewModelApplier(NewLoadBalancerManager(getMockCloudProvider()),
		NewListenerManager(getMockCloudProvider()), getTestVGroupManager())

	// create new lb
	svc := getDefaultService()
	svc.Spec.Ports = []v1.ServicePort{
		{
			Name:       "tcp",
			Port:       80,
			TargetPort: intstr.FromInt(80),
			NodePort:   80,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "udp",
			Port:       53,
			TargetPort: intstr.FromInt(53),
			NodePort:   53,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "http",
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
			NodePort:   8080,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "https",
			Port:       443,
			TargetPort: intstr.FromInt(443),
			NodePort:   443,
			Protocol:   v1.ProtocolTCP,
		},
	}
	svc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "tcp:80,udp:53,http:8080,https:443"

	reqCtx := getReqCtx(svc)
	localModel, err := builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}
}

func TestSyncLB(t *testing.T) {
	builder := &ModelBuilder{
		LoadBalancerMgr: NewLoadBalancerManager(getMockCloudProvider()),
		ListenerMgr:     NewListenerManager(getMockCloudProvider()),
		VGroupMgr:       getTestVGroupManager(),
	}

	applier := NewModelApplier(NewLoadBalancerManager(getMockCloudProvider()),
		NewListenerManager(getMockCloudProvider()), getTestVGroupManager())

	svc := getDefaultService()
	svc.Spec.Ports = []v1.ServicePort{
		{
			Name:       "tcp",
			Port:       80,
			TargetPort: intstr.FromInt(80),
			NodePort:   80,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "udp",
			Port:       53,
			TargetPort: intstr.FromInt(53),
			NodePort:   53,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "http",
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
			NodePort:   8080,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "https",
			Port:       443,
			TargetPort: intstr.FromInt(443),
			NodePort:   443,
			Protocol:   v1.ProtocolTCP,
		},
	}
	svc.Annotations[annotation.Annotation(annotation.LoadBalancerId)] = vmock.ExistLBID
	svc.Annotations[annotation.Annotation(annotation.OverrideListener)] = "true"
	svc.Annotations[annotation.Annotation(annotation.Spec)] = "slb.s2.small"
	svc.Annotations[annotation.Annotation(annotation.ChargeType)] = "paybybandwidth"
	svc.Annotations[annotation.Annotation(annotation.Bandwidth)] = "5"
	svc.Annotations[annotation.Annotation(annotation.DeleteProtection)] = "off"
	svc.Annotations[annotation.Annotation(annotation.LoadBalancerName)] = "new-lb-name"
	svc.Annotations[annotation.Annotation(annotation.ModificationProtection)] = "NonProtection"

	svc.Annotations[annotation.Annotation(annotation.AclID)] = "acl-id"
	svc.Annotations[annotation.Annotation(annotation.AclStatus)] = string(model.OnFlag)
	svc.Annotations[annotation.Annotation(annotation.AclType)] = "white"
	svc.Annotations[annotation.Annotation(annotation.Scheduler)] = "wrr"
	svc.Annotations[annotation.Annotation(annotation.PersistenceTimeout)] = "10"
	svc.Annotations[annotation.Annotation(annotation.EstablishedTimeout)] = "12"
	svc.Annotations[annotation.Annotation(annotation.EnableHttp2)] = "false"
	svc.Annotations[annotation.Annotation(annotation.IdleTimeout)] = "60"
	svc.Annotations[annotation.Annotation(annotation.RequestTimeout)] = "30"
	svc.Annotations[annotation.Annotation(annotation.ConnectionDrain)] = "on"
	svc.Annotations[annotation.Annotation(annotation.ConnectionDrainTimeout)] = "30"
	svc.Annotations[annotation.Annotation(annotation.Cookie)] = "test-cookie"
	svc.Annotations[annotation.Annotation(annotation.CookieTimeout)] = "60"
	svc.Annotations[annotation.Annotation(annotation.SessionStick)] = "on"
	svc.Annotations[annotation.Annotation(annotation.SessionStickType)] = "insert"
	svc.Annotations[annotation.Annotation(annotation.XForwardedForProto)] = "on"

	svc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "tcp:80,udp:53,http:8080,https:443"
	svc.Annotations[annotation.Annotation(annotation.CertID)] = "cert-id"
	svc.Annotations[annotation.Annotation(annotation.ForwardPort)] = "8080:443"

	svc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "6"
	svc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "5"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckConnectTimeout)] = "3"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckConnectPort)] = "80"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "5"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckDomain)] = "foo2.bar.com"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckURI)] = "/test2/index.html"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckHTTPCode)] = "http_2xx,http_3xx"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = "tcp"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = "on"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckTimeout)] = "3"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckMethod)] = "get"

	reqCtx := getReqCtx(svc)
	localModel, err := builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteLB(t *testing.T) {
	builder := &ModelBuilder{
		LoadBalancerMgr: NewLoadBalancerManager(getMockCloudProvider()),
		ListenerMgr:     NewListenerManager(getMockCloudProvider()),
		VGroupMgr:       getTestVGroupManager(),
	}

	applier := NewModelApplier(NewLoadBalancerManager(getMockCloudProvider()),
		NewListenerManager(getMockCloudProvider()), getTestVGroupManager())

	// delete auto-created lb
	svc := getDefaultService()
	svc.UID = types.UID(SvcUID)
	svc.ObjectMeta.Finalizers = []string{helper.ServiceFinalizer}
	svc.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	svc.Labels = map[string]string{helper.LabelServiceHash: helper.GetServiceHash(svc)}
	reqCtx := getReqCtx(svc)
	localModel, err := builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}

	// reuse slb
	reqCtx = getReqCtx(getDefaultService())
	reqCtx.Service.Annotations[annotation.Annotation(annotation.LoadBalancerId)] = "exist-lb"
	reqCtx.Service.Annotations[annotation.Annotation(annotation.OverrideListener)] = "true"
	reqCtx.Service.ObjectMeta.Finalizers = []string{"service.k8s.alibaba/resources"}
	reqCtx.Service.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: time.Now()}

	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}
}

func TestBuildRemoteModel(t *testing.T) {
	builder := &ModelBuilder{
		LoadBalancerMgr: NewLoadBalancerManager(getMockCloudProvider()),
		ListenerMgr:     NewListenerManager(getMockCloudProvider()),
		VGroupMgr:       getTestVGroupManager(),
	}
	svc := getDefaultService()
	svc.UID = types.UID(SvcUID)
	_, err := builder.Instance(RemoteModel).Build(getReqCtx(svc))
	if err != nil {
		t.Error(err)
	}
}

func TestSyncVGroups(t *testing.T) {
	builder := &ModelBuilder{
		LoadBalancerMgr: NewLoadBalancerManager(getMockCloudProvider()),
		ListenerMgr:     NewListenerManager(getMockCloudProvider()),
		VGroupMgr:       getTestVGroupManager(),
	}

	applier := NewModelApplier(NewLoadBalancerManager(getMockCloudProvider()),
		NewListenerManager(getMockCloudProvider()), getTestVGroupManager())

	svc := getDefaultService()
	svc.Spec.Ports = []v1.ServicePort{
		{
			Name:       "tcp",
			Port:       80,
			TargetPort: intstr.FromInt(80),
			NodePort:   80,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "udp",
			Port:       53,
			TargetPort: intstr.FromInt(53),
			NodePort:   53,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "http",
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
			NodePort:   8080,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "https",
			Port:       443,
			TargetPort: intstr.FromInt(443),
			NodePort:   443,
			Protocol:   v1.ProtocolTCP,
		},
	}

	// cluster mode
	reqCtx := getReqCtx(svc.DeepCopy())
	reqCtx.Service.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyType(v1.ServiceInternalTrafficPolicyCluster)
	reqCtx.Service.UID = types.UID(SvcUID)
	reqCtx.Service.Annotations = nil
	localModel, err := builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}

	//local mode
	reqCtx = getReqCtx(svc.DeepCopy())
	reqCtx.Service.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyType(v1.ServiceInternalTrafficPolicyLocal)
	reqCtx.Service.UID = types.UID(SvcUID)
	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}

	// eni mode
	reqCtx = getReqCtx(svc.DeepCopy())
	reqCtx.Service.Annotations = map[string]string{annotation.BackendType: model.ENIBackendType}
	reqCtx.Service.UID = types.UID(SvcUID)
	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}

	// reuse vServerGroup: cluster traffic policy
	reqCtx = getReqCtx(getDefaultService())
	reqCtx.Service.Annotations[annotation.Annotation(annotation.LoadBalancerId)] = vmock.ExistLBID
	reqCtx.Service.Annotations[annotation.Annotation(annotation.VGroupPort)] = fmt.Sprintf("%s:%d", vmock.ExistVGroupID, 80)
	reqCtx.Service.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "80"
	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}

	// reuse vServerGroup: local traffic policy
	reqCtx = getReqCtx(getDefaultService())
	reqCtx.Service.Annotations[annotation.Annotation(annotation.LoadBalancerId)] = vmock.ExistLBID
	reqCtx.Service.Annotations[annotation.Annotation(annotation.VGroupPort)] = fmt.Sprintf("%s:%d", vmock.ExistVGroupID, 80)
	reqCtx.Service.Annotations[annotation.Annotation(annotation.VGroupWeight)] = "0"
	reqCtx.Service.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}

	// EndpointSlice
	_ = utilfeature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{string(ctrlCfg.EndpointSlice): true})
	reqCtx = getReqCtx(getDefaultService())
	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}
	_ = utilfeature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{string(ctrlCfg.EndpointSlice): false})

	// filter out by label
	reqCtx = getReqCtx(getDefaultService())
	reqCtx.Service.Annotations[annotation.Annotation(annotation.BackendLabel)] = "app=nginx"
	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}

	// string targetPort
	reqCtx = getReqCtx(getDefaultService())
	reqCtx.Service.Spec.Ports = []v1.ServicePort{
		{
			Name:       "tcp",
			Port:       80,
			TargetPort: intstr.FromString("tcp"),
			NodePort:   80,
			Protocol:   v1.ProtocolTCP,
		},
	}
	reqCtx.Service.Annotations[annotation.BackendType] = model.ENIBackendType
	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, localModel.VServerGroups[0].VGroupName, "k8s/tcp/test/default/clusterid")
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}
}
