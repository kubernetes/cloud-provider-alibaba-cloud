package service

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
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
	svc.Annotations[Annotation(ProtocolPort)] = "tcp:80,udp:53,http:8080,https:443"

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
	svc.Annotations[Annotation(LoadBalancerId)] = vmock.ExistLBID
	svc.Annotations[Annotation(OverrideListener)] = "true"
	svc.Annotations[Annotation(Spec)] = "slb.s2.small"
	svc.Annotations[Annotation(ChargeType)] = "paybybandwidth"
	svc.Annotations[Annotation(Bandwidth)] = "5"
	svc.Annotations[Annotation(DeleteProtection)] = "off"
	svc.Annotations[Annotation(LoadBalancerName)] = "new-lb-name"
	svc.Annotations[Annotation(ModificationProtection)] = "NonProtection"

	svc.Annotations[Annotation(AclID)] = "acl-id"
	svc.Annotations[Annotation(AclStatus)] = string(model.OnFlag)
	svc.Annotations[Annotation(AclType)] = "white"
	svc.Annotations[Annotation(Scheduler)] = "wrr"
	svc.Annotations[Annotation(PersistenceTimeout)] = "10"
	svc.Annotations[Annotation(EstablishedTimeout)] = "12"
	svc.Annotations[Annotation(EnableHttp2)] = "false"
	svc.Annotations[Annotation(IdleTimeout)] = "60"
	svc.Annotations[Annotation(RequestTimeout)] = "30"
	svc.Annotations[Annotation(ConnectionDrain)] = "on"
	svc.Annotations[Annotation(ConnectionDrainTimeout)] = "30"
	svc.Annotations[Annotation(Cookie)] = "test-cookie"
	svc.Annotations[Annotation(CookieTimeout)] = "60"
	svc.Annotations[Annotation(SessionStick)] = "on"
	svc.Annotations[Annotation(SessionStickType)] = "insert"
	svc.Annotations[Annotation(XForwardedForProto)] = "on"

	svc.Annotations[Annotation(ProtocolPort)] = "tcp:80,udp:53,http:8080,https:443"
	svc.Annotations[Annotation(CertID)] = "cert-id"
	svc.Annotations[Annotation(ForwardPort)] = "8080:443"

	svc.Annotations[Annotation(HealthyThreshold)] = "6"
	svc.Annotations[Annotation(UnhealthyThreshold)] = "5"
	svc.Annotations[Annotation(HealthCheckConnectTimeout)] = "3"
	svc.Annotations[Annotation(HealthCheckConnectPort)] = "80"
	svc.Annotations[Annotation(HealthCheckInterval)] = "5"
	svc.Annotations[Annotation(HealthCheckDomain)] = "foo2.bar.com"
	svc.Annotations[Annotation(HealthCheckURI)] = "/test2/index.html"
	svc.Annotations[Annotation(HealthCheckHTTPCode)] = "http_2xx,http_3xx"
	svc.Annotations[Annotation(HealthCheckType)] = "tcp"
	svc.Annotations[Annotation(HealthCheckFlag)] = "on"
	svc.Annotations[Annotation(HealthCheckTimeout)] = "3"
	svc.Annotations[Annotation(HealthCheckMethod)] = "get"

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
	svc.ObjectMeta.Finalizers = []string{ServiceFinalizer}
	svc.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	svc.Labels = map[string]string{LabelServiceHash: getServiceHash(svc)}
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
	reqCtx.Service.Annotations[Annotation(LoadBalancerId)] = "exist-lb"
	reqCtx.Service.Annotations[Annotation(OverrideListener)] = "true"
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
	reqCtx.Service.Annotations = map[string]string{BackendType: model.ENIBackendType}
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
	reqCtx.Service.Annotations[Annotation(LoadBalancerId)] = vmock.ExistLBID
	reqCtx.Service.Annotations[Annotation(VGroupPort)] = fmt.Sprintf("%s:%d", vmock.ExistVGroupID, 80)
	reqCtx.Service.Annotations[Annotation(VGroupWeight)] = "80"
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
	reqCtx.Service.Annotations[Annotation(LoadBalancerId)] = vmock.ExistLBID
	reqCtx.Service.Annotations[Annotation(VGroupPort)] = fmt.Sprintf("%s:%d", vmock.ExistVGroupID, 80)
	reqCtx.Service.Annotations[Annotation(VGroupWeight)] = "0"
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
	_ = utilfeature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{string(helper.EndpointSlice): true})
	reqCtx = getReqCtx(getDefaultService())
	localModel, err = builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)
	}
	_ = utilfeature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{string(helper.EndpointSlice): false})

	// filter out by label
	reqCtx = getReqCtx(getDefaultService())
	reqCtx.Service.Annotations[Annotation(BackendLabel)] = "app=nginx"
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
	reqCtx.Service.Annotations[BackendType] = model.ENIBackendType
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
