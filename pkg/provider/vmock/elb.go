package vmock

import (
	"context"
	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

const (
	ELBHOSTName = "ELBHost"
)

const (
	MockPrefix = "mock-"
	MockIntVal = "9"

	//MockELB
	MockLoadBalancerId     = MockPrefix + "lb-id"
	MockLoadBalancerName   = MockPrefix + "lb-name"
	MockNetworkId          = MockPrefix + "network-id"
	MockVSwitchId          = MockPrefix + "vswitch-id"
	MockRegionId           = MockPrefix + "region-id"
	MockLoadBalanceSpec    = MockPrefix + "elb-spec"
	MockLoadBalancePayType = MockPrefix + "pay-type"

	//MockEIP
	MockEipId                 = MockPrefix + "eip-id"
	MockEipBandwidth          = MockIntVal
	MockEipInternetChargeType = MockPrefix + "eip-internet-chargetype"
	MockEipInstanceChargeType = MockPrefix + "eip-instance-chargetype"

	//MockListener
	MockScheduler                 = MockPrefix + "scheduler"
	MockHealthCheckConnectPort    = MockIntVal
	MockHealthyThreshold          = MockIntVal
	MockUnhealthyThreshold        = MockIntVal
	MockHealthCheckInterval       = MockIntVal
	MockHealthCheckConnectTimeout = MockIntVal
	MockHealthCheckTimeout        = MockIntVal
	MockPersistenceTimeout        = MockIntVal
	MockEstablishedTimeout        = MockIntVal

	//Mock Backend
	MockBackendLabel          = MockPrefix + "backend-label"
	MockBackendUseLbLabel     = MockPrefix + "use-lb"
	MockBackendUseLbValue     = MockPrefix + "true"
	MockBackendExcludeLbLabel = MockPrefix + "exclude-lb"
	MockBackendExcludeLbValue = MockPrefix + "true"
	MockRemoveUnscheduled     = MockPrefix + "remove-unscheduled-backend"
)

// ENS Status
const (
	MockENSInstance1 = MockPrefix + "ens-1"
	MockENSInstance2 = MockPrefix + "ens-2"
	MockENSInstance3 = MockPrefix + "ens-3"
	MockENSInstance4 = MockPrefix + "ens-4"

	ENSRunning = "Running"
	ENSStopped = "Stopped"
	ENSExpired = "Expired"
)

func NewMockELB(auth *base.ClientMgr) *MockELB {
	return &MockELB{auth: auth}
}

type MockELB struct {
	auth *base.ClientMgr
}

func (m MockELB) FindEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) CreateEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) DeleteEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) DescribeEdgeLoadBalancerById(ctx context.Context, lbId string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) DescribeEdgeLoadBalancerByName(ctx context.Context, lbName string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) SetEdgeLoadBalancerStatus(ctx context.Context, status string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) CreateEip(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) ReleaseEip(ctx context.Context, eipId string) error {
	return nil
}

func (m MockELB) AssociateElbEipAddress(ctx context.Context, eipId, slbId string) error {
	return nil
}

func (m MockELB) UnAssociateElbEipAddress(ctx context.Context, eipId string) error {
	return nil
}

func (m MockELB) ModifyEipAttribute(ctx context.Context, eipId string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) DescribeEnsEipAddressesById(ctx context.Context, eipId string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) DescribeEnsEipAddressesByName(ctx context.Context, eipName string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) FindAssociatedInstance(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) FindEdgeLoadBalancerListener(ctx context.Context, lbId string, listeners *elbmodel.EdgeListeners) error {
	return nil
}

func (m MockELB) DescribeEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (m MockELB) DescribeEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (m MockELB) DescribeEdgeLoadBalancerHTTPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (m MockELB) DescribeEdgeLoadBalancerHTTPSListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (m MockELB) StartEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, protocol string) error {
	return nil
}

func (m MockELB) StopEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, protocol string) error {
	return nil
}

func (m MockELB) ModifyEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (m MockELB) ModifyEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (m MockELB) CreateEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (m MockELB) CreateEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (m MockELB) DeleteEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, listenerProtocol string) error {
	return nil
}

func (m MockELB) AddBackendToEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	return nil
}

func (d MockELB) UpdateEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	return nil
}

func (m MockELB) RemoveBackendFromEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	return nil
}

func (m MockELB) FindBackendFromLoadBalancer(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	return nil
}

func (m MockELB) DescribeNetwork(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (m MockELB) GetEnsRegionIdByNetwork(ctx context.Context, networkId string) (string, error) {
	return MockRegionId, nil
}

func (m MockELB) FindNetWorkAndVSwitchByLoadBalancerId(ctx context.Context, lbId string) ([]string, error) {
	ret := make([]string, 0)
	ret = append(ret, MockNetworkId, MockVSwitchId)
	return ret, nil
}

func (m MockELB) FindEnsInstancesByNetwork(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) (map[string]string, error) {
	return map[string]string{
		MockENSInstance1: ENSRunning,
		MockENSInstance2: ENSRunning,
		MockENSInstance3: ENSStopped,
		MockENSInstance4: ENSExpired,
	}, nil
}
