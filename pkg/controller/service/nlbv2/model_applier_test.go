package nlbv2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
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

func TestModelApplier_BuildServerGroupCreateAndUpdateActions(t *testing.T) {
	tests := []struct {
		name                   string
		localServerGroupType   nlbmodel.ServerGroupType
		remoteServerGroupType  nlbmodel.ServerGroupType
		localAddressIPVersion  string
		remoteAddressIPVersion string
		userManaged            bool
		wantAction             serverGroupActionType
		wantError              string
	}{
		{
			name:                   "empty local address IP version equals remote IPv4",
			localServerGroupType:   nlbmodel.IpServerGroupType,
			remoteServerGroupType:  nlbmodel.IpServerGroupType,
			remoteAddressIPVersion: nlbmodel.IPv4,
			wantAction:             serverGroupActionUpdate,
		},
		{
			name:                  "local IPv4 equals empty remote address IP version",
			localServerGroupType:  nlbmodel.IpServerGroupType,
			remoteServerGroupType: nlbmodel.IpServerGroupType,
			localAddressIPVersion: nlbmodel.IPv4,
			wantAction:            serverGroupActionUpdate,
		},
		{
			name:                   "IPv4 to DualStack recreates server group",
			localServerGroupType:   nlbmodel.IpServerGroupType,
			remoteServerGroupType:  nlbmodel.IpServerGroupType,
			localAddressIPVersion:  nlbmodel.DualStack,
			remoteAddressIPVersion: nlbmodel.IPv4,
			wantAction:             serverGroupActionCreateAndAddBackendServers,
		},
		{
			name:                   "DualStack to default IPv4 recreates server group",
			localServerGroupType:   nlbmodel.IpServerGroupType,
			remoteServerGroupType:  nlbmodel.IpServerGroupType,
			remoteAddressIPVersion: nlbmodel.DualStack,
			wantAction:             serverGroupActionCreateAndAddBackendServers,
		},
		{
			name:                   "address IP version comparison ignores case",
			localServerGroupType:   nlbmodel.IpServerGroupType,
			remoteServerGroupType:  nlbmodel.IpServerGroupType,
			localAddressIPVersion:  "dualstack",
			remoteAddressIPVersion: nlbmodel.DualStack,
			wantAction:             serverGroupActionUpdate,
		},
		{
			name:                   "server group type change still recreates server group",
			localServerGroupType:   nlbmodel.IpServerGroupType,
			remoteServerGroupType:  nlbmodel.InstanceServerGroupType,
			localAddressIPVersion:  nlbmodel.IPv4,
			remoteAddressIPVersion: nlbmodel.IPv4,
			wantAction:             serverGroupActionCreateAndAddBackendServers,
		},
		{
			name:                   "user managed address IP version mismatch returns error",
			localServerGroupType:   nlbmodel.IpServerGroupType,
			remoteServerGroupType:  nlbmodel.IpServerGroupType,
			localAddressIPVersion:  nlbmodel.DualStack,
			remoteAddressIPVersion: nlbmodel.IPv4,
			userManaged:            true,
			wantError:              "AddressIPVersion of user managed server group sg-id should be [DualStack], but [ipv4]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			local := &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{},
				ServerGroups: []*nlbmodel.ServerGroup{
					{
						IsUserManaged:    tt.userManaged,
						ServerGroupId:    "sg-id",
						ServerGroupName:  "sg-name",
						ServerGroupType:  tt.localServerGroupType,
						AddressIPVersion: tt.localAddressIPVersion,
					},
				},
			}
			remote := &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{},
				ServerGroups: []*nlbmodel.ServerGroup{
					{
						ServerGroupId:    "sg-id",
						ServerGroupName:  "sg-name",
						ServerGroupType:  tt.remoteServerGroupType,
						AddressIPVersion: tt.remoteAddressIPVersion,
					},
				},
			}

			actions, err := (&ModelApplier{}).buildServerGroupCreateAndUpdateActions(
				getReqCtx(&v1.Service{}), local, remote)
			if tt.wantError != "" {
				assert.EqualError(t, err, tt.wantError)
				assert.Nil(t, actions)
				return
			}

			assert.NoError(t, err)
			if assert.Len(t, actions, 1) {
				assert.Equal(t, tt.wantAction, actions[0].Action)
			}
		})
	}
}

func TestIsNLBReusable(t *testing.T) {
	tests := []struct {
		name     string
		service  *v1.Service
		tags     []tag.Tag
		dnsName  string
		expected bool
		reason   string
	}{
		{
			name: "reusable with no tags",
			service: &v1.Service{
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{},
				},
			},
			tags:     []tag.Tag{},
			dnsName:  "test-dns",
			expected: true,
			reason:   "",
		},
		{
			name: "not reusable with TAGKEY tag",
			service: &v1.Service{
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{},
				},
			},
			tags: []tag.Tag{
				{
					Key:   helper.TAGKEY,
					Value: "cluster-id",
				},
			},
			dnsName:  "test-dns",
			expected: false,
			reason:   "can not reuse loadbalancer created by kubernetes.",
		},
		{
			name: "not reusable with ClusterTagKey tag",
			service: &v1.Service{
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{},
				},
			},
			tags: []tag.Tag{
				{
					Key:   util.ClusterTagKey,
					Value: "cluster-id",
				},
			},
			dnsName:  "test-dns",
			expected: false,
			reason:   "can not reuse loadbalancer created by kubernetes.",
		},
		{
			name: "reusable with matching dnsName",
			service: &v1.Service{
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								Hostname: "test-dns",
							},
						},
					},
				},
			},
			tags:     []tag.Tag{},
			dnsName:  "test-dns",
			expected: true,
			reason:   "",
		},
		{
			name: "not reusable with different dnsName",
			service: &v1.Service{
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								Hostname: "other-dns",
							},
						},
					},
				},
			},
			tags:     []tag.Tag{},
			dnsName:  "test-dns",
			expected: false,
			reason:   "service has been associated with dnsname",
		},
		{
			name: "reusable with multiple ingress matching dnsName",
			service: &v1.Service{
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								Hostname: "other-dns",
							},
							{
								Hostname: "test-dns",
							},
						},
					},
				},
			},
			tags:     []tag.Tag{},
			dnsName:  "test-dns",
			expected: true,
			reason:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reusable, reason := isNLBReusable(tt.service, tt.tags, tt.dnsName)
			assert.Equal(t, tt.expected, reusable)
			if tt.reason != "" {
				assert.Contains(t, reason, tt.reason)
			}
		})
	}
}

func TestBuildActionsForListenersPreserveOnDelete(t *testing.T) {
	deletionTimestamp := metav1.Now()
	tests := []struct {
		name           string
		service        *v1.Service
		preserve       bool
		localListeners []*nlbmodel.ListenerAttribute
		wantActions    []listenerActionType
	}{
		{
			name: "preserve deleting service",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &deletionTimestamp},
				Spec:       v1.ServiceSpec{Type: v1.ServiceTypeLoadBalancer},
			},
			preserve: true,
		},
		{
			name: "delete non-preserved service listener",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &deletionTimestamp},
				Spec:       v1.ServiceSpec{Type: v1.ServiceTypeLoadBalancer},
			},
			wantActions: []listenerActionType{listenerActionDelete},
		},
		{
			name: "reconcile preserved service normally",
			service: &v1.Service{
				Spec: v1.ServiceSpec{Type: v1.ServiceTypeLoadBalancer},
			},
			preserve: true,
			localListeners: []*nlbmodel.ListenerAttribute{
				{ListenerProtocol: nlbmodel.TCP, ListenerPort: 80, ServerGroupId: "sg-id"},
			},
			wantActions: []listenerActionType{listenerActionUpdate},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			local := &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{PreserveOnDelete: tt.preserve},
				Listeners:             tt.localListeners,
			}
			remote := &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{LoadBalancerId: "nlb-id"},
				Listeners: []*nlbmodel.ListenerAttribute{
					{ListenerId: "listener-id", ListenerProtocol: nlbmodel.TCP, ListenerPort: 80},
				},
			}

			actions, err := buildActionsForListeners(getReqCtx(tt.service), local, remote)
			assert.NoError(t, err)

			var gotActions []listenerActionType
			for _, action := range actions {
				gotActions = append(gotActions, action.Action)
			}
			assert.Equal(t, tt.wantActions, gotActions)
		})
	}
}

func TestIsListenerPortMatch(t *testing.T) {
	t.Run("ListenerPort match", func(t *testing.T) {
		l := &nlbmodel.ListenerAttribute{ListenerPort: 80}
		r := &nlbmodel.ListenerAttribute{ListenerPort: 80}
		assert.True(t, isListenerPortMatch(l, r))
	})
	t.Run("ListenerPort mismatch", func(t *testing.T) {
		l := &nlbmodel.ListenerAttribute{ListenerPort: 80}
		r := &nlbmodel.ListenerAttribute{ListenerPort: 443}
		assert.False(t, isListenerPortMatch(l, r))
	})
	t.Run("StartPort EndPort match", func(t *testing.T) {
		l := &nlbmodel.ListenerAttribute{StartPort: 1000, EndPort: 2000}
		r := &nlbmodel.ListenerAttribute{StartPort: 1000, EndPort: 2000}
		assert.True(t, isListenerPortMatch(l, r))
	})
	t.Run("StartPort EndPort mismatch", func(t *testing.T) {
		l := &nlbmodel.ListenerAttribute{StartPort: 1000, EndPort: 2000}
		r := &nlbmodel.ListenerAttribute{StartPort: 1000, EndPort: 3000}
		assert.False(t, isListenerPortMatch(l, r))
	})
}
