package dryrun

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewDryRunCloud(auth *base.ClientAuth) prvd.Provider {

	return &DryRunCloud{
		IMetaData: auth.Meta,
		MockECS:   NewMockECS(auth),
		MockCLB:   NewMockCLB(auth),
		MockPVTZ:  NewMockPVTZ(auth),
		MockVPC:   NewMockVPC(auth, nil),
	}
}

var _ prvd.Provider = alibaba.AlibabaCloud{}

// MockCloud for unit test
type DryRunCloud struct {
	*DryRunECS
	*MockPVTZ
	*MockVPC
	*MockCLB
	prvd.IMetaData
}
