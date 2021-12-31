package alb

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/pkg/errors"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
)

func (m *ALBProvider) CreateALBServerGroup(ctx context.Context, resSGP *alb.ServerGroup, trackingProvider tracking.TrackingProvider) (alb.ServerGroupStatus, error) {
	traceID := ctx.Value(util.TraceID)

	createSgpReq, err := buildSDKServerGroupCreateRequest(resSGP.Spec)
	if err != nil {
		return alb.ServerGroupStatus{}, err
	}

	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("creating server group",
		"stackID", resSGP.Stack().StackID(),
		"resourceID", resSGP.ID(),
		"traceID", traceID,
		"startTime", startTime,
		util.Action, util.CreateALBServerGroup)
	createSgpResp, err := m.auth.ALB.CreateServerGroup(createSgpReq)
	if err != nil {
		return alb.ServerGroupStatus{}, err
	}
	m.logger.V(util.MgrLogLevel).Info("created server group",
		"stackID", resSGP.Stack().StackID(),
		"resourceID", resSGP.ID(),
		"traceID", traceID,
		"serverGroupID", createSgpResp.ServerGroupId,
		"requestID", createSgpResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.CreateALBServerGroup)

	sgpTags := trackingProvider.ResourceTags(resSGP.Stack(), resSGP, transTagListToMap(resSGP.Spec.Tags))
	tags := transTagMapToSDKTagResourcesTagList(sgpTags)
	resIDs := make([]string, 0)
	resIDs = append(resIDs, createSgpResp.ServerGroupId)
	tagReq := albsdk.CreateTagResourcesRequest()
	tagReq.Tag = &tags
	tagReq.ResourceId = &resIDs
	tagReq.ResourceType = util.ServerGroupResourceType
	startTime = time.Now()
	m.logger.V(util.MgrLogLevel).Info("tagging resource",
		"stackID", resSGP.Stack().StackID(),
		"resourceID", resSGP.ID(),
		"traceID", traceID,
		"serverGroupID", createSgpResp.ServerGroupId,
		"startTime", startTime,
		util.Action, util.TagALBResource)
	tagResp, err := m.auth.ALB.TagResources(tagReq)
	if err != nil {
		if errTmp := m.DeleteALBServerGroup(ctx, createSgpResp.ServerGroupId); errTmp != nil {
			m.logger.V(util.MgrLogLevel).Error(errTmp, "roll back server group failed",
				"stackID", resSGP.Stack().StackID(),
				"resourceID", resSGP.ID(),
				"traceID", traceID,
				"serverGroupID", createSgpResp.ServerGroupId,
				util.Action, util.TagALBResource)
		}
		return alb.ServerGroupStatus{}, err
	}
	m.logger.V(util.MgrLogLevel).Info("tagged resource",
		"stackID", resSGP.Stack().StackID(),
		"resourceID", resSGP.ID(),
		"traceID", traceID,
		"serverGroupID", createSgpResp.ServerGroupId,
		"requestID", tagResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.TagALBResource)

	return buildReServerGroupStatus(createSgpResp.ServerGroupId), nil
}

func (m *ALBProvider) UpdateALBServerGroup(ctx context.Context, resSGP *alb.ServerGroup, sdkSGP alb.ServerGroupWithTags) (alb.ServerGroupStatus, error) {
	_, err := m.updateServerGroupAttribute(ctx, resSGP, &sdkSGP.ServerGroup)
	if err != nil {
		return alb.ServerGroupStatus{}, err
	}
	return buildReServerGroupStatus(sdkSGP.ServerGroupId), nil
}

func (m *ALBProvider) DeleteALBServerGroup(ctx context.Context, serverGroupID string) error {
	traceID := ctx.Value(util.TraceID)

	deleteSgpReq := albsdk.CreateDeleteServerGroupRequest()
	deleteSgpReq.ServerGroupId = serverGroupID

	if err := util.RetryImmediateOnError(m.waitSGPDeletionPollInterval, m.waitSGPDeletionTimeout, isServerGroupResourceInUseError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("deleting server group",
			"serverGroupID", serverGroupID,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.DeleteALBServerGroup)
		deleteSgpResp, err := m.auth.ALB.DeleteServerGroup(deleteSgpReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("deleting server group",
				"serverGroupID", serverGroupID,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.DeleteALBServerGroup)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("deleted server group",
			"serverGroupID", serverGroupID,
			"traceID", traceID,
			"requestID", deleteSgpResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.DeleteALBServerGroup)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to delete serverGroup")
	}

	return nil
}

func (m *ALBProvider) updateServerGroupAttribute(ctx context.Context, resSGP *alb.ServerGroup, sdkSGP *albsdk.ServerGroup) (*albsdk.UpdateServerGroupAttributeResponse, error) {
	traceID := ctx.Value(util.TraceID)

	var (
		isHealthCheckConfigNeedUpdate,
		isStickySessionConfigNeedUpdate,
		isServerGroupNameNeedUpdate,
		isSchedulerNeedUpdate bool
	)
	if resSGP.Spec.ServerGroupName != sdkSGP.ServerGroupName {
		m.logger.V(util.MgrLogLevel).Info("ServerGroupName update:",
			"res", resSGP.Spec.ServerGroupName,
			"sdk", sdkSGP.ServerGroupName,
			"serverGroupID", sdkSGP.ServerGroupId,
			"traceID", traceID)
		isServerGroupNameNeedUpdate = true
	}
	if !isServerGroupSchedulerValid(resSGP.Spec.Scheduler) {
		return nil, fmt.Errorf("invalid server group scheduler: %s", resSGP.Spec.Scheduler)
	}
	if !strings.EqualFold(resSGP.Spec.Scheduler, sdkSGP.Scheduler) {
		m.logger.V(util.MgrLogLevel).Info("Scheduler update:",
			"res", resSGP.Spec.Scheduler,
			"sdk", sdkSGP.Scheduler,
			"serverGroupID", sdkSGP.ServerGroupId,
			"traceID", traceID)
		isSchedulerNeedUpdate = true
	}

	if err := checkHealthCheckConfigValid(resSGP.Spec.HealthCheckConfig); err != nil {
		return nil, err
	}
	if resSGP.Spec.HealthCheckConfig.HealthCheckEnabled {
		if !reflect.DeepEqual(resSGP.Spec.HealthCheckConfig, sdkSGP.HealthCheckConfig) {
			m.logger.V(util.MgrLogLevel).Info("HealthCheckConfig update:",
				"res", resSGP.Spec.HealthCheckConfig,
				"sdk", sdkSGP.HealthCheckConfig,
				"serverGroupID", sdkSGP.ServerGroupId,
				"traceID", traceID)
			isHealthCheckConfigNeedUpdate = true
		}
	} else if !resSGP.Spec.HealthCheckConfig.HealthCheckEnabled && sdkSGP.HealthCheckConfig.HealthCheckEnabled {
		m.logger.V(util.MgrLogLevel).Info("HealthCheckConfig update:",
			"res", resSGP.Spec.HealthCheckConfig,
			"sdk", sdkSGP.HealthCheckConfig,
			"serverGroupID", sdkSGP.ServerGroupId,
			"traceID", traceID)
		isHealthCheckConfigNeedUpdate = true
	}

	if err := checkStickySessionConfigValid(resSGP.Spec.StickySessionConfig); err != nil {
		return nil, err
	}
	if resSGP.Spec.StickySessionConfig.StickySessionEnabled {
		if !reflect.DeepEqual(resSGP.Spec.StickySessionConfig, sdkSGP.StickySessionConfig) {
			m.logger.V(util.MgrLogLevel).Info("StickySessionConfig update:",
				"res", resSGP.Spec.StickySessionConfig,
				"sdk", sdkSGP.StickySessionConfig,
				"serverGroupID", sdkSGP.ServerGroupId,
				"traceID", traceID)
			isStickySessionConfigNeedUpdate = true
		}
	} else if !resSGP.Spec.StickySessionConfig.StickySessionEnabled && sdkSGP.StickySessionConfig.StickySessionEnabled {
		m.logger.V(util.MgrLogLevel).Info("StickySessionConfig update:",
			"res", resSGP.Spec.StickySessionConfig,
			"sdk", sdkSGP.StickySessionConfig,
			"serverGroupID", sdkSGP.ServerGroupId,
			"traceID", traceID)
		isStickySessionConfigNeedUpdate = true
	}

	if !isServerGroupNameNeedUpdate && !isSchedulerNeedUpdate &&
		!isHealthCheckConfigNeedUpdate && !isStickySessionConfigNeedUpdate {
		return nil, nil
	}

	updateSgpReq := albsdk.CreateUpdateServerGroupAttributeRequest()
	updateSgpReq.ServerGroupId = sdkSGP.ServerGroupId

	if isServerGroupNameNeedUpdate {
		updateSgpReq.ServerGroupName = resSGP.Spec.ServerGroupName
	}
	if isSchedulerNeedUpdate {
		updateSgpReq.Scheduler = resSGP.Spec.Scheduler
	}
	if isHealthCheckConfigNeedUpdate {
		updateSgpReq.HealthCheckConfig = *transSDKHealthCheckConfigToUpdateSGP(resSGP.Spec.HealthCheckConfig)
	}
	if isStickySessionConfigNeedUpdate {
		updateSgpReq.StickySessionConfig = *transSDKStickySessionConfigToUpdateSGP(resSGP.Spec.StickySessionConfig)
	}

	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("updating server group attribute",
		"stackID", resSGP.Stack().StackID(),
		"resourceID", resSGP.ID(),
		"traceID", traceID,
		"serverGroupID", sdkSGP.ServerGroupId,
		"startTime", startTime,
		util.Action, util.UpdateALBServerGroupAttribute)
	updateSgpResp, err := m.auth.ALB.UpdateServerGroupAttribute(updateSgpReq)
	if err != nil {
		return nil, err
	}
	m.logger.V(util.MgrLogLevel).Info("updated server group attribute",
		"stackID", resSGP.Stack().StackID(),
		"resourceID", resSGP.ID(),
		"traceID", traceID,
		"serverGroupID", sdkSGP.ServerGroupId,
		"requestID", updateSgpResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.UpdateALBServerGroupAttribute)

	return updateSgpResp, nil
}

func buildSDKServerGroupCreateRequest(sgpSpec alb.ServerGroupSpec) (*albsdk.CreateServerGroupRequest, error) {
	sgpReq := albsdk.CreateCreateServerGroupRequest()

	sgpReq.ServerGroupName = sgpSpec.ServerGroupName

	if len(sgpSpec.VpcId) == 0 {
		return nil, fmt.Errorf("invalid server group vpc id: %s", sgpSpec.VpcId)
	}
	sgpReq.VpcId = sgpSpec.VpcId

	if !isServerGroupSchedulerValid(sgpSpec.Scheduler) {
		return nil, fmt.Errorf("invalid server group scheduler: %s", sgpSpec.Scheduler)
	}
	sgpReq.Scheduler = sgpSpec.Scheduler

	if !isServerGroupProtocolValid(sgpSpec.Protocol) {
		return nil, fmt.Errorf("invalid server group protocol: %s", sgpSpec.Protocol)
	}
	sgpReq.Protocol = sgpSpec.Protocol

	sgpReq.ResourceGroupId = sgpSpec.ResourceGroupId
	if err := checkHealthCheckConfigValid(sgpSpec.HealthCheckConfig); err != nil {
		return nil, err
	}
	sgpReq.HealthCheckConfig = *transSDKHealthCheckConfigToCreateSGP(sgpSpec.HealthCheckConfig)
	if err := checkStickySessionConfigValid(sgpSpec.StickySessionConfig); err != nil {
		return nil, err
	}
	sgpReq.StickySessionConfig = *transSDKStickySessionConfigToCreateSGP(sgpSpec.StickySessionConfig)
	sgpReq.ServerGroupType = sgpSpec.ServerGroupType

	return sgpReq, nil
}

func checkHealthCheckConfigValid(conf alb.HealthCheckConfig) error {
	if !conf.HealthCheckEnabled {
		return nil
	}

	if conf.HealthCheckConnectPort < 0 || conf.HealthCheckConnectPort > 65535 {
		return fmt.Errorf("invalid server group HealthCheckConnectPort: %v", conf.HealthCheckConnectPort)
	}
	if !strings.EqualFold(conf.HealthCheckHttpVersion, util.ServerGroupHealthCheckHttpVersion10) &&
		!strings.EqualFold(conf.HealthCheckHttpVersion, util.ServerGroupHealthCheckHttpVersion11) {
		return fmt.Errorf("invalid server group HealthCheckHttpVersion: %v", conf.HealthCheckHttpVersion)
	}
	if conf.HealthCheckInterval < 1 || conf.HealthCheckInterval > 50 {
		return fmt.Errorf("invalid server group HealthCheckInterval: %v", conf.HealthCheckInterval)
	}
	if !strings.EqualFold(conf.HealthCheckMethod, util.ServerGroupHealthCheckMethodGET) &&
		!strings.EqualFold(conf.HealthCheckMethod, util.ServerGroupHealthCheckMethodHEAD) {
		return fmt.Errorf("invalid server group HealthCheckMethod: %v", conf.HealthCheckMethod)
	}
	if !strings.EqualFold(conf.HealthCheckProtocol, util.ServerGroupHealthCheckProtocolHTTP) &&
		!strings.EqualFold(conf.HealthCheckProtocol, util.ServerGroupHealthCheckProtocolHTTPS) {
		return fmt.Errorf("invalid server group HealthCheckProtocol: %v", conf.HealthCheckProtocol)
	}
	if conf.HealthCheckTimeout < 1 || conf.HealthCheckTimeout > 300 {
		return fmt.Errorf("invalid server group HealthCheckTimeout: %v", conf.HealthCheckTimeout)
	}
	if conf.HealthyThreshold < 2 || conf.HealthyThreshold > 10 {
		return fmt.Errorf("invalid server group HealthyThreshold: %v", conf.HealthyThreshold)
	}
	if conf.UnhealthyThreshold < 2 || conf.UnhealthyThreshold > 10 {
		return fmt.Errorf("invalid server group UnhealthyThreshold: %v", conf.UnhealthyThreshold)
	}

	return nil
}

func checkStickySessionConfigValid(conf alb.StickySessionConfig) error {
	if !conf.StickySessionEnabled {
		return nil
	}

	if !strings.EqualFold(conf.StickySessionType, util.ServerGroupStickySessionTypeInsert) &&
		!strings.EqualFold(conf.StickySessionType, util.ServerGroupStickySessionTypeServer) {
		return fmt.Errorf("invalid server group StickySessionType: %v", conf.StickySessionType)
	}

	if strings.EqualFold(conf.StickySessionType, util.ServerGroupStickySessionTypeInsert) {
		if conf.CookieTimeout < 1 || conf.CookieTimeout > 86400 {
			return fmt.Errorf("invalid server group CookieTimeout: %v", conf.CookieTimeout)
		}
	}

	if strings.EqualFold(conf.StickySessionType, util.ServerGroupStickySessionTypeServer) {
		if len(conf.Cookie) == 0 {
			return fmt.Errorf("invalid server group Cookie: %v", conf.Cookie)
		}
	}

	return nil
}

func isServerGroupResourceInUseError(err error) bool {
	if strings.Contains(err.Error(), "ResourceInUse.ServerGroup") ||
		strings.Contains(err.Error(), "IncorrectStatus.ServerGroup") {
		return true
	}
	return false
}

func isServerGroupSchedulerValid(scheduler string) bool {
	if strings.EqualFold(scheduler, util.ServerGroupSchedulerWrr) ||
		strings.EqualFold(scheduler, util.ServerGroupSchedulerWlc) ||
		strings.EqualFold(scheduler, util.ServerGroupSchedulerSch) {
		return true
	}
	return false
}
func isServerGroupProtocolValid(protocol string) bool {
	if strings.EqualFold(protocol, util.ServerGroupProtocolHTTP) ||
		strings.EqualFold(protocol, util.ServerGroupProtocolHTTPS) {
		return true
	}
	return false
}

func buildReServerGroupStatus(serverGroupID string) alb.ServerGroupStatus {
	return alb.ServerGroupStatus{
		ServerGroupID: serverGroupID,
	}
}

func transSDKHealthCheckConfigToCreateSGP(conf alb.HealthCheckConfig) *albsdk.CreateServerGroupHealthCheckConfig {
	if !conf.HealthCheckEnabled {
		return &albsdk.CreateServerGroupHealthCheckConfig{
			HealthCheckEnabled: strconv.FormatBool(conf.HealthCheckEnabled),
		}
	}

	return &albsdk.CreateServerGroupHealthCheckConfig{
		HealthCheckCodes:               &conf.HealthCheckHttpCodes,
		HealthCheckEnabled:             strconv.FormatBool(conf.HealthCheckEnabled),
		HealthCheckTimeout:             strconv.Itoa(conf.HealthCheckTimeout),
		HealthCheckMethod:              conf.HealthCheckMethod,
		HealthCheckHost:                conf.HealthCheckHost,
		HealthCheckProtocol:            conf.HealthCheckProtocol,
		UnhealthyThreshold:             strconv.Itoa(conf.UnhealthyThreshold),
		HealthyThreshold:               strconv.Itoa(conf.HealthyThreshold),
		HealthCheckTcpFastCloseEnabled: strconv.FormatBool(conf.HealthCheckTcpFastCloseEnabled),
		HealthCheckPath:                conf.HealthCheckPath,
		HealthCheckInterval:            strconv.Itoa(conf.HealthCheckInterval),
		HealthCheckHttpCodes:           &conf.HealthCheckHttpCodes,
		HealthCheckHttpVersion:         conf.HealthCheckHttpVersion,
		HealthCheckConnectPort:         strconv.Itoa(conf.HealthCheckConnectPort),
	}
}

func transSDKStickySessionConfigToCreateSGP(conf alb.StickySessionConfig) *albsdk.CreateServerGroupStickySessionConfig {
	if !conf.StickySessionEnabled {
		return &albsdk.CreateServerGroupStickySessionConfig{
			StickySessionEnabled: strconv.FormatBool(conf.StickySessionEnabled),
		}
	}

	return &albsdk.CreateServerGroupStickySessionConfig{
		StickySessionEnabled: strconv.FormatBool(conf.StickySessionEnabled),
		Cookie:               conf.Cookie,
		CookieTimeout:        strconv.Itoa(conf.CookieTimeout),
		StickySessionType:    conf.StickySessionType,
	}
}

func transSDKHealthCheckConfigToUpdateSGP(conf alb.HealthCheckConfig) *albsdk.UpdateServerGroupAttributeHealthCheckConfig {
	if !conf.HealthCheckEnabled {
		return &albsdk.UpdateServerGroupAttributeHealthCheckConfig{
			HealthCheckEnabled: strconv.FormatBool(conf.HealthCheckEnabled),
		}
	}

	return &albsdk.UpdateServerGroupAttributeHealthCheckConfig{
		HealthCheckCodes:               &conf.HealthCheckHttpCodes,
		HealthCheckEnabled:             strconv.FormatBool(conf.HealthCheckEnabled),
		HealthCheckTimeout:             strconv.Itoa(conf.HealthCheckTimeout),
		HealthCheckMethod:              conf.HealthCheckMethod,
		HealthCheckHost:                conf.HealthCheckHost,
		HealthCheckProtocol:            conf.HealthCheckProtocol,
		UnhealthyThreshold:             strconv.Itoa(conf.UnhealthyThreshold),
		HealthyThreshold:               strconv.Itoa(conf.HealthyThreshold),
		HealthCheckTcpFastCloseEnabled: strconv.FormatBool(conf.HealthCheckTcpFastCloseEnabled),
		HealthCheckPath:                conf.HealthCheckPath,
		HealthCheckInterval:            strconv.Itoa(conf.HealthCheckInterval),
		HealthCheckHttpCodes:           &conf.HealthCheckHttpCodes,
		HealthCheckHttpVersion:         conf.HealthCheckHttpVersion,
		HealthCheckConnectPort:         strconv.Itoa(conf.HealthCheckConnectPort),
	}
}

func transSDKStickySessionConfigToUpdateSGP(conf alb.StickySessionConfig) *albsdk.UpdateServerGroupAttributeStickySessionConfig {
	if !conf.StickySessionEnabled {
		return &albsdk.UpdateServerGroupAttributeStickySessionConfig{
			StickySessionEnabled: strconv.FormatBool(conf.StickySessionEnabled),
		}
	}

	return &albsdk.UpdateServerGroupAttributeStickySessionConfig{
		StickySessionEnabled: strconv.FormatBool(conf.StickySessionEnabled),
		Cookie:               conf.Cookie,
		CookieTimeout:        strconv.Itoa(conf.CookieTimeout),
		StickySessionType:    conf.StickySessionType,
	}
}
