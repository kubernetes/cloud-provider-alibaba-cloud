package alb

import (
	"context"
	"fmt"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
)

func transTagFilterToListServerGroupTags(tagFilters map[string]string) []albsdk.ListServerGroupsTag {
	listTags := make([]albsdk.ListServerGroupsTag, 0)
	for k, v := range tagFilters {
		listTags = append(listTags, albsdk.ListServerGroupsTag{
			Key:   k,
			Value: v,
		})
	}
	return listTags
}
func transTagFilterToListAlbLoadBalancersTags(tagFilters map[string]string) []albsdk.ListLoadBalancersTag {
	listTags := make([]albsdk.ListLoadBalancersTag, 0)
	for k, v := range tagFilters {
		listTags = append(listTags, albsdk.ListLoadBalancersTag{
			Key:   k,
			Value: v,
		})
	}
	return listTags
}

func (m *ALBProvider) ListALBServerGroupsByTag(ctx context.Context, tagFilters map[string]string) ([]albsdk.ServerGroup, error) {
	traceID := ctx.Value(util.TraceID)

	if len(tagFilters) == 0 {
		return nil, fmt.Errorf("invalid tag filter: %v for listing server groups", tagFilters)
	}

	listTags := transTagFilterToListServerGroupTags(tagFilters)

	var (
		nextToken    string
		serverGroups []albsdk.ServerGroup
	)

	sgpReq := albsdk.CreateListServerGroupsRequest()
	sgpReq.Tag = &listTags

	for {
		sgpReq.NextToken = nextToken

		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("listing server groups by tag",
			"tags", listTags,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.ListALBServerGroups)
		sgpResp, err := m.auth.ALB.ListServerGroups(sgpReq)
		if err != nil {
			return nil, err
		}
		m.logger.V(util.MgrLogLevel).Info("listed server groups by tag",
			"requestID", sgpResp.RequestId,
			"traceID", traceID,
			"serverGroups", sgpResp.ServerGroups,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListALBServerGroups)

		serverGroups = append(serverGroups, sgpResp.ServerGroups...)

		if sgpResp.NextToken == "" {
			break
		} else {
			nextToken = sgpResp.NextToken
		}
	}

	return serverGroups, nil
}
func (m *ALBProvider) ListAlbLoadBalancersByTag(ctx context.Context, tagFilters map[string]string) ([]albsdk.LoadBalancer, error) {
	traceID := ctx.Value(util.TraceID)

	if len(tagFilters) == 0 {
		return nil, fmt.Errorf("invalid tag filter: %v for listing load balancers", tagFilters)
	}

	listTags := transTagFilterToListAlbLoadBalancersTags(tagFilters)

	var (
		nextToken     string
		loadBalancers []albsdk.LoadBalancer
	)

	lbReq := albsdk.CreateListLoadBalancersRequest()
	lbReq.Tag = &listTags

	for {
		lbReq.NextToken = nextToken

		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("listing loadBalancers by tag",
			"tags", listTags,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.ListALBLoadBalancers)
		lbResp, err := m.auth.ALB.ListLoadBalancers(lbReq)
		if err != nil {
			return nil, err
		}
		m.logger.V(util.MgrLogLevel).Info("listed loadBalancers by tag",
			"traceID", traceID,
			"requestID", lbResp.RequestId,
			"loadBalancers", lbResp.LoadBalancers,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListALBLoadBalancers)

		loadBalancers = append(loadBalancers, lbResp.LoadBalancers...)

		if lbResp.NextToken == "" {
			break
		} else {
			nextToken = lbResp.NextToken
		}
	}

	return loadBalancers, nil
}

func (m *ALBProvider) ListALBServerGroupsWithTags(ctx context.Context, tagFilters map[string]string) ([]alb.ServerGroupWithTags, error) {
	serverGroups, err := m.ListALBServerGroupsByTag(ctx, tagFilters)
	if err != nil {
		return nil, err
	}

	serverGroupsWithTags := make([]alb.ServerGroupWithTags, 0)
	for _, serverGroup := range serverGroups {
		tagMap := transSDKTagListToMap(serverGroup.Tags)

		serverGroupsWithTags = append(serverGroupsWithTags, alb.ServerGroupWithTags{
			ServerGroup: serverGroup,
			Tags:        tagMap,
		})
	}

	return serverGroupsWithTags, nil
}
func (m *ALBProvider) ListALBsWithTags(ctx context.Context, tagFilters map[string]string) ([]alb.AlbLoadBalancerWithTags, error) {
	traceID := ctx.Value(util.TraceID)

	lbs, err := m.ListAlbLoadBalancersByTag(ctx, tagFilters)
	if err != nil {
		return nil, err
	}

	lbsWithTags := make([]alb.AlbLoadBalancerWithTags, 0)
	for _, lb := range lbs {
		getLbReq := albsdk.CreateGetLoadBalancerAttributeRequest()
		getLbReq.LoadBalancerId = lb.LoadBalancerId
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("getting loadBalancer attribute",
			"loadBalancerID", lb.LoadBalancerId,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.GetALBLoadBalancerAttribute)
		getLbResp, err := m.auth.ALB.GetLoadBalancerAttribute(getLbReq)
		if err != nil {
			return nil, err
		}
		m.logger.V(util.MgrLogLevel).Info("got loadBalancer attribute",
			"loadBalancerID", lb.LoadBalancerId,
			"traceID", traceID,
			"requestID", getLbResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.GetALBLoadBalancerAttribute)

		tagMap := transSDKTagListToMap(lb.Tags)

		lbsWithTags = append(lbsWithTags, alb.AlbLoadBalancerWithTags{
			LoadBalancer: *transSDKGetAlbLoadBalancerAttributeResponseToNormal(getLbResp),
			Tags:         tagMap,
		})
	}

	return lbsWithTags, nil
}

func transSDKGetAlbLoadBalancerAttributeResponseToNormal(resp *albsdk.GetLoadBalancerAttributeResponse) *albsdk.LoadBalancer {
	return &albsdk.LoadBalancer{
		AddressAllocatedMode:         resp.AddressAllocatedMode,
		AddressType:                  resp.AddressType,
		BandwidthCapacity:            resp.BandwidthCapacity,
		BandwidthPackageId:           resp.BandwidthPackageId,
		CreateTime:                   resp.CreateTime,
		DNSName:                      resp.DNSName,
		LoadBalancerBussinessStatus:  resp.LoadBalancerStatus,
		LoadBalancerEdition:          resp.LoadBalancerEdition,
		LoadBalancerId:               resp.LoadBalancerId,
		LoadBalancerName:             resp.LoadBalancerName,
		LoadBalancerStatus:           resp.LoadBalancerStatus,
		ResourceGroupId:              resp.ResourceGroupId,
		VpcId:                        resp.VpcId,
		AccessLogConfig:              resp.AccessLogConfig,
		DeletionProtectionConfig:     resp.DeletionProtectionConfig,
		LoadBalancerBillingConfig:    resp.LoadBalancerBillingConfig,
		ModificationProtectionConfig: resp.ModificationProtectionConfig,
		LoadBalancerOperationLocks:   resp.LoadBalancerOperationLocks,
		Tags:                         resp.Tags,
	}
}
