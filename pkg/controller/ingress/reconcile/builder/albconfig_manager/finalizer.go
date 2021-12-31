package albconfigmanager

import (
	"context"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	networking "k8s.io/api/networking/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
)

type FinalizerManager interface {
	AddGroupFinalizer(ctx context.Context, members []*networking.Ingress) error

	RemoveGroupFinalizer(ctx context.Context, inactiveMembers []*networking.Ingress) error
}

func NewDefaultFinalizerManager(k8sFinalizerManager helper.FinalizerManager) *defaultFinalizerManager {
	return &defaultFinalizerManager{
		k8sFinalizerManager: k8sFinalizerManager,
	}
}

var _ FinalizerManager = (*defaultFinalizerManager)(nil)

type defaultFinalizerManager struct {
	k8sFinalizerManager helper.FinalizerManager
}

func (m *defaultFinalizerManager) AddGroupFinalizer(ctx context.Context, members []*networking.Ingress) error {
	for _, member := range members {
		if err := m.k8sFinalizerManager.AddFinalizers(ctx, member, GetIngressFinalizer()); err != nil {
			return err
		}
	}
	return nil
}

func (m *defaultFinalizerManager) RemoveGroupFinalizer(ctx context.Context, inactiveMembers []*networking.Ingress) error {
	for _, ing := range inactiveMembers {
		if err := m.k8sFinalizerManager.RemoveFinalizers(ctx, ing, GetIngressFinalizer()); err != nil {
			return err
		}
	}
	return nil
}

func GetIngressFinalizer() string {
	return util.IngressFinalizer
}
