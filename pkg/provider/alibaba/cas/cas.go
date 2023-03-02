package cas

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/cache"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/go-logr/logr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	ctrl "sigs.k8s.io/controller-runtime"
)

func NewCASProvider(
	auth *base.ClientMgr,
) *CASProvider {
	logger := ctrl.Log.WithName("controllers").WithName("CASProvider")
	return &CASProvider{
		auth:          auth,
		logger:        logger,
		loadCertMutex: &sync.Mutex{},
		certsCache:    cache.NewExpiring(),
		certsCacheTTL: 3 * time.Minute,
	}
}

var _ prvd.ICAS = &CASProvider{}

type CASProvider struct {
	auth          *base.ClientMgr
	logger        logr.Logger
	loadCertMutex *sync.Mutex
	certsCache    *cache.Expiring
	certsCacheTTL time.Duration
}

const (
	CASDomain   = "cas.aliyuncs.com"
	CASShowSize = 50
)

const (
	certsCacheKey                         = "CertificateInfo"
	DescribeSSLCertificateList            = "DescribeSSLCertificateList"
	DescribeSSLCertificatePublicKeyDetail = "DescribeSSLCertificatePublicKeyDetail"
	DefaultSSLCertificatePollInterval     = 30 * time.Second
	DefaultSSLCertificateTimeout          = 60 * time.Second
)

func (c CASProvider) casDoAction(request requests.AcsRequest, response responses.AcsResponse) (err error) {
	return c.auth.CAS.Client.DoAction(request, response)
}

func (c CASProvider) DescribeSSLCertificatePublicKeyDetail(ctx context.Context, certId string) (*model.CertificateInfo, error) {
	traceID := ctx.Value(util.TraceID)
	c.loadCertMutex.Lock()
	defer c.loadCertMutex.Unlock()

	rpcRequest := &requests.RpcRequest{}
	rpcRequest.InitWithApiInfo("cas", "2020-06-19", DescribeSSLCertificatePublicKeyDetail, "cas", "openAPI")
	rpcRequest.Method = requests.POST
	rpcRequest.Domain = CASDomain
	rpcRequest.QueryParams = map[string]string{
		"CertIdentifier": certId,
	}

	response := responses.NewCommonResponse()
	var retErr error

	if err := util.RetryImmediateOnError(DefaultSSLCertificatePollInterval, DefaultSSLCertificateTimeout, func(err error) bool {
		return false
	}, func() error {
		startTime := time.Now()
		c.logger.Info("deleting ssl certificate",
			"traceID", traceID,
			"startTime", startTime,
			"action", DescribeSSLCertificatePublicKeyDetail)
		retErr = c.casDoAction(rpcRequest, response)
		if retErr != nil {
			return retErr
		}
		if !response.IsSuccess() {
			c.logger.Error(retErr, "DescribeSSLCertificatePublicKeyDetail error")
			return retErr
		}

		c.logger.Info("got ssl certificate",
			"traceID", traceID,
			"certID", certId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			"response", response,
			"action", DescribeSSLCertificatePublicKeyDetail)
		return nil
	}); err != nil {
		return nil, errors.Wrap(retErr, "failed to describeSSLCertificatePublicKeyDetail")
	}

	resp := struct {
		CertificateInfo *model.CertificateInfo `json:"CertificateInfo"`
	}{}
	err := json.Unmarshal(response.GetHttpContentBytes(), &resp)
	if err != nil {
		return nil, err
	}

	return resp.CertificateInfo, nil
}

func (c CASProvider) DescribeSSLCertificateList(ctx context.Context) ([]model.CertificateInfo, error) {
	traceID := ctx.Value(util.TraceID)
	c.loadCertMutex.Lock()
	defer c.loadCertMutex.Unlock()

	if rawCacheItem, ok := c.certsCache.Get(certsCacheKey); ok {
		return rawCacheItem.([]model.CertificateInfo), nil
	}

	rpcRequest := &requests.RpcRequest{}
	rpcRequest.InitWithApiInfo("cas", "2020-06-19", "DescribeSSLCertificateList", "cas", "openAPI")
	rpcRequest.Method = requests.POST
	rpcRequest.Domain = CASDomain

	response := responses.NewCommonResponse()

	certificateInfos := make([]model.CertificateInfo, 0)
	pageNumber := 1
	for {
		rpcRequest.QueryParams = map[string]string{
			"ShowSize":    strconv.Itoa(CASShowSize),
			"CurrentPage": strconv.Itoa(pageNumber),
		}

		startTime := time.Now()
		c.logger.Info("listing ssl certificate",
			"traceID", traceID,
			"startTime", startTime,
			"action", DescribeSSLCertificateList)
		err := c.casDoAction(rpcRequest, response)
		if err != nil {
			c.logger.Error(err, "DescribeUserCertificateList error")
			return nil, err
		}
		c.logger.Info("listed ssl certificate",
			"traceID", traceID,
			"certMetaList", response.GetHttpContentString(),
			"elapsedTime", time.Since(startTime).Milliseconds(),
			"response", response,
			"action", DescribeSSLCertificateList)
		resp := map[string]interface{}{}
		err = json.Unmarshal(response.GetHttpContentBytes(), &resp)
		if err != nil {
			return nil, err
		}
		certBytes, _ := json.Marshal(resp["CertMetaList"])
		certs := []model.CertificateInfo{}
		err = json.Unmarshal(certBytes, &certs)
		if err != nil {
			return nil, err
		}
		certificateInfos = append(certificateInfos, certs...)
		pageCount := int(resp["PageCount"].(float64))
		if pageNumber < pageCount {
			pageNumber++
		} else {
			break
		}
	}
	c.certsCache.Set(certsCacheKey, certificateInfos, c.certsCacheTTL)
	return certificateInfos, nil
}
