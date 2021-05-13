package alibaba

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
	"k8s.io/klog"
	"reflect"
	"strconv"
)

func NewLBProvider(
	auth *metadata.ClientAuth,
) *ProviderSLB {
	return &ProviderSLB{auth: auth}
}

var _ prvd.ILoadBalancer = &ProviderSLB{}

type ProviderSLB struct {
	auth *metadata.ClientAuth
}

func (p ProviderSLB) FindSLB(ctx context.Context, mdl *model.LoadBalancer) (bool, error) {

	// 1. find slb by loadbalancer id
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] try to find slb by loadbalancer id %s", mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		req := slb.CreateDescribeLoadBalancersRequest()
		req.LoadBalancerId = mdl.LoadBalancerAttribute.LoadBalancerId
		resp, err := p.auth.SLB.DescribeLoadBalancers(req)
		if err != nil {
			return false, err
		}
		if resp != nil && len(resp.LoadBalancers.LoadBalancer) > 0 {
			return getModelFromResponse(p, mdl, resp)
		}
	}

	// 2. find slb by tag
	items, err := json.Marshal(mdl.LoadBalancerAttribute.Tags)
	if err != nil {
		return false, err
	}
	klog.Infof("[%s] try to find slb by tag %s", mdl.NamespacedName, items)
	req := slb.CreateDescribeLoadBalancersRequest()
	req.Tags = string(items)
	resp, err := p.auth.SLB.DescribeLoadBalancers(req)
	if err != nil {
		return false, err
	}
	if resp != nil && len(resp.LoadBalancers.LoadBalancer) > 0 {
		return getModelFromResponse(p, mdl, resp)
	}

	// 3. find slb by name
	if mdl.LoadBalancerAttribute.LoadBalancerName != nil {
		klog.Infof("[%s] try to find slb by name %s", mdl.NamespacedName, *mdl.LoadBalancerAttribute.LoadBalancerName)
		req := slb.CreateDescribeLoadBalancersRequest()
		req.LoadBalancerName = *mdl.LoadBalancerAttribute.LoadBalancerName
		resp, err := p.auth.SLB.DescribeLoadBalancers(req)
		if err != nil {
			return false, err
		}
		if resp != nil && len(resp.LoadBalancers.LoadBalancer) > 0 {
			return getModelFromResponse(p, mdl, resp)
		}
	}

	return false, nil
}

func getModelFromResponse(p ProviderSLB, mdl *model.LoadBalancer, response *slb.DescribeLoadBalancersResponse) (bool, error) {
	if len(response.LoadBalancers.LoadBalancer) > 1 {
		klog.Warningf("find %d load balances by model, use the first one", len(response.LoadBalancers.LoadBalancer))
	}
	// Describe LoadBalancer Attribute
	req := slb.CreateDescribeLoadBalancerAttributeRequest()
	req.LoadBalancerId = response.LoadBalancers.LoadBalancer[0].LoadBalancerId
	resp, err := p.auth.SLB.DescribeLoadBalancerAttribute(req)
	if err != nil {
		return true, err
	}
	if resp == nil {
		// find the slb, but describe error. return true
		return true, fmt.Errorf("describe loadbalancer attribute error: slb [%s] is nil", req.LoadBalancerId)
	}
	setModelFromLoadBalancerAttributeResponse(resp, mdl)
	return true, nil

}

func (p ProviderSLB) CreateSLB(ctx context.Context, mdl *model.LoadBalancer) error {
	req := slb.CreateCreateLoadBalancerRequest()
	setCreateSLBReqFromModel(req, mdl)
	req.ClientToken = utils.GetUUID()
	resp, err := p.auth.SLB.CreateLoadBalancer(req)
	if err != nil {
		return err
	}
	mdl.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	mdl.LoadBalancerAttribute.Address = resp.Address
	return nil
}

func setCreateSLBReqFromModel(request *slb.CreateLoadBalancerRequest, mdl *model.LoadBalancer) {
	if mdl.LoadBalancerAttribute.AddressType != nil {
		request.AddressType = *mdl.LoadBalancerAttribute.AddressType
	}
	if mdl.LoadBalancerAttribute.InternetChargeType != nil {
		request.InternetChargeType = *mdl.LoadBalancerAttribute.InternetChargeType
	}
	if mdl.LoadBalancerAttribute.Bandwidth != nil {
		request.Bandwidth = requests.Integer(*mdl.LoadBalancerAttribute.Bandwidth)
	}
	if mdl.LoadBalancerAttribute.LoadBalancerName != nil {
		request.LoadBalancerName = *mdl.LoadBalancerAttribute.LoadBalancerName
	}
	if mdl.LoadBalancerAttribute.VpcId != "" {
		request.VpcId = mdl.LoadBalancerAttribute.VpcId
	}
	if mdl.LoadBalancerAttribute.VSwitchId != nil {
		request.VSwitchId = *mdl.LoadBalancerAttribute.VSwitchId
	}
	if mdl.LoadBalancerAttribute.MasterZoneId != nil {
		request.MasterZoneId = *mdl.LoadBalancerAttribute.MasterZoneId
	}
	if mdl.LoadBalancerAttribute.SlaveZoneId != nil {
		request.SlaveZoneId = *mdl.LoadBalancerAttribute.SlaveZoneId
	}
	if mdl.LoadBalancerAttribute.LoadBalancerSpec != nil {
		request.LoadBalancerSpec = *mdl.LoadBalancerAttribute.LoadBalancerSpec
	}
	if mdl.LoadBalancerAttribute.ResourceGroupId != nil {
		request.ResourceGroupId = *mdl.LoadBalancerAttribute.ResourceGroupId
	}
	if mdl.LoadBalancerAttribute.AddressIPVersion != nil {
		request.AddressIPVersion = *mdl.LoadBalancerAttribute.AddressIPVersion
	}
	if mdl.LoadBalancerAttribute.DeleteProtection != nil {
		request.DeleteProtection = *mdl.LoadBalancerAttribute.DeleteProtection
	}
	if mdl.LoadBalancerAttribute.ModificationProtectionStatus != nil {
		request.ModificationProtectionStatus = *mdl.LoadBalancerAttribute.ModificationProtectionStatus
		request.ModificationProtectionReason = mdl.LoadBalancerAttribute.ModificationProtectionReason
	}
}

func (p ProviderSLB) DeleteSLB(ctx context.Context, mdl *model.LoadBalancer) error {
	request := slb.CreateDeleteLoadBalancerRequest()
	request.LoadBalancerId = mdl.LoadBalancerAttribute.LoadBalancerId
	_, err := p.auth.SLB.DeleteLoadBalancer(request)
	return err
}

// TODO
func (p ProviderSLB) DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error) {
	req := slb.CreateDescribeLoadBalancerListenersRequest()
	req.LoadBalancerId = &[]string{lbId}
	resp, err := p.auth.SLB.DescribeLoadBalancerListeners(req)
	if err != nil {
		return nil, err
	}
	var listeners []model.ListenerAttribute
	for _, lis := range resp.Listeners {
		listeners = append(listeners, model.ListenerAttribute{
			Description:  lis.Description,
			ListenerPort: lis.ListenerPort,
			Protocol:     lis.ListenerProtocol,
			Bandwidth:    &lis.Bandwidth,
			Scheduler:    &lis.Scheduler,
			VGroupId:     lis.VServerGroupId,
		})
	}
	return listeners, nil
}

func (p ProviderSLB) StartLoadBalancerListener(ctx context.Context, loadBalancerId string, port int) error {
	req := slb.CreateStartLoadBalancerListenerRequest()
	req.LoadBalancerId = loadBalancerId
	req.ListenerPort = requests.NewInteger(port)
	_, err := p.auth.SLB.StartLoadBalancerListener(req)
	return err
}

func (p ProviderSLB) DescribeSLB(ctx context.Context, mdl *model.LoadBalancer) error {
	req := slb.CreateDescribeLoadBalancerAttributeRequest()
	req.LoadBalancerId = mdl.LoadBalancerAttribute.LoadBalancerId
	resp, err := p.auth.SLB.DescribeLoadBalancerAttribute(req)
	if err != nil {
		return err
	}
	setModelFromLoadBalancerAttributeResponse(resp, mdl)
	return nil
}

func setModelFromLoadBalancerAttributeResponse(resp *slb.DescribeLoadBalancerAttributeResponse, lb *model.LoadBalancer) {
	lb.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	lb.LoadBalancerAttribute.LoadBalancerName = &resp.LoadBalancerName
	lb.LoadBalancerAttribute.Address = resp.Address
	lb.LoadBalancerAttribute.AddressType = &resp.AddressType
	lb.LoadBalancerAttribute.NetworkType = &resp.NetworkType
	lb.LoadBalancerAttribute.VpcId = resp.VpcId
	lb.LoadBalancerAttribute.VSwitchId = &resp.VSwitchId
	lb.LoadBalancerAttribute.Bandwidth = &resp.Bandwidth
	lb.LoadBalancerAttribute.MasterZoneId = &resp.MasterZoneId
	lb.LoadBalancerAttribute.SlaveZoneId = &resp.SlaveZoneId
	lb.LoadBalancerAttribute.DeleteProtection = &resp.DeleteProtection
	lb.LoadBalancerAttribute.InternetChargeType = &resp.InternetChargeType
	lb.LoadBalancerAttribute.LoadBalancerSpec = &resp.LoadBalancerSpec
	lb.LoadBalancerAttribute.ModificationProtectionStatus = &resp.ModificationProtectionStatus
	lb.LoadBalancerAttribute.ResourceGroupId = &resp.ResourceGroupId
}

func (p ProviderSLB) CreateVServerGroup(ctx context.Context, vg *model.VServerGroup, lbId string) error {
	req := slb.CreateCreateVServerGroupRequest()
	req.LoadBalancerId = lbId
	req.VServerGroupName = vg.VGroupName
	backendJson, err := json.Marshal(vg.Backends)
	if err != nil {
		return err
	}
	req.BackendServers = string(backendJson)
	resp, err := p.auth.SLB.CreateVServerGroup(req)
	if err != nil {
		return err
	}
	vg.VGroupId = resp.VServerGroupId
	return nil
}

func (p ProviderSLB) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (*model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupAttributeRequest()
	req.VServerGroupId = vGroupId
	resp, err := p.auth.SLB.DescribeVServerGroupAttribute(req)
	if err != nil {
		return nil, err
	}
	vg := setVServerGroupFromResponse(resp)
	return vg, nil

}

func setVServerGroupFromResponse(resp *slb.DescribeVServerGroupAttributeResponse) *model.VServerGroup {
	vg := model.VServerGroup{
		VGroupId:   resp.VServerGroupId,
		VGroupName: resp.VServerGroupName,
		Backends:   nil,
	}
	// TODO
	return &vg

}

func (p ProviderSLB) DescribeVServerGroups(ctx context.Context, lbId string) ([]model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupsRequest()
	req.LoadBalancerId = lbId
	resp, err := p.auth.SLB.DescribeVServerGroups(req)
	if err != nil {
		return nil, err
	}
	var vgs []model.VServerGroup
	for _, v := range resp.VServerGroups.VServerGroup {
		vg := model.VServerGroup{
			VGroupId:   v.VServerGroupId,
			VGroupName: v.VServerGroupName,
		}
		vgs = append(vgs, vg)
	}
	return vgs, nil
}

func (p ProviderSLB) DeleteLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	req := slb.CreateDeleteLoadBalancerListenerRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)

	_, err := p.auth.SLB.DeleteLoadBalancerListener(req)
	return err
}

func (p ProviderSLB) StopLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	req := slb.CreateStopLoadBalancerListenerRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	_, err := p.auth.SLB.StopLoadBalancerListener(req)
	return err
}

func (p ProviderSLB) CreateLoadBalancerTCPListener(ctx context.Context, lbId string, port *model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerTCPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, port)
	_, err := p.auth.SLB.CreateLoadBalancerTCPListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerTCPListenerAttribute(ctx context.Context, lbId string, port *model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerTCPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, port)
	_, err := p.auth.SLB.SetLoadBalancerTCPListenerAttribute(req)
	return err
}

func (p ProviderSLB) CreateLoadBalancerUDPListener(ctx context.Context, lbId string, port *model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerUDPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, port)
	_, err := p.auth.SLB.CreateLoadBalancerUDPListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerUDPListenerAttribute(ctx context.Context, lbId string, port *model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerUDPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, port)
	_, err := p.auth.SLB.SetLoadBalancerUDPListenerAttribute(req)
	return err
}
func (p ProviderSLB) CreateLoadBalancerHTTPListener(ctx context.Context, lbId string, port *model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerHTTPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, port)
	_, err := p.auth.SLB.CreateLoadBalancerHTTPListener(req)
	return err
}
func (p ProviderSLB) SetLoadBalancerHTTPListenerAttribute(ctx context.Context, lbId string, port *model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerHTTPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, port)
	_, err := p.auth.SLB.SetLoadBalancerHTTPListenerAttribute(req)
	return err
}
func (p ProviderSLB) CreateLoadBalancerHTTPSListener(ctx context.Context, lbId string, port *model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerHTTPSListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, port)
	if port.CertId != nil {
		req.ServerCertificateId = *port.CertId
	}
	_, err := p.auth.SLB.CreateLoadBalancerHTTPSListener(req)
	return err
}
func (p ProviderSLB) SetLoadBalancerHTTPSListenerAttribute(ctx context.Context, lbId string, port *model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerHTTPSListenerAttributeRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, port)
	_, err := p.auth.SLB.SetLoadBalancerHTTPSListenerAttribute(req)
	return err
}

func setGenericListenerValue(req interface{}, port *model.ListenerAttribute) {
	v := reflect.ValueOf(req).Elem()

	listenerPort := v.FieldByName("ListenerPort")
	listenerPort.SetString(strconv.Itoa(port.ListenerPort))

	vGroupId := v.FieldByName("VServerGroupId")
	vGroupId.SetString(port.VGroupId)

	description := v.FieldByName("Description")
	description.SetString(port.Description)

	if port.Bandwidth != nil {
		bandwidth := v.FieldByName("Bandwidth")
		bandwidth.SetString(strconv.Itoa(*port.Bandwidth))
	}
	if port.Scheduler != nil {
		scheduler := v.FieldByName("Scheduler")
		scheduler.SetString(*port.Scheduler)
	}
	if port.HealthyThreshold != nil {
		healthyThreshold := v.FieldByName("HealthyThreshold")
		healthyThreshold.SetString(strconv.Itoa(*port.HealthyThreshold))
	}
	if port.UnhealthyThreshold != nil {
		unhealthyThreshold := v.FieldByName("UnhealthyThreshold")
		unhealthyThreshold.SetString(strconv.Itoa(*port.UnhealthyThreshold))
	}
	if port.HealthCheckConnectTimeout != nil {
		connectTimeout := v.FieldByName("HealthCheckConnectTimeout")
		connectTimeout.SetString(strconv.Itoa(*port.HealthCheckConnectTimeout))
	}
	if port.HealthCheckConnectPort != nil {
		connectPort := v.FieldByName("HealthCheckConnectPort")
		connectPort.SetString(strconv.Itoa(*port.HealthCheckConnectPort))
	}
	if port.HealthCheckInterval != nil {
		interval := v.FieldByName("HealthCheckInterval")
		interval.SetString(strconv.Itoa(*port.HealthCheckInterval))
	}
	if port.HealthCheckDomain != nil {
		domain := v.FieldByName("HealthCheckDomain")
		domain.SetString(*port.HealthCheckDomain)
	}
	if port.HealthCheckURI != nil {
		uri := v.FieldByName("HealthCheckURI")
		uri.SetString(*port.HealthCheckURI)
	}
	if port.HealthCheckHttpCode != nil {
		httpCode := v.FieldByName("HealthCheckHttpCode")
		httpCode.SetString(*port.HealthCheckHttpCode)
	}
	if port.HealthCheckType != nil {
		healthCheckType := v.FieldByName("HealthCheckType")
		healthCheckType.SetString(*port.HealthCheckType)
	}
	if port.AclType != nil {
		aclType := v.FieldByName("AclType")
		aclType.SetString(*port.AclType)
	}
	if port.AclStatus != nil {
		aclStatus := v.FieldByName("AclStatus")
		aclStatus.SetString(*port.AclStatus)
	}
}
