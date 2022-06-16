package tracking

import (
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

type TrackingProvider interface {
	ResourceIDTagKey() string

	AlbConfigTagKey() string

	ClusterNameTagKey() string

	StackTags(stack core.Manager) map[string]string

	ResourceTags(stack core.Manager, res core.Resource, additionalTags map[string]string) map[string]string
}

func NewDefaultProvider(tagPrefix, clusterID string) *defaultTrackingProvider {
	return &defaultTrackingProvider{
		tagPrefix: tagPrefix,
		clusterID: clusterID,
	}
}

var _ TrackingProvider = &defaultTrackingProvider{}

type defaultTrackingProvider struct {
	tagPrefix string
	clusterID string
}

func (p *defaultTrackingProvider) ResourceIDTagKey() string {
	return p.prefixedTrackingKey("resource")
}

func (p *defaultTrackingProvider) AlbConfigTagKey() string {
	return p.prefixedTrackingKey(util.AlbConfigTagKey)
}

func (p *defaultTrackingProvider) ClusterNameTagKey() string {
	return util.ClusterTagKey
}

func (p *defaultTrackingProvider) StackTags(stack core.Manager) map[string]string {
	stackID := stack.StackID()
	return map[string]string{
		p.ClusterNameTagKey(): p.clusterID,
		p.AlbConfigTagKey():   stackID.String(),
	}
}

// ResourceTags
// ack.aliyun.com ${cluster_id}
// ingress.k8s.alibaba/resource ApplicationLoadBalancer
// ingress.k8s.alibaba/albconfig ${albconfig_ns}/${albconfig_name}
func (p *defaultTrackingProvider) ResourceTags(stack core.Manager, res core.Resource, additionalTags map[string]string) map[string]string {
	stackTags := p.StackTags(stack)
	resourceIDTags := map[string]string{
		p.ResourceIDTagKey(): res.ID(),
	}
	return util.MergeStringMap(stackTags, resourceIDTags, additionalTags)
}

func (p *defaultTrackingProvider) prefixedTrackingKey(tag string) string {
	return fmt.Sprintf("%v/%v", p.tagPrefix, tag)
}
