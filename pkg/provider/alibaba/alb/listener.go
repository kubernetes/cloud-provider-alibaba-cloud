package alb

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"

	"k8s.io/apimachinery/pkg/util/sets"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

func (m *ALBProvider) CreateALBListener(ctx context.Context, resLS *albmodel.Listener) (albmodel.ListenerStatus, error) {
	traceID := ctx.Value(util.TraceID)

	createLsReq, err := buildSDKCreateListenerRequest(resLS.Spec)
	if err != nil {
		return albmodel.ListenerStatus{}, err
	}

	var createLsResp *albsdk.CreateListenerResponse
	if err := util.RetryImmediateOnError(m.waitLSExistencePollInterval, m.waitLSExistenceTimeout, isIncorrectStatusLoadBalancerError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("creating listener",
			"stackID", resLS.Stack().StackID(),
			"resourceID", resLS.ID(),
			"traceID", traceID,
			"listenerPort", resLS.Spec.ListenerPort,
			"listenerProtocol", resLS.Spec.ListenerProtocol,
			"startTime", startTime,
			util.Action, util.CreateALBListener)
		createLsResp, err = m.auth.ALB.CreateListener(createLsReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("creating listener",
				"stackID", resLS.Stack().StackID(),
				"resourceID", resLS.ID(),
				"traceID", traceID,
				"listenerID", createLsResp.ListenerId,
				"requestID", createLsResp.RequestId,
				"error", err.Error(),
				util.Action, util.CreateALBListener)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("created listener",
			"stackID", resLS.Stack().StackID(),
			"resourceID", resLS.ID(),
			"traceID", traceID,
			"listenerID", createLsResp.ListenerId,
			"requestID", createLsResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.CreateALBListener)
		return nil
	}); err != nil {
		return albmodel.ListenerStatus{}, errors.Wrap(err, "failed to create listener")
	}

	asynchronousStartTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("creating listener asynchronous",
		"stackID", resLS.Stack().StackID(),
		"resourceID", resLS.ID(),
		"traceID", traceID,
		"listenerID", createLsResp.ListenerId,
		"startTime", asynchronousStartTime,
		util.Action, util.CreateALBListenerAsynchronous)
	var getLsResp *albsdk.GetListenerAttributeResponse
	for i := 0; i < util.CreateListenerWaitRunningMaxRetryTimes; i++ {
		time.Sleep(util.CreateListenerWaitRunningRetryInterval)

		getLsResp, err = getALBListenerAttributeFunc(ctx, createLsResp.ListenerId, m.auth, m.logger)
		if err != nil {
			return albmodel.ListenerStatus{}, err
		}
		if isListenerListenerStatusRunning(getLsResp.ListenerStatus) {
			break
		}
	}
	m.logger.V(util.MgrLogLevel).Info("created listener asynchronous",
		"stackID", resLS.Stack().StackID(),
		"resourceID", resLS.ID(),
		"traceID", traceID,
		"listenerID", createLsResp.ListenerId,
		"listenerStatus", getLsResp.ListenerStatus,
		"requestID", getLsResp.RequestId,
		"elapsedTime", time.Since(asynchronousStartTime).Milliseconds(),
		util.Action, util.CreateALBListenerAsynchronous)

	if isHTTPSListenerProtocol(resLS.Spec.ListenerProtocol) {
		if err := util.RetryImmediateOnError(m.waitLSExistencePollInterval, m.waitLSExistenceTimeout, isIncorrectStatusListenerError, func() error {
			if err := m.updateListenerExtraCertificates(ctx, createLsResp.ListenerId, resLS); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return albmodel.ListenerStatus{}, errors.Wrap(err, "failed to update listener extra certificates")
		}
	}
	return buildResListenerStatus(createLsResp.ListenerId), nil
}

func isListenerListenerStatusRunning(status string) bool {
	return strings.EqualFold(status, util.ListenerStatusRunning)
}

var getALBListenerAttributeFunc = func(ctx context.Context, lsID string, auth *base.ClientMgr, logger logr.Logger) (*albsdk.GetListenerAttributeResponse, error) {
	traceID := ctx.Value(util.TraceID)

	getLsReq := albsdk.CreateGetListenerAttributeRequest()
	getLsReq.ListenerId = lsID
	startTime := time.Now()
	logger.V(util.MgrLogLevel).Info("getting listener attribute",
		"traceID", traceID,
		"listenerID", lsID,
		"startTime", startTime,
		util.Action, util.GetALBListenerAttribute)
	getLsResp, err := auth.ALB.GetListenerAttribute(getLsReq)
	if err != nil {
		return nil, err
	}
	logger.V(util.MgrLogLevel).Info("got listener attribute",
		"traceID", traceID,
		"listenerID", lsID,
		"listenerStatus", getLsResp.ListenerStatus,
		"requestID", getLsResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.GetALBListenerAttribute)
	return getLsResp, nil
}

func (m *ALBProvider) listListenerCerts(ctx context.Context, lsID string) ([]albsdk.CertificateModel, error) {
	traceID := ctx.Value(util.TraceID)

	var (
		nextToken         string
		certificateModels []albsdk.CertificateModel
	)
	listLsCertificateReq := albsdk.CreateListListenerCertificatesRequest()
	listLsCertificateReq.ListenerId = lsID
	for {
		listLsCertificateReq.NextToken = nextToken

		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("listing listener Certificates",
			"traceID", traceID,
			"listenerID", lsID,
			"startTime", startTime,
			util.Action, util.ListALBListenerCertificates)
		listLsCertificateResp, err := m.auth.ALB.ListListenerCertificates(listLsCertificateReq)
		if err != nil {
			return nil, err
		}
		m.logger.V(util.MgrLogLevel).Info("listed listener Certificates",
			"traceID", traceID,
			"listenerID", lsID,
			"certificates", listLsCertificateResp.Certificates,
			"requestID", listLsCertificateResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListALBListenerCertificates)

		certificateModels = append(certificateModels, listLsCertificateResp.Certificates...)

		if len(listLsCertificateResp.NextToken) == 0 {
			break
		} else {
			nextToken = listLsCertificateResp.NextToken
		}
	}

	return certificateModels, nil
}

func (m *ALBProvider) UpdateALBListener(ctx context.Context, resLS *albmodel.Listener, sdkLS *albsdk.Listener) (albmodel.ListenerStatus, error) {
	if isHTTPSListenerProtocol(sdkLS.ListenerProtocol) {
		certs, err := m.listListenerCerts(ctx, sdkLS.ListenerId)
		if err != nil {
			return albmodel.ListenerStatus{}, err
		}
		m.sdkCerts = certs
	}

	if err := m.updateListenerAttribute(ctx, resLS, sdkLS); err != nil {
		return albmodel.ListenerStatus{}, err
	}

	if isHTTPSListenerProtocol(sdkLS.ListenerProtocol) {
		if err := m.updateListenerExtraCertificates(ctx, sdkLS.ListenerId, resLS); err != nil {
			return albmodel.ListenerStatus{}, err
		}
	}

	return buildResListenerStatus(sdkLS.ListenerId), nil
}

func (m *ALBProvider) DeleteALBListener(ctx context.Context, sdkLSId string) error {
	traceID := ctx.Value(util.TraceID)

	deleteLsReq := albsdk.CreateDeleteListenerRequest()
	deleteLsReq.ListenerId = sdkLSId

	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("deleting listener",
		"listenerID", sdkLSId,
		"traceID", traceID,
		"startTime", startTime,
		util.Action, util.DeleteALBListener)
	deleteLsResp, err := m.auth.ALB.DeleteListener(deleteLsReq)
	if err != nil {
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("deleted listener",
		"listenerID", sdkLSId,
		"traceID", traceID,
		"requestID", deleteLsResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.DeleteALBListener)
	return nil
}

func (m *ALBProvider) ListALBListeners(ctx context.Context, lbID string) ([]albsdk.Listener, error) {
	traceID := ctx.Value(util.TraceID)

	if len(lbID) == 0 {
		return nil, fmt.Errorf("invalid load balancer id: %s for listing listeners", lbID)
	}

	var (
		nextToken string
		listeners []albsdk.Listener
	)

	listLsReq := albsdk.CreateListListenersRequest()
	listLsReq.LoadBalancerIds = &[]string{lbID}

	for {
		listLsReq.NextToken = nextToken

		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("listing listeners",
			"loadBalancerID", lbID,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.ListALBListeners)
		listLsResp, err := m.auth.ALB.ListListeners(listLsReq)
		if err != nil {
			return nil, err
		}
		m.logger.V(util.MgrLogLevel).Info("listed listeners",
			"loadBalancerID", lbID,
			"traceID", traceID,
			"requestID", listLsResp.RequestId,
			"listeners", listLsResp.Listeners,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListALBListeners)

		listeners = append(listeners, listLsResp.Listeners...)

		if len(listLsResp.NextToken) == 0 {
			break
		} else {
			nextToken = listLsResp.NextToken
		}
	}

	return listeners, nil
}

func transSDKCertificateToAssociate(certs []albsdk.Certificate) *[]albsdk.AssociateAdditionalCertificatesWithListenerCertificates {
	associateCerts := make([]albsdk.AssociateAdditionalCertificatesWithListenerCertificates, 0)
	for _, cert := range certs {
		associateCerts = append(associateCerts, albsdk.AssociateAdditionalCertificatesWithListenerCertificates{
			CertificateId: cert.CertificateId,
		})
	}
	return &associateCerts
}

func (m *ALBProvider) AssociateALBAdditionalCertificatesWithListener(lsID string, certs []albsdk.Certificate) (*albsdk.AssociateAdditionalCertificatesWithListenerResponse, error) {
	lsReq := albsdk.CreateAssociateAdditionalCertificatesWithListenerRequest()
	lsReq.ListenerId = lsID
	lsReq.Certificates = transSDKCertificateToAssociate(certs)
	resp, err := m.auth.ALB.AssociateAdditionalCertificatesWithListener(lsReq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func transSDKCertificateToDissociate(certs []albsdk.Certificate) *[]albsdk.DissociateAdditionalCertificatesFromListenerCertificates {
	dissociateCerts := make([]albsdk.DissociateAdditionalCertificatesFromListenerCertificates, 0)
	for _, cert := range certs {
		dissociateCerts = append(dissociateCerts, albsdk.DissociateAdditionalCertificatesFromListenerCertificates{
			CertificateId: cert.CertificateId,
		})
	}
	return &dissociateCerts
}

func (m *ALBProvider) DissociateALBAdditionalCertificatesFromListener(lsID string, certs []albsdk.Certificate) (*albsdk.DissociateAdditionalCertificatesFromListenerResponse, error) {
	lsReq := albsdk.CreateDissociateAdditionalCertificatesFromListenerRequest()
	lsReq.ListenerId = lsID
	lsReq.Certificates = transSDKCertificateToDissociate(certs)
	lsResp, err := m.auth.ALB.DissociateAdditionalCertificatesFromListener(lsReq)
	if err != nil {
		return nil, err
	}
	return lsResp, nil
}

func buildSDKCreateListenerRequest(lsSpec albmodel.ListenerSpec) (*albsdk.CreateListenerRequest, error) {
	ctx := context.Background()
	lbID, err := lsSpec.LoadBalancerID.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	createLsReq := albsdk.CreateCreateListenerRequest()

	createLsReq.LoadBalancerId = lbID

	if !isListenerProtocolValid(lsSpec.ListenerProtocol) {
		return nil, fmt.Errorf("invalid listener protocol: %s", lsSpec.ListenerProtocol)
	}
	createLsReq.ListenerProtocol = lsSpec.ListenerProtocol

	if !isListenerPortValid(lsSpec.ListenerPort) {
		return nil, fmt.Errorf("invalid listener port: %d", lsSpec.ListenerPort)
	}
	createLsReq.ListenerPort = requests.NewInteger(lsSpec.ListenerPort)

	createLsReq.ListenerDescription = lsSpec.ListenerDescription

	if !isListenerRequestTimeoutValid(lsSpec.RequestTimeout) {
		return nil, fmt.Errorf("invalid listener RequestTimeout: %d", lsSpec.RequestTimeout)
	}
	createLsReq.RequestTimeout = requests.NewInteger(lsSpec.RequestTimeout)

	if !isListenerIdleTimeoutValid(lsSpec.IdleTimeout) {
		return nil, fmt.Errorf("invalid listener IdleTimeout: %d", lsSpec.IdleTimeout)
	}
	createLsReq.IdleTimeout = requests.NewInteger(lsSpec.IdleTimeout)

	createLsReq.GzipEnabled = requests.NewBoolean(lsSpec.GzipEnabled)
	createLsReq.QuicConfig = transSDKQuicConfigToCreateLs(lsSpec.QuicConfig)

	if len(lsSpec.DefaultActions) == 0 {
		return nil, fmt.Errorf("empty listener default actions: %v", lsSpec.DefaultActions)
	}
	defaultActions, err := transModelActionToSDKCreateLs(lsSpec.DefaultActions)
	if err != nil {
		return nil, err
	}
	if len(*defaultActions) == 0 {
		return nil, fmt.Errorf("empty listener default actions: %v", *defaultActions)
	}
	createLsReq.DefaultActions = defaultActions

	createLsReq.XForwardedForConfig = transSDKXForwardedForConfigToCreateLs(lsSpec.XForwardedForConfig)

	if isHTTPSListenerProtocol(lsSpec.ListenerProtocol) {
		if len(lsSpec.SecurityPolicyId) == 0 {
			return nil, fmt.Errorf("invalid https listener SecurityPolicyId: %s", lsSpec.SecurityPolicyId)
		}
		createLsReq.SecurityPolicyId = lsSpec.SecurityPolicyId

		createLsReq.CaCertificates = transSDKCaCertificatesToCreateLs(lsSpec.CaCertificates)

		if len(lsSpec.Certificates) == 0 {
			return nil, fmt.Errorf("empty https listener default certs ")
		}
		defaultCerts, _ := buildSDKCertificates(lsSpec.Certificates)
		if len(defaultCerts) != 1 {
			return nil, fmt.Errorf("empty https listener default certs")
		}
		createLsReq.Certificates = transSDKCertificatesToCreateLs(defaultCerts)

		createLsReq.Http2Enabled = requests.NewBoolean(lsSpec.Http2Enabled)
	}

	return createLsReq, nil
}

func (m *ALBProvider) updateListenerExtraCertificates(ctx context.Context, lsID string, resLs *albmodel.Listener) error {
	traceID := ctx.Value(util.TraceID)

	desiredExtraCertIDs := sets.NewString()
	_, desiredExtraCerts := buildSDKCertificates(resLs.Spec.Certificates)
	for _, cert := range desiredExtraCerts {
		desiredExtraCertIDs.Insert(cert.CertificateId)
	}

	currentExtraCertIDs := sets.NewString()
	_, currentExtraCerts := buildSDKCertificatesModel(m.sdkCerts)
	for _, cert := range currentExtraCerts {
		currentExtraCertIDs.Insert(cert.CertificateId)
	}

	unmatchedResCerts := desiredExtraCertIDs.Difference(currentExtraCertIDs).List()
	if len(unmatchedResCerts) != 0 {
		certs := make([]albsdk.Certificate, 0)
		for _, unmatchedResCert := range unmatchedResCerts {
			certs = append(certs, albsdk.Certificate{
				CertificateId: unmatchedResCert,
			})
		}
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("associating additional certificates to listener",
			"stackID", resLs.Stack().StackID(),
			"resourceID", resLs.ID(),
			"listenerID", lsID,
			"traceID", traceID,
			"certificates", certs,
			"startTime", startTime,
			util.Action, util.AssociateALBAdditionalCertificatesWithListener)
		resp, err := m.AssociateALBAdditionalCertificatesWithListener(lsID, certs)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("associating additional certificates to listener",
				"stackID", resLs.Stack().StackID(),
				"resourceID", resLs.ID(),
				"listenerID", lsID,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.AssociateALBAdditionalCertificatesWithListener)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("associated additional certificates to listener",
			"stackID", resLs.Stack().StackID(),
			"resourceID", resLs.ID(),
			"listenerID", lsID,
			"traceID", traceID,
			"certificates", certs,
			"requestID", resp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.AssociateALBAdditionalCertificatesWithListener)
	}

	unmatchedSDKCerts := currentExtraCertIDs.Difference(desiredExtraCertIDs).List()
	if len(unmatchedSDKCerts) != 0 {
		certs := make([]albsdk.Certificate, 0)
		for _, unmatchedSDKCert := range unmatchedSDKCerts {
			certs = append(certs, albsdk.Certificate{
				CertificateId: unmatchedSDKCert,
			})
		}
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("dissociating additional certificates from listener",
			"stackID", resLs.Stack().StackID(),
			"resourceID", resLs.ID(),
			"listenerID", lsID,
			"traceID", traceID,
			"certificates", certs,
			"startTime", startTime,
			util.Action, util.DissociateALBAdditionalCertificatesFromListener)
		resp, err := m.DissociateALBAdditionalCertificatesFromListener(lsID, certs)
		if err != nil {
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("dissociated additional certificates from listener",
			"stackID", resLs.Stack().StackID(),
			"resourceID", resLs.ID(),
			"listenerID", lsID,
			"traceID", traceID,
			"certificates", certs,
			"requestID", resp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.DissociateALBAdditionalCertificatesFromListener)
	}
	return nil
}

func (m *ALBProvider) updateListenerAttribute(ctx context.Context, resLS *albmodel.Listener, sdkLs *albsdk.Listener) error {
	traceID := ctx.Value(util.TraceID)

	var (
		isGzipEnabledNeedUpdate,
		isQuicConfigUpdate,
		isHttp2EnabledNeedUpdate,
		isDefaultActionsNeedUpdate,
		isRequestTimeoutNeedUpdate,
		isXForwardedForConfigNeedUpdate,
		isSecurityPolicyIdNeedUpdate,
		isIdleTimeoutNeedUpdate,
		isListenerDescriptionNeedUpdate,
		isCertificatesNeedUpdate bool
	)

	if resLS.Spec.GzipEnabled != sdkLs.GzipEnabled {
		m.logger.V(util.MgrLogLevel).Info("GzipEnabled update",
			"res", resLS.Spec.GzipEnabled,
			"sdk", sdkLs.GzipEnabled,
			"listenerID", sdkLs.ListenerId,
			"traceID", traceID)
		isGzipEnabledNeedUpdate = true
	}
	quicConfig := transQuicConfigToSDK(resLS.Spec.QuicConfig)
	if quicConfig != sdkLs.QuicConfig {
		m.logger.V(util.MgrLogLevel).Info("QuicConfig update",
			"res", resLS.Spec.QuicConfig,
			"sdk", sdkLs.QuicConfig,
			"listenerID", sdkLs.ListenerId,
			"traceID", traceID)
		isQuicConfigUpdate = true
	}
	if len(resLS.Spec.DefaultActions) == 0 {
		return fmt.Errorf("empty listener default action: %v", resLS.Spec.DefaultActions)
	}
	resAction, err := transModelActionsToSDKLs(resLS.Spec.DefaultActions)
	if err != nil {
		return err
	}
	if len(*resAction) == 0 {
		return fmt.Errorf("empty listener default action: %v", *resAction)
	}
	if !reflect.DeepEqual((*resAction)[0], sdkLs.DefaultActions[0]) {
		m.logger.V(util.MgrLogLevel).Info("DefaultActions update",
			"res", (*resAction)[0],
			"sdk", sdkLs.DefaultActions[0],
			"listenerID", sdkLs.ListenerId,
			"traceID", traceID)
		isDefaultActionsNeedUpdate = true
	}

	if !isListenerRequestTimeoutValid(resLS.Spec.RequestTimeout) {
		return fmt.Errorf("invalid listener RequestTimeout: %d", resLS.Spec.RequestTimeout)
	}
	if resLS.Spec.RequestTimeout != sdkLs.RequestTimeout {
		m.logger.V(util.MgrLogLevel).Info("RequestTimeout update",
			"res", resLS.Spec.RequestTimeout,
			"sdk", sdkLs.RequestTimeout,
			"listenerID", sdkLs.ListenerId,
			"traceID", traceID)
		isRequestTimeoutNeedUpdate = true
	}
	xForwardedForConfig := transXForwardedForConfigToSDK(resLS.Spec.XForwardedForConfig)
	if xForwardedForConfig != sdkLs.XForwardedForConfig {
		m.logger.V(util.MgrLogLevel).Info("XForwardedForConfig update",
			"res", resLS.Spec.XForwardedForConfig,
			"sdk", sdkLs.XForwardedForConfig,
			"listenerID", sdkLs.ListenerId,
			"traceID", traceID)
		isXForwardedForConfigNeedUpdate = true
	}

	if !isListenerIdleTimeoutValid(resLS.Spec.IdleTimeout) {
		return fmt.Errorf("invalid listener IdleTimeout: %d", resLS.Spec.IdleTimeout)
	}
	if resLS.Spec.IdleTimeout != sdkLs.IdleTimeout {
		m.logger.V(util.MgrLogLevel).Info("IdleTimeout update",
			"res", resLS.Spec.IdleTimeout,
			"sdk", sdkLs.IdleTimeout,
			"listenerID", sdkLs.ListenerId,
			"traceID", traceID)
		isIdleTimeoutNeedUpdate = true
	}
	if resLS.Spec.ListenerDescription != sdkLs.ListenerDescription {
		m.logger.V(util.MgrLogLevel).Info("ListenerDescription update",
			"res", resLS.Spec.ListenerDescription,
			"sdk", sdkLs.ListenerDescription,
			"listenerID", sdkLs.ListenerId,
			"traceID", traceID)
		isListenerDescriptionNeedUpdate = true
	}

	if isHTTPSListenerProtocol(sdkLs.ListenerProtocol) {
		if len(resLS.Spec.SecurityPolicyId) == 0 {
			return fmt.Errorf("invalid https listener SecurityPolicyId: %s", resLS.Spec.SecurityPolicyId)
		}
		if resLS.Spec.SecurityPolicyId != sdkLs.SecurityPolicyId {
			m.logger.V(util.MgrLogLevel).Info("SecurityPolicyId update",
				"res", resLS.Spec.SecurityPolicyId,
				"sdk", sdkLs.SecurityPolicyId,
				"listenerID", sdkLs.ListenerId,
				"traceID", traceID)
			isSecurityPolicyIdNeedUpdate = true
		}

		desiredDefaultCerts, _ := buildSDKCertificates(resLS.Spec.Certificates)
		if len(desiredDefaultCerts) != 1 {
			return fmt.Errorf("invalid res https listener default certs len: %d", len(desiredDefaultCerts))
		}
		currentDefaultCerts, _ := buildSDKCertificatesModel(m.sdkCerts)
		if len(currentDefaultCerts) != 1 {
			return fmt.Errorf("invalid sdk https listener default certs len: %d", len(currentDefaultCerts))
		}
		if desiredDefaultCerts[0].CertificateId != currentDefaultCerts[0].CertificateId {
			m.logger.V(util.MgrLogLevel).Info("DefaultCerts update",
				"res", desiredDefaultCerts[0],
				"sdk", currentDefaultCerts[0],
				"listenerID", sdkLs.ListenerId,
				"traceID", traceID)
			isCertificatesNeedUpdate = true
		}

		if resLS.Spec.Http2Enabled != sdkLs.Http2Enabled {
			m.logger.V(util.MgrLogLevel).Info("Http2Enabled update",
				"res", resLS.Spec.Http2Enabled,
				"sdk", sdkLs.Http2Enabled,
				"listenerID", sdkLs.ListenerId,
				"traceID", traceID)
			isHttp2EnabledNeedUpdate = true
		}
	}

	if !isGzipEnabledNeedUpdate && !isQuicConfigUpdate && !isHttp2EnabledNeedUpdate &&
		!isDefaultActionsNeedUpdate && !isRequestTimeoutNeedUpdate && !isXForwardedForConfigNeedUpdate &&
		!isSecurityPolicyIdNeedUpdate && !isIdleTimeoutNeedUpdate && !isListenerDescriptionNeedUpdate &&
		!isCertificatesNeedUpdate {
		return nil
	}

	updateLsReq := albsdk.CreateUpdateListenerAttributeRequest()
	updateLsReq.ListenerId = sdkLs.ListenerId

	if isGzipEnabledNeedUpdate {
		updateLsReq.GzipEnabled = requests.NewBoolean(resLS.Spec.GzipEnabled)
	}
	if isQuicConfigUpdate {
		updateLsReq.QuicConfig = transSDKQuicConfigToUpdateLs(resLS.Spec.QuicConfig)
	}
	if isHttp2EnabledNeedUpdate {
		updateLsReq.Http2Enabled = requests.NewBoolean(resLS.Spec.Http2Enabled)
	}
	if isDefaultActionsNeedUpdate {
		defaultAction, err := transModelActionToSDKUpdateLs(resLS.Spec.DefaultActions)
		if err != nil {
			return err
		}
		updateLsReq.DefaultActions = defaultAction
	}
	if isRequestTimeoutNeedUpdate {
		updateLsReq.RequestTimeout = requests.NewInteger(resLS.Spec.RequestTimeout)
	}
	if isXForwardedForConfigNeedUpdate {
		updateLsReq.XForwardedForConfig = transSDKXForwardedForConfigToUpdateLs(resLS.Spec.XForwardedForConfig)
	}
	if isSecurityPolicyIdNeedUpdate {
		updateLsReq.SecurityPolicyId = resLS.Spec.SecurityPolicyId
	}
	if isIdleTimeoutNeedUpdate {
		updateLsReq.IdleTimeout = requests.NewInteger(resLS.Spec.IdleTimeout)
	}
	if isListenerDescriptionNeedUpdate {
		updateLsReq.ListenerDescription = resLS.Spec.ListenerDescription
	}
	if isCertificatesNeedUpdate {
		updateLsReq.Certificates = transSDKCertificatesToUpdateLs(resLS.Spec.Certificates)
	}

	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("updating listener attribute",
		"stackID", resLS.Stack().StackID(),
		"resourceID", resLS.ID(),
		"traceID", traceID,
		"listenerID", sdkLs.ListenerId,
		"startTime", startTime,
		util.Action, util.UpdateALBListenerAttribute)
	updateLsResp, err := m.auth.ALB.UpdateListenerAttribute(updateLsReq)
	if err != nil {
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("updated listener attribute",
		"stackID", resLS.Stack().StackID(),
		"resourceID", resLS.ID(),
		"traceID", traceID,
		"listenerID", sdkLs.ListenerId,
		"requestID", updateLsResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.UpdateALBListenerAttribute)

	return nil
}

func transModelActionToSDKCreateLs(actions []albmodel.Action) (*[]albsdk.CreateListenerDefaultActions, error) {
	createLsActions := make([]albsdk.CreateListenerDefaultActions, 0)
	for _, action := range actions {
		sgTuples, err := transModelServerGroupTupleToSDKCreateLs(action.ForwardConfig.ServerGroups)
		if err != nil {
			return nil, err
		}
		createLsAction := albsdk.CreateListenerDefaultActions{
			ForwardGroupConfig: albsdk.CreateListenerDefaultActionsForwardGroupConfig{
				ServerGroupTuples: sgTuples,
			},
			Type: action.Type,
		}
		createLsActions = append(createLsActions, createLsAction)
	}
	return &createLsActions, nil
}

func transModelActionToSDKUpdateLs(actions []albmodel.Action) (*[]albsdk.UpdateListenerAttributeDefaultActions, error) {
	updateLsActions := make([]albsdk.UpdateListenerAttributeDefaultActions, 0)

	for _, action := range actions {
		sgTuples, err := transModelServerGroupTupleToSDKUpdateLs(action.ForwardConfig.ServerGroups)
		if err != nil {
			return nil, err
		}
		updateLsAction := albsdk.UpdateListenerAttributeDefaultActions{
			ForwardGroupConfig: albsdk.UpdateListenerAttributeDefaultActionsForwardGroupConfig{
				ServerGroupTuples: sgTuples,
			},
			Type: action.Type,
		}
		updateLsActions = append(updateLsActions, updateLsAction)
	}
	return &updateLsActions, nil
}

func transModelServerGroupTuplesToSDKLs(ms []albmodel.ServerGroupTuple) (*[]albsdk.ServerGroupTuple, error) {
	var updateRuleServerGroupTuples []albsdk.ServerGroupTuple
	if len(ms) != 0 {
		updateRuleServerGroupTuples = make([]albsdk.ServerGroupTuple, 0)
		for _, m := range ms {
			serverGroupID, err := m.ServerGroupID.Resolve(context.Background())
			if err != nil {
				return nil, err
			}
			updateRuleServerGroupTuples = append(updateRuleServerGroupTuples, albsdk.ServerGroupTuple{
				ServerGroupId: serverGroupID,
			})
		}
	}
	return &updateRuleServerGroupTuples, nil
}

func transModelActionsToSDKLs(actions []albmodel.Action) (*[]albsdk.DefaultAction, error) {
	lsActions := make([]albsdk.DefaultAction, 0)

	for _, action := range actions {
		sgTuples, err := transModelServerGroupTuplesToSDKLs(action.ForwardConfig.ServerGroups)
		if err != nil {
			return nil, err
		}
		updateLsAction := albsdk.DefaultAction{
			ForwardGroupConfig: albsdk.ForwardGroupConfig{
				ServerGroupTuples: *sgTuples,
			},
			Type: action.Type,
		}
		lsActions = append(lsActions, updateLsAction)
	}
	return &lsActions, nil
}

func transModelServerGroupTupleToSDKCreateLs(serverGroupTuples []albmodel.ServerGroupTuple) (*[]albsdk.CreateListenerDefaultActionsForwardGroupConfigServerGroupTuplesItem, error) {
	createLsServerGroupTuples := make([]albsdk.CreateListenerDefaultActionsForwardGroupConfigServerGroupTuplesItem, 0)

	for _, serverGroupTuple := range serverGroupTuples {
		serverGroupID, err := serverGroupTuple.ServerGroupID.Resolve(context.Background())
		if err != nil {
			return nil, err
		}
		createServerGroupTuple := albsdk.CreateListenerDefaultActionsForwardGroupConfigServerGroupTuplesItem{
			ServerGroupId: serverGroupID,
		}
		createLsServerGroupTuples = append(createLsServerGroupTuples, createServerGroupTuple)
	}
	return &createLsServerGroupTuples, nil
}

func transModelServerGroupTupleToSDKUpdateLs(serverGroupTuples []albmodel.ServerGroupTuple) (*[]albsdk.UpdateListenerAttributeDefaultActionsForwardGroupConfigServerGroupTuplesItem, error) {
	updateLsServerGroupTuples := make([]albsdk.UpdateListenerAttributeDefaultActionsForwardGroupConfigServerGroupTuplesItem, 0)

	for _, serverGroupTuple := range serverGroupTuples {
		serverGroupID, err := serverGroupTuple.ServerGroupID.Resolve(context.Background())
		if err != nil {
			return nil, err
		}
		createServerGroupTuple := albsdk.UpdateListenerAttributeDefaultActionsForwardGroupConfigServerGroupTuplesItem{
			ServerGroupId: serverGroupID,
		}
		updateLsServerGroupTuples = append(updateLsServerGroupTuples, createServerGroupTuple)
	}
	return &updateLsServerGroupTuples, nil
}

func transSDKQuicConfigToCreateLs(c albmodel.QuicConfig) albsdk.CreateListenerQuicConfig {
	return albsdk.CreateListenerQuicConfig{
		QuicUpgradeEnabled: strconv.FormatBool(c.QuicUpgradeEnabled),
		QuicListenerId:     c.QuicListenerId,
	}
}
func transQuicConfigToSDK(c albmodel.QuicConfig) albsdk.QuicConfig {
	return albsdk.QuicConfig{
		QuicUpgradeEnabled: c.QuicUpgradeEnabled,
		QuicListenerId:     c.QuicListenerId,
	}
}
func transSDKQuicConfigToUpdateLs(c albmodel.QuicConfig) albsdk.UpdateListenerAttributeQuicConfig {
	return albsdk.UpdateListenerAttributeQuicConfig{
		QuicUpgradeEnabled: strconv.FormatBool(c.QuicUpgradeEnabled),
		QuicListenerId:     c.QuicListenerId,
	}
}

func transSDKCaCertificatesToCreateLs(certificates []albmodel.Certificate) *[]albsdk.CreateListenerCaCertificates {
	createListenerAttributeCaCertificates := make([]albsdk.CreateListenerCaCertificates, 0)
	for _, certificate := range certificates {
		createListenerAttributeCaCertificates = append(createListenerAttributeCaCertificates, albsdk.CreateListenerCaCertificates{
			CertificateId: certificate.CertificateId,
		})
	}
	return &createListenerAttributeCaCertificates
}

func transSDKCertificatesToCreateLs(certificates []albsdk.Certificate) *[]albsdk.CreateListenerCertificates {
	createListenerAttributeCertificates := make([]albsdk.CreateListenerCertificates, 0)
	for _, certificate := range certificates {
		createListenerAttributeCertificates = append(createListenerAttributeCertificates, albsdk.CreateListenerCertificates{
			CertificateId: certificate.CertificateId,
		})
	}
	return &createListenerAttributeCertificates
}

func transSDKCertificatesToUpdateLs(certificates []albmodel.Certificate) *[]albsdk.UpdateListenerAttributeCertificates {
	updateListenerAttributeCertificates := make([]albsdk.UpdateListenerAttributeCertificates, 0)
	for _, certificate := range certificates {
		updateListenerAttributeCertificates = append(updateListenerAttributeCertificates, albsdk.UpdateListenerAttributeCertificates{
			CertificateId: certificate.CertificateId,
		})
	}
	return &updateListenerAttributeCertificates
}

func transSDKXForwardedForConfigToCreateLs(c albmodel.XForwardedForConfig) albsdk.CreateListenerXForwardedForConfig {
	return albsdk.CreateListenerXForwardedForConfig{
		XForwardedForClientCertSubjectDNAlias:      c.XForwardedForClientCertIssuerDNAlias,
		XForwardedForClientCertIssuerDNEnabled:     strconv.FormatBool(c.XForwardedForClientCertIssuerDNEnabled),
		XForwardedForClientCertFingerprintEnabled:  strconv.FormatBool(c.XForwardedForClientCertFingerprintEnabled),
		XForwardedForClientCertIssuerDNAlias:       c.XForwardedForClientCertIssuerDNAlias,
		XForwardedForProtoEnabled:                  strconv.FormatBool(c.XForwardedForProtoEnabled),
		XForwardedForClientCertFingerprintAlias:    c.XForwardedForClientCertFingerprintAlias,
		XForwardedForClientCertClientVerifyEnabled: strconv.FormatBool(c.XForwardedForClientCertClientVerifyEnabled),
		XForwardedForSLBPortEnabled:                strconv.FormatBool(c.XForwardedForSLBPortEnabled),
		XForwardedForClientCertSubjectDNEnabled:    strconv.FormatBool(c.XForwardedForClientCertSubjectDNEnabled),
		XForwardedForClientCertClientVerifyAlias:   c.XForwardedForClientCertClientVerifyAlias,
		XForwardedForClientSrcPortEnabled:          strconv.FormatBool(c.XForwardedForClientSrcPortEnabled),
		XForwardedForEnabled:                       strconv.FormatBool(c.XForwardedForEnabled),
		XForwardedForSLBIdEnabled:                  strconv.FormatBool(c.XForwardedForSLBIdEnabled),
	}
}

func transXForwardedForConfigToSDK(c albmodel.XForwardedForConfig) albsdk.XForwardedForConfig {
	return albsdk.XForwardedForConfig{
		XForwardedForClientCertSubjectDNAlias:      c.XForwardedForClientCertIssuerDNAlias,
		XForwardedForClientCertIssuerDNEnabled:     c.XForwardedForClientCertIssuerDNEnabled,
		XForwardedForClientCertFingerprintEnabled:  c.XForwardedForClientCertFingerprintEnabled,
		XForwardedForClientCertIssuerDNAlias:       c.XForwardedForClientCertIssuerDNAlias,
		XForwardedForProtoEnabled:                  c.XForwardedForProtoEnabled,
		XForwardedForClientCertFingerprintAlias:    c.XForwardedForClientCertFingerprintAlias,
		XForwardedForClientCertClientVerifyEnabled: c.XForwardedForClientCertClientVerifyEnabled,
		XForwardedForSLBPortEnabled:                c.XForwardedForSLBPortEnabled,
		XForwardedForClientCertSubjectDNEnabled:    c.XForwardedForClientCertSubjectDNEnabled,
		XForwardedForClientCertClientVerifyAlias:   c.XForwardedForClientCertClientVerifyAlias,
		XForwardedForClientSrcPortEnabled:          c.XForwardedForClientSrcPortEnabled,
		XForwardedForEnabled:                       c.XForwardedForEnabled,
		XForwardedForSLBIdEnabled:                  c.XForwardedForSLBIdEnabled,
	}
}

func transSDKXForwardedForConfigToUpdateLs(c albmodel.XForwardedForConfig) albsdk.UpdateListenerAttributeXForwardedForConfig {
	return albsdk.UpdateListenerAttributeXForwardedForConfig{
		XForwardedForClientCertSubjectDNAlias:      c.XForwardedForClientCertIssuerDNAlias,
		XForwardedForClientCertIssuerDNEnabled:     strconv.FormatBool(c.XForwardedForClientCertIssuerDNEnabled),
		XForwardedForClientCertFingerprintEnabled:  strconv.FormatBool(c.XForwardedForClientCertFingerprintEnabled),
		XForwardedForClientCertIssuerDNAlias:       c.XForwardedForClientCertIssuerDNAlias,
		XForwardedForProtoEnabled:                  strconv.FormatBool(c.XForwardedForProtoEnabled),
		XForwardedForClientCertFingerprintAlias:    c.XForwardedForClientCertFingerprintAlias,
		XForwardedForClientCertClientVerifyEnabled: strconv.FormatBool(c.XForwardedForClientCertClientVerifyEnabled),
		XForwardedForSLBPortEnabled:                strconv.FormatBool(c.XForwardedForSLBPortEnabled),
		XForwardedForClientCertSubjectDNEnabled:    strconv.FormatBool(c.XForwardedForClientCertSubjectDNEnabled),
		XForwardedForClientCertClientVerifyAlias:   c.XForwardedForClientCertClientVerifyAlias,
		XForwardedForClientSrcPortEnabled:          strconv.FormatBool(c.XForwardedForClientSrcPortEnabled),
		XForwardedForEnabled:                       strconv.FormatBool(c.XForwardedForEnabled),
		XForwardedForSLBIdEnabled:                  strconv.FormatBool(c.XForwardedForSLBIdEnabled),
	}
}

func buildSDKCertificatesModel(modelCerts []albsdk.CertificateModel) ([]albsdk.CertificateModel, []albsdk.CertificateModel) {
	if len(modelCerts) == 0 {
		return nil, nil
	}

	var defaultSDKCerts []albsdk.CertificateModel
	var extraSDKCerts []albsdk.CertificateModel
	for _, cert := range modelCerts {
		if cert.IsDefault {
			defaultSDKCerts = append(defaultSDKCerts, cert)
		} else {
			extraSDKCerts = append(extraSDKCerts, cert)
		}
	}
	return defaultSDKCerts, extraSDKCerts
}

func buildSDKCertificates(modelCerts []albmodel.Certificate) ([]albsdk.Certificate, []albsdk.Certificate) {
	if len(modelCerts) == 0 {
		return nil, nil
	}

	var defaultSDKCerts []albsdk.Certificate
	var extraSDKCerts []albsdk.Certificate
	for _, cert := range modelCerts {
		if cert.IsDefault {
			defaultSDKCerts = append(defaultSDKCerts, albsdk.Certificate{
				IsDefault:     cert.IsDefault,
				CertificateId: cert.CertificateId,
				Status:        cert.Status,
			})
		} else {
			extraSDKCerts = append(extraSDKCerts, albsdk.Certificate{
				IsDefault:     cert.IsDefault,
				CertificateId: cert.CertificateId,
				Status:        cert.Status,
			})
		}
	}
	return defaultSDKCerts, extraSDKCerts
}

func isIncorrectStatusLoadBalancerError(err error) bool {
	if strings.Contains(err.Error(), "IncorrectStatus.LoadBalancer") {
		return true
	}
	return false
}

func isIncorrectStatusListenerError(err error) bool {
	if strings.Contains(err.Error(), "IncorrectStatus.Listener") {
		return true
	}
	return false
}

func isHTTPSListenerProtocol(protocol string) bool {
	return strings.EqualFold(protocol, util.ListenerProtocolHTTPS)
}

func isListenerProtocolValid(protocol string) bool {
	if strings.EqualFold(protocol, util.ListenerProtocolHTTP) ||
		strings.EqualFold(protocol, util.ListenerProtocolHTTPS) ||
		strings.EqualFold(protocol, util.ListenerProtocolQUIC) {
		return true
	}
	return false
}

func isListenerPortValid(port int) bool {
	if port < 1 || port > 65535 {
		return false
	}
	return true
}

func isListenerRequestTimeoutValid(requestTimeout int) bool {
	if requestTimeout < 1 || requestTimeout > 180 {
		return false
	}
	return true
}

func isListenerIdleTimeoutValid(idleTimeout int) bool {
	if idleTimeout < 1 || idleTimeout > 60 {
		return false
	}
	return true
}

func buildResListenerStatus(lsID string) albmodel.ListenerStatus {
	return albmodel.ListenerStatus{
		ListenerID: lsID,
	}
}
