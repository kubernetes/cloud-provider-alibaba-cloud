package service

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/hash"
	"k8s.io/klog/v2"
	"os"
	"reflect"
	"strings"
	"time"
)

func findNodeByNodeName(nodes []v1.Node, nodeName string) *v1.Node {
	for _, n := range nodes {
		if n.Name == nodeName {
			return &n
		}
	}
	klog.Infof("node %s not found ", nodeName)
	return nil
}

// providerID
// 1) the id of the instance in the alicloud API. Use '.' to separate providerID which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7'. The format of "REGION.NODEID"
// 2) the id for an instance in the kubernetes API, which has 'alicloud://' prefix. e.g. alicloud://cn-hangzhou.i-v98dklsmnxkkgiiil7
func NodeFromProviderID(providerID string) (string, string, error) {
	if strings.HasPrefix(providerID, "alicloud://") {
		k8sName := strings.Split(providerID, "://")
		if len(k8sName) < 2 {
			return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
		} else {
			providerID = k8sName[1]
		}
	}

	name := strings.Split(providerID, ".")
	if len(name) < 2 {
		return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
	}
	return name[0], name[1], nil
}

// only for node event
func canNodeSkipEventHandler(node *v1.Node) bool {
	if node == nil || node.Labels == nil {
		return false
	}

	if helper.HasExcludeLabel(node) {
		klog.V(5).Infof("node %s has exclude label, skip", node.Name)
		return true
	}
	if isMasterNode(node) {
		klog.V(5).Infof("node %s is master node, skip", node.Name)
		return true
	}
	return false
}

func isMasterNode(node *v1.Node) bool {
	if _, isMaster := node.Labels[LabelNodeRoleMaster]; isMaster {
		return true
	}
	return false
}

func isLocalModeService(svc *v1.Service) bool {
	return svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeLocal
}

func isClusterIPService(svc *v1.Service) bool {
	return svc.Spec.Type == v1.ServiceTypeClusterIP
}

func isENIBackendType(svc *v1.Service) bool {
	if svc.Annotations[BackendType] != "" {
		return svc.Annotations[BackendType] == model.ENIBackendType
	}

	if os.Getenv("SERVICE_FORCE_BACKEND_ENI") != "" {
		return os.Getenv("SERVICE_FORCE_BACKEND_ENI") == "true"
	}

	return ctrlCfg.CloudCFG.Global.ServiceBackendType == model.ENIBackendType
}

func isNodeExcludeFromLoadBalancer(node *v1.Node) bool {
	if _, exclude := node.Labels[LabelNodeExcludeBalancer]; exclude {
		return true
	}

	if helper.HasExcludeLabel(node) {
		return true
	}
	return false
}

func needDeleteLoadBalancer(svc *v1.Service) bool {
	return svc.DeletionTimestamp != nil || svc.Spec.Type != v1.ServiceTypeLoadBalancer
}

func needLoadBalancer(service *v1.Service) bool {
	return service.Spec.Type == v1.ServiceTypeLoadBalancer
}

func getServiceHash(svc *v1.Service) string {
	var op []interface{}
	op = append(op, svc.Spec, svc.Annotations, svc.DeletionTimestamp)
	return hash.HashObject(op)
}

func isServiceHashChanged(service *v1.Service) bool {
	if oldHash, ok := service.Labels[LabelServiceHash]; ok {
		newHash := getServiceHash(service)
		return !strings.EqualFold(oldHash, newHash)
	}
	return true
}

func isLoadBalancerReusable(service *v1.Service, tags []model.Tag, lbIp string) (bool, string) {
	for _, tag := range tags {
		// the tag of the apiserver slb is "ack.aliyun.com": "${clusterid}",
		// so can not reuse slbs which have ack.aliyun.com tag key.
		if tag.TagKey == TAGKEY || tag.TagKey == util.ClusterTagKey {
			return false, "can not reuse loadbalancer created by kubernetes."
		}
	}

	if len(service.Status.LoadBalancer.Ingress) > 0 {
		found := false
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.IP == lbIp || (ingress.Hostname != "" && ingress.IP == "") {
				found = true
			}
		}
		if !found {
			return false, fmt.Sprintf("service has been associated with ip [%v], cannot be bound to ip [%s]",
				service.Status.LoadBalancer.Ingress[0].IP, lbIp)
		}
	}

	return true, ""
}

// check if the service exists in service definition
func isServiceOwnIngress(service *v1.Service) bool {
	if service == nil {
		return false
	}
	if len(service.Status.LoadBalancer.Ingress) == 0 {
		return false
	}
	return true
}

// MAX_BACKEND_NUM max batch backend num
const (
	MaxBackendNum = 39
	MaxLBTagNum   = 10
)

type Func func([]interface{}) error

// Batch batch process `object` m with func `func`
// for general purpose
func Batch(m interface{}, cnt int, batch Func) error {
	if cnt <= 0 {
		cnt = MaxBackendNum
	}
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("non-slice type for %v", m)
	}

	// need to convert interface to []interface
	// see https://github.com/golang/go/wiki/InterfaceSlice
	target := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		target[i] = v.Index(i).Interface()
	}
	klog.Infof("batch process ,total length %d", len(target))
	for len(target) > cnt {
		if err := batch(target[0:cnt]); err != nil {

			return err
		}
		target = target[cnt:]
	}
	if len(target) <= 0 {
		return nil
	}

	klog.Infof("batch process ,total length %d last section", len(target))
	return batch(target)
}

const TRY_AGAIN = "try again"

func retry(
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

func Is7LayerProtocol(protocol string) bool {
	return protocol == model.HTTP || protocol == model.HTTPS
}

func Is4LayerProtocol(protocol string) bool {
	return protocol == model.TCP || protocol == model.UDP
}

func LogEndpoints(eps *v1.Endpoints) string {
	if eps == nil {
		return "endpoints is nil"
	}
	var epAddrList []string
	for _, subSet := range eps.Subsets {
		for _, addr := range subSet.Addresses {
			epAddrList = append(epAddrList, addr.IP)
		}
	}
	return strings.Join(epAddrList, ",")
}

func LogEndpointSlice(es *discovery.EndpointSlice) string {
	if es == nil {
		return "endpointSlice is nil"
	}
	var epAddrList []string
	for _, ep := range es.Endpoints {
		epAddrList = append(epAddrList, ep.Addresses...)
	}

	return strings.Join(epAddrList, ",")
}

func LogEndpointSliceList(esList []discovery.EndpointSlice) string {
	if esList == nil {
		return "endpointSliceList is nil"
	}
	var epAddrList []string
	for _, es := range esList {
		for _, ep := range es.Endpoints {
			if ep.Conditions.Ready != nil && !*ep.Conditions.Ready {
				continue
			}
			epAddrList = append(epAddrList, ep.Addresses...)
		}
	}

	return strings.Join(epAddrList, ",")
}

func GetServiceTrafficPolicy(svc *v1.Service) (TrafficPolicy, error) {
	if isENIBackendType(svc) {
		return ENITrafficPolicy, nil
	}
	if isClusterIPService(svc) {
		return "", fmt.Errorf("cluster service type just support eni mode for alb ingress")
	}
	if isLocalModeService(svc) {
		return LocalTrafficPolicy, nil
	}
	return ClusterTrafficPolicy, nil
}
