package applier

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/sets"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

func NewListenerRuleApplier(albProvider prvd.Provider, stack core.Manager, logger logr.Logger, errRes core.ErrResult) *listenerRuleApplier {
	return &listenerRuleApplier{
		stack:       stack,
		albProvider: albProvider,
		logger:      logger,
		errRes:      errRes,
	}
}

type listenerRuleApplier struct {
	albProvider prvd.Provider
	stack       core.Manager
	logger      logr.Logger
	errRes      core.ErrResult
}

func (s *listenerRuleApplier) Apply(ctx context.Context) error {
	var resLRs []*albmodel.ListenerRule
	_ = s.stack.ListResources(&resLRs)

	resLRsByLsID, err := mapResListenerRuleByListenerID(ctx, resLRs)
	if err != nil {
		return err
	}

	var resLSs []*albmodel.Listener
	_ = s.stack.ListResources(&resLSs)

	resLSsByLsID, err := mapResListenerByListenerID(ctx, resLSs)
	if err != nil {
		return err
	}

	var (
		errApply error
		wgApply  sync.WaitGroup
		chApply  = make(chan struct{}, util.ListenerConcurrentNum)
	)

	for lsID, resLSs := range resLSsByLsID {
		err := s.errRes.CheckErrMsgsByListenerPort(resLSs.Spec.ListenerPort)
		if err != nil {
			s.logger.V(util.SynLogLevel).Error(err, "CheckErrMsgsByListenerPort succ")
			continue
		}
		chApply <- struct{}{}
		wgApply.Add(1)

		go func(listenerID string) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer func() {
				wgApply.Done()
				<-chApply
			}()

			rules := resLRsByLsID[listenerID]
			if errOnce := s.applyListenerRulesOnListenerBatch(ctx, listenerID, rules); errApply == nil && errOnce != nil {
				s.logger.Error(errOnce, "apply listener rules failed", "listener", listenerID)
				errApply = errOnce
				return
			}
		}(lsID)
	}
	wgApply.Wait()
	if errApply != nil {
		return errApply
	}

	return nil
}

func (s *listenerRuleApplier) PostApply(ctx context.Context) error {
	return nil
}

func (s *listenerRuleApplier) applyListenerRulesOnListenerBatch(ctx context.Context, lsID string, resLRs []*albmodel.ListenerRule) error {
	if len(lsID) == 0 {
		return fmt.Errorf("empty listener id when apply rules error")
	}

	traceID := ctx.Value(util.TraceID)

	sdkLRs, err := s.findSDKListenersRulesOnLS(ctx, lsID)
	if err != nil {
		return err
	}

	matchedResAndSDKLRs, unmatchedResLRs, unmatchedSDKLRs := matchResAndSDKListenerRules(resLRs, sdkLRs)

	if len(matchedResAndSDKLRs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply rules batch",
			"matchedResAndSDKLRs", matchedResAndSDKLRs,
			"traceID", traceID)
	}
	if len(unmatchedResLRs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply rules batch",
			"unmatchedResLRs", unmatchedResLRs,
			"traceID", traceID)
	}
	if len(unmatchedSDKLRs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply rules batch",
			"unmatchedSDKLBs", unmatchedSDKLRs,
			"traceID", traceID)
	}
	/** 1. 先新增、再修改、最后执行删除
	 *  2. 修改操作从后往前批量处理
	 */

	lrStatus, err := s.albProvider.CreateALBListenerRules(ctx, unmatchedResLRs)
	if err != nil {
		return err
	}
	for _, resLR := range unmatchedResLRs {
		status, ok := lrStatus[resLR.Spec.Priority]
		if !ok {
			return fmt.Errorf("failed create rule with priority: %d", resLR.Spec.Priority)
		}
		resLR.SetStatus(status)
	}

	resAndSDKListenerRulePairs := make([]albmodel.ResAndSDKListenerRulePair, 0)
	for _, matchedResAndSDKLR := range matchedResAndSDKLRs {
		resAndSDKListenerRulePairs = append(resAndSDKListenerRulePairs, albmodel.ResAndSDKListenerRulePair{
			ResLR: matchedResAndSDKLR.ResLR,
			SdkLR: matchedResAndSDKLR.SdkLR,
		})
	}
	err = s.albProvider.UpdateALBListenerRules(ctx, resAndSDKListenerRulePairs)
	if err != nil {
		return err
	}
	for _, matchedResAndSDKLR := range matchedResAndSDKLRs {
		matchedResAndSDKLR.ResLR.SetStatus(albmodel.ListenerRuleStatus{
			RuleID: matchedResAndSDKLR.SdkLR.RuleId,
		})
	}

	unmatchedSDKLRIDs := make([]string, 0)
	for _, sdkLR := range unmatchedSDKLRs {
		unmatchedSDKLRIDs = append(unmatchedSDKLRIDs, sdkLR.RuleId)
	}
	if err := s.albProvider.DeleteALBListenerRules(ctx, unmatchedSDKLRIDs); err != nil {
		return err
	}

	return nil
}

func (s *listenerRuleApplier) findSDKListenersRulesOnLS(ctx context.Context, lsID string) ([]albsdk.Rule, error) {
	rules, err := s.albProvider.ListALBListenerRules(ctx, lsID)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func matchResAndSDKListenerRules(resLRs []*albmodel.ListenerRule, sdkLRs []albsdk.Rule) ([]albmodel.ResAndSDKListenerRulePair, []*albmodel.ListenerRule, []albsdk.Rule) {
	var matchedResAndSDKLRs []albmodel.ResAndSDKListenerRulePair
	var unmatchedResLRs []*albmodel.ListenerRule
	var unmatchedSDKLRs []albsdk.Rule

	resLRByPriorityDirection := mapResListenerRuleByPriority(resLRs)
	sdkLRByPriorityDirection := mapSDKListenerRuleByPriority(sdkLRs)
	resLRPriorityDirections := sets.StringKeySet(resLRByPriorityDirection)
	sdkLRPriorityDirections := sets.StringKeySet(sdkLRByPriorityDirection)
	for _, priorityDirection := range resLRPriorityDirections.Intersection(sdkLRPriorityDirections).List() {
		resLR := resLRByPriorityDirection[priorityDirection]
		sdkLR := sdkLRByPriorityDirection[priorityDirection]
		matchedResAndSDKLRs = append(matchedResAndSDKLRs, albmodel.ResAndSDKListenerRulePair{
			ResLR: resLR,
			SdkLR: &sdkLR,
		})
	}
	for _, priorityDirection := range resLRPriorityDirections.Difference(sdkLRPriorityDirections).List() {
		unmatchedResLRs = append(unmatchedResLRs, resLRByPriorityDirection[priorityDirection])
	}
	for _, priorityDirection := range sdkLRPriorityDirections.Difference(resLRPriorityDirections).List() {
		unmatchedSDKLRs = append(unmatchedSDKLRs, sdkLRByPriorityDirection[priorityDirection])
	}

	return matchedResAndSDKLRs, unmatchedResLRs, unmatchedSDKLRs
}

func mapResListenerRuleByPriority(resLRs []*albmodel.ListenerRule) map[string]*albmodel.ListenerRule {
	resLRByPriorityDirection := make(map[string]*albmodel.ListenerRule)
	for _, resLR := range resLRs {
		resLRByPriorityDirection[strconv.Itoa(resLR.Spec.Priority)+resLR.Spec.RuleDirection] = resLR
	}
	return resLRByPriorityDirection
}

func mapSDKListenerRuleByPriority(sdkLRs []albsdk.Rule) map[string]albsdk.Rule {
	sdkLRByPriorityDirection := make(map[string]albsdk.Rule)
	for _, sdkLR := range sdkLRs {
		priority := strconv.Itoa(sdkLR.Priority) + sdkLR.Direction
		sdkLRByPriorityDirection[priority] = sdkLR
	}
	return sdkLRByPriorityDirection
}

func mapResListenerRuleByListenerID(ctx context.Context, resLRs []*albmodel.ListenerRule) (map[string][]*albmodel.ListenerRule, error) {
	resLRsByLsID := make(map[string][]*albmodel.ListenerRule)
	for _, lr := range resLRs {
		lsID, err := lr.Spec.ListenerID.Resolve(ctx)
		if err != nil {
			return nil, err
		}
		resLRsByLsID[lsID] = append(resLRsByLsID[lsID], lr)
	}
	return resLRsByLsID, nil
}

func mapResListenerByListenerID(ctx context.Context, resLSs []*albmodel.Listener) (map[string]*albmodel.Listener, error) {
	resLSByID := make(map[string]*albmodel.Listener, len(resLSs))
	for _, ls := range resLSs {
		lsID, err := ls.ListenerID().Resolve(ctx)
		if err != nil {
			return nil, err
		}
		resLSByID[lsID] = ls
	}
	return resLSByID, nil
}
