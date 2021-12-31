package alb

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/pkg/errors"
)

func (m *ALBProvider) CreateALBListenerRule(ctx context.Context, resLR *alb.ListenerRule) (alb.ListenerRuleStatus, error) {
	traceID := ctx.Value(util.TraceID)

	createRuleReq, err := buildSDKCreateListenerRuleRequest(resLR.Spec)
	if err != nil {
		return alb.ListenerRuleStatus{}, err
	}

	var createRuleResp *albsdk.CreateRuleResponse
	if err := util.RetryImmediateOnError(m.waitLSExistencePollInterval, m.waitLSExistenceTimeout, isVipStatusNotSupportError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("creating rule",
			"stackID", resLR.Stack().StackID(),
			"resourceID", resLR.ID(),
			"listenerID", createRuleReq.ListenerId,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.CreateALBRule)
		createRuleResp, err = m.auth.ALB.CreateRule(createRuleReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("creating rule",
				"stackID", resLR.Stack().StackID(),
				"resourceID", resLR.ID(),
				"listenerID", createRuleReq.ListenerId,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.CreateALBRule)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("created rule",
			"stackID", resLR.Stack().StackID(),
			"resourceID", resLR.ID(),
			"listenerID", createRuleReq.ListenerId,
			"traceID", traceID,
			"ruleID", createRuleResp.RuleId,
			"requestID", createRuleResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.CreateALBRule)
		return nil
	}); err != nil {
		return alb.ListenerRuleStatus{}, errors.Wrap(err, "failed to create listener rule")
	}

	return buildResListenerRuleStatus(createRuleResp.RuleId), nil
}

func (m *ALBProvider) UpdateALBListenerRule(ctx context.Context, resLR *alb.ListenerRule, sdkLR *albsdk.Rule) (alb.ListenerRuleStatus, error) {
	updateAnalyzer := new(ListenerRuleUpdateAnalyzer)
	if err := updateAnalyzer.analysis(ctx, resLR, sdkLR); err != nil {
		return alb.ListenerRuleStatus{}, err
	}
	if !updateAnalyzer.needUpdate {
		return buildResListenerRuleStatus(sdkLR.RuleId), nil
	}

	_, err := m.updateALBRuleAttribute(ctx, resLR, sdkLR, *updateAnalyzer)
	if err != nil {
		return alb.ListenerRuleStatus{}, err
	}

	return buildResListenerRuleStatus(sdkLR.RuleId), nil
}

func (m *ALBProvider) DeleteALBListenerRule(ctx context.Context, sdkLRId string) error {
	traceID := ctx.Value(util.TraceID)

	deleteRuleReq := albsdk.CreateDeleteRuleRequest()
	deleteRuleReq.RuleId = sdkLRId

	if err := util.RetryImmediateOnError(m.waitLSExistencePollInterval, m.waitLSExistenceTimeout, isVipStatusNotSupportError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("deleting rule",
			"ruleID", sdkLRId,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.DeleteALBRule)
		deleteRuleResp, err := m.auth.ALB.DeleteRule(deleteRuleReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("deleting rule",
				"ruleID", sdkLRId,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.DeleteALBRule)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("deleted rule",
			"ruleID", sdkLRId,
			"traceID", traceID,
			"requestID", deleteRuleResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.DeleteALBRule)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to delete listener rule")
	}

	return nil
}

func (m *ALBProvider) ListALBListenerRules(ctx context.Context, lsID string) ([]albsdk.Rule, error) {
	traceID := ctx.Value(util.TraceID)

	if len(lsID) == 0 {
		return nil, fmt.Errorf("invalid listener id: %s for listing rules", lsID)
	}

	var (
		nextToken string
		rules     []albsdk.Rule
	)

	listRuleReq := albsdk.CreateListRulesRequest()
	listRuleReq.ListenerId = lsID

	for {
		listRuleReq.NextToken = nextToken

		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("listing rules",
			"listenerID", lsID,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.ListALBRules)
		listRuleResp, err := m.auth.ALB.ListRules(listRuleReq)
		if err != nil {
			return nil, err
		}
		m.logger.V(util.MgrLogLevel).Info("listed rules",
			"listenerID", lsID,
			"traceID", traceID,
			"requestID", listRuleResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListALBRules)

		rules = append(rules, listRuleResp.Rules...)

		if listRuleResp.NextToken == "" {
			break
		} else {
			nextToken = listRuleResp.NextToken
		}
	}

	return rules, nil
}

func (m *ALBProvider) updateALBRuleAttribute(ctx context.Context, resLR *alb.ListenerRule, sdkLR *albsdk.Rule, judger ListenerRuleUpdateAnalyzer) (*albsdk.UpdateRuleAttributeResponse, error) {
	traceID := ctx.Value(util.TraceID)

	ruleReq := albsdk.CreateUpdateRuleAttributeRequest()
	ruleReq.RuleId = sdkLR.RuleId

	if judger.ruleNameNeedUpdate {
		m.logger.V(util.MgrLogLevel).Info("RuleName update",
			"res", resLR.Spec.RuleName,
			"sdk", sdkLR.RuleName,
			"ruleID", sdkLR.RuleId,
			"traceID", traceID)
		ruleReq.RuleName = resLR.Spec.RuleName
	}
	if judger.ruleActionsNeedUpdate {
		m.logger.V(util.MgrLogLevel).Info("Actions update",
			"res", resLR.Spec.RuleActions,
			"sdk", sdkLR.RuleActions,
			"ruleID", sdkLR.RuleId,
			"traceID", traceID)
		actions, err := transModelActionsToSDKUpdateRule(resLR.Spec.RuleActions)
		if err != nil {
			return nil, err
		}
		ruleReq.RuleActions = actions
	}
	if judger.ruleConditionsNeedUpdate {
		m.logger.V(util.MgrLogLevel).Info("Conditions update",
			"res", resLR.Spec.RuleConditions,
			"sdk", sdkLR.RuleConditions,
			"ruleID", sdkLR.RuleId,
			"traceID", traceID)
		ruleReq.RuleConditions = transSDKConditionsToUpdateRule(resLR.Spec.RuleConditions)
	}
	if judger.priorityNeedUpdate {
		m.logger.V(util.MgrLogLevel).Info("Priority update",
			"res", resLR.Spec.Priority,
			"sdk", sdkLR.Priority,
			"ruleID", sdkLR.RuleId,
			"traceID", traceID)
		ruleReq.Priority = requests.NewInteger(resLR.Spec.Priority)
	}

	var updateRuleResp *albsdk.UpdateRuleAttributeResponse
	if err := util.RetryImmediateOnError(m.waitLSExistencePollInterval, m.waitLSExistenceTimeout, isVipStatusNotSupportError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("updating rule attribute",
			"stackID", resLR.Stack().StackID(),
			"resourceID", resLR.ID(),
			"traceID", traceID,
			"ruleID", sdkLR.RuleId,
			"startTime", startTime,
			util.Action, util.UpdateALBRuleAttribute)
		var err error
		updateRuleResp, err = m.auth.ALB.UpdateRuleAttribute(ruleReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("updating rule attribute",
				"stackID", resLR.Stack().StackID(),
				"resourceID", resLR.ID(),
				"traceID", traceID,
				"ruleID", sdkLR.RuleId,
				"error", err.Error(),
				"requestID", updateRuleResp.RequestId,
				"startTime", startTime,
				util.Action, util.UpdateALBRuleAttribute)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("updated rule attribute",
			"stackID", resLR.Stack().StackID(),
			"resourceID", resLR.ID(),
			"traceID", traceID,
			"ruleID", sdkLR.RuleId,
			"requestID", updateRuleResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.UpdateALBRuleAttribute)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to update listener rule")
	}

	return updateRuleResp, nil
}

func transModelServerGroupTuplesToSDKCreateRules(sgpTuples []alb.ServerGroupTuple) (*[]albsdk.CreateRulesRulesRuleActionsItemForwardGroupConfigServerGroupTuplesItem, error) {
	var createRulesServerGroupTuples []albsdk.CreateRulesRulesRuleActionsItemForwardGroupConfigServerGroupTuplesItem
	if len(sgpTuples) != 0 {
		createRulesServerGroupTuples = make([]albsdk.CreateRulesRulesRuleActionsItemForwardGroupConfigServerGroupTuplesItem, 0)
		for _, m := range sgpTuples {
			serverGroupID, err := m.ServerGroupID.Resolve(context.Background())
			if err != nil {
				return nil, err
			}
			createRulesServerGroupTuples = append(createRulesServerGroupTuples, albsdk.CreateRulesRulesRuleActionsItemForwardGroupConfigServerGroupTuplesItem{
				ServerGroupId: serverGroupID,
				Weight:        strconv.Itoa(m.Weight),
			})
		}
	}
	return &createRulesServerGroupTuples, nil
}
func transModelServerGroupTuplesToSDKUpdateRules(sgpTuples []alb.ServerGroupTuple) (*[]albsdk.UpdateRulesAttributeRulesRuleActionsItemForwardGroupConfigServerGroupTuplesItem, error) {
	var updateRulesServerGroupTuples []albsdk.UpdateRulesAttributeRulesRuleActionsItemForwardGroupConfigServerGroupTuplesItem
	if len(sgpTuples) != 0 {
		updateRulesServerGroupTuples = make([]albsdk.UpdateRulesAttributeRulesRuleActionsItemForwardGroupConfigServerGroupTuplesItem, 0)
		for _, m := range sgpTuples {
			serverGroupID, err := m.ServerGroupID.Resolve(context.Background())
			if err != nil {
				return nil, err
			}
			updateRulesServerGroupTuples = append(updateRulesServerGroupTuples, albsdk.UpdateRulesAttributeRulesRuleActionsItemForwardGroupConfigServerGroupTuplesItem{
				ServerGroupId: serverGroupID,
				Weight:        strconv.Itoa(m.Weight),
			})
		}
	}
	return &updateRulesServerGroupTuples, nil
}
func transModelServerGroupTuplesToSDKCreateRule(sgpTuples []alb.ServerGroupTuple) (*[]albsdk.CreateRuleRuleActionsForwardGroupConfigServerGroupTuplesItem, error) {
	var createRuleServerGroupTuples []albsdk.CreateRuleRuleActionsForwardGroupConfigServerGroupTuplesItem
	if len(sgpTuples) != 0 {
		createRuleServerGroupTuples = make([]albsdk.CreateRuleRuleActionsForwardGroupConfigServerGroupTuplesItem, 0)
		for _, m := range sgpTuples {
			serverGroupID, err := m.ServerGroupID.Resolve(context.Background())
			if err != nil {
				return nil, err
			}
			createRuleServerGroupTuples = append(createRuleServerGroupTuples, albsdk.CreateRuleRuleActionsForwardGroupConfigServerGroupTuplesItem{
				ServerGroupId: serverGroupID,
				Weight:        strconv.Itoa(m.Weight),
			})
		}
	}
	return &createRuleServerGroupTuples, nil
}
func transModelServerGroupTuplesToSDKUpdateRule(sgpTuples []alb.ServerGroupTuple) (*[]albsdk.UpdateRuleAttributeRuleActionsForwardGroupConfigServerGroupTuplesItem, error) {
	var updateRuleServerGroupTuples []albsdk.UpdateRuleAttributeRuleActionsForwardGroupConfigServerGroupTuplesItem
	if len(sgpTuples) != 0 {
		updateRuleServerGroupTuples = make([]albsdk.UpdateRuleAttributeRuleActionsForwardGroupConfigServerGroupTuplesItem, 0)
		for _, m := range sgpTuples {
			serverGroupID, err := m.ServerGroupID.Resolve(context.Background())
			if err != nil {
				return nil, err
			}
			updateRuleServerGroupTuples = append(updateRuleServerGroupTuples, albsdk.UpdateRuleAttributeRuleActionsForwardGroupConfigServerGroupTuplesItem{
				ServerGroupId: serverGroupID,
				Weight:        strconv.Itoa(m.Weight),
			})
		}
	}
	return &updateRuleServerGroupTuples, nil
}
func transModelServerGroupTuplesToSDKRule(sgpTuples []alb.ServerGroupTuple) (*[]albsdk.ServerGroupTuple, error) {
	var sdkServerGroupTuples []albsdk.ServerGroupTuple
	if len(sgpTuples) != 0 {
		sdkServerGroupTuples = make([]albsdk.ServerGroupTuple, 0)
		for _, m := range sgpTuples {
			serverGroupID, err := m.ServerGroupID.Resolve(context.Background())
			if err != nil {
				return nil, err
			}
			sdkServerGroupTuples = append(sdkServerGroupTuples, albsdk.ServerGroupTuple{
				ServerGroupId: serverGroupID,
				Weight:        m.Weight,
			})
		}
	}
	return &sdkServerGroupTuples, nil
}
func transSDKFixedResponseConfigToCreateRules(fixedRespConf alb.FixedResponseConfig) albsdk.CreateRulesRulesRuleActionsItemFixedResponseConfig {
	return albsdk.CreateRulesRulesRuleActionsItemFixedResponseConfig{
		HttpCode:    fixedRespConf.HttpCode,
		Content:     fixedRespConf.Content,
		ContentType: fixedRespConf.ContentType,
	}
}
func transSDKFixedResponseConfigToUpdateRules(fixedRespConf alb.FixedResponseConfig) albsdk.UpdateRulesAttributeRulesRuleActionsItemFixedResponseConfig {
	return albsdk.UpdateRulesAttributeRulesRuleActionsItemFixedResponseConfig{
		HttpCode:    fixedRespConf.HttpCode,
		Content:     fixedRespConf.Content,
		ContentType: fixedRespConf.ContentType,
	}
}
func transSDKFixedResponseConfigToCreateRule(fixedRespConf alb.FixedResponseConfig) albsdk.CreateRuleRuleActionsFixedResponseConfig {
	return albsdk.CreateRuleRuleActionsFixedResponseConfig{
		HttpCode:    fixedRespConf.HttpCode,
		Content:     fixedRespConf.Content,
		ContentType: fixedRespConf.ContentType,
	}
}
func transSDKFixedResponseConfigToUpdateRule(fixedRespConf alb.FixedResponseConfig) albsdk.UpdateRuleAttributeRuleActionsFixedResponseConfig {
	return albsdk.UpdateRuleAttributeRuleActionsFixedResponseConfig{
		HttpCode:    fixedRespConf.HttpCode,
		Content:     fixedRespConf.Content,
		ContentType: fixedRespConf.ContentType,
	}
}
func transSDKRedirectConfigToCreateRules(redirectConf alb.RedirectConfig) albsdk.CreateRulesRulesRuleActionsItemRedirectConfig {
	return albsdk.CreateRulesRulesRuleActionsItemRedirectConfig{
		Path:     redirectConf.Path,
		Protocol: redirectConf.Protocol,
		Port:     redirectConf.Port,
		Query:    redirectConf.Query,
		Host:     redirectConf.Host,
		HttpCode: redirectConf.HttpCode,
	}
}
func transSDKRedirectConfigToUpdateRules(redirectConf alb.RedirectConfig) albsdk.UpdateRulesAttributeRulesRuleActionsItemRedirectConfig {
	return albsdk.UpdateRulesAttributeRulesRuleActionsItemRedirectConfig{
		Path:     redirectConf.Path,
		Protocol: redirectConf.Protocol,
		Port:     redirectConf.Port,
		Query:    redirectConf.Query,
		Host:     redirectConf.Host,
		HttpCode: redirectConf.HttpCode,
	}
}
func transSDKRedirectConfigToCreateRule(redirectConf alb.RedirectConfig) albsdk.CreateRuleRuleActionsRedirectConfig {
	return albsdk.CreateRuleRuleActionsRedirectConfig{
		Path:     redirectConf.Path,
		Protocol: redirectConf.Protocol,
		Port:     redirectConf.Port,
		Query:    redirectConf.Query,
		Host:     redirectConf.Host,
		HttpCode: redirectConf.HttpCode,
	}
}
func transSDKRedirectConfigToUpdateRule(redirectConf alb.RedirectConfig) albsdk.UpdateRuleAttributeRuleActionsRedirectConfig {
	return albsdk.UpdateRuleAttributeRuleActionsRedirectConfig{
		Path:     redirectConf.Path,
		Protocol: redirectConf.Protocol,
		Port:     redirectConf.Port,
		Query:    redirectConf.Query,
		Host:     redirectConf.Host,
		HttpCode: redirectConf.HttpCode,
	}
}
func transModelForwardActionConfigToSDKCreateRules(forwardActionConf alb.ForwardActionConfig) (*albsdk.CreateRulesRulesRuleActionsItemForwardGroupConfig, error) {
	sdkSGPs, err := transModelServerGroupTuplesToSDKCreateRules(forwardActionConf.ServerGroups)
	if err != nil {
		return nil, err
	}
	r := &albsdk.CreateRulesRulesRuleActionsItemForwardGroupConfig{
		ServerGroupTuples: sdkSGPs,
	}
	if forwardActionConf.ServerGroupStickySession != nil {
		r.ServerGroupStickySession = albsdk.CreateRulesRulesRuleActionsItemForwardGroupConfigServerGroupStickySession{
			Enabled: strconv.FormatBool(forwardActionConf.ServerGroupStickySession.Enabled),
			Timeout: strconv.Itoa(forwardActionConf.ServerGroupStickySession.Timeout),
		}
	}
	return r, nil
}
func transModelForwardActionConfigToSDKUpdateRules(forwardActionConf alb.ForwardActionConfig) (*albsdk.UpdateRulesAttributeRulesRuleActionsItemForwardGroupConfig, error) {
	sdkSGPs, err := transModelServerGroupTuplesToSDKUpdateRules(forwardActionConf.ServerGroups)
	if err != nil {
		return nil, err
	}
	r := &albsdk.UpdateRulesAttributeRulesRuleActionsItemForwardGroupConfig{
		ServerGroupTuples: sdkSGPs,
	}
	if forwardActionConf.ServerGroupStickySession != nil {
		r.ServerGroupStickySession = albsdk.UpdateRulesAttributeRulesRuleActionsItemForwardGroupConfigServerGroupStickySession{
			Enabled: strconv.FormatBool(forwardActionConf.ServerGroupStickySession.Enabled),
			Timeout: strconv.Itoa(forwardActionConf.ServerGroupStickySession.Timeout),
		}
	}
	return r, nil
}
func transModelForwardActionConfigToSDKCreateRule(forwardActionConf alb.ForwardActionConfig) (*albsdk.CreateRuleRuleActionsForwardGroupConfig, error) {
	sdkSGPs, err := transModelServerGroupTuplesToSDKCreateRule(forwardActionConf.ServerGroups)
	if err != nil {
		return nil, err
	}
	r := &albsdk.CreateRuleRuleActionsForwardGroupConfig{
		ServerGroupTuples: sdkSGPs,
	}
	if forwardActionConf.ServerGroupStickySession != nil {
		r.ServerGroupStickySession = albsdk.CreateRuleRuleActionsForwardGroupConfigServerGroupStickySession{
			Enabled: strconv.FormatBool(forwardActionConf.ServerGroupStickySession.Enabled),
			Timeout: strconv.Itoa(forwardActionConf.ServerGroupStickySession.Timeout),
		}
	}
	return r, nil
}
func transModelForwardActionConfigToSDKUpdateRule(forwardActionConf alb.ForwardActionConfig) (*albsdk.UpdateRuleAttributeRuleActionsForwardGroupConfig, error) {
	sdkSGPs, err := transModelServerGroupTuplesToSDKUpdateRule(forwardActionConf.ServerGroups)
	if err != nil {
		return nil, err
	}
	r := &albsdk.UpdateRuleAttributeRuleActionsForwardGroupConfig{
		ServerGroupTuples: sdkSGPs,
	}
	if forwardActionConf.ServerGroupStickySession != nil {
		r.ServerGroupStickySession = albsdk.UpdateRuleAttributeRuleActionsForwardGroupConfigServerGroupStickySession{
			Enabled: strconv.FormatBool(forwardActionConf.ServerGroupStickySession.Enabled),
			Timeout: strconv.Itoa(forwardActionConf.ServerGroupStickySession.Timeout),
		}
	}
	return r, nil
}
func transModelForwardActionConfigToSDKRule(forwardActionConf alb.ForwardActionConfig) (*albsdk.ForwardGroupConfigInListRules, error) {
	sdkSGPs, err := transModelServerGroupTuplesToSDKRule(forwardActionConf.ServerGroups)
	if err != nil {
		return nil, err
	}
	r := &albsdk.ForwardGroupConfigInListRules{
		ServerGroupTuples: *sdkSGPs,
	}
	if forwardActionConf.ServerGroupStickySession != nil {
		r.ServerGroupStickySession = albsdk.ServerGroupStickySession{
			Enabled: forwardActionConf.ServerGroupStickySession.Enabled,
			Timeout: forwardActionConf.ServerGroupStickySession.Timeout,
		}
	}
	return r, nil
}

func transModelActionToSDKCreateRules(action alb.Action) (*albsdk.CreateRulesRulesRuleActionsItem, error) {
	sdkObj := &albsdk.CreateRulesRulesRuleActionsItem{}
	sdkObj.Type = action.Type
	sdkObj.Order = strconv.Itoa(action.Order)
	switch sdkObj.Type {
	case util.RuleActionTypeFixedResponse:
		sdkObj.FixedResponseConfig = transSDKFixedResponseConfigToCreateRules(*action.FixedResponseConfig)
	case util.RuleActionTypeRedirect:
		sdkObj.RedirectConfig = transSDKRedirectConfigToCreateRules(*action.RedirectConfig)
	case util.RuleActionTypeForward:
		forwardActionConfig, err := transModelForwardActionConfigToSDKCreateRules(*action.ForwardConfig)
		if err != nil {
			return nil, err
		}
		sdkObj.ForwardGroupConfig = *forwardActionConfig
	}
	return sdkObj, nil
}
func transModelActionToSDKUpdateRules(action alb.Action) (*albsdk.UpdateRulesAttributeRulesRuleActionsItem, error) {
	sdkObj := &albsdk.UpdateRulesAttributeRulesRuleActionsItem{}
	sdkObj.Type = action.Type
	sdkObj.Order = strconv.Itoa(action.Order)
	switch sdkObj.Type {
	case util.RuleActionTypeFixedResponse:
		sdkObj.FixedResponseConfig = transSDKFixedResponseConfigToUpdateRules(*action.FixedResponseConfig)
	case util.RuleActionTypeRedirect:
		sdkObj.RedirectConfig = transSDKRedirectConfigToUpdateRules(*action.RedirectConfig)
	case util.RuleActionTypeForward:
		forwardActionConfig, err := transModelForwardActionConfigToSDKUpdateRules(*action.ForwardConfig)
		if err != nil {
			return nil, err
		}
		sdkObj.ForwardGroupConfig = *forwardActionConfig
	}
	return sdkObj, nil
}
func transModelActionToSDKCreateRule(action alb.Action) (*albsdk.CreateRuleRuleActions, error) {
	sdkObj := &albsdk.CreateRuleRuleActions{}
	sdkObj.Type = action.Type
	sdkObj.Order = strconv.Itoa(action.Order)
	switch sdkObj.Type {
	case util.RuleActionTypeFixedResponse:
		sdkObj.FixedResponseConfig = transSDKFixedResponseConfigToCreateRule(*action.FixedResponseConfig)
	case util.RuleActionTypeRedirect:
		sdkObj.RedirectConfig = transSDKRedirectConfigToCreateRule(*action.RedirectConfig)
	case util.RuleActionTypeForward:
		forwardActionConfig, err := transModelForwardActionConfigToSDKCreateRule(*action.ForwardConfig)
		if err != nil {
			return nil, err
		}
		sdkObj.ForwardGroupConfig = *forwardActionConfig
	}
	return sdkObj, nil
}
func transModelActionToSDKUpdateRule(action alb.Action) (*albsdk.UpdateRuleAttributeRuleActions, error) {
	sdkObj := &albsdk.UpdateRuleAttributeRuleActions{}
	sdkObj.Type = action.Type
	sdkObj.Order = strconv.Itoa(action.Order)
	switch sdkObj.Type {
	case util.RuleActionTypeFixedResponse:
		sdkObj.FixedResponseConfig = transSDKFixedResponseConfigToUpdateRule(*action.FixedResponseConfig)
	case util.RuleActionTypeRedirect:
		sdkObj.RedirectConfig = transSDKRedirectConfigToUpdateRule(*action.RedirectConfig)
	case util.RuleActionTypeForward:
		forwardActionConfig, err := transModelForwardActionConfigToSDKUpdateRule(*action.ForwardConfig)
		if err != nil {
			return nil, err
		}
		sdkObj.ForwardGroupConfig = *forwardActionConfig
	}
	return sdkObj, nil
}
func transModelActionsToSDKCreateRules(actions []alb.Action) (*[]albsdk.CreateRulesRulesRuleActionsItem, error) {
	var createRulesActions []albsdk.CreateRulesRulesRuleActionsItem
	if len(actions) != 0 {
		createRulesActions = make([]albsdk.CreateRulesRulesRuleActionsItem, 0)
		for index, m := range actions {
			m.Order = index + 1
			action, err := transModelActionToSDKCreateRules(m)
			if err != nil {
				return nil, err
			}
			createRulesActions = append(createRulesActions, *action)
		}
	}
	return &createRulesActions, nil
}
func transModelActionsToSDKUpdateRules(actions []alb.Action) (*[]albsdk.UpdateRulesAttributeRulesRuleActionsItem, error) {
	var updateRulesActions []albsdk.UpdateRulesAttributeRulesRuleActionsItem
	if len(actions) != 0 {
		updateRulesActions = make([]albsdk.UpdateRulesAttributeRulesRuleActionsItem, 0)
		for index, m := range actions {
			m.Order = index + 1
			action, err := transModelActionToSDKUpdateRules(m)
			if err != nil {
				return nil, err
			}
			updateRulesActions = append(updateRulesActions, *action)
		}
	}
	return &updateRulesActions, nil
}
func transModelActionsToSDKCreateRule(actions []alb.Action) (*[]albsdk.CreateRuleRuleActions, error) {
	var createRuleActions []albsdk.CreateRuleRuleActions
	if len(actions) != 0 {
		createRuleActions = make([]albsdk.CreateRuleRuleActions, 0)
		for index, m := range actions {
			m.Order = index + 1
			createRuleAction, err := transModelActionToSDKCreateRule(m)
			if err != nil {
				return nil, err
			}
			createRuleActions = append(createRuleActions, *createRuleAction)
		}
	}
	return &createRuleActions, nil
}
func transModelActionsToSDKUpdateRule(actions []alb.Action) (*[]albsdk.UpdateRuleAttributeRuleActions, error) {
	var updateRuleActions []albsdk.UpdateRuleAttributeRuleActions
	if len(actions) != 0 {
		updateRuleActions = make([]albsdk.UpdateRuleAttributeRuleActions, 0)
		for index, m := range actions {
			m.Order = index + 1
			updateRuleAction, err := transModelActionToSDKUpdateRule(m)
			if err != nil {
				return nil, err
			}
			updateRuleActions = append(updateRuleActions, *updateRuleAction)
		}
	}

	return &updateRuleActions, nil
}
func transModelActionToSDK(action alb.Action) (*albsdk.Action, error) {
	sdkObj := &albsdk.Action{}
	sdkObj.Type = action.Type
	sdkObj.Order = action.Order
	switch sdkObj.Type {
	case util.RuleActionTypeFixedResponse:
		sdkObj.FixedResponseConfig = albsdk.FixedResponseConfig{
			HttpCode:    action.FixedResponseConfig.HttpCode,
			Content:     action.FixedResponseConfig.Content,
			ContentType: action.FixedResponseConfig.ContentType,
		}
	case util.RuleActionTypeRedirect:
		sdkObj.RedirectConfig = albsdk.RedirectConfig{
			HttpCode: action.RedirectConfig.HttpCode,
			Host:     action.RedirectConfig.Host,
			Path:     action.RedirectConfig.Path,
			Protocol: action.RedirectConfig.Protocol,
			Port:     action.RedirectConfig.Port,
			Query:    action.RedirectConfig.Query,
		}
	case util.RuleActionTypeForward:
		forwardActionConfig, err := transModelForwardActionConfigToSDKRule(*action.ForwardConfig)
		if err != nil {
			return nil, err
		}
		sdkObj.ForwardGroupConfig = *forwardActionConfig
	}
	return sdkObj, nil
}
func transModelActionsToSDK(actions []alb.Action) (*[]albsdk.Action, error) {
	var sdkActions []albsdk.Action
	if len(actions) != 0 {
		sdkActions = make([]albsdk.Action, 0)
		for index, m := range actions {
			m.Order = index + 1
			ruleAction, err := transModelActionToSDK(m)
			if err != nil {
				return nil, err
			}
			sdkActions = append(sdkActions, *ruleAction)
		}
	}

	return &sdkActions, nil
}

func transSDKQueryStringConfigValueToCreateRule(value alb.Value) *albsdk.CreateRuleRuleConditionsQueryStringConfigValuesItem {
	return &albsdk.CreateRuleRuleConditionsQueryStringConfigValuesItem{
		Value: value.Value,
		Key:   value.Key,
	}
}
func transSDKQueryStringConfigValueToUpdateRule(value alb.Value) *albsdk.UpdateRuleAttributeRuleConditionsQueryStringConfigValuesItem {
	return &albsdk.UpdateRuleAttributeRuleConditionsQueryStringConfigValuesItem{
		Value: value.Value,
		Key:   value.Key,
	}
}
func transSDKQueryStringConfigValuesToCreateRule(values []alb.Value) *[]albsdk.CreateRuleRuleConditionsQueryStringConfigValuesItem {
	r := make([]albsdk.CreateRuleRuleConditionsQueryStringConfigValuesItem, 0)
	for _, v := range values {
		r = append(r, *transSDKQueryStringConfigValueToCreateRule(v))
	}
	return &r
}
func transSDKQueryStringConfigValuesToUpdateRule(values []alb.Value) *[]albsdk.UpdateRuleAttributeRuleConditionsQueryStringConfigValuesItem {
	r := make([]albsdk.UpdateRuleAttributeRuleConditionsQueryStringConfigValuesItem, 0)
	for _, v := range values {
		r = append(r, *transSDKQueryStringConfigValueToUpdateRule(v))
	}
	return &r
}
func transSDKQueryStringConfigToCreateRule(queryStringConf alb.QueryStringConfig) albsdk.CreateRuleRuleConditionsQueryStringConfig {
	return albsdk.CreateRuleRuleConditionsQueryStringConfig{
		Values: transSDKQueryStringConfigValuesToCreateRule(queryStringConf.Values),
	}
}
func transSDKQueryStringConfigToUpdateRule(queryStringConf alb.QueryStringConfig) albsdk.UpdateRuleAttributeRuleConditionsQueryStringConfig {
	return albsdk.UpdateRuleAttributeRuleConditionsQueryStringConfig{
		Values: transSDKQueryStringConfigValuesToUpdateRule(queryStringConf.Values),
	}
}
func transSDKQueryStringConfigValuesToCreateRules(values []alb.Value) *[]albsdk.CreateRulesRulesRuleConditionsItemQueryStringConfigValuesItem {
	r := make([]albsdk.CreateRulesRulesRuleConditionsItemQueryStringConfigValuesItem, 0)
	for _, v := range values {
		r = append(r, *transSDKQueryStringConfigValueToCreateRules(v))
	}
	return &r
}
func transSDKQueryStringConfigValuesToUpdateRules(values []alb.Value) *[]albsdk.UpdateRulesAttributeRulesRuleConditionsItemQueryStringConfigValuesItem {
	r := make([]albsdk.UpdateRulesAttributeRulesRuleConditionsItemQueryStringConfigValuesItem, 0)
	for _, v := range values {
		r = append(r, *transSDKQueryStringConfigValueToUpdateRules(v))
	}
	return &r
}
func transSDKQueryStringConfigValueToCreateRules(value alb.Value) *albsdk.CreateRulesRulesRuleConditionsItemQueryStringConfigValuesItem {
	return &albsdk.CreateRulesRulesRuleConditionsItemQueryStringConfigValuesItem{
		Value: value.Value,
		Key:   value.Key,
	}
}
func transSDKQueryStringConfigValueToUpdateRules(value alb.Value) *albsdk.UpdateRulesAttributeRulesRuleConditionsItemQueryStringConfigValuesItem {
	return &albsdk.UpdateRulesAttributeRulesRuleConditionsItemQueryStringConfigValuesItem{
		Value: value.Value,
		Key:   value.Key,
	}
}
func transSDKQueryStringConfigToCreateRules(queryStringConf alb.QueryStringConfig) albsdk.CreateRulesRulesRuleConditionsItemQueryStringConfig {
	return albsdk.CreateRulesRulesRuleConditionsItemQueryStringConfig{
		Values: transSDKQueryStringConfigValuesToCreateRules(queryStringConf.Values),
	}
}
func transSDKQueryStringConfigToUpdateRules(queryStringConf alb.QueryStringConfig) albsdk.UpdateRulesAttributeRulesRuleConditionsItemQueryStringConfig {
	return albsdk.UpdateRulesAttributeRulesRuleConditionsItemQueryStringConfig{
		Values: transSDKQueryStringConfigValuesToUpdateRules(queryStringConf.Values),
	}
}

func transSDKCookieConfigValueToCreateRules(value alb.Value) *albsdk.CreateRulesRulesRuleConditionsItemCookieConfigValuesItem {
	return &albsdk.CreateRulesRulesRuleConditionsItemCookieConfigValuesItem{
		Value: value.Value,
		Key:   value.Key,
	}
}
func transSDKCookieConfigValueToUpdateRules(value alb.Value) *albsdk.UpdateRulesAttributeRulesRuleConditionsItemCookieConfigValuesItem {
	return &albsdk.UpdateRulesAttributeRulesRuleConditionsItemCookieConfigValuesItem{
		Value: value.Value,
		Key:   value.Key,
	}
}
func transSDKCookieConfigValueToCreateRule(values alb.Value) *albsdk.CreateRuleRuleConditionsCookieConfigValuesItem {
	return &albsdk.CreateRuleRuleConditionsCookieConfigValuesItem{
		Value: values.Value,
		Key:   values.Key,
	}
}
func transSDKCookieConfigValueToUpdateRule(value alb.Value) *albsdk.UpdateRuleAttributeRuleConditionsCookieConfigValuesItem {
	return &albsdk.UpdateRuleAttributeRuleConditionsCookieConfigValuesItem{
		Value: value.Value,
		Key:   value.Key,
	}
}
func transSDKCookieConfigValuesToCreateRules(values []alb.Value) *[]albsdk.CreateRulesRulesRuleConditionsItemCookieConfigValuesItem {
	r := make([]albsdk.CreateRulesRulesRuleConditionsItemCookieConfigValuesItem, 0)
	for _, v := range values {
		r = append(r, *transSDKCookieConfigValueToCreateRules(v))
	}
	return &r
}
func transSDKCookieConfigValuesToUpdateRules(values []alb.Value) *[]albsdk.UpdateRulesAttributeRulesRuleConditionsItemCookieConfigValuesItem {
	r := make([]albsdk.UpdateRulesAttributeRulesRuleConditionsItemCookieConfigValuesItem, 0)
	for _, v := range values {
		r = append(r, *transSDKCookieConfigValueToUpdateRules(v))
	}
	return &r
}
func transSDKCookieConfigValuesToCreateRule(values []alb.Value) *[]albsdk.CreateRuleRuleConditionsCookieConfigValuesItem {
	r := make([]albsdk.CreateRuleRuleConditionsCookieConfigValuesItem, 0)
	for _, v := range values {
		r = append(r, *transSDKCookieConfigValueToCreateRule(v))
	}
	return &r
}
func transSDKCookieConfigValuesToUpdateRule(values []alb.Value) *[]albsdk.UpdateRuleAttributeRuleConditionsCookieConfigValuesItem {
	r := make([]albsdk.UpdateRuleAttributeRuleConditionsCookieConfigValuesItem, 0)
	for _, v := range values {
		r = append(r, *transSDKCookieConfigValueToUpdateRule(v))
	}
	return &r
}
func transSDKCookieConfigToCreateRules(cookieConfig alb.CookieConfig) albsdk.CreateRulesRulesRuleConditionsItemCookieConfig {
	return albsdk.CreateRulesRulesRuleConditionsItemCookieConfig{
		Values: transSDKCookieConfigValuesToCreateRules(cookieConfig.Values),
	}
}
func transSDKCookieConfigToUpdateRules(cookieConfig alb.CookieConfig) albsdk.UpdateRulesAttributeRulesRuleConditionsItemCookieConfig {
	return albsdk.UpdateRulesAttributeRulesRuleConditionsItemCookieConfig{
		Values: transSDKCookieConfigValuesToUpdateRules(cookieConfig.Values),
	}
}
func transSDKCookieConfigToCreateRule(cookieConfig alb.CookieConfig) albsdk.CreateRuleRuleConditionsCookieConfig {
	return albsdk.CreateRuleRuleConditionsCookieConfig{
		Values: transSDKCookieConfigValuesToCreateRule(cookieConfig.Values),
	}
}
func transSDKCookieConfigToUpdateRule(cookieConfig alb.CookieConfig) albsdk.UpdateRuleAttributeRuleConditionsCookieConfig {
	return albsdk.UpdateRuleAttributeRuleConditionsCookieConfig{
		Values: transSDKCookieConfigValuesToUpdateRule(cookieConfig.Values),
	}
}

func transSDKHostConfigToCreateRules(hostConf alb.HostConfig) albsdk.CreateRulesRulesRuleConditionsItemHostConfig {
	return albsdk.CreateRulesRulesRuleConditionsItemHostConfig{
		Values: &hostConf.Values,
	}
}
func transSDKHostConfigToUpdateRules(hostConf alb.HostConfig) albsdk.UpdateRulesAttributeRulesRuleConditionsItemHostConfig {
	return albsdk.UpdateRulesAttributeRulesRuleConditionsItemHostConfig{
		Values: &hostConf.Values,
	}
}
func transSDKHostConfigToCreateRule(hostConf alb.HostConfig) albsdk.CreateRuleRuleConditionsHostConfig {
	return albsdk.CreateRuleRuleConditionsHostConfig{
		Values: &hostConf.Values,
	}
}
func transSDKHostConfigToUpdateRule(hostConf alb.HostConfig) albsdk.UpdateRuleAttributeRuleConditionsHostConfig {
	return albsdk.UpdateRuleAttributeRuleConditionsHostConfig{
		Values: &hostConf.Values,
	}
}

func transSDKHeaderConfigToCreateRules(headerConfig alb.HeaderConfig) albsdk.CreateRulesRulesRuleConditionsItemHeaderConfig {
	return albsdk.CreateRulesRulesRuleConditionsItemHeaderConfig{
		Values: &headerConfig.Values,
		Key:    headerConfig.Key,
	}
}
func transSDKHeaderConfigToUpdateRules(headerConfig alb.HeaderConfig) albsdk.UpdateRulesAttributeRulesRuleConditionsItemHeaderConfig {
	return albsdk.UpdateRulesAttributeRulesRuleConditionsItemHeaderConfig{
		Values: &headerConfig.Values,
		Key:    headerConfig.Key,
	}
}
func transSDKHeaderConfigToCreateRule(headerConfig alb.HeaderConfig) albsdk.CreateRuleRuleConditionsHeaderConfig {
	return albsdk.CreateRuleRuleConditionsHeaderConfig{
		Values: &headerConfig.Values,
		Key:    headerConfig.Key,
	}
}
func transSDKHeaderConfigToUpdateRule(headerConfig alb.HeaderConfig) albsdk.UpdateRuleAttributeRuleConditionsHeaderConfig {
	return albsdk.UpdateRuleAttributeRuleConditionsHeaderConfig{
		Values: &headerConfig.Values,
		Key:    headerConfig.Key,
	}
}

func transSDKMethodConfigToCreateRules(methodConfig alb.MethodConfig) albsdk.CreateRulesRulesRuleConditionsItemMethodConfig {
	return albsdk.CreateRulesRulesRuleConditionsItemMethodConfig{
		Values: &methodConfig.Values,
	}
}
func transSDKMethodConfigToCreateRule(methodConfig alb.MethodConfig) albsdk.CreateRuleRuleConditionsMethodConfig {
	return albsdk.CreateRuleRuleConditionsMethodConfig{
		Values: &methodConfig.Values,
	}
}
func transSDKMethodConfigToUpdateRules(methodConfig alb.MethodConfig) albsdk.UpdateRulesAttributeRulesRuleConditionsItemMethodConfig {
	return albsdk.UpdateRulesAttributeRulesRuleConditionsItemMethodConfig{
		Values: &methodConfig.Values,
	}
}
func transSDKMethodConfigToUpdateRule(methodConfig alb.MethodConfig) albsdk.UpdateRuleAttributeRuleConditionsMethodConfig {
	return albsdk.UpdateRuleAttributeRuleConditionsMethodConfig{
		Values: &methodConfig.Values,
	}
}

func transSDKPathConfigToCreateRules(pathConfig alb.PathConfig) albsdk.CreateRulesRulesRuleConditionsItemPathConfig {
	return albsdk.CreateRulesRulesRuleConditionsItemPathConfig{
		Values: &pathConfig.Values,
	}
}
func transSDKPathConfigToUpdateRules(pathConfig alb.PathConfig) albsdk.UpdateRulesAttributeRulesRuleConditionsItemPathConfig {
	return albsdk.UpdateRulesAttributeRulesRuleConditionsItemPathConfig{
		Values: &pathConfig.Values,
	}
}
func transSDKPathConfigToCreateRule(pathConfig alb.PathConfig) albsdk.CreateRuleRuleConditionsPathConfig {
	return albsdk.CreateRuleRuleConditionsPathConfig{
		Values: &pathConfig.Values,
	}
}
func transSDKPathConfigToUpdateRule(pathConfig alb.PathConfig) albsdk.UpdateRuleAttributeRuleConditionsPathConfig {
	return albsdk.UpdateRuleAttributeRuleConditionsPathConfig{
		Values: &pathConfig.Values,
	}
}
func transSDKConditionToCreateRules(condition alb.Condition) *albsdk.CreateRulesRulesRuleConditionsItem {
	resCondition := &albsdk.CreateRulesRulesRuleConditionsItem{}
	resCondition.Type = condition.Type

	switch condition.Type {
	case util.RuleConditionFieldHost:
		resCondition.HostConfig = transSDKHostConfigToCreateRules(condition.HostConfig)
	case util.RuleConditionFieldHeader:
		resCondition.HeaderConfig = transSDKHeaderConfigToCreateRules(condition.HeaderConfig)
	case util.RuleConditionFieldMethod:
		resCondition.MethodConfig = transSDKMethodConfigToCreateRules(condition.MethodConfig)
	case util.RuleConditionFieldPath:
		resCondition.PathConfig = transSDKPathConfigToCreateRules(condition.PathConfig)
	case util.RuleConditionFieldQueryString:
		resCondition.QueryStringConfig = transSDKQueryStringConfigToCreateRules(condition.QueryStringConfig)
	case util.RuleConditionFieldCookie:
		resCondition.CookieConfig = transSDKCookieConfigToCreateRules(condition.CookieConfig)
	}

	return resCondition
}
func transSDKConditionToUpdateRules(condition alb.Condition) *albsdk.UpdateRulesAttributeRulesRuleConditionsItem {
	resCondition := &albsdk.UpdateRulesAttributeRulesRuleConditionsItem{}
	resCondition.Type = condition.Type

	switch condition.Type {
	case util.RuleConditionFieldHost:
		resCondition.HostConfig = transSDKHostConfigToUpdateRules(condition.HostConfig)
	case util.RuleConditionFieldHeader:
		resCondition.HeaderConfig = transSDKHeaderConfigToUpdateRules(condition.HeaderConfig)
	case util.RuleConditionFieldMethod:
		resCondition.MethodConfig = transSDKMethodConfigToUpdateRules(condition.MethodConfig)
	case util.RuleConditionFieldPath:
		resCondition.PathConfig = transSDKPathConfigToUpdateRules(condition.PathConfig)
	case util.RuleConditionFieldQueryString:
		resCondition.QueryStringConfig = transSDKQueryStringConfigToUpdateRules(condition.QueryStringConfig)
	case util.RuleConditionFieldCookie:
		resCondition.CookieConfig = transSDKCookieConfigToUpdateRules(condition.CookieConfig)
	}

	return resCondition
}

func transSDKConditionToCreateRule(condition alb.Condition) *albsdk.CreateRuleRuleConditions {
	resCondition := &albsdk.CreateRuleRuleConditions{}
	resCondition.Type = condition.Type

	switch condition.Type {
	case util.RuleConditionFieldHost:
		resCondition.HostConfig = transSDKHostConfigToCreateRule(condition.HostConfig)
	case util.RuleConditionFieldHeader:
		resCondition.HeaderConfig = transSDKHeaderConfigToCreateRule(condition.HeaderConfig)
	case util.RuleConditionFieldMethod:
		resCondition.MethodConfig = transSDKMethodConfigToCreateRule(condition.MethodConfig)
	case util.RuleConditionFieldPath:
		resCondition.PathConfig = transSDKPathConfigToCreateRule(condition.PathConfig)
	case util.RuleConditionFieldQueryString:
		resCondition.QueryStringConfig = transSDKQueryStringConfigToCreateRule(condition.QueryStringConfig)
	case util.RuleConditionFieldCookie:
		resCondition.CookieConfig = transSDKCookieConfigToCreateRule(condition.CookieConfig)
	}

	return resCondition
}
func transSDKConditionToUpdateRule(condition alb.Condition) *albsdk.UpdateRuleAttributeRuleConditions {
	resCondition := &albsdk.UpdateRuleAttributeRuleConditions{}
	resCondition.Type = condition.Type

	switch condition.Type {
	case util.RuleConditionFieldHost:
		resCondition.HostConfig = transSDKHostConfigToUpdateRule(condition.HostConfig)
	case util.RuleConditionFieldHeader:
		resCondition.HeaderConfig = transSDKHeaderConfigToUpdateRule(condition.HeaderConfig)
	case util.RuleConditionFieldMethod:
		resCondition.MethodConfig = transSDKMethodConfigToUpdateRule(condition.MethodConfig)
	case util.RuleConditionFieldPath:
		resCondition.PathConfig = transSDKPathConfigToUpdateRule(condition.PathConfig)
	case util.RuleConditionFieldQueryString:
		resCondition.QueryStringConfig = transSDKQueryStringConfigToUpdateRule(condition.QueryStringConfig)
	case util.RuleConditionFieldCookie:
		resCondition.CookieConfig = transSDKCookieConfigToUpdateRule(condition.CookieConfig)
	}

	return resCondition
}

func transSDKConditionsToCreateRules(conditions []alb.Condition) *[]albsdk.CreateRulesRulesRuleConditionsItem {
	var createRuleConditions []albsdk.CreateRulesRulesRuleConditionsItem
	if len(conditions) != 0 {
		createRuleConditions = make([]albsdk.CreateRulesRulesRuleConditionsItem, 0)
		for _, condition := range conditions {
			createRuleCondition := transSDKConditionToCreateRules(condition)
			createRuleConditions = append(createRuleConditions, *createRuleCondition)
		}
	}
	return &createRuleConditions
}
func transSDKConditionsToUpdateRules(conditions []alb.Condition) *[]albsdk.UpdateRulesAttributeRulesRuleConditionsItem {
	var updateRulesConditions []albsdk.UpdateRulesAttributeRulesRuleConditionsItem
	if len(conditions) != 0 {
		updateRulesConditions = make([]albsdk.UpdateRulesAttributeRulesRuleConditionsItem, 0)
		for _, condition := range conditions {
			updateRuleCondition := transSDKConditionToUpdateRules(condition)
			updateRulesConditions = append(updateRulesConditions, *updateRuleCondition)
		}
	}
	return &updateRulesConditions
}

func transSDKConditionsToCreateRule(conditions []alb.Condition) *[]albsdk.CreateRuleRuleConditions {
	var createRuleConditions []albsdk.CreateRuleRuleConditions
	if len(conditions) != 0 {
		createRuleConditions = make([]albsdk.CreateRuleRuleConditions, 0)
		for _, condition := range conditions {
			createRuleCondition := transSDKConditionToCreateRule(condition)
			createRuleConditions = append(createRuleConditions, *createRuleCondition)
		}
	}
	return &createRuleConditions
}
func transSDKConditionsToUpdateRule(conditions []alb.Condition) *[]albsdk.UpdateRuleAttributeRuleConditions {
	var updateRuleConditions []albsdk.UpdateRuleAttributeRuleConditions
	if len(conditions) != 0 {
		updateRuleConditions = make([]albsdk.UpdateRuleAttributeRuleConditions, 0)
		for _, condition := range conditions {
			updateRuleCondition := transSDKConditionToUpdateRule(condition)
			updateRuleConditions = append(updateRuleConditions, *updateRuleCondition)
		}
	}
	return &updateRuleConditions
}

func buildSDKCreateListenerRuleRequest(lrSpec alb.ListenerRuleSpec) (*albsdk.CreateRuleRequest, error) {
	lsID, err := lrSpec.ListenerID.Resolve(context.Background())
	if err != nil {
		return nil, err
	}

	ruleReq := albsdk.CreateCreateRuleRequest()
	ruleReq.RuleName = lrSpec.RuleName
	ruleReq.ListenerId = lsID
	ruleReq.RuleConditions = transSDKConditionsToCreateRule(lrSpec.RuleConditions)
	actions, err := transModelActionsToSDKCreateRule(lrSpec.RuleActions)
	if err != nil {
		return nil, err
	}
	ruleReq.RuleActions = actions
	ruleReq.Priority = requests.NewInteger(lrSpec.Priority)

	return ruleReq, nil
}

func isVipStatusNotSupportError(err error) bool {
	if strings.Contains(err.Error(), "IncorrectStatus.Listener") ||
		strings.Contains(err.Error(), "VipStatusNotSupport") ||
		strings.Contains(err.Error(), "IncorrectStatus.Rule") {
		return true
	}
	return false
}

func buildResListenerRuleStatus(ruleID string) alb.ListenerRuleStatus {
	return alb.ListenerRuleStatus{
		RuleID: ruleID,
	}
}

type ListenerRuleUpdateAnalyzer struct {
	ruleNameNeedUpdate       bool
	ruleActionsNeedUpdate    bool
	ruleConditionsNeedUpdate bool
	priorityNeedUpdate       bool
	needUpdate               bool
}

func (r *ListenerRuleUpdateAnalyzer) isNeedUpdate() bool {
	if !r.ruleNameNeedUpdate &&
		!r.ruleActionsNeedUpdate &&
		!r.ruleConditionsNeedUpdate &&
		!r.priorityNeedUpdate {
		return false
	}
	return true
}

func (r *ListenerRuleUpdateAnalyzer) analysis(_ context.Context, resLR *alb.ListenerRule, sdkLR *albsdk.Rule) error {
	if resLR.Spec.RuleName != sdkLR.RuleName {
		r.ruleNameNeedUpdate = true
	}

	if resLR.Spec.Priority != sdkLR.Priority {
		r.priorityNeedUpdate = true
	}

	resActions, err := transModelActionsToSDK(resLR.Spec.RuleActions)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(*resActions, sdkLR.RuleActions) {
		r.ruleActionsNeedUpdate = true
	}

	if !reflect.DeepEqual(resLR.Spec.RuleConditions, sdkLR.RuleConditions) {
		r.ruleConditionsNeedUpdate = true
	}

	r.needUpdate = r.isNeedUpdate()

	return nil
}

var createRulesFunc = func(ctx context.Context, ruleMgr *ALBProvider, lsID string, rules []albsdk.CreateRulesRules) ([]albsdk.RuleId, error) {
	if len(rules) == 0 {
		return nil, nil
	}

	traceID := ctx.Value(util.TraceID)

	createRulesReq := albsdk.CreateCreateRulesRequest()
	createRulesReq.ListenerId = lsID
	createRulesReq.Rules = &rules

	var createRuleResp *albsdk.CreateRulesResponse
	if err := util.RetryImmediateOnError(ruleMgr.waitLSExistencePollInterval, ruleMgr.waitLSExistenceTimeout, isVipStatusNotSupportError, func() error {
		startTime := time.Now()
		ruleMgr.logger.V(util.MgrLogLevel).Info("creating rules",
			"listenerID", lsID,
			"rules", rules,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.CreateALBRules)
		var err error
		createRuleResp, err = ruleMgr.auth.ALB.CreateRules(createRulesReq)
		if err != nil {
			ruleMgr.logger.V(util.MgrLogLevel).Info("creating rules",
				"listenerID", createRulesReq.ListenerId,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.CreateALBRules)
			return err
		}
		ruleMgr.logger.V(util.MgrLogLevel).Info("created rules",
			"listenerID", createRulesReq.ListenerId,
			"traceID", traceID,
			"rules", createRuleResp.RuleIds,
			"requestID", createRuleResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.CreateALBRules)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to create listener rules")
	}

	return createRuleResp.RuleIds, nil
}

type BatchCreateRulesFunc func(context.Context, *ALBProvider, string, []albsdk.CreateRulesRules) ([]albsdk.RuleId, error)

func BatchCreateRules(ctx context.Context, ruleMgr *ALBProvider, lsID string, rules []albsdk.CreateRulesRules, cnt int, batch BatchCreateRulesFunc) ([]albsdk.RuleId, error) {
	if cnt <= 0 || cnt >= util.BatchCreateDeleteUpdateRulesMaxNum {
		cnt = util.BatchCreateRulesDefaultNum
	}

	var resRules []albsdk.RuleId

	for len(rules) > cnt {
		resp, err := batch(ctx, ruleMgr, lsID, rules[0:cnt])
		if err != nil {
			return nil, err
		}
		resRules = append(resRules, resp...)
		rules = rules[cnt:]
	}
	if len(rules) <= 0 {
		return resRules, nil
	}

	resp, err := batch(ctx, ruleMgr, lsID, rules)
	if err != nil {
		return nil, err
	}
	resRules = append(resRules, resp...)

	return resRules, nil
}

func transModelRulesToSDKCreateRules(resLRs []*alb.ListenerRule) ([]albsdk.CreateRulesRules, error) {
	creatRules := make([]albsdk.CreateRulesRules, 0)
	for _, resLR := range resLRs {
		actions, err := transModelActionsToSDKCreateRules(resLR.Spec.RuleActions)
		if err != nil {
			return nil, err
		}
		creatRules = append(creatRules, albsdk.CreateRulesRules{
			RuleConditions: transSDKConditionsToCreateRules(resLR.Spec.RuleConditions),
			RuleName:       resLR.Spec.RuleName,
			Priority:       strconv.Itoa(resLR.Spec.Priority),
			RuleActions:    actions,
		})
	}
	return creatRules, nil
}

func (m *ALBProvider) CreateALBListenerRules(ctx context.Context, resLRs []*alb.ListenerRule) (map[int]alb.ListenerRuleStatus, error) {
	if len(resLRs) == 0 {
		return nil, nil
	}

	lsID, err := (*resLRs[0]).Spec.ListenerID.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	rules, err := transModelRulesToSDKCreateRules(resLRs)
	if err != nil {
		return nil, err
	}

	createdRules, err := BatchCreateRules(ctx, m, lsID, rules, util.BatchCreateRulesDefaultNum, createRulesFunc)
	if err != nil {
		return nil, err
	}

	priorityToStatus := make(map[int]alb.ListenerRuleStatus)
	for _, createdRule := range createdRules {
		priorityToStatus[createdRule.Priority] = alb.ListenerRuleStatus{
			RuleID: createdRule.RuleId,
		}
	}
	return priorityToStatus, nil
}

var updateRulesFunc = func(ctx context.Context, ruleMgr *ALBProvider, rules []albsdk.UpdateRulesAttributeRules) error {
	if len(rules) == 0 {
		return nil
	}

	traceID := ctx.Value(util.TraceID)

	updateRulesReq := albsdk.CreateUpdateRulesAttributeRequest()
	updateRulesReq.Rules = &rules

	if err := util.RetryImmediateOnError(ruleMgr.waitLSExistencePollInterval, ruleMgr.waitLSExistenceTimeout, isVipStatusNotSupportError, func() error {
		startTime := time.Now()
		ruleMgr.logger.V(util.MgrLogLevel).Info("updating rules attribute",
			"rules", rules,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.UpdateALBRulesAttribute)
		updateRulesResp, err := ruleMgr.auth.ALB.UpdateRulesAttribute(updateRulesReq)
		if err != nil {
			ruleMgr.logger.V(util.MgrLogLevel).Info("updated rules attribute",
				"traceID", traceID,
				"requestID", updateRulesResp.RequestId,
				"error", err.Error(),
				util.Action, util.UpdateALBRulesAttribute)
			return err
		}
		ruleMgr.logger.V(util.MgrLogLevel).Info("updated rules attribute",
			"traceID", traceID,
			"requestID", updateRulesResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.UpdateALBRulesAttribute)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to update listener rules")
	}

	return nil
}

type BatchUpdateRulesFunc func(context.Context, *ALBProvider, []albsdk.UpdateRulesAttributeRules) error

func BatchUpdateRules(ctx context.Context, ruleMgr *ALBProvider, rules []albsdk.UpdateRulesAttributeRules, cnt int, batch BatchUpdateRulesFunc) error {
	if cnt <= 0 || cnt >= util.BatchCreateDeleteUpdateRulesMaxNum {
		cnt = util.BatchUpdateRulesDefaultNum
	}

	for len(rules) > cnt {
		if err := batch(ctx, ruleMgr, rules[0:cnt]); err != nil {
			return err
		}
		rules = rules[cnt:]
	}
	if len(rules) <= 0 {
		return nil
	}

	return batch(ctx, ruleMgr, rules)
}

func (m *ALBProvider) transRulePairsToSDKUpdateRules(ctx context.Context, rulePairs []alb.ResAndSDKListenerRulePair) ([]albsdk.UpdateRulesAttributeRules, error) {
	if len(rulePairs) == 0 {
		return nil, nil
	}

	traceID := ctx.Value(util.TraceID)

	rules := make([]albsdk.UpdateRulesAttributeRules, 0)

	for _, rulePair := range rulePairs {
		updateAnalyzer := new(ListenerRuleUpdateAnalyzer)
		if err := updateAnalyzer.analysis(ctx, rulePair.ResLR, rulePair.SdkLR); err != nil {
			return nil, err
		}

		if !updateAnalyzer.needUpdate {
			continue
		}

		var rule albsdk.UpdateRulesAttributeRules
		rule.RuleId = rulePair.SdkLR.RuleId

		if updateAnalyzer.ruleNameNeedUpdate {
			m.logger.V(util.MgrLogLevel).Info("RuleName update",
				"res", rulePair.ResLR.Spec.RuleName,
				"sdk", rulePair.SdkLR.RuleName,
				"ruleID", rulePair.SdkLR.RuleId,
				"traceID", traceID)
			rule.RuleName = rulePair.ResLR.Spec.RuleName
		}
		if updateAnalyzer.ruleActionsNeedUpdate {
			m.logger.V(util.MgrLogLevel).Info("Actions update",
				"res", rulePair.ResLR.Spec.RuleActions,
				"sdk", rulePair.SdkLR.RuleActions,
				"ruleID", rulePair.SdkLR.RuleId,
				"traceID", traceID)
			actions, err := transModelActionsToSDKUpdateRules(rulePair.ResLR.Spec.RuleActions)
			if err != nil {
				return nil, err
			}
			rule.RuleActions = actions
		}
		if updateAnalyzer.ruleConditionsNeedUpdate {
			m.logger.V(util.MgrLogLevel).Info("Conditions update",
				"res", rulePair.ResLR.Spec.RuleConditions,
				"sdk", rulePair.SdkLR.RuleConditions,
				"ruleID", rulePair.SdkLR.RuleId,
				"traceID", traceID)
			rule.RuleConditions = transSDKConditionsToUpdateRules(rulePair.ResLR.Spec.RuleConditions)
		}
		if updateAnalyzer.priorityNeedUpdate {
			m.logger.V(util.MgrLogLevel).Info("Priority update",
				"res", rulePair.ResLR.Spec.Priority,
				"sdk", rulePair.SdkLR.Priority,
				"ruleID", rulePair.SdkLR.RuleId,
				"traceID", traceID)
			rule.Priority = strconv.Itoa(rulePair.ResLR.Spec.Priority)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (m *ALBProvider) UpdateALBListenerRules(ctx context.Context, rulePairs []alb.ResAndSDKListenerRulePair) error {
	if len(rulePairs) == 0 {
		return nil
	}

	updateRules, err := m.transRulePairsToSDKUpdateRules(ctx, rulePairs)
	if err != nil {
		return err
	}

	if len(updateRules) == 0 {
		return nil
	}

	if err := BatchUpdateRules(ctx, m, updateRules, util.BatchUpdateRulesDefaultNum, updateRulesFunc); err != nil {
		return err
	}

	return nil
}

var deleteRulesFunc = func(ctx context.Context, ruleMgr *ALBProvider, ruleIDs []string) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	traceID := ctx.Value(util.TraceID)

	deleteRulesReq := albsdk.CreateDeleteRulesRequest()
	deleteRulesReq.RuleIds = &ruleIDs

	if err := util.RetryImmediateOnError(ruleMgr.waitLSExistencePollInterval, ruleMgr.waitLSExistenceTimeout, isVipStatusNotSupportError, func() error {
		startTime := time.Now()
		ruleMgr.logger.V(util.MgrLogLevel).Info("deleting rules",
			"ruleIDs", ruleIDs,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.DeleteALBRules)
		deleteRulesResp, err := ruleMgr.auth.ALB.DeleteRules(deleteRulesReq)
		if err != nil {
			ruleMgr.logger.V(util.MgrLogLevel).Info("deleting rules",
				"ruleIDs", ruleIDs,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.DeleteALBRules)
			return err
		}
		ruleMgr.logger.V(util.MgrLogLevel).Info("deleted rules",
			"traceID", traceID,
			"requestID", deleteRulesResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.DeleteALBRules)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to delete listener rules")
	}

	return nil
}

type BatchDeleteRulesFunc func(context.Context, *ALBProvider, []string) error

func BatchDeleteRules(ctx context.Context, ruleMgr *ALBProvider, ruleIDs []string, cnt int, batch BatchDeleteRulesFunc) error {
	if cnt <= 0 || cnt >= util.BatchCreateDeleteUpdateRulesMaxNum {
		cnt = util.BatchDeleteRulesDefaultNum
	}

	for len(ruleIDs) > cnt {
		if err := batch(ctx, ruleMgr, ruleIDs[0:cnt]); err != nil {
			return err
		}
		ruleIDs = ruleIDs[cnt:]
	}
	if len(ruleIDs) <= 0 {
		return nil
	}

	return batch(ctx, ruleMgr, ruleIDs)
}

func (m *ALBProvider) DeleteALBListenerRules(ctx context.Context, sdkLRIds []string) error {
	if len(sdkLRIds) == 0 {
		return nil
	}

	if err := BatchDeleteRules(ctx, m, sdkLRIds, util.BatchDeleteRulesDefaultNum, deleteRulesFunc); err != nil {
		return err
	}

	return nil
}
