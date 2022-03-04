package albconfigmanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
)

func (t *defaultModelBuildTask) buildServerGroup(ctx context.Context,
	ing *networking.Ingress, svc *corev1.Service, port int) (*alb.ServerGroup, error) {
	sgpResID := t.buildServerGroupResourceID(util.NamespacedName(ing), util.NamespacedName(svc), port)
	if sgp, exists := t.sgpByResID[sgpResID]; exists {
		return sgp, nil
	}
	sgpSpec, err := t.buildServerGroupSpec(ctx, ing, svc, port)
	if err != nil {
		return nil, err
	}
	sgp := alb.NewServerGroup(t.stack, sgpResID, sgpSpec)
	t.sgpByResID[sgpResID] = sgp
	return sgp, nil
}

func calServerGroupResourceIDHashUUID(resourceID string) string {
	uuidHash := sha256.New()
	_, _ = uuidHash.Write([]byte(resourceID))
	uuid := hex.EncodeToString(uuidHash.Sum(nil))
	return uuid
}

func (t *defaultModelBuildTask) buildServerGroupResourceID(ingKey types.NamespacedName, svcKey types.NamespacedName, port int) string {
	resourceID := fmt.Sprintf("%s/%s-%s:%s", ingKey.Namespace, ingKey.Name, svcKey.Name, fmt.Sprintf("%v", port))
	return calServerGroupResourceIDHashUUID(resourceID)
}

func (t *defaultModelBuildTask) buildServerGroupName(ing *networking.Ingress, svc *corev1.Service, port int) string {
	return fmt.Sprintf("%s-%s-%s", svc.Namespace, svc.Name, fmt.Sprintf("%v", port))
}

func (t *defaultModelBuildTask) buildServerGroupSpec(_ context.Context,
	ing *networking.Ingress, svc *corev1.Service, port int) (alb.ServerGroupSpec, error) {

	// preCheck tag value
	if len(ing.Namespace) == 0 ||
		len(ing.Name) == 0 ||
		len(svc.Name) == 0 {
		return alb.ServerGroupSpec{}, fmt.Errorf("ingress namespace, ingress name and service name cant be empty")
	}
	tags := make([]alb.ALBTag, 0)
	tags = append(tags, []alb.ALBTag{
		{
			Key:   util.ServiceNamespaceTagKey,
			Value: ing.Namespace,
		},
		{
			Key:   util.IngressNameTagKey,
			Value: ing.Name,
		},
		{
			Key:   util.ServiceNameTagKey,
			Value: svc.Name,
		},
		{
			Key:   util.ServicePortTagKey,
			Value: fmt.Sprintf("%v", port),
		},
	}...)

	sgpNameKey := alb.ServerGroupNamedKey{
		ClusterID:   t.clusterID,
		IngressName: ing.GetName(),
		ServiceName: svc.GetName(),
		Namespace:   svc.GetNamespace(),
		ServicePort: port,
	}

	var sgpSpec alb.ServerGroupSpec
	sgpSpec.ServerGroupNamedKey = sgpNameKey
	sgpSpec.Tags = tags
	sgpSpec.HealthCheckConfig = buildServerGroupHealthCheckConfig(ing)
	sgpSpec.ServerGroupName = t.buildServerGroupName(ing, svc, port)
	sgpSpec.Scheduler = t.defaultServerGroupScheduler
	sgpSpec.Protocol = t.defaultServerGroupProtocol
	sgpSpec.StickySessionConfig = buildServerGroupStickySessionConfig(ing)
	sgpSpec.ServerGroupType = t.defaultServerGroupType
	sgpSpec.VpcId = t.vpcID
	return sgpSpec, nil
}

func buildServerGroupHealthCheckConfig(ing *networking.Ingress) alb.HealthCheckConfig {
	healthCheckEnabled := util.DefaultServerGroupHealthCheckEnabled
	if v, ok := ing.Annotations[annotations.HealthCheckEnabled]; ok && v == "true" {
		healthCheckEnabled = true
	}

	healthcheckPath := util.DefaultServerGroupHealthCheckPath
	if v, ok := ing.Annotations[annotations.HealthCheckPath]; ok {
		healthcheckPath = v
	}
	healthcheckMethod := util.DefaultServerGroupHealthCheckMethod
	if v, ok := ing.Annotations[annotations.HealthCheckMethod]; ok {
		healthcheckMethod = v
	}
	healthcheckProtocol := util.DefaultServerGroupHealthCheckProtocol
	if v, ok := ing.Annotations[annotations.HealthCheckProtocol]; ok {
		healthcheckProtocol = v
	}
	healthcheckCode := util.DefaultServerGroupHealthCheckHTTPCodes
	if v, ok := ing.Annotations[annotations.HealthCheckHTTPCode]; ok {
		healthcheckCode = v
	}
	healthcheckTimeout := util.DefaultServerGroupHealthCheckTimeout
	if v, ok := ing.Annotations[annotations.HealthCheckTimeout]; ok {
		if val, err := strconv.Atoi(v); err != nil {
			klog.Error(err.Error())
		} else {
			healthcheckTimeout = val
		}
	}
	healthCheckInterval := util.DefaultServerGroupHealthCheckInterval
	if v, ok := ing.Annotations[annotations.HealthCheckInterval]; ok {
		if val, err := strconv.Atoi(v); err != nil {
			klog.Error(err.Error())
		} else {
			healthCheckInterval = val
		}
	}
	healthyThreshold := util.DefaultServerGroupHealthyThreshold
	if v, ok := ing.Annotations[annotations.HealthThreshold]; ok {
		if val, err := strconv.Atoi(v); err != nil {
			klog.Error(err.Error())
		} else {
			healthyThreshold = val
		}
	}
	unhealthyThreshold := util.DefaultServerGroupUnhealthyThreshold
	if v, ok := ing.Annotations[annotations.UnHealthThreshold]; ok {
		if val, err := strconv.Atoi(v); err != nil {
			klog.Error(err.Error())
		} else {
			unhealthyThreshold = val
		}
	}
	return alb.HealthCheckConfig{
		HealthCheckConnectPort:         util.DefaultServerGroupHealthCheckConnectPort,
		HealthCheckEnabled:             healthCheckEnabled,
		HealthCheckHost:                util.DefaultServerGroupHealthCheckHost,
		HealthCheckHttpVersion:         util.DefaultServerGroupHealthCheckHttpVersion,
		HealthCheckInterval:            healthCheckInterval,
		HealthCheckMethod:              healthcheckMethod,
		HealthCheckPath:                healthcheckPath,
		HealthCheckProtocol:            healthcheckProtocol,
		HealthCheckTimeout:             healthcheckTimeout,
		HealthyThreshold:               healthyThreshold,
		UnhealthyThreshold:             unhealthyThreshold,
		HealthCheckTcpFastCloseEnabled: util.DefaultServerGroupHealthCheckTcpFastCloseEnabled,
		HealthCheckHttpCodes: []string{
			healthcheckCode,
		},
		HealthCheckCodes: []string{
			util.DefaultServerGroupHealthCheckCodes,
		},
	}
}

func buildServerGroupStickySessionConfig(ing *networking.Ingress) alb.StickySessionConfig {
	return alb.StickySessionConfig{
		Cookie:               "",
		CookieTimeout:        util.DefaultServerGroupStickySessionCookieTimeout,
		StickySessionEnabled: util.DefaultServerGroupStickySessionEnabled,
		StickySessionType:    util.DefaultServerGroupStickySessionType,
	}
}
