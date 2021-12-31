package applier

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/sets"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

func NewListenerRuleApplier(albProvider prvd.Provider, stack core.Manager, logger logr.Logger) *listenerRuleApplier {
	return &listenerRuleApplier{
		stack:       stack,
		albProvider: albProvider,
		logger:      logger,
	}
}

type listenerRuleApplier struct {
	albProvider prvd.Provider
	stack       core.Manager
	logger      logr.Logger
}

func (s *listenerRuleApplier) Apply(ctx context.Context) error {
	var resLRs []*albmodel.ListenerRule
	s.stack.ListResources(&resLRs)

	resLRsByLsID, err := mapResListenerRuleByListenerID(ctx, resLRs)
	if err != nil {
		return err
	}

	var resLSs []*albmodel.Listener
	s.stack.ListResources(&resLSs)

	resLSsByLsID, err := mapResListenerByListenerID(ctx, resLSs)
	if err != nil {
		return err
	}

	var (
		errApply error
		wgApply  sync.WaitGroup
		chApply  = make(chan struct{}, util.ListenerConcurrentNum)
	)

	for lsID := range resLSsByLsID {
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

func (s *listenerRuleApplier) applyListenerRulesOnListener(ctx context.Context, lsID string, resLRs []*albmodel.ListenerRule) error {
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
		s.logger.V(util.SynLogLevel).Info("apply rules",
			"matchedResAndSDKLRs", matchedResAndSDKLRs,
			"traceID", traceID)
	}
	if len(unmatchedResLRs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply rules",
			"unmatchedResLRs", unmatchedResLRs,
			"traceID", traceID)
	}
	if len(unmatchedSDKLRs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply rules",
			"unmatchedSDKLBs", unmatchedSDKLRs,
			"traceID", traceID)
	}

	for _, sdkLR := range unmatchedSDKLRs {
		if err := s.albProvider.DeleteALBListenerRule(ctx, sdkLR.RuleId); err != nil {
			return err
		}
	}
	for _, resLR := range unmatchedResLRs {
		lrStatus, err := s.albProvider.CreateALBListenerRule(ctx, resLR)
		if err != nil {
			return err
		}
		resLR.SetStatus(lrStatus)
	}
	for _, resAndSDKLR := range matchedResAndSDKLRs {
		lsStatus, err := s.albProvider.UpdateALBListenerRule(ctx, resAndSDKLR.ResLR, resAndSDKLR.SdkLR)
		if err != nil {
			return err
		}
		resAndSDKLR.ResLR.SetStatus(lsStatus)
	}
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

	unmatchedSDKLRIDs := make([]string, 0)
	for _, sdkLR := range unmatchedSDKLRs {
		unmatchedSDKLRIDs = append(unmatchedSDKLRIDs, sdkLR.RuleId)
	}
	if err := s.albProvider.DeleteALBListenerRules(ctx, unmatchedSDKLRIDs); err != nil {
		return err
	}

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

	resLRByPriority := mapResListenerRuleByPriority(resLRs)
	sdkLRByPriority := mapSDKListenerRuleByPriority(sdkLRs)
	resLRPriorities := sets.Int64KeySet(resLRByPriority)
	sdkLRPriorities := sets.Int64KeySet(sdkLRByPriority)
	for _, priority := range resLRPriorities.Intersection(sdkLRPriorities).List() {
		resLR := resLRByPriority[priority]
		sdkLR := sdkLRByPriority[priority]
		matchedResAndSDKLRs = append(matchedResAndSDKLRs, albmodel.ResAndSDKListenerRulePair{
			ResLR: resLR,
			SdkLR: &sdkLR,
		})
	}
	for _, priority := range resLRPriorities.Difference(sdkLRPriorities).List() {
		unmatchedResLRs = append(unmatchedResLRs, resLRByPriority[priority])
	}
	for _, priority := range sdkLRPriorities.Difference(resLRPriorities).List() {
		unmatchedSDKLRs = append(unmatchedSDKLRs, sdkLRByPriority[priority])
	}

	return matchedResAndSDKLRs, unmatchedResLRs, unmatchedSDKLRs
}

func mapResListenerRuleByPriority(resLRs []*albmodel.ListenerRule) map[int64]*albmodel.ListenerRule {
	resLRByPriority := make(map[int64]*albmodel.ListenerRule, 0)
	for _, resLR := range resLRs {
		resLRByPriority[int64(resLR.Spec.Priority)] = resLR
	}
	return resLRByPriority
}

func mapSDKListenerRuleByPriority(sdkLRs []albsdk.Rule) map[int64]albsdk.Rule {
	sdkLRByPriority := make(map[int64]albsdk.Rule, 0)
	for _, sdkLR := range sdkLRs {
		priority := int64(sdkLR.Priority)
		sdkLRByPriority[priority] = sdkLR
	}
	return sdkLRByPriority
}

func mapResListenerRuleByListenerID(ctx context.Context, resLRs []*albmodel.ListenerRule) (map[string][]*albmodel.ListenerRule, error) {
	resLRsByLsID := make(map[string][]*albmodel.ListenerRule, 0)
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
