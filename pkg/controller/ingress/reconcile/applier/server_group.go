package applier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper/k8s"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/backend"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

func NewServerGroupApplier(kubeClient client.Client, backendManager backend.Manager, albProvider prvd.Provider, trackingProvider tracking.TrackingProvider, stack core.Manager, logger logr.Logger) *serverGroupApplier {
	return &serverGroupApplier{
		kubeClient:       kubeClient,
		trackingProvider: trackingProvider,
		stack:            stack,
		albProvider:      albProvider,
		backendManager:   backendManager,
		logger:           logger,
	}
}

type serverGroupApplier struct {
	albProvider      prvd.Provider
	trackingProvider tracking.TrackingProvider
	stack            core.Manager
	backendManager   backend.Manager
	kubeClient       client.Client
	unmatchedSDKSGPs []albmodel.ServerGroupWithTags
	logger           logr.Logger
}

func (s *serverGroupApplier) addServerToServerGroup(ctx context.Context, serverGroupID string, svcKey types.NamespacedName, port intstr.IntOrString) error {
	backends, _, err := s.backendManager.BuildServicePortSDKBackends(ctx, svcKey, port)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if len(backends) == 0 {
		return nil
	}
	if err := s.albProvider.RegisterALBServers(ctx, serverGroupID, backends); err != nil {
		return err
	}
	err = updateTargetHealthPodCondition(ctx, s.kubeClient, k8s.BuildReadinessGatePodConditionType(), backends)
	if err != nil {
		return err
	}
	return nil
}

func (s *serverGroupApplier) removeServerFromServerGroup(ctx context.Context, serverGroupID string, svcKey types.NamespacedName, port intstr.IntOrString) error {
	_, _, err := s.backendManager.BuildServicePortSDKBackends(ctx, svcKey, port)
	if err != nil {
		if apierrors.IsNotFound(err) {
			servers, err := s.albProvider.ListALBServers(ctx, serverGroupID)
			if err != nil {
				return err
			}
			if len(servers) == 0 {
				return nil
			}
			if err := s.albProvider.DeregisterALBServers(ctx, serverGroupID, servers); err != nil {
				return err
			}

			return nil
		}
		return err
	}
	return nil
}

func (s *serverGroupApplier) Apply(ctx context.Context) error {
	traceID := ctx.Value(util.TraceID)

	var resSGPs []*albmodel.ServerGroup
	s.stack.ListResources(&resSGPs)

	sdkSGPs, err := s.findSDKServerGroups(ctx)
	if err != nil {
		return err
	}

	matchedResAndSDKSGPs, unmatchedResSGPs, unmatchedSDKSGPs, err := matchResAndSDKServerGroupsSGP(resSGPs, sdkSGPs, s.trackingProvider.ResourceIDTagKey())
	if err != nil {
		return err
	}

	if len(matchedResAndSDKSGPs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply serverGroups",
			"matchedResAndSDKSGPs", matchedResAndSDKSGPs,
			"traceID", traceID)
	}
	if len(unmatchedResSGPs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply serverGroups",
			"unmatchedResSGPs", unmatchedResSGPs,
			"traceID", traceID)
	}
	if len(unmatchedSDKSGPs) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply serverGroups",
			"unmatchedSDKSGPs", unmatchedSDKSGPs,
			"traceID", traceID)
	}

	s.unmatchedSDKSGPs = unmatchedSDKSGPs

	var (
		errCreate error
		wgCreate  sync.WaitGroup
		chCreate  = make(chan struct{}, util.ServerGroupConcurrentNum)
	)
	for _, resSGP := range unmatchedResSGPs {
		chCreate <- struct{}{}
		wgCreate.Add(1)

		go func(res *albmodel.ServerGroup) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer func() {
				wgCreate.Done()
				<-chCreate
			}()

			sgpStatus, errOnce := s.albProvider.CreateALBServerGroup(ctx, res, s.trackingProvider)
			if errCreate == nil && errOnce != nil {
				errCreate = errOnce
				return
			}
			res.SetStatus(sgpStatus)

			if strings.Contains(res.Spec.ServerGroupNamedKey.IngressName, util.DefaultListenerFlag) {
				return
			}
			if errOnce = s.addServerToServerGroup(ctx, sgpStatus.ServerGroupID, types.NamespacedName{
				Namespace: res.Spec.ServerGroupNamedKey.Namespace,
				Name:      res.Spec.ServerGroupNamedKey.ServiceName},
				intstr.FromInt(res.Spec.ServerGroupNamedKey.ServicePort)); errCreate == nil && errOnce != nil {
				if errTmp := s.albProvider.DeleteALBServerGroup(ctx, sgpStatus.ServerGroupID); errTmp != nil {
					s.logger.V(util.SynLogLevel).Error(errTmp, "apply serverGroups roll back server group failed",
						"serverGroupID", sgpStatus.ServerGroupID, "traceID", traceID)
				}
				errCreate = errOnce
				return
			}
		}(resSGP)
	}
	wgCreate.Wait()
	if errCreate != nil {
		return errCreate
	}

	var (
		errUpdate error
		wgUpdate  sync.WaitGroup
		chUpdate  = make(chan struct{}, util.ServerGroupConcurrentNum)
	)
	for _, resAndSDKSGP := range matchedResAndSDKSGPs {
		chUpdate <- struct{}{}
		wgUpdate.Add(1)

		go func(resSGP *albmodel.ServerGroup, sdkSGP albmodel.ServerGroupWithTags) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer func() {
				wgUpdate.Done()
				<-chUpdate
			}()

			sgpStatus, errOnce := s.albProvider.UpdateALBServerGroup(ctx, resSGP, sdkSGP)
			if errUpdate == nil && errOnce != nil {
				errUpdate = errOnce
			}
			resSGP.SetStatus(sgpStatus)

			if strings.Contains(resSGP.Spec.ServerGroupNamedKey.IngressName, util.DefaultListenerFlag) {
				return
			}
			if errOnce = s.removeServerFromServerGroup(ctx, sgpStatus.ServerGroupID, types.NamespacedName{
				Namespace: resSGP.Spec.ServerGroupNamedKey.Namespace,
				Name:      resSGP.Spec.ServerGroupNamedKey.ServiceName},
				intstr.FromInt(resSGP.Spec.ServerGroupNamedKey.ServicePort)); errCreate == nil && errOnce != nil {
				errCreate = errOnce
				return
			}
		}(resAndSDKSGP.ResSGP, resAndSDKSGP.SdkSGP)
	}
	wgUpdate.Wait()
	if errUpdate != nil {
		return errUpdate
	}

	return nil
}

func (s *serverGroupApplier) PostApply(ctx context.Context) error {
	var (
		errDelete error
		wgDelete  sync.WaitGroup
		chDelete  = make(chan struct{}, util.ServerGroupConcurrentNum)
	)
	for _, sdkSGP := range s.unmatchedSDKSGPs {
		chDelete <- struct{}{}
		wgDelete.Add(1)

		go func(sgpID string) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer func() {
				wgDelete.Done()
				<-chDelete
			}()

			if errOnce := s.albProvider.DeleteALBServerGroup(ctx, sgpID); errDelete == nil && errOnce != nil {
				errDelete = errOnce
				return
			}
		}(sdkSGP.ServerGroup.ServerGroupId)
	}
	wgDelete.Wait()
	if errDelete != nil {
		return errDelete
	}

	return nil
}

func (s *serverGroupApplier) findSDKServerGroups(ctx context.Context) ([]albmodel.ServerGroupWithTags, error) {
	stackTags := s.trackingProvider.StackTags(s.stack)
	return s.albProvider.ListALBServerGroupsWithTags(ctx, stackTags)
}

type resAndSDKServerGroupPairSGP struct {
	ResSGP *albmodel.ServerGroup
	SdkSGP albmodel.ServerGroupWithTags
}

func matchResAndSDKServerGroupsSGP(resSGPs []*albmodel.ServerGroup, sdkSGPs []albmodel.ServerGroupWithTags, resourceIDTagKey string) ([]resAndSDKServerGroupPairSGP, []*albmodel.ServerGroup, []albmodel.ServerGroupWithTags, error) {
	var matchedResAndSDKSGPs []resAndSDKServerGroupPairSGP
	var unmatchedResSGPs []*albmodel.ServerGroup
	var unmatchedSDKSGPs []albmodel.ServerGroupWithTags

	resSGPsByID := mapResServerGroupByResourceIDSGP(resSGPs)
	sdkSGPsByID, err := mapSDKServerGroupByResourceIDSGP(sdkSGPs, resourceIDTagKey)
	if err != nil {
		return nil, nil, nil, err
	}

	resSGPIDs := sets.StringKeySet(resSGPsByID)
	sdkSGPIDs := sets.StringKeySet(sdkSGPsByID)

	for _, resID := range resSGPIDs.Intersection(sdkSGPIDs).List() {
		resSGP := resSGPsByID[resID]
		sdkSGPs := sdkSGPsByID[resID]
		for _, sdkSGP := range sdkSGPs {
			matchedResAndSDKSGPs = append(matchedResAndSDKSGPs, resAndSDKServerGroupPairSGP{
				ResSGP: resSGP,
				SdkSGP: sdkSGP,
			})
		}
	}
	for _, resID := range resSGPIDs.Difference(sdkSGPIDs).List() {
		unmatchedResSGPs = append(unmatchedResSGPs, resSGPsByID[resID])
	}
	for _, resID := range sdkSGPIDs.Difference(resSGPIDs).List() {
		unmatchedSDKSGPs = append(unmatchedSDKSGPs, sdkSGPsByID[resID]...)
	}

	return matchedResAndSDKSGPs, unmatchedResSGPs, unmatchedSDKSGPs, nil
}

func mapResServerGroupByResourceIDSGP(resSGPs []*albmodel.ServerGroup) map[string]*albmodel.ServerGroup {
	resSGPsByID := make(map[string]*albmodel.ServerGroup, len(resSGPs))
	for _, resSGP := range resSGPs {
		resSGPsByID[resSGP.ID()] = resSGP
	}
	return resSGPsByID
}

func mapSDKServerGroupByResourceIDSGP(sdkSGPs []albmodel.ServerGroupWithTags, resourceIDTagKey string) (map[string][]albmodel.ServerGroupWithTags, error) {
	sdkSGPsByID := make(map[string][]albmodel.ServerGroupWithTags, 0)
	for _, sdkSGP := range sdkSGPs {
		resourceID, ok := sdkSGP.Tags[resourceIDTagKey]
		if !ok {
			return nil, errors.Errorf("unexpected serverGroup with no resourceID: %v", resourceIDTagKey)
		}
		if isOldResourceID(resourceID) {
			resourceID = calServerGroupResourceIDHashUUID(resourceID)
		}
		sdkSGPsByID[resourceID] = append(sdkSGPsByID[resourceID], sdkSGP)
	}
	return sdkSGPsByID, nil
}

func isOldResourceID(resourceID string) bool {
	return strings.Contains(resourceID, "/") &&
		strings.Contains(resourceID, "-") &&
		strings.Contains(resourceID, ":")
}

func calServerGroupResourceIDHashUUID(resourceID string) string {
	uuidHash := sha256.New()
	_, _ = uuidHash.Write([]byte(resourceID))
	uuid := hex.EncodeToString(uuidHash.Sum(nil))
	return uuid
}
