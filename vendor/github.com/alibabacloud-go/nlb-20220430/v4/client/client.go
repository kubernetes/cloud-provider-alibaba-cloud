// This file is auto-generated, don't edit it. Thanks.
package client

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
)

type Client struct {
	openapi.Client
	DisableSDKError *bool
}

func NewClient(config *openapiutil.Config) (*Client, error) {
	client := new(Client)
	err := client.Init(config)
	return client, err
}

func (client *Client) Init(config *openapiutil.Config) (_err error) {
	_err = client.Client.Init(config)
	if _err != nil {
		return _err
	}
	client.EndpointRule = dara.String("regional")
	_err = client.CheckConfig(config)
	if _err != nil {
		return _err
	}
	client.Endpoint, _err = client.GetEndpoint(dara.String("nlb"), client.RegionId, client.EndpointRule, client.Network, client.Suffix, client.EndpointMap, client.Endpoint)
	if _err != nil {
		return _err
	}

	return nil
}

func (client *Client) GetEndpoint(productId *string, regionId *string, endpointRule *string, network *string, suffix *string, endpointMap map[string]*string, endpoint *string) (_result *string, _err error) {
	if !dara.IsNil(endpoint) {
		_result = endpoint
		return _result, _err
	}

	if !dara.IsNil(endpointMap) && !dara.IsNil(endpointMap[dara.StringValue(regionId)]) {
		_result = endpointMap[dara.StringValue(regionId)]
		return _result, _err
	}

	_body, _err := openapiutil.GetEndpointRules(productId, regionId, endpointRule, network, suffix)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

// Summary:
//
// Adds backend servers to a specified server group.
//
// @param request - AddServersToServerGroupRequest
//
// @param runtime - runtime options for this request RuntimeOptions
//
// @return AddServersToServerGroupResponse
func (client *Client) AddServersToServerGroupWithOptions(request *AddServersToServerGroupRequest, runtime *dara.RuntimeOptions) (_result *AddServersToServerGroupResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// Summary:
//
// Adds backend servers to a specified server group.
//
// @param request - AddServersToServerGroupRequest
//
// @return AddServersToServerGroupResponse
func (client *Client) AddServersToServerGroup(request *AddServersToServerGroupRequest) (_result *AddServersToServerGroupResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &AddServersToServerGroupResponse{}
	_body, _err := client.AddServersToServerGroupWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) AssociateAdditionalCertificatesWithListenerWithOptions(request *AssociateAdditionalCertificatesWithListenerRequest, runtime *dara.RuntimeOptions) (_result *AssociateAdditionalCertificatesWithListenerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return AssociateAdditionalCertificatesWithListenerResponse
func (client *Client) AssociateAdditionalCertificatesWithListener(request *AssociateAdditionalCertificatesWithListenerRequest) (_result *AssociateAdditionalCertificatesWithListenerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &AssociateAdditionalCertificatesWithListenerResponse{}
	_body, _err := client.AssociateAdditionalCertificatesWithListenerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) AttachCommonBandwidthPackageToLoadBalancerWithOptions(request *AttachCommonBandwidthPackageToLoadBalancerRequest, runtime *dara.RuntimeOptions) (_result *AttachCommonBandwidthPackageToLoadBalancerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return AttachCommonBandwidthPackageToLoadBalancerResponse
func (client *Client) AttachCommonBandwidthPackageToLoadBalancer(request *AttachCommonBandwidthPackageToLoadBalancerRequest) (_result *AttachCommonBandwidthPackageToLoadBalancerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &AttachCommonBandwidthPackageToLoadBalancerResponse{}
	_body, _err := client.AttachCommonBandwidthPackageToLoadBalancerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) CancelShiftLoadBalancerZonesWithOptions(request *CancelShiftLoadBalancerZonesRequest, runtime *dara.RuntimeOptions) (_result *CancelShiftLoadBalancerZonesResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return CancelShiftLoadBalancerZonesResponse
func (client *Client) CancelShiftLoadBalancerZones(request *CancelShiftLoadBalancerZonesRequest) (_result *CancelShiftLoadBalancerZonesResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &CancelShiftLoadBalancerZonesResponse{}
	_body, _err := client.CancelShiftLoadBalancerZonesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) CreateListenerWithOptions(tmpReq *CreateListenerRequest, runtime *dara.RuntimeOptions) (_result *CreateListenerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @param request - CreateListenerRequest
//
// @return CreateListenerResponse
func (client *Client) CreateListener(request *CreateListenerRequest) (_result *CreateListenerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &CreateListenerResponse{}
	_body, _err := client.CreateListenerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) CreateLoadBalancerWithOptions(request *CreateLoadBalancerRequest, runtime *dara.RuntimeOptions) (_result *CreateLoadBalancerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return CreateLoadBalancerResponse
func (client *Client) CreateLoadBalancer(request *CreateLoadBalancerRequest) (_result *CreateLoadBalancerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &CreateLoadBalancerResponse{}
	_body, _err := client.CreateLoadBalancerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) CreateSecurityPolicyWithOptions(request *CreateSecurityPolicyRequest, runtime *dara.RuntimeOptions) (_result *CreateSecurityPolicyResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return CreateSecurityPolicyResponse
func (client *Client) CreateSecurityPolicy(request *CreateSecurityPolicyRequest) (_result *CreateSecurityPolicyResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &CreateSecurityPolicyResponse{}
	_body, _err := client.CreateSecurityPolicyWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) CreateServerGroupWithOptions(request *CreateServerGroupRequest, runtime *dara.RuntimeOptions) (_result *CreateServerGroupResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return CreateServerGroupResponse
func (client *Client) CreateServerGroup(request *CreateServerGroupRequest) (_result *CreateServerGroupResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &CreateServerGroupResponse{}
	_body, _err := client.CreateServerGroupWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DeleteListenerWithOptions(request *DeleteListenerRequest, runtime *dara.RuntimeOptions) (_result *DeleteListenerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DeleteListenerResponse
func (client *Client) DeleteListener(request *DeleteListenerRequest) (_result *DeleteListenerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DeleteListenerResponse{}
	_body, _err := client.DeleteListenerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DeleteLoadBalancerWithOptions(request *DeleteLoadBalancerRequest, runtime *dara.RuntimeOptions) (_result *DeleteLoadBalancerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DeleteLoadBalancerResponse
func (client *Client) DeleteLoadBalancer(request *DeleteLoadBalancerRequest) (_result *DeleteLoadBalancerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DeleteLoadBalancerResponse{}
	_body, _err := client.DeleteLoadBalancerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DeleteSecurityPolicyWithOptions(request *DeleteSecurityPolicyRequest, runtime *dara.RuntimeOptions) (_result *DeleteSecurityPolicyResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DeleteSecurityPolicyResponse
func (client *Client) DeleteSecurityPolicy(request *DeleteSecurityPolicyRequest) (_result *DeleteSecurityPolicyResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DeleteSecurityPolicyResponse{}
	_body, _err := client.DeleteSecurityPolicyWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DeleteServerGroupWithOptions(request *DeleteServerGroupRequest, runtime *dara.RuntimeOptions) (_result *DeleteServerGroupResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DeleteServerGroupResponse
func (client *Client) DeleteServerGroup(request *DeleteServerGroupRequest) (_result *DeleteServerGroupResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DeleteServerGroupResponse{}
	_body, _err := client.DeleteServerGroupWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DescribeHdMonitorRegionConfigWithOptions(request *DescribeHdMonitorRegionConfigRequest, runtime *dara.RuntimeOptions) (_result *DescribeHdMonitorRegionConfigResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DescribeHdMonitorRegionConfigResponse
func (client *Client) DescribeHdMonitorRegionConfig(request *DescribeHdMonitorRegionConfigRequest) (_result *DescribeHdMonitorRegionConfigResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DescribeHdMonitorRegionConfigResponse{}
	_body, _err := client.DescribeHdMonitorRegionConfigWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DescribeRegionsWithOptions(request *DescribeRegionsRequest, runtime *dara.RuntimeOptions) (_result *DescribeRegionsResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DescribeRegionsResponse
func (client *Client) DescribeRegions(request *DescribeRegionsRequest) (_result *DescribeRegionsResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DescribeRegionsResponse{}
	_body, _err := client.DescribeRegionsWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DescribeZonesWithOptions(request *DescribeZonesRequest, runtime *dara.RuntimeOptions) (_result *DescribeZonesResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DescribeZonesResponse
func (client *Client) DescribeZones(request *DescribeZonesRequest) (_result *DescribeZonesResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DescribeZonesResponse{}
	_body, _err := client.DescribeZonesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DetachCommonBandwidthPackageFromLoadBalancerWithOptions(request *DetachCommonBandwidthPackageFromLoadBalancerRequest, runtime *dara.RuntimeOptions) (_result *DetachCommonBandwidthPackageFromLoadBalancerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DetachCommonBandwidthPackageFromLoadBalancerResponse
func (client *Client) DetachCommonBandwidthPackageFromLoadBalancer(request *DetachCommonBandwidthPackageFromLoadBalancerRequest) (_result *DetachCommonBandwidthPackageFromLoadBalancerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DetachCommonBandwidthPackageFromLoadBalancerResponse{}
	_body, _err := client.DetachCommonBandwidthPackageFromLoadBalancerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DisableLoadBalancerIpv6InternetWithOptions(request *DisableLoadBalancerIpv6InternetRequest, runtime *dara.RuntimeOptions) (_result *DisableLoadBalancerIpv6InternetResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DisableLoadBalancerIpv6InternetResponse
func (client *Client) DisableLoadBalancerIpv6Internet(request *DisableLoadBalancerIpv6InternetRequest) (_result *DisableLoadBalancerIpv6InternetResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DisableLoadBalancerIpv6InternetResponse{}
	_body, _err := client.DisableLoadBalancerIpv6InternetWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) DisassociateAdditionalCertificatesWithListenerWithOptions(request *DisassociateAdditionalCertificatesWithListenerRequest, runtime *dara.RuntimeOptions) (_result *DisassociateAdditionalCertificatesWithListenerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return DisassociateAdditionalCertificatesWithListenerResponse
func (client *Client) DisassociateAdditionalCertificatesWithListener(request *DisassociateAdditionalCertificatesWithListenerRequest) (_result *DisassociateAdditionalCertificatesWithListenerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &DisassociateAdditionalCertificatesWithListenerResponse{}
	_body, _err := client.DisassociateAdditionalCertificatesWithListenerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) EnableLoadBalancerIpv6InternetWithOptions(request *EnableLoadBalancerIpv6InternetRequest, runtime *dara.RuntimeOptions) (_result *EnableLoadBalancerIpv6InternetResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return EnableLoadBalancerIpv6InternetResponse
func (client *Client) EnableLoadBalancerIpv6Internet(request *EnableLoadBalancerIpv6InternetRequest) (_result *EnableLoadBalancerIpv6InternetResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &EnableLoadBalancerIpv6InternetResponse{}
	_body, _err := client.EnableLoadBalancerIpv6InternetWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) GetJobStatusWithOptions(request *GetJobStatusRequest, runtime *dara.RuntimeOptions) (_result *GetJobStatusResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return GetJobStatusResponse
func (client *Client) GetJobStatus(request *GetJobStatusRequest) (_result *GetJobStatusResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &GetJobStatusResponse{}
	_body, _err := client.GetJobStatusWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) GetListenerAttributeWithOptions(request *GetListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *GetListenerAttributeResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return GetListenerAttributeResponse
func (client *Client) GetListenerAttribute(request *GetListenerAttributeRequest) (_result *GetListenerAttributeResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &GetListenerAttributeResponse{}
	_body, _err := client.GetListenerAttributeWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) GetListenerHealthStatusWithOptions(request *GetListenerHealthStatusRequest, runtime *dara.RuntimeOptions) (_result *GetListenerHealthStatusResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return GetListenerHealthStatusResponse
func (client *Client) GetListenerHealthStatus(request *GetListenerHealthStatusRequest) (_result *GetListenerHealthStatusResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &GetListenerHealthStatusResponse{}
	_body, _err := client.GetListenerHealthStatusWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) GetLoadBalancerAttributeWithOptions(request *GetLoadBalancerAttributeRequest, runtime *dara.RuntimeOptions) (_result *GetLoadBalancerAttributeResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return GetLoadBalancerAttributeResponse
func (client *Client) GetLoadBalancerAttribute(request *GetLoadBalancerAttributeRequest) (_result *GetLoadBalancerAttributeResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &GetLoadBalancerAttributeResponse{}
	_body, _err := client.GetLoadBalancerAttributeWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListAsynJobsWithOptions(request *ListAsynJobsRequest, runtime *dara.RuntimeOptions) (_result *ListAsynJobsResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListAsynJobsResponse
func (client *Client) ListAsynJobs(request *ListAsynJobsRequest) (_result *ListAsynJobsResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListAsynJobsResponse{}
	_body, _err := client.ListAsynJobsWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListListenerCertificatesWithOptions(request *ListListenerCertificatesRequest, runtime *dara.RuntimeOptions) (_result *ListListenerCertificatesResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListListenerCertificatesResponse
func (client *Client) ListListenerCertificates(request *ListListenerCertificatesRequest) (_result *ListListenerCertificatesResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListListenerCertificatesResponse{}
	_body, _err := client.ListListenerCertificatesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListListenersWithOptions(request *ListListenersRequest, runtime *dara.RuntimeOptions) (_result *ListListenersResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListListenersResponse
func (client *Client) ListListeners(request *ListListenersRequest) (_result *ListListenersResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListListenersResponse{}
	_body, _err := client.ListListenersWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListLoadBalancersWithOptions(request *ListLoadBalancersRequest, runtime *dara.RuntimeOptions) (_result *ListLoadBalancersResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListLoadBalancersResponse
func (client *Client) ListLoadBalancers(request *ListLoadBalancersRequest) (_result *ListLoadBalancersResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListLoadBalancersResponse{}
	_body, _err := client.ListLoadBalancersWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListSecurityPolicyWithOptions(request *ListSecurityPolicyRequest, runtime *dara.RuntimeOptions) (_result *ListSecurityPolicyResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListSecurityPolicyResponse
func (client *Client) ListSecurityPolicy(request *ListSecurityPolicyRequest) (_result *ListSecurityPolicyResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListSecurityPolicyResponse{}
	_body, _err := client.ListSecurityPolicyWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListServerGroupServersWithOptions(request *ListServerGroupServersRequest, runtime *dara.RuntimeOptions) (_result *ListServerGroupServersResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListServerGroupServersResponse
func (client *Client) ListServerGroupServers(request *ListServerGroupServersRequest) (_result *ListServerGroupServersResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListServerGroupServersResponse{}
	_body, _err := client.ListServerGroupServersWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListServerGroupsWithOptions(request *ListServerGroupsRequest, runtime *dara.RuntimeOptions) (_result *ListServerGroupsResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListServerGroupsResponse
func (client *Client) ListServerGroups(request *ListServerGroupsRequest) (_result *ListServerGroupsResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListServerGroupsResponse{}
	_body, _err := client.ListServerGroupsWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListSystemSecurityPolicyWithOptions(request *ListSystemSecurityPolicyRequest, runtime *dara.RuntimeOptions) (_result *ListSystemSecurityPolicyResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListSystemSecurityPolicyResponse
func (client *Client) ListSystemSecurityPolicy(request *ListSystemSecurityPolicyRequest) (_result *ListSystemSecurityPolicyResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListSystemSecurityPolicyResponse{}
	_body, _err := client.ListSystemSecurityPolicyWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) ListTagResourcesWithOptions(request *ListTagResourcesRequest, runtime *dara.RuntimeOptions) (_result *ListTagResourcesResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return ListTagResourcesResponse
func (client *Client) ListTagResources(request *ListTagResourcesRequest) (_result *ListTagResourcesResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &ListTagResourcesResponse{}
	_body, _err := client.ListTagResourcesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) LoadBalancerJoinSecurityGroupWithOptions(request *LoadBalancerJoinSecurityGroupRequest, runtime *dara.RuntimeOptions) (_result *LoadBalancerJoinSecurityGroupResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return LoadBalancerJoinSecurityGroupResponse
func (client *Client) LoadBalancerJoinSecurityGroup(request *LoadBalancerJoinSecurityGroupRequest) (_result *LoadBalancerJoinSecurityGroupResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &LoadBalancerJoinSecurityGroupResponse{}
	_body, _err := client.LoadBalancerJoinSecurityGroupWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) LoadBalancerLeaveSecurityGroupWithOptions(request *LoadBalancerLeaveSecurityGroupRequest, runtime *dara.RuntimeOptions) (_result *LoadBalancerLeaveSecurityGroupResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return LoadBalancerLeaveSecurityGroupResponse
func (client *Client) LoadBalancerLeaveSecurityGroup(request *LoadBalancerLeaveSecurityGroupRequest) (_result *LoadBalancerLeaveSecurityGroupResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &LoadBalancerLeaveSecurityGroupResponse{}
	_body, _err := client.LoadBalancerLeaveSecurityGroupWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) MoveResourceGroupWithOptions(request *MoveResourceGroupRequest, runtime *dara.RuntimeOptions) (_result *MoveResourceGroupResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return MoveResourceGroupResponse
func (client *Client) MoveResourceGroup(request *MoveResourceGroupRequest) (_result *MoveResourceGroupResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &MoveResourceGroupResponse{}
	_body, _err := client.MoveResourceGroupWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) RemoveServersFromServerGroupWithOptions(request *RemoveServersFromServerGroupRequest, runtime *dara.RuntimeOptions) (_result *RemoveServersFromServerGroupResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return RemoveServersFromServerGroupResponse
func (client *Client) RemoveServersFromServerGroup(request *RemoveServersFromServerGroupRequest) (_result *RemoveServersFromServerGroupResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &RemoveServersFromServerGroupResponse{}
	_body, _err := client.RemoveServersFromServerGroupWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) SetHdMonitorRegionConfigWithOptions(request *SetHdMonitorRegionConfigRequest, runtime *dara.RuntimeOptions) (_result *SetHdMonitorRegionConfigResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return SetHdMonitorRegionConfigResponse
func (client *Client) SetHdMonitorRegionConfig(request *SetHdMonitorRegionConfigRequest) (_result *SetHdMonitorRegionConfigResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &SetHdMonitorRegionConfigResponse{}
	_body, _err := client.SetHdMonitorRegionConfigWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) StartListenerWithOptions(request *StartListenerRequest, runtime *dara.RuntimeOptions) (_result *StartListenerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return StartListenerResponse
func (client *Client) StartListener(request *StartListenerRequest) (_result *StartListenerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &StartListenerResponse{}
	_body, _err := client.StartListenerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) StartShiftLoadBalancerZonesWithOptions(request *StartShiftLoadBalancerZonesRequest, runtime *dara.RuntimeOptions) (_result *StartShiftLoadBalancerZonesResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return StartShiftLoadBalancerZonesResponse
func (client *Client) StartShiftLoadBalancerZones(request *StartShiftLoadBalancerZonesRequest) (_result *StartShiftLoadBalancerZonesResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &StartShiftLoadBalancerZonesResponse{}
	_body, _err := client.StartShiftLoadBalancerZonesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) StopListenerWithOptions(request *StopListenerRequest, runtime *dara.RuntimeOptions) (_result *StopListenerResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return StopListenerResponse
func (client *Client) StopListener(request *StopListenerRequest) (_result *StopListenerResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &StopListenerResponse{}
	_body, _err := client.StopListenerWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) TagResourcesWithOptions(request *TagResourcesRequest, runtime *dara.RuntimeOptions) (_result *TagResourcesResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return TagResourcesResponse
func (client *Client) TagResources(request *TagResourcesRequest) (_result *TagResourcesResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &TagResourcesResponse{}
	_body, _err := client.TagResourcesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UntagResourcesWithOptions(request *UntagResourcesRequest, runtime *dara.RuntimeOptions) (_result *UntagResourcesResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return UntagResourcesResponse
func (client *Client) UntagResources(request *UntagResourcesRequest) (_result *UntagResourcesResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UntagResourcesResponse{}
	_body, _err := client.UntagResourcesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UpdateListenerAttributeWithOptions(tmpReq *UpdateListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateListenerAttributeResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @param request - UpdateListenerAttributeRequest
//
// @return UpdateListenerAttributeResponse
func (client *Client) UpdateListenerAttribute(request *UpdateListenerAttributeRequest) (_result *UpdateListenerAttributeResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UpdateListenerAttributeResponse{}
	_body, _err := client.UpdateListenerAttributeWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UpdateLoadBalancerAddressTypeConfigWithOptions(request *UpdateLoadBalancerAddressTypeConfigRequest, runtime *dara.RuntimeOptions) (_result *UpdateLoadBalancerAddressTypeConfigResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return UpdateLoadBalancerAddressTypeConfigResponse
func (client *Client) UpdateLoadBalancerAddressTypeConfig(request *UpdateLoadBalancerAddressTypeConfigRequest) (_result *UpdateLoadBalancerAddressTypeConfigResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UpdateLoadBalancerAddressTypeConfigResponse{}
	_body, _err := client.UpdateLoadBalancerAddressTypeConfigWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UpdateLoadBalancerAttributeWithOptions(request *UpdateLoadBalancerAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateLoadBalancerAttributeResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return UpdateLoadBalancerAttributeResponse
func (client *Client) UpdateLoadBalancerAttribute(request *UpdateLoadBalancerAttributeRequest) (_result *UpdateLoadBalancerAttributeResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UpdateLoadBalancerAttributeResponse{}
	_body, _err := client.UpdateLoadBalancerAttributeWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UpdateLoadBalancerProtectionWithOptions(request *UpdateLoadBalancerProtectionRequest, runtime *dara.RuntimeOptions) (_result *UpdateLoadBalancerProtectionResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return UpdateLoadBalancerProtectionResponse
func (client *Client) UpdateLoadBalancerProtection(request *UpdateLoadBalancerProtectionRequest) (_result *UpdateLoadBalancerProtectionResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UpdateLoadBalancerProtectionResponse{}
	_body, _err := client.UpdateLoadBalancerProtectionWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UpdateLoadBalancerZonesWithOptions(request *UpdateLoadBalancerZonesRequest, runtime *dara.RuntimeOptions) (_result *UpdateLoadBalancerZonesResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return UpdateLoadBalancerZonesResponse
func (client *Client) UpdateLoadBalancerZones(request *UpdateLoadBalancerZonesRequest) (_result *UpdateLoadBalancerZonesResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UpdateLoadBalancerZonesResponse{}
	_body, _err := client.UpdateLoadBalancerZonesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UpdateSecurityPolicyAttributeWithOptions(request *UpdateSecurityPolicyAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateSecurityPolicyAttributeResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return UpdateSecurityPolicyAttributeResponse
func (client *Client) UpdateSecurityPolicyAttribute(request *UpdateSecurityPolicyAttributeRequest) (_result *UpdateSecurityPolicyAttributeResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UpdateSecurityPolicyAttributeResponse{}
	_body, _err := client.UpdateSecurityPolicyAttributeWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UpdateServerGroupAttributeWithOptions(request *UpdateServerGroupAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateServerGroupAttributeResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return UpdateServerGroupAttributeResponse
func (client *Client) UpdateServerGroupAttribute(request *UpdateServerGroupAttributeRequest) (_result *UpdateServerGroupAttributeResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UpdateServerGroupAttributeResponse{}
	_body, _err := client.UpdateServerGroupAttributeWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
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
func (client *Client) UpdateServerGroupServersAttributeWithOptions(request *UpdateServerGroupServersAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateServerGroupServersAttributeResponse, _err error) {
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
	_body, _err := client.CallApi(params, req, runtime)
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
// @return UpdateServerGroupServersAttributeResponse
func (client *Client) UpdateServerGroupServersAttribute(request *UpdateServerGroupServersAttributeRequest) (_result *UpdateServerGroupServersAttributeResponse, _err error) {
	runtime := &dara.RuntimeOptions{}
	_result = &UpdateServerGroupServersAttributeResponse{}
	_body, _err := client.UpdateServerGroupServersAttributeWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}
