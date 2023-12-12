package nlb

import (
	nlb "github.com/alibabacloud-go/nlb-20220430/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"testing"
)

func TestNLBProvider_ListNLBServerGroups(t *testing.T) {
	client, err := NewNLBClient()
	if err != nil {
		t.Skip("fail to create slb client, skip")
	}

	req := &nlb.ListServerGroupsRequest{}
	req.Tag = []*nlb.ListServerGroupsRequestTag{
		{
			Key:   tea.String("kubernetes.do.not.delete"),
			Value: tea.String("a533dfa96107c490bbeab77efbcf5698"),
		},
		{
			Key:   tea.String("ack.aliyun.com"),
			Value: tea.String("c8f23ea1a520f4cbc936cfeed3a57a8fd"),
		},
	}

	resp, err := client.ListServerGroups(req)
	if err != nil {
		t.Error(err)
	}

	var sgs []nlbmodel.ServerGroup
	for _, ret := range resp.Body.ServerGroups {
		sg := nlbmodel.ServerGroup{
			ServerGroupId:          *ret.ServerGroupId,
			ServerGroupType:        nlbmodel.ServerGroupType(*ret.ServerGroupType),
			ServerGroupName:        *ret.ServerGroupName,
			AddressIPVersion:       *ret.AddressIPVersion,
			Scheduler:              *ret.Scheduler,
			ConnectionDrainEnabled: ret.ConnectionDrainEnabled,
			ConnectionDrainTimeout: *ret.ConnectionDrainTimeout,
			ResourceGroupId:        *ret.ResourceGroupId,
		}

		serverReq := &nlb.ListServerGroupServersRequest{}
		serverReq.ServerGroupId = tea.String(sg.ServerGroupId)
		serverResp, err := client.ListServerGroupServers(serverReq)
		if err != nil {
			t.Error(err)
		}

		for _, s := range serverResp.Body.Servers {
			sg.Servers = append(sg.Servers, nlbmodel.ServerGroupServer{
				ServerGroupId: *s.ServerGroupId,
				Description:   *s.Description,
				ServerId:      *s.ServerId,
				ServerIp:      *s.ServerIp,
				ServerType:    nlbmodel.ServerType(*s.ServerType),
				Port:          *s.Port,
				Weight:        *s.Weight,
			})
		}
		sgs = append(sgs, sg)
	}

	t.Logf("%+v", sgs)
}
