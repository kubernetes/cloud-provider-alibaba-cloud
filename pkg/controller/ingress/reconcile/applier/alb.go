package applier

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

func NewAlbLoadBalancerApplier(albProvider prvd.Provider, trackingProvider tracking.TrackingProvider, stack core.Manager, logger logr.Logger) *albLoadBalancerApplier {
	return &albLoadBalancerApplier{
		albProvider:      albProvider,
		trackingProvider: trackingProvider,
		stack:            stack,
		logger:           logger,
	}
}

type albLoadBalancerApplier struct {
	albProvider      prvd.Provider
	trackingProvider tracking.TrackingProvider
	stack            core.Manager
	logger           logr.Logger
}

func (s *albLoadBalancerApplier) Apply(ctx context.Context) error {
	traceID := ctx.Value(util.TraceID)

	var resLBs []*albmodel.AlbLoadBalancer
	_ = s.stack.ListResources(&resLBs)

	if len(resLBs) > 1 {
		return fmt.Errorf("invalid res loadBalancers, at most one loadBalancer for stack: %s", s.stack.StackID())
	}

	sdkLBs, err := s.findSDKAlbLoadBalancers(ctx)
	if err != nil {
		return err
	}
	if len(sdkLBs) > 1 {
		stackID := s.stack.StackID()
		return fmt.Errorf("invalid sdk loadBalancers: %v, at most one loadBalancer can find by stack tag: %s, must delete manually", sdkLBs, stackID)
	}

	matchedResAndSDKLBs, unmatchedResLBs, unmatchedSDKLBs, err := matchResAndSDKAlbLoadBalancers(resLBs, sdkLBs, s.trackingProvider.ResourceIDTagKey())
	if err != nil {
		return err
	}

	if len(matchedResAndSDKLBs) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize loadBalancers",
			"matchedResAndSDKLBs", matchedResAndSDKLBs,
			"traceID", traceID)
	}
	if len(unmatchedResLBs) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize loadBalancers",
			"unmatchedResLBs", unmatchedResLBs,
			"traceID", traceID)
	}
	if len(unmatchedSDKLBs) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize loadBalancers",
			"unmatchedSDKLBs", unmatchedSDKLBs,
			"traceID", traceID)
	}

	for _, sdkLB := range unmatchedSDKLBs {
		if err := s.albProvider.DeleteALB(ctx, sdkLB.LoadBalancerId); err != nil {
			return err
		}
	}
	for _, resLB := range unmatchedResLBs {
		var isReuseLb bool
		if len(resLB.Spec.LoadBalancerId) != 0 {
			isReuseLb = true
		}
		if isReuseLb {
			lbStatus, err := s.albProvider.ReuseALB(ctx, resLB, resLB.Spec.LoadBalancerId, s.trackingProvider)
			if err != nil {
				return err
			}
			resLB.SetStatus(lbStatus)
			continue
		}

		lbStatus, err := s.albProvider.CreateALB(ctx, resLB, s.trackingProvider)
		if err != nil {
			return err
		}
		resLB.SetStatus(lbStatus)
	}
	for _, resAndSDKLB := range matchedResAndSDKLBs {
		var isReuseLb bool
		if len(resAndSDKLB.resLB.Spec.LoadBalancerId) != 0 {
			isReuseLb = true
		}
		if isReuseLb {
			if resAndSDKLB.resLB.Spec.ForceOverride != nil && !*resAndSDKLB.resLB.Spec.ForceOverride {
				resAndSDKLB.resLB.SetStatus(albmodel.LoadBalancerStatus{
					LoadBalancerID: resAndSDKLB.sdkLB.LoadBalancerId,
					DNSName:        resAndSDKLB.sdkLB.DNSName,
				})
				continue
			}
		}

		lbStatus, err := s.albProvider.UpdateALB(ctx, resAndSDKLB.resLB, resAndSDKLB.sdkLB.LoadBalancer)
		if err != nil {
			return err
		}
		resAndSDKLB.resLB.SetStatus(lbStatus)
	}
	return nil
}

func (s *albLoadBalancerApplier) PostApply(ctx context.Context) error {
	return nil
}

func (s *albLoadBalancerApplier) findSDKAlbLoadBalancers(ctx context.Context) ([]albmodel.AlbLoadBalancerWithTags, error) {
	stackTags := s.trackingProvider.StackTags(s.stack)
	return s.albProvider.ListALBsWithTags(ctx, stackTags)
}

type resAndSDKLoadBalancerPair struct {
	resLB *albmodel.AlbLoadBalancer
	sdkLB *albmodel.AlbLoadBalancerWithTags
}

func matchResAndSDKAlbLoadBalancers(resLBs []*albmodel.AlbLoadBalancer, sdkLBs []albmodel.AlbLoadBalancerWithTags, resourceIDTagKey string) ([]resAndSDKLoadBalancerPair, []*albmodel.AlbLoadBalancer, []albmodel.AlbLoadBalancerWithTags, error) {
	var matchedResAndSDKLBs []resAndSDKLoadBalancerPair
	var unmatchedResLBs []*albmodel.AlbLoadBalancer
	var unmatchedSDKLBs []albmodel.AlbLoadBalancerWithTags

	resLBsByID := mapResAlbLoadBalancerByResourceID(resLBs)
	sdkLBsByID, err := mapSDKAlbLoadBalancerByResourceID(sdkLBs, resourceIDTagKey)
	if err != nil {
		return nil, nil, nil, err
	}

	resLBIDs := sets.StringKeySet(resLBsByID)
	sdkLBIDs := sets.StringKeySet(sdkLBsByID)
	for _, resID := range resLBIDs.Intersection(sdkLBIDs).List() {
		resLB := resLBsByID[resID]
		sdkLBs := sdkLBsByID[resID]
		for _, sdkLB := range sdkLBs {
			matchedResAndSDKLBs = append(matchedResAndSDKLBs, resAndSDKLoadBalancerPair{
				resLB: resLB,
				sdkLB: &sdkLB,
			})
		}
	}
	for _, resID := range resLBIDs.Difference(sdkLBIDs).List() {
		unmatchedResLBs = append(unmatchedResLBs, resLBsByID[resID])
	}
	for _, resID := range sdkLBIDs.Difference(resLBIDs).List() {
		unmatchedSDKLBs = append(unmatchedSDKLBs, sdkLBsByID[resID]...)
	}

	return matchedResAndSDKLBs, unmatchedResLBs, unmatchedSDKLBs, nil
}

func mapResAlbLoadBalancerByResourceID(resLBs []*albmodel.AlbLoadBalancer) map[string]*albmodel.AlbLoadBalancer {
	resLBsByID := make(map[string]*albmodel.AlbLoadBalancer, len(resLBs))
	for _, resLB := range resLBs {
		resLBsByID[resLB.ID()] = resLB
	}
	return resLBsByID
}

func mapSDKAlbLoadBalancerByResourceID(sdkLBs []albmodel.AlbLoadBalancerWithTags, resourceIDTagKey string) (map[string][]albmodel.AlbLoadBalancerWithTags, error) {
	sdkLBsByID := make(map[string][]albmodel.AlbLoadBalancerWithTags, len(sdkLBs))
	for _, sdkLB := range sdkLBs {
		resourceID, ok := sdkLB.Tags[resourceIDTagKey]
		if !ok {
			return nil, errors.Errorf("unexpected loadBalancer with no resourceID: %v", sdkLB.LoadBalancer.LoadBalancerId)
		}
		sdkLBsByID[resourceID] = append(sdkLBsByID[resourceID], sdkLB)
	}
	return sdkLBsByID, nil
}
