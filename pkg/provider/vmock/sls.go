package vmock

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	slsprvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/sls"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sls"
)

func NewMockSLS(
	auth *base.ClientMgr,
) *MockSLS {
	return &MockSLS{auth: auth}
}

type MockSLS struct {
	auth *base.ClientMgr
	sls  *slsprvd.SLSProvider
}

func (s MockSLS) AnalyzeProductLog(request *sls.AnalyzeProductLogRequest) (response *sls.AnalyzeProductLogResponse, err error) {
	return nil, nil
}
