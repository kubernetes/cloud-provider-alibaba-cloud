package cas

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"testing"
)

func NewCASClient() (*cas.Client, error) {
	var ak, sk, regionId string
	if ak == "" || sk == "" {
		return nil, fmt.Errorf("ak or sk is empty")
	}
	return cas.NewClientWithAccessKey(regionId, ak, sk)
}

func Test_DescribeSSLCertificateList(t *testing.T) {

	client, err := NewCASClient()
	if err != nil {
		t.Skip("fail to create cas client, skip")
		return
	}

	casProvider := NewCASProvider(&base.ClientMgr{
		CAS: client,
	})

	certs, err := casProvider.DescribeSSLCertificateList(context.TODO())
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	for _, cert := range certs {
		t.Logf("cert id: %s", cert.CertIdentifier)
	}
}

func Test_DescribeSSLCertificatePublicKeyDetail(t *testing.T) {
	client, err := NewCASClient()
	if err != nil {
		t.Skip("fail to create ecs client, skip")
		return
	}

	casProvider := NewCASProvider(&base.ClientMgr{
		CAS: client,
	})

	certId := "8xxxx-cn-hangzhou"

	certInfo, err := casProvider.DescribeSSLCertificatePublicKeyDetail(context.TODO(), certId)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	domains := sets.NewString(certInfo.CommonName, certInfo.Sans)
	t.Logf("cert domain: %+v", domains)

}
