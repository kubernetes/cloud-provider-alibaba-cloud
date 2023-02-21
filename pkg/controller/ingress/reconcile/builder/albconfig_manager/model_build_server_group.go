package albconfigmanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	if err := checkIngressProtocolAnnotations(ing); err != nil {
		return alb.ServerGroupSpec{}, err
	}
	if err := checkBackendSchedulerAnnotations(ing); err != nil {
		return alb.ServerGroupSpec{}, err
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
	sgpSpec.Scheduler = t.buildServerGroupScheduler(ing)
	sgpSpec.Protocol = t.buildServerGroupProtocol(ing)
	sgpSpec.StickySessionConfig = buildServerGroupStickySessionConfig(ing)
	sgpSpec.ServerGroupType = t.defaultServerGroupType
	sgpSpec.VpcId = t.vpcID
	return sgpSpec, nil
}

func checkBackendSchedulerAnnotations(ing *networking.Ingress) error {
	if v, ok := ing.Annotations[annotations.AlbBackendScheduler]; ok {
		switch v {
		case "wrr":
			return nil
		case "wlc":
			return nil
		case "sch":
			return nil
		default:
			return fmt.Errorf("unkown backend scheduler [%s]", v)
		}
	}
	return nil
}

func checkIngressProtocolAnnotations(ing *networking.Ingress) error {
	if v, ok := ing.Annotations[annotations.AlbBackendProtocol]; ok && v == "grpc" {
		if len(ing.Spec.TLS) == 0 {
			return fmt.Errorf("'grpc' backend protocol must be use with TLS listener(Ingress.Spec.TLS)")
		}
		if rawListenPorts, err := annotations.GetStringAnnotation(annotations.ListenPorts, ing); err == nil {
			var entries []map[string]int32
			if err := json.Unmarshal([]byte(rawListenPorts), &entries); err != nil {
				return fmt.Errorf("failed to parse listen-ports configuration: `%s` [%v]", rawListenPorts, err)
			}
			if len(entries) == 0 {
				return fmt.Errorf("empty listen-ports configuration: `%s`", rawListenPorts)
			}

			for _, entry := range entries {
				for protocol, port := range entry {
					if port < 1 || port > 65535 {
						return fmt.Errorf("listen port must be within [1, 65535]: %v", port)
					}
					switch protocol {
					case string(ProtocolHTTP):
						if v := annotations.GetStringAnnotationMutil(annotations.NginxSslRedirect, annotations.AlbSslRedirect, ing); v == "true" && port == 80 {
							break
						}
						return fmt.Errorf("'grpc' backend protocol must be use TLS listener only or redirect to")
					case string(ProtocolHTTPS):
						continue
					default:
						return fmt.Errorf("listen protocol must be within [%v, %v]: %v", ProtocolHTTP, ProtocolHTTPS, protocol)
					}
				}
			}
		}
	}
	return nil
}

func (t *defaultModelBuildTask) buildServerGroupScheduler(ing *networking.Ingress) string {
	backendScheduler := t.defaultServerGroupScheduler
	if v, ok := ing.Annotations[annotations.AlbBackendScheduler]; ok {
		switch v {
		case "wrr":
			backendScheduler = util.ServerGroupSchedulerWrr
		case "wlc":
			backendScheduler = util.ServerGroupSchedulerWlc
		case "sch":
			backendScheduler = util.ServerGroupSchedulerSch
		default:
			backendScheduler = util.ServerGroupSchedulerWrr
		}
	}
	return backendScheduler
}

func (t *defaultModelBuildTask) buildServerGroupProtocol(ing *networking.Ingress) string {
	backendProtocol := t.defaultServerGroupProtocol
	if v, ok := ing.Annotations[annotations.AlbBackendProtocol]; ok {
		switch v {
		case "grpc":
			backendProtocol = util.ServerGroupProtocolGRPC
		case "https":
			backendProtocol = util.ServerGroupProtocolHTTPS
		default:
			backendProtocol = util.ServerGroupProtocolHTTP
		}
	}
	return backendProtocol
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
	healthyCheckConnectPort := util.DefaultServerGroupHealthCheckConnectPort
	if v, ok := ing.Annotations[annotations.HealthCheckConnectPort]; ok {
		if val, err := strconv.Atoi(v); err != nil {
			klog.Error(err.Error())
		} else {
			healthyCheckConnectPort = val
		}
	}
	return alb.HealthCheckConfig{
		HealthCheckConnectPort:         healthyCheckConnectPort,
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
	sessionStickEnabled := util.DefaultServerGroupStickySessionEnabled
	if v, ok := ing.Annotations[annotations.SessionStick]; ok && v == "true" {
		sessionStickEnabled = true
	}
	sessionStickType := util.DefaultServerGroupStickySessionType
	if v, ok := ing.Annotations[annotations.SessionStickType]; ok {
		sessionStickType = v
	}
	cookieTimeout := util.DefaultServerGroupStickySessionCookieTimeout
	if v, ok := ing.Annotations[annotations.CookieTimeout]; ok {
		if val, err := strconv.Atoi(v); err != nil {
			klog.Error(err.Error())
		} else {
			cookieTimeout = val
		}
	}
	return alb.StickySessionConfig{
		Cookie:               "",
		CookieTimeout:        cookieTimeout,
		StickySessionEnabled: sessionStickEnabled,
		StickySessionType:    sessionStickType,
	}
}
