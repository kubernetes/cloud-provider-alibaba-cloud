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

func NewListenerApplier(albProvider prvd.Provider, stack core.Manager, logger logr.Logger) *listenerApplier {
	return &listenerApplier{
		albProvider: albProvider,
		stack:       stack,
		logger:      logger,
	}
}

type listenerApplier struct {
	albProvider prvd.Provider
	stack       core.Manager
	logger      logr.Logger
}

func (s *listenerApplier) Apply(ctx context.Context) error {
	var resLSs []*albmodel.Listener
	s.stack.ListResources(&resLSs)

	resLSsByLbID, err := mapResListenerByAlbLoadBalancerID(ctx, resLSs)
	if err != nil {
		return err
	}

	if len(resLSsByLbID) == 0 {
		resLSsByLbID = make(map[string][]*albmodel.Listener)
		var resLBs []*albmodel.AlbLoadBalancer
		s.stack.ListResources(&resLBs)
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
		errDelete error
		wgDelete  sync.WaitGroup
	)
	for _, sdkLS := range unmatchedSDKLSs {
		wgDelete.Add(1)
		go func(sdkLS albsdk.Listener) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer wgDelete.Done()
			if err := s.albProvider.DeleteALB(ctx, sdkLS.ListenerId); errDelete == nil && err != nil {
				errDelete = err
			}
		}(sdkLS)
	}
	wgDelete.Wait()
	if errDelete != nil {
		return errDelete
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
			}
			resLS.SetStatus(lsStatus)
		}(resLS)
	}
	wgCreate.Wait()
	if errCreate != nil {
		return errCreate
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
			}
			resLs.SetStatus(lsStatus)
		}(resAndSDKLS.resLS, resAndSDKLS.sdkLS)
	}
	wgUpdate.Wait()
	if errUpdate != nil {
		return errUpdate
	}

	return nil
}

func (s *listenerApplier) findSDKListenersOnLB(ctx context.Context, lbID string) ([]albsdk.Listener, error) {
	listeners, err := s.albProvider.ListALBListeners(ctx, lbID)
	if err != nil {
		return nil, err
	}
	return listeners, nil
}

type resAndSDKListenerPair struct {
	resLS *albmodel.Listener
	sdkLS *albsdk.Listener
}

func matchResAndSDKListeners(resLSs []*albmodel.Listener, sdkLSs []albsdk.Listener) ([]resAndSDKListenerPair, []*albmodel.Listener, []albsdk.Listener) {
	var matchedResAndSDKLSs []resAndSDKListenerPair
	var unmatchedResLSs []*albmodel.Listener
	var unmatchedSDKLSs []albsdk.Listener

	resLSByPort := mapResListenerByPort(resLSs)
	sdkLSByPort := mapSDKListenerByPort(sdkLSs)
	resLSPorts := sets.Int64KeySet(resLSByPort)
	sdkLSPorts := sets.Int64KeySet(sdkLSByPort)
	for _, port := range resLSPorts.Intersection(sdkLSPorts).List() {
		resLS := resLSByPort[port]
		sdkLS := sdkLSByPort[port]
		matchedResAndSDKLSs = append(matchedResAndSDKLSs, resAndSDKListenerPair{
			resLS: resLS,
			sdkLS: &sdkLS,
		})
	}
	for _, port := range resLSPorts.Difference(sdkLSPorts).List() {
		unmatchedResLSs = append(unmatchedResLSs, resLSByPort[port])
	}
	for _, port := range sdkLSPorts.Difference(resLSPorts).List() {
		unmatchedSDKLSs = append(unmatchedSDKLSs, sdkLSByPort[port])
	}

	return matchedResAndSDKLSs, unmatchedResLSs, unmatchedSDKLSs
}

func mapResListenerByPort(resLSs []*albmodel.Listener) map[int64]*albmodel.Listener {
	resLSByPort := make(map[int64]*albmodel.Listener, len(resLSs))
	for _, ls := range resLSs {
		resLSByPort[int64(ls.Spec.ListenerPort)] = ls
	}
	return resLSByPort
}

func mapSDKListenerByPort(sdkLSs []albsdk.Listener) map[int64]albsdk.Listener {
	sdkLSByPort := make(map[int64]albsdk.Listener, len(sdkLSs))
	for _, ls := range sdkLSs {
		sdkLSByPort[int64(ls.ListenerPort)] = ls
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
