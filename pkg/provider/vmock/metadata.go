package vmock

import (
	"fmt"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

// render meta data from cloud config file
var _ prvd.IMetaData = &fakeMetaData{}

type fakeMetaData struct {
	base prvd.IMetaData
}

func (m *fakeMetaData) HostName() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) ImageID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) InstanceID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) Mac() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) NetworkType() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) OwnerAccountID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) PrivateIPv4() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) Region() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) SerialNumber() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) SourceAddress() (string, error) {

	return "", fmt.Errorf("unimplemented")

}

func (m *fakeMetaData) VpcCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) VpcID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) VswitchCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) VswitchID() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) EIPv4() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) DNSNameServers() ([]string, error) {

	return []string{""}, fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) NTPConfigServers() ([]string, error) {

	return []string{""}, fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) Zone() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) RoleName() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) RamRoleToken(role string) (prvd.RoleAuth, error) {

	return m.base.RamRoleToken(role)
}

func (m *fakeMetaData) ClusterID() string {
	return "clusterid"
}
