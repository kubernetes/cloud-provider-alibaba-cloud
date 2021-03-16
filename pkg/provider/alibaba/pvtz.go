package alibaba

import "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"

func NewPVTZProvider(
	auth *metadata.ClientAuth,
) *PVTZProvider {
	return &PVTZProvider{auth: auth}
}

type PVTZProvider struct {
	auth *metadata.ClientAuth
}

func (*PVTZProvider) CreatePVTZ() {
	panic("implement me")
}
