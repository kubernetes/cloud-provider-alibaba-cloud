package elb

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewListenerManager(cloud prvd.Provider) *ListenerManager {
	return &ListenerManager{
		cloud: cloud,
	}
}

type ListenerManager struct {
	cloud prvd.Provider
}

func (mgr *ListenerManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	for _, port := range reqCtx.Service.Spec.Ports {
		listener, err := mgr.buildListenerFromServicePort(reqCtx, port)
		if err != nil {
			return fmt.Errorf("build local listener from servicePort %d error: %s", port.Port, err.Error())
		}
		mdl.Listeners.BackListener = append(mdl.Listeners.BackListener, listener)
	}
	return nil
}

func (mgr *ListenerManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	err := mgr.cloud.FindEdgeLoadBalancerListener(reqCtx.Ctx, mdl.GetLoadBalancerId(), &mdl.Listeners)
	if err != nil {
		return fmt.Errorf("build remote listener for load balancer %s error: %s", mdl.GetLoadBalancerId(), err.Error())
	}

	err = mgr.buildListenerFromCloud(reqCtx, mdl.GetLoadBalancerId(), &mdl.Listeners)
	if err != nil {
		return fmt.Errorf("build remote listener for load balancer %s error: %s", mdl.GetLoadBalancerId(), err.Error())
	}

	return nil
}

func (mgr *ListenerManager) buildListenerFromServicePort(reqCtx *svcCtx.RequestContext, port v1.ServicePort) (elbmodel.EdgeListenerAttribute, error) {

	listener := elbmodel.EdgeListenerAttribute{
		NamedKey: &elbmodel.ListenerNamedKey{
			NamedKey: elbmodel.NamedKey{
				Prefix:      model.DEFAULT_PREFIX,
				CID:         base.CLUSTER_ID,
				Namespace:   reqCtx.Service.Namespace,
				ServiceName: reqCtx.Service.Name,
			},
			Port: port.NodePort,
		},
		ListenerPort:     int(port.NodePort),
		ListenerProtocol: strings.ToLower(string(port.Protocol)),
		IsUserManaged:    false,
	}

	err := setListenerFromDefaultConfig(&listener)
	if err != nil {
		return listener, fmt.Errorf("build listener from default config error: %s", err.Error())
	}

	err = setListenerFromAnnotation(reqCtx, &listener)
	if err != nil {
		return listener, fmt.Errorf("build listener from annotation error: %s", err.Error())
	}
	return listener, nil
}

func (mgr *ListenerManager) buildListenerFromCloud(reqCtx *svcCtx.RequestContext, lbId string, listeners *elbmodel.EdgeListeners) error {
	if len(listeners.BackListener) < 1 {
		return nil
	}
	backListeners := make([]elbmodel.EdgeListenerAttribute, 0, len(listeners.BackListener))
	for _, listener := range listeners.BackListener {
		ls := new(elbmodel.EdgeListenerAttribute)
		switch listener.ListenerProtocol {
		case elbmodel.ProtocolTCP:
			if err := mgr.cloud.DescribeEdgeLoadBalancerTCPListener(reqCtx.Ctx, lbId, listener.ListenerPort, ls); err != nil {
				return err
			}
		case elbmodel.ProtocolUDP:
			if err := mgr.cloud.DescribeEdgeLoadBalancerUDPListener(reqCtx.Ctx, lbId, listener.ListenerPort, ls); err != nil {
				return err
			}
		case elbmodel.ProtocolHTTP:
			if err := mgr.cloud.DescribeEdgeLoadBalancerHTTPListener(reqCtx.Ctx, lbId, listener.ListenerPort, ls); err != nil {
				return err
			}
		case elbmodel.ProtocolHTTPS:
			if err := mgr.cloud.DescribeEdgeLoadBalancerHTTPSListener(reqCtx.Ctx, lbId, listener.ListenerPort, ls); err != nil {
				return err
			}
		default:
			continue
		}
		nameKey, err := elbmodel.LoadListenerNamedKey(ls.Description)
		if err != nil {
			ls.IsUserManaged = true
			reqCtx.Log.Info("listener description [%s], not expected format. skip user managed port", ls.Description)
		} else {
			ls.IsUserManaged = false
		}
		ls.NamedKey = nameKey
		backListeners = append(backListeners, *ls)
	}
	listeners.BackListener = backListeners
	return nil
}

func (mgr *ListenerManager) batchAddListeners(reqCtx *svcCtx.RequestContext, lbId string, listeners *[]elbmodel.EdgeListenerAttribute) error {
	listenPorts := *listeners
	for _, listener := range listenPorts {
		switch listener.ListenerProtocol {
		case elbmodel.ProtocolTCP:
			if err := mgr.cloud.CreateEdgeLoadBalancerTCPListener(reqCtx.Ctx, lbId, &listener); err != nil {
				return err
			}
		case elbmodel.ProtocolUDP:
			if err := mgr.cloud.CreateEdgeLoadBalancerUDPListener(reqCtx.Ctx, lbId, &listener); err != nil {
				return err
			}
		default:
			continue
		}
		if err := mgr.waitFindListeners(reqCtx, lbId, listener.ListenerPort, listener.ListenerProtocol); err != nil {
			return fmt.Errorf("batch add listener error %s", err.Error())
		}
		reqCtx.Log.Info("finish add listeners %s:%d from load balancer %s", listener.ListenerProtocol, listener.ListenerPort, lbId)
	}
	return nil
}

func (mgr *ListenerManager) batchRemoveListeners(reqCtx *svcCtx.RequestContext, lbId string, listeners *[]elbmodel.EdgeListenerAttribute) error {
	listenPorts := *listeners
	for _, listener := range listenPorts {
		err := mgr.cloud.DeleteEdgeLoadBalancerListener(reqCtx.Ctx, lbId, listener.ListenerPort, listener.ListenerProtocol)
		if err != nil {
			return err
		}
		reqCtx.Log.Info("finish remove listeners %s:%d from load balancer %s", listener.ListenerProtocol, listener.ListenerPort, lbId)
	}

	return nil
}

func (mgr *ListenerManager) batchUpdateListeners(reqCtx *svcCtx.RequestContext, lbId string, listeners *[]elbmodel.EdgeListenerAttribute) error {
	listenPorts := *listeners
	for _, listener := range listenPorts {
		switch listener.ListenerProtocol {
		case elbmodel.ProtocolTCP:
			if err := mgr.cloud.ModifyEdgeLoadBalancerTCPListener(reqCtx.Ctx, lbId, &listener); err != nil {
				return err
			}
		case elbmodel.ProtocolUDP:
			if err := mgr.cloud.ModifyEdgeLoadBalancerUDPListener(reqCtx.Ctx, lbId, &listener); err != nil {
				return err
			}
		default:
			continue
		}
		if err := mgr.waitFindListeners(reqCtx, lbId, listener.ListenerPort, listener.ListenerProtocol); err != nil {
			return err
		}
		reqCtx.Log.Info("finish update listeners %s:%d from load balancer %s", listener.ListenerProtocol, listener.ListenerPort, lbId)
	}
	return nil
}

func (mgr *ListenerManager) waitFindListeners(reqCtx *svcCtx.RequestContext, lbId string, port int, protocol string) error {
	waitErr := util.RetryImmediateOnError(10*time.Second, time.Minute, canSkipError, func() error {
		lis := new(elbmodel.EdgeListenerAttribute)
		switch protocol {
		case elbmodel.ProtocolTCP:
			if err := mgr.cloud.DescribeEdgeLoadBalancerTCPListener(reqCtx.Ctx, lbId, port, lis); err != nil {
				return fmt.Errorf("find no edge loadbalancer listener err %s", err.Error())
			}
		case elbmodel.ProtocolUDP:
			if err := mgr.cloud.DescribeEdgeLoadBalancerUDPListener(reqCtx.Ctx, lbId, port, lis); err != nil {
				return fmt.Errorf("find no edge loadbalancer listener err %s", err.Error())
			}
		}
		if lis == nil || lis.Status == "" {
			return fmt.Errorf("find no edge loadbalancer listener by port %s", fmt.Sprintf("%s:%d", protocol, port))
		}

		if lis.Status == ListenerRunning || lis.Status == ListenerStarting {
			return nil
		}
		if lis.Status == ListenerConfiguring || lis.Status == ListenerStopping {
			return fmt.Errorf("listener %s status is aberrant", lis.NamedKey.String())
		}
		if lis.Status == ListenerStopped {
			if err := mgr.cloud.StartEdgeLoadBalancerListener(reqCtx.Ctx, lbId, port, protocol); err != nil {
				return err
			}
		}
		return nil
	})
	if waitErr != nil {
		return waitErr
	}
	return nil
}

func setListenerFromDefaultConfig(listener *elbmodel.EdgeListenerAttribute) error {
	listener.Scheduler = elbmodel.ListenerDefaultScheduler
	listener.HealthThreshold = elbmodel.ListenerDefaultHealthThreshold
	listener.UnhealthyThreshold = elbmodel.ListenerDefaultUnhealthyThreshold
	listener.Description = listener.NamedKey.Key()
	switch listener.ListenerProtocol {
	case elbmodel.ProtocolTCP:
		listener.PersistenceTimeout = elbmodel.ListenerDefaultPersistenceTimeout
		listener.EstablishedTimeout = elbmodel.ListenerDefaultEstablishedTimeout
		listener.HealthCheckConnectTimeout = elbmodel.ListenerTCPDefaultHealthCheckConnectTimeout
		listener.HealthCheckInterval = elbmodel.ListenerTCPDefaultHealthCheckInterval
		listener.HealthCheckType = elbmodel.ProtocolTCP
		return nil
	case elbmodel.ProtocolUDP:
		listener.HealthCheckConnectTimeout = elbmodel.ListenerUDPDefaultHealthCheckConnectTimeout
		listener.HealthCheckInterval = elbmodel.ListenerUDPDefaultHealthCheckInterval
		return nil
	}
	return fmt.Errorf("unknown listening protocol %s", listener.ListenerProtocol)
}

func setListenerFromAnnotation(reqCtx *svcCtx.RequestContext, listener *elbmodel.EdgeListenerAttribute) error {
	var err error
	if strings.ToLower(reqCtx.Anno.Get(annotation.Scheduler)) != "" {
		listener.Scheduler = reqCtx.Anno.Get(annotation.Scheduler)
	}

	if reqCtx.Anno.Get(annotation.HealthCheckConnectPort) != "" {
		healthCheckConnectPort, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthCheckConnectPort))
		if err != nil {
			return fmt.Errorf("Annotation healthy threshold must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthCheckConnectPort), err.Error())
		}
		listener.HealthCheckConnectPort = healthCheckConnectPort
	}

	if reqCtx.Anno.Get(annotation.HealthyThreshold) != "" {
		healthyThreshold, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthyThreshold))
		if err != nil {
			return fmt.Errorf("Annotation healthy threshold must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthyThreshold), err.Error())
		}
		listener.HealthThreshold = healthyThreshold
	}

	if reqCtx.Anno.Get(annotation.UnhealthyThreshold) != "" {
		unhealthyThreshold, err := strconv.Atoi(reqCtx.Anno.Get(annotation.UnhealthyThreshold))
		if err != nil {
			return fmt.Errorf("Annotation unhealthy threshold must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.UnhealthyThreshold), err.Error())
		}
		listener.UnhealthyThreshold = unhealthyThreshold
	}

	if reqCtx.Anno.Get(annotation.HealthCheckInterval) != "" {
		healthyCheckInterval, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthCheckInterval))
		if err != nil {
			return fmt.Errorf("Annotation healthy check interval must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthCheckInterval), err.Error())
		}
		listener.HealthCheckInterval = healthyCheckInterval
	}

	if reqCtx.Anno.Get(annotation.HealthCheckConnectTimeout) != "" {
		healthCheckConnectTimeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthCheckConnectTimeout))
		if err != nil {
			return fmt.Errorf("Annotation health check connect timeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthCheckConnectTimeout), err.Error())
		}
		listener.HealthCheckConnectTimeout = healthCheckConnectTimeout
	}

	if listener.ListenerProtocol == elbmodel.ProtocolUDP {
		return nil
	}

	if reqCtx.Anno.Get(annotation.PersistenceTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.PersistenceTimeout))
		if err != nil {
			return fmt.Errorf("annotation persistence timeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.PersistenceTimeout), err.Error())
		}
		if timeout != 0 && reqCtx.Service.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeCluster {
			return fmt.Errorf("session persistence not support external traffic policy using cluster mode")
		}
		listener.PersistenceTimeout = timeout
	}

	if reqCtx.Anno.Get(annotation.EstablishedTimeout) != "" {
		establishedTimeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.EstablishedTimeout))
		if err != nil {
			return fmt.Errorf("Annotation established timeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.EstablishedTimeout), err.Error())
		}
		listener.EstablishedTimeout = establishedTimeout
	}

	return err
}

func getListeners(localModel, remoteModel *elbmodel.EdgeLoadBalancer) (add, remove, update []elbmodel.EdgeListenerAttribute, err error) {
	add = make([]elbmodel.EdgeListenerAttribute, 0)
	remove = make([]elbmodel.EdgeListenerAttribute, 0)
	update = make([]elbmodel.EdgeListenerAttribute, 0)
	if len(localModel.Listeners.BackListener) < 1 && len(remoteModel.Listeners.BackListener) < 1 {
		return
	}
	localMap := make(map[string]int)
	remoteMap := make(map[string]int)
	httpPortMap := make(map[string]string)
	for idx, local := range localModel.Listeners.BackListener {
		localMap[fmt.Sprintf("%s:%d", local.ListenerProtocol, local.ListenerPort)] = idx
	}

	for idx, remote := range remoteModel.Listeners.BackListener {
		switch remote.ListenerProtocol {
		case elbmodel.ProtocolHTTP:
			httpPortMap[strconv.Itoa(remote.ListenerPort)] = elbmodel.ProtocolHTTP
		case elbmodel.ProtocolHTTPS:
			httpPortMap[strconv.Itoa(remote.ListenerPort)] = elbmodel.ProtocolHTTPS
		default:
			remoteMap[fmt.Sprintf("%s:%d", remote.ListenerProtocol, remote.ListenerPort)] = idx
		}
	}
	for _, remote := range remoteModel.Listeners.BackListener {
		_, ok := localMap[fmt.Sprintf("%s:%d", remote.ListenerProtocol, remote.ListenerPort)]
		if !ok {
			if canRemoveListener(remoteModel, remote) {
				remove = append(remove, remote)
			}
		}
	}

	for _, local := range localModel.Listeners.BackListener {
		// if conflict with http/https port, return err
		if proto, ok := httpPortMap[strconv.Itoa(local.ListenerPort)]; ok {
			err = fmt.Errorf("listener %s:%d conflict with %s:%d",
				local.ListenerProtocol, local.ListenerPort, proto, local.ListenerPort)
			return
		}
		// add the listener if protocol:port is not found in remote,
		// update the listener if protocol:port is found in remote, but it is not manager by user
		idx, ok := remoteMap[fmt.Sprintf("%s:%d", local.ListenerProtocol, local.ListenerPort)]
		if !ok {
			add = append(add, local)
			continue
		}
		if ok && local.IsUserManaged {
			continue
		}

		if ok && !local.IsUserManaged {
			if listenerIsChanged(&local, &remoteModel.Listeners.BackListener[idx]) {
				update = append(update, local)
			}
		}
	}
	return
}

func canRemoveListener(mdl *elbmodel.EdgeLoadBalancer, listener elbmodel.EdgeListenerAttribute) bool {
	if listener.IsUserManaged {
		return false
	}
	if listener.NamedKey.CID != base.CLUSTER_ID {
		return false
	}
	if listener.NamedKey.Namespace != mdl.NamespacedName.Namespace {
		return false
	}
	if listener.NamedKey.ServiceName != mdl.NamespacedName.Name {
		return false
	}
	return true
}

func listenerIsChanged(local, remote *elbmodel.EdgeListenerAttribute) bool {
	if !isEqual(local.Scheduler, remote.Scheduler) {
		return true
	}
	if !isEqual(local.HealthThreshold, remote.HealthThreshold) {
		return true
	}
	if !isEqual(local.UnhealthyThreshold, remote.UnhealthyThreshold) {
		return true
	}
	if !isEqual(local.HealthCheckConnectTimeout, remote.HealthCheckConnectTimeout) {
		return true
	}
	if !isEqual(local.HealthCheckInterval, remote.HealthCheckInterval) {
		return true
	}
	if !isEqual(local.HealthCheckConnectPort, remote.HealthCheckConnectPort) {
		return true
	}
	if local.ListenerProtocol == elbmodel.ProtocolTCP && remote.ListenerProtocol == elbmodel.ProtocolTCP {
		if !isEqual(local.EstablishedTimeout, remote.EstablishedTimeout) {
			return true
		}
		if !isEqual(local.PersistenceTimeout, remote.PersistenceTimeout) {
			return true
		}
		if !isEqual(local.HealthCheckType, remote.HealthCheckType) {
			return true
		}
	}
	return false
}
