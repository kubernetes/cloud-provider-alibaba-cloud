package backend

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/store"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/backend"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	pkgModel "k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("backend not found")

type EndpointResolver interface {
	ResolveENIEndpoints(ctx context.Context, svcKey types.NamespacedName, port intstr.IntOrString) ([]NodePortEndpoint, bool, error)

	ResolveLocalEndpoints(ctx context.Context, svcKey types.NamespacedName, port intstr.IntOrString) ([]NodePortEndpoint, bool, error)

	ResolveClusterEndpoints(ctx context.Context, svcKey types.NamespacedName, port intstr.IntOrString) ([]NodePortEndpoint, bool, error)
}

type PodEndpoint struct {
	IP       string
	Port     int
	NodeName *string
	Pod      *corev1.Pod
}

type NodePortEndpoint alb.BackendItem

func NewDefaultEndpointResolver(store store.Storer, k8sClient client.Client, cloud prvd.Provider, logger logr.Logger) *defaultEndpointResolver {
	return &defaultEndpointResolver{
		k8sClient: k8sClient,
		logger:    logger,
		cloud:     cloud,
		store:     store,
	}
}

var _ EndpointResolver = &defaultEndpointResolver{}

type defaultEndpointResolver struct {
	store     store.Storer
	k8sClient client.Client
	cloud     prvd.Provider
	logger    logr.Logger
}

func (r *defaultEndpointResolver) findServiceAndServicePort(ctx context.Context, svcKey types.NamespacedName, port intstr.IntOrString) (*corev1.Service, corev1.ServicePort, error) {
	svc := &corev1.Service{}
	if err := r.k8sClient.Get(ctx, svcKey, svc); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, corev1.ServicePort{}, fmt.Errorf("%w: %v", ErrNotFound, err.Error())
		}
		return nil, corev1.ServicePort{}, err
	}
	svcPort, err := LookupServicePort(svc, port)
	if err != nil {
		return nil, corev1.ServicePort{}, fmt.Errorf("%w: %v", ErrNotFound, err.Error())
	}

	return svc, svcPort, nil
}

func (r *defaultEndpointResolver) resolvePodEndpoints(ctx context.Context, svc *corev1.Service, svcPort corev1.ServicePort) ([]PodEndpoint, bool, error) {
	epsKey := util.NamespacedName(svc)
	eps := &corev1.Endpoints{}
	if err := r.k8sClient.Get(ctx, epsKey, eps); err != nil {
		klog.Errorf("resolvePodEndpoints: %v", err)
		if apierrors.IsNotFound(err) {
			return nil, false, fmt.Errorf("%w: %v", ErrNotFound, err.Error())
		}
		return nil, false, err
	}
	var endpoints []PodEndpoint
	containsPotentialReadyEndpoints := false
	klog.Infof("resolvePodEndpoints = %v", eps)

	for _, ep := range eps.Subsets {
		var backendPort int
		for _, p := range ep.Ports {
			if p.Name == svcPort.Name {
				backendPort = int(p.Port)
				break
			}
		}

		for _, addr := range ep.Addresses {
			if addr.TargetRef == nil || addr.TargetRef.Kind != "Pod" {
				continue
			}
			pod, err := r.findPodByReference(ctx, svc.Namespace, *addr.TargetRef)
			if err != nil {
				return nil, false, err
			}
			endpoints = append(endpoints, buildPodEndpoint(addr, backendPort, pod))
		}
		// readiness gates
		for _, epAddr := range ep.NotReadyAddresses {
			if epAddr.TargetRef == nil || epAddr.TargetRef.Kind != "Pod" {
				continue
			}
			pod, err := r.findPodByReference(ctx, svc.Namespace, *epAddr.TargetRef)
			if err != nil {
				klog.Errorf("findPodByReference error: %s", err.Error())
				return nil, false, err
			}

			if !helper.IsPodHasReadinessGate(pod) {
				continue
			}
			if !helper.IsPodContainersReady(pod) {
				containsPotentialReadyEndpoints = true
				continue
			}
			endpoints = append(endpoints, buildPodEndpoint(epAddr, backendPort, pod))
		}

	}

	return endpoints, containsPotentialReadyEndpoints, nil
}
func (r *defaultEndpointResolver) findPodByReference(ctx context.Context, namespace string, podRef corev1.ObjectReference) (*corev1.Pod, error) {
	podKey := fmt.Sprintf("%s/%s", podRef.Namespace, podRef.Name)
	return r.store.GetPod(podKey)
}
func (r *defaultEndpointResolver) ResolveENIEndpoints(ctx context.Context, svcKey types.NamespacedName, port intstr.IntOrString) ([]NodePortEndpoint, bool, error) {
	svc, svcPort, err := r.findServiceAndServicePort(ctx, svcKey, port)
	if err != nil {
		return nil, false, err
	}

	podEndpoints, containsPotentialReadyEndpoints, err := r.resolvePodEndpoints(ctx, svc, svcPort)
	if err != nil {
		return nil, containsPotentialReadyEndpoints, err
	}

	eps, err := r.transPodEndpointsToEnis(podEndpoints)
	if err != nil {
		return nil, containsPotentialReadyEndpoints, err
	}
	return eps, containsPotentialReadyEndpoints, nil
}

func (r *defaultEndpointResolver) ResolveLocalEndpoints(ctx context.Context, svcKey types.NamespacedName, port intstr.IntOrString) ([]NodePortEndpoint, bool, error) {
	svc, svcPort, err := r.findServiceAndServicePort(ctx, svcKey, port)
	if err != nil {
		return nil, false, err
	}

	podEndPoints, containsPotentialReadyEndpoints, err := r.resolvePodEndpoints(ctx, svc, svcPort)
	if err != nil {
		return nil, containsPotentialReadyEndpoints, err
	}

	svcNodePort := svcPort.NodePort

	reqCtx := &svcCtx.RequestContext{
		Ctx:     ctx,
		Service: svc,
		Anno:    &annotation.AnnotationRequest{Service: svc},
		Log:     r.logger,
	}
	nodes, err := backend.GetNodes(reqCtx, r.k8sClient)
	if err != nil {
		return nil, containsPotentialReadyEndpoints, err
	}
	nodesByName := nodesByName(nodes)

	ecsEndpoints := make([]NodePortEndpoint, 0)
	eciEndpoints := make([]PodEndpoint, 0)

	for _, podEndPoint := range podEndPoints {
		if podEndPoint.NodeName == nil {
			return nil, containsPotentialReadyEndpoints, errors.New("empty node name")
		}

		node, ok := nodesByName[*podEndPoint.NodeName]
		if !ok {
			continue
		}

		if node.Labels["type"] == util.LabelNodeTypeVK {
			eciEndpoints = append(eciEndpoints, podEndPoint)
			continue
		}

		if helper.HasExcludeLabel(&node) {
			continue
		}

		_, id, err := helper.NodeFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return nil, containsPotentialReadyEndpoints, err
		}

		ecsEndpoints = append(ecsEndpoints, buildNodePortEndpoint(id, "", int(svcNodePort), alb.ECSBackendType, util.DefaultServerWeight, podEndPoint.Pod))
	}

	if len(eciEndpoints) != 0 {
		eniEps, err := r.transPodEndpointsToEnis(eciEndpoints)
		if err != nil {
			return nil, containsPotentialReadyEndpoints, err
		}
		ecsEndpoints = append(ecsEndpoints, eniEps...)
	}

	return RemoteDuplicatedBackends(ecsEndpoints), containsPotentialReadyEndpoints, nil
}

func (r *defaultEndpointResolver) ResolveClusterEndpoints(ctx context.Context, svcKey types.NamespacedName, port intstr.IntOrString) ([]NodePortEndpoint, bool, error) {
	svc, svcPort, err := r.findServiceAndServicePort(ctx, svcKey, port)
	if err != nil {
		return nil, false, err
	}

	podEndPoints, containsPotentialReadyEndpoints, err := r.resolvePodEndpoints(ctx, svc, svcPort)
	if err != nil {
		return nil, containsPotentialReadyEndpoints, err
	}

	svcNodePort := svcPort.NodePort
	reqCtx := &svcCtx.RequestContext{
		Ctx:     ctx,
		Service: svc,
		Anno:    &annotation.AnnotationRequest{Service: svc},
		Log:     r.logger,
	}
	nodes, err := backend.GetNodes(reqCtx, r.k8sClient)
	if err != nil {
		return nil, containsPotentialReadyEndpoints, err
	}
	nodesByName := nodesByName(nodes)

	ecsEndpoints := make([]NodePortEndpoint, 0)
	for _, node := range nodes {
		if helper.HasExcludeLabel(&node) {
			continue
		}
		_, id, err := helper.NodeFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return nil, containsPotentialReadyEndpoints, fmt.Errorf("normal parse providerid: %s. "+
				"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
		}

		ecsEndpoints = append(ecsEndpoints, buildNodePortEndpoint(id, "", int(svcNodePort), alb.ECSBackendType, util.DefaultServerWeight, nil))
	}

	eciEndpoints := make([]PodEndpoint, 0)
	for _, podEndPoint := range podEndPoints {
		if podEndPoint.NodeName == nil {
			return nil, containsPotentialReadyEndpoints, errors.New("empty node name")
		}

		node, ok := nodesByName[*podEndPoint.NodeName]
		if !ok {
			continue
		}

		if node.Labels["type"] == util.LabelNodeTypeVK {
			eciEndpoints = append(eciEndpoints, podEndPoint)
		}
	}

	if len(eciEndpoints) != 0 {
		eniEndpointsFromEci, err := r.transPodEndpointsToEnis(eciEndpoints)
		if err != nil {
			return nil, containsPotentialReadyEndpoints, err
		}
		ecsEndpoints = append(ecsEndpoints, eniEndpointsFromEci...)
	}

	return ecsEndpoints, containsPotentialReadyEndpoints, nil
}

func nodesByName(nodes []corev1.Node) map[string]corev1.Node {
	nodesByName := make(map[string]corev1.Node)
	for _, node := range nodes {
		nodesByName[node.Name] = node
	}
	return nodesByName
}

func (r *defaultEndpointResolver) transPodEndpointsToEnis(backends []PodEndpoint) ([]NodePortEndpoint, error) {
	vpcId, err := r.cloud.VpcID()
	if err != nil {
		return nil, fmt.Errorf("get vpc id from metadata error:%s", err.Error())
	}

	var ips []string
	for _, b := range backends {
		ips = append(ips, b.IP)
	}

	result, err := r.cloud.DescribeNetworkInterfaces(vpcId, ips, pkgModel.IPv4)
	if err != nil {
		return nil, fmt.Errorf("call DescribeNetworkInterfaces: %s", err.Error())
	}
	var nodePortEndpoints []NodePortEndpoint
	for i := range backends {
		eniid, ok := result[backends[i].IP]
		if !ok {
			return nil, fmt.Errorf("can not find eniid for ip %s in vpc %s", backends[i].IP, vpcId)
		}
		// for ENI backend type, port should be set to targetPort (default value), no need to update
		nodePortEndpoints = append(nodePortEndpoints, buildNodePortEndpoint(eniid, backends[i].IP, backends[i].Port, alb.ENIBackendType, util.DefaultServerWeight, backends[i].Pod))
	}

	return nodePortEndpoints, nil

}

func buildPodEndpoint(epAddr corev1.EndpointAddress, port int, pod *corev1.Pod) PodEndpoint {
	return PodEndpoint{
		IP:       epAddr.IP,
		Port:     port,
		NodeName: epAddr.NodeName,
		Pod:      pod,
	}
}

func buildNodePortEndpoint(instanceID string, serverIP string, port int, tp string, weight int, pod *corev1.Pod) NodePortEndpoint {
	nodePortEndpoint := NodePortEndpoint{
		ServerId: instanceID,
		ServerIp: serverIP,
		Weight:   weight,
		Port:     port,
		Type:     tp,
		Pod:      pod,
	}
	return nodePortEndpoint
}

func LookupServicePort(svc *corev1.Service, port intstr.IntOrString) (corev1.ServicePort, error) {
	if port.Type == intstr.String {
		for _, p := range svc.Spec.Ports {
			if p.Name == port.StrVal {
				return p, nil
			}
		}
	} else {
		for _, p := range svc.Spec.Ports {
			if p.Port == port.IntVal {
				return p, nil
			}
		}
	}

	return corev1.ServicePort{}, errors.Errorf("unable to find port %s on service %s", port.String(), util.NamespacedName(svc))
}

func RemoteDuplicatedBackends(backends []NodePortEndpoint) []NodePortEndpoint {
	nodeMap := make(map[string]struct{})
	var uniqBackends []NodePortEndpoint
	for _, backend := range backends {
		if _, ok := nodeMap[backend.ServerId]; ok {
			continue
		}
		nodeMap[backend.ServerId] = struct{}{}
		uniqBackends = append(uniqBackends, backend)
	}
	return uniqBackends
}
