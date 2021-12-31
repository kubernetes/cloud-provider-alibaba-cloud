package annotations

import (
	networking "k8s.io/api/networking/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
)

type ClassAnnotationMatcher interface {
	Matches(ingClassAnnotation string) bool
}

func NewDefaultClassAnnotationMatcher(ingressClass string) *defaultClassAnnotationMatcher {
	return &defaultClassAnnotationMatcher{
		ingressClass: ingressClass,
	}
}

var _ ClassAnnotationMatcher = &defaultClassAnnotationMatcher{}

type defaultClassAnnotationMatcher struct {
	ingressClass string
}

func (m *defaultClassAnnotationMatcher) Matches(ingClassAnnotation string) bool {
	if m.ingressClass == "" && ingClassAnnotation == util.IngressClassALB {
		return true
	}
	return ingClassAnnotation == m.ingressClass
}

func IsIngressAlbClass(ing networking.Ingress) bool {
	classAnnotationMatcher := NewDefaultClassAnnotationMatcher(util.IngressClassALB)

	if ingClassAnnotation, exists := ing.Annotations[util.IngressClass]; exists {
		if matchesIngressClass := classAnnotationMatcher.Matches(ingClassAnnotation); matchesIngressClass {
			return true
		}
	}
	return false
}
