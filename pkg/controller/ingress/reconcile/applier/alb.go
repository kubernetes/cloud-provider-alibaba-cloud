package applier

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

func NewAlbLoadBalancerApplier(albProvider prvd.Provider, trackingProvider tracking.TrackingProvider, stack core.Manager, logger logr.Logger, commonReuse bool) *albLoadBalancerApplier {
	return &albLoadBalancerApplier{
		albProvider:      albProvider,
		trackingProvider: trackingProvider,
		stack:            stack,
		logger:           logger,
		commonReuse:      commonReuse,
	}
}

type albLoadBalancerApplier struct {
	albProvider      prvd.Provider
	trackingProvider tracking.TrackingProvider
	stack            core.Manager
	logger           logr.Logger
	commonReuse      bool
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
		if err := s.DeleteALB(ctx, sdkLB.LoadBalancerId); err != nil {
			return err
		}
	}
	for _, resLB := range unmatchedResLBs {
		lbStatus, err := s.CreateALB(ctx, resLB, s.trackingProvider)
		if err != nil {
			return err
		}
		resLB.SetStatus(lbStatus)
	}
	for _, resAndSDKLB := range matchedResAndSDKLBs {
		lbStatus, err := s.UpdateALB(ctx, resAndSDKLB.resLB, resAndSDKLB.sdkLB.LoadBalancer)
		if err != nil {
			return err
		}
		resAndSDKLB.resLB.SetStatus(lbStatus)
	}
	return nil
}

func (s *albLoadBalancerApplier) UpdateALB(ctx context.Context, resLB *albmodel.AlbLoadBalancer, sdkLB alb.LoadBalancer) (albmodel.LoadBalancerStatus, error) {
	var albStatus albmodel.LoadBalancerStatus
	var err error
	if s.commonReuse {
		albStatus = albmodel.LoadBalancerStatus{
			LoadBalancerID: sdkLB.LoadBalancerId,
			DNSName:        sdkLB.DNSName,
		}
	} else {
		albStatus, err = s.albProvider.UpdateALB(ctx, resLB, sdkLB, s.trackingProvider)
	}
	return albStatus, err
}

func (s *albLoadBalancerApplier) CreateALB(ctx context.Context, resLB *albmodel.AlbLoadBalancer, trackingProvider tracking.TrackingProvider) (albmodel.LoadBalancerStatus, error) {
	var albStatus albmodel.LoadBalancerStatus
	var err error
	var isReuseLb bool
	if v, ok := ctx.Value(util.IsReuseLb).(bool); ok {
		isReuseLb = v
	}
	if isReuseLb {
		albStatus, err = s.albProvider.ReuseALB(ctx, resLB, resLB.Spec.LoadBalancerId, s.trackingProvider)
	} else {
		albStatus, err = s.albProvider.CreateALB(ctx, resLB, s.trackingProvider)
	}
	return albStatus, err
}

func (s *albLoadBalancerApplier) DeleteALB(ctx context.Context, lbID string) error {
	var err error
	var isReuseLb bool
	if v, ok := ctx.Value(util.IsReuseLb).(bool); ok {
		isReuseLb = v
	}
	if isReuseLb {
		err = s.albProvider.UnReuseALB(ctx, lbID, s.trackingProvider)
	} else {
		err = s.albProvider.DeleteALB(ctx, lbID)
	}
	return err
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
