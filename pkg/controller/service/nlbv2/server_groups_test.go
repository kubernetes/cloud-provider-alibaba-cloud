package nlbv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
)

func makeNLBBackends(n int) []nlbmodel.ServerGroupServer {
	backends := make([]nlbmodel.ServerGroupServer, n)
	for i := range backends {
		backends[i].ServerId = "ecs-default"
		backends[i].ServerType = nlbmodel.EcsServerType
	}
	return backends
}

func newSvcWithAnnotations(annos map[string]string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ServiceName,
			Namespace:   v1.NamespaceDefault,
			Annotations: annos,
		},
	}
}

// --- setWeightBackends ---

func TestNLBSetWeightBackends_NilWeightNilDefault_UsesDefaultServerWeight(t *testing.T) {
	backends := makeNLBBackends(3)
	result := setWeightBackends(helper.ENITrafficPolicy, backends, nil, nil)
	for _, b := range result {
		assert.Equal(t, int32(DefaultServerWeight), b.Weight)
	}
}

func TestNLBSetWeightBackends_NilWeightWithCustomDefault_UsesCustomDefaultWeight(t *testing.T) {
	backends := makeNLBBackends(3)
	customDefault := 40
	result := setWeightBackends(helper.ENITrafficPolicy, backends, nil, &customDefault)
	for _, b := range result {
		assert.Equal(t, int32(40), b.Weight)
	}
}

func TestNLBSetWeightBackends_WithWeight_IgnoresDefaultWeight(t *testing.T) {
	backends := makeNLBBackends(2)
	weight := 60
	customDefault := 40
	// podPercentAlgorithm: 60 total / 2 backends = 30 each
	result := setWeightBackends(helper.ENITrafficPolicy, backends, &weight, &customDefault)
	for _, b := range result {
		assert.Equal(t, int32(30), b.Weight)
	}
}

// --- podNumberAlgorithm ---

func TestNLBPodNumberAlgorithm_ENIMode_AppliesCustomDefaultWeight(t *testing.T) {
	backends := makeNLBBackends(3)
	result := podNumberAlgorithm(helper.ENITrafficPolicy, backends, 75)
	for _, b := range result {
		assert.Equal(t, int32(75), b.Weight)
	}
}

func TestNLBPodNumberAlgorithm_ClusterMode_AppliesCustomDefaultWeight(t *testing.T) {
	backends := makeNLBBackends(3)
	result := podNumberAlgorithm(helper.ClusterTrafficPolicy, backends, 30)
	for _, b := range result {
		assert.Equal(t, int32(30), b.Weight)
	}
}

func TestNLBPodNumberAlgorithm_LocalMode_WeightByPodCount(t *testing.T) {
	backends := []nlbmodel.ServerGroupServer{
		{ServerId: "ecs-1", ServerType: nlbmodel.EcsServerType},
		{ServerId: "ecs-1", ServerType: nlbmodel.EcsServerType},
		{ServerId: "ecs-2", ServerType: nlbmodel.EcsServerType},
	}
	result := podNumberAlgorithm(helper.LocalTrafficPolicy, backends, 50)
	assert.Equal(t, int32(2), result[0].Weight) // ecs-1 has 2 pods
	assert.Equal(t, int32(2), result[1].Weight)
	assert.Equal(t, int32(1), result[2].Weight) // ecs-2 has 1 pod
}

// --- setServerGroupAttributeFromAnno ---

func TestSetServerGroupAttributeFromAnno_ValidDefaultWeight(t *testing.T) {
	svc := newSvcWithAnnotations(map[string]string{
		annotation.Annotation(annotation.DefaultWeight): "50",
	})
	sg := &nlbmodel.ServerGroup{}
	err := setServerGroupAttributeFromAnno(sg, annotation.NewAnnotationRequest(svc))
	assert.NoError(t, err)
	assert.NotNil(t, sg.DefaultWeight)
	assert.Equal(t, 50, *sg.DefaultWeight)
}

func TestSetServerGroupAttributeFromAnno_ZeroDefaultWeight(t *testing.T) {
	svc := newSvcWithAnnotations(map[string]string{
		annotation.Annotation(annotation.DefaultWeight): "0",
	})
	sg := &nlbmodel.ServerGroup{}
	err := setServerGroupAttributeFromAnno(sg, annotation.NewAnnotationRequest(svc))
	assert.NoError(t, err)
	assert.NotNil(t, sg.DefaultWeight)
	assert.Equal(t, 0, *sg.DefaultWeight)
}

func TestSetServerGroupAttributeFromAnno_NoDefaultWeight_DoesNotSetField(t *testing.T) {
	svc := newSvcWithAnnotations(map[string]string{})
	sg := &nlbmodel.ServerGroup{}
	err := setServerGroupAttributeFromAnno(sg, annotation.NewAnnotationRequest(svc))
	assert.NoError(t, err)
	assert.Nil(t, sg.DefaultWeight)
}

func TestSetServerGroupAttributeFromAnno_InvalidDefaultWeight_NotInteger(t *testing.T) {
	svc := newSvcWithAnnotations(map[string]string{
		annotation.Annotation(annotation.DefaultWeight): "abc",
	})
	sg := &nlbmodel.ServerGroup{}
	err := setServerGroupAttributeFromAnno(sg, annotation.NewAnnotationRequest(svc))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default weight parse error")
}

func TestSetServerGroupAttributeFromAnno_InvalidDefaultWeight_TooLarge(t *testing.T) {
	svc := newSvcWithAnnotations(map[string]string{
		annotation.Annotation(annotation.DefaultWeight): "200",
	})
	sg := &nlbmodel.ServerGroup{}
	err := setServerGroupAttributeFromAnno(sg, annotation.NewAnnotationRequest(svc))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default weight must be integer in range [0,100]")
}

func TestSetServerGroupAttributeFromAnno_InvalidDefaultWeight_Negative(t *testing.T) {
	svc := newSvcWithAnnotations(map[string]string{
		annotation.Annotation(annotation.DefaultWeight): "-5",
	})
	sg := &nlbmodel.ServerGroup{}
	err := setServerGroupAttributeFromAnno(sg, annotation.NewAnnotationRequest(svc))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default weight must be integer in range [0,100]")
}
