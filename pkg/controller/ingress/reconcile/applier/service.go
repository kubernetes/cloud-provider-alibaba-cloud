package applier

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	"k8s.io/apimachinery/pkg/util/sets"

	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

type ServiceManagerApplier interface {
	Apply(ctx context.Context, albProvider prvd.Provider, serviceStack *albmodel.ServiceManager) error
}

var _ ServiceManagerApplier = &defaultServiceManagerApplier{}

func NewServiceManagerApplier(kubeClient client.Client, albProvider prvd.Provider, logger logr.Logger) *defaultServiceManagerApplier {
	return &defaultServiceManagerApplier{
		kubeClient:  kubeClient,
		albProvider: albProvider,
		logger:      logger,
	}
}

type defaultServiceManagerApplier struct {
	kubeClient  client.Client
	albProvider prvd.Provider

	logger logr.Logger
}

func (m *defaultServiceManagerApplier) Apply(ctx context.Context, albProvider prvd.Provider, serviceStack *albmodel.ServiceManager) error {
	serverGroupApplier := NewServiceStackApplier(albProvider, serviceStack, m.logger)
	if err := serverGroupApplier.Apply(ctx); err != nil {
		return err
	}

	matchedResAndSDKSGPs := serverGroupApplier.MatchedResAndSDKSGPs

	var (
		err     error
		wg      sync.WaitGroup
		chApply = make(chan struct{}, util.ServerGroupConcurrentNum)
	)
	for _, v := range matchedResAndSDKSGPs {
		chApply <- struct{}{}
		wg.Add(1)

		go func(serverGroupID string, backends []albmodel.BackendItem) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer func() {
				wg.Done()
				<-chApply
			}()

			serverApplier := NewServerApplier(m.kubeClient, albProvider, serverGroupID, backends, serviceStack.TrafficPolicy, m.logger)
			if errOnce := serverApplier.Apply(ctx); err == nil && errOnce != nil {
				m.logger.Error(errOnce, "synthesize servers failed", "serverGroupID", serverGroupID)
				err = errOnce
			}
		}(v.SdkSGP.ServerGroupId, v.ResSGP.Backends)
	}
	wg.Wait()
	if err != nil {
		return err
	}

	return nil
}

func NewServiceStackApplier(albProvider prvd.Provider, serviceStack *albmodel.ServiceManager, logger logr.Logger) *serviceStackApplier {
	tagFilters := make(map[string]string)
	tagFilters[util.ClusterNameTagKey] = serviceStack.ClusterID
	tagFilters[util.ServiceNamespaceTagKey] = serviceStack.Namespace
	tagFilters[util.ServiceNameTagKey] = serviceStack.Name

	return &serviceStackApplier{
		serviceStack: serviceStack,
		albProvider:  albProvider,
		tagFilters:   tagFilters,
		logger:       logger,
	}
}

type serviceStackApplier struct {
	serviceStack         *albmodel.ServiceManager
	albProvider          prvd.Provider
	tagFilters           map[string]string
	MatchedResAndSDKSGPs []resAndSDKServerGroupPair

	logger logr.Logger
}

func (s *serviceStackApplier) Apply(ctx context.Context) error {
	traceID := ctx.Value(util.TraceID)

	serverGroupsWithNameKey := transServiceStackToServerGroupsWithNameKey(s.serviceStack)
	serverGroupsWithTags, err := s.albProvider.ListALBServerGroupsWithTags(ctx, s.tagFilters)
	if err != nil {
		return err
	}

	matchedResAndSDKSGPs, unmatchedResSGPs, unmatchedSDKSGPs, err := matchResAndSDKServerGroups(serverGroupsWithNameKey, serverGroupsWithTags)
	if err != nil {
		return err
	}

	if len(matchedResAndSDKSGPs) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize serviceStack",
			"matchedResAndSDKSGPs", matchedResAndSDKSGPs,
			"traceID", traceID)
	}
	if len(unmatchedResSGPs) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize serviceStack",
			"unmatchedResSGPs", unmatchedResSGPs,
			"traceID", traceID)
	}
	if len(unmatchedSDKSGPs) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize serviceStack",
			"unmatchedSDKSGPs", unmatchedSDKSGPs,
			"traceID", traceID)
	}

	s.MatchedResAndSDKSGPs = matchedResAndSDKSGPs
	return nil
}

func (s *serviceStackApplier) PostApply(ctx context.Context) error {
	return nil
}

func transServiceStackToServerGroupsWithNameKey(serviceStack *albmodel.ServiceManager) []albmodel.ServiceGroupWithNameKey {
	serverGroups := make([]albmodel.ServiceGroupWithNameKey, 0)

	for port, serverGroup := range serviceStack.PortToServerGroup {
		for _, ingressName := range serverGroup.IngressNames {
			serverGroupNamedKey := &albmodel.ServerGroupNamedKey{
				ClusterID:   serviceStack.ClusterID,
				Namespace:   serviceStack.Namespace,
				IngressName: ingressName,
				ServiceName: serviceStack.Name,
				ServicePort: int(port),
			}
			albconfig := serviceStack.IngressAlbConfigMap[serviceStack.Namespace+"/"+ingressName]

			serverGroups = append(serverGroups, albmodel.ServiceGroupWithNameKey{
				NamedKey:     serverGroupNamedKey,
				AlbConfigKey: albconfig,
				Backends:     serverGroup.Backends,
			})
		}
	}

	return serverGroups
}

type resAndSDKServerGroupPair struct {
	ResSGP albmodel.ServiceGroupWithNameKey
	SdkSGP albmodel.ServerGroupWithTags
}

func mapResServerGroupByResourceID(resSGPs []albmodel.ServiceGroupWithNameKey) map[string]albmodel.ServiceGroupWithNameKey {
	resSGPsByID := make(map[string]albmodel.ServiceGroupWithNameKey)
	for _, resSGP := range resSGPs {
		resID := fmt.Sprintf("%s_%s", resSGP.NamedKey.Key(), resSGP.AlbConfigKey)
		resSGPsByID[resID] = resSGP
	}
	return resSGPsByID
}

func mapSDKServerGroupByResourceID(sdkSGPs []albmodel.ServerGroupWithTags) map[string]albmodel.ServerGroupWithTags {
	resSGPsByID := make(map[string]albmodel.ServerGroupWithTags)
	for _, sdkSGP := range sdkSGPs {
		var svcNameKey albmodel.ServerGroupNamedKey
		var svcAlbConfig string

		tags := sdkSGP.Tags
		if v, ok := tags[util.AlbConfigFullTagKey]; ok {
			svcAlbConfig = v
		} else {
			continue
		}
		if v, ok := tags[util.ClusterNameTagKey]; ok {
			svcNameKey.ClusterID = v
		} else {
			continue
		}

		if v, ok := tags[util.ServiceNamespaceTagKey]; ok {
			svcNameKey.Namespace = v
		} else {
			continue
		}

		if v, ok := tags[util.IngressNameTagKey]; ok {
			svcNameKey.IngressName = v
		} else {
			continue
		}

		if v, ok := tags[util.ServiceNameTagKey]; ok {
			svcNameKey.ServiceName = v
		} else {
			continue
		}

		if v, ok := tags[util.ServicePortTagKey]; ok {
			intV, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			svcNameKey.ServicePort = intV
		} else {
			continue
		}
		sdkTagId := fmt.Sprintf("%s_%s", svcNameKey.Key(), svcAlbConfig)
		resSGPsByID[sdkTagId] = sdkSGP
	}
	return resSGPsByID
}

func matchResAndSDKServerGroups(resSGPs []albmodel.ServiceGroupWithNameKey, sdkSGPs []albmodel.ServerGroupWithTags) ([]resAndSDKServerGroupPair, []albmodel.ServiceGroupWithNameKey, []albmodel.ServerGroupWithTags, error) {
	var matchedResAndSDKSGPs []resAndSDKServerGroupPair
	var unmatchedResSGPs []albmodel.ServiceGroupWithNameKey
	var unmatchedSDKSGPs []albmodel.ServerGroupWithTags

	resSGPsByID := mapResServerGroupByResourceID(resSGPs)
	sdkSGPsByID := mapSDKServerGroupByResourceID(sdkSGPs)

	resSGPIDs := sets.StringKeySet(resSGPsByID)
	sdkSGPIDs := sets.StringKeySet(sdkSGPsByID)

	for _, resID := range resSGPIDs.Intersection(sdkSGPIDs).List() {
		resSGP := resSGPsByID[resID]
		sdkSGPs := sdkSGPsByID[resID]
		matchedResAndSDKSGPs = append(matchedResAndSDKSGPs, resAndSDKServerGroupPair{
			ResSGP: resSGP,
			SdkSGP: sdkSGPs,
		})
	}
	for _, resID := range resSGPIDs.Difference(sdkSGPIDs).List() {
		unmatchedResSGPs = append(unmatchedResSGPs, resSGPsByID[resID])
	}
	for _, resID := range sdkSGPIDs.Difference(resSGPIDs).List() {
		unmatchedSDKSGPs = append(unmatchedSDKSGPs, sdkSGPsByID[resID])
	}

	return matchedResAndSDKSGPs, unmatchedResSGPs, unmatchedSDKSGPs, nil
}
