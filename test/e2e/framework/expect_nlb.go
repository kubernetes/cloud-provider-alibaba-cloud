package framework

import (
	"context"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"strconv"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/clbv1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/nlbv2"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
	"k8s.io/klog/v2"
)

func (f *Framework) ExpectNetworkLoadBalancerEqual(svc *v1.Service) error {
	reqCtx := &svcCtx.RequestContext{
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
	}

	var retErr error
	_ = wait.PollImmediate(10*time.Second, 3*time.Minute, func() (done bool, err error) {
		defer func() {
			if retErr != nil {
				klog.Infof("error in this try: %s", retErr.Error())
			}
		}()

		svc, remote, err := f.FindNetworkLoadBalancer()
		if err != nil {
			retErr = fmt.Errorf("find loadbalancer: %w", err)
			return false, nil
		}

		// check whether the nlb and svc is reconciled
		klog.Infof("check whether the nlb %s has been synced", remote.LoadBalancerAttribute.LoadBalancerId)
		if err := networkLoadBalancerAttrEqual(f, reqCtx.Anno, svc, remote.LoadBalancerAttribute); err != nil {
			retErr = fmt.Errorf("check nlb attr: %w", err)
			return false, nil
		}

		if reqCtx.Anno.Get(annotation.LoadBalancerId) == "" || isOverride(reqCtx.Anno) {
			if err := nlbListenerAttrEqual(reqCtx, remote.Listeners); err != nil {
				retErr = fmt.Errorf("check nlb listener attr: %w", err)
				return false, nil
			}
		}

		if err := nlbVsgAttrEqual(f, reqCtx, remote); err != nil {
			retErr = fmt.Errorf("check nlb vsg attr: %w", err)
			return false, nil
		}

		klog.Infof("nlb %s sync successfully", remote.LoadBalancerAttribute.LoadBalancerId)
		retErr = nil
		return true, nil
	})

	return retErr
}

func (f *Framework) ExpectNetworkLoadBalancerClean(svc *v1.Service, remote *nlbmodel.NetworkLoadBalancer) error {
	for _, lis := range remote.Listeners {
		if lis.IsUserManaged || lis.NamedKey == nil {
			continue
		}
		if lis.NamedKey.ServiceName == svc.Name &&
			lis.NamedKey.Namespace == svc.Namespace &&
			lis.NamedKey.CID == options.TestConfig.ClusterId {
			return fmt.Errorf("nlb %s listener %d is managed by ccm, but do not deleted",
				remote.LoadBalancerAttribute.LoadBalancerId, lis.ListenerPort)
		}
	}

	for _, sg := range remote.ServerGroups {
		if sg.IsUserManaged || sg.NamedKey == nil {
			continue
		}

		if sg.NamedKey.ServiceName == svc.Name &&
			sg.NamedKey.Namespace == svc.Namespace &&
			sg.NamedKey.CID == options.TestConfig.ClusterId {

			hasUserManagedNode := false
			for _, b := range sg.Servers {
				if b.Description != sg.ServerGroupName {
					hasUserManagedNode = true
				}
			}
			if !hasUserManagedNode {
				return fmt.Errorf("nlb %s server group %s is managed by ccm, but do not deleted",
					remote.LoadBalancerAttribute.LoadBalancerId, sg.ServerGroupId)
			}
		}
	}

	return nil
}

func (f *Framework) ExpectNetworkLoadBalancerDeleted(svc *v1.Service) error {
	reqCtx := &svcCtx.RequestContext{
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
	}
	lbManager := nlbv2.NewNLBManager(f.Client.CloudClient)

	return wait.PollImmediate(5*time.Second, 120*time.Second, func() (done bool, err error) {
		lbMdl := &nlbmodel.NetworkLoadBalancer{
			NamespacedName:        util.NamespacedName(svc),
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{},
		}
		err = lbManager.Find(reqCtx, lbMdl)
		if err != nil {
			return false, err
		}
		if lbMdl.LoadBalancerAttribute.LoadBalancerId != "" {
			return false, nil
		}
		return true, nil
	})
}

func (f *Framework) FindNetworkLoadBalancer() (*v1.Service, *nlbmodel.NetworkLoadBalancer, error) {
	// wait until service created successfully
	var svc *v1.Service
	err := wait.PollImmediate(10*time.Second, 60*time.Second, func() (done bool, err error) {
		svc, err = f.Client.KubeClient.GetService()
		if err != nil {
			return false, nil
		}
		klog.Infof("wait nlb service running, ingress: %+v", svc.Status.LoadBalancer.Ingress)
		if len(svc.Status.LoadBalancer.Ingress) == 1 &&
			(svc.Status.LoadBalancer.Ingress[0].IP != "" ||
				svc.Status.LoadBalancer.Ingress[0].Hostname != "") {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return svc, nil, err
	}

	klog.Infof("try get nlb from svc %s", svc.Name)

	remote, err := buildNLBRemoteModel(f, svc)
	if err != nil {
		return svc, nil, fmt.Errorf("build nlb remote model error: %s", err.Error())
	}
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return svc, nil, fmt.Errorf("nlb is nil")
	}
	return svc, remote, nil
}

func networkLoadBalancerAttrEqual(f *Framework, anno *annotation.AnnotationRequest, svc *v1.Service, nlb *nlbmodel.LoadBalancerAttribute) error {
	if id := anno.Get(annotation.LoadBalancerId); id != "" {
		if id != nlb.LoadBalancerId {
			return fmt.Errorf("expected nlb id %s, got %s", id, annotation.LoadBalancerId)
		}
	}

	if zoneMappings := anno.Get(annotation.ZoneMaps); zoneMappings != "" {
		localMappings, err := parseZoneMappings(zoneMappings)
		if err != nil {
			return fmt.Errorf("parse nlb local zone maps error: %s", err)
		}
		for _, local := range localMappings {
			found := false
			for _, remote := range nlb.ZoneMappings {
				if local.ZoneId == remote.ZoneId && local.VSwitchId == remote.VSwitchId {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected nlb zoneMappings %+v, got %+v", localMappings, nlb.ZoneMappings)
			}
		}
	}

	if addressType := nlbmodel.GetAddressType(anno.Get(annotation.AddressType)); addressType != "" {
		if addressType != nlb.AddressType {
			return fmt.Errorf("expected nlb address type %s, got %s", addressType, nlb.AddressType)
		}
	}

	if ipv6AddressType := anno.Get(annotation.IPv6AddressType); ipv6AddressType != "" {
		if !strings.EqualFold(ipv6AddressType, nlb.IPv6AddressType) {
			return fmt.Errorf("expected nlb ipv6 address type %s, got %s", ipv6AddressType, nlb.IPv6AddressType)
		}
	}

	if resourceGroupId := anno.Get(annotation.ResourceGroupId); resourceGroupId != "" {
		if resourceGroupId != nlb.ResourceGroupId {
			return fmt.Errorf("expected nlb resource group id %s, got %s", resourceGroupId, nlb.ResourceGroupId)
		}
	}

	if addressIpVersion := nlbmodel.GetAddressIpVersion(anno.Get(annotation.IPVersion)); addressIpVersion != "" {
		if !strings.EqualFold(addressIpVersion, nlb.AddressIpVersion) {
			return fmt.Errorf("expected nlb ip version %s, got %s", addressIpVersion, nlb.AddressIpVersion)
		}
	}

	if name := anno.Get(annotation.LoadBalancerName); name != "" {
		if name != nlb.Name {
			return fmt.Errorf("expected nlb name %s, got %s", name, nlb.Name)
		}
	}

	if additionalTags := anno.Get(annotation.AdditionalTags); additionalTags != "" {
		tags, err := f.Client.CloudClient.ListNLBTagResources(context.TODO(), nlb.LoadBalancerId)
		if err != nil {
			return err
		}
		defaultTags := anno.GetDefaultTags()
		defaultTags = append(defaultTags, tag.Tag{Key: helper.REUSEKEY, Value: "true"})
		var remoteTags []tag.Tag
		for _, r := range tags {
			found := false
			for _, t := range defaultTags {
				if t.Key == r.Key && t.Value == r.Value {
					found = true
					break
				}
			}
			if !found {
				remoteTags = append(remoteTags, r)
			}
		}
		if !tagsEqual(additionalTags, remoteTags) {
			return fmt.Errorf("expected nlb additional tags %s, got %v", additionalTags, remoteTags)
		}
	}

	if anno.Has(annotation.SecurityGroupIds) {
		id := anno.Get(annotation.SecurityGroupIds)
		var ids []string
		if id != "" {
			ids = strings.Split(id, ",")
		}
		if !util.IsStringSliceEqual(ids, nlb.SecurityGroupIds) {
			return fmt.Errorf("expected nlb security group ids %v, got %v", ids, nlb.SecurityGroupIds)
		}
	}
	return nil
}

func nlbListenerAttrEqual(reqCtx *svcCtx.RequestContext, remote []*nlbmodel.ListenerAttribute) error {
	portRange, err := nlbListenerPortRange(reqCtx)
	if err != nil {
		return err
	}
	for _, port := range reqCtx.Service.Spec.Ports {
		proto, err := nlbListenerProtocol(reqCtx.Anno.Get(annotation.ProtocolPort), port)
		if err != nil {
			return err
		}
		find := false
		for _, r := range remote {
			if nlbListenerPortEqual(r, port, portRange) && r.ListenerProtocol == proto {
				find = true
				switch proto {
				case nlbmodel.TCP:
					if err := nlbTCPEqual(reqCtx, port, r); err != nil {
						return err
					}
				case nlbmodel.UDP:
					if err := nlbUDPEqual(reqCtx, port, r); err != nil {
						return err
					}
				case nlbmodel.TCPSSL:
					if err := nlbTCPSSLEqual(reqCtx, port, r); err != nil {
						return err
					}
				}
			}
		}

		if !find {
			return fmt.Errorf("not found nlb listener %d, proto %s", port.Port, proto)
		}
	}
	return nil
}

func nlbListenerPortEqual(r *nlbmodel.ListenerAttribute, port v1.ServicePort, portRange map[int32][2]int32) bool {
	if r.ListenerPort != 0 {
		return r.ListenerPort == port.Port
	}
	if len(portRange) == 0 {
		return false
	}
	if ranges, ok := portRange[port.Port]; ok {
		return r.StartPort == ranges[0] && r.EndPort == ranges[1]
	}
	return false
}

func nlbListenerPortRange(reqCtx *svcCtx.RequestContext) (map[int32][2]int32, error) {
	ret := make(map[int32][2]int32)
	// 80-88:80
	pr := reqCtx.Anno.Get(annotation.ListenerPortRange)
	if pr == "" {
		return ret, nil
	}
	splits := strings.Split(pr, ",")
	for _, s := range splits {
		maps := strings.Split(s, ":")
		if len(maps) != 2 {
			return nil, fmt.Errorf("parse listener port range %s error", pr)
		}
		servicePort, err := strconv.Atoi(maps[1])
		if err != nil {
			return nil, fmt.Errorf("parse listener port range %s error: %s", pr, err.Error())
		}
		targetPorts := strings.Split(maps[0], "-")
		if len(targetPorts) != 2 {
			return nil, fmt.Errorf("parse listener port range %s error", pr)
		}
		startPort, err := strconv.Atoi(targetPorts[0])
		if err != nil {
			return nil, fmt.Errorf("parse listener port range %s error: %s", pr, err.Error())
		}
		endPort, err := strconv.Atoi(targetPorts[1])
		if err != nil {
			return nil, fmt.Errorf("parse listener port range %s error: %s", pr, err.Error())
		}

		ret[int32(servicePort)] = [2]int32{int32(startPort), int32(endPort)}
	}
	return ret, nil
}

func nlbVsgAttrEqual(f *Framework, reqCtx *svcCtx.RequestContext, remote *nlbmodel.NetworkLoadBalancer) error {
	portRange, err := nlbListenerPortRange(reqCtx)
	if err != nil {
		return err
	}
	for _, port := range reqCtx.Service.Spec.Ports {
		var (
			groupId string
			err     error
			weight  *int
		)
		proto, err := nlbListenerProtocol(reqCtx.Anno.Get(annotation.ProtocolPort), port)
		if err != nil {
			return err
		}
		var name string
		if portRange, ok := portRange[port.Port]; ok {
			name = getAnyPortServerGroupName(reqCtx.Service, proto, portRange[0], portRange[1])
		} else {
			name = getServerGroupName(reqCtx.Service, proto, &port)
		}
		if vGroupAnno := reqCtx.Anno.Get(annotation.VGroupPort); vGroupAnno != "" {
			groupId, err = getVGroupID(reqCtx.Anno.Get(annotation.VGroupPort), port)
			if err != nil {
				return fmt.Errorf("parse vgroup port annotation %s error: %s", vGroupAnno, err.Error())
			}

			if weightAnno := reqCtx.Anno.Get(annotation.VGroupWeight); weightAnno != "" {
				w, err := strconv.Atoi(weightAnno)
				if err != nil {
					return fmt.Errorf("parse vgroup weight annotation %s error: %s", weightAnno, err.Error())
				}
				weight = &w
			}
		}

		found := false
		for _, sg := range remote.ServerGroups {
			if sg.ServerGroupName == name {
				found = true
			}
			if sg.ServerGroupId == groupId {
				found = true
				sg.IsUserManaged = true
				sg.ServerGroupName = name
			}
			if found {
				sgType := reqCtx.Anno.Get(annotation.ServerGroupType)
				if sgType != "" && nlbmodel.ServerGroupType(sgType) != sg.ServerGroupType {
					return fmt.Errorf("server group %s type not equal, local: %s, remote: %s",
						sg.ServerGroupName, reqCtx.Anno.Get(annotation.ServerGroupType), sg.ServerGroupType)
				}

				sg.ServicePort = &port
				sg.ServicePort.Protocol = v1.Protocol(proto)
				sg.Weight = weight
				if isOverride(reqCtx.Anno) && !isNLBServerGroupUsedByPort(sg, remote.Listeners, portRange) {
					return fmt.Errorf("port %d do not use vgroup id: %s", port.Port, sg.ServerGroupId)
				}
				equal, err := isNLBBackendEqual(f, reqCtx, sg)
				if err != nil || !equal {
					return fmt.Errorf("port %d and vgroup %s do not have equal backends, error: CreateNLBServiceByAnno, %s",
						port.Port, sg.ServerGroupId, err)
				}
				err = serverGroupAttrEqual(reqCtx, sg)
				if err != nil {
					return err
				}
				break
			}
		}
		if !found {
			return fmt.Errorf("cannot found server group %s", name)
		}
	}
	return nil
}

func buildNLBRemoteModel(f *Framework, svc *v1.Service) (*nlbmodel.NetworkLoadBalancer, error) {
	sgMgr, err := nlbv2.NewServerGroupManager(f.Client.RuntimeClient, f.Client.CloudClient)
	if err != nil {
		return nil, err
	}
	builder := &nlbv2.ModelBuilder{
		NLBMgr: nlbv2.NewNLBManager(f.Client.CloudClient),
		LisMgr: nlbv2.NewListenerManager(f.Client.CloudClient),
		SGMgr:  sgMgr,
	}

	reqCtx := &svcCtx.RequestContext{
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
	}

	return builder.Instance(nlbv2.RemoteModel).Build(reqCtx)
}

func parseZoneMappings(zoneMaps string) ([]nlbmodel.ZoneMapping, error) {
	var ret []nlbmodel.ZoneMapping
	attrs := strings.Split(zoneMaps, ",")
	for _, attr := range attrs {
		items := strings.Split(attr, ":")
		if len(items) < 2 {
			return nil, fmt.Errorf("ZoneMapping format error, expect [zone-a:vsw-id-1,zone-b:vsw-id-2], got %s", zoneMaps)
		}
		zoneMap := nlbmodel.ZoneMapping{
			ZoneId:    items[0],
			VSwitchId: items[1],
		}

		if len(items) > 2 {
			zoneMap.IPv4Addr = items[2]
		}

		if len(items) > 3 {
			zoneMap.AllocationId = items[3]
		}
		ret = append(ret, zoneMap)
	}

	if len(ret) < 0 {
		return nil, fmt.Errorf("ZoneMapping format error, expect [zone-a:vsw-id-1,zone-b:vsw-id-2], got %s", zoneMaps)
	}
	return ret, nil
}

func nlbListenerProtocol(annotation string, port v1.ServicePort) (string, error) {
	if annotation == "" {
		return strings.ToUpper(string(port.Protocol)), nil
	}
	for _, v := range strings.Split(annotation, ",") {
		pp := strings.Split(v, ":")
		if len(pp) < 2 {
			return "", fmt.Errorf("port and "+
				"protocol format must be like 'https:443' with colon separated. got=[%+v]", pp)
		}

		if strings.ToUpper(pp[0]) != string(nlbmodel.TCP) &&
			strings.ToUpper(pp[0]) != string(nlbmodel.UDP) &&
			strings.ToUpper(pp[0]) != string(nlbmodel.TCPSSL) {
			return "", fmt.Errorf("port protocol"+
				" format must be either [TCP|UDP|TCPSSL], protocol not supported wit [%s]\n", pp[0])
		}

		if pp[1] == fmt.Sprintf("%d", port.Port) {
			util.ServiceLog.Info(fmt.Sprintf("port [%d] transform protocol from %s to %s", port.Port, port.Protocol, pp[0]))
			return strings.ToUpper(pp[0]), nil
		}
	}
	return strings.ToUpper(string(port.Protocol)), nil
}

func isNLBServerGroupUsedByPort(sg *nlbmodel.ServerGroup, listeners []*nlbmodel.ListenerAttribute, portRange map[int32][2]int32) bool {
	for _, l := range listeners {
		if nlbListenerPortEqual(l, *sg.ServicePort, portRange) &&
			strings.EqualFold(l.ListenerProtocol, string(sg.ServicePort.Protocol)) {
			return sg.ServerGroupId == l.ServerGroupId
		}
	}
	return false
}

func isNLBBackendEqual(f *Framework, reqCtx *svcCtx.RequestContext, sg *nlbmodel.ServerGroup) (bool, error) {
	policy := getTrafficPolicy(reqCtx)
	endpoints, err := f.Client.KubeClient.GetEndpoint()
	if err != nil {
		if !errors.IsNotFound(err) {
			return false, err
		}
		klog.Infof("endpoint is nil")
	}

	nodes, err := f.Client.KubeClient.ListNodes()
	if err != nil {
		return false, err
	}

	var backends []nlbmodel.ServerGroupServer
	switch policy {
	case helper.ENITrafficPolicy:
		backends, err = buildServerGroupENIBackends(f, reqCtx.Anno, endpoints, sg, model.IPv4)
		if err != nil {
			return false, err
		}
	case helper.LocalTrafficPolicy:
		backends, err = buildServerGroupLocalBackends(reqCtx.Anno, endpoints, nodes, sg)
		if err != nil {
			return false, err
		}
	case helper.ClusterTrafficPolicy:
		backends, err = buildServerGroupClusterBackends(reqCtx.Anno, endpoints, nodes, sg)
		if err != nil {
			return false, err
		}
	}
	for _, l := range backends {
		found := false
		for _, r := range sg.Servers {
			if isServerEqual(l, r) {
				if l.Port != r.Port {
					return false, fmt.Errorf("expected servergroup [%s] backend %s port not equal,"+
						" expect %d, got %d", sg.ServerGroupId, r.ServerId, l.Port, r.Port)
				}
				if l.Weight != r.Weight {
					return false, fmt.Errorf("expected servergroup [%s] backend %s weight not equal,"+
						" expect %d, got %d", sg.ServerGroupId, r.ServerId, l.Weight, r.Weight)
				}
				if l.Description != r.Description {
					return false, fmt.Errorf("expected servergroup [%s] backend %s description not equal,"+
						" expect %s, got %s", sg.ServerGroupId, r.ServerId, l.Description, r.Description)
				}
				found = true
				break
			}
		}
		if !found {
			return false, fmt.Errorf("mode %s expected vgroup [%s] has backend [%+v], got nil, backends [%s]",
				policy, sg.ServerGroupId, l, sg.BackendInfo())
		}
	}
	return true, nil
}

func isServerEqual(a, b nlbmodel.ServerGroupServer) bool {
	if a.ServerType != b.ServerType {
		return false
	}

	switch a.ServerType {
	case nlbmodel.EniServerType:
		return a.ServerIp == b.ServerIp
		//return a.ServerId == b.ServerId && a.ServerIp == b.ServerIp
	case nlbmodel.EcsServerType:
		return a.ServerId == b.ServerId
	case nlbmodel.IpServerType:
		return a.ServerId == b.ServerId && a.ServerIp == b.ServerIp
	default:
		klog.Errorf("%s is not supported, skip", a.ServerType)
		return false
	}
}

func nlbTCPEqual(reqCtx *svcCtx.RequestContext, local v1.ServicePort, remote *nlbmodel.ListenerAttribute) error {
	if err := genericNLBServerEqual(reqCtx, local, remote); err != nil {
		return err
	}
	return nil
}

func nlbUDPEqual(reqCtx *svcCtx.RequestContext, local v1.ServicePort, remote *nlbmodel.ListenerAttribute) error {
	if err := genericNLBServerEqual(reqCtx, local, remote); err != nil {
		return err
	}
	return nil
}

func nlbTCPSSLEqual(reqCtx *svcCtx.RequestContext, local v1.ServicePort, remote *nlbmodel.ListenerAttribute) error {
	if err := genericNLBServerEqual(reqCtx, local, remote); err != nil {
		return err
	}

	if certId := reqCtx.Anno.Get(annotation.CertID); certId != "" {
		localCerts := strings.Split(certId, ",")
		for _, local := range localCerts {
			found := false
			for _, remote := range remote.CertificateIds {
				if local == remote {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected nlb cert ids %v, got %v", localCerts, remote.CertificateIds)
			}
		}
	}

	if tlsCipherPolicy := reqCtx.Anno.Get(annotation.TLSCipherPolicy); tlsCipherPolicy != "" {
		if tlsCipherPolicy != remote.SecurityPolicyId {
			return fmt.Errorf("excpected nlb tls cipher policy %s, got %s", tlsCipherPolicy, remote.SecurityPolicyId)
		}
	}

	if cacert := reqCtx.Anno.Get(annotation.CaCert); cacert != "" {
		localEnabled := strings.EqualFold(cacert, string(model.OnFlag))
		remoteEnabled := remote.CaEnabled != nil && *remote.CaEnabled
		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb cacert %t, got %+v", localEnabled, remote.CaEnabled)
		}
	}

	if cacertId := reqCtx.Anno.Get(annotation.CaCertID); cacertId != "" {
		localCerts := strings.Split(cacertId, ",")
		for _, local := range localCerts {
			found := false
			for _, remote := range remote.CaCertificateIds {
				if local == remote {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected nlb cacert ids %v, got %v", localCerts, remote.CertificateIds)
			}
		}
	}

	if alpnEnabled := reqCtx.Anno.Get(annotation.AlpnEnabled); alpnEnabled != "" {
		localEnabled := strings.EqualFold(alpnEnabled, string(model.OnFlag))
		remoteEnabled := tea.BoolValue(remote.AlpnEnabled)
		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb alpn enabled %t, got %t(%+v)", localEnabled, tea.BoolValue(remote.AlpnEnabled), remote.AlpnEnabled)
		}

		if localEnabled && remote.AlpnPolicy != reqCtx.Anno.Get(annotation.AlpnPolicy) {
			return fmt.Errorf("expected nlb alpn policy %s, got %s", remote.AlpnPolicy, reqCtx.Anno.Get(annotation.AlpnPolicy))
		}
	}

	return nil
}

func getServerGroupName(svc *v1.Service, protocol string, servicePort *v1.ServicePort) string {
	sgPort := ""
	if helper.IsENIBackendType(svc) {
		switch servicePort.TargetPort.Type {
		case intstr.Int:
			sgPort = fmt.Sprintf("%d", servicePort.TargetPort.IntValue())
		case intstr.String:
			sgPort = servicePort.TargetPort.StrVal
		}
	} else {
		sgPort = fmt.Sprintf("%d", servicePort.NodePort)
	}
	namedKey := &nlbmodel.SGNamedKey{
		NamedKey: nlbmodel.NamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			Namespace:   svc.Namespace,
			CID:         base.CLUSTER_ID,
			ServiceName: svc.Name,
		},
		Protocol:    protocol,
		SGGroupPort: sgPort,
	}

	return namedKey.Key()
}

func getAnyPortServerGroupName(svc *v1.Service, protocol string, startPort, endPort int32) string {
	namedKey := &nlbmodel.SGNamedKey{
		NamedKey: nlbmodel.NamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			Namespace:   svc.Namespace,
			CID:         base.CLUSTER_ID,
			ServiceName: svc.Name,
		},
		Protocol:    protocol,
		SGGroupPort: fmt.Sprintf("%d_%d", startPort, endPort),
	}

	return namedKey.Key()
}

func buildServerGroupENIBackends(f *Framework, anno *annotation.AnnotationRequest, ep *v1.Endpoints, sg *nlbmodel.ServerGroup, ipVersion model.AddressIPVersionType) ([]nlbmodel.ServerGroupServer, error) {
	var ret []nlbmodel.ServerGroupServer
	for _, subset := range ep.Subsets {
		backendPort := getBackendPort(*sg.ServicePort, subset)
		if backendPort == 0 {
			continue
		}
		for _, address := range subset.Addresses {
			ret = append(ret, nlbmodel.ServerGroupServer{
				Description: sg.ServerGroupName,
				ServerIp:    address.IP,
				Port:        int32(backendPort),
			})
		}
	}

	var ips []string
	for _, b := range ret {
		ips = append(ips, b.ServerIp)
	}

	result, err := f.Client.CloudClient.DescribeNetworkInterfaces(options.TestConfig.VPCID, ips, ipVersion)
	if err != nil {
		return nil, fmt.Errorf("call DescribeNetworkInterfaces: %s", err.Error())
	}

	if sg.ServerGroupType == nlbmodel.IpServerGroupType {
		for i := range ret {
			ret[i].ServerId = ret[i].ServerIp
			ret[i].ServerType = nlbmodel.IpServerType
		}
	} else {
		for i := range ret {
			eniid, ok := result[ret[i].ServerIp]
			if !ok {
				return nil, fmt.Errorf("can not find eniid for ip %s in vpc %s", ret[i].ServerIp, options.TestConfig.VPCID)
			}
			ret[i].ServerId = eniid
			ret[i].ServerType = nlbmodel.EniServerType
			if sg.AnyPortEnabled {
				ret[i].Port = 0
			}
		}
	}
	return setServerGroupWeightBackends(helper.ENITrafficPolicy, ret, sg.Weight), nil
}

func buildServerGroupLocalBackends(anno *annotation.AnnotationRequest, ep *v1.Endpoints, nodes []v1.Node, sg *nlbmodel.ServerGroup) ([]nlbmodel.ServerGroupServer, error) {
	var ret []nlbmodel.ServerGroupServer
	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName == nil {
				return nil, fmt.Errorf("%s node name is nil", addr.IP)
			}
			node := findNodeByNodeName(nodes, *addr.NodeName)
			if node == nil {
				continue
			}
			if isNodeExcludeFromLoadBalancer(node, anno) {
				continue
			}

			_, id, err := helper.NodeFromProviderID(node.Spec.ProviderID)
			if err != nil {
				return nil, fmt.Errorf("providerID %s parse error: %s", node.Spec.ProviderID, err.Error())
			}
			if sg.ServerGroupType == nlbmodel.IpServerGroupType {
				ip, err := helper.GetNodeInternalIP(node)
				if err != nil {
					return nil, fmt.Errorf("get node address err: %s", err.Error())
				}
				ret = append(ret, nlbmodel.ServerGroupServer{
					Description: sg.ServerGroupName,
					ServerId:    ip,
					ServerIp:    ip,
					Port:        sg.ServicePort.NodePort,
					ServerType:  nlbmodel.IpServerType,
				})
			} else {
				ret = append(ret, nlbmodel.ServerGroupServer{
					Description: sg.ServerGroupName,
					ServerId:    id,
					Port:        sg.ServicePort.NodePort,
					ServerType:  nlbmodel.EcsServerType,
				})
			}

		}
	}

	eciBackends, err := buildServerGroupECIBackends(ep, nodes, sg)
	if err != nil {
		return nil, fmt.Errorf("build eci backends error: %s", err.Error())
	}

	return setServerGroupWeightBackends(helper.LocalTrafficPolicy, append(ret, eciBackends...), sg.Weight), nil
}

func buildServerGroupECIBackends(ep *v1.Endpoints, nodes []v1.Node, sg *nlbmodel.ServerGroup) ([]nlbmodel.ServerGroupServer, error) {
	var ret []nlbmodel.ServerGroupServer
	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName == nil {
				return nil, fmt.Errorf("%s node name is nil", addr.IP)
			}
			node := findNodeByNodeName(nodes, *addr.NodeName)
			if node == nil {
				continue
			}
			if isVKNode(*node) {
				backendPort := getBackendPort(*sg.ServicePort, subset)
				if sg.ServerGroupType == nlbmodel.IpServerGroupType {
					ret = append(ret, nlbmodel.ServerGroupServer{
						Description: sg.ServerGroupName,
						ServerId:    addr.IP,
						ServerIp:    addr.IP,
						Port:        int32(backendPort),
						ServerType:  nlbmodel.IpServerType,
					})
				} else {
					ret = append(ret, nlbmodel.ServerGroupServer{
						Description: sg.ServerGroupName,
						ServerIp:    addr.IP,
						Port:        int32(backendPort),
						ServerType:  model.ENIBackendType,
					})
				}
			}
		}
	}
	return ret, nil
}

func buildServerGroupClusterBackends(anno *annotation.AnnotationRequest, ep *v1.Endpoints, nodes []v1.Node, sg *nlbmodel.ServerGroup) ([]nlbmodel.ServerGroupServer, error) {
	var ret []nlbmodel.ServerGroupServer
	for _, n := range nodes {
		if isNodeExcludeFromLoadBalancer(&n, anno) {
			continue
		}
		_, id, err := helper.NodeFromProviderID(n.Spec.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("providerID %s parse error: %s", n.Spec.ProviderID, err.Error())
		}

		if sg.ServerGroupType == nlbmodel.IpServerGroupType {
			ip, err := helper.GetNodeInternalIP(&n)
			if err != nil {
				return nil, fmt.Errorf("get node address err: %s", err.Error())
			}
			ret = append(ret, nlbmodel.ServerGroupServer{
				Description: sg.ServerGroupName,
				ServerId:    ip,
				ServerIp:    ip,
				ServerType:  nlbmodel.IpServerType,
				Port:        sg.ServicePort.NodePort,
			})
		} else {
			ret = append(ret, nlbmodel.ServerGroupServer{
				Description: sg.ServerGroupName,
				ServerId:    id,
				ServerType:  nlbmodel.EcsServerType,
				Port:        sg.ServicePort.NodePort,
			})
		}

	}

	eciBackends, err := buildServerGroupECIBackends(ep, nodes, sg)
	if err != nil {
		return nil, fmt.Errorf("build eci backends error: %s", err.Error())
	}
	return setServerGroupWeightBackends(helper.ClusterTrafficPolicy, append(ret, eciBackends...), sg.Weight), nil
}

func setServerGroupWeightBackends(mode helper.TrafficPolicy, backends []nlbmodel.ServerGroupServer, weight *int) []nlbmodel.ServerGroupServer {
	// use default
	if weight == nil {
		return nlbPodNumberAlgorithm(mode, backends)
	}

	return nlbPodPercentAlgorithm(mode, backends, *weight)
}

func nlbPodNumberAlgorithm(mode helper.TrafficPolicy, backends []nlbmodel.ServerGroupServer) []nlbmodel.ServerGroupServer {
	if mode == helper.ENITrafficPolicy || mode == helper.ClusterTrafficPolicy {
		for i := range backends {
			backends[i].Weight = clbv1.DefaultServerWeight
		}
		return backends
	}

	// LocalTrafficPolicy
	ecsPods := make(map[string]int32)
	for _, b := range backends {
		ecsPods[b.ServerId] += 1
	}
	for i := range backends {
		backends[i].Weight = ecsPods[backends[i].ServerId]
	}
	return backends
}

func nlbPodPercentAlgorithm(mode helper.TrafficPolicy, backends []nlbmodel.ServerGroupServer, weight int) []nlbmodel.ServerGroupServer {
	if len(backends) == 0 {
		return backends
	}

	if weight == 0 {
		for i := range backends {
			backends[i].Weight = 0
		}
		return backends
	}

	if mode == helper.ENITrafficPolicy || mode == helper.ClusterTrafficPolicy {
		per := weight / len(backends)
		if per < 1 {
			per = 1
		}

		for i := range backends {
			backends[i].Weight = int32(per)
		}
		return backends
	}

	// LocalTrafficPolicy
	ecsPods := make(map[string]int)
	for _, b := range backends {
		ecsPods[b.ServerId] += 1
	}
	for i := range backends {
		backends[i].Weight = int32(weight * ecsPods[backends[i].ServerId] / len(backends))
		if backends[i].Weight < 1 {
			backends[i].Weight = 1
		}
	}
	return backends
}

func genericNLBServerEqual(reqCtx *svcCtx.RequestContext, local v1.ServicePort, remote *nlbmodel.ListenerAttribute) error {
	portRange, err := nlbListenerPortRange(reqCtx)
	if err != nil {
		return err
	}
	proto, err := nlbListenerProtocol(reqCtx.Anno.Get(annotation.ProtocolPort), local)
	if err != nil {
		return err
	}
	nameKey := &nlbmodel.ListenerNamedKey{
		NamedKey: nlbmodel.NamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			CID:         base.CLUSTER_ID,
			Namespace:   reqCtx.Service.Namespace,
			ServiceName: reqCtx.Service.Name,
		},
		Port:     local.Port,
		Protocol: proto,
	}
	if pr, ok := portRange[local.Port]; ok {
		nameKey.Port = 0
		nameKey.StartPort = pr[0]
		nameKey.EndPort = pr[1]
	}

	if remote.ListenerDescription != nameKey.Key() {
		return fmt.Errorf("expected listener description %s, got %s", nameKey.Key(), remote.ListenerDescription)
	}

	if cps := reqCtx.Anno.Get(annotation.Cps); cps != "" {
		cps, err := strconv.Atoi(cps)
		if err != nil {
			return fmt.Errorf("cps %s parse error: %s", cps, err.Error())
		}

		if remote.Cps == nil || int32(cps) != *remote.Cps {
			return fmt.Errorf("expected nlb cps %d, got %+v", cps, remote.Cps)
		}
	}

	if proxyProtocol := reqCtx.Anno.Get(annotation.ProxyProtocol); proxyProtocol != "" {
		localEnabled := strings.EqualFold(proxyProtocol, string(model.OnFlag))
		remoteEnabled := remote.ProxyProtocolEnabled != nil && *remote.ProxyProtocolEnabled
		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb proxy protocol %t, got %+v", localEnabled, remote.ProxyProtocolEnabled)
		}
	}

	if epIDEnabled := reqCtx.Anno.Get(annotation.Ppv2PrivateLinkEpIdEnabled); epIDEnabled != "" {
		localEnabled := strings.EqualFold(epIDEnabled, string(model.OnFlag))
		remoteEnabled := tea.BoolValue(remote.ProxyProtocolV2Config.PrivateLinkEpIdEnabled)
		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb ppv2 privatelink ep id enabled %t, got %t(%+v)", localEnabled, remoteEnabled, remote.ProxyProtocolV2Config.PrivateLinkEpIdEnabled)
		}
	}

	if epsIDEnabled := reqCtx.Anno.Get(annotation.Ppv2PrivateLinkEpsIdEnabled); epsIDEnabled != "" {
		localEnabled := strings.EqualFold(epsIDEnabled, string(model.OnFlag))
		remoteEnabled := tea.BoolValue(remote.ProxyProtocolV2Config.PrivateLinkEpsIdEnabled)
		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb ppv2 privatelink eps id enabled %t, got %t(%+v)", localEnabled, remoteEnabled, remote.ProxyProtocolV2Config.PrivateLinkEpsIdEnabled)
		}
	}

	if vpcIDEnabled := reqCtx.Anno.Get(annotation.Ppv2VpcIdEnabled); vpcIDEnabled != "" {
		localEnabled := strings.EqualFold(vpcIDEnabled, string(model.OnFlag))
		remoteEnabled := tea.BoolValue(remote.ProxyProtocolV2Config.VpcIdEnabled)
		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb ppv2 privatelink vpc id enabled %t, got %t(%+v)", localEnabled, remoteEnabled, remote.ProxyProtocolV2Config.VpcIdEnabled)
		}
	}

	if idleTimeout := reqCtx.Anno.Get(annotation.IdleTimeout); idleTimeout != "" {
		timeout, err := strconv.Atoi(idleTimeout)
		if err != nil {
			return fmt.Errorf("idle timeout %s parse error: %s", idleTimeout, err.Error())
		}

		if remote.IdleTimeout != int32(timeout) {
			return fmt.Errorf("expected nlb idle timeout %d, got %d", timeout, remote.IdleTimeout)
		}
	}

	return nil
}

func serverGroupAttrEqual(reqCtx *svcCtx.RequestContext, remote *nlbmodel.ServerGroup) error {
	if scheduler := reqCtx.Anno.Get(annotation.Scheduler); scheduler != "" {
		if !strings.EqualFold(scheduler, remote.Scheduler) {
			return fmt.Errorf("expected nlb listener scheduler %s, got %s", scheduler, remote.Scheduler)
		}
	}

	if connectionDrain := reqCtx.Anno.Get(annotation.ConnectionDrain); connectionDrain != "" {
		localEnabled := strings.EqualFold(connectionDrain, string(model.OnFlag))
		remoteEnabled := remote.ConnectionDrainEnabled != nil && *remote.ConnectionDrainEnabled
		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb listener connection drain %t, got %+v", localEnabled, remote.ConnectionDrainEnabled)
		}
	}

	if connectionDrainTimeout := reqCtx.Anno.Get(annotation.ConnectionDrainTimeout); connectionDrainTimeout != "" {
		timeout, err := strconv.Atoi(connectionDrainTimeout)
		if err != nil {
			return fmt.Errorf("error convert timeout to int: %s", err.Error())
		}

		if int32(timeout) != remote.ConnectionDrainTimeout {
			return fmt.Errorf("expected nlb listener connection drain timeout %d, got %d", timeout, remote.ConnectionDrainTimeout)
		}
	}

	if preserveClientIp := reqCtx.Anno.Get(annotation.PreserveClientIp); preserveClientIp != "" {
		localEnabled := strings.EqualFold(preserveClientIp, string(model.OnFlag))
		remoteEnabled := remote.PreserveClientIpEnabled != nil && *remote.PreserveClientIpEnabled
		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb listener preserve client ip %t, got %+v", localEnabled, remoteEnabled)
		}
	}

	if healthCheckFlag := reqCtx.Anno.Get(annotation.HealthCheckFlag); healthCheckFlag != "" {
		localEnabled := strings.EqualFold(healthCheckFlag, string(model.OnFlag))
		remoteEnabled := remote.HealthCheckConfig != nil && remote.HealthCheckConfig.HealthCheckEnabled != nil &&
			*remote.HealthCheckConfig.HealthCheckEnabled

		if localEnabled != remoteEnabled {
			return fmt.Errorf("expected nlb listener health check flag %t, got %t", localEnabled, remoteEnabled)
		}
	}

	if healthCheckType := reqCtx.Anno.Get(annotation.HealthCheckType); healthCheckType != "" {
		if remote.HealthCheckConfig == nil || !strings.EqualFold(healthCheckType, remote.HealthCheckConfig.HealthCheckType) {
			return fmt.Errorf("expected nlb listener health check type %s, got %+v", healthCheckType, remote.HealthCheckConfig)
		}
	}

	if healthCheckConnectTimeout := reqCtx.Anno.Get(annotation.HealthCheckConnectTimeout); healthCheckConnectTimeout != "" {
		timeout, err := strconv.Atoi(healthCheckConnectTimeout)
		if err != nil {
			return fmt.Errorf("error convert timeout to int: %s", healthCheckConnectTimeout)
		}

		if remote.HealthCheckConfig == nil || int32(timeout) != remote.HealthCheckConfig.HealthCheckConnectTimeout {
			return fmt.Errorf("expected nlb listener health check connect timeout %d, got %+v", timeout, remote.HealthCheckConfig)
		}
	}

	if healthyThreshold := reqCtx.Anno.Get(annotation.HealthyThreshold); healthyThreshold != "" {
		threshold, err := strconv.Atoi(healthyThreshold)
		if err != nil {
			return fmt.Errorf("error convert threshold to int: %s", healthyThreshold)
		}

		if remote.HealthCheckConfig == nil || int32(threshold) != remote.HealthCheckConfig.HealthyThreshold {
			return fmt.Errorf("expected healthy threshold %d, got %+v", threshold, remote.HealthCheckConfig)
		}
	}

	if unhealthyThreshold := reqCtx.Anno.Get(annotation.UnhealthyThreshold); unhealthyThreshold != "" {
		threshold, err := strconv.Atoi(unhealthyThreshold)
		if err != nil {
			return fmt.Errorf("error convert threshold to int: %s", unhealthyThreshold)
		}

		if remote.HealthCheckConfig == nil || int32(threshold) != remote.HealthCheckConfig.UnhealthyThreshold {
			return fmt.Errorf("expected unhealthy threshold %d, got %+v", threshold, remote.HealthCheckConfig)
		}
	}

	if healthCheckInterval := reqCtx.Anno.Get(annotation.HealthCheckInterval); healthCheckInterval != "" {
		interval, err := strconv.Atoi(healthCheckInterval)
		if err != nil {
			return fmt.Errorf("error convert interval to int: %s", healthCheckInterval)
		}

		if remote.HealthCheckConfig == nil || int32(interval) != remote.HealthCheckConfig.HealthCheckInterval {
			return fmt.Errorf("expected health check interval %d, got %+v", interval, remote.HealthCheckConfig)
		}
	}

	if healthCheckUri := reqCtx.Anno.Get(annotation.HealthCheckURI); healthCheckUri != "" {
		if remote.HealthCheckConfig == nil || healthCheckUri != remote.HealthCheckConfig.HealthCheckUrl {
			return fmt.Errorf("expected health check uri %s, got %+v", healthCheckUri, remote.HealthCheckConfig)
		}
	}

	if healthCheckDomain := reqCtx.Anno.Get(annotation.HealthCheckDomain); healthCheckDomain != "" {
		if remote.HealthCheckConfig == nil || healthCheckDomain != remote.HealthCheckConfig.HealthCheckDomain {
			return fmt.Errorf("expected health check uri %s, got %+v", healthCheckDomain, remote.HealthCheckConfig)
		}
	}

	if healCheckMethod := reqCtx.Anno.Get(annotation.HealthCheckMethod); healCheckMethod != "" {
		if remote.HealthCheckConfig == nil || healCheckMethod != remote.HealthCheckConfig.HttpCheckMethod {
			return fmt.Errorf("expected health check method %s, got %+v", healCheckMethod, remote.HealthCheckConfig)
		}
	}

	return nil
}
