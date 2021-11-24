package vmock

import (
	"fmt"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

var _ prvd.IMetaData = &MockMetaData{}

type MockMetaData struct {
	base  prvd.IMetaData
	vpcID string
}

func NewMockMetaData(vpcID string) prvd.IMetaData {
	return &MockMetaData{
		vpcID: vpcID,
	}
}

func (m *MockMetaData) HostName() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) ImageID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) InstanceID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) Mac() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) NetworkType() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) OwnerAccountID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) PrivateIPv4() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) Region() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) SerialNumber() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) SourceAddress() (string, error) {

	return "", fmt.Errorf("unimplemented")

}

func (m *MockMetaData) VpcCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) VpcID() (string, error) {

	return m.vpcID, nil
}

func (m *MockMetaData) VswitchCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) VswitchID() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) EIPv4() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) DNSNameServers() ([]string, error) {

	return []string{""}, fmt.Errorf("unimplemented")
}

func (m *MockMetaData) NTPConfigServers() ([]string, error) {

	return []string{""}, fmt.Errorf("unimplemented")
}

func (m *MockMetaData) Zone() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) RoleName() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *MockMetaData) RamRoleToken(role string) (prvd.RoleAuth, error) {

	return m.base.RamRoleToken(role)
}

func (m *MockMetaData) ClusterID() string {
	return "clusterid"
}
