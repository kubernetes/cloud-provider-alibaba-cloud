package alb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sls"
	"github.com/go-logr/logr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	DefaultWaitSGPDeletionPollInterval = 2 * time.Second
	DefaultWaitSGPDeletionTimeout      = 50 * time.Second
	DefaultWaitLSExistencePollInterval = 2 * time.Second
	DefaultWaitLSExistenceTimeout      = 20 * time.Second
)

func NewALBProvider(
	auth *base.ClientMgr,
) *ALBProvider {
	logger := ctrl.Log.WithName("controllers").WithName("ALBProvider")
	return &ALBProvider{
		logger:                      logger,
		auth:                        auth,
		waitSGPDeletionPollInterval: DefaultWaitSGPDeletionPollInterval,
		waitSGPDeletionTimeout:      DefaultWaitSGPDeletionTimeout,
		waitLSExistenceTimeout:      DefaultWaitLSExistenceTimeout,
		waitLSExistencePollInterval: DefaultWaitLSExistencePollInterval,
	}
}

var _ prvd.IALB = &ALBProvider{}

type ALBProvider struct {
	auth     *base.ClientMgr
	logger   logr.Logger
	sdkCerts []albsdk.CertificateModel

	waitLSExistencePollInterval time.Duration
	waitLSExistenceTimeout      time.Duration
	waitSGPDeletionPollInterval time.Duration
	waitSGPDeletionTimeout      time.Duration
}

func (m *ALBProvider) CreateALB(ctx context.Context, resLB *alb.AlbLoadBalancer, trackingProvider tracking.TrackingProvider) (alb.LoadBalancerStatus, error) {
	traceID := ctx.Value(util.TraceID)

	createLbReq, err := buildSDKCreateAlbLoadBalancerRequest(resLB.Spec)
	if err != nil {
		return alb.LoadBalancerStatus{}, err
	}

	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("creating loadBalancer",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"startTime", startTime,
		util.Action, util.CreateALBLoadBalancer)
	createLbResp, err := m.auth.ALB.CreateLoadBalancer(createLbReq)
	if err != nil {
		return alb.LoadBalancerStatus{}, err
	}
	m.logger.V(util.MgrLogLevel).Info("created loadBalancer",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"loadBalancerID", createLbResp.LoadBalancerId,
		"traceID", traceID,
		"requestID", createLbResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.CreateALBLoadBalancer)

	asynchronousStartTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("creating loadBalancer asynchronous",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", createLbResp.LoadBalancerId,
		"startTime", asynchronousStartTime,
		util.Action, util.CreateALBLoadBalancerAsynchronous)
	var getLbResp *albsdk.GetLoadBalancerAttributeResponse
	for i := 0; i < util.CreateLoadBalancerWaitActiveMaxRetryTimes; i++ {
		time.Sleep(util.CreateLoadBalancerWaitActiveRetryInterval)

		getLbResp, err = getALBLoadBalancerAttributeFunc(ctx, createLbResp.LoadBalancerId, m.auth, m.logger)
		if err != nil {
			return alb.LoadBalancerStatus{}, err
		}
		if isAlbLoadBalancerActive(getLbResp.LoadBalancerStatus) {
			break
		}
	}
	m.logger.V(util.MgrLogLevel).Info("created loadBalancer asynchronous",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", createLbResp.LoadBalancerId,
		"requestID", getLbResp.RequestId,
		"elapsedTime", time.Since(asynchronousStartTime).Milliseconds(),
		util.Action, util.CreateALBLoadBalancerAsynchronous)

	if err := m.Tag(ctx, resLB, createLbResp.LoadBalancerId, trackingProvider); err != nil {
		if errTmp := m.DeleteALB(ctx, createLbResp.LoadBalancerId); errTmp != nil {
			m.logger.V(util.MgrLogLevel).Error(errTmp, "roll back load balancer failed",
				"stackID", resLB.Stack().StackID(),
				"resourceID", resLB.ID(),
				"traceID", traceID,
				"loadBalancerID", createLbResp.LoadBalancerId,
				util.Action, util.TagALBResource)
		}
		return alb.LoadBalancerStatus{}, err
	}

	if len(resLB.Spec.AccessLogConfig.LogProject) != 0 && len(resLB.Spec.AccessLogConfig.LogStore) != 0 {
		if err := m.AnalyzeAndAssociateAccessLogToALB(ctx, createLbResp.LoadBalancerId, resLB); err != nil {
			return alb.LoadBalancerStatus{}, err
		}
	}

	return buildResAlbLoadBalancerStatus(createLbResp.LoadBalancerId, getLbResp.DNSName), nil
}

func isAlbLoadBalancerActive(status string) bool {
	return strings.EqualFold(status, util.LoadBalancerStatusActive)
}

func (m *ALBProvider) UpdateALB(ctx context.Context, resLB *alb.AlbLoadBalancer, sdkLB albsdk.LoadBalancer) (alb.LoadBalancerStatus, error) {
	if err := m.updateAlbLoadBalancerAttribute(ctx, resLB, &sdkLB); err != nil {
		return alb.LoadBalancerStatus{}, err
	}
	if err := m.updateAlbLoadBalancerDeletionProtection(ctx, resLB, &sdkLB); err != nil {
		return alb.LoadBalancerStatus{}, err
	}
	if err := m.updateAlbLoadBalancerAccessLogConfig(ctx, resLB, &sdkLB); err != nil {
		return alb.LoadBalancerStatus{}, err
	}
	if err := m.updateAlbLoadBalancerEdition(ctx, resLB, &sdkLB); err != nil {
		return alb.LoadBalancerStatus{}, err
	}

	return buildResAlbLoadBalancerStatus(sdkLB.LoadBalancerId, sdkLB.DNSName), nil
}

func (m *ALBProvider) Tag(ctx context.Context, resLB *alb.AlbLoadBalancer, lbID string, trackingProvider tracking.TrackingProvider) error {
	traceID := ctx.Value(util.TraceID)

	lbTags := trackingProvider.ResourceTags(resLB.Stack(), resLB, transTagListToMap(resLB.Spec.Tags))
	lbIDs := make([]string, 0)
	lbIDs = append(lbIDs, lbID)
	tags := transTagMapToSDKTagResourcesTagList(lbTags)
	tagReq := albsdk.CreateTagResourcesRequest()
	tagReq.Tag = &tags
	tagReq.ResourceId = &lbIDs
	tagReq.ResourceType = util.LoadBalancerResourceType
	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("tagging resource",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", lbID,
		"startTime", startTime,
		util.Action, util.TagALBResource)
	tagResp, err := m.auth.ALB.TagResources(tagReq)
	if err != nil {
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("tagged resource",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", lbID,
		"requestID", tagResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.TagALBResource)

	return nil
}

func (m *ALBProvider) preCheckTagConflictForReuse(ctx context.Context, sdkLB *albsdk.GetLoadBalancerAttributeResponse, resLB *alb.AlbLoadBalancer, trackingProvider tracking.TrackingProvider) error {
	if sdkLB.VpcId != resLB.Spec.VpcId {
		return fmt.Errorf("the vpc %s of reused alb %s is not same with cluster vpc %s", sdkLB.VpcId, sdkLB.LoadBalancerId, resLB.Spec.VpcId)
	}

	if len(sdkLB.Tags) == 0 {
		return nil
	}
	sdkTags := transSDKTagListToMap(sdkLB.Tags)
	resTags := trackingProvider.ResourceTags(resLB.Stack(), resLB, transTagListToMap(resLB.Spec.Tags))

	sdkClusterID, okSdkClusterID := sdkTags[trackingProvider.ClusterNameTagKey()]
	resClusterID, okResClusterID := resTags[trackingProvider.ClusterNameTagKey()]
	if okSdkClusterID && okResClusterID {
		if sdkClusterID != resClusterID {
			return fmt.Errorf("alb %s belongs to cluster: %s, cant reuse alb to another cluster: %s",
				sdkLB.LoadBalancerId, sdkClusterID, resClusterID)
		}

		sdkAlbConfig, okSdkAlbConfig := sdkTags[trackingProvider.AlbConfigTagKey()]
		resAlbConfig, okResAlbConfig := resTags[trackingProvider.AlbConfigTagKey()]
		if okSdkAlbConfig && okResAlbConfig {
			if sdkAlbConfig != resAlbConfig {
				return fmt.Errorf("alb %s belongs to albconfig: %s, cant reuse alb to another albconfig: %s",
					sdkLB.LoadBalancerId, sdkAlbConfig, resAlbConfig)
			}
		}
	}

	return nil
}

func (m *ALBProvider) ReuseALB(ctx context.Context, resLB *alb.AlbLoadBalancer, lbID string, trackingProvider tracking.TrackingProvider) (alb.LoadBalancerStatus, error) {
	getLbResp, err := getALBLoadBalancerAttributeFunc(ctx, lbID, m.auth, m.logger)
	if err != nil {
		return alb.LoadBalancerStatus{}, err
	}

	if err := m.preCheckTagConflictForReuse(ctx, getLbResp, resLB, trackingProvider); err != nil {
		return alb.LoadBalancerStatus{}, err
	}

	if err = m.Tag(ctx, resLB, lbID, trackingProvider); err != nil {
		return alb.LoadBalancerStatus{}, err
	}

	if resLB.Spec.ForceOverride != nil && *resLB.Spec.ForceOverride {
		loadBalancer := transSDKGetLoadBalancerAttributeResponseToLoadBalancer(*getLbResp)
		return m.UpdateALB(ctx, resLB, loadBalancer)
	}

	return buildResAlbLoadBalancerStatus(lbID, getLbResp.DNSName), nil
}

func transSDKGetLoadBalancerAttributeResponseToLoadBalancer(albAttr albsdk.GetLoadBalancerAttributeResponse) albsdk.LoadBalancer {
	return albsdk.LoadBalancer{
		AddressAllocatedMode:         albAttr.AddressAllocatedMode,
		AddressType:                  albAttr.AddressType,
		BandwidthCapacity:            albAttr.BandwidthCapacity,
		BandwidthPackageId:           albAttr.BandwidthPackageId,
		CreateTime:                   albAttr.CreateTime,
		DNSName:                      albAttr.DNSName,
		ServiceManagedEnabled:        albAttr.ServiceManagedEnabled,
		ServiceManagedMode:           albAttr.ServiceManagedMode,
		LoadBalancerBussinessStatus:  albAttr.LoadBalancerBussinessStatus,
		LoadBalancerEdition:          albAttr.LoadBalancerEdition,
		LoadBalancerId:               albAttr.LoadBalancerId,
		LoadBalancerName:             albAttr.LoadBalancerName,
		LoadBalancerStatus:           albAttr.LoadBalancerStatus,
		ResourceGroupId:              albAttr.ResourceGroupId,
		VpcId:                        albAttr.VpcId,
		AccessLogConfig:              albAttr.AccessLogConfig,
		DeletionProtectionConfig:     albAttr.DeletionProtectionConfig,
		LoadBalancerBillingConfig:    albAttr.LoadBalancerBillingConfig,
		ModificationProtectionConfig: albAttr.ModificationProtectionConfig,
		LoadBalancerOperationLocks:   albAttr.LoadBalancerOperationLocks,
		Tags:                         albAttr.Tags,
	}
}

var getALBLoadBalancerAttributeFunc = func(ctx context.Context, lbID string, auth *base.ClientMgr, logger logr.Logger) (*albsdk.GetLoadBalancerAttributeResponse, error) {
	traceID := ctx.Value(util.TraceID)

	getLbReq := albsdk.CreateGetLoadBalancerAttributeRequest()
	getLbReq.LoadBalancerId = lbID
	startTime := time.Now()
	logger.V(util.MgrLogLevel).Info("getting loadBalancer attribute",
		"traceID", traceID,
		"loadBalancerID", lbID,
		"startTime", startTime,
		util.Action, util.GetALBLoadBalancerAttribute)
	getLbResp, err := auth.ALB.GetLoadBalancerAttribute(getLbReq)
	if err != nil {
		return nil, err
	}
	logger.V(util.MgrLogLevel).Info("got loadBalancer attribute",
		"traceID", traceID,
		"loadBalancerID", lbID,
		"loadBalancerStatus", getLbResp.LoadBalancerStatus,
		"requestID", getLbResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.GetALBLoadBalancerAttribute)
	return getLbResp, nil
}

var disableALBDeletionProtectionFunc = func(ctx context.Context, lbID string, auth *base.ClientMgr, logger logr.Logger) (*albsdk.DisableDeletionProtectionResponse, error) {
	traceID := ctx.Value(util.TraceID)

	updateLbReq := albsdk.CreateDisableDeletionProtectionRequest()
	updateLbReq.ResourceId = lbID
	startTime := time.Now()
	logger.V(util.MgrLogLevel).Info("disabling delete protection",
		"traceID", traceID,
		"loadBalancerID", lbID,
		"startTime", startTime,
		util.Action, util.DisableALBDeletionProtection)
	updateLbResp, err := auth.ALB.DisableDeletionProtection(updateLbReq)
	if err != nil {
		return nil, err
	}
	logger.V(util.MgrLogLevel).Info("disabled delete protection",
		"traceID", traceID,
		"loadBalancerID", lbID,
		"requestID", updateLbResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.DisableALBDeletionProtection)

	return updateLbResp, nil
}

var enableALBDeletionProtectionFunc = func(ctx context.Context, lbID string, auth *base.ClientMgr, logger logr.Logger) (*albsdk.EnableDeletionProtectionResponse, error) {
	traceID := ctx.Value(util.TraceID)

	updateLbReq := albsdk.CreateEnableDeletionProtectionRequest()
	updateLbReq.ResourceId = lbID
	startTime := time.Now()
	logger.V(util.MgrLogLevel).Info("enabling delete protection",
		"traceID", traceID,
		"loadBalancerID", lbID,
		"startTime", startTime,
		util.Action, util.EnableALBDeletionProtection)
	updateLbResp, err := auth.ALB.EnableDeletionProtection(updateLbReq)
	if err != nil {
		return nil, err
	}
	logger.V(util.MgrLogLevel).Info("enabled delete protection",
		"traceID", traceID,
		"loadBalancerID", lbID,
		"requestID", updateLbResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.EnableALBDeletionProtection)

	return updateLbResp, nil
}

var deleteALBLoadBalancerFunc = func(ctx context.Context, m *ALBProvider, lbID string) (*albsdk.DeleteLoadBalancerResponse, error) {
	traceID := ctx.Value(util.TraceID)

	lbReq := albsdk.CreateDeleteLoadBalancerRequest()
	lbReq.LoadBalancerId = lbID

	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("deleting loadBalancer",
		"loadBalancerID", lbID,
		"traceID", traceID,
		"startTime", startTime,
		util.Action, util.DeleteALBLoadBalancer)
	lsResp, err := m.auth.ALB.DeleteLoadBalancer(lbReq)
	if err != nil {
		return nil, err
	}
	m.logger.V(util.MgrLogLevel).Info("deleted loadBalancer",
		"loadBalancerID", lbID,
		"traceID", traceID,
		"requestID", lsResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.DeleteALBLoadBalancer)

	return lsResp, nil
}

func (m *ALBProvider) DeleteALB(ctx context.Context, lbID string) error {
	getLbResp, err := getALBLoadBalancerAttributeFunc(ctx, lbID, m.auth, m.logger)
	if err != nil {
		return err
	}

	if getLbResp.DeletionProtectionConfig.Enabled {
		_, err := disableALBDeletionProtectionFunc(ctx, lbID, m.auth, m.logger)
		if err != nil {
			return err
		}
	}

	if _, err = deleteALBLoadBalancerFunc(ctx, m, lbID); err != nil {
		return err
	}

	return nil
}

func transSDKModificationProtectionConfigToCreateLb(mpc alb.ModificationProtectionConfig) albsdk.CreateLoadBalancerModificationProtectionConfig {
	return albsdk.CreateLoadBalancerModificationProtectionConfig{
		Reason: mpc.Reason,
		Status: mpc.Status,
	}
}

func transSDKLoadBalancerBillingConfigToCreateLb(lbc alb.LoadBalancerBillingConfig) albsdk.CreateLoadBalancerLoadBalancerBillingConfig {
	return albsdk.CreateLoadBalancerLoadBalancerBillingConfig{
		PayType: lbc.PayType,
	}
}

func transSDKZoneMappingsToCreateLb(zoneMappings []alb.ZoneMapping) *[]albsdk.CreateLoadBalancerZoneMappings {
	createLbZoneMappings := make([]albsdk.CreateLoadBalancerZoneMappings, 0)

	for _, zoneMapping := range zoneMappings {
		createLbZoneMapping := albsdk.CreateLoadBalancerZoneMappings{
			VSwitchId: zoneMapping.VSwitchId,
			ZoneId:    zoneMapping.ZoneId,
		}
		createLbZoneMappings = append(createLbZoneMappings, createLbZoneMapping)
	}

	return &createLbZoneMappings
}

func buildSDKCreateAlbLoadBalancerRequest(lbSpec alb.ALBLoadBalancerSpec) (*albsdk.CreateLoadBalancerRequest, error) {
	createLbReq := albsdk.CreateCreateLoadBalancerRequest()
	if len(lbSpec.VpcId) == 0 {
		return nil, fmt.Errorf("invalid load balancer vpc id: %s", lbSpec.VpcId)
	}
	createLbReq.VpcId = lbSpec.VpcId

	if !isAlbLoadBalancerAddressTypeValid(lbSpec.AddressType) {
		return nil, fmt.Errorf("invalid load balancer address type: %s", lbSpec.AddressType)
	}
	createLbReq.AddressType = lbSpec.AddressType

	createLbReq.LoadBalancerName = lbSpec.LoadBalancerName

	createLbReq.DeletionProtectionEnabled = requests.NewBoolean(lbSpec.DeletionProtectionConfig.Enabled)

	if !isAlbLoadBalancerModificationProtectionStatusValid(lbSpec.ModificationProtectionConfig.Status) {
		return nil, fmt.Errorf("invalid load balancer modification protection config: %v", lbSpec.ModificationProtectionConfig)
	}
	createLbReq.ModificationProtectionConfig = transSDKModificationProtectionConfigToCreateLb(lbSpec.ModificationProtectionConfig)

	if len(lbSpec.ZoneMapping) == 0 {
		return nil, fmt.Errorf("empty load balancer zone mapping")
	}
	createLbReq.ZoneMappings = transSDKZoneMappingsToCreateLb(lbSpec.ZoneMapping)

	if !isLoadBalancerAddressAllocatedModeValid(lbSpec.AddressAllocatedMode) {
		return nil, fmt.Errorf("invalid load balancer address allocate mode: %s", lbSpec.AddressAllocatedMode)
	}
	createLbReq.AddressAllocatedMode = lbSpec.AddressAllocatedMode

	createLbReq.ResourceGroupId = lbSpec.ResourceGroupId

	if !isAlbLoadBalancerEditionValid(lbSpec.LoadBalancerEdition) {
		return nil, fmt.Errorf("invalid load balancer edition: %s", lbSpec.LoadBalancerEdition)
	}
	createLbReq.LoadBalancerEdition = lbSpec.LoadBalancerEdition

	if !isAlbLoadBalancerLoadBalancerPayTypeValid(lbSpec.LoadBalancerBillingConfig.PayType) {
		return nil, fmt.Errorf("invalid load balancer paytype: %s", lbSpec.LoadBalancerBillingConfig.PayType)
	}
	createLbReq.LoadBalancerBillingConfig = transSDKLoadBalancerBillingConfigToCreateLb(lbSpec.LoadBalancerBillingConfig)

	return createLbReq, nil
}

func transSDKModificationProtectionConfigToUpdateLb(mpc alb.ModificationProtectionConfig) albsdk.UpdateLoadBalancerAttributeModificationProtectionConfig {
	return albsdk.UpdateLoadBalancerAttributeModificationProtectionConfig{
		Reason: mpc.Reason,
		Status: mpc.Status,
	}
}

func (m *ALBProvider) updateAlbLoadBalancerAttribute(ctx context.Context, resLB *alb.AlbLoadBalancer, sdkLB *albsdk.LoadBalancer) error {
	traceID := ctx.Value(util.TraceID)

	var (
		isModificationProtectionConfigModifyNeedUpdate = false
		isLoadBalancerNameNeedUpdate                   = false
	)

	if !isAlbLoadBalancerModificationProtectionStatusValid(resLB.Spec.ModificationProtectionConfig.Status) {
		return fmt.Errorf("invalid load balancer modification protection config: %v", resLB.Spec.ModificationProtectionConfig)
	}
	modificationProtectionConfig := transModificationProtectionConfigToSDK(resLB.Spec.ModificationProtectionConfig)
	if modificationProtectionConfig != sdkLB.ModificationProtectionConfig {
		m.logger.V(util.MgrLogLevel).Info("ModificationProtectionConfig update",
			"res", resLB.Spec.ModificationProtectionConfig,
			"sdk", sdkLB.ModificationProtectionConfig,
			"loadBalancerID", sdkLB.LoadBalancerId,
			"traceID", traceID)
		isModificationProtectionConfigModifyNeedUpdate = true
	}
	if resLB.Spec.LoadBalancerName != sdkLB.LoadBalancerName {
		m.logger.V(util.MgrLogLevel).Info("LoadBalancerName update",
			"res", resLB.Spec.LoadBalancerName,
			"sdk", sdkLB.LoadBalancerName,
			"loadBalancerID", sdkLB.LoadBalancerId,
			"traceID", traceID)
		isLoadBalancerNameNeedUpdate = true
	}

	if !isLoadBalancerNameNeedUpdate && !isModificationProtectionConfigModifyNeedUpdate {
		return nil
	}

	updateLbReq := albsdk.CreateUpdateLoadBalancerAttributeRequest()
	updateLbReq.LoadBalancerId = sdkLB.LoadBalancerId
	if isModificationProtectionConfigModifyNeedUpdate {
		updateLbReq.ModificationProtectionConfig = transSDKModificationProtectionConfigToUpdateLb(resLB.Spec.ModificationProtectionConfig)
	}
	if isLoadBalancerNameNeedUpdate {
		updateLbReq.LoadBalancerName = resLB.Spec.LoadBalancerName
	}

	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("updating loadBalancer attribute",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", sdkLB.LoadBalancerId,
		"startTime", startTime,
		util.Action, util.UpdateALBLoadBalancerAttribute)
	updateLbResp, err := m.auth.ALB.UpdateLoadBalancerAttribute(updateLbReq)
	if err != nil {
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("updating loadBalancer attribute",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", sdkLB.LoadBalancerId,
		"requestID", updateLbResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.UpdateALBLoadBalancerAttribute)

	return nil
}

func (m *ALBProvider) updateAlbLoadBalancerDeletionProtection(ctx context.Context, resLB *alb.AlbLoadBalancer, sdkLB *albsdk.LoadBalancer) error {
	traceID := ctx.Value(util.TraceID)

	var (
		isDeletionProtectionNeedUpdate = false
	)

	if resLB.Spec.DeletionProtectionConfig.Enabled != sdkLB.DeletionProtectionConfig.Enabled {
		m.logger.V(util.MgrLogLevel).Info("DeletionProtectionConfig update",
			"res", resLB.Spec.DeletionProtectionConfig.Enabled,
			"sdk", sdkLB.DeletionProtectionConfig.Enabled,
			"loadBalancerID", sdkLB.LoadBalancerId,
			"traceID", traceID)
		isDeletionProtectionNeedUpdate = true
	}
	if !isDeletionProtectionNeedUpdate {
		return nil
	}

	if resLB.Spec.DeletionProtectionConfig.Enabled && !sdkLB.DeletionProtectionConfig.Enabled {
		_, err := enableALBDeletionProtectionFunc(ctx, sdkLB.LoadBalancerId, m.auth, m.logger)
		if err != nil {
			return err
		}
	} else if !resLB.Spec.DeletionProtectionConfig.Enabled && sdkLB.DeletionProtectionConfig.Enabled {
		_, err := disableALBDeletionProtectionFunc(ctx, sdkLB.LoadBalancerId, m.auth, m.logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func reCorrectRegion(region string) string {
	switch region {
	case "cn-shenzhen-finance-1":
		return "cn-shenzhen-finance"
	}
	return region
}

func (m *ALBProvider) DissociateAccessLogFromALB(ctx context.Context, lbID string, resLB *alb.AlbLoadBalancer) error {
	traceID := ctx.Value(util.TraceID)

	updateLbReq := albsdk.CreateDisableLoadBalancerAccessLogRequest()
	updateLbReq.LoadBalancerId = lbID
	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("disabling loadBalancer access log",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", lbID,
		"startTime", startTime,
		util.Action, util.DisableALBLoadBalancerAccessLog)
	updateLbResp, err := m.auth.ALB.DisableLoadBalancerAccessLog(updateLbReq)
	if err != nil {
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("disabled loadBalancer access log",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", lbID,
		"requestID", updateLbResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.DisableALBLoadBalancerAccessLog)
	return nil
}

func (m *ALBProvider) AnalyzeAndAssociateAccessLogToALB(ctx context.Context, lbID string, resLB *alb.AlbLoadBalancer) error {
	traceID := ctx.Value(util.TraceID)

	logProject := resLB.Spec.AccessLogConfig.LogProject
	logStore := resLB.Spec.AccessLogConfig.LogStore

	if !isLogProjectNameValid(logProject) || !isLogStoreNameValid(logStore) {
		return fmt.Errorf("invalid name of logProject: %s or logStore: %s", logProject, logStore)
	}

	logReq := sls.CreateAnalyzeProductLogRequest()
	region, err := m.auth.Meta.Region()
	if err != nil {
		return err
	}
	logReq.Region = reCorrectRegion(region)
	logReq.Logstore = logStore
	logReq.Project = logProject
	logReq.AcceptFormat = util.DefaultLogAcceptFormat
	logReq.CloudProduct = util.DefaultLogCloudProduct
	logReq.Lang = util.DefaultLogLang
	logReq.Domain = fmt.Sprintf("%s%s", logReq.Region, util.DefaultLogDomainSuffix)
	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("analyzing product log",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"request", logReq,
		"loadBalancerID", lbID,
		"startTime", startTime,
		util.Action, util.AnalyzeProductLog)
	logResp, err := m.auth.SLS.AnalyzeProductLog(logReq)
	if err != nil {
		m.logger.V(util.MgrLogLevel).Info("analyzing product log",
			"stackID", resLB.Stack().StackID(),
			"resourceID", resLB.ID(),
			"loadBalancerID", lbID,
			"traceID", traceID,
			"requestID", logResp.RequestId,
			"error", err.Error(),
			util.Action, util.AnalyzeProductLog)
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("analyzed product log",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"loadBalancerID", lbID,
		"traceID", traceID,
		"requestID", logResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.AnalyzeProductLog)

	lbReq := albsdk.CreateEnableLoadBalancerAccessLogRequest()
	lbReq.LoadBalancerId = lbID
	lbReq.LogProject = logProject
	lbReq.LogStore = logStore
	startTime = time.Now()
	m.logger.V(util.MgrLogLevel).Info("enabling loadBalancer access log",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", lbID,
		"startTime", startTime,
		util.Action, util.EnableALBLoadBalancerAccessLog)
	lbResp, err := m.auth.ALB.EnableLoadBalancerAccessLog(lbReq)
	if err != nil {
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("enabled loadBalancer access log",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"loadBalancerID", lbID,
		"traceID", traceID,
		"requestID", lbResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.EnableALBLoadBalancerAccessLog)
	return nil
}

func (m *ALBProvider) updateAlbLoadBalancerAccessLogConfig(ctx context.Context, resLB *alb.AlbLoadBalancer, sdkLB *albsdk.LoadBalancer) error {
	traceID := ctx.Value(util.TraceID)

	var (
		isAccessLogConfigNeedUpdate = false
	)
	accessLogConfig := transAccessLogConfigToSDK(resLB.Spec.AccessLogConfig)
	if accessLogConfig != sdkLB.AccessLogConfig {
		m.logger.Info("LoadBalancer AccessLogConfig update",
			"res", resLB.Spec.AccessLogConfig,
			"sdk", sdkLB.AccessLogConfig,
			"traceID", traceID)
		isAccessLogConfigNeedUpdate = true
	}
	if !isAccessLogConfigNeedUpdate {
		return nil
	}

	if len(sdkLB.AccessLogConfig.LogProject) != 0 && len(sdkLB.AccessLogConfig.LogStore) != 0 {
		if err := m.DissociateAccessLogFromALB(ctx, sdkLB.LoadBalancerId, resLB); err != nil {
			return err
		}
	}

	if len(resLB.Spec.AccessLogConfig.LogProject) != 0 && len(resLB.Spec.AccessLogConfig.LogStore) != 0 {
		if err := m.AnalyzeAndAssociateAccessLogToALB(ctx, sdkLB.LoadBalancerId, resLB); err != nil {
			return err
		}
	}
	return nil
}

func (m *ALBProvider) updateAlbLoadBalancerEdition(ctx context.Context, resLB *alb.AlbLoadBalancer, sdkLB *albsdk.LoadBalancer) error {
	traceID := ctx.Value(util.TraceID)

	var (
		isLoadBalancerEditionNeedUpdate = false
	)

	if !isAlbLoadBalancerEditionValid(resLB.Spec.LoadBalancerEdition) {
		return fmt.Errorf("invalid load balancer edition: %s", resLB.Spec.LoadBalancerEdition)
	}
	if strings.EqualFold(resLB.Spec.LoadBalancerEdition, util.LoadBalancerEditionBasic) &&
		strings.EqualFold(sdkLB.LoadBalancerEdition, util.LoadBalancerEditionStandard) {
		return errors.New("downgrade not allowed for alb from standard to basic")
	}
	if !strings.EqualFold(resLB.Spec.LoadBalancerEdition, sdkLB.LoadBalancerEdition) {
		m.logger.V(util.MgrLogLevel).Info("LoadBalancer Edition update",
			"res", resLB.Spec.LoadBalancerEdition,
			"sdk", sdkLB.LoadBalancerEdition,
			"loadBalancerID", sdkLB.LoadBalancerId,
			"traceID", traceID)
		isLoadBalancerEditionNeedUpdate = true
	}
	if !isLoadBalancerEditionNeedUpdate {
		return nil
	}

	updateLbReq := albsdk.CreateUpdateLoadBalancerEditionRequest()
	updateLbReq.LoadBalancerId = sdkLB.LoadBalancerId
	updateLbReq.LoadBalancerEdition = resLB.Spec.LoadBalancerEdition
	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("updating loadBalancer edition",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"startTime", startTime,
		"traceID", traceID,
		"loadBalancerID", sdkLB.LoadBalancerId,
		util.Action, util.UpdateALBLoadBalancerEdition)
	updateLbResp, err := m.auth.ALB.UpdateLoadBalancerEdition(updateLbReq)
	if err != nil {
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("updated loadBalancer edition",
		"stackID", resLB.Stack().StackID(),
		"resourceID", resLB.ID(),
		"traceID", traceID,
		"loadBalancerID", sdkLB.LoadBalancerId,
		"requestID", updateLbResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.UpdateALBLoadBalancerEdition)

	return nil
}

func transTagMapToSDKTagResourcesTagList(tagMap map[string]string) []albsdk.TagResourcesTag {
	tagList := make([]albsdk.TagResourcesTag, 0)
	for k, v := range tagMap {
		tagList = append(tagList, albsdk.TagResourcesTag{
			Key:   k,
			Value: v,
		})
	}
	return tagList
}

func transTagListToMap(tagList []alb.ALBTag) map[string]string {
	tagMap := make(map[string]string, 0)
	for _, tag := range tagList {
		tagMap[tag.Key] = tag.Value
	}
	return tagMap
}

func transSDKTagListToMap(tagList []albsdk.Tag) map[string]string {
	tagMap := make(map[string]string, 0)
	for _, tag := range tagList {
		tagMap[tag.Key] = tag.Value
	}
	return tagMap
}
func buildResAlbLoadBalancerStatus(lbID, DNSName string) alb.LoadBalancerStatus {
	return alb.LoadBalancerStatus{
		LoadBalancerID: lbID,
		DNSName:        DNSName,
	}
}

func isAlbLoadBalancerAddressTypeValid(addressType string) bool {
	if strings.EqualFold(addressType, util.LoadBalancerAddressTypeInternet) ||
		strings.EqualFold(addressType, util.LoadBalancerAddressTypeIntranet) {
		return true
	}
	return false
}

func isAlbLoadBalancerModificationProtectionStatusValid(modificationProtectionStatus string) bool {
	if strings.EqualFold(modificationProtectionStatus, util.LoadBalancerModificationProtectionStatusNonProtection) ||
		strings.EqualFold(modificationProtectionStatus, util.LoadBalancerModificationProtectionStatusConsoleProtection) {
		return true
	}
	return false
}

func isLoadBalancerAddressAllocatedModeValid(addressAllocatedMode string) bool {
	if strings.EqualFold(addressAllocatedMode, util.LoadBalancerAddressAllocatedModeFixed) ||
		strings.EqualFold(addressAllocatedMode, util.LoadBalancerAddressAllocatedModeDynamic) {
		return true
	}
	return false
}

func (p ALBProvider) TagALBResources(request *albsdk.TagResourcesRequest) (response *albsdk.TagResourcesResponse, err error) {
	return p.auth.ALB.TagResources(request)
}
func (p ALBProvider) DescribeALBZones(request *albsdk.DescribeZonesRequest) (response *albsdk.DescribeZonesResponse, err error) {
	return p.auth.ALB.DescribeZones(request)
}

func isAlbLoadBalancerEditionValid(edition string) bool {
	if strings.EqualFold(edition, util.LoadBalancerEditionBasic) ||
		strings.EqualFold(edition, util.LoadBalancerEditionStandard) {
		return true
	}
	return false
}

func isAlbLoadBalancerLoadBalancerPayTypeValid(payType string) bool {
	if strings.EqualFold(payType, util.LoadBalancerPayTypePostPay) {
		return true
	}
	return false
}

func isLogProjectNameValid(logProject string) bool {
	if len(logProject) < util.MinLogProjectNameLen || len(logProject) > util.MaxLogProjectNameLen {
		return false
	}
	return true
}

func isLogStoreNameValid(logStore string) bool {
	if len(logStore) < util.MinLogStoreNameLen || len(logStore) > util.MaxLogStoreNameLen {
		return false
	}
	return true
}
func transAccessLogConfigToSDK(a alb.AccessLogConfig) albsdk.AccessLogConfig {
	return albsdk.AccessLogConfig{
		LogProject: a.LogProject,
		LogStore:   a.LogStore,
	}
}
func transModificationProtectionConfigToSDK(m alb.ModificationProtectionConfig) albsdk.ModificationProtectionConfig {
	return albsdk.ModificationProtectionConfig{
		Reason: m.Reason,
		Status: m.Status,
	}
}
