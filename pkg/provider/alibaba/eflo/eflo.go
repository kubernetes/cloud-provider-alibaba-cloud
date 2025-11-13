package eflo

import (
	"context"
	"strings"

	efloController "github.com/alibabacloud-go/eflo-controller-20221215/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/klog/v2"
)

func NewEFLOProvider(
	auth *base.ClientMgr,
) *EFLOProvider {
	return &EFLOProvider{
		auth: auth,
	}
}

type EFLOProvider struct {
	auth *base.ClientMgr
}

func (e *EFLOProvider) DescribeLingJunNode(ctx context.Context, id string) (*prvd.EFLONodeAttribute, error) {
	req := &efloController.DescribeNodeRequest{}
	req.NodeId = tea.String(id)

	resp, err := e.auth.EFLO.DescribeNode(req)
	if err != nil {
		if strings.Contains(err.Error(), "RESOURCE_NOT_FOUND") {
			return nil, nil
		}
		return nil, err
	}
	klog.V(5).Infof("RequestId: %s, API: %s, NodeId: %s", tea.StringValue(resp.Body.RequestId), "DescribeNode", id)

	return &prvd.EFLONodeAttribute{
		NodeID:          tea.StringValue(resp.Body.NodeId),
		NodeGroupName:   tea.StringValue(resp.Body.NodeGroupName),
		ClusterName:     tea.StringValue(resp.Body.ClusterName),
		ClusterID:       tea.StringValue(resp.Body.ClusterId),
		ZoneID:          tea.StringValue(resp.Body.ZoneId),
		CreateTime:      tea.StringValue(resp.Body.CreateTime),
		NodeGroupID:     tea.StringValue(resp.Body.NodeGroupId),
		Hostname:        tea.StringValue(resp.Body.Hostname),
		ImageID:         tea.StringValue(resp.Body.ImageId),
		MachineType:     tea.StringValue(resp.Body.MachineType),
		SN:              tea.StringValue(resp.Body.Sn),
		OperatingState:  tea.StringValue(resp.Body.OperatingState),
		ExpiredTime:     tea.StringValue(resp.Body.ExpiredTime),
		ImageName:       tea.StringValue(resp.Body.ImageName),
		ResourceGroupID: tea.StringValue(resp.Body.ResourceGroupId),
		NodeType:        tea.StringValue(resp.Body.NodeType),
		HyperNodeID:     tea.StringValue(resp.Body.HyperNodeId),
	}, nil
}
