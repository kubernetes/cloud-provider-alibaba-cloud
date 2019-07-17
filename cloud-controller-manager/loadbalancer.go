/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alicloud

import (
	"errors"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"os"
	"reflect"
	"strings"

	"encoding/json"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

// AnnotationRequest annotated parameters.
type AnnotationRequest struct {
	Loadbalancerid string
	BackendLabel   string

	SSLPorts       string
	AddressType    slb.AddressType
	AclStatus      string
	AclID          string
	AclType        string
	VswitchID      string
	ForwardPort    string
	SLBNetworkType string

	ChargeType slb.InternetChargeType
	//Region     		common.Region
	Bandwidth int
	CertID    string

	MasterZoneID string
	SlaveZoneID  string

	HealthCheck            slb.FlagType
	HealthCheckURI         string
	HealthCheckConnectPort int
	HealthyThreshold       int
	UnhealthyThreshold     int
	HealthCheckInterval    int
	HealthCheckDomain      string
	HealthCheckHttpCode    slb.HealthCheckHttpCodeType

	HealthCheckConnectTimeout int                 // for tcp
	HealthCheckType           slb.HealthCheckType // for tcp, Type could be http tcp
	HealthCheckTimeout        int                 // for https and http

	LoadBalancerSpec slb.LoadBalancerSpecType
	Scheduler        string

	StickySession      slb.FlagType
	StickySessionType  slb.StickySessionType
	Cookie             string
	CookieTimeout      int
	PersistenceTimeout int
	AddressIPVersion   slb.AddressIPVersionType

	OverrideListeners string

	PrivateZoneName       string
	PrivateZoneId         string
	PrivateZoneRecordName string
	PrivateZoneRecordTTL  int
}

// TAGKEY Default tag key.
const TAGKEY = "kubernetes.do.not.delete"

// ClientSLBSDK client sdk for slb
type ClientSLBSDK interface {
	DescribeLoadBalancers(args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error)
	CreateLoadBalancer(args *slb.CreateLoadBalancerArgs) (response *slb.CreateLoadBalancerResponse, err error)
	DeleteLoadBalancer(loadBalancerId string) (err error)
	ModifyLoadBalancerInstanceSpec(args *slb.ModifyLoadBalancerInstanceSpecArgs) (err error)
	ModifyLoadBalancerInternetSpec(args *slb.ModifyLoadBalancerInternetSpecArgs) (err error)
	DescribeLoadBalancerAttribute(loadBalancerId string) (loadBalancer *slb.LoadBalancerType, err error)
	RemoveBackendServers(loadBalancerId string, backendServers []string) (result []slb.BackendServerType, err error)
	AddBackendServers(loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error)

	StopLoadBalancerListener(loadBalancerId string, port int) (err error)
	StartLoadBalancerListener(loadBalancerId string, port int) (err error)
	CreateLoadBalancerTCPListener(args *slb.CreateLoadBalancerTCPListenerArgs) (err error)
	CreateLoadBalancerUDPListener(args *slb.CreateLoadBalancerUDPListenerArgs) (err error)
	DeleteLoadBalancerListener(loadBalancerId string, port int) (err error)
	CreateLoadBalancerHTTPSListener(args *slb.CreateLoadBalancerHTTPSListenerArgs) (err error)
	CreateLoadBalancerHTTPListener(args *slb.CreateLoadBalancerHTTPListenerArgs) (err error)
	DescribeLoadBalancerHTTPSListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error)
	DescribeLoadBalancerTCPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error)
	DescribeLoadBalancerUDPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerUDPListenerAttributeResponse, err error)
	DescribeLoadBalancerHTTPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPListenerAttributeResponse, err error)

	SetLoadBalancerHTTPListenerAttribute(args *slb.SetLoadBalancerHTTPListenerAttributeArgs) (err error)
	SetLoadBalancerHTTPSListenerAttribute(args *slb.SetLoadBalancerHTTPSListenerAttributeArgs) (err error)
	SetLoadBalancerTCPListenerAttribute(args *slb.SetLoadBalancerTCPListenerAttributeArgs) (err error)
	SetLoadBalancerUDPListenerAttribute(args *slb.SetLoadBalancerUDPListenerAttributeArgs) (err error)

	RemoveTags(args *slb.RemoveTagsArgs) error
	DescribeTags(args *slb.DescribeTagsArgs) (tags []slb.TagItemType, pagination *common.PaginationResult, err error)
	AddTags(args *slb.AddTagsArgs) error

	CreateVServerGroup(args *slb.CreateVServerGroupArgs) (response *slb.CreateVServerGroupResponse, err error)
	DescribeVServerGroups(args *slb.DescribeVServerGroupsArgs) (response *slb.DescribeVServerGroupsResponse, err error)
	DeleteVServerGroup(args *slb.DeleteVServerGroupArgs) (response *slb.DeleteVServerGroupResponse, err error)
	SetVServerGroupAttribute(args *slb.SetVServerGroupAttributeArgs) (response *slb.SetVServerGroupAttributeResponse, err error)
	DescribeVServerGroupAttribute(args *slb.DescribeVServerGroupAttributeArgs) (response *slb.DescribeVServerGroupAttributeResponse, err error)
	ModifyVServerGroupBackendServers(args *slb.ModifyVServerGroupBackendServersArgs) (response *slb.ModifyVServerGroupBackendServersResponse, err error)
	AddVServerGroupBackendServers(args *slb.AddVServerGroupBackendServersArgs) (response *slb.AddVServerGroupBackendServersResponse, err error)
	RemoveVServerGroupBackendServers(args *slb.RemoveVServerGroupBackendServersArgs) (response *slb.RemoveVServerGroupBackendServersResponse, err error)
}

// LoadBalancerClient slb client wrapper
type LoadBalancerClient struct {
	region string
	vpcid  string
	c      ClientSLBSDK
	// known service resource version
	ins ClientInstanceSDK
}

func (s *LoadBalancerClient) findLoadBalancer(service *v1.Service) (bool, *slb.LoadBalancerType, error) {
	def, _ := ExtractAnnotationRequest(service)

	//loadbalancer := service.Status.LoadBalancer
	//glog.V(4).Infof("alicloud: find loadbalancer [%v] with user defined annotations [%v]\n",loadbalancer,def)
	//if len(loadbalancer.Ingress) > 0 {
	//	if lbid := fromLoadBalancerStatus(service); lbid != "" {
	//		if def.Loadbalancerid != "" && def.Loadbalancerid != lbid {
	//			glog.Errorf("alicloud: changing loadbalancer id was not allowed after loadbalancer"+
	//				" was already created ! please remove service annotation: [%s]\n", ServiceAnnotationLoadBalancerId)
	//			return false, nil, fmt.Errorf("alicloud: change loadbalancer id after service " +
	//				"has been created is not supported. remove service annotation[%s] and retry!", ServiceAnnotationLoadBalancerId)
	//		}
	//		// found loadbalancer id in service ingress status.
	//		// this id was set previously when loadbalancer was created.
	//		return s.findLoadBalancerByID(lbid)
	//	}
	//}

	// User assigned lobadbalancer id go first.
	if def.Loadbalancerid != "" {
		return s.findLoadBalancerByID(def.Loadbalancerid)
	}
	// if not, find by slb tags
	return s.findLoadBalancerByTags(service)
}

func (s *LoadBalancerClient) findLoadBalancerByID(lbid string) (bool, *slb.LoadBalancerType, error) {

	lbs, err := s.c.DescribeLoadBalancers(
		&slb.DescribeLoadBalancersArgs{
			RegionId:       DEFAULT_REGION,
			LoadBalancerId: lbid,
		},
	)
	glog.Infof("find loadbalancer with id [%s], %d found.", lbid, len(lbs))
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		return false, nil, nil
	}
	if len(lbs) > 1 {
		glog.Warningf("multiple loadbalancer returned with id [%s], using the first one with IP=%s", lbid, lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(lbs[0].LoadBalancerId)
	return err == nil, lb, err
}

func (s *LoadBalancerClient) findLoadBalancerByTags(service *v1.Service) (bool, *slb.LoadBalancerType, error) {
	if service.UID == "" {
		return false, nil, fmt.Errorf("unexpected empty service uid")
	}
	lbn := cloudprovider.GetLoadBalancerName(service)
	items, err := json.Marshal(
		[]slb.TagItem{
			{
				TagKey:   TAGKEY,
				TagValue: lbn,
			},
		},
	)
	if err != nil {
		return false, nil, err
	}
	lbs, err := s.c.DescribeLoadBalancers(
		&slb.DescribeLoadBalancersArgs{
			RegionId: DEFAULT_REGION,
			Tags:     string(items),
		},
	)
	utils.Logf(service, "alicloud: fallback to find loadbalancer by tags [%s]", string(items))
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		// here we need to fallback on finding by name for compatible reason
		// the old service slb may not have a tag.
		return s.findLoadBalancerByName(lbn)
	}
	if len(lbs) > 1 {
		utils.Logf(service, "Warning: multiple loadbalancer returned with tags [%s], "+
			"using the first one with IP=%s", string(items), lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(lbs[0].LoadBalancerId)
	return err == nil, lb, err
}

func (s *LoadBalancerClient) findLoadBalancerByName(name string) (bool, *slb.LoadBalancerType, error) {
	lbs, err := s.c.DescribeLoadBalancers(
		&slb.DescribeLoadBalancersArgs{
			RegionId:         DEFAULT_REGION,
			LoadBalancerName: name,
		},
	)
	glog.V(2).Infof("fallback to find loadbalancer by name [%s]", name)
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		return false, nil, nil
	}
	if len(lbs) > 1 {
		glog.Warningf("alicloud: multiple loadbalancer returned with name [%s], "+
			"using the first one with IP=%s", name, lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(lbs[0].LoadBalancerId)
	return err == nil, lb, err
}

// getLoadBalancerAdditionalTags converts the comma separated list of key-value
// pairs in the ServiceAnnotationLoadBalancerAdditionalTags annotation and returns
// it as a map.
func getLoadBalancerAdditionalTags(annotations map[string]string) map[string]string {
	additionalTags := make(map[string]string)
	if additionalTagsList, ok := annotations[ServiceAnnotationLoadBalancerAdditionalTags]; ok {
		additionalTagsList = strings.TrimSpace(additionalTagsList)

		// Break up list of "Key1=Val,Key2=Val2"
		tagList := strings.Split(additionalTagsList, ",")

		// Break up "Key=Val"
		for _, tagSet := range tagList {
			tag := strings.Split(strings.TrimSpace(tagSet), "=")

			// Accept "Key=val" or "Key=" or just "Key"
			if len(tag) >= 2 && len(tag[0]) != 0 {
				// There is a key and a value, so save it
				additionalTags[tag[0]] = tag[1]
			} else if len(tag) == 1 && len(tag[0]) != 0 {
				// Just "Key"
				additionalTags[tag[0]] = ""
			}
		}
	}

	return additionalTags
}

func equalsAddressIPVersion(request, origined slb.AddressIPVersionType) bool {
	if request == "" {
		request = slb.IPv4
	}

	if origined == "" {
		origined = slb.IPv4
	}
	return request == origined
}

// EnsureLoadBalancer make sure slb is reconciled nodes []*v1.Node
func (s *LoadBalancerClient) EnsureLoadBalancer(service *v1.Service, nodes interface{}, vswitchid string) (*slb.LoadBalancerType, error) {
	glog.V(4).Infof("alicloud: ensure loadbalancer with service details, \n%+v", PrettyJson(service))

	exists, origined, err := s.findLoadBalancer(service)
	if err != nil {
		return nil, err
	}
	glog.V(4).Infof("alicloud: find "+
		"loadbalancer with result, exist=%v\n %s\n", exists, PrettyJson(origined))
	_, request := ExtractAnnotationRequest(service)

	var derr error
	// this is a workaround for issue: https://github.com/kubernetes/kubernetes/issues/59084
	if !exists {
		// If need created, double check if the resource id has been deleted
		if isServiceDeleted(service) {
			glog.V(2).Infof("alicloud: isServiceDeleted report that this service has been " +
				"deleted before. see issue: https://github.com/kubernetes/kubernetes/issues/59084")
			os.Exit(1)
		}
		if isLoadbalancerOwnIngress(service) {
			return nil, fmt.Errorf("alicloud: not able to find loadbalancer "+
				"named [%s] in openapi, but it's defined in service.loaderbalancer.ingress. "+
				"this may happen when you removed loadbalancerid annotation.\n", service.Name)
		}
		if request.Loadbalancerid != "" {
			return nil, fmt.Errorf("alicloud: user specified "+
				"loadbalancer[%s] does not exist. pls check", request.Loadbalancerid)
		}

		// From here, we need to create a new loadbalancer
		glog.V(5).Infof("alicloud: can not find a "+
			"loadbalancer with service name [%s/%s], creating a new one\n", service.Namespace, service.Name)
		opts := s.getLoadBalancerOpts(service, vswitchid)
		lbr, err := s.c.CreateLoadBalancer(opts)
		if err != nil {
			return nil, err
		}

		//deal with loadBalancer tags
		tags := getLoadBalancerAdditionalTags(getBackwardsCompatibleAnnotation(service.Annotations))
		loadbalancerName := cloudprovider.GetLoadBalancerName(service)
		// Add default tags
		tags[TAGKEY] = loadbalancerName

		tagItemArr := make([]slb.TagItem, 0)
		for key, value := range tags {
			tagItemArr = append(tagItemArr, slb.TagItem{TagKey: key, TagValue: value})
		}
		// TODO if the key size or value size > slb's tag limit or number of tags  >limit(20), tags will insert error.

		tagItems, err := json.Marshal(tagItemArr)

		if err != nil {
			return nil, err
		}
		if err := s.c.AddTags(&slb.AddTagsArgs{
			RegionId:       opts.RegionId,
			LoadBalancerID: lbr.LoadBalancerId,
			Tags:           string(tagItems),
		}); err != nil {
			return nil, err
		}

		origined, derr = s.c.DescribeLoadBalancerAttribute(lbr.LoadBalancerId)
	} else {
		// Need to verify loadbalancer.
		// Reuse SLB is not allowed when the SLB is created by k8s service.
		tags, _, err := s.c.DescribeTags(
			&slb.DescribeTagsArgs{
				RegionId:       origined.RegionId,
				LoadBalancerID: origined.LoadBalancerId,
			})
		if err != nil {
			return origined, err
		}
		if isLoadBalancerCreatedByKubernetes(tags) &&
			isUserDefinedLoadBalancer(service) {
			return origined, fmt.Errorf("alicloud: can not reuse loadbalancer created by kubernetes. %s", origined.LoadBalancerId)
		}
		needUpdate, charge, bandwidth := false, origined.InternetChargeType, origined.Bandwidth
		glog.V(5).Infof("alicloud: found "+
			"an exist loadbalancer[%s], check to see whether update is needed.", origined.LoadBalancerId)
		if request.MasterZoneID != "" && request.MasterZoneID != origined.MasterZoneId {
			return nil, fmt.Errorf("alicloud: can not change LoadBalancer master zone id once created")
		}
		if request.SlaveZoneID != "" && request.SlaveZoneID != origined.SlaveZoneId {
			return nil, fmt.Errorf("alicloud: can not change LoadBalancer slave zone id once created")
		}
		if !equalsAddressIPVersion(request.AddressIPVersion, origined.AddressIPVersion) {
			return nil, fmt.Errorf("alicloud: can not change LoadBalancer AddressIPVersion once created")
		}
		if request.ChargeType != "" && request.ChargeType != origined.InternetChargeType {
			needUpdate = true
			charge = request.ChargeType
			// Todo: here we need to compare loadbalance
			utils.Logf(service, "internet charge type([%s] -> [%s]), update loadbalancer [%s]\n",
				string(origined.InternetChargeType), string(request.ChargeType), origined.LoadBalancerName)
		}
		if request.ChargeType == slb.PayByBandwidth &&
			request.Bandwidth != DEFAULT_BANDWIDTH &&
			request.Bandwidth != origined.Bandwidth {

			needUpdate = true
			bandwidth = request.Bandwidth
			utils.Logf(service, "bandwidth changed from ([%d] -> [%d]), update [%s]", origined.Bandwidth, request.Bandwidth, origined.LoadBalancerId)
		}
		if request.AddressType != "" && request.AddressType != origined.AddressType {
			glog.Errorf("alicloud: warning! can not change "+
				"loadbalancer address type after it has been created! please "+
				"recreate the service.[%s]->[%s],[%s]\n", origined.AddressType, request.AddressType, origined.LoadBalancerName)
			return nil, errors.New("alicloud: change loadbalancer " +
				"address type after service has been created is not supported. delete and retry")
		}
		if needUpdate {
			if origined.AddressType == "internet" {
				utils.Logf(service, "modify loadbalancer: chargetype=%s, bandwidth=%d", charge, bandwidth)
				if err := s.c.ModifyLoadBalancerInternetSpec(
					&slb.ModifyLoadBalancerInternetSpecArgs{
						LoadBalancerId:     origined.LoadBalancerId,
						InternetChargeType: charge,
						Bandwidth:          bandwidth,
					}); err != nil {
					return nil, err
				}
			} else {
				utils.Logf(service, "only internet loadbalancer is allowed to modify bandwidth and pay type")
			}
		}

		// update instance spec
		if request.LoadBalancerSpec != "" && request.LoadBalancerSpec != origined.LoadBalancerSpec {
			glog.Infof("alicloud: loadbalancerspec([%s] -> [%s]) changed, update loadbalancer [%s]\n",
				origined.LoadBalancerSpec, request.LoadBalancerSpec, origined.LoadBalancerName)
			if err := s.c.ModifyLoadBalancerInstanceSpec(&slb.ModifyLoadBalancerInstanceSpecArgs{
				RegionId:         origined.RegionId,
				LoadBalancerId:   origined.LoadBalancerId,
				LoadBalancerSpec: request.LoadBalancerSpec,
			}); err != nil {
				return nil, err
			}
		}
		origined, derr = s.c.DescribeLoadBalancerAttribute(origined.LoadBalancerId)
	}
	if derr != nil {
		glog.Errorf("alicloud: can not get loadbalancer[%s] attribute. ", origined.LoadBalancerId)
		return nil, err
	}
	vgs := BuildVirturalGroupFromService(s, service, origined)

	// Make sure virtual server backend group has been updated.
	if err := EnsureVirtualGroups(vgs, nodes); err != nil {
		return origined, fmt.Errorf("update backend servers: error %s", err.Error())
	}
	// Apply listener when
	//   1. user does not assign loadbalancer id by themselves.
	//   2. force-override-listener annotation is set.
	if (!isUserDefinedLoadBalancer(service)) ||
		(isUserDefinedLoadBalancer(service) && isOverrideListeners(service)) {
		utils.Logf(service, "not user defined loadbalancer[%s], start to apply listener.\n", origined.LoadBalancerId)
		// If listener update is needed. Switch to vserver group immediately.
		// No longer update default backend servers.
		if err := EnsureListeners(s, service, origined, vgs); err != nil {

			return origined, fmt.Errorf("ensure listener error: %s", err.Error())
		}
	}
	return origined, s.UpdateLoadBalancer(service, nodes, false)
}

func isLoadBalancerCreatedByKubernetes(tags []slb.TagItemType) bool {
	for _, tag := range tags {
		if tag.TagKey == TAGKEY {
			return true
		}
	}
	return false
}

//UpdateLoadBalancer make sure slb backend is reconciled
func (s *LoadBalancerClient) UpdateLoadBalancer(service *v1.Service, nodes interface{}, withVgroup bool) error {

	exists, lb, err := s.findLoadBalancer(service)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the loadbalance you specified by name [%s] does not exist", service.Name)
	}
	if withVgroup {
		vgs := BuildVirturalGroupFromService(s, service, lb)
		if err := EnsureVirtualGroups(vgs, nodes); err != nil {
			return fmt.Errorf("update backend servers: error %s", err.Error())
		}
	}
	if !needUpdateDefaultBackend(service, lb) {
		return nil
	}
	utils.Logf(service, "update default backend server group")
	return s.UpdateDefaultServerGroup(nodes, lb)
}

// needUpdateDefaultBackend when listeners have legacy listener.
// which means legacy listener
// we do not update default backend server group when listener enables virtual server group mode.
func needUpdateDefaultBackend(svc *v1.Service, lb *slb.LoadBalancerType) bool {
	listeners := lb.ListenerPortsAndProtocol.ListenerPortAndProtocol

	// legacy listener has a empty name or '-'
	for _, lis := range listeners {
		if lis.Description == "" ||
			lis.Description == "-" {
			if !hasPort(svc, int32(lis.ListenerPort)) {
				continue
			}
			utils.Logf(svc, "port %d has legacy listener, "+
				"apply default backend server group. [%s]", lis.ListenerPort, lis.Description)
			return true
		}
	}
	return false
}

func hasPort(svc *v1.Service, port int32) bool {
	for _, p := range svc.Spec.Ports {
		if p.Port == port {
			return true
		}
	}
	return false
}

// EnsureLoadBalanceDeleted make sure slb is deleted
func (s *LoadBalancerClient) EnsureLoadBalanceDeleted(service *v1.Service) error {
	// need to save the resource version when deleted event
	err := keepResourceVesion(service)
	if err != nil {
		utils.Logf(service, "Warning: failed to save deleted service resourceVersion,due to [%s] ", err.Error())
	}
	exists, lb, err := s.findLoadBalancer(service)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	// skip delete user defined loadbalancer
	if isUserDefinedLoadBalancer(service) {
		utils.Logf(service, "user managed loadbalancer will not be deleted by cloudprovider.")
		return EnsureListenersDeleted(s.c, service, lb, BuildVirturalGroupFromService(s, service, lb))
	}

	return s.c.DeleteLoadBalancer(lb.LoadBalancerId)
}

func (s *LoadBalancerClient) getLoadBalancerOpts(service *v1.Service, vswitchid string) (args *slb.CreateLoadBalancerArgs) {
	ar, req := ExtractAnnotationRequest(service)
	args = &slb.CreateLoadBalancerArgs{
		AddressType:        ar.AddressType,
		InternetChargeType: ar.ChargeType,
		RegionId:           DEFAULT_REGION,
		LoadBalancerSpec:   ar.LoadBalancerSpec,
		MasterZoneId:       ar.MasterZoneID,
		SlaveZoneId:        ar.SlaveZoneID,
		AddressIPVersion:   ar.AddressIPVersion,
	}
	// paybybandwidth need a default bandwidth args, while paybytraffic doesnt.
	if ar.ChargeType == slb.PayByBandwidth ||
		(ar.ChargeType == slb.PayByTraffic && req.Bandwidth != 0) {
		glog.V(5).Infof("alicloud: %s, set bandwidth to %d", ar.ChargeType, ar.Bandwidth)
		args.Bandwidth = ar.Bandwidth
	}
	if ar.SLBNetworkType != "classic" &&
		strings.Compare(string(ar.AddressType), string(slb.IntranetAddressType)) == 0 {

		utils.Logf(service, "intranet vpc "+
			"loadbalancer will be created. address type=%s, switchid=%s", ar.AddressType, vswitchid)
		args.VSwitchId = vswitchid
	}
	args.LoadBalancerName = cloudprovider.GetLoadBalancerName(service)
	return
}

// DEFAULT_SERVER_WEIGHT default server weight
const DEFAULT_SERVER_WEIGHT = 100

// UpdateDefaultServerGroup update default server group
func (s *LoadBalancerClient) UpdateDefaultServerGroup(backends interface{}, lb *slb.LoadBalancerType) error {
	nodes, ok := backends.([]*v1.Node)
	if !ok {
		glog.Infof("skip default server group update for type %s", reflect.TypeOf(backends))
		return nil
	}
	additions, deletions := []slb.BackendServerType{}, []string{}
	glog.V(5).Infof("alicloud: try to update loadbalancer backend servers. [%s]\n", lb.LoadBalancerId)
	// checkout for newly added servers
	for _, n1 := range nodes {
		found := false
		_, id, err := nodeFromProviderID(n1.Spec.ProviderID)
		for _, n2 := range lb.BackendServers.BackendServer {
			if err != nil {
				glog.Errorf("alicloud: node providerid=%s is not"+
					" in the correct form, expect regionid.instanceid. skip add op", n1.Spec.ProviderID)
				continue
			}
			if id == n2.ServerId {
				found = true
				break
			}
		}
		if !found {
			additions = append(additions, slb.BackendServerType{ServerId: string(id), Weight: DEFAULT_SERVER_WEIGHT, Type: "ecs"})
		}
	}
	if len(additions) > 0 {
		glog.V(5).Infof("alicloud: add loadbalancer backend for [%s] \n %s\n", lb.LoadBalancerId, PrettyJson(additions))
		// only 20 backend servers is accepted per delete.
		for len(additions) > 0 {
			var target []slb.BackendServerType
			if len(additions) > MAX_LOADBALANCER_BACKEND {
				target = additions[0:MAX_LOADBALANCER_BACKEND]
				additions = additions[MAX_LOADBALANCER_BACKEND:]
				glog.V(5).Infof("alicloud: batch add backend servers, %v", target)
			} else {
				target = additions
				additions = []slb.BackendServerType{}
				glog.V(5).Infof("alicloud: batch add backend servers, else %v", target)
			}
			if _, err := s.c.AddBackendServers(lb.LoadBalancerId, target); err != nil {
				return err
			}
		}
	}

	// check for removed backend servers
	for _, n1 := range lb.BackendServers.BackendServer {
		found := false
		for _, n2 := range nodes {
			_, id, err := nodeFromProviderID(n2.Spec.ProviderID)
			if err != nil {
				glog.Errorf("alicloud: node providerid=%s is not "+
					"in the correct form, expect regionid.instanceid.. skip delete op... [%s]", n2.Spec.ProviderID, err.Error())
				continue
			}
			if n1.ServerId == id {
				found = true
				break
			}
		}
		if !found {
			deletions = append(deletions, n1.ServerId)
		}
	}
	if len(deletions) > 0 {
		glog.V(5).Infof("alicloud: delete loadbalancer backend for [%s] %v\n", lb.LoadBalancerId, deletions)
		// only 20 backend servers is accepted per delete.
		for len(deletions) > 0 {
			var target []string
			if len(deletions) > MAX_LOADBALANCER_BACKEND {
				target = deletions[0:MAX_LOADBALANCER_BACKEND]
				deletions = deletions[MAX_LOADBALANCER_BACKEND:]
				glog.V(5).Infof("alicloud: batch delete backend servers, %s", target)
			} else {
				target = deletions
				deletions = []string{}
				glog.V(5).Infof("alicloud: batch delete backend servers else, %s", target)
			}
			if _, err := s.c.RemoveBackendServers(lb.LoadBalancerId, target); err != nil {
				return err
			}
		}
	}
	if len(additions) <= 0 && len(deletions) <= 0 {
		glog.V(5).Infof("alicloud: no backend servers need to be updated.[%s]\n", lb.LoadBalancerId)
		return nil
	}
	glog.V(5).Infof("alicloud: update loadbalancer`s backend servers finished! [%s]\n", lb.LoadBalancerId)
	return nil
}

// check to see if user has assigned any loadbalancer
func isUserDefinedLoadBalancer(svc *v1.Service) bool {
	return serviceAnnotation(svc, ServiceAnnotationLoadBalancerId) != ""
}

func isOverrideListeners(svc *v1.Service) bool {
	return strings.ToLower(serviceAnnotation(svc, ServiceAnnotationLoadBalancerOverrideListener)) == "true"
}

// check if the service exists in service definition
func isLoadbalancerOwnIngress(service *v1.Service) bool {
	if service == nil ||
		len(service.Status.LoadBalancer.Ingress) == 0 {
		utils.Logf(service, "service %s doesn't have ingresses", service.Name)
		return false
	}
	utils.Logf(service, "service %s has ingresses=%v", service.Name, service.Status.LoadBalancer.Ingress)
	return true
}

// ensure service resource version properly and update last known resource
// version to the largest one, for now only keep create and delete behavior
func isServiceDeleted(service *v1.Service) bool {
	//if service == nil {
	//	return fmt.Errorf("alicloud:service is nil")
	//}
	serviceUID := string(service.GetUID())
	keeper := GetLocalService()
	deleted := keeper.get(serviceUID)
	if deleted {
		utils.Logf(service, "service "+
			"%s's uid %v has been deleted, shouldn't be created again.", service.Name, serviceUID)
		return true
	}
	utils.Logf(service, "service %s's uid %v "+
		"hasn't been deleted, first time to process, as expected.", service.Name, serviceUID)
	return false
}

// save the deleted service's uid
func keepResourceVesion(service *v1.Service) error {
	if service == nil {
		return fmt.Errorf("alicloud: failed to save deleted service resource version , for service is nil")
	}

	serviceUID := string(service.GetUID())
	keeper := GetLocalService()
	if !keeper.get(serviceUID) {
		keeper.set(serviceUID)
		utils.Logf(service, "service %s's uid %v is kept in the DeletedSvcKeeper successfully.", service.Name, serviceUID)
	} else {
		utils.Logf(service, "service %s's uid %v has "+
			"already been kept in the DeletedSvcKeeper, no need to update.", service.Name, serviceUID)
	}
	// keeper.set(serviceUIDNoneExist, currentVersion)
	return nil
}
