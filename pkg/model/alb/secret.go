package alb

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

var _ core.Resource = &Listener{}

type SecretCertificate struct {
	core.ResourceMeta `json:"-"`

	Spec SecretCertificateSpec `json:"spec"`

	Status *SecretCertificateStatus `json:"status,omitempty"`
}

func NewSecretCertificate(stack core.Manager, id string, spec SecretCertificateSpec) *SecretCertificate {
	sc := &SecretCertificate{
		ResourceMeta: core.NewResourceMeta(stack, "ALIYUN::ALB::CERTIFICATE", id),
		Spec:         spec,
		Status:       nil,
	}
	_ = stack.AddResource(sc)
	return sc
}

func (sc *SecretCertificate) SetStatus(status SecretCertificateStatus) {
	sc.Status = &status
}

func (sc *SecretCertificate) CertIdentifier() core.StringToken {
	return core.NewResourceFieldStringToken(sc, "status/certIdentifier",
		func(ctx context.Context, res core.Resource, fieldPath string) (s string, err error) {
			sc := res.(*SecretCertificate)
			if sc.Status == nil {
				return "", errors.Errorf("SecretCertificate is not fulfilled yet: %v", sc.ID())
			}
			return sc.Status.CertIdentifier, nil
		},
	)
}

type SecretCertificateSpec struct {
	CertName    string `json:"certName"`
	IsDefault   bool   `json:"IsDefault" xml:"IsDefault"`
	Certificate string `json:"-"`
	PrivateKey  string `json:"-"`
}
type SecretCertificateStatus struct {
	CertIdentifier string `json:"certIdentifier"`
}

func (sc *SecretCertificate) SetDefault() {
	sc.Spec.IsDefault = true
}

func (sc *SecretCertificate) GetIsDefault() bool {
	return sc.Spec.IsDefault
}

func (sc *SecretCertificate) GetCertificateId(ctx context.Context) (string, error) {
	return sc.CertIdentifier().Resolve(ctx)
}

type FixedCertificate struct {
	IsDefault     bool   `json:"IsDefault" xml:"IsDefault"`
	CertificateId string `json:"CertificateId" xml:"CertificateId"`
	Status        string `json:"Status" xml:"Status"`
}

func (f *FixedCertificate) SetDefault() {
	f.IsDefault = true
}

func (f *FixedCertificate) GetIsDefault() bool {
	return f.IsDefault
}

func (f *FixedCertificate) GetCertificateId(ctx context.Context) (string, error) {
	return f.CertificateId, nil
}
