package clbv1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/klog/v2/klogr"
)

func TestIsListenerACLIDsEqual(t *testing.T) {
	cases := []struct {
		name     string
		local    model.ListenerAttribute
		remote   model.ListenerAttribute
		expected bool
	}{
		{
			name:     "empty",
			local:    model.ListenerAttribute{},
			remote:   model.ListenerAttribute{},
			expected: true,
		},
		{
			name: "local one",
			local: model.ListenerAttribute{
				AclId: "acl-123456",
			},
			remote:   model.ListenerAttribute{},
			expected: false,
		},
		{
			name:  "remote one",
			local: model.ListenerAttribute{},
			remote: model.ListenerAttribute{
				AclIds: []string{"acl-123456"},
			},
			expected: false,
		},
		{
			name: "local remote one",
			local: model.ListenerAttribute{
				AclId: "acl-123456",
			},
			remote: model.ListenerAttribute{
				AclIds: []string{"acl-123456"},
			},
			expected: true,
		},
		{
			name: "local retmoe multi equal",
			local: model.ListenerAttribute{
				AclId: "acl-123456,acl-1234567",
			},
			remote: model.ListenerAttribute{
				AclIds: []string{"acl-1234567", "acl-123456"},
			},
			expected: true,
		},
		{
			name: "local remote multi not equal",
			local: model.ListenerAttribute{
				AclId: "acl-123456,acl-1234567",
			},
			remote: model.ListenerAttribute{
				AclIds: []string{"acl-12345678", "acl-123456"},
			},
			expected: false,
		},
		{
			name: "remote aclid",
			local: model.ListenerAttribute{
				AclId: "acl-123456",
			},
			remote: model.ListenerAttribute{
				AclId: "acl-123456",
			},
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, IsListenerACLIDsEqual(c.local, c.remote))
		})
	}
}

func TestForwardPort(t *testing.T) {
	cases := []struct {
		name        string
		port        string
		target      int
		expected    int
		expectError bool
	}{
		{
			name:        "empty port",
			port:        "",
			target:      80,
			expected:    0,
			expectError: true,
		},
		{
			name:        "valid forward",
			port:        "80:443",
			target:      80,
			expected:    443,
			expectError: false,
		},
		{
			name:        "multiple forwards",
			port:        "80:443,8080:8443",
			target:      8080,
			expected:    8443,
			expectError: false,
		},
		{
			name:        "no match",
			port:        "80:443",
			target:      8080,
			expected:    0,
			expectError: false,
		},
		{
			name:        "invalid format",
			port:        "80",
			target:      80,
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid number",
			port:        "80:abc",
			target:      80,
			expected:    0,
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := forwardPort(c.port, c.target)
			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, result)
			}
		})
	}
}

func TestProtocol(t *testing.T) {
	cases := []struct {
		name        string
		annotation  string
		port        v1.ServicePort
		expected    string
		expectError bool
	}{
		{
			name:        "no annotation",
			annotation:  "",
			port:        v1.ServicePort{Port: 80, Protocol: v1.ProtocolTCP},
			expected:    "tcp",
			expectError: false,
		},
		{
			name:        "https protocol",
			annotation:  "https:443",
			port:        v1.ServicePort{Port: 443, Protocol: v1.ProtocolTCP},
			expected:    "https",
			expectError: false,
		},
		{
			name:        "http protocol",
			annotation:  "http:80",
			port:        v1.ServicePort{Port: 80, Protocol: v1.ProtocolTCP},
			expected:    "http",
			expectError: false,
		},
		{
			name:        "udp protocol",
			annotation:  "udp:53",
			port:        v1.ServicePort{Port: 53, Protocol: v1.ProtocolUDP},
			expected:    "udp",
			expectError: false,
		},
		{
			name:        "multiple protocols",
			annotation:  "http:80,https:443",
			port:        v1.ServicePort{Port: 443, Protocol: v1.ProtocolTCP},
			expected:    "https",
			expectError: false,
		},
		{
			name:        "invalid format",
			annotation:  "https",
			port:        v1.ServicePort{Port: 443, Protocol: v1.ProtocolTCP},
			expected:    "",
			expectError: true,
		},
		{
			name:        "unsupported protocol",
			annotation:  "ftp:21",
			port:        v1.ServicePort{Port: 21, Protocol: v1.ProtocolTCP},
			expected:    "",
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := protocol(c.annotation, c.port)
			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, result)
			}
		})
	}
}

func TestVgroup(t *testing.T) {
	cases := []struct {
		name        string
		annotation  string
		port        v1.ServicePort
		expected    string
		expectError bool
	}{
		{
			name:        "valid vgroup",
			annotation:  "vsp-xxx:80",
			port:        v1.ServicePort{Port: 80},
			expected:    "vsp-xxx",
			expectError: false,
		},
		{
			name:        "multiple vgroups",
			annotation:  "vsp-xxx:80,vsp-yyy:443",
			port:        v1.ServicePort{Port: 443},
			expected:    "vsp-yyy",
			expectError: false,
		},
		{
			name:        "no match",
			annotation:  "vsp-xxx:80",
			port:        v1.ServicePort{Port: 443},
			expected:    "",
			expectError: false,
		},
		{
			name:        "invalid format",
			annotation:  "vsp-xxx",
			port:        v1.ServicePort{Port: 80},
			expected:    "",
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := vgroup(c.annotation, c.port)
			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, result)
			}
		})
	}
}

func TestSetDefaultValueForListener(t *testing.T) {
	t.Run("set default scheduler", func(t *testing.T) {
		listener := &model.ListenerAttribute{}
		setDefaultValueForListener(listener)
		assert.Equal(t, "rr", listener.Scheduler)
	})

	t.Run("keep existing scheduler", func(t *testing.T) {
		listener := &model.ListenerAttribute{Scheduler: "wrr"}
		setDefaultValueForListener(listener)
		assert.Equal(t, "wrr", listener.Scheduler)
	})

	t.Run("set default bandwidth for tcp", func(t *testing.T) {
		listener := &model.ListenerAttribute{Protocol: model.TCP}
		setDefaultValueForListener(listener)
		assert.Equal(t, DefaultListenerBandwidth, listener.Bandwidth)
	})

	t.Run("set default health check for http", func(t *testing.T) {
		listener := &model.ListenerAttribute{Protocol: model.HTTP}
		setDefaultValueForListener(listener)
		assert.Equal(t, model.OffFlag, listener.HealthCheck)
		assert.Equal(t, model.OffFlag, listener.StickySession)
	})
}

func TestFindVServerGroup(t *testing.T) {
	t.Run("found vgroup", func(t *testing.T) {
		vgs := []model.VServerGroup{
			{VGroupId: "vsg-1", VGroupName: "group-1"},
			{VGroupId: "vsg-2", VGroupName: "group-2"},
		}
		listener := &model.ListenerAttribute{VGroupName: "group-1"}
		err := findVServerGroup(vgs, listener)
		assert.NoError(t, err)
		assert.Equal(t, "vsg-1", listener.VGroupId)
	})

	t.Run("not found vgroup", func(t *testing.T) {
		vgs := []model.VServerGroup{
			{VGroupId: "vsg-1", VGroupName: "group-1"},
		}
		listener := &model.ListenerAttribute{VGroupName: "group-2"}
		err := findVServerGroup(vgs, listener)
		assert.Error(t, err)
	})
}

func TestIsPortManagedByMyService(t *testing.T) {
	t.Run("managed by my service", func(t *testing.T) {
		local := &model.LoadBalancer{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}
		listener := model.ListenerAttribute{
			NamedKey: &model.ListenerNamedKey{
				ServiceName: "test",
				Namespace:   "default",
				CID:         base.CLUSTER_ID,
			},
		}
		result := isPortManagedByMyService(local, listener)
		assert.True(t, result)
	})

	t.Run("not managed - user managed", func(t *testing.T) {
		local := &model.LoadBalancer{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}
		listener := model.ListenerAttribute{
			IsUserManaged: true,
			NamedKey: &model.ListenerNamedKey{
				ServiceName: "test",
				Namespace:   "default",
				CID:         base.CLUSTER_ID,
			},
		}
		result := isPortManagedByMyService(local, listener)
		assert.False(t, result)
	})

	t.Run("not managed - different service", func(t *testing.T) {
		local := &model.LoadBalancer{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}
		listener := model.ListenerAttribute{
			NamedKey: &model.ListenerNamedKey{
				ServiceName: "other",
				Namespace:   "default",
				CID:         base.CLUSTER_ID,
			},
		}
		result := isPortManagedByMyService(local, listener)
		assert.False(t, result)
	})
}

func TestGetVGroupNamedKey(t *testing.T) {
	t.Run("ecs backend", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		port := v1.ServicePort{
			Port:       80,
			TargetPort: intstr.FromInt(8080),
			NodePort:   30080,
		}
		key := getVGroupNamedKey(svc, port)
		assert.Equal(t, "default", key.Namespace)
		assert.Equal(t, "test", key.ServiceName)
		assert.Equal(t, "30080", key.VGroupPort)
	})

	t.Run("eni backend with int target port", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test",
				Namespace:   "default",
				Annotations: map[string]string{annotation.BackendType: model.ENIBackendType},
			},
		}
		port := v1.ServicePort{
			Port:       80,
			TargetPort: intstr.FromInt(8080),
			NodePort:   30080,
		}
		key := getVGroupNamedKey(svc, port)
		assert.Equal(t, "8080", key.VGroupPort)
	})

	t.Run("eni backend with string target port", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test",
				Namespace:   "default",
				Annotations: map[string]string{annotation.BackendType: model.ENIBackendType},
			},
		}
		port := v1.ServicePort{
			Port:       80,
			TargetPort: intstr.FromString("http"),
			NodePort:   30080,
		}
		key := getVGroupNamedKey(svc, port)
		assert.Equal(t, "http", key.VGroupPort)
	})
}

func TestHttpUpdate(t *testing.T) {
	mgr := NewListenerManager(getMockCloudProvider())
	httpListener := &http{mgr: mgr}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     klogr.New(),
	}

	t.Run("remote listener is stopped", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort: 80,
			Protocol:     model.HTTP,
			Status:       model.Running,
		}
		remote := model.ListenerAttribute{
			ListenerPort: 80,
			Protocol:     model.HTTP,
			Status:       model.Stopped,
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})

	t.Run("forward port changed need recreate", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort:    80,
			Protocol:        model.HTTP,
			ForwardPort:     443,
			ListenerForward: model.OffFlag,
		}
		remote := model.ListenerAttribute{
			ListenerPort:    80,
			Protocol:        model.HTTP,
			ForwardPort:     8080,
			ListenerForward: model.OnFlag,
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})

	t.Run("forward port disabled need recreate", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort:    80,
			Protocol:        model.HTTP,
			ForwardPort:     0,
			ListenerForward: model.OffFlag,
		}
		remote := model.ListenerAttribute{
			ListenerPort:    80,
			Protocol:        model.HTTP,
			ForwardPort:     443,
			ListenerForward: model.OnFlag,
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})

	t.Run("listener forward is on skip update", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort: 80,
			Protocol:     model.HTTP,
			ForwardPort:  0,
		}
		remote := model.ListenerAttribute{
			ListenerPort:    80,
			Protocol:        model.HTTP,
			ListenerForward: model.OnFlag,
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})

	t.Run("no change skip update", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort: 80,
			Protocol:     model.HTTP,
			Scheduler:    "rr",
		}
		remote := model.ListenerAttribute{
			ListenerPort: 80,
			Protocol:     model.HTTP,
			Scheduler:    "rr",
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})

	t.Run("normal update", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort: 80,
			Protocol:     model.HTTP,
			Scheduler:    "wrr",
		}
		remote := model.ListenerAttribute{
			ListenerPort: 80,
			Protocol:     model.HTTP,
			Scheduler:    "rr",
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})
}

func TestHttpsUpdate(t *testing.T) {
	mgr := NewListenerManager(getMockCloudProvider())
	httpsListener := &https{mgr: mgr}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     klogr.New(),
	}

	t.Run("remote listener is stopped", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort: 443,
			Protocol:     model.HTTPS,
			Status:       model.Running,
		}
		remote := model.ListenerAttribute{
			ListenerPort: 443,
			Protocol:     model.HTTPS,
			Status:       model.Stopped,
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpsListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})

	t.Run("cert id changed", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort: 443,
			Protocol:     model.HTTPS,
			CertId:       "cert-new-id",
		}
		remote := model.ListenerAttribute{
			ListenerPort: 443,
			Protocol:     model.HTTPS,
			CertId:       "cert-old-id",
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpsListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})

	t.Run("no change skip update", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort: 443,
			Protocol:     model.HTTPS,
			Scheduler:    "rr",
		}
		remote := model.ListenerAttribute{
			ListenerPort: 443,
			Protocol:     model.HTTPS,
			Scheduler:    "rr",
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpsListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})

	t.Run("normal update", func(t *testing.T) {
		local := model.ListenerAttribute{
			ListenerPort: 443,
			Protocol:     model.HTTPS,
			Scheduler:    "wrr",
		}
		remote := model.ListenerAttribute{
			ListenerPort: 443,
			Protocol:     model.HTTPS,
			Scheduler:    "rr",
		}
		action := UpdateAction{
			lbId:   "lb-test-id",
			local:  local,
			remote: remote,
		}
		err := httpsListener.Update(reqCtx, action)
		assert.NoError(t, err)
	})
}

func TestCheckCertValidity(t *testing.T) {
	cloud := getMockCloudProvider()

	t.Run("old cert is nil", func(t *testing.T) {
		err := checkCertValidity(cloud, "cert-not-found", "cert-new-id")
		assert.NoError(t, err)
	})

	t.Run("new cert is nil", func(t *testing.T) {
		err := checkCertValidity(cloud, "cert-old-id", "cert-not-found")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not found cert by id")
	})

	t.Run("both certs are nil", func(t *testing.T) {
		err := checkCertValidity(cloud, "cert-not-found", "cert-not-found-1")
		assert.NoError(t, err)
	})

	t.Run("old cert is expired", func(t *testing.T) {
		err := checkCertValidity(cloud, "cert-expired", "cert-test-1")
		assert.NoError(t, err)
	})

	t.Run("new cert is expired", func(t *testing.T) {
		err := checkCertValidity(cloud, "cert-test-1", "cert-expired")
		assert.Error(t, err)
	})
}

func TestParseDomainExtensionsAnnotation(t *testing.T) {
	t.Run("empty annotation", func(t *testing.T) {
		result, err := parseDomainExtensionsAnnotation("")
		assert.NoError(t, err)
		assert.Equal(t, []model.DomainExtension{}, result)
	})

	t.Run("single domain extension", func(t *testing.T) {
		result, err := parseDomainExtensionsAnnotation("domain1:certId1")
		assert.NoError(t, err)
		expected := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "certId1"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("multiple domain extensions", func(t *testing.T) {
		result, err := parseDomainExtensionsAnnotation("domain1:certId1,domain2:certId2")
		assert.NoError(t, err)
		expected := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "certId1"},
			{Domain: "domain2", ServerCertificateId: "certId2"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("domain extensions with spaces", func(t *testing.T) {
		result, err := parseDomainExtensionsAnnotation("domain1:certId1, domain2:certId2 ")
		assert.NoError(t, err)
		expected := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "certId1"},
			{Domain: "domain2", ServerCertificateId: "certId2"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("invalid format - missing separator", func(t *testing.T) {
		result, err := parseDomainExtensionsAnnotation("domain1certId1")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid domain extension")
	})

	t.Run("invalid format - extra separator", func(t *testing.T) {
		result, err := parseDomainExtensionsAnnotation("domain1:certId1:extra")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid domain extension")
	})
}

func TestDiffDomainExtensions(t *testing.T) {
	t.Run("no changes", func(t *testing.T) {
		local := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1"},
		}
		remote := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1"},
		}
		toAdd, toDelete, toUpdate := diffDomainExtensions(local, remote)
		assert.Empty(t, toAdd)
		assert.Empty(t, toDelete)
		assert.Empty(t, toUpdate)
	})

	t.Run("add domain extension", func(t *testing.T) {
		local := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1"},
			{Domain: "domain2", ServerCertificateId: "cert2"},
		}
		remote := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1"},
		}
		toAdd, toDelete, toUpdate := diffDomainExtensions(local, remote)
		assert.Equal(t, []model.DomainExtension{{Domain: "domain2", ServerCertificateId: "cert2"}}, toAdd)
		assert.Empty(t, toDelete)
		assert.Empty(t, toUpdate)
	})

	t.Run("delete domain extension", func(t *testing.T) {
		local := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1"},
		}
		remote := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1"},
			{Domain: "domain2", ServerCertificateId: "cert2"},
		}
		toAdd, toDelete, toUpdate := diffDomainExtensions(local, remote)
		assert.Empty(t, toAdd)
		assert.Equal(t, []model.DomainExtension{{Domain: "domain2", ServerCertificateId: "cert2"}}, toDelete)
		assert.Empty(t, toUpdate)
	})

	t.Run("update domain extension", func(t *testing.T) {
		local := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1_new"},
		}
		remote := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1_old"},
		}
		toAdd, toDelete, toUpdate := diffDomainExtensions(local, remote)
		// Note: In the original function, the update logic modifies the local object
		// So the ServerCertificateId becomes the remote value
		expectedUpdate := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1_old"},
		}
		assert.Empty(t, toAdd)
		assert.Empty(t, toDelete)
		assert.Equal(t, expectedUpdate, toUpdate)
	})

	t.Run("complex scenario", func(t *testing.T) {
		local := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1_new"}, // update
			{Domain: "domain3", ServerCertificateId: "cert3"},    // add
		}
		remote := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1_old"}, // update
			{Domain: "domain2", ServerCertificateId: "cert2"},     // delete
		}
		toAdd, toDelete, toUpdate := diffDomainExtensions(local, remote)
		assert.Equal(t, []model.DomainExtension{{Domain: "domain3", ServerCertificateId: "cert3"}}, toAdd)
		assert.Equal(t, []model.DomainExtension{{Domain: "domain2", ServerCertificateId: "cert2"}}, toDelete)
		expectedUpdate := []model.DomainExtension{
			{Domain: "domain1", ServerCertificateId: "cert1_old"},
		}
		assert.Equal(t, expectedUpdate, toUpdate)
	})
}
