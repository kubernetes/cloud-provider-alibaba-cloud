package albconfigmanager

import (
	"context"
	"k8s.io/klog/v2"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apimachinery/pkg/util/sets"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

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

func (d *casCertDiscovery) loadDomainsForAllCertificates(ctx context.Context) (map[string]sets.Set[string], error) {
	d.loadDomainsByCertIDMutex.Lock()
	defer d.loadDomainsByCertIDMutex.Unlock()

	certIDs, err := d.loadAllCertificateIDs(ctx)
	if err != nil {
		klog.Errorf("loadAllCertificateIDs error: %v", err)
		return nil, err
	}
	domainsByCertID := make(map[string]sets.Set[string], len(certIDs))
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

	if rawCacheItem, ok := d.certIDsCache.Get(certIdentifierCacheKey); ok {
		return rawCacheItem.([]string), nil
	}

	certs, err := d.cloud.DescribeSSLCertificateList(ctx)
	if err != nil {
		klog.Errorf("loadAllCertificates error: %v", err)
		return nil, err
	}

	var certIDs []string
	for _, certSummary := range certs {
		certIDs = append(certIDs, certSummary.CertIdentifier)
	}

	d.certIDsCache.Set(certIdentifierCacheKey, certIDs, d.certIDsCacheTTL)

	return certIDs, nil
}

func (d *casCertDiscovery) loadDomainsForCertificate(ctx context.Context, certID string) (sets.Set[string], error) {

	if rawCacheItem, ok := d.certDomainsCache.Get(certID); ok {
		return rawCacheItem.(sets.Set[string]), nil
	}

	resp, err := d.cloud.DescribeSSLCertificatePublicKeyDetail(ctx, certID)
	if err != nil {
		klog.Errorf("DescribeUserCertificateDetail error: %v", err)
		return nil, err
	}

	domains := sets.New(resp.CommonName, resp.Sans)
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
