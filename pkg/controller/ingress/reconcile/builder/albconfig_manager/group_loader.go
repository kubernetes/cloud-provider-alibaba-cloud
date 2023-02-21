package albconfigmanager

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiextcli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/store"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pkg/errors"
)

var (
	errInvalidIngressGroup = errors.New("invalid ingress group")
	DefaultGroupName       = "default"
	ALBConfigNamespace     = "kube-system"
)

type GroupID types.NamespacedName

func (groupID GroupID) String() string {
	return fmt.Sprintf("%s/%s", groupID.Namespace, groupID.Name)
}

type Group struct {
	ID GroupID

	Members []*networking.Ingress

	InactiveMembers []*networking.Ingress
}

type GroupLoader interface {
	Load(ctx context.Context, groupID GroupID, ingress []*store.Ingress) (*Group, error)

	LoadGroupID(ctx context.Context, ing *networking.Ingress) (*GroupID, error)
}

func NewDefaultGroupLoader(kubeClient client.Client, kubeClientCache cache.Cache, client apiextcli.Interface, annotationParser annotations.Parser) *defaultGroupLoader {
	return &defaultGroupLoader{
		annotationParser: annotationParser,
		kubeClient:       kubeClient,
		kubeClientCache:  kubeClientCache,
		client:           client,
	}
}

var _ GroupLoader = (*defaultGroupLoader)(nil)

type defaultGroupLoader struct {
	annotationParser annotations.Parser
	kubeClient       client.Client
	kubeClientCache  cache.Cache
	client           apiextcli.Interface
}

func (m *defaultGroupLoader) Load(ctx context.Context, groupID GroupID, ingress []*store.Ingress) (*Group, error) {

	var members []*networking.Ingress
	var inactiveMembers []*networking.Ingress
	for _, ing := range ingress {
		groupName := ""
		if exists := m.annotationParser.ParseStringAnnotation(util.IngressSuffixAlbConfigName, &groupName, ing.Annotations); !exists {
			// default config：default
			groupName = DefaultGroupName
			albconfigName := m.IngressClass(&ing.Ingress)
			if albconfigName != "" {
				groupName = albconfigName
			}
		}
		if groupID.Namespace == ALBConfigNamespace {
			if groupID.Name != groupName {
				continue
			}
		} else {
			acrd, err := m.client.ApiextensionsV1().
				CustomResourceDefinitions().Get(ctx, "albconfigs.alibabacloud.com", metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if acrd.Spec.Scope == apiextv1.NamespaceScoped {
				if groupID.Name != groupName {
					continue
				}
			}
		}

		isGroupMember, err := m.isGroupMember(ctx, groupID, &ing.Ingress)
		if err != nil {
			return nil, errors.Wrapf(err, "ingress: %v", util.NamespacedName(ing))
		}
		if isGroupMember {
			members = append(members, &ing.Ingress)
		} else if m.containsGroupFinalizer(GetIngressFinalizer(), &ing.Ingress) {
			inactiveMembers = append(inactiveMembers, &ing.Ingress)
		}
	}

	klog.Infof("groupID: %v, members: %d, inactiveMembers: %d", groupID, len(members), len(inactiveMembers))

	sortedMembers, err := m.sortGroupMembers(members)
	if err != nil {
		return nil, err
	}

	return &Group{
		ID:              groupID,
		Members:         sortedMembers,
		InactiveMembers: inactiveMembers,
	}, nil
}

func (m *defaultGroupLoader) isGroupMember(ctx context.Context, groupID GroupID, ing *networking.Ingress) (bool, error) {
	if !ing.DeletionTimestamp.IsZero() {
		return false, nil
	}

	ingGroupID, err := m.LoadGroupID(ctx, ing)
	if err != nil {
		return false, err
	}
	if ingGroupID == nil || ingGroupID.Name == "" {
		return false, nil
	}

	return groupID == *ingGroupID, nil
}

func (m *defaultGroupLoader) LoadGroupID(ctx context.Context, ing *networking.Ingress) (*GroupID, error) {
	groupName := ""
	if exists := m.annotationParser.ParseStringAnnotation(util.IngressSuffixAlbConfigName, &groupName, ing.Annotations); !exists {
		groupName = DefaultGroupName
		albconfigName := m.IngressClass(ing)
		if albconfigName != "" {
			groupName = albconfigName
		}
	}
	acrd, err := m.client.ApiextensionsV1().
		CustomResourceDefinitions().Get(ctx, "albconfigs.alibabacloud.com", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if acrd.Spec.Scope == apiextv1.ClusterScoped {
		//if err := m.kubeClient.Get(ctx, types.NamespacedName{
		//	Name: groupName,
		//}, albconfig); err != nil {
		//	return nil, err
		//}
		groupID := GroupID(types.NamespacedName{
			Namespace: ALBConfigNamespace,
			Name:      groupName,
		})
		return &groupID, nil
	}
	albconfig := &v1.AlbConfig{}
	if err := m.kubeClientCache.Get(ctx, types.NamespacedName{
		Namespace: ing.Namespace,
		Name:      groupName,
	}, albconfig); err == nil {
		groupID := GroupID(types.NamespacedName{
			Namespace: ing.Namespace,
			Name:      groupName,
		})
		return &groupID, nil
	}
	alist := &v1.AlbConfigList{}
	err = m.kubeClientCache.List(ctx, alist)
	if err == nil {
		hasAlbConfig := ""
		for _, item := range alist.Items {
			if item.Name == groupName {
				if item.Status.LoadBalancer.DNSName != "" {
					groupID := GroupID(types.NamespacedName{
						Namespace: item.Namespace,
						Name:      groupName,
					})
					return &groupID, nil
				}
				hasAlbConfig = item.Namespace
			}
		}
		if hasAlbConfig != "" {
			groupID := GroupID(types.NamespacedName{
				Namespace: hasAlbConfig,
				Name:      groupName,
			})
			return &groupID, nil
		}
	}
	groupID := GroupID(types.NamespacedName{
		Namespace: ALBConfigNamespace,
		Name:      groupName,
	})
	return &groupID, nil
}

func (m *defaultGroupLoader) containsGroupFinalizer(finalizer string, ing *networking.Ingress) bool {
	return helper.HasFinalizer(ing, finalizer)
}

type groupMemberWithOrder struct {
	member *networking.Ingress
	order  int64
}

const (
	defaultGroupOrder int64 = 10
	minGroupOrder     int64 = 1
	maxGroupOder      int64 = 1000
)

func (m *defaultGroupLoader) sortGroupMembers(members []*networking.Ingress) ([]*networking.Ingress, error) {
	if len(members) == 0 {
		return nil, nil
	}

	groupMemberWithOrderList := make([]groupMemberWithOrder, 0, len(members))
	explicitOrders := sets.NewInt64()
	for _, member := range members {
		var order = defaultGroupOrder
		exists := false
		v := annotations.GetStringAnnotationMutil(util.IngressSuffixAlbConfigOrder, annotations.Order, member)
		if v != "" {
			exists = true
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to load Ingress group order for ingress: %v", util.NamespacedName(member))
			}
			order = i
		}
		if exists {
			if order < minGroupOrder || order > maxGroupOder {
				return nil, errors.Errorf("explicit Ingress group order must be within [%v:%v], Ingress: %v, order: %v",
					minGroupOrder, maxGroupOder, util.NamespacedName(member), order)
			}
			if explicitOrders.Has(order) {
				return nil, errors.Errorf("conflict Ingress group order: %v", order)
			}
			explicitOrders.Insert(order)
		}

		groupMemberWithOrderList = append(groupMemberWithOrderList, groupMemberWithOrder{member: member, order: order})
	}

	sort.Slice(groupMemberWithOrderList, func(i, j int) bool {
		orderI := groupMemberWithOrderList[i].order
		orderJ := groupMemberWithOrderList[j].order
		if orderI != orderJ {
			return orderI < orderJ
		}

		nameI := util.NamespacedName(groupMemberWithOrderList[i].member).String()
		nameJ := util.NamespacedName(groupMemberWithOrderList[j].member).String()
		return nameI < nameJ
	})

	sortedMembers := make([]*networking.Ingress, 0, len(groupMemberWithOrderList))
	for _, item := range groupMemberWithOrderList {
		sortedMembers = append(sortedMembers, item.member)
	}
	return sortedMembers, nil
}

func (m *defaultGroupLoader) IngressClass(ing *networking.Ingress) string {
	alb := ing.Spec.IngressClassName
	if alb == nil {
		albStr := ing.GetAnnotations()[store.IngressKey]
		alb = &albStr
	}
	ic := &networking.IngressClass{}
	err := m.kubeClientCache.Get(context.TODO(), types.NamespacedName{Name: *alb}, ic)
	if err != nil {
		klog.Errorf("Get IngressClass %s:%s, error: %s", "class", alb, err.Error())
		return ""
	}
	return ic.Spec.Parameters.Name
}
