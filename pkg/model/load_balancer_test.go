package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLoadBalancerId(t *testing.T) {
	var lb *LoadBalancer
	id := lb.GetLoadBalancerId()
	assert.Equal(t, id, "")

	lb2 := &LoadBalancer{}
	id2 := lb2.GetLoadBalancerId()
	assert.Equal(t, id2, "")

	lb3 := &LoadBalancer{
		LoadBalancerAttribute: LoadBalancerAttribute{
			LoadBalancerId: "lb-xxxx",
		},
	}
	id3 := lb3.GetLoadBalancerId()
	assert.Equal(t, id3, "lb-xxxx")
}

func TestParseFlagType(t *testing.T) {
	cases := []struct {
		flag    string
		value   FlagType
		wantErr bool
	}{
		{
			flag:  "on",
			value: OnFlag,
		},
		{
			flag:  "off",
			value: OffFlag,
		},
		{
			flag:  "ON",
			value: OnFlag,
		},
		{
			flag:  "Off",
			value: OffFlag,
		},
		{
			flag:    "true",
			wantErr: true,
		},
	}

	for _, c := range cases {
		value, err := ParseFlagType(c.flag)
		if c.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, c.value, value)
		}
	}
}

func TestInstanceChargeType_IsPayBySpec(t *testing.T) {
	tests := []struct {
		name string
		t    InstanceChargeType
		want bool
	}{
		{
			name: "PayBySpec uppercase",
			t:    PayBySpec,
			want: true,
		},
		{
			name: "PayBySpec lowercase",
			t:    "paybyspec",
			want: true,
		},
		{
			name: "PayByCLCU",
			t:    PayByCLCU,
			want: false,
		},
		{
			name: "empty string",
			t:    "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.t.IsPayBySpec())
		})
	}
}

func TestInstanceChargeType_IsPayByCLCU(t *testing.T) {
	tests := []struct {
		name string
		t    InstanceChargeType
		want bool
	}{
		{
			name: "PayByCLCU uppercase",
			t:    PayByCLCU,
			want: true,
		},
		{
			name: "PayByCLCU lowercase",
			t:    "paybyclcu",
			want: true,
		},
		{
			name: "empty string",
			t:    "",
			want: true,
		},
		{
			name: "PayBySpec",
			t:    PayBySpec,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.t.IsPayByCLCU())
		})
	}
}

func TestVServerGroup_BackendInfo(t *testing.T) {
	tests := []struct {
		name     string
		backends []BackendAttribute
		wantJSON bool
	}{
		{
			name: "few backends",
			backends: []BackendAttribute{
				{ServerId: "i-123", ServerIp: "192.168.1.1", Port: 80, Weight: 100},
				{ServerId: "i-456", ServerIp: "192.168.1.2", Port: 80, Weight: 100},
			},
			wantJSON: true,
		},
		{
			name:     "empty backends",
			backends: []BackendAttribute{},
			wantJSON: true,
		},
		{
			name: "more than 100 backends",
			backends: func() []BackendAttribute {
				var backends []BackendAttribute
				for i := 0; i < 150; i++ {
					backends = append(backends, BackendAttribute{
						ServerId: "i-" + string(rune(i)),
						Port:     80,
					})
				}
				return backends
			}(),
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vg := &VServerGroup{Backends: tt.backends}
			result := vg.BackendInfo()
			assert.NotEmpty(t, result)

			if tt.wantJSON {
				var backends []BackendAttribute
				err := json.Unmarshal([]byte(result), &backends)
				assert.NoError(t, err)
				if len(tt.backends) <= 100 {
					assert.Equal(t, len(tt.backends), len(backends))
				} else {
					assert.Equal(t, 100, len(backends))
				}
			}
		})
	}
}

func TestListenerNamedKey_String(t *testing.T) {
	tests := []struct {
		name string
		key  *ListenerNamedKey
		want string
	}{
		{
			name: "nil key",
			key:  nil,
			want: "",
		},
		{
			name: "valid key with default prefix",
			key: &ListenerNamedKey{
				CID:         "c123",
				Namespace:   "default",
				ServiceName: "nginx",
				Port:        80,
			},
			want: "k8s/80/nginx/default/c123",
		},
		{
			name: "valid key with custom prefix",
			key: &ListenerNamedKey{
				Prefix:      "custom",
				CID:         "c123",
				Namespace:   "default",
				ServiceName: "nginx",
				Port:        80,
			},
			want: "custom/80/nginx/default/c123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.key.String())
		})
	}
}

func TestListenerNamedKey_Key(t *testing.T) {
	key := &ListenerNamedKey{
		CID:         "c123",
		Namespace:   "default",
		ServiceName: "nginx",
		Port:        80,
	}

	result := key.Key()
	assert.Equal(t, "k8s/80/nginx/default/c123", result)
	assert.Equal(t, DEFAULT_PREFIX, key.Prefix) // Should set default prefix
}

func TestLoadListenerNamedKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    *ListenerNamedKey
		wantErr bool
	}{
		{
			name: "valid key",
			key:  "k8s/80/nginx/default/c123",
			want: &ListenerNamedKey{
				Prefix:      DEFAULT_PREFIX,
				CID:         "c123",
				Namespace:   "default",
				ServiceName: "nginx",
				Port:        80,
			},
			wantErr: false,
		},
		{
			name:    "invalid prefix",
			key:     "custom/80/nginx/default/c123",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - too few parts",
			key:     "k8s/80/nginx",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid port",
			key:     "k8s/abc/nginx/default/c123",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadListenerNamedKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestVGroupNamedKey_String(t *testing.T) {
	tests := []struct {
		name string
		key  *VGroupNamedKey
		want string
	}{
		{
			name: "nil key",
			key:  nil,
			want: "",
		},
		{
			name: "valid key with default prefix",
			key: &VGroupNamedKey{
				CID:         "c123",
				Namespace:   "default",
				ServiceName: "nginx",
				VGroupPort:  "80",
			},
			want: "k8s/80/nginx/default/c123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.key.String())
		})
	}
}

func TestVGroupNamedKey_Key(t *testing.T) {
	key := &VGroupNamedKey{
		CID:         "c123",
		Namespace:   "default",
		ServiceName: "nginx",
		VGroupPort:  "80",
	}

	result := key.Key()
	assert.Equal(t, "k8s/80/nginx/default/c123", result)
	assert.Equal(t, DEFAULT_PREFIX, key.Prefix)
}

func TestLoadVGroupNamedKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    *VGroupNamedKey
		wantErr bool
	}{
		{
			name: "valid key",
			key:  "k8s/80/nginx/default/c123",
			want: &VGroupNamedKey{
				Prefix:      DEFAULT_PREFIX,
				CID:         "c123",
				Namespace:   "default",
				ServiceName: "nginx",
				VGroupPort:  "80",
			},
			wantErr: false,
		},
		{
			name:    "invalid prefix",
			key:     "custom/80/nginx/default/c123",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - too few parts",
			key:     "k8s/80/nginx",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadVGroupNamedKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFormatError_Error(t *testing.T) {
	err := formatError{key: "invalid/key"}
	result := err.Error()
	assert.Contains(t, result, FORMAT_ERROR)
	assert.Contains(t, result, "invalid/key")
}
