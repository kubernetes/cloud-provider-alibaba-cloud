package nlb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAddressType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Internet address type",
			input:    "Internet",
			expected: InternetAddressType,
		},
		{
			name:     "internet address type lowercase",
			input:    "internet",
			expected: InternetAddressType,
		},
		{
			name:     "Intranet address type",
			input:    "Intranet",
			expected: IntranetAddressType,
		},
		{
			name:     "intranet address type lowercase",
			input:    "intranet",
			expected: IntranetAddressType,
		},
		{
			name:     "unknown address type",
			input:    "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAddressType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAddressIpVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 version",
			input:    "ipv4",
			expected: IPv4,
		},
		{
			name:     "IPv4 version uppercase",
			input:    "IPv4",
			expected: IPv4,
		},
		{
			name:     "DualStack version",
			input:    "DualStack",
			expected: DualStack,
		},
		{
			name:     "dualstack version lowercase",
			input:    "dualstack",
			expected: DualStack,
		},
		{
			name:     "unknown version",
			input:    "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAddressIpVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetListenerProtocolType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "TCP protocol",
			input:    "TCP",
			expected: TCP,
		},
		{
			name:     "tcp protocol lowercase",
			input:    "tcp",
			expected: TCP,
		},
		{
			name:     "UDP protocol",
			input:    "UDP",
			expected: UDP,
		},
		{
			name:     "udp protocol lowercase",
			input:    "udp",
			expected: UDP,
		},
		{
			name:     "TCPSSL protocol",
			input:    "TCPSSL",
			expected: TCPSSL,
		},
		{
			name:     "tcpssl protocol lowercase",
			input:    "tcpssl",
			expected: TCPSSL,
		},
		{
			name:     "unknown protocol",
			input:    "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetListenerProtocolType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNetworkLoadBalancer_GetLoadBalancerId(t *testing.T) {
	// Test normal case
	lb := &NetworkLoadBalancer{
		LoadBalancerAttribute: &LoadBalancerAttribute{
			LoadBalancerId: "lb-12345",
		},
	}
	assert.Equal(t, "lb-12345", lb.GetLoadBalancerId())

	// Test nil LoadBalancerAttribute
	lbNil := &NetworkLoadBalancer{}
	assert.Equal(t, "", lbNil.GetLoadBalancerId())

	// Test nil NetworkLoadBalancer
	var nilLb *NetworkLoadBalancer
	assert.Equal(t, "", nilLb.GetLoadBalancerId())
}

func TestListenerAttribute_PortString(t *testing.T) {
	// Test with ListenerPort
	listener1 := &ListenerAttribute{
		ListenerPort: 80,
	}
	assert.Equal(t, "80", listener1.PortString())

	// Test with StartPort and EndPort
	listener2 := &ListenerAttribute{
		StartPort: 1000,
		EndPort:   2000,
	}
	assert.Equal(t, "1000-2000", listener2.PortString())

	// Test with all ports zero
	listener3 := &ListenerAttribute{}
	assert.Equal(t, "0-0", listener3.PortString())
}

func TestServerGroup_BackendInfo(t *testing.T) {
	// Test with few servers
	servers := make([]ServerGroupServer, 2)
	servers[0] = ServerGroupServer{
		ServerId: "server-1",
		Port:     80,
	}
	servers[1] = ServerGroupServer{
		ServerId: "server-2",
		Port:     8080,
	}

	sg := &ServerGroup{
		Servers: servers,
	}

	jsonData, _ := json.Marshal(servers)
	expected := string(jsonData)
	assert.Equal(t, expected, sg.BackendInfo())

	// Test with many servers (more than 100)
	manyServers := make([]ServerGroupServer, 150)
	for i := 0; i < 150; i++ {
		manyServers[i] = ServerGroupServer{
			ServerId: "server-" + string(rune(i)),
			Port:     int32(i),
		}
	}

	sgMany := &ServerGroup{
		Servers: manyServers,
	}

	// Should only include first 100
	truncatedServers := manyServers[:100]
	jsonData, _ = json.Marshal(truncatedServers)
	expected = string(jsonData)
	assert.Equal(t, expected, sgMany.BackendInfo())
}

func TestNamedKey_IsManagedByService(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
	}

	key := &NamedKey{
		CID:         "cluster-id",
		Namespace:   "test-namespace",
		ServiceName: "test-service",
	}

	// Test positive case
	assert.True(t, key.IsManagedByService(svc, "cluster-id"))

	// Test wrong cluster id
	assert.False(t, key.IsManagedByService(svc, "wrong-cluster-id"))

	// Test wrong namespace
	key.Namespace = "wrong-namespace"
	assert.False(t, key.IsManagedByService(svc, "cluster-id"))

	// Test wrong service name
	key.Namespace = "test-namespace"
	key.ServiceName = "wrong-service"
	assert.False(t, key.IsManagedByService(svc, "cluster-id"))

	// Test nil key
	var nilKey *NamedKey
	assert.False(t, nilKey.IsManagedByService(svc, "cluster-id"))
}

func TestListenerNamedKey_String(t *testing.T) {
	// Test normal case
	key := &ListenerNamedKey{
		NamedKey: NamedKey{
			Prefix:      "k8s",
			CID:         "cluster-id",
			Namespace:   "test-namespace",
			ServiceName: "test-service",
		},
		Port:     80,
		Protocol: "TCP",
	}

	expected := "k8s.80.TCP.test-service.test-namespace.cluster-id"
	assert.Equal(t, expected, key.String())
	assert.Equal(t, expected, key.Key())

	// Test with range ports
	keyRange := &ListenerNamedKey{
		NamedKey: NamedKey{
			Prefix:      "k8s",
			CID:         "cluster-id",
			Namespace:   "test-namespace",
			ServiceName: "test-service",
		},
		StartPort: 1000,
		EndPort:   2000,
		Protocol:  "TCP",
	}

	expectedRange := "k8s.1000_2000.TCP.test-service.test-namespace.cluster-id"
	assert.Equal(t, expectedRange, keyRange.String())
	assert.Equal(t, expectedRange, keyRange.Key())

	// Test nil key
	var nilKey *ListenerNamedKey
	assert.Equal(t, "", nilKey.String())
}

func TestLoadNLBListenerNamedKey(t *testing.T) {
	// Test valid key with single port
	validKey := "k8s.80.TCP.test-service.test-namespace.cluster-id"
	result, err := LoadNLBListenerNamedKey(validKey)
	assert.NoError(t, err)
	assert.Equal(t, int32(80), result.Port)
	assert.Equal(t, "TCP", result.Protocol)
	assert.Equal(t, "test-service", result.ServiceName)
	assert.Equal(t, "test-namespace", result.Namespace)
	assert.Equal(t, "cluster-id", result.CID)

	// Test valid key with port range
	validRangeKey := "k8s.1000_2000.TCP.test-service.test-namespace.cluster-id"
	resultRange, err := LoadNLBListenerNamedKey(validRangeKey)
	assert.NoError(t, err)
	assert.Equal(t, int32(1000), resultRange.StartPort)
	assert.Equal(t, int32(2000), resultRange.EndPort)
	assert.Equal(t, "TCP", resultRange.Protocol)
	assert.Equal(t, "test-service", resultRange.ServiceName)
	assert.Equal(t, "test-namespace", resultRange.Namespace)
	assert.Equal(t, "cluster-id", resultRange.CID)

	// Test invalid key format
	invalidKey := "invalid-key-format"
	_, err = LoadNLBListenerNamedKey(invalidKey)
	assert.Error(t, err)

	// Test invalid port range format
	invalidPortKey := "k8s.invalid_port.TCP.test-service.test-namespace.cluster-id"
	_, err = LoadNLBListenerNamedKey(invalidPortKey)
	assert.Error(t, err)
}

func TestSGNamedKey_String(t *testing.T) {
	// Test normal case
	key := &SGNamedKey{
		NamedKey: NamedKey{
			Prefix:      "k8s",
			CID:         "cluster-id",
			Namespace:   "test-namespace",
			ServiceName: "test-service",
		},
		Protocol:    "TCP",
		SGGroupPort: "80",
	}

	expected := "k8s.80.TCP.test-service.test-namespace.cluster-id"
	assert.Equal(t, expected, key.String())
	assert.Equal(t, expected, key.Key())

	// Test nil key
	var nilKey *SGNamedKey
	assert.Equal(t, "", nilKey.String())
}

func TestSGNamedKey_Key_EmptyPrefix(t *testing.T) {
	key := &SGNamedKey{
		NamedKey: NamedKey{
			Prefix:      "",
			CID:         "cluster-id",
			Namespace:   "test-namespace",
			ServiceName: "test-service",
		},
		Protocol:    "TCP",
		SGGroupPort: "80",
	}
	expected := "k8s.80.TCP.test-service.test-namespace.cluster-id"
	assert.Equal(t, expected, key.Key())
	assert.Equal(t, "k8s", key.Prefix)
}

func TestLoadNLBSGNamedKey(t *testing.T) {
	// Test valid key
	validKey := "k8s.80.TCP.test-service.test-namespace.cluster-id"
	result, err := LoadNLBSGNamedKey(validKey)
	assert.NoError(t, err)
	assert.Equal(t, "80", result.SGGroupPort)
	assert.Equal(t, "TCP", result.Protocol)
	assert.Equal(t, "test-service", result.ServiceName)
	assert.Equal(t, "test-namespace", result.Namespace)
	assert.Equal(t, "cluster-id", result.CID)

	// Test invalid key format
	invalidKey := "invalid-key-format"
	_, err = LoadNLBSGNamedKey(invalidKey)
	assert.Error(t, err)
}

func TestParseListenerPortKey(t *testing.T) {
	// Test single port
	port, startPort, endPort, err := parseListenerPortKey("80")
	assert.NoError(t, err)
	assert.Equal(t, int32(80), port)
	assert.Equal(t, int32(0), startPort)
	assert.Equal(t, int32(0), endPort)

	// Test port range
	port, startPort, endPort, err = parseListenerPortKey("1000_2000")
	assert.NoError(t, err)
	assert.Equal(t, int32(0), port)
	assert.Equal(t, int32(1000), startPort)
	assert.Equal(t, int32(2000), endPort)

	// Test invalid format
	_, _, _, err = parseListenerPortKey("invalid")
	assert.Error(t, err)

	// Test invalid range format
	_, _, _, err = parseListenerPortKey("1000_2000_3000")
	assert.Error(t, err)

	// Test invalid start port
	_, _, _, err = parseListenerPortKey("invalid_2000")
	assert.Error(t, err)

	// Test invalid end port
	_, _, _, err = parseListenerPortKey("1000_invalid")
	assert.Error(t, err)
}
