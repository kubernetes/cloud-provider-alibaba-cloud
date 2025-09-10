package vmock

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
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
	for _, id := range ids {
		if strings.Contains(id, "list-instances-error") {
			return nil, fmt.Errorf("mock list instances error")
		}
	}
	mins := make(map[string]*prvd.NodeAttribute)
	for _, id := range ids {
		if strings.Contains(id, "not-exists") {
			continue
		}
		_, node, err := util.NodeFromProviderID(id)
		if err != nil {
			return nil, err
		}
		mins[id] = &prvd.NodeAttribute{
			InstanceID:   node,
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
			PrimaryNetworkInterfaceID: fmt.Sprintf("eni-%s", node),
		}
	}
	return mins, nil
}

func (d *MockECS) GetInstancesByIP(ctx context.Context, ips []string) (*prvd.NodeAttribute, error) {
	return nil, nil
}

func (d *MockECS) DescribeNetworkInterfaces(vpcId string, ips []string, ipVersionType model.AddressIPVersionType) (map[string]string, error) {
	if vpcId == "vpc-describe-eni-error" {
		return nil, fmt.Errorf("mock DescribeNetworkInterfaces error")
	}
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
	if strings.Contains(id, "error") {
		return fmt.Errorf("error")
	}
	return nil
}
