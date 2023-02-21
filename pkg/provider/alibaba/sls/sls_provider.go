package sls

import (
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sls"
)

func NewSLSProvider(
	auth *base.ClientMgr,
) *SLSProvider {
	return &SLSProvider{auth: auth}
}

var _ prvd.ISLS = &SLSProvider{}

type SLSProvider struct {
	auth *base.ClientMgr
}

func (p SLSProvider) SLSDoAction(request requests.AcsRequest, response responses.AcsResponse) (err error) {
	return p.auth.SLS.Client.DoAction(request, response)

}
func (p SLSProvider) AnalyzeProductLog(request *sls.AnalyzeProductLogRequest) (response *sls.AnalyzeProductLogResponse, err error) {
	return p.auth.SLS.AnalyzeProductLog(request)
}
