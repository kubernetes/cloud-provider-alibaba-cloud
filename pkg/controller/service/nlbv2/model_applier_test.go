package nlbv2

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"testing"
)

func TestModelApplier_Apply_CreateNLB(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)
	svc := &v1.Service{}
	_ = recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: v1.NamespaceDefault, Name: ServiceName}, svc)
	svc.UID = "ec0b5d7a-2764-4593-ba6c-fc2a57fa3884"
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
			Name:       "tcpssl",
			Port:       443,
			TargetPort: intstr.FromInt(443),
			NodePort:   443,
			Protocol:   v1.ProtocolTCP,
		},
	}
	svc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "udp:53,tcpssl:443"
	svc.Annotations[annotation.Annotation(annotation.CertID)] = "cert-id"
	svc.Annotations[annotation.Annotation(annotation.ZoneMaps)] = "cn-hangzhou-a:vsw-1,cn-hangzhou-b:vsw-2"
	svc.Annotations[annotation.Annotation(annotation.LoadBalancerName)] = "nlb-name"
	svc.Annotations[annotation.Annotation(annotation.ResourceGroupId)] = "rg-id"

	reqCtx := getReqCtx(svc)
	localModel, err := recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)

	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)
}

func TestModelApplier_Apply_UpdateNLB(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)
	svc := &v1.Service{}
	_ = recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: v1.NamespaceDefault, Name: ServiceName}, svc)
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
			Name:       "tcpssl",
			Port:       443,
			TargetPort: intstr.FromInt(443),
			NodePort:   443,
			Protocol:   v1.ProtocolTCP,
		},
	}

	svc.Annotations[annotation.Annotation(annotation.LoadBalancerId)] = vmock.ExistNLBID
	svc.Annotations[annotation.Annotation(annotation.OverrideListener)] = "true"
	svc.Annotations[annotation.Annotation(annotation.LoadBalancerName)] = "nlb-name"
	svc.Annotations[annotation.Annotation(annotation.ResourceGroupId)] = "rg-id"

	svc.Annotations[annotation.Annotation(annotation.ProtocolPort)] = "udp:53,tcpssl:443"
	svc.Annotations[annotation.Annotation(annotation.CertID)] = "cert-id"
	svc.Annotations[annotation.Annotation(annotation.CaCertID)] = "cacert-id"
	svc.Annotations[annotation.Annotation(annotation.CaCert)] = "on"
	svc.Annotations[annotation.Annotation(annotation.TLSCipherPolicy)] = "tls_cipher_policy_1_2"
	svc.Annotations[annotation.Annotation(annotation.ProxyProtocol)] = "on"
	svc.Annotations[annotation.Annotation(annotation.Cps)] = "60"
	svc.Annotations[annotation.Annotation(annotation.IdleTimeout)] = "15"

	svc.Annotations[annotation.Annotation(annotation.Scheduler)] = "rr"
	svc.Annotations[annotation.Annotation(annotation.ConnectionDrain)] = "on"
	svc.Annotations[annotation.Annotation(annotation.ConnectionDrainTimeout)] = "30"
	svc.Annotations[annotation.Annotation(annotation.PreserveClientIp)] = "on"

	svc.Annotations[annotation.Annotation(annotation.HealthCheckFlag)] = "on"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckType)] = "tcp"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckConnectPort)] = "8080"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckConnectTimeout)] = "3"
	svc.Annotations[annotation.Annotation(annotation.HealthyThreshold)] = "6"
	svc.Annotations[annotation.Annotation(annotation.UnhealthyThreshold)] = "5"
	svc.Annotations[annotation.Annotation(annotation.HealthCheckInterval)] = "5"

	reqCtx := getReqCtx(svc)
	localModel, err := recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)

}

func TestModelApplier_Apply_DeleteNLB(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)

	// auto created lb
	svc := &v1.Service{}
	_ = recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: v1.NamespaceDefault, Name: DelServiceName}, svc)
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
			Name:       "tcpssl",
			Port:       443,
			TargetPort: intstr.FromInt(443),
			NodePort:   443,
			Protocol:   v1.ProtocolTCP,
		},
	}
	reqCtx := getReqCtx(svc)
	localModel, err := recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)

	// reused lb
	svc.Annotations[annotation.Annotation(annotation.LoadBalancerId)] = vmock.ExistNLBID
	svc.Annotations[annotation.Annotation(annotation.OverrideListener)] = "true"

	reqCtx = getReqCtx(svc)
	localModel, err = recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)
}

func TestModelApplier_Apply_VServerGroup(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)

	svc := &v1.Service{}
	_ = recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: v1.NamespaceDefault, Name: ServiceName}, svc)
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
			Name:       "tcpssl",
			Port:       443,
			TargetPort: intstr.FromInt(443),
			NodePort:   443,
			Protocol:   v1.ProtocolTCP,
		},
	}

	// cluster mode
	clusterSvc := svc.DeepCopy()
	clusterSvc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
	reqCtx := getReqCtx(clusterSvc)
	localModel, err := recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)

	// local mode
	localSvc := svc.DeepCopy()
	localSvc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
	reqCtx = getReqCtx(localSvc)
	localModel, err = recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)

	// eni mode
	eniSvc := svc.DeepCopy()
	eniSvc.Annotations[annotation.BackendType] = "eni"
	reqCtx = getReqCtx(eniSvc)
	localModel, err = recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)

	// EndpointSlice
	_ = utilfeature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{string(ctrlCfg.EndpointSlice): true})
	esSvc := svc.DeepCopy()
	//esSvc.Annotations[annotation.Annotation(annotation.BackendIPVersion)] = nlbmodel.DualStack
	reqCtx = getReqCtx(esSvc)
	localModel, err = recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)
	_ = utilfeature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{string(ctrlCfg.EndpointSlice): false})

	// filter by label
	labelSvc := svc.DeepCopy()
	labelSvc.Annotations[annotation.Annotation(annotation.BackendLabel)] = "app=nginx"
	reqCtx = getReqCtx(labelSvc)
	localModel, err = recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)

	// string targetPort
	stringSvc := svc.DeepCopy()
	stringSvc.Spec.Ports = []v1.ServicePort{
		{
			Name:       "tcp",
			Port:       80,
			TargetPort: intstr.FromString("tcp"),
			NodePort:   80,
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "udp",
			Port:       53,
			TargetPort: intstr.FromString("udp"),
			NodePort:   53,
			Protocol:   v1.ProtocolTCP,
		},
	}
	reqCtx = getReqCtx(stringSvc)
	localModel, err = recon.builder.Instance(LocalModel).Build(reqCtx)
	assert.Equal(t, nil, err)
	_, err = recon.applier.Apply(reqCtx, localModel)
	assert.Equal(t, nil, err)
}
