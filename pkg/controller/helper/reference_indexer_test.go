package helper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	networking "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewDefaultReferenceIndexer(t *testing.T) {
	indexer := NewDefaultReferenceIndexer()
	assert.NotNil(t, indexer)
	assert.IsType(t, &defaultReferenceIndexer{}, indexer)
}

func TestDefaultReferenceIndexer_BuildServiceRefIndexes_EmptyIngress(t *testing.T) {
	indexer := NewDefaultReferenceIndexer()
	ingress := &networking.Ingress{}

	result := indexer.BuildServiceRefIndexes(context.Background(), ingress)
	assert.Empty(t, result)
}

func TestDefaultReferenceIndexer_BuildServiceRefIndexes_OnlyDefaultBackend(t *testing.T) {
	indexer := NewDefaultReferenceIndexer()
	ingress := &networking.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "default-service",
				},
			},
		},
	}

	result := indexer.BuildServiceRefIndexes(context.Background(), ingress)
	assert.Len(t, result, 1)
	assert.Contains(t, result, "default-service")
}

func TestDefaultReferenceIndexer_BuildServiceRefIndexes_OnlyHTTPPaths(t *testing.T) {
	indexer := NewDefaultReferenceIndexer()
	ingress := &networking.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: "service1",
										},
									},
								},
								{
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: "service2",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := indexer.BuildServiceRefIndexes(context.Background(), ingress)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "service1")
	assert.Contains(t, result, "service2")
}

func TestDefaultReferenceIndexer_BuildServiceRefIndexes_MixedBackends(t *testing.T) {
	indexer := NewDefaultReferenceIndexer()
	ingress := &networking.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "default-service",
				},
			},
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: "path-service1",
										},
									},
								},
								{
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: "path-service2",
										},
									},
								},
							},
						},
					},
				},
				{
					// Rule without HTTP
				},
			},
		},
	}

	result := indexer.BuildServiceRefIndexes(context.Background(), ingress)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "default-service")
	assert.Contains(t, result, "path-service1")
	assert.Contains(t, result, "path-service2")
}

func TestDefaultReferenceIndexer_BuildServiceRefIndexes_DuplicateServices(t *testing.T) {
	indexer := NewDefaultReferenceIndexer()
	ingress := &networking.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "duplicate-service",
				},
			},
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: "duplicate-service",
										},
									},
								},
								{
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: "other-service",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := indexer.BuildServiceRefIndexes(context.Background(), ingress)
	assert.Len(t, result, 2) // Should deduplicate
	assert.Contains(t, result, "duplicate-service")
	assert.Contains(t, result, "other-service")
}
