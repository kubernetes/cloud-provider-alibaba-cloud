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
	"os"
	"strings"

	"encoding/json"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

type AnnotationRequest struct {
	Loadbalancerid string
	BackendLabel   string

	SSLPorts       string
	AddressType    slb.AddressType
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

	StickySession      slb.FlagType
	StickySessionType  slb.StickySessionType
	Cookie             string
	CookieTimeout      int
	PersistenceTimeout int

	OverrideListeners string
}

const TAGKEY = "kubernetes.do.not.delete"

type ClientSLBSDK interface {
	DescribeLoadBalancers(args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error)
	CreateLoadBalancer(args *slb.CreateLoadBalancerArgs) (response *slb.CreateLoadBalancerResponse, err error)
	DeleteLoadBalancer(loadBalancerId string) (err error)
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
}

type LoadBalancerClient struct {
	c ClientSLBSDK
	// known service resource version
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
	glog.V(4).Infof("alicloud: find loadbalancer with id [%s], %d found.", lbid, len(lbs))
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		return false, nil, nil
	}
	if len(lbs) > 1 {
		glog.Warningf("alicloud: "+
			"multiple loadbalancer returned with id [%s], using the first one with IP=%s", lbid, lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(lbs[0].LoadBalancerId)
	return err == nil, lb, err
}

func (s *LoadBalancerClient) findLoadBalancerByTags(service *v1.Service) (bool, *slb.LoadBalancerType, error) {
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
	glog.V(2).Infof("alicloud: fallback to find loadbalancer by tags [%s]\n", string(items))
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		// here we need to fallback on finding by name for compatible reason
		// the old service slb may not have a tag.
		return s.findLoadBalancerByName(lbn)
	}
	if len(lbs) > 1 {
		glog.Warningf("alicloud: multiple loadbalancer returned with tags [%s], "+
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
	glog.V(2).Infof("alicloud: fallback to find loadbalancer by name [%s]\n", name)
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

func (s *LoadBalancerClient) EnsureLoadBalancer(service *v1.Service, nodes []*v1.Node, vswitchid string) (*slb.LoadBalancerType, error) {
	glog.V(4).Infof("alicloud: ensure loadbalancer with service details, \n%+v", PrettyJson(service))

	exists, origined, err := s.findLoadBalancer(service)
	if err != nil {
		return nil, err
	}
	glog.V(4).Infof("alicloud: find loadbalancer with result, exist=%v\n %s\n", exists, PrettyJson(origined))
	_, request := ExtractAnnotationRequest(service)

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
				"loadbalancer[%s] does not exist. pls check!", request.Loadbalancerid)
		}

		// From here, we need to create a new loadbalancer
		glog.V(5).Infof("alicloud: can not find a loadbalancer with service name [%s/%s], creating a new one\n", service.Namespace, service.Name)
		opts := s.getLoadBalancerOpts(service, vswitchid)
		lbr, err := s.c.CreateLoadBalancer(opts)
		if err != nil {
			return nil, err
		}

		// Tag the loadbalancer.
		items, err := json.Marshal(
			[]slb.TagItem{
				{
					TagKey:   TAGKEY,
					TagValue: opts.LoadBalancerName,
				},
			},
		)
		if err != nil {
			return nil, err
		}
		if err := s.c.AddTags(&slb.AddTagsArgs{
			RegionId:       opts.RegionId,
			LoadBalancerID: lbr.LoadBalancerId,
			Tags:           string(items),
		}); err != nil {
			return nil, err
		}
		origined, err = s.c.DescribeLoadBalancerAttribute(lbr.LoadBalancerId)
	} else {
		needUpdate, charge, bandwidth := false, origined.InternetChargeType, origined.Bandwidth
		glog.V(5).Infof("alicloud: found an exist loadbalancer[%s], check to see whether update is needed.", origined.LoadBalancerId)
		if request.MasterZoneID != "" && request.MasterZoneID != origined.MasterZoneId {
			return nil, fmt.Errorf("alicloud: can not change LoadBalancer master zone id once created.")
		}
		if request.SlaveZoneID != "" && request.SlaveZoneID != origined.SlaveZoneId {
			return nil, fmt.Errorf("alicloud: can not change LoadBalancer slave zone id once created.")
		}
		if request.ChargeType != "" && request.ChargeType != origined.InternetChargeType {
			needUpdate = true
			charge = request.ChargeType
			// Todo: here we need to compare loadbalance
			glog.Infof("alicloud: internet charge type([%s] -> [%s]), update loadbalancer [%s]\n",
				string(origined.InternetChargeType), string(request.ChargeType), origined.LoadBalancerName)
		}
		if request.ChargeType == slb.PayByBandwidth &&
			request.Bandwidth != DEFAULT_BANDWIDTH &&
			request.Bandwidth != origined.Bandwidth {

			needUpdate = true
			bandwidth = request.Bandwidth
			glog.Infof("alicloud: bandwidth([%d] -> [%d]) changed, update loadbalancer [%s]\n",
				origined.Bandwidth, request.Bandwidth, origined.LoadBalancerName)
		}
		if needUpdate {
			if err := s.c.ModifyLoadBalancerInternetSpec(
				&slb.ModifyLoadBalancerInternetSpecArgs{
					LoadBalancerId:     origined.LoadBalancerId,
					InternetChargeType: charge,
					Bandwidth:          bandwidth,
				}); err != nil {
				return nil, err
			}
		}
		if request.AddressType != "" && request.AddressType != origined.AddressType {
			glog.Errorf("alicloud: warning! can not change "+
				"loadbalancer address type after it has been created! please "+
				"recreate the service.[%s]->[%s],[%s]\n", origined.AddressType, request.AddressType, origined.LoadBalancerName)
			return nil, errors.New("alicloud: change loadbalancer address type after service has been created is not supported. delete and retry")
		}
		origined, err = s.c.DescribeLoadBalancerAttribute(origined.LoadBalancerId)
	}
	if err != nil {
		glog.Errorf("alicloud: can not get loadbalancer[%s] attribute. ", origined.LoadBalancerId)
		return nil, err
	}
	// Apply listener when
	//   1. user does not assign loadbalancer id by themselves.
	//   2. force-override-listener annotation is set.
	if (!isUserDefinedLoadBalancer(request)) ||
		(isUserDefinedLoadBalancer(request) && isOverrideListeners(request)) {
		glog.V(5).Infof("alicloud: not user defined loadbalancer[%s], start to apply listener.\n", origined.LoadBalancerId)
		err = NewListenerManager(s.c, service, origined).Apply()
		if err != nil {
			return nil, err
		}
	}
	return origined, s.UpdateBackendServers(nodes, origined)
}

func (s *LoadBalancerClient) UpdateLoadBalancer(service *v1.Service, nodes []*v1.Node) error {

	exists, lb, err := s.findLoadBalancer(service)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the loadbalance you specified by name [%s] does not exist!", service.Name)
	}

	return s.UpdateBackendServers(nodes, lb)
}

func (s *LoadBalancerClient) EnsureLoadBalanceDeleted(service *v1.Service) error {

	// need to save the resource version when deleted event
	err := keepResourceVesion(service)
	if err != nil {
		glog.Warningf("alicloud: failed to save "+
			"deleted service resourceVersion, [%s] due to [%s] ", service.Name, err.Error())
	}
	_, request := ExtractAnnotationRequest(service)
	// skip delete user defined loadbalancer
	if isUserDefinedLoadBalancer(request) {
		glog.V(2).Infof("alicloud: user created loadbalancer "+
			"will not be deleted by cloudprovider. service [%s]", service.Name)
		return nil
	}

	exists, lb, err := s.findLoadBalancer(service)
	if err != nil {
		return err
	}
	if !exists {
		return nil
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
	}
	// paybybandwidth need a default bandwidth args, while paybytraffic doesnt.
	if ar.ChargeType == slb.PayByBandwidth ||
		(ar.ChargeType == slb.PayByTraffic && req.Bandwidth != 0) {
		glog.V(5).Infof("alicloud: %s, set bandwidth to %d", ar.ChargeType, ar.Bandwidth)
		args.Bandwidth = ar.Bandwidth
	}
	if ar.SLBNetworkType != "classic" &&
		strings.Compare(string(ar.AddressType), string(slb.IntranetAddressType)) == 0 {

		glog.Infof("alicloud: intranet vpc "+
			"loadbalancer will be created. address type=%s, switchid=%s\n", ar.AddressType, vswitchid)
		args.VSwitchId = vswitchid
	}
	args.LoadBalancerName = cloudprovider.GetLoadBalancerName(service)
	return
}

const DEFAULT_SERVER_WEIGHT = 100

func (s *LoadBalancerClient) UpdateBackendServers(nodes []*v1.Node, lb *slb.LoadBalancerType) error {
	additions, deletions := []slb.BackendServerType{}, []string{}
	glog.V(5).Infof("alicloud: try to update loadbalancer backend servers. [%s]\n", lb.LoadBalancerId)
	// checkout for newly added servers
	for _, n1 := range nodes {
		found := false
		_, id, err := nodeinfo(types.NodeName(n1.Spec.ProviderID))
		for _, n2 := range lb.BackendServers.BackendServer {
			if err != nil {
				glog.Errorf("alicloud: node providerid=%s is not"+
					" in the correct form, expect regionid.instanceid. skip add op", n1.Spec.ProviderID)
				continue
			}
			if string(id) == n2.ServerId {
				found = true
				break
			}
		}
		if !found {
			additions = append(additions, slb.BackendServerType{ServerId: string(id), Weight: DEFAULT_SERVER_WEIGHT})
		}
	}
	if len(additions) > 0 {
		glog.V(5).Infof("alicloud: add loadbalancer backend for [%s] \n %s\n", lb.LoadBalancerId, PrettyJson(additions))
		// only 20 backend servers is accepted per delete.
		for len(additions) > 0 {
			target := []slb.BackendServerType{}
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
			_, id, err := nodeinfo(types.NodeName(n2.Spec.ProviderID))
			if err != nil {
				glog.Errorf("alicloud: node providerid=%s is not "+
					"in the correct form, expect regionid.instanceid.. skip delete op... [%s]", n2.Spec.ProviderID, err.Error())
				continue
			}
			if n1.ServerId == string(id) {
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
			target := []string{}
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
func isUserDefinedLoadBalancer(request *AnnotationRequest) bool {

	return request.Loadbalancerid != ""
}

func isOverrideListeners(request *AnnotationRequest) bool {

	return strings.ToLower(request.OverrideListeners) == "true"
}

// check if the service exists in service definition
func isLoadbalancerOwnIngress(service *v1.Service) bool {
	if service == nil ||
		len(service.Status.LoadBalancer.Ingress) == 0 {
		glog.V(2).Infof("alicloud: service "+
			"%s doesn't have ingresses\n", service.Name)
		return false
	}
	glog.V(2).Infof("alicloud: service %s has"+
		" ingresses=%v\n", service.Name, service.Status.LoadBalancer.Ingress)
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
		glog.V(2).Infof("alicloud: service "+
			"%s's uid %v has been deleted, shouldn't be created again.\n", service.Name, serviceUID)
		return true
	}
	glog.V(2).Infof("alicloud: service %s's uid %v "+
		"hasn't been deleted, first time to process, as expected.\n", service.Name, serviceUID)
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
		glog.V(2).Infof("alicloud: "+
			"service %s's uid %v is kept in the DeletedSvcKeeper successfully.\n", service.Name, serviceUID)
	} else {
		glog.V(2).Infof("alicloud: service %s's uid %v has "+
			"already been kept in the DeletedSvcKeeper, no need to update.\n", service.Name, serviceUID)
	}
	// keeper.set(serviceUID, currentVersion)
	return nil
}
