package vmock

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewMockECS(
	auth *base.ClientMgr,
) *MockECS {
	return &MockECS{auth: auth}
}

type MockECS struct {
	auth *base.ClientMgr
}

var _ prvd.IInstance = &MockECS{}

const (
	ZoneID             = "cn-hangzhou-a"
	RegionID           = "cn-hangzhou"
	InstanceIP         = "192.0.168.68"
	InstanceType       = "ecs.c6.xlarge"
	NodePoolID         = "np-123456"
	InstanceChargeType = "PostPaid"
	SpotStrategy       = "SpotAsPriceGo"

	tagKeyNodePoolID = "ack.alibabacloud.com/nodepool-id"
)

func (d *MockECS) ListInstances(ctx context.Context, ids []string) (map[string]*prvd.NodeAttribute, error) {
	mins := make(map[string]*prvd.NodeAttribute)
	for _, id := range ids {
		mins[id] = &prvd.NodeAttribute{
			InstanceID:   id,
			InstanceType: InstanceType,
			Addresses: []v1.NodeAddress{
				{
					Type:    v1.NodeInternalIP,
					Address: InstanceIP,
				},
			},
			Zone:               ZoneID,
			Region:             RegionID,
			InstanceChargeType: InstanceChargeType,
			SpotStrategy:       SpotStrategy,
			Tags: map[string]string{
				tagKeyNodePoolID: NodePoolID,
			},
		}
	}
	return mins, nil
}

func (d *MockECS) GetInstancesByIP(ctx context.Context, ips []string) (*prvd.NodeAttribute, error) {
	return nil, nil
}

func (d *MockECS) DescribeNetworkInterfaces(vpcId string, ips []string, ipVersionType model.AddressIPVersionType) (map[string]string, error) {
	eniids := make(map[string]string)
	for _, ip := range ips {
		eniids[ip] = "eni-id"
	}
	return eniids, nil
}

func (d *MockECS) DescribeNetworkInterfacesByIDs(ids []string) ([]*prvd.EniAttribute, error) {
	var ret []*prvd.EniAttribute
	for _, i := range ids {
		ret = append(ret, &prvd.EniAttribute{
			NetworkInterfaceID: i,
			Status:             "Attached",
			PrivateIPAddress:   "192.168.0.1",
			SourceDestCheck:    true,
		})
	}
	return ret, nil
}

func (d *MockECS) ModifyNetworkInterfaceSourceDestCheck(id string, enabled bool) error {
	return nil
}
