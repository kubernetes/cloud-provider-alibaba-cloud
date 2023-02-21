package albconfigmanager

import (
	"context"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/util/sets"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
)

type CertDiscovery interface {
	Discover(ctx context.Context, tlsHosts []string) ([]string, error)
}

func NewCASCertDiscovery(cloud prvd.Provider, logger logr.Logger) *casCertDiscovery {
	return &casCertDiscovery{
		logger: logger,
		cloud:  cloud,
	}
}

var _ CertDiscovery = &casCertDiscovery{}

type casCertDiscovery struct {
	cloud  prvd.Provider
	logger logr.Logger
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

		if len(certIDsForHost) == 0 {
			return nil, errors.Errorf("none certificate found for host: %s", host)
		}
		certIDs.Insert(certIDsForHost...)
	}
	return certIDs.List(), nil
}

func (d *casCertDiscovery) loadDomainsForAllCertificates(ctx context.Context) (map[string]sets.String, error) {
	certs, err := d.cloud.DescribeSSLCertificateList(ctx)
	if err != nil {
		klog.Errorf("loadAllCertificates error: %v", err)
		return nil, err
	}
	domainsByCertID := make(map[string]sets.String, len(certs))
	for _, cert := range certs {
		domains := sets.NewString(cert.CommonName, cert.Sans)
		domainsByCertID[cert.CertIdentifier] = domains
	}
	return domainsByCertID, nil
}

func (d *casCertDiscovery) domainMatchesHost(domainName string, tlsHost string) bool {
	isMatch := false
	domains := strings.Split(domainName, ",")
	lower_host := strings.ToLower(tlsHost)
	for _, dom := range domains {
		lower_dom := strings.ToLower(dom)
		if strings.HasPrefix(lower_dom, "*.") {
			ds := strings.Split(lower_dom, ".")
			hs := strings.Split(lower_host, ".")
			if len(ds) != len(hs) {
				continue
			}

			if cmp.Equal(ds[1:], hs[1:], cmpopts.EquateEmpty()) {
				isMatch = true
				break
			}
		}
		if lower_dom == lower_host {
			isMatch = true
			break
		}
	}
	return isMatch
}
