package helper

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/util/retry"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/hash"

	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
)

// Finalizer
const (
	ServiceFinalizer = "service.k8s.alibaba/resources"
	NLBFinalizer     = "service.k8s.alibaba/nlb"
)

// annotation
const (
	BackendType       = "service.beta.kubernetes.io/backend-type"
	LoadBalancerClass = "service.beta.kubernetes.io/class"
	LoadBalancerType  = "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-type"
	TunnelType        = "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-tunnel-type"
)

// load balancer class
const NLBClass = "alibabacloud.com/nlb"

// label
const (
	LabelServiceHash    = "service.beta.kubernetes.io/hash"
	LabelLoadBalancerId = "service.k8s.alibaba/loadbalancer-id"
)

const (
	TAGKEY   = "kubernetes.do.not.delete"
	REUSEKEY = "kubernetes.reused.by.user"
)

type TrafficPolicy string

const (
	// LocalTrafficPolicy externalTrafficPolicy=Local
	LocalTrafficPolicy = TrafficPolicy("Local")
	// ClusterTrafficPolicy externalTrafficPolicy=Cluster
	ClusterTrafficPolicy = TrafficPolicy("Cluster")
	// ENITrafficPolicy is forwarded to pod directly
	ENITrafficPolicy = TrafficPolicy("ENI")
)

func GetServiceTrafficPolicy(svc *v1.Service) (TrafficPolicy, error) {
	if IsENIBackendType(svc) {
		return ENITrafficPolicy, nil
	}
	if IsClusterIPService(svc) {
		return "", fmt.Errorf("cluster service type just support eni mode for alb ingress")
	}
	if IsLocalModeService(svc) {
		return LocalTrafficPolicy, nil
	}
	return ClusterTrafficPolicy, nil
}

func IsLocalModeService(svc *v1.Service) bool {
	return svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeLocal
}

func IsENIBackendType(svc *v1.Service) bool {
	if svc.Annotations[BackendType] != "" {
		return svc.Annotations[BackendType] == model.ENIBackendType
	}

	if os.Getenv("SERVICE_FORCE_BACKEND_ENI") != "" {
		return os.Getenv("SERVICE_FORCE_BACKEND_ENI") == "true"
	}

	return ctrlCfg.CloudCFG.Global.ServiceBackendType == model.ENIBackendType
}

func IsClusterIPService(svc *v1.Service) bool {
	return svc.Spec.Type == v1.ServiceTypeClusterIP
}

func NeedDeleteLoadBalancer(svc *v1.Service) bool {
	return svc.DeletionTimestamp != nil || svc.Spec.Type != v1.ServiceTypeLoadBalancer
}

func NeedCLB(service *v1.Service) bool {
	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		return false
	}
	if service.Spec.LoadBalancerClass != nil {
		return false
	}
	if feature.DefaultFeatureGate.Enabled(ctrlCfg.LoadBalancerTypeAnnotation) && service.Annotations[LoadBalancerType] != "" {
		return false
	}
	if IsTunnelTypeService(service) {
		return false
	}
	return service.Annotations[LoadBalancerClass] == ""
}

func NeedNLB(service *v1.Service) bool {
	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		return false
	}
	if service.Spec.LoadBalancerClass != nil && *service.Spec.LoadBalancerClass == NLBClass {
		return true
	}
	if feature.DefaultFeatureGate.Enabled(ctrlCfg.LoadBalancerTypeAnnotation) && service.Annotations[LoadBalancerType] == "nlb" {
		return service.Annotations[LoadBalancerType] == "nlb"
	}
	return false
}

func GetServiceHash(svc *v1.Service) string {
	var op []interface{}
	// ServiceSpec
	op = append(op, svc.Spec.Ports, svc.Spec.Type, svc.Spec.ExternalTrafficPolicy, svc.Spec.LoadBalancerClass)
	op = append(op, svc.Annotations, svc.DeletionTimestamp)
	return hash.HashObject(op)
}

func IsServiceHashChanged(service *v1.Service) bool {
	if oldHash, ok := service.Labels[LabelServiceHash]; ok {
		newHash := GetServiceHash(service)
		return !strings.EqualFold(oldHash, newHash)
	}
	return true
}

const TRY_AGAIN = "try again"

func Retry(
	backoff *wait.Backoff,
	fun func(svc *v1.Service) error,
	svc *v1.Service,
) error {
	if backoff == nil {
		backoff = &wait.Backoff{
			Duration: 1 * time.Second,
			Steps:    8,
			Factor:   2,
			Jitter:   4,
		}
	}
	return wait.ExponentialBackoff(
		*backoff,
		func() (bool, error) {
			err := fun(svc)
			if err != nil &&
				strings.Contains(err.Error(), TRY_AGAIN) {
				klog.Errorf("retry with error: %s", err.Error())
				return false, nil
			}
			if err != nil {
				klog.Errorf("retry error: NotRetry, %s", err.Error())
			}
			return true, nil
		},
	)
}

func RetryOnErrorContains(contains string, f func() error) error {
	return retry.OnError(retry.DefaultRetry, func(err error) bool {
		return strings.Contains(err.Error(), contains)
	}, f)
}

func Is7LayerProtocol(protocol string) bool {
	return protocol == model.HTTP || protocol == model.HTTPS
}

func Is4LayerProtocol(protocol string) bool {
	return protocol == model.TCP || protocol == model.UDP
}

// check if the service exists in service definition
func IsServiceOwnIngress(service *v1.Service) bool {
	if service == nil {
		return false
	}
	if len(service.Status.LoadBalancer.Ingress) == 0 {
		return false
	}
	return true
}

func IsTunnelTypeService(svc *v1.Service) bool {
	if svc.Annotations == nil {
		return false
	}
	if _, ok := svc.Annotations[TunnelType]; ok {
		return true
	}
	return false
}
