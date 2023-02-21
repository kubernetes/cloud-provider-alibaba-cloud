package dryrun

import (
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	slsprvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/sls"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sls"
)

func NewDryRunSLS(
	auth *base.ClientMgr, sls *slsprvd.SLSProvider,
) *DryRunSLS {
	return &DryRunSLS{auth: auth, sls: sls}
}

var _ prvd.ISLS = &DryRunSLS{}

type DryRunSLS struct {
	auth *base.ClientMgr
	sls  *slsprvd.SLSProvider
}

func (p DryRunSLS) SLSDoAction(request requests.AcsRequest, response responses.AcsResponse) (err error) {
	return p.auth.ALB.Client.DoAction(request, response)
}
func (s DryRunSLS) AnalyzeProductLog(request *sls.AnalyzeProductLogRequest) (response *sls.AnalyzeProductLogResponse, err error) {
	return s.auth.SLS.AnalyzeProductLog(request)
}
