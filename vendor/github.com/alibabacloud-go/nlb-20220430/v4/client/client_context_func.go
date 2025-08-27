// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"context"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
)

// Summary:
//
// Adds backend servers to a specified server group.
//
// @param request - AddServersToServerGroupRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return AddServersToServerGroupResponse
func (client *Client) AddServersToServerGroupWithContext(ctx context.Context, request *AddServersToServerGroupRequest, runtime *dara.RuntimeOptions) (_result *AddServersToServerGroupResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	bodyFlat := map[string]interface{}{}
	if !dara.IsNil(request.Servers) {
		bodyFlat["Servers"] = request.Servers
	}

	body = dara.ToMap(body,
		openapiutil.Query(bodyFlat))
	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("AddServersToServerGroup"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &AddServersToServerGroupResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Associates additional certificates with a listener that uses SSL over TCP.
//
// Description:
//
// *AssociateAdditionalCertificatesWithListener*	- is an asynchronous operation. After a request is sent, the system returns a request ID and runs the task in the background. You can call the [ListListenerCertificates](https://help.aliyun.com/document_detail/615175.html) operation to query the status of the task:
//
//   - If the listener is in the **Associating*	- state, the additional certificates are being associated.
//
//   - If the listener is in the **Associated*	- state, the additional certificates are associated.
//
// @param request - AssociateAdditionalCertificatesWithListenerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return AssociateAdditionalCertificatesWithListenerResponse
func (client *Client) AssociateAdditionalCertificatesWithListenerWithContext(ctx context.Context, request *AssociateAdditionalCertificatesWithListenerRequest, runtime *dara.RuntimeOptions) (_result *AssociateAdditionalCertificatesWithListenerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.AdditionalCertificateIds) {
		body["AdditionalCertificateIds"] = request.AdditionalCertificateIds
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		body["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("AssociateAdditionalCertificatesWithListener"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &AssociateAdditionalCertificatesWithListenerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Associates an Internet Shared Bandwidth instance with a Network Load Balancer (NLB) instance.
//
// @param request - AttachCommonBandwidthPackageToLoadBalancerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return AttachCommonBandwidthPackageToLoadBalancerResponse
func (client *Client) AttachCommonBandwidthPackageToLoadBalancerWithContext(ctx context.Context, request *AttachCommonBandwidthPackageToLoadBalancerRequest, runtime *dara.RuntimeOptions) (_result *AttachCommonBandwidthPackageToLoadBalancerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.BandwidthPackageId) {
		body["BandwidthPackageId"] = request.BandwidthPackageId
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("AttachCommonBandwidthPackageToLoadBalancer"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &AttachCommonBandwidthPackageToLoadBalancerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Adds the elastic IP address (EIP) and virtual IP address (VIP) of a zone to the DNS record.
//
// Description:
//
// Before you call this operation, the zone of the Network Load Balancer (NLB) instance is removed from the DNS record by using the console or calling the [StartShiftLoadBalancerZones](https://help.aliyun.com/document_detail/2411999.html) API operation.
//
// @param request - CancelShiftLoadBalancerZonesRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return CancelShiftLoadBalancerZonesResponse
func (client *Client) CancelShiftLoadBalancerZonesWithContext(ctx context.Context, request *CancelShiftLoadBalancerZonesRequest, runtime *dara.RuntimeOptions) (_result *CancelShiftLoadBalancerZonesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ZoneMappings) {
		body["ZoneMappings"] = request.ZoneMappings
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("CancelShiftLoadBalancerZones"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &CancelShiftLoadBalancerZonesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Creates a TCP or UDP listener, or a listener that uses SSL over TCP for a Network Load Balancer (NLB) instance.
//
// @param tmpReq - CreateListenerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return CreateListenerResponse
func (client *Client) CreateListenerWithContext(ctx context.Context, tmpReq *CreateListenerRequest, runtime *dara.RuntimeOptions) (_result *CreateListenerResponse, _err error) {
	_err = tmpReq.Validate()
	if _err != nil {
		return _result, _err
	}
	request := &CreateListenerShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !dara.IsNil(tmpReq.ProxyProtocolV2Config) {
		request.ProxyProtocolV2ConfigShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.ProxyProtocolV2Config, dara.String("ProxyProtocolV2Config"), dara.String("json"))
	}

	body := map[string]interface{}{}
	if !dara.IsNil(request.AlpnEnabled) {
		body["AlpnEnabled"] = request.AlpnEnabled
	}

	if !dara.IsNil(request.AlpnPolicy) {
		body["AlpnPolicy"] = request.AlpnPolicy
	}

	if !dara.IsNil(request.CaCertificateIds) {
		body["CaCertificateIds"] = request.CaCertificateIds
	}

	if !dara.IsNil(request.CaEnabled) {
		body["CaEnabled"] = request.CaEnabled
	}

	if !dara.IsNil(request.CertificateIds) {
		body["CertificateIds"] = request.CertificateIds
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.Cps) {
		body["Cps"] = request.Cps
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.EndPort) {
		body["EndPort"] = request.EndPort
	}

	if !dara.IsNil(request.IdleTimeout) {
		body["IdleTimeout"] = request.IdleTimeout
	}

	if !dara.IsNil(request.ListenerDescription) {
		body["ListenerDescription"] = request.ListenerDescription
	}

	if !dara.IsNil(request.ListenerPort) {
		body["ListenerPort"] = request.ListenerPort
	}

	if !dara.IsNil(request.ListenerProtocol) {
		body["ListenerProtocol"] = request.ListenerProtocol
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.Mss) {
		body["Mss"] = request.Mss
	}

	if !dara.IsNil(request.ProxyProtocolEnabled) {
		body["ProxyProtocolEnabled"] = request.ProxyProtocolEnabled
	}

	if !dara.IsNil(request.ProxyProtocolV2ConfigShrink) {
		body["ProxyProtocolV2Config"] = request.ProxyProtocolV2ConfigShrink
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecSensorEnabled) {
		body["SecSensorEnabled"] = request.SecSensorEnabled
	}

	if !dara.IsNil(request.SecurityPolicyId) {
		body["SecurityPolicyId"] = request.SecurityPolicyId
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	if !dara.IsNil(request.StartPort) {
		body["StartPort"] = request.StartPort
	}

	if !dara.IsNil(request.Tag) {
		body["Tag"] = request.Tag
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("CreateListener"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &CreateListenerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Creates a Network Load Balancer (NLB) instance in a specified region.
//
// Description:
//
//	  When you create an NLB instance, the service-linked role AliyunServiceRoleForNlb is automatically created and assigned to you.
//
//		- **CreateLoadBalancer*	- is an asynchronous operation. After you send a request, the system returns an instance ID and runs the task in the background. You can call [GetLoadBalancerAttribute](https://help.aliyun.com/document_detail/445873.html) to query the status of an NLB instance.
//
//	    	- If an NLB instance is in the **Provisioning*	- state, the NLB instance is being created.
//
//	    	- If an NLB instance is in the **Active*	- state, the NLB instance is created.
//
// @param request - CreateLoadBalancerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return CreateLoadBalancerResponse
func (client *Client) CreateLoadBalancerWithContext(ctx context.Context, request *CreateLoadBalancerRequest, runtime *dara.RuntimeOptions) (_result *CreateLoadBalancerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.AddressIpVersion) {
		body["AddressIpVersion"] = request.AddressIpVersion
	}

	if !dara.IsNil(request.AddressType) {
		body["AddressType"] = request.AddressType
	}

	if !dara.IsNil(request.BandwidthPackageId) {
		body["BandwidthPackageId"] = request.BandwidthPackageId
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	bodyFlat := map[string]interface{}{}
	if !dara.IsNil(request.DeletionProtectionConfig) {
		bodyFlat["DeletionProtectionConfig"] = request.DeletionProtectionConfig
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerBillingConfig) {
		bodyFlat["LoadBalancerBillingConfig"] = request.LoadBalancerBillingConfig
	}

	if !dara.IsNil(request.LoadBalancerName) {
		body["LoadBalancerName"] = request.LoadBalancerName
	}

	if !dara.IsNil(request.LoadBalancerType) {
		body["LoadBalancerType"] = request.LoadBalancerType
	}

	if !dara.IsNil(request.ModificationProtectionConfig) {
		bodyFlat["ModificationProtectionConfig"] = request.ModificationProtectionConfig
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceGroupId) {
		body["ResourceGroupId"] = request.ResourceGroupId
	}

	if !dara.IsNil(request.Tag) {
		body["Tag"] = request.Tag
	}

	if !dara.IsNil(request.VpcId) {
		body["VpcId"] = request.VpcId
	}

	if !dara.IsNil(request.ZoneMappings) {
		bodyFlat["ZoneMappings"] = request.ZoneMappings
	}

	body = dara.ToMap(body,
		openapiutil.Query(bodyFlat))
	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("CreateLoadBalancer"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &CreateLoadBalancerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Creates a custom security policy for a TCP/SSL listener.
//
// @param request - CreateSecurityPolicyRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return CreateSecurityPolicyResponse
func (client *Client) CreateSecurityPolicyWithContext(ctx context.Context, request *CreateSecurityPolicyRequest, runtime *dara.RuntimeOptions) (_result *CreateSecurityPolicyResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.Ciphers) {
		body["Ciphers"] = request.Ciphers
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceGroupId) {
		body["ResourceGroupId"] = request.ResourceGroupId
	}

	if !dara.IsNil(request.SecurityPolicyName) {
		body["SecurityPolicyName"] = request.SecurityPolicyName
	}

	if !dara.IsNil(request.Tag) {
		body["Tag"] = request.Tag
	}

	if !dara.IsNil(request.TlsVersions) {
		body["TlsVersions"] = request.TlsVersions
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("CreateSecurityPolicy"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &CreateSecurityPolicyResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Creates a server group in a region.
//
// Description:
//
// *CreateServerGroup*	- is an asynchronous operation. After a request is sent, the system returns a request ID and runs the task in the background. You can call the [GetJobStatus](https://help.aliyun.com/document_detail/445904.html) operation to query the creation status of the task.
//
//   - If the task is in the **Succeeded*	- status, the server group is created.
//
// -    If the task is in the **Processing*	- status, the server group is being created.
//
// @param request - CreateServerGroupRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return CreateServerGroupResponse
func (client *Client) CreateServerGroupWithContext(ctx context.Context, request *CreateServerGroupRequest, runtime *dara.RuntimeOptions) (_result *CreateServerGroupResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.AddressIPVersion) {
		body["AddressIPVersion"] = request.AddressIPVersion
	}

	if !dara.IsNil(request.AnyPortEnabled) {
		body["AnyPortEnabled"] = request.AnyPortEnabled
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.ConnectionDrainEnabled) {
		body["ConnectionDrainEnabled"] = request.ConnectionDrainEnabled
	}

	if !dara.IsNil(request.ConnectionDrainTimeout) {
		body["ConnectionDrainTimeout"] = request.ConnectionDrainTimeout
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	bodyFlat := map[string]interface{}{}
	if !dara.IsNil(request.HealthCheckConfig) {
		bodyFlat["HealthCheckConfig"] = request.HealthCheckConfig
	}

	if !dara.IsNil(request.PreserveClientIpEnabled) {
		body["PreserveClientIpEnabled"] = request.PreserveClientIpEnabled
	}

	if !dara.IsNil(request.Protocol) {
		body["Protocol"] = request.Protocol
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceGroupId) {
		body["ResourceGroupId"] = request.ResourceGroupId
	}

	if !dara.IsNil(request.Scheduler) {
		body["Scheduler"] = request.Scheduler
	}

	if !dara.IsNil(request.ServerGroupName) {
		body["ServerGroupName"] = request.ServerGroupName
	}

	if !dara.IsNil(request.ServerGroupType) {
		body["ServerGroupType"] = request.ServerGroupType
	}

	if !dara.IsNil(request.Tag) {
		body["Tag"] = request.Tag
	}

	if !dara.IsNil(request.VpcId) {
		body["VpcId"] = request.VpcId
	}

	body = dara.ToMap(body,
		openapiutil.Query(bodyFlat))
	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("CreateServerGroup"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &CreateServerGroupResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Deletes a Network Load Balancer (NLB) listener.
//
// @param request - DeleteListenerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DeleteListenerResponse
func (client *Client) DeleteListenerWithContext(ctx context.Context, request *DeleteListenerRequest, runtime *dara.RuntimeOptions) (_result *DeleteListenerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		body["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DeleteListener"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DeleteListenerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Deletes a Network Load Balancer (NLB) instance.
//
// @param request - DeleteLoadBalancerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DeleteLoadBalancerResponse
func (client *Client) DeleteLoadBalancerWithContext(ctx context.Context, request *DeleteLoadBalancerRequest, runtime *dara.RuntimeOptions) (_result *DeleteLoadBalancerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DeleteLoadBalancer"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DeleteLoadBalancerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Deletes a custom TLS security policy from a Network Load Balancer (NLB) instance.
//
// @param request - DeleteSecurityPolicyRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DeleteSecurityPolicyResponse
func (client *Client) DeleteSecurityPolicyWithContext(ctx context.Context, request *DeleteSecurityPolicyRequest, runtime *dara.RuntimeOptions) (_result *DeleteSecurityPolicyResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecurityPolicyId) {
		body["SecurityPolicyId"] = request.SecurityPolicyId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DeleteSecurityPolicy"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DeleteSecurityPolicyResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Deletes a Network Load Balancer (NLB) server group.
//
// Description:
//
// You can delete server groups that are not associated with listeners.
//
// @param request - DeleteServerGroupRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DeleteServerGroupResponse
func (client *Client) DeleteServerGroupWithContext(ctx context.Context, request *DeleteServerGroupRequest, runtime *dara.RuntimeOptions) (_result *DeleteServerGroupResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DeleteServerGroup"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DeleteServerGroupResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the storage configurations of fine-grained monitoring.
//
// @param request - DescribeHdMonitorRegionConfigRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DescribeHdMonitorRegionConfigResponse
func (client *Client) DescribeHdMonitorRegionConfigWithContext(ctx context.Context, request *DescribeHdMonitorRegionConfigRequest, runtime *dara.RuntimeOptions) (_result *DescribeHdMonitorRegionConfigResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeHdMonitorRegionConfig"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DescribeHdMonitorRegionConfigResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries regions that support Network Load Balancer (NLB) instances.
//
// @param request - DescribeRegionsRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DescribeRegionsResponse
func (client *Client) DescribeRegionsWithContext(ctx context.Context, request *DescribeRegionsRequest, runtime *dara.RuntimeOptions) (_result *DescribeRegionsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.AcceptLanguage) {
		query["AcceptLanguage"] = request.AcceptLanguage
	}

	if !dara.IsNil(request.ServiceCode) {
		query["ServiceCode"] = request.ServiceCode
	}

	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
		Body:  openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeRegions"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DescribeRegionsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the zones of a region in which a Network Load Balancer (NLB) instance is deployed.
//
// @param request - DescribeZonesRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DescribeZonesResponse
func (client *Client) DescribeZonesWithContext(ctx context.Context, request *DescribeZonesRequest, runtime *dara.RuntimeOptions) (_result *DescribeZonesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.AcceptLanguage) {
		query["AcceptLanguage"] = request.AcceptLanguage
	}

	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ServiceCode) {
		query["ServiceCode"] = request.ServiceCode
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeZones"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DescribeZonesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Disassociates a Network Load Balancer (NLB) instance from an Internet Shared Bandwidth instance.
//
// @param request - DetachCommonBandwidthPackageFromLoadBalancerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DetachCommonBandwidthPackageFromLoadBalancerResponse
func (client *Client) DetachCommonBandwidthPackageFromLoadBalancerWithContext(ctx context.Context, request *DetachCommonBandwidthPackageFromLoadBalancerRequest, runtime *dara.RuntimeOptions) (_result *DetachCommonBandwidthPackageFromLoadBalancerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.BandwidthPackageId) {
		body["BandwidthPackageId"] = request.BandwidthPackageId
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DetachCommonBandwidthPackageFromLoadBalancer"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DetachCommonBandwidthPackageFromLoadBalancerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Changes the public IPv6 address of a dual-stack Network Load Balancer (NLB) instance to a private IPv6 address.
//
// @param request - DisableLoadBalancerIpv6InternetRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DisableLoadBalancerIpv6InternetResponse
func (client *Client) DisableLoadBalancerIpv6InternetWithContext(ctx context.Context, request *DisableLoadBalancerIpv6InternetRequest, runtime *dara.RuntimeOptions) (_result *DisableLoadBalancerIpv6InternetResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DisableLoadBalancerIpv6Internet"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DisableLoadBalancerIpv6InternetResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Disassociates additional certificates from a listener that uses SSL over TCP.
//
// Description:
//
// *DisassociateAdditionalCertificatesWithListener*	- is an asynchronous operation. After a request is sent, the system returns a request ID and runs the task in the background. You can call the [ListListenerCertificates](https://help.aliyun.com/document_detail/615175.html) operation to query the status of the task:
//
//   - If an additional certificate is in the **Dissociating*	- state, the additional certificate is being disassociated.
//
//   - If an additional certificate is in the **Dissociated*	- state, the additional certificate is disassociated.
//
// @param request - DisassociateAdditionalCertificatesWithListenerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return DisassociateAdditionalCertificatesWithListenerResponse
func (client *Client) DisassociateAdditionalCertificatesWithListenerWithContext(ctx context.Context, request *DisassociateAdditionalCertificatesWithListenerRequest, runtime *dara.RuntimeOptions) (_result *DisassociateAdditionalCertificatesWithListenerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.AdditionalCertificateIds) {
		body["AdditionalCertificateIds"] = request.AdditionalCertificateIds
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		body["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DisassociateAdditionalCertificatesWithListener"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &DisassociateAdditionalCertificatesWithListenerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Changes the network type of the IPv6 address of a dual-stack Network Load Balancer (NLB) instance from internal-facing to Internet-facing.
//
// @param request - EnableLoadBalancerIpv6InternetRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return EnableLoadBalancerIpv6InternetResponse
func (client *Client) EnableLoadBalancerIpv6InternetWithContext(ctx context.Context, request *EnableLoadBalancerIpv6InternetRequest, runtime *dara.RuntimeOptions) (_result *EnableLoadBalancerIpv6InternetResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("EnableLoadBalancerIpv6Internet"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &EnableLoadBalancerIpv6InternetResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the result of an asynchronous operation performed on a Network Load Balancer (NLB) instance.
//
// @param request - GetJobStatusRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return GetJobStatusResponse
func (client *Client) GetJobStatusWithContext(ctx context.Context, request *GetJobStatusRequest, runtime *dara.RuntimeOptions) (_result *GetJobStatusResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.JobId) {
		query["JobId"] = request.JobId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetJobStatus"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &GetJobStatusResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the details of a Network Load Balancer (NLB) listener.
//
// @param request - GetListenerAttributeRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return GetListenerAttributeResponse
func (client *Client) GetListenerAttributeWithContext(ctx context.Context, request *GetListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *GetListenerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		query["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		query["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetListenerAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &GetListenerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the health check status of a Network Load Balancer (NLB) listener.
//
// @param request - GetListenerHealthStatusRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return GetListenerHealthStatusResponse
func (client *Client) GetListenerHealthStatusWithContext(ctx context.Context, request *GetListenerHealthStatusRequest, runtime *dara.RuntimeOptions) (_result *GetListenerHealthStatusResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.ListenerId) {
		query["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetListenerHealthStatus"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &GetListenerHealthStatusResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the details about a Network Load Balancer (NLB) instance.
//
// @param request - GetLoadBalancerAttributeRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return GetLoadBalancerAttributeResponse
func (client *Client) GetLoadBalancerAttributeWithContext(ctx context.Context, request *GetLoadBalancerAttributeRequest, runtime *dara.RuntimeOptions) (_result *GetLoadBalancerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		query["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		query["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetLoadBalancerAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &GetLoadBalancerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the results of multiple asynchronous operations performed on a Network Load Balancer (NLB) instance.
//
// @param request - ListAsynJobsRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListAsynJobsResponse
func (client *Client) ListAsynJobsWithContext(ctx context.Context, request *ListAsynJobsRequest, runtime *dara.RuntimeOptions) (_result *ListAsynJobsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.JobIds) {
		query["JobIds"] = request.JobIds
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListAsynJobs"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListAsynJobsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the server certificate of a TCP/SSL listener.
//
// @param request - ListListenerCertificatesRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListListenerCertificatesResponse
func (client *Client) ListListenerCertificatesWithContext(ctx context.Context, request *ListListenerCertificatesRequest, runtime *dara.RuntimeOptions) (_result *ListListenerCertificatesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.CertType) {
		body["CertType"] = request.CertType
	}

	if !dara.IsNil(request.CertificateIds) {
		body["CertificateIds"] = request.CertificateIds
	}

	if !dara.IsNil(request.ListenerId) {
		body["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.MaxResults) {
		body["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		body["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListListenerCertificates"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListListenerCertificatesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries listeners added to a Network Load Balancer (NLB) instance.
//
// @param request - ListListenersRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListListenersResponse
func (client *Client) ListListenersWithContext(ctx context.Context, request *ListListenersRequest, runtime *dara.RuntimeOptions) (_result *ListListenersResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.ListenerIds) {
		query["ListenerIds"] = request.ListenerIds
	}

	if !dara.IsNil(request.ListenerProtocol) {
		query["ListenerProtocol"] = request.ListenerProtocol
	}

	if !dara.IsNil(request.LoadBalancerIds) {
		query["LoadBalancerIds"] = request.LoadBalancerIds
	}

	if !dara.IsNil(request.MaxResults) {
		query["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		query["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecSensorEnabled) {
		query["SecSensorEnabled"] = request.SecSensorEnabled
	}

	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListListeners"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListListenersResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the basic information about Network Load Balancer (NLB) instances.
//
// @param request - ListLoadBalancersRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListLoadBalancersResponse
func (client *Client) ListLoadBalancersWithContext(ctx context.Context, request *ListLoadBalancersRequest, runtime *dara.RuntimeOptions) (_result *ListLoadBalancersResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.AddressIpVersion) {
		query["AddressIpVersion"] = request.AddressIpVersion
	}

	if !dara.IsNil(request.AddressType) {
		query["AddressType"] = request.AddressType
	}

	if !dara.IsNil(request.DNSName) {
		query["DNSName"] = request.DNSName
	}

	if !dara.IsNil(request.Ipv6AddressType) {
		query["Ipv6AddressType"] = request.Ipv6AddressType
	}

	if !dara.IsNil(request.LoadBalancerBusinessStatus) {
		query["LoadBalancerBusinessStatus"] = request.LoadBalancerBusinessStatus
	}

	if !dara.IsNil(request.LoadBalancerIds) {
		query["LoadBalancerIds"] = request.LoadBalancerIds
	}

	if !dara.IsNil(request.LoadBalancerNames) {
		query["LoadBalancerNames"] = request.LoadBalancerNames
	}

	if !dara.IsNil(request.LoadBalancerStatus) {
		query["LoadBalancerStatus"] = request.LoadBalancerStatus
	}

	if !dara.IsNil(request.LoadBalancerType) {
		query["LoadBalancerType"] = request.LoadBalancerType
	}

	if !dara.IsNil(request.MaxResults) {
		query["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		query["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceGroupId) {
		query["ResourceGroupId"] = request.ResourceGroupId
	}

	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	if !dara.IsNil(request.VpcIds) {
		query["VpcIds"] = request.VpcIds
	}

	if !dara.IsNil(request.ZoneId) {
		query["ZoneId"] = request.ZoneId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListLoadBalancers"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListLoadBalancersResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the TLS security policies set for a Network Load Balancer (NLB) instance.
//
// @param request - ListSecurityPolicyRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListSecurityPolicyResponse
func (client *Client) ListSecurityPolicyWithContext(ctx context.Context, request *ListSecurityPolicyRequest, runtime *dara.RuntimeOptions) (_result *ListSecurityPolicyResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	body := map[string]interface{}{}
	if !dara.IsNil(request.MaxResults) {
		body["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		body["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceGroupId) {
		body["ResourceGroupId"] = request.ResourceGroupId
	}

	if !dara.IsNil(request.SecurityPolicyIds) {
		body["SecurityPolicyIds"] = request.SecurityPolicyIds
	}

	if !dara.IsNil(request.SecurityPolicyNames) {
		body["SecurityPolicyNames"] = request.SecurityPolicyNames
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
		Body:  openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListSecurityPolicy"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListSecurityPolicyResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries servers in a server group of a Network Load Balancer (NLB) instance.
//
// @param request - ListServerGroupServersRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListServerGroupServersResponse
func (client *Client) ListServerGroupServersWithContext(ctx context.Context, request *ListServerGroupServersRequest, runtime *dara.RuntimeOptions) (_result *ListServerGroupServersResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.MaxResults) {
		body["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		body["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	if !dara.IsNil(request.ServerIds) {
		body["ServerIds"] = request.ServerIds
	}

	if !dara.IsNil(request.ServerIps) {
		body["ServerIps"] = request.ServerIps
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListServerGroupServers"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListServerGroupServersResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the server groups of a Network Load Balancer (NLB) instance.
//
// @param request - ListServerGroupsRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListServerGroupsResponse
func (client *Client) ListServerGroupsWithContext(ctx context.Context, request *ListServerGroupsRequest, runtime *dara.RuntimeOptions) (_result *ListServerGroupsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	body := map[string]interface{}{}
	if !dara.IsNil(request.MaxResults) {
		body["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		body["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceGroupId) {
		body["ResourceGroupId"] = request.ResourceGroupId
	}

	if !dara.IsNil(request.ServerGroupIds) {
		body["ServerGroupIds"] = request.ServerGroupIds
	}

	if !dara.IsNil(request.ServerGroupNames) {
		body["ServerGroupNames"] = request.ServerGroupNames
	}

	if !dara.IsNil(request.ServerGroupType) {
		body["ServerGroupType"] = request.ServerGroupType
	}

	if !dara.IsNil(request.VpcId) {
		body["VpcId"] = request.VpcId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
		Body:  openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListServerGroups"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListServerGroupsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the default TLS policy.
//
// @param request - ListSystemSecurityPolicyRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListSystemSecurityPolicyResponse
func (client *Client) ListSystemSecurityPolicyWithContext(ctx context.Context, request *ListSystemSecurityPolicyRequest, runtime *dara.RuntimeOptions) (_result *ListSystemSecurityPolicyResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListSystemSecurityPolicy"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListSystemSecurityPolicyResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Queries the tags of a resource.
//
// @param request - ListTagResourcesRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return ListTagResourcesResponse
func (client *Client) ListTagResourcesWithContext(ctx context.Context, request *ListTagResourcesRequest, runtime *dara.RuntimeOptions) (_result *ListTagResourcesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.MaxResults) {
		body["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		body["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	bodyFlat := map[string]interface{}{}
	if !dara.IsNil(request.ResourceId) {
		bodyFlat["ResourceId"] = request.ResourceId
	}

	if !dara.IsNil(request.ResourceType) {
		body["ResourceType"] = request.ResourceType
	}

	if !dara.IsNil(request.Tag) {
		bodyFlat["Tag"] = request.Tag
	}

	body = dara.ToMap(body,
		openapiutil.Query(bodyFlat))
	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListTagResources"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListTagResourcesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Associates a security group with a Network Load Balancer (NLB) instance.
//
// Description:
//
//	  Make sure that you have created a security group. For more information about how to create a security group, see [CreateSecurityGroup](https://help.aliyun.com/document_detail/25553.html).
//
//		- An NLB instance can be associated with up to four security groups.
//
//		- You can query the security groups that are associated with an NLB instance by calling the [GetLoadBalancerAttribute](https://help.aliyun.com/document_detail/214362.html) operation.
//
//		- LoadBalancerJoinSecurityGroup is an asynchronous operation. After a request is sent, the system returns a request ID and runs the task in the background. You can call the [GetJobStatus](https://help.aliyun.com/document_detail/445904.html) operation to query the status of a task.
//
//	    	- If the task is in the **Succeeded*	- state, the security group is associated.
//
//	    	- If the task is in the **Processing*	- state, the security group is being associated. In this case, you can perform only query operations.
//
// @param request - LoadBalancerJoinSecurityGroupRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return LoadBalancerJoinSecurityGroupResponse
func (client *Client) LoadBalancerJoinSecurityGroupWithContext(ctx context.Context, request *LoadBalancerJoinSecurityGroupRequest, runtime *dara.RuntimeOptions) (_result *LoadBalancerJoinSecurityGroupResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecurityGroupIds) {
		body["SecurityGroupIds"] = request.SecurityGroupIds
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("LoadBalancerJoinSecurityGroup"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &LoadBalancerJoinSecurityGroupResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Disassociates a Network Load Balancer (NLB) instance from a security group.
//
// Description:
//
// LoadBalancerLeaveSecurityGroup is an asynchronous operation. After a request is sent, the system returns a request ID and runs the task in the background. You can call the [GetJobStatus](https://help.aliyun.com/document_detail/445904.html) operation to query the status of a task.
//
//   - If the task is in the **Succeeded*	- state, the security group is disassociated.
//
//   - If the task is in the **Processing*	- state, the security group is being disassociated. In this case, you can perform only query operations.
//
// @param request - LoadBalancerLeaveSecurityGroupRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return LoadBalancerLeaveSecurityGroupResponse
func (client *Client) LoadBalancerLeaveSecurityGroupWithContext(ctx context.Context, request *LoadBalancerLeaveSecurityGroupRequest, runtime *dara.RuntimeOptions) (_result *LoadBalancerLeaveSecurityGroupResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecurityGroupIds) {
		body["SecurityGroupIds"] = request.SecurityGroupIds
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("LoadBalancerLeaveSecurityGroup"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &LoadBalancerLeaveSecurityGroupResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Modify the group of resource.
//
// @param request - MoveResourceGroupRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return MoveResourceGroupResponse
func (client *Client) MoveResourceGroupWithContext(ctx context.Context, request *MoveResourceGroupRequest, runtime *dara.RuntimeOptions) (_result *MoveResourceGroupResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.NewResourceGroupId) {
		body["NewResourceGroupId"] = request.NewResourceGroupId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceId) {
		body["ResourceId"] = request.ResourceId
	}

	if !dara.IsNil(request.ResourceType) {
		body["ResourceType"] = request.ResourceType
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("MoveResourceGroup"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &MoveResourceGroupResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Deletes server groups from a Network Load Balancer (NLB) instance.
//
// @param request - RemoveServersFromServerGroupRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return RemoveServersFromServerGroupResponse
func (client *Client) RemoveServersFromServerGroupWithContext(ctx context.Context, request *RemoveServersFromServerGroupRequest, runtime *dara.RuntimeOptions) (_result *RemoveServersFromServerGroupResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	if !dara.IsNil(request.Servers) {
		body["Servers"] = request.Servers
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("RemoveServersFromServerGroup"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &RemoveServersFromServerGroupResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Configures storage for fine-grained monitoring.
//
// Description:
//
// This operation is used to configure a data warehouse for the fine-grained monitoring feature. If a listener in the current region has fine-grained monitoring enabled, you must disable fine-grained monitoring before you can configure a warehouse.
//
// @param request - SetHdMonitorRegionConfigRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return SetHdMonitorRegionConfigResponse
func (client *Client) SetHdMonitorRegionConfigWithContext(ctx context.Context, request *SetHdMonitorRegionConfigRequest, runtime *dara.RuntimeOptions) (_result *SetHdMonitorRegionConfigResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.LogProject) {
		query["LogProject"] = request.LogProject
	}

	if !dara.IsNil(request.MetricStore) {
		query["MetricStore"] = request.MetricStore
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("SetHdMonitorRegionConfig"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &SetHdMonitorRegionConfigResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Enables a Network Load Balancer (NLB) listener.
//
// @param request - StartListenerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return StartListenerResponse
func (client *Client) StartListenerWithContext(ctx context.Context, request *StartListenerRequest, runtime *dara.RuntimeOptions) (_result *StartListenerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		body["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("StartListener"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &StartListenerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Removes the elastic IP address (EIP) or virtual IP address (VIP) used in a zone from the DNS record.
//
// Description:
//
// >  If the NLB instance is deployed in only one zone, you cannot remove the EIP or VIP from the DNS record.
//
// @param request - StartShiftLoadBalancerZonesRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return StartShiftLoadBalancerZonesResponse
func (client *Client) StartShiftLoadBalancerZonesWithContext(ctx context.Context, request *StartShiftLoadBalancerZonesRequest, runtime *dara.RuntimeOptions) (_result *StartShiftLoadBalancerZonesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ZoneMappings) {
		body["ZoneMappings"] = request.ZoneMappings
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("StartShiftLoadBalancerZones"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &StartShiftLoadBalancerZonesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Stops a listener of a Network Load Balancer (NLB) instance.
//
// @param request - StopListenerRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return StopListenerResponse
func (client *Client) StopListenerWithContext(ctx context.Context, request *StopListenerRequest, runtime *dara.RuntimeOptions) (_result *StopListenerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		body["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("StopListener"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &StopListenerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Adds tags to specified resources.
//
// @param request - TagResourcesRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return TagResourcesResponse
func (client *Client) TagResourcesWithContext(ctx context.Context, request *TagResourcesRequest, runtime *dara.RuntimeOptions) (_result *TagResourcesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	bodyFlat := map[string]interface{}{}
	if !dara.IsNil(request.ResourceId) {
		bodyFlat["ResourceId"] = request.ResourceId
	}

	if !dara.IsNil(request.ResourceType) {
		body["ResourceType"] = request.ResourceType
	}

	if !dara.IsNil(request.Tag) {
		bodyFlat["Tag"] = request.Tag
	}

	body = dara.ToMap(body,
		openapiutil.Query(bodyFlat))
	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("TagResources"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &TagResourcesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Removes tags from resources.
//
// @param request - UntagResourcesRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UntagResourcesResponse
func (client *Client) UntagResourcesWithContext(ctx context.Context, request *UntagResourcesRequest, runtime *dara.RuntimeOptions) (_result *UntagResourcesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.All) {
		body["All"] = request.All
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	bodyFlat := map[string]interface{}{}
	if !dara.IsNil(request.ResourceId) {
		bodyFlat["ResourceId"] = request.ResourceId
	}

	if !dara.IsNil(request.ResourceType) {
		body["ResourceType"] = request.ResourceType
	}

	if !dara.IsNil(request.TagKey) {
		bodyFlat["TagKey"] = request.TagKey
	}

	body = dara.ToMap(body,
		openapiutil.Query(bodyFlat))
	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UntagResources"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UntagResourcesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Updates the attributes of a listener, such as the name and the idle connection timeout period.
//
// @param tmpReq - UpdateListenerAttributeRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UpdateListenerAttributeResponse
func (client *Client) UpdateListenerAttributeWithContext(ctx context.Context, tmpReq *UpdateListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateListenerAttributeResponse, _err error) {
	_err = tmpReq.Validate()
	if _err != nil {
		return _result, _err
	}
	request := &UpdateListenerAttributeShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !dara.IsNil(tmpReq.ProxyProtocolV2Config) {
		request.ProxyProtocolV2ConfigShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.ProxyProtocolV2Config, dara.String("ProxyProtocolV2Config"), dara.String("json"))
	}

	body := map[string]interface{}{}
	if !dara.IsNil(request.AlpnEnabled) {
		body["AlpnEnabled"] = request.AlpnEnabled
	}

	if !dara.IsNil(request.AlpnPolicy) {
		body["AlpnPolicy"] = request.AlpnPolicy
	}

	if !dara.IsNil(request.CaCertificateIds) {
		body["CaCertificateIds"] = request.CaCertificateIds
	}

	if !dara.IsNil(request.CaEnabled) {
		body["CaEnabled"] = request.CaEnabled
	}

	if !dara.IsNil(request.CertificateIds) {
		body["CertificateIds"] = request.CertificateIds
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.Cps) {
		body["Cps"] = request.Cps
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.IdleTimeout) {
		body["IdleTimeout"] = request.IdleTimeout
	}

	if !dara.IsNil(request.ListenerDescription) {
		body["ListenerDescription"] = request.ListenerDescription
	}

	if !dara.IsNil(request.ListenerId) {
		body["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.Mss) {
		body["Mss"] = request.Mss
	}

	if !dara.IsNil(request.ProxyProtocolEnabled) {
		body["ProxyProtocolEnabled"] = request.ProxyProtocolEnabled
	}

	if !dara.IsNil(request.ProxyProtocolV2ConfigShrink) {
		body["ProxyProtocolV2Config"] = request.ProxyProtocolV2ConfigShrink
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecSensorEnabled) {
		body["SecSensorEnabled"] = request.SecSensorEnabled
	}

	if !dara.IsNil(request.SecurityPolicyId) {
		body["SecurityPolicyId"] = request.SecurityPolicyId
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateListenerAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateListenerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Changes the network type of the IPv4 address for a Network Load Balancer (NLB) instance.
//
// Description:
//
//	  Make sure that an NLB instance is created. For more information, see [CreateLoadBalancer](https://help.aliyun.com/document_detail/445868.html).
//
//		- You can call the [GetLoadBalancerAttribute](https://help.aliyun.com/document_detail/445873.html) operation to query the **AddressType*	- value of an NLB instance after you change the network type.
//
//		- **UpdateLoadBalancerAddressTypeConfig*	- is an asynchronous operation. After a request is sent, the system returns a request ID and runs the task in the background. You can call the [GetJobStatus](https://help.aliyun.com/document_detail/445904.html) operation to query the task status:
//
//	    	- If the task is in the **Succeeded*	- state, the network type of the IPv4 address of the NLB instance is changed.
//
//	    	- If the task is in the **Processing*	- state, the network type of the IPv4 address of the NLB instance is being changed. In this case, you can perform only query operations.
//
// @param request - UpdateLoadBalancerAddressTypeConfigRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UpdateLoadBalancerAddressTypeConfigResponse
func (client *Client) UpdateLoadBalancerAddressTypeConfigWithContext(ctx context.Context, request *UpdateLoadBalancerAddressTypeConfigRequest, runtime *dara.RuntimeOptions) (_result *UpdateLoadBalancerAddressTypeConfigResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.AddressType) {
		body["AddressType"] = request.AddressType
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ZoneMappings) {
		body["ZoneMappings"] = request.ZoneMappings
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateLoadBalancerAddressTypeConfig"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateLoadBalancerAddressTypeConfigResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Updates the attributes, including the name, of a Network Load Balancer (NLB) instance.
//
// @param request - UpdateLoadBalancerAttributeRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UpdateLoadBalancerAttributeResponse
func (client *Client) UpdateLoadBalancerAttributeWithContext(ctx context.Context, request *UpdateLoadBalancerAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateLoadBalancerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.Cps) {
		body["Cps"] = request.Cps
	}

	if !dara.IsNil(request.CrossZoneEnabled) {
		body["CrossZoneEnabled"] = request.CrossZoneEnabled
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.LoadBalancerName) {
		body["LoadBalancerName"] = request.LoadBalancerName
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateLoadBalancerAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateLoadBalancerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Enables or disables the deletion protection feature.
//
// Description:
//
// > You can call the [GetLoadBalancerAttribute](https://help.aliyun.com/document_detail/445873.html) operation to query the details about deletion protection and the configuration read-only mode.
//
// @param request - UpdateLoadBalancerProtectionRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UpdateLoadBalancerProtectionResponse
func (client *Client) UpdateLoadBalancerProtectionWithContext(ctx context.Context, request *UpdateLoadBalancerProtectionRequest, runtime *dara.RuntimeOptions) (_result *UpdateLoadBalancerProtectionResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DeletionProtectionEnabled) {
		body["DeletionProtectionEnabled"] = request.DeletionProtectionEnabled
	}

	if !dara.IsNil(request.DeletionProtectionReason) {
		body["DeletionProtectionReason"] = request.DeletionProtectionReason
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.ModificationProtectionReason) {
		body["ModificationProtectionReason"] = request.ModificationProtectionReason
	}

	if !dara.IsNil(request.ModificationProtectionStatus) {
		body["ModificationProtectionStatus"] = request.ModificationProtectionStatus
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateLoadBalancerProtection"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateLoadBalancerProtectionResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Modifies the zones and zone attributes of a Network Load Balancer (NLB) instance.
//
// Description:
//
// When you call this operation, make sure that you specify all the zones of the NLB instance, including the existing zones and new zones. If you do not specify the existing zones, the existing zones are removed.
//
// Prerequisites
//
//   - An NLB instance is created. For more information, see [CreateLoadBalancer](https://help.aliyun.com/document_detail/445868.html).
//
//   - You can call the [GetLoadBalancerAttribute](https://help.aliyun.com/document_detail/445873.html) operation to query the zones and zone attributes of an NLB instance.
//
//   - **UpdateLoadBalancerZones*	- is an asynchronous operation. After a request is sent, the system returns a request ID and runs the task in the background. You can call the [GetJobStatus](https://help.aliyun.com/document_detail/445904.html) operation query to query the status of a task:
//
//   - If the task is in the **Succeeded*	- state, the zones and zone attributes are modified.
//
//   - If the task is in the **Processing*	- state, the zones and zone attributes are being modified. In this case, you can perform only query operations.
//
// @param request - UpdateLoadBalancerZonesRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UpdateLoadBalancerZonesResponse
func (client *Client) UpdateLoadBalancerZonesWithContext(ctx context.Context, request *UpdateLoadBalancerZonesRequest, runtime *dara.RuntimeOptions) (_result *UpdateLoadBalancerZonesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		body["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ZoneMappings) {
		body["ZoneMappings"] = request.ZoneMappings
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateLoadBalancerZones"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateLoadBalancerZonesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Modifies the configurations of a security policy for a Network Load Balancer (NLB) instance.
//
// @param request - UpdateSecurityPolicyAttributeRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UpdateSecurityPolicyAttributeResponse
func (client *Client) UpdateSecurityPolicyAttributeWithContext(ctx context.Context, request *UpdateSecurityPolicyAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateSecurityPolicyAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.Ciphers) {
		body["Ciphers"] = request.Ciphers
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecurityPolicyId) {
		body["SecurityPolicyId"] = request.SecurityPolicyId
	}

	if !dara.IsNil(request.SecurityPolicyName) {
		body["SecurityPolicyName"] = request.SecurityPolicyName
	}

	if !dara.IsNil(request.TlsVersions) {
		body["TlsVersions"] = request.TlsVersions
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateSecurityPolicyAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateSecurityPolicyAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Modifies the configurations of a Network Load Balancer (NLB) server group.
//
// @param request - UpdateServerGroupAttributeRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UpdateServerGroupAttributeResponse
func (client *Client) UpdateServerGroupAttributeWithContext(ctx context.Context, request *UpdateServerGroupAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateServerGroupAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.ConnectionDrainEnabled) {
		body["ConnectionDrainEnabled"] = request.ConnectionDrainEnabled
	}

	if !dara.IsNil(request.ConnectionDrainTimeout) {
		body["ConnectionDrainTimeout"] = request.ConnectionDrainTimeout
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	bodyFlat := map[string]interface{}{}
	if !dara.IsNil(request.HealthCheckConfig) {
		bodyFlat["HealthCheckConfig"] = request.HealthCheckConfig
	}

	if !dara.IsNil(request.PreserveClientIpEnabled) {
		body["PreserveClientIpEnabled"] = request.PreserveClientIpEnabled
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.Scheduler) {
		body["Scheduler"] = request.Scheduler
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	if !dara.IsNil(request.ServerGroupName) {
		body["ServerGroupName"] = request.ServerGroupName
	}

	body = dara.ToMap(body,
		openapiutil.Query(bodyFlat))
	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateServerGroupAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateServerGroupAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Modifies the weights and descriptions of backend servers in a server group of a Network Load Balancer (NLB) instance.
//
// Description:
//
// *UpdateServerGroupServersAttribute*	- is an asynchronous operation. After a request is sent, the system returns a request ID and runs the task in the background.
//
// 1.  You can call the [ListServerGroups](https://help.aliyun.com/document_detail/445895.html) operation to query the status of a server group.
//
//   - If a server group is in the **Configuring*	- state, the server group is being modified.
//
//   - If a server group is in the **Available*	- state, the server group is running.
//
// 2.  You can call the [ListServerGroupServers](https://help.aliyun.com/document_detail/445896.html) operation to query the status of a backend server.
//
//   - If a backend server is in the **Configuring*	- state, it indicates that the backend server is being modified.
//
//   - If a backend server is in the **Available*	- state, it indicates that the backend server is running.
//
// @param request - UpdateServerGroupServersAttributeRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return UpdateServerGroupServersAttributeResponse
func (client *Client) UpdateServerGroupServersAttributeWithContext(ctx context.Context, request *UpdateServerGroupServersAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateServerGroupServersAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	if !dara.IsNil(request.Servers) {
		body["Servers"] = request.Servers
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateServerGroupServersAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateServerGroupServersAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
