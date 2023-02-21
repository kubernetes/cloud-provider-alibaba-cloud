package applier

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/sets"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

func NewListenerApplier(albProvider prvd.Provider, stack core.Manager, logger logr.Logger, commonReuse bool, errRes core.ErrResult, listenerCommonReuse bool) *listenerApplier {
	return &listenerApplier{
		albProvider:         albProvider,
		stack:               stack,
		logger:              logger,
		commonReuse:         commonReuse,
		errRes:              errRes,
		listenerCommonReuse: listenerCommonReuse,
	}
}

type listenerApplier struct {
	albProvider         prvd.Provider
	stack               core.Manager
	logger              logr.Logger
	commonReuse         bool
	errRes              core.ErrResult
	listenerCommonReuse bool
}

func (s *listenerApplier) Apply(ctx context.Context) error {
	var resLSs []*albmodel.Listener
	_ = s.stack.ListResources(&resLSs)

	resLSsByLbID, err := mapResListenerByAlbLoadBalancerID(ctx, resLSs)
	if err != nil {
		return err
	}

	if len(resLSsByLbID) == 0 {
		resLSsByLbID = make(map[string][]*albmodel.Listener)
		var resLBs []*albmodel.AlbLoadBalancer
		_ = s.stack.ListResources(&resLBs)
		if len(resLBs) == 0 {
			return nil
		}
		lbID, err := resLBs[0].LoadBalancerID().Resolve(ctx)
		if err != nil {
			return err
		}
		resLSsByLbID[lbID] = make([]*albmodel.Listener, 0)
	}

	for lbID, resLSs := range resLSsByLbID {
		if err := s.applyListenersOnLB(ctx, lbID, resLSs); err != nil {
			return err
		}
	}

	return nil
}

func (s *listenerApplier) PostApply(ctx context.Context) error {
	return nil
}

func (s *listenerApplier) applyListenersOnLB(ctx context.Context, lbID string, resLSs []*albmodel.Listener) error {
	if len(lbID) == 0 {
		return fmt.Errorf("empty loadBalancer id when apply listeners error")
	}

	traceID := ctx.Value(util.TraceID)

	sdkLSs, err := s.findSDKListenersOnLB(ctx, lbID)
	if err != nil {
		return err
	}
	matchedResAndSDKLSs, unmatchedResLSs, unmatchedSDKLSs := matchResAndSDKListeners(resLSs, sdkLSs)

	if len(matchedResAndSDKLSs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply listeners",
			"matchedResAndSDKLSs", matchedResAndSDKLSs,
			"traceID", traceID)
	}
	if len(unmatchedResLSs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply listeners",
			"unmatchedResLSs", unmatchedResLSs,
			"traceID", traceID)
	}
	if len(unmatchedSDKLSs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply listeners",
			"unmatchedSDKLSs", unmatchedSDKLSs,
			"traceID", traceID)
	}

	var (
		errCreate error
		wgCreate  sync.WaitGroup
	)
	for _, resLS := range unmatchedResLSs {
		wgCreate.Add(1)
		go func(resLS *albmodel.Listener) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer wgCreate.Done()
			lsStatus, err := s.albProvider.CreateALBListener(ctx, resLS)
			if errCreate == nil && err != nil {
				errCreate = err
				s.errRes.AddErrMsgsWithListenerPort(resLS.Spec.ListenerPort, errCreate)
				s.logger.V(util.SynLogLevel).Error(errCreate, "CreateALBListener AddErrMsg")
			}
			resLS.SetStatus(lsStatus)
		}(resLS)
	}
	wgCreate.Wait()
	if errCreate != nil {
		if strings.Contains(errCreate.Error(), "ResourceAlreadyExist.Listener") {
			return fmt.Errorf("listener already in use, please set forceOverride=true and reconcile")
		}
		//return errCreate
	}

	var (
		errUpdate error
		wgUpdate  sync.WaitGroup
	)
	for _, resAndSDKLS := range matchedResAndSDKLSs {
		wgUpdate.Add(1)
		go func(resLs *albmodel.Listener, sdkLs *albsdk.Listener) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer wgUpdate.Done()
			lsStatus, err := s.albProvider.UpdateALBListener(ctx, resLs, sdkLs)
			if errUpdate == nil && err != nil {
				errUpdate = err
				s.errRes.AddErrMsgsWithListenerPort(resAndSDKLS.resLS.Spec.ListenerPort, errUpdate)
				s.logger.V(util.SynLogLevel).Error(errUpdate, "UpdateALBListener AddErrMsg")
			}
			resLs.SetStatus(lsStatus)
		}(resAndSDKLS.resLS, resAndSDKLS.sdkLS)
	}
	wgUpdate.Wait()
	if errUpdate != nil {
		//return errUpdate
	}
	var (
		errDelete error
		wgDelete  sync.WaitGroup
	)
	for _, sdkLS := range unmatchedSDKLSs {
		wgDelete.Add(1)
		go func(sdkLS albsdk.Listener) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)
			defer wgDelete.Done()
			if err := s.albProvider.DeleteALBListener(ctx, sdkLS.ListenerId); errDelete == nil && err != nil {
				errDelete = err
			}
		}(sdkLS)
	}
	wgDelete.Wait()
	if errDelete != nil {
		return errDelete
	}
	return nil
}

func (s *listenerApplier) findSDKListenersOnLB(ctx context.Context, lbID string) ([]albsdk.Listener, error) {
	listeners, err := s.albProvider.ListALBListeners(ctx, lbID)
	if err != nil {
		return nil, err
	}
	if !s.commonReuse {
		return listeners, nil
	}
	if !s.listenerCommonReuse {
		return listeners, nil
	}
	filteredListeners := make([]albsdk.Listener, 0)
	for _, ls := range listeners {
		if strings.HasPrefix(ls.ListenerDescription, util.ListenerDescriptionPrefix) {
			filteredListeners = append(filteredListeners, ls)
		}
	}
	return filteredListeners, nil
}

type resAndSDKListenerPair struct {
	resLS *albmodel.Listener
	sdkLS *albsdk.Listener
}

func matchResAndSDKListeners(resLSs []*albmodel.Listener, sdkLSs []albsdk.Listener) ([]resAndSDKListenerPair, []*albmodel.Listener, []albsdk.Listener) {
	var matchedResAndSDKLSs []resAndSDKListenerPair
	var unmatchedResLSs []*albmodel.Listener
	var unmatchedSDKLSs []albsdk.Listener

	resLSByPP := mapResListenerByPortProtocol(resLSs)
	sdkLSByPP := mapSDKListenerByPortProtocol(sdkLSs)
	resLSPP := sets.StringKeySet(resLSByPP)
	sdkLSPP := sets.StringKeySet(sdkLSByPP)
	for _, pp := range resLSPP.Intersection(sdkLSPP).List() {
		resLS := resLSByPP[pp]
		sdkLS := sdkLSByPP[pp]
		matchedResAndSDKLSs = append(matchedResAndSDKLSs, resAndSDKListenerPair{
			resLS: resLS,
			sdkLS: &sdkLS,
		})
	}
	for _, port := range resLSPP.Difference(sdkLSPP).List() {
		unmatchedResLSs = append(unmatchedResLSs, resLSByPP[port])
	}
	for _, port := range sdkLSPP.Difference(resLSPP).List() {
		unmatchedSDKLSs = append(unmatchedSDKLSs, sdkLSByPP[port])
	}

	return matchedResAndSDKLSs, unmatchedResLSs, unmatchedSDKLSs
}

func mapResListenerByPortProtocol(resLSs []*albmodel.Listener) map[string]*albmodel.Listener {
	resLSByPort := make(map[string]*albmodel.Listener, len(resLSs))
	for _, ls := range resLSs {
		pp := fmt.Sprintf("%s+%d", ls.Spec.ListenerProtocol, ls.Spec.ListenerPort)
		resLSByPort[pp] = ls
	}
	return resLSByPort
}

func mapSDKListenerByPortProtocol(sdkLSs []albsdk.Listener) map[string]albsdk.Listener {
	sdkLSByPort := make(map[string]albsdk.Listener, len(sdkLSs))
	for _, ls := range sdkLSs {
		pp := fmt.Sprintf("%s+%d", ls.ListenerProtocol, ls.ListenerPort)
		sdkLSByPort[pp] = ls
	}
	return sdkLSByPort
}

func mapResListenerByAlbLoadBalancerID(ctx context.Context, resLSs []*albmodel.Listener) (map[string][]*albmodel.Listener, error) {
	resLSsByLbID := make(map[string][]*albmodel.Listener, len(resLSs))
	for _, ls := range resLSs {
		lbID, err := ls.Spec.LoadBalancerID.Resolve(ctx)
		if err != nil {
			return nil, err
		}
		resLSsByLbID[lbID] = append(resLSsByLbID[lbID], ls)
	}
	return resLSsByLbID, nil
}
