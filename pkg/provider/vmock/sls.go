package vmock

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sls"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewMockSLS(
	auth *base.ClientMgr,
) *MockSLS {
	return &MockSLS{auth: auth}
}

type MockSLS struct {
	auth *base.ClientMgr
}

func (s MockSLS) SLSDoAction(request requests.AcsRequest, response responses.AcsResponse) (err error) {
	return nil
}
func (s MockSLS) AnalyzeProductLog(request *sls.AnalyzeProductLogRequest) (response *sls.AnalyzeProductLogResponse, err error) {
	return nil, nil
}
