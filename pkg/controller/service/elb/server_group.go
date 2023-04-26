package elb

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/backend"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"

	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

func NewServerGroupManager(kubeClient client.Client, cloud prvd.Provider) *ServerGroupManager {
	manager := &ServerGroupManager{
		kubeClient: kubeClient,
		cloud:      cloud,
	}
	return manager
}

type ServerGroupManager struct {
	kubeClient client.Client
	cloud      prvd.Provider
}

func (mgr *ServerGroupManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	candidates, err := backend.NewEdgeEndpoints(reqCtx, mgr.kubeClient)
	if err != nil {
		return fmt.Errorf("build edge endpoints error: %s", err.Error())
	}
	candidates.Nodes, err = mgr.getEdgeNetworkENSNode(reqCtx, candidates, mdl)
	if err != nil {
		return fmt.Errorf("get ens nodes error: %s", err.Error())
	}

	mdl.ServerGroup, err = mgr.buildLocalServerGroup(reqCtx, candidates)
	if err != nil {
		return fmt.Errorf("build local server group error: %s", err.Error())
	}

	return nil
}

func (mgr *ServerGroupManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	return mgr.cloud.FindBackendFromLoadBalancer(reqCtx.Ctx, mdl.GetLoadBalancerId(), &mdl.ServerGroup)
}

func (mgr *ServerGroupManager) getEdgeNetworkENSNode(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI, mdl *elbmodel.EdgeLoadBalancer) ([]v1.Node, error) {
	ret := make([]v1.Node, 0)
	ensMap, err := mgr.cloud.FindEnsInstancesByNetwork(reqCtx.Ctx, mdl)
	if err != nil {
		return ret, fmt.Errorf("find ens instances by network %s error: %s", mdl.GetNetworkId(), err.Error())
	}
	for _, node := range candidates.Nodes {
		status, ok := ensMap[node.Name]
		if ok && status == ENSRunning {
			ret = append(ret, node)
		}
	}
	return ret, nil
}

func (mgr *ServerGroupManager) buildLocalServerGroup(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI) (elbmodel.EdgeServerGroup, error) {
	var err error
	ret := new(elbmodel.EdgeServerGroup)

	switch candidates.TrafficPolicy {
	case helper.LocalTrafficPolicy:
		reqCtx.Log.Info("local mode, build backends for loadbalancer")
		ret.Backends, err = mgr.buildLocalBackends(reqCtx, candidates)
		if err != nil {
			return *ret, fmt.Errorf("build local backends error: %s", err.Error())
		}
	case helper.ClusterTrafficPolicy:
		reqCtx.Log.Info("cluster mode, build backends for loadbalancer")
		ret.Backends, err = mgr.buildClusterBackends(reqCtx, candidates)
		if err != nil {
			return *ret, fmt.Errorf("build cluster backends error: %s", err.Error())
		}
	}
	return *ret, nil
}

func (mgr *ServerGroupManager) buildLocalBackends(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI) ([]elbmodel.EdgeBackendAttribute, error) {
	ret := make([]elbmodel.EdgeBackendAttribute, 0)
	if reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse) != "" && isTrue(reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse)) {
		return ret, fmt.Errorf("service %s reuse elb not support external traffic policy using local mode", util.Key(reqCtx.Service))
	}
	var initBackends []elbmodel.EdgeBackendAttribute
	if utilfeature.DefaultMutableFeatureGate.Enabled(ctrlCfg.EndpointSlice) {
		initBackends = setBackendFromEndpointSlice(reqCtx, candidates)
	} else {
		initBackends = setBackendFromEndpoints(reqCtx, candidates)
	}
	candidateBackends := getCandidateBackend(candidates)
	for _, backend := range initBackends {
		if _, ok := candidateBackends[backend.ServerId]; ok {
			ret = append(ret, backend)
		}
	}
	return ret, nil
}

func (mgr *ServerGroupManager) buildClusterBackends(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI) ([]elbmodel.EdgeBackendAttribute, error) {
	ret := make([]elbmodel.EdgeBackendAttribute, 0)
	weightMap := splitBackendWeight(reqCtx)
	for _, node := range candidates.Nodes {
		if node.Name == "" {
			klog.Warningf("[%s], Invalid node without instance name", util.Key(reqCtx.Service))
			continue
		}
		ret = append(ret, elbmodel.EdgeBackendAttribute{
			ServerId: node.Name,
			Type:     elbmodel.ServerGroupDefaultType,
			Weight:   getBackendWeight(node.Name, weightMap),
			Port:     elbmodel.ServerGroupDefaultPort,
		})
	}
	return ret, nil
}

func (mgr *ServerGroupManager) batchAddServerGroup(reqCtx *svcCtx.RequestContext, lbId string, sg *elbmodel.EdgeServerGroup) error {
	iter := int(math.Ceil(float64(len(sg.Backends)) / float64(ENSBatchAddMaxNumber)))
	for i := 0; i < iter; i++ {
		subSg := new(elbmodel.EdgeServerGroup)
		if i == iter-1 {
			subSg.Backends = sg.Backends[i*ENSBatchAddMaxNumber:]
		} else {
			subSg.Backends = sg.Backends[i*ENSBatchAddMaxNumber : (i+1)*ENSBatchAddMaxNumber]
		}
		err := mgr.cloud.AddBackendToEdgeServerGroup(reqCtx.Ctx, lbId, subSg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mgr *ServerGroupManager) batchUpdateServerGroup(reqCtx *svcCtx.RequestContext, lbId string, sg *elbmodel.EdgeServerGroup) error {

	iter := int(math.Ceil(float64(len(sg.Backends)) / float64(ENSBatchAddMaxNumber)))
	for i := 0; i < iter; i++ {
		subSg := new(elbmodel.EdgeServerGroup)
		if i == iter-1 {
			subSg.Backends = sg.Backends[i*ENSBatchAddMaxNumber:]
		} else {
			subSg.Backends = sg.Backends[i*ENSBatchAddMaxNumber : (i+1)*ENSBatchAddMaxNumber]
		}
		err := mgr.cloud.UpdateEdgeServerGroup(reqCtx.Ctx, lbId, subSg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mgr *ServerGroupManager) batchRemoveServerGroup(reqCtx *svcCtx.RequestContext, lbId string, sg *elbmodel.EdgeServerGroup) error {
	iter := int(math.Ceil(float64(len(sg.Backends)) / float64(ENSBatchAddMaxNumber)))
	for i := 0; i < iter; i++ {
		subSg := new(elbmodel.EdgeServerGroup)
		if i == iter-1 {
			subSg.Backends = sg.Backends[i*ENSBatchAddMaxNumber:]
		} else {
			subSg.Backends = sg.Backends[i*ENSBatchAddMaxNumber : (i+1)*ENSBatchAddMaxNumber]
		}
		err := mgr.cloud.RemoveBackendFromEdgeServerGroup(reqCtx.Ctx, lbId, subSg)
		if err != nil {
			return err
		}
	}
	return nil
}

func setBackendFromEndpointSlice(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI) []elbmodel.EdgeBackendAttribute {
	ret := make([]elbmodel.EdgeBackendAttribute, 0)
	endpointMap := make(map[string]struct{})
	weightMap := splitBackendWeight(reqCtx)
	for _, es := range candidates.EndpointSlices {
		for _, ep := range es.Endpoints {
			if ep.Conditions.Ready == nil || !*ep.Conditions.Ready {
				continue
			}
			for _, addr := range ep.Addresses {
				if _, ok := endpointMap[addr]; ok {
					continue
				}
				endpointMap[addr] = struct{}{}
				if *ep.NodeName == "" {
					klog.Warningf("[%s], Invalid endpoints without node name", util.Key(reqCtx.Service))
					continue
				}
				ret = append(ret, elbmodel.EdgeBackendAttribute{
					ServerId: *ep.NodeName,
					Type:     elbmodel.ServerGroupDefaultType,
					Weight:   getBackendWeight(*ep.NodeName, weightMap),
					Port:     elbmodel.ServerGroupDefaultPort,
				})
			}
		}
	}
	return ret
}

func setBackendFromEndpoints(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI) []elbmodel.EdgeBackendAttribute {
	ret := make([]elbmodel.EdgeBackendAttribute, 0)
	if len(candidates.Endpoints.Subsets) == 0 {
		return ret
	}
	endpointMap := make(map[string]struct{})
	weightMap := splitBackendWeight(reqCtx)
	for _, ep := range candidates.Endpoints.Subsets {
		for _, addr := range ep.Addresses {
			if _, ok := endpointMap[addr.IP]; ok {
				continue
			}
			endpointMap[addr.IP] = struct{}{}
			if *addr.NodeName == "" {
				klog.Warningf("[%s], Invalid endpoints without node name", util.Key(reqCtx.Service))
				continue
			}
			ret = append(ret, elbmodel.EdgeBackendAttribute{
				ServerId: *addr.NodeName,
				Type:     elbmodel.ServerGroupDefaultType,
				Weight:   getBackendWeight(*addr.NodeName, weightMap),
				Port:     elbmodel.ServerGroupDefaultPort,
			})
		}
	}
	return ret
}

func getCandidateBackend(candidates *backend.EndpointWithENI) map[string]struct{} {
	ret := make(map[string]struct{})
	for _, node := range candidates.Nodes {
		ret[node.Name] = struct{}{}
	}
	return ret
}

func getUpdateServerGroup(localModel, remoteModel *elbmodel.EdgeLoadBalancer) (addServerGroup, removeServerGroup, updateServerGroup *elbmodel.EdgeServerGroup) {
	addServerGroup = new(elbmodel.EdgeServerGroup)
	removeServerGroup = new(elbmodel.EdgeServerGroup)
	updateServerGroup = new(elbmodel.EdgeServerGroup)
	addServerGroup.Backends = make([]elbmodel.EdgeBackendAttribute, 0)
	removeServerGroup.Backends = make([]elbmodel.EdgeBackendAttribute, 0)
	updateServerGroup.Backends = make([]elbmodel.EdgeBackendAttribute, 0)

	if len(remoteModel.ServerGroup.Backends) == 0 {
		addServerGroup.Backends = append(addServerGroup.Backends, localModel.ServerGroup.Backends...)
		return
	}
	remoteServerMap := make(map[string]int)
	localServerMap := make(map[string]int)
	for idx, remoteServer := range remoteModel.ServerGroup.Backends {
		remoteServerMap[remoteServer.ServerId] = idx
	}
	for idx, localServer := range localModel.ServerGroup.Backends {
		localServerMap[localServer.ServerId] = idx
	}

	for _, localBackend := range localModel.ServerGroup.Backends {
		if _, ok := remoteServerMap[localBackend.ServerId]; !ok {
			addServerGroup.Backends = append(addServerGroup.Backends, localBackend)
			continue
		}
		if idx, ok := remoteServerMap[localBackend.ServerId]; ok {
			if localBackend.Weight != remoteModel.ServerGroup.Backends[idx].Weight {
				updateServerGroup.Backends = append(updateServerGroup.Backends, localBackend)
			}
		}
	}
	for _, remoteBackend := range remoteModel.ServerGroup.Backends {
		if _, ok := localServerMap[remoteBackend.ServerId]; !ok {
			removeServerGroup.Backends = append(removeServerGroup.Backends, remoteBackend)
		}
	}
	return
}

func splitBackendWeight(reqCtx *svcCtx.RequestContext) map[string]int {
	str := reqCtx.Anno.Get(annotation.EdgeServerWeight)
	ret := make(map[string]int)
	if str == "" {
		return ret
	}
	kv := strings.Split(str, ",")
	for _, e := range kv {
		r := strings.Split(e, "=")
		v, err := strconv.Atoi(r[1])
		if err != nil {
			reqCtx.Log.Error(err, fmt.Sprintf("parse backend weight [%s] error %s", e, err.Error()))
			v = elbmodel.ServerGroupDefaultServerWeight
		}
		if v < 0 || v > 100 {
			v = elbmodel.ServerGroupDefaultServerWeight
		}
		ret[r[0]] = v
	}
	return ret
}

func getBackendWeight(name string, weights map[string]int) int {
	ret := elbmodel.ServerGroupDefaultServerWeight
	if w, ok := weights[BaseBackendWeight]; ok {
		ret = w
	}
	if w, ok := weights[name]; ok {
		ret = w
	}
	return ret
}
