package vmock

import (
	"context"
	"strings"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewMockEFLO(
	auth *base.ClientMgr,
) *MockEFLO {
	return &MockEFLO{
		auth: auth,
	}
}

var _ prvd.IEFLO = &MockEFLO{}

type MockEFLO struct {
	auth *base.ClientMgr
}

func (m *MockEFLO) DescribeLingJunNode(ctx context.Context, id string) (*prvd.EFLONodeAttribute, error) {
	if strings.Contains(id, "notfound") {
		return nil, nil
	}
	return &prvd.EFLONodeAttribute{
		ClusterID:     "cluster-id",
		ClusterName:   "cluster-name",
		CreateTime:    "2023-01-01T00:00:00Z",
		ExpiredTime:   "2023-01-01T00:00:00Z",
		Hostname:      "hostname",
		HyperNodeID:   "hyper-node-id",
		ImageID:       "image-id",
		ImageName:     "image-name",
		MachineType:   "machine-type",
		NodeGroupID:   "node-group-id",
		NodeGroupName: "node-group-name",
		NodeID:        "node-id",
	}, nil
}
