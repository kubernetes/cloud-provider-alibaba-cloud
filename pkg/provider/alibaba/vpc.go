package alibaba

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

type AssociatedInstanceType string

const SlbInstance = AssociatedInstanceType("SlbInstance")

func NewVPCProvider(
	auth *ClientAuth,
) *VPCProvider {
	return &VPCProvider{auth: auth}
}

var _ prvd.IVPC = &VPCProvider{}

type VPCProvider struct {
	auth *ClientAuth
}

func (p *VPCProvider) CreateRoute() {
	panic("implement me")
}

func (p *VPCProvider) DeleteRoute() {
	panic("implement me")
}

func (p *VPCProvider) ListRoute() {
	panic("implement me")
}

func (p *VPCProvider) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) ([]string, error) {
	req := vpc.CreateDescribeEipAddressesRequest()
	req.AssociatedInstanceType = instanceType
	req.AssociatedInstanceId = instanceId
	var ips []string
	next := &Pagination{
		PageNumber: 1,
		PageSize:   10,
	}

	for {
		req.PageSize = requests.NewInteger(next.PageSize)
		req.PageNumber = requests.NewInteger(next.PageNumber)
		resp, err := p.auth.VPC.DescribeEipAddresses(req)
		if err != nil {
			return nil, err
		}

		for _, eip := range resp.EipAddresses.EipAddress {
			ips = append(ips, eip.IpAddress)
		}

		pageResult := &PaginationResult{
			PageNumber: resp.PageNumber,
			PageSize:   resp.PageSize,
			TotalCount: resp.TotalCount,
		}
		next := pageResult.NextPage()
		if next == nil {
			break
		}
	}
	return ips, nil
}
