package alb

import (
	"context"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"

	"github.com/pkg/errors"
)

var _ core.Resource = &ServerGroup{}

type ServerGroup struct {
	core.ResourceMeta `json:"-"`

	Spec ServerGroupSpec `json:"spec"`

	Status *ServerGroupStatus `json:"status,omitempty"`
}

func NewServerGroup(stack core.Manager, id string, spec ServerGroupSpec) *ServerGroup {
	sgp := &ServerGroup{
		ResourceMeta: core.NewResourceMeta(stack, "ALIYUN::ALB::SERVERGROUP", id),
		Spec:         spec,
		Status:       nil,
	}
	stack.AddResource(sgp)
	return sgp
}

func (sgp *ServerGroup) ServerGroupID() core.StringToken {
	return core.NewResourceFieldStringToken(sgp, "status/serverGroupID",
		func(ctx context.Context, res core.Resource, fieldPath string) (s string, err error) {
			sgp := res.(*ServerGroup)
			if sgp.Status == nil {
				return "", errors.Errorf("ServerGroup is not fulfilled yet: %v", sgp.ID())
			}
			return sgp.Status.ServerGroupID, nil
		},
	)
}

type ServerGroupSpec struct {
	ServerGroupNamedKey
	ALBServerGroupSpec
}

type ServerGroupStatus struct {
	ServerGroupID string `json:"serverGroupID"`
}

func (sgp *ServerGroup) SetStatus(status ServerGroupStatus) {
	sgp.Status = &status
}
