package vmock

import (
	"context"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"strings"
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

func (d *MockECS) ListInstances(ctx context.Context, ids []string) (map[string]*prvd.NodeAttribute, error) {
	return nil, nil
}

func (d *MockECS) SetInstanceTags(ctx context.Context, id string, tags map[string]string) error {
	return nil
}

func (d *MockECS) DescribeNetworkInterfaces(vpcId string, ips []string, ipVersionType model.AddressIPVersionType) (map[string]string, error) {
	eniids := make(map[string]string)
	for _, ip := range ips {
		eniids[ip] = fmt.Sprintf("eni-id-%s", strings.Replace(ip, ".", "-", -1))
	}
	return eniids, nil
}
