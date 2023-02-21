package applier

import (
	"context"
	"sync"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"

	"github.com/go-logr/logr"
)

func NewSecretApplier(albProvider prvd.Provider, stack core.Manager, logger logr.Logger) *secretStackApplier {
	return &secretStackApplier{
		stack:       stack,
		albProvider: albProvider,
		logger:      logger,
	}
}

type secretStackApplier struct {
	albProvider prvd.Provider
	stack       core.Manager
	logger      logr.Logger
}

func (s *secretStackApplier) Apply(ctx context.Context) error {
	traceID := ctx.Value(util.TraceID)

	var resCerts []*albmodel.SecretCertificate
	_ = s.stack.ListResources(&resCerts)
	if len(resCerts) == 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize secretStack: SecretCertificate not found, skip", "traceID", traceID)
		return nil
	}
	sdkCerts, err := s.albProvider.DescribeSSLCertificateList(ctx)
	if err != nil {
		return err
	}
	matchedResAndSDKCerts, unmatchedResCerts, _ := matchResAndSDKCertificates(resCerts, sdkCerts)

	if len(matchedResAndSDKCerts) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize secretStack",
			"matchedResAndSDKCertss", matchedResAndSDKCerts,
			"traceID", traceID)
	}
	if len(unmatchedResCerts) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize secretStack",
			"unmatchedResCerts", unmatchedResCerts,
			"traceID", traceID)
	}
	var (
		errCreate error
		wgCreate  sync.WaitGroup
	)
	for _, cert := range unmatchedResCerts {
		wgCreate.Add(1)
		go func(cert *albmodel.SecretCertificate) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer wgCreate.Done()
			certId, err := s.albProvider.CreateSSLCertificateWithName(ctx, cert.Spec.CertName, cert.Spec.Certificate, cert.Spec.PrivateKey)
			if errCreate == nil && err != nil {
				errCreate = err
			}
			cert.SetStatus(albmodel.SecretCertificateStatus{
				CertIdentifier: certId,
			})
		}(cert)
	}
	wgCreate.Wait()
	if errCreate != nil {
		return errCreate
	}

	var (
		errUpdate error
		wgUpdate  sync.WaitGroup
	)
	for _, certPair := range matchedResAndSDKCerts {
		wgUpdate.Add(1)
		go func(certPair resAndSDKCertificatePair) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer wgUpdate.Done()
			certPair.ResCert.SetStatus(albmodel.SecretCertificateStatus{
				CertIdentifier: certPair.SdkCert.CertIdentifier,
			})
		}(certPair)
	}
	wgUpdate.Wait()
	if errUpdate != nil {
		return errUpdate
	}

	return nil
}

func (s *secretStackApplier) PostApply(ctx context.Context) error {
	return nil
}

type resAndSDKCertificatePair struct {
	ResCert *albmodel.SecretCertificate
	SdkCert model.CertificateInfo
}

func matchResAndSDKCertificates(resCerts []*albmodel.SecretCertificate, sdkCerts []model.CertificateInfo) ([]resAndSDKCertificatePair, []*albmodel.SecretCertificate, []model.CertificateInfo) {
	var matchedResAndSDKCerts []resAndSDKCertificatePair
	var unmatchedResCerts []*albmodel.SecretCertificate
	var unmatchedSDKCerts []model.CertificateInfo
	resCertsByName := mapResCertByName(resCerts)
	sdkCertsByName := mapSDKCertByName(sdkCerts)

	resCertNames := sets.StringKeySet(resCertsByName)
	sdkCertNames := sets.StringKeySet(sdkCertsByName)
	for _, name := range resCertNames.Intersection(sdkCertNames).List() {
		matchedResAndSDKCerts = append(matchedResAndSDKCerts, resAndSDKCertificatePair{
			ResCert: resCertsByName[name],
			SdkCert: sdkCertsByName[name],
		})
	}

	for _, name := range resCertNames.Difference(sdkCertNames).List() {
		unmatchedResCerts = append(unmatchedResCerts, resCertsByName[name])
	}

	for _, name := range sdkCertNames.Difference(resCertNames).List() {
		unmatchedSDKCerts = append(unmatchedSDKCerts, sdkCertsByName[name])
	}

	return matchedResAndSDKCerts, unmatchedResCerts, unmatchedSDKCerts
}

func mapResCertByName(resCerts []*albmodel.SecretCertificate) map[string]*albmodel.SecretCertificate {
	resCertsByName := make(map[string]*albmodel.SecretCertificate)
	for _, cert := range resCerts {
		resCertsByName[cert.Spec.CertName] = cert
	}
	return resCertsByName
}
func mapSDKCertByName(sdkCerts []model.CertificateInfo) map[string]model.CertificateInfo {
	sdkCertsByName := make(map[string]model.CertificateInfo)
	for _, cert := range sdkCerts {
		sdkCertsByName[cert.CertName] = cert
	}
	return sdkCertsByName
}
