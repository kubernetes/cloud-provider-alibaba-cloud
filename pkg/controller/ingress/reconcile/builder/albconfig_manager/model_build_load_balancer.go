package albconfigmanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"

	"github.com/pkg/errors"
)

const (
	ApplicationLoadBalancerResource = "ApplicationLoadBalancer"
)

func (t *defaultModelBuildTask) buildAlbLoadBalancer(ctx context.Context, albconfig *v1.AlbConfig) (*alb.AlbLoadBalancer, error) {
	lbSpec, err := t.buildAlbLoadBalancerSpec(ctx, albconfig)
	if err != nil {
		return nil, err
	}
	lb := alb.NewAlbLoadBalancer(t.stack, ApplicationLoadBalancerResource, lbSpec)
	t.loadBalancer = lb
	return lb, nil
}

func (t *defaultModelBuildTask) buildAlbLoadBalancerSpec(ctx context.Context, albConfig *v1.AlbConfig) (alb.ALBLoadBalancerSpec, error) {
	lbModel := alb.ALBLoadBalancerSpec{}
	lbModel.LoadBalancerId = albConfig.Spec.LoadBalancer.Id
	lbModel.ForceOverride = albConfig.Spec.LoadBalancer.ForceOverride
	forceOverride := false
	if lbModel.ForceOverride == nil {
		lbModel.ForceOverride = &forceOverride
	}
	lbModel.ListenerForceOverride = albConfig.Spec.LoadBalancer.ListenerForceOverride
	listenerForceOverride := false
	if lbModel.ListenerForceOverride == nil {
		lbModel.ListenerForceOverride = &listenerForceOverride
	}
	if len(albConfig.Spec.LoadBalancer.Name) != 0 {
		lbModel.LoadBalancerName = albConfig.Spec.LoadBalancer.Name
	} else {
		lbName, err := t.buildAlbLoadBalancerName()
		if err != nil {
			return alb.ALBLoadBalancerSpec{}, nil
		}
		lbModel.LoadBalancerName = lbName
	}
	lbModel.VpcId = t.vpcID
	lbModel.AccessLogConfig = alb.AccessLogConfig{
		LogStore:   albConfig.Spec.LoadBalancer.AccessLogConfig.LogStore,
		LogProject: albConfig.Spec.LoadBalancer.AccessLogConfig.LogProject,
	}

	zoneMappings := make([]alb.ZoneMapping, 0)
	if albConfig.Spec.LoadBalancer.Id == "" {
		if len(albConfig.Spec.LoadBalancer.ZoneMappings) != 0 {
			vSwitchIds := make([]string, 0)
			for _, zm := range albConfig.Spec.LoadBalancer.ZoneMappings {
				vSwitchIds = append(vSwitchIds, zm.VSwitchId)
			}
			vSwitches, err := t.vSwitchResolver.ResolveViaIDSlice(ctx, vSwitchIds)
			if err != nil {
				return alb.ALBLoadBalancerSpec{}, err
			}
			if len(vSwitches) < 2 {
				return alb.ALBLoadBalancerSpec{}, errors.New("unable to discover at least two vswitchs for alb")
			}
			for _, vSwitch := range vSwitches {
				zoneMappings = append(zoneMappings, alb.ZoneMapping{
					VSwitchId: vSwitch.VSwitchId,
					ZoneId:    vSwitch.ZoneId,
				})
			}
		} else {
			vSwitches, err := t.vSwitchResolver.ResolveViaDiscovery(ctx)
			if err != nil {
				return alb.ALBLoadBalancerSpec{}, err
			}
			if len(vSwitches) < 2 {
				return alb.ALBLoadBalancerSpec{}, errors.New("unable to discover at least two albZones for alb")
			}
			for _, vSwitch := range vSwitches {
				zoneMappings = append(zoneMappings, alb.ZoneMapping{
					VSwitchId: vSwitch.VSwitchId,
					ZoneId:    vSwitch.ZoneId,
				})
			}
		}
	}
	lbModel.ZoneMapping = zoneMappings

	lbModel.AddressAllocatedMode = albConfig.Spec.LoadBalancer.AddressAllocatedMode
	if lbModel.AddressAllocatedMode == "" {
		lbModel.AddressAllocatedMode = util.LoadBalancerAddressAllocatedModeDynamic
	}
	lbModel.AddressType = albConfig.Spec.LoadBalancer.AddressType
	if lbModel.AddressType == "" {
		lbModel.AddressType = util.LoadBalancerAddressTypeInternet
	}
	lbModel.DeletionProtectionConfig = alb.DeletionProtectionConfig{
		Enabled:     true,
		EnabledTime: "",
	}
	lbModel.ModificationProtectionConfig = alb.ModificationProtectionConfig{
		Reason: "",
		Status: util.LoadBalancerModificationProtectionStatusConsoleProtection,
	}
	payType := albConfig.Spec.LoadBalancer.BillingConfig.PayType
	if payType == "" {
		payType = util.LoadBalancerPayTypePostPay
	}
	lbModel.LoadBalancerBillingConfig = alb.LoadBalancerBillingConfig{
		InternetBandwidth:  albConfig.Spec.LoadBalancer.BillingConfig.InternetBandwidth,
		InternetChargeType: albConfig.Spec.LoadBalancer.BillingConfig.InternetChargeType,
		PayType:            payType,
	}
	lbModel.LoadBalancerEdition = albConfig.Spec.LoadBalancer.Edition
	if lbModel.LoadBalancerEdition == "" {
		lbModel.LoadBalancerEdition = util.LoadBalancerEditionStandard
	}

	// build customer tags
	if len(albConfig.Spec.LoadBalancer.Tags) != 0 {
		tags := make([]alb.ALBTag, 0)
		for _, tag := range albConfig.Spec.LoadBalancer.Tags {
			tags = append(tags, alb.ALBTag{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}
		lbModel.Tags = tags
	}

	return lbModel, nil
}

var invalidLoadBalancerNamePattern = regexp.MustCompile("[[:^alnum:]]")

func (t *defaultModelBuildTask) buildAlbLoadBalancerName() (string, error) {
	uuidHash := sha256.New()
	_, _ = uuidHash.Write([]byte(t.clusterID))
	_, _ = uuidHash.Write([]byte(t.ingGroup.ID.String()))
	uuid := hex.EncodeToString(uuidHash.Sum(nil))

	sanitizedNamespace := invalidLoadBalancerNamePattern.ReplaceAllString(t.ingGroup.ID.Namespace, "")
	sanitizedName := invalidLoadBalancerNamePattern.ReplaceAllString(t.ingGroup.ID.Name, "")
	return fmt.Sprintf("k8s-%s-%s-%.10s", sanitizedNamespace, sanitizedName, uuid), nil
}
