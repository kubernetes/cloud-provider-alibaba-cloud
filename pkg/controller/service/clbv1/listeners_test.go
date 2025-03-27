package clbv1

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"testing"
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
