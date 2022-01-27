package albconfigmanager

import (
	"context"
	"k8s.io/klog/v2"
	"strings"
	"sync"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apimachinery/pkg/util/sets"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	cassdk "github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
)

const (
	certIdentifierCacheKey             = "CertIdentifier"
	defaultCertIDsCacheTTL             = 1 * time.Minute
	defaultImportedCertDomainsCacheTTL = 5 * time.Minute
	defaultPrivateCertDomainsCacheTTL  = 10 * time.Hour
)

const (
	CASVersion  = "2021-06-19"
	CASDomain   = "cas.aliyuncs.com"
	CASShowSize = 50
)

type CertDiscovery interface {
	Discover(ctx context.Context, tlsHosts []string) ([]string, error)
}

func NewCASCertDiscovery(cloud prvd.Provider, logger logr.Logger) *casCertDiscovery {
	return &casCertDiscovery{
		logger:                      logger,
		cloud:                       cloud,
		loadDomainsByCertIDMutex:    sync.Mutex{},
		certIDsCache:                cache.NewExpiring(),
		certIDsCacheTTL:             defaultCertIDsCacheTTL,
		certDomainsCache:            cache.NewExpiring(),
		importedCertDomainsCacheTTL: defaultImportedCertDomainsCacheTTL,
		privateCertDomainsCacheTTL:  defaultPrivateCertDomainsCacheTTL,
	}
}

var _ CertDiscovery = &casCertDiscovery{}

type casCertDiscovery struct {
	cloud  prvd.Provider
	logger logr.Logger

	loadDomainsByCertIDMutex    sync.Mutex
	certIDsCache                *cache.Expiring
	certIDsCacheTTL             time.Duration
	certDomainsCache            *cache.Expiring
	importedCertDomainsCacheTTL time.Duration
	privateCertDomainsCacheTTL  time.Duration
}

func (d *casCertDiscovery) Discover(ctx context.Context, tlsHosts []string) ([]string, error) {
	domainsByCertID, err := d.loadDomainsForAllCertificates(ctx)
	if err != nil {
		klog.Errorf("loadDomainsForAllCertificates err: %v", err)
		return nil, err
	}
	certIDs := sets.NewString()
	for _, host := range tlsHosts {
		var certIDsForHost []string
		for certID, domains := range domainsByCertID {
			for domain := range domains {
				if d.domainMatchesHost(domain, host) {
					certIDsForHost = append(certIDsForHost, certID)
					break
				}
			}
		}

		if len(certIDsForHost) > 1 {
			return nil, errors.Errorf("multiple certificate found for host: %s, certIDs: %v", host, certIDsForHost)
		}
		if len(certIDsForHost) == 0 {
			return nil, errors.Errorf("none certificate found for host: %s", host)
		}
		certIDs.Insert(certIDsForHost...)
	}
	return certIDs.List(), nil
}

const (
	DescribeSSLCertificateList            = "DescribeSSLCertificateList"
	DescribeSSLCertificatePublicKeyDetail = "DescribeSSLCertificatePublicKeyDetail"
)

func (d *casCertDiscovery) loadDomainsForAllCertificates(ctx context.Context) (map[string]sets.String, error) {
	d.loadDomainsByCertIDMutex.Lock()
	defer d.loadDomainsByCertIDMutex.Unlock()

	certIDs, err := d.loadAllCertificateIDs(ctx)
	if err != nil {
		klog.Errorf("loadAllCertificateIDs error: %v", err)
		return nil, err
	}
	domainsByCertID := make(map[string]sets.String, len(certIDs))
	for _, certID := range certIDs {
		certDomains, err := d.loadDomainsForCertificate(ctx, certID)
		if err != nil {
			klog.Errorf("loadDomainsForCertificate error: %v", err)
			return nil, err
		}
		domainsByCertID[certID] = certDomains
	}
	return domainsByCertID, nil
}

func (d *casCertDiscovery) loadAllCertificateIDs(ctx context.Context) ([]string, error) {
	traceID := ctx.Value(util.TraceID)

	if rawCacheItem, ok := d.certIDsCache.Get(certIdentifierCacheKey); ok {
		return rawCacheItem.([]string), nil
	}

	req := cassdk.CreateDescribeSSLCertificateListRequest()
	req.SetVersion(CASVersion)
	req.Domain = CASDomain
	req.ShowSize = requests.NewInteger(CASShowSize)

	certificateInfos := make([]cassdk.CertificateInfo, 0)
	pageNumber := 1
	for {
		req.CurrentPage = requests.NewInteger(pageNumber)

		startTime := time.Now()
		d.logger.Info("listing ssl certificate",
			"traceID", traceID,
			"startTime", startTime,
			"action", DescribeSSLCertificateList)
		resp, err := d.cloud.DescribeSSLCertificateList(ctx, req)
		if err != nil {
			klog.Errorf("DescribeUserCertificateList error: %v", err)
			return nil, err
		}
		d.logger.Info("listed ssl certificate",
			"traceID", traceID,
			"certMetaList", resp.CertMetaList,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			"requestID", resp.RequestId,
			"action", DescribeSSLCertificateList)

		certificateInfos = append(certificateInfos, resp.CertMetaList...)

		if pageNumber < resp.PageCount {
			pageNumber++
		} else {
			break
		}
	}

	var certIDs []string
	for _, certSummary := range certificateInfos {
		certIDs = append(certIDs, certSummary.CertIdentifier)
	}

	d.certIDsCache.Set(certIdentifierCacheKey, certIDs, d.certIDsCacheTTL)

	return certIDs, nil
}

func (d *casCertDiscovery) loadDomainsForCertificate(ctx context.Context, certID string) (sets.String, error) {
	traceID := ctx.Value(util.TraceID)

	if rawCacheItem, ok := d.certDomainsCache.Get(certID); ok {
		return rawCacheItem.(sets.String), nil
	}
	req := cassdk.CreateDescribeSSLCertificatePublicKeyDetailRequest()
	req.CertIdentifier = certID
	req.SetVersion(CASVersion)
	req.Domain = CASDomain

	startTime := time.Now()
	d.logger.Info("getting ssl certificate",
		"traceID", traceID,
		"certID", certID,
		"startTime", startTime,
		"action", DescribeSSLCertificatePublicKeyDetail)
	resp, err := d.cloud.DescribeSSLCertificatePublicKeyDetail(ctx, req)
	if err != nil {
		klog.Errorf("DescribeUserCertificateDetail error: %v", err)
		return nil, err
	}
	d.logger.Info("got ssl certificate",
		"traceID", traceID,
		"certID", certID,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		"certificateInfo", resp.CertificateInfo,
		"requestID", resp.RequestId,
		"action", DescribeSSLCertificatePublicKeyDetail)

	domains := sets.NewString(resp.CertificateInfo.CommonName, resp.CertificateInfo.Sans)
	d.certDomainsCache.Set(certID, domains, d.importedCertDomainsCacheTTL)

	return domains, nil
}

func (d *casCertDiscovery) domainMatchesHost(domainName string, tlsHost string) bool {
	isMatch := false
	domains := strings.Split(domainName, ",")
	for _, dom := range domains {
		if strings.HasPrefix(dom, "*.") {
			ds := strings.Split(dom, ".")
			hs := strings.Split(tlsHost, ".")
			if len(ds) != len(hs) {
				continue
			}

			if cmp.Equal(ds[1:], hs[1:], cmpopts.EquateEmpty()) {
				isMatch = true
				break
			}
		}
		if dom == tlsHost {
			isMatch = true
			break
		}
	}
	return isMatch
}
