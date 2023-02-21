package cas

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/cache"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	cassdk "github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
	CASVersion  = "2021-06-19"
	CASDomain   = "cas.aliyuncs.com"
	CASShowSize = 50
)

const (
	certsCacheKey                         = "CertificateInfo"
	DescribeSSLCertificateList            = "DescribeSSLCertificateList"
	DescribeSSLCertificatePublicKeyDetail = "DescribeSSLCertificatePublicKeyDetail"
	CreateSSLCertificateWithName          = "CreateSSLCertificateWithName"
	DeleteSSLCertificate                  = "DeleteSSLCertificate"
	DefaultSSLCertificatePollInterval     = 30 * time.Second
	DefaultSSLCertificateTimeout          = 60 * time.Second
)

func (c CASProvider) casDoAction(request requests.AcsRequest, response responses.AcsResponse) (err error) {
	return c.auth.CAS.Client.DoAction(request, response)
}

func (c CASProvider) DeleteSSLCertificate(ctx context.Context, certId string) error {
	traceID := ctx.Value(util.TraceID)
	rpcRequest := &requests.RpcRequest{}
	rpcRequest.InitWithApiInfo("cas", "2020-06-19", "DeleteSSLCertificate", "cas", "openAPI")
	rpcRequest.Method = requests.POST
	rpcRequest.Domain = CASDomain
	rpcRequest.QueryParams = map[string]string{
		"CertIdentifier": certId,
	}

	response := responses.NewCommonResponse()
	var err error
	if err := util.RetryImmediateOnError(DefaultSSLCertificatePollInterval, DefaultSSLCertificateTimeout, func(err error) bool {
		return false
	}, func() error {
		startTime := time.Now()
		c.logger.Info("deleting ssl certificate",
			"traceID", traceID,
			"startTime", startTime,
			"action", DeleteSSLCertificate)
		err = c.casDoAction(rpcRequest, response)
		if err != nil {
			return err
		}
		if !response.IsSuccess() {
			c.logger.Error(err, "DeleteSSLCertificate error")
			return err
		}
		c.logger.Info("deleted ssl certificate",
			"traceID", traceID,
			"CertIdentifier", certId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			"response", response,
			"action", DeleteSSLCertificate)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to deleteSSLCertificate")
	}
	c.loadCertMutex.Lock()
	defer c.loadCertMutex.Unlock()
	c.certsCache.Delete(certsCacheKey)
	return nil
}

func (c CASProvider) CreateSSLCertificateWithName(ctx context.Context, certName, certificate, privateKey string) (string, error) {
	traceID := ctx.Value(util.TraceID)

	rpcRequest := &requests.RpcRequest{}
	rpcRequest.InitWithApiInfo("cas", "2020-06-19", "CreateSSLCertificateWithName", "cas", "openAPI")
	rpcRequest.Method = requests.POST
	rpcRequest.Domain = CASDomain
	rpcRequest.QueryParams = map[string]string{
		"CertName":    certName,
		"PrivateKey":  privateKey,
		"Certificate": certificate,
	}
	response := responses.NewCommonResponse()
	var err error
	if err := util.RetryImmediateOnError(DefaultSSLCertificatePollInterval, DefaultSSLCertificateTimeout, func(err error) bool {
		return false
	}, func() error {
		startTime := time.Now()
		c.logger.Info("creating ssl certificate",
			"traceID", traceID,
			"startTime", startTime,
			"action", CreateSSLCertificateWithName)
		err = c.casDoAction(rpcRequest, response)
		if err != nil {
			return err
		}
		if !response.IsSuccess() {
			c.logger.Error(err, "CreateSSLCertificateWithName error")
			return err
		}
		c.logger.Info("created ssl certificate",
			"traceID", traceID,
			"certName", certName,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			"response", response,
			"action", CreateSSLCertificateWithName)
		return nil
	}); err != nil {
		return "", errors.Wrap(err, "failed to createSSLCertificateWithName")
	}
	c.loadCertMutex.Lock()
	defer c.loadCertMutex.Unlock()
	c.certsCache.Delete(certsCacheKey)
	resp := map[string]interface{}{}
	json.Unmarshal(response.GetHttpContentBytes(), &resp)
	return resp["CertIdentifier"].(string), nil
}

func (c CASProvider) DescribeSSLCertificateList(ctx context.Context) ([]model.CertificateInfo, error) {
	traceID := ctx.Value(util.TraceID)
	c.loadCertMutex.Lock()
	defer c.loadCertMutex.Unlock()

	if rawCacheItem, ok := c.certsCache.Get(certsCacheKey); ok {
		return rawCacheItem.([]model.CertificateInfo), nil
	}

	req := cassdk.CreateDescribeSSLCertificateListRequest()
	req.SetVersion(CASVersion)
	req.Domain = CASDomain
	req.ShowSize = requests.NewInteger(CASShowSize)

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
		json.Unmarshal(response.GetHttpContentBytes(), &resp)
		certBytes, _ := json.Marshal(resp["CertMetaList"])
		certs := []model.CertificateInfo{}
		json.Unmarshal(certBytes, &certs)
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
