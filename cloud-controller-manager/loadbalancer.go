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
	"context"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"k8s.io/klog"
	"os"
	"reflect"
	"strings"

	"encoding/json"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/api/core/v1"
)

// AnnotationRequest annotated parameters.
type AnnotationRequest struct {
	Loadbalancerid   string
	LoadBalancerName string
	BackendLabel     string
	BackendType      string

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
	PersistenceTimeout *int
	AddressIPVersion   slb.AddressIPVersionType

	OverrideListeners string

	PrivateZoneName       string
	PrivateZoneId         string
	PrivateZoneRecordName string
	PrivateZoneRecordTTL  int

	RemoveUnscheduledBackend string
	ResourceGroupId          string

	DeleteProtection             slb.FlagType
	ModificationProtectionStatus slb.ModificationProtectionType
	ExternalIPType               string
}

// TAGKEY Default tag key.
const TAGKEY = "kubernetes.do.not.delete"
const ACKKEY = "ack.aliyun.com"
const MDSKEY = "managed.by.ack"

// ClientSLBSDK client sdk for slb
type ClientSLBSDK interface {
	DescribeLoadBalancers(ctx context.Context, args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error)
	CreateLoadBalancer(ctx context.Context, args *slb.CreateLoadBalancerArgs) (response *slb.CreateLoadBalancerResponse, err error)
	SetLoadBalancerName(ctx context.Context, loadBalancerId string, loadBalancerName string) (err error)
	DeleteLoadBalancer(ctx context.Context, loadBalancerId string) (err error)
	SetLoadBalancerDeleteProtection(ctx context.Context, args *slb.SetLoadBalancerDeleteProtectionArgs) (err error)
	ModifyLoadBalancerInstanceSpec(ctx context.Context, args *slb.ModifyLoadBalancerInstanceSpecArgs) (err error)
	ModifyLoadBalancerInternetSpec(ctx context.Context, args *slb.ModifyLoadBalancerInternetSpecArgs) (err error)
	DescribeLoadBalancerAttribute(ctx context.Context, loadBalancerId string) (loadBalancer *slb.LoadBalancerType, err error)
	RemoveBackendServers(ctx context.Context, loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error)
	AddBackendServers(ctx context.Context, loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error)
	SetLoadBalancerModificationProtection(ctx context.Context, args *slb.SetLoadBalancerModificationProtectionArgs) (err error)

	StopLoadBalancerListener(ctx context.Context, loadBalancerId string, port int) (err error)
	StartLoadBalancerListener(ctx context.Context, loadBalancerId string, port int) (err error)
	CreateLoadBalancerTCPListener(ctx context.Context, args *slb.CreateLoadBalancerTCPListenerArgs) (err error)
	CreateLoadBalancerUDPListener(ctx context.Context, args *slb.CreateLoadBalancerUDPListenerArgs) (err error)
	DeleteLoadBalancerListener(ctx context.Context, loadBalancerId string, port int) (err error)
	CreateLoadBalancerHTTPSListener(ctx context.Context, args *slb.CreateLoadBalancerHTTPSListenerArgs) (err error)
	CreateLoadBalancerHTTPListener(ctx context.Context, args *slb.CreateLoadBalancerHTTPListenerArgs) (err error)
	DescribeLoadBalancerHTTPSListenerAttribute(ctx context.Context, loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error)
	DescribeLoadBalancerTCPListenerAttribute(ctx context.Context, loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error)
	DescribeLoadBalancerUDPListenerAttribute(ctx context.Context, loadBalancerId string, port int) (response *slb.DescribeLoadBalancerUDPListenerAttributeResponse, err error)
	DescribeLoadBalancerHTTPListenerAttribute(ctx context.Context, loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPListenerAttributeResponse, err error)

	SetLoadBalancerHTTPListenerAttribute(ctx context.Context, args *slb.SetLoadBalancerHTTPListenerAttributeArgs) (err error)
	SetLoadBalancerHTTPSListenerAttribute(ctx context.Context, args *slb.SetLoadBalancerHTTPSListenerAttributeArgs) (err error)
	SetLoadBalancerTCPListenerAttribute(ctx context.Context, args *slb.SetLoadBalancerTCPListenerAttributeArgs) (err error)
	SetLoadBalancerUDPListenerAttribute(ctx context.Context, args *slb.SetLoadBalancerUDPListenerAttributeArgs) (err error)

	RemoveTags(ctx context.Context, args *slb.RemoveTagsArgs) error
	DescribeTags(ctx context.Context, args *slb.DescribeTagsArgs) (tags []slb.TagItemType, pagination *common.PaginationResult, err error)
	AddTags(ctx context.Context, args *slb.AddTagsArgs) error

	CreateVServerGroup(ctx context.Context, args *slb.CreateVServerGroupArgs) (response *slb.CreateVServerGroupResponse, err error)
	DescribeVServerGroups(ctx context.Context, args *slb.DescribeVServerGroupsArgs) (response *slb.DescribeVServerGroupsResponse, err error)
	DeleteVServerGroup(ctx context.Context, args *slb.DeleteVServerGroupArgs) (response *slb.DeleteVServerGroupResponse, err error)
	SetVServerGroupAttribute(ctx context.Context, args *slb.SetVServerGroupAttributeArgs) (response *slb.SetVServerGroupAttributeResponse, err error)
	DescribeVServerGroupAttribute(ctx context.Context, args *slb.DescribeVServerGroupAttributeArgs) (response *slb.DescribeVServerGroupAttributeResponse, err error)
	ModifyVServerGroupBackendServers(ctx context.Context, args *slb.ModifyVServerGroupBackendServersArgs) (response *slb.ModifyVServerGroupBackendServersResponse, err error)
	AddVServerGroupBackendServers(ctx context.Context, args *slb.AddVServerGroupBackendServersArgs) (response *slb.AddVServerGroupBackendServersResponse, err error)
	RemoveVServerGroupBackendServers(ctx context.Context, args *slb.RemoveVServerGroupBackendServersArgs) (response *slb.RemoveVServerGroupBackendServersResponse, err error)
}

// LoadBalancerClient slb client wrapper
type LoadBalancerClient struct {
	region string
	vpcid  string
	c      ClientSLBSDK
	// known service resource version
	ins ClientInstanceSDK
}

func (s *LoadBalancerClient) FindLoadBalancer(ctx context.Context, service *v1.Service) (bool, *slb.LoadBalancerType, error) {
	def, _ := ExtractAnnotationRequest(service)

	// User assigned lobadbalancer id go first.
	if def.Loadbalancerid != "" {
		return s.FindLoadBalancerByID(ctx, def.Loadbalancerid)
	}
	// if not, find by slb tags
	return s.FindLoadBalancerByTags(ctx, service)
}

func (s *LoadBalancerClient) FindLoadBalancerByID(ctx context.Context, lbid string) (bool, *slb.LoadBalancerType, error) {

	lb, err := s.c.DescribeLoadBalancerAttribute(ctx, lbid)
	if err != nil && strings.Contains(err.Error(), "InvalidLoadBalancerId.NotFound") {
		return false, nil, nil
	}
	return err == nil, lb, err
}

func (s *LoadBalancerClient) FindLoadBalancerByTags(ctx context.Context, service *v1.Service) (bool, *slb.LoadBalancerType, error) {
	if service.UID == "" {
		return false, nil, fmt.Errorf("unexpected empty service uid")
	}
	lbn := GetLoadBalancerName(service)
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
		ctx,
		&slb.DescribeLoadBalancersArgs{
			Tags:     string(items),
			RegionId: DEFAULT_REGION,
		},
	)
	utils.Logf(service, "alicloud: fallback to find loadbalancer by tags [%s]", string(items))
	if err != nil {
		return false, nil, err
	}

	if len(lbs) == 0 {
		// here we need to fallback on finding by name for compatible reason
		// the old service slb may not have a tag.
		return s.FindLoadBalancerByName(ctx, lbn)
	}
	if len(lbs) > 1 {
		utils.Logf(service, "Warning: multiple loadbalancer returned with tags [%s], "+
			"using the first one with IP=%s", string(items), lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(ctx, lbs[0].LoadBalancerId)
	return err == nil, lb, err
}

func (s *LoadBalancerClient) FindLoadBalancerByName(ctx context.Context, name string) (bool, *slb.LoadBalancerType, error) {
	lbs, err := s.c.DescribeLoadBalancers(
		ctx,
		&slb.DescribeLoadBalancersArgs{
			RegionId:         DEFAULT_REGION,
			LoadBalancerName: name,
		},
	)
	klog.V(2).Infof("fallback to find loadbalancer by name [%s]", name)
	if err != nil {
		return false, nil, err
	}

	if len(lbs) == 0 {
		return false, nil, nil
	}
	if len(lbs) > 1 {
		klog.Warningf("alicloud: multiple loadbalancer returned with name [%s], "+
			"using the first one with IP=%s", name, lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(ctx, lbs[0].LoadBalancerId)
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
func (s *LoadBalancerClient) EnsureLoadBalancer(ctx context.Context, service *v1.Service, nodes *EndpointWithENI, vswitchid string) (*slb.LoadBalancerType, error) {
	utils.Logf(service, "ensure loadbalancer with service details, \n%+v", PrettyJson(service))

	exists, origined, err := s.FindLoadBalancer(ctx, service)
	if err != nil {
		return nil, err
	}
	utils.Logf(service, "find loadbalancer with result, exist=%v, %s\n", exists, PrettyJson(origined))
	_, request := ExtractAnnotationRequest(service)

	var derr error
	serviceHashChanged := true
	// this is a workaround for issue: https://github.com/kubernetes/kubernetes/issues/59084
	if !exists {
		// If need created, double check if the resource id has been deleted
		if isServiceDeleted(service) {
			klog.V(2).Infof("alicloud: isServiceDeleted report that this service has been " +
				"deleted before. see issue: https://github.com/kubernetes/kubernetes/issues/59084")
			os.Exit(1)
		}
		if isLoadbalancerOwnIngress(service) {
			return nil, fmt.Errorf("alicloud: not able to find loadbalancer "+
				"named [%s] in openapi, but it's defined in service.loaderbalancer.ingress. "+
				"this may happen when you removed loadbalancerid annotation", service.Name)
		}
		if request.Loadbalancerid != "" {
			return nil, fmt.Errorf("alicloud: user specified "+
				"loadbalancer[%s] does not exist. pls check", request.Loadbalancerid)
		}

		// From here, we need to create a new loadbalancer
		klog.V(5).Infof("alicloud: can not find a "+
			"loadbalancer with service name [%s/%s], creating a new one", service.Namespace, service.Name)
		opts := s.getLoadBalancerOpts(service, vswitchid)
		lbr, err := s.c.CreateLoadBalancer(ctx, opts)
		if err != nil {
			return nil, err
		}

		//deal with loadBalancer tags
		tags := getLoadBalancerAdditionalTags(getBackwardsCompatibleAnnotation(service.Annotations))
		loadbalancerName := GetLoadBalancerName(service)
		// Add default tags
		tags[TAGKEY] = loadbalancerName
		tags[ACKKEY] = CLUSTER_ID

		tagItemArr := make([]slb.TagItem, 0)
		for key, value := range tags {
			tagItemArr = append(tagItemArr, slb.TagItem{TagKey: key, TagValue: value})
		}
		// TODO if the key size or value size > slb's tag limit or number of tags  >limit(20), tags will insert error.

		tagItems, err := json.Marshal(tagItemArr)

		if err != nil {
			return nil, err
		}
		if err := s.c.AddTags(
			ctx,
			&slb.AddTagsArgs{
				RegionId:       opts.RegionId,
				LoadBalancerID: lbr.LoadBalancerId,
				Tags:           string(tagItems),
			},
		); err != nil {
			return nil, err
		}

		origined, derr = s.c.DescribeLoadBalancerAttribute(ctx, lbr.LoadBalancerId)
	} else {
		// Need to verify loadbalancer.
		// Reuse SLB is not allowed when the SLB is created by k8s service.
		tags, _, err := s.c.DescribeTags(
			ctx,
			&slb.DescribeTagsArgs{
				RegionId:       origined.RegionId,
				LoadBalancerID: origined.LoadBalancerId,
			})
		if err != nil {
			return origined, err
		}
		if ok, reason := isLoadBalancerNonReusable(tags, service); ok {
			return origined, fmt.Errorf("alicloud: the loadbalancer %s can not be reused, %s", origined.LoadBalancerId, reason)
		}

		serviceHashChanged, err = utils.IsServiceHashChanged(service)
		if err != nil {
			return origined, fmt.Errorf("compute svc hash error :%s", err.Error())
		}
		if serviceHashChanged {
			if err := updateLoadBalancerByAnnotations(ctx, s.c, origined, service, request, tags); err != nil {
				return origined, err
			}
			origined, derr = s.c.DescribeLoadBalancerAttribute(ctx, origined.LoadBalancerId)
		}
	}
	if derr != nil {
		utils.Logf(service, "alicloud: can not get loadbalancer attribute. ")
		return nil, derr
	}
	vgs := BuildVirtualGroupFromService(s, service, origined)

	// Make sure virtual server backend group has been updated.
	if err := EnsureVirtualGroups(ctx, vgs, nodes); err != nil {
		return origined, fmt.Errorf("update backend servers: error %s", err.Error())
	}
	// Apply listener when
	//   1. user does not assign loadbalancer id by themselves.
	//   2. force-override-listener annotation is set.
	if serviceHashChanged {
		if (!isUserDefinedLoadBalancer(service)) ||
			(isUserDefinedLoadBalancer(service) && isOverrideListeners(service)) {
			utils.Logf(service, "not user defined loadbalancer[%s], start to apply listener.", origined.LoadBalancerId)
			// If listener update is needed. Switch to vserver group immediately.
			// No longer update default backend servers.
			if err := EnsureListeners(ctx, s, service, origined, vgs); err != nil {

				return origined, fmt.Errorf("ensure listener error: %s", err.Error())
			}
		}
	}
	return origined, s.UpdateLoadBalancer(ctx, service, nodes, false)
}

func isLoadBalancerNonReusable(tags []slb.TagItemType, service *v1.Service) (bool, string) {
	for _, tag := range tags {
		if isUserDefinedLoadBalancer(service) &&
			(tag.TagKey == TAGKEY || tag.TagKey == ACKKEY) {
			return true, "can not reuse loadbalancer created by kubernetes."
		}
	}
	return false, ""
}

func isLoadBalancerHasTag(tags []slb.TagItemType) bool {
	for _, tag := range tags {
		if tag.TagKey == TAGKEY {
			return true
		}
	}
	return false
}

func updateLoadBalancerByAnnotations(context context.Context, slbClient ClientSLBSDK, lb *slb.LoadBalancerType,
	service *v1.Service, request *AnnotationRequest, tags []slb.TagItemType) error {
	klog.V(5).Infof("alicloud: found "+
		"an exist loadbalancer[%s], check to see whether update is needed.", lb.LoadBalancerId)

	if request.MasterZoneID != "" && request.MasterZoneID != lb.MasterZoneId {
		return fmt.Errorf("alicloud: can not change LoadBalancer master zone id once created")
	}
	if request.SlaveZoneID != "" && request.SlaveZoneID != lb.SlaveZoneId {
		return fmt.Errorf("alicloud: can not change LoadBalancer slave zone id once created")
	}
	if request.AddressType != "" && request.AddressType != lb.AddressType {
		return fmt.Errorf("alicloud: can not change LoadBalancer AddressType once created. delete and retry")
	}
	if !equalsAddressIPVersion(request.AddressIPVersion, lb.AddressIPVersion) {
		return fmt.Errorf("alicloud: can not change LoadBalancer AddressIPVersion once created")
	}

	// update chargeType & bandwidth
	needUpdate, charge, bandwidth := false, lb.InternetChargeType, lb.Bandwidth
	if request.ChargeType != "" && request.ChargeType != lb.InternetChargeType {
		needUpdate = true
		charge = request.ChargeType
		utils.Logf(service, "internet chargeType changed([%s] -> [%s]), update loadbalancer [%s]",
			string(lb.InternetChargeType), string(request.ChargeType), lb.LoadBalancerId)
	}
	if request.ChargeType == slb.PayByBandwidth && request.Bandwidth != lb.Bandwidth && request.Bandwidth != 0 {
		needUpdate = true
		bandwidth = request.Bandwidth
		utils.Logf(service, "bandwidth changed([%d] -> [%d]), update loadbalancer[%s]",
			lb.Bandwidth, request.Bandwidth, lb.LoadBalancerId)
	}
	if needUpdate {
		if lb.AddressType == slb.InternetAddressType {
			utils.Logf(service, "modify loadbalancer: chargeType=%s, bandwidth=%d", charge, bandwidth)
			if err := slbClient.ModifyLoadBalancerInternetSpec(
				context,
				&slb.ModifyLoadBalancerInternetSpecArgs{
					LoadBalancerId:     lb.LoadBalancerId,
					InternetChargeType: charge,
					Bandwidth:          bandwidth,
				}); err != nil {
				return err
			}
		} else {
			utils.Logf(service, "only internet loadbalancer is allowed to modify bandwidth and pay type")
		}
	}

	// update instance spec
	if request.LoadBalancerSpec != "" && request.LoadBalancerSpec != lb.LoadBalancerSpec {
		klog.Infof("alicloud: loadbalancerSpec changed ([%s] -> [%s]), update loadbalancer [%s]",
			lb.LoadBalancerSpec, request.LoadBalancerSpec, lb.LoadBalancerId)
		if err := slbClient.ModifyLoadBalancerInstanceSpec(
			context,
			&slb.ModifyLoadBalancerInstanceSpecArgs{
				RegionId:         lb.RegionId,
				LoadBalancerId:   lb.LoadBalancerId,
				LoadBalancerSpec: request.LoadBalancerSpec,
			},
		); err != nil {
			return err
		}
	}

	// update slb delete protection
	if request.DeleteProtection != "" && request.DeleteProtection != lb.DeleteProtection {
		utils.Logf(service, "delete protection changed([%d] -> [%d]), update loadbalancer [%s]",
			lb.DeleteProtection, request.DeleteProtection, lb.LoadBalancerId)
		if err := slbClient.SetLoadBalancerDeleteProtection(
			context,
			&slb.SetLoadBalancerDeleteProtectionArgs{
				RegionId:         lb.RegionId,
				LoadBalancerId:   lb.LoadBalancerId,
				DeleteProtection: request.DeleteProtection,
			},
		); err != nil {
			return err
		}
	}

	// update modification protection
	if request.ModificationProtectionStatus != "" && request.ModificationProtectionStatus != lb.ModificationProtectionStatus {
		klog.Infof("alicloud: loadbalancer modification protection changed([%s] -> [%s]) changed, update loadbalancer [%s]",
			lb.ModificationProtectionStatus, request.ModificationProtectionStatus, lb.LoadBalancerName)
		args := slb.SetLoadBalancerModificationProtectionArgs{
			RegionId:                     lb.RegionId,
			LoadBalancerId:               lb.LoadBalancerId,
			ModificationProtectionStatus: request.ModificationProtectionStatus,
			ModificationProtectionReason: MDSKEY,
		}
		if err := slbClient.SetLoadBalancerModificationProtection(context, &args); err != nil {
			return err
		}
	}

	// update slb name
	// only user defined slb or slb which has "kubernetes.do.not.delete" tag can update name
	if request.LoadBalancerName != "" && request.LoadBalancerName != lb.LoadBalancerName {
		if isLoadBalancerHasTag(tags) || isUserDefinedLoadBalancer(service) {
			klog.Infof("alicloud: LoadBalancer name (%s -> %s) changed, update loadbalancer [%s]",
				lb.LoadBalancerName, request.LoadBalancerName, lb.LoadBalancerId)
			if err := slbClient.SetLoadBalancerName(context, lb.LoadBalancerId, request.LoadBalancerName); err != nil {
				return err
			}
		} else {
			record, err := utils.GetRecorderFromContext(context)
			if err != nil {
				klog.Warningf("get recorder error: %s", err.Error())
				klog.Warningf("alicloud: LoadBalancer name (%s -> %s) changed, try to update loadbalancer [%s],"+
					" warning: only user defined slb or slb which has 'kubernetes.do.not.delete' tag can update name",
					lb.LoadBalancerName, request.LoadBalancerName, lb.LoadBalancerId)
			} else {
				record.Eventf(
					service,
					v1.EventTypeWarning,
					"SetLoadBalancerNameFailed",
					"Error setting load balancer %s name: "+
						"only user defined slb or slb which has 'kubernetes.do.not.delete' tag can update name",
					lb.LoadBalancerName, request.LoadBalancerName, lb.LoadBalancerId,
				)
			}
		}
	}

	return nil
}

//UpdateLoadBalancer make sure slb backend is reconciled
func (s *LoadBalancerClient) UpdateLoadBalancer(ctx context.Context, service *v1.Service, nodes *EndpointWithENI, withVgroup bool) error {

	exists, lb, err := s.FindLoadBalancer(ctx, service)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the loadbalance you specified by name [%s] does not exist", service.Name)
	}
	if withVgroup {
		vgs := BuildVirtualGroupFromService(s, service, lb)
		if err := EnsureVirtualGroups(ctx, vgs, nodes); err != nil {
			return fmt.Errorf("update backend servers: error %s", err.Error())
		}
	}
	if !needUpdateDefaultBackend(service, lb) {
		return nil
	}
	utils.Logf(service, "update default backend server group")
	return s.UpdateDefaultServerGroup(ctx, nodes, lb)
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
func (s *LoadBalancerClient) EnsureLoadBalanceDeleted(ctx context.Context, service *v1.Service) error {
	// need to save the resource version when deleted event
	// need to remove. when svc type changed (LoadBalancer -> ClusterIP -> LoadBalancer), ccm will restart.
	err := keepResourceVersion(service)
	if err != nil {
		utils.Logf(service, "Warning: failed to save deleted service resourceVersion,due to [%s] ", err.Error())
	}
	exists, lb, err := s.FindLoadBalancer(ctx, service)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	// skip delete user defined loadbalancer
	if isUserDefinedLoadBalancer(service) {
		utils.Logf(service, "user managed loadbalancer will not be deleted by cloudprovider.")
		return EnsureListenersDeleted(ctx, s.c, service, lb, BuildVirtualGroupFromService(s, service, lb))
	}

	// set delete protection off
	if lb.DeleteProtection == slb.OnFlag {
		if err := s.c.SetLoadBalancerDeleteProtection(
			ctx,
			&slb.SetLoadBalancerDeleteProtectionArgs{
				RegionId:         lb.RegionId,
				LoadBalancerId:   lb.LoadBalancerId,
				DeleteProtection: slb.OffFlag,
			},
		); err != nil {
			return fmt.Errorf("error to set slb id [%s] delete protection off, svc [%s], err: %s", lb.LoadBalancerId, service.Name, err.Error())
		}
	}

	return s.c.DeleteLoadBalancer(ctx, lb.LoadBalancerId)
}

func (s *LoadBalancerClient) getLoadBalancerOpts(service *v1.Service, vswitchid string) (args *slb.CreateLoadBalancerArgs) {
	ar, req := ExtractAnnotationRequest(service)
	args = &slb.CreateLoadBalancerArgs{
		AddressType:                  ar.AddressType,
		InternetChargeType:           ar.ChargeType,
		RegionId:                     DEFAULT_REGION,
		LoadBalancerSpec:             ar.LoadBalancerSpec,
		MasterZoneId:                 ar.MasterZoneID,
		SlaveZoneId:                  ar.SlaveZoneID,
		AddressIPVersion:             ar.AddressIPVersion,
		DeleteProtection:             ar.DeleteProtection,
		ResourceGroupId:              ar.ResourceGroupId,
		ModificationProtectionStatus: ar.ModificationProtectionStatus,
		ModificationProtectionReason: MDSKEY,
	}
	// paybybandwidth need a default bandwidth args, while paybytraffic doesnt.
	if ar.ChargeType == slb.PayByBandwidth ||
		(ar.ChargeType == slb.PayByTraffic && req.Bandwidth != 0) {
		klog.V(5).Infof("alicloud: %s, set bandwidth to %d", ar.ChargeType, ar.Bandwidth)
		args.Bandwidth = ar.Bandwidth
	}
	if ar.SLBNetworkType != "classic" &&
		strings.Compare(string(ar.AddressType), string(slb.IntranetAddressType)) == 0 {

		utils.Logf(service, "intranet vpc "+
			"loadbalancer will be created. address type=%s, switchid=%s", ar.AddressType, vswitchid)
		args.VSwitchId = vswitchid
	}
	if req.LoadBalancerName == "" {
		args.LoadBalancerName = GetLoadBalancerName(service)
	} else {
		args.LoadBalancerName = req.LoadBalancerName
	}
	return
}

// DEFAULT_SERVER_WEIGHT default server weight
const DEFAULT_SERVER_WEIGHT = 100

// UpdateDefaultServerGroup update default server group
func (s *LoadBalancerClient) UpdateDefaultServerGroup(ctx context.Context, backends interface{}, lb *slb.LoadBalancerType) error {
	nodes, ok := backends.([]*v1.Node)
	if !ok {
		klog.Infof("skip default server group update for type %s", reflect.TypeOf(backends))
		return nil
	}
	additions, deletions := []slb.BackendServerType{}, []string{}
	klog.V(5).Infof("alicloud: try to update loadbalancer backend servers. [%s]", lb.LoadBalancerId)
	// checkout for newly added servers
	for _, n1 := range nodes {
		found := false
		_, id, err := nodeFromProviderID(n1.Spec.ProviderID)
		for _, n2 := range lb.BackendServers.BackendServer {
			if err != nil {
				klog.Errorf("alicloud: node providerid=%s is not"+
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
		klog.V(5).Infof("alicloud: add loadbalancer backend for [%s] \n %s\n", lb.LoadBalancerId, PrettyJson(additions))
		// only 20 backend servers is accepted per delete.
		for len(additions) > 0 {
			var target []slb.BackendServerType
			if len(additions) > MAX_LOADBALANCER_BACKEND {
				target = additions[0:MAX_LOADBALANCER_BACKEND]
				additions = additions[MAX_LOADBALANCER_BACKEND:]
				klog.V(5).Infof("alicloud: batch add backend servers, %v", target)
			} else {
				target = additions
				additions = []slb.BackendServerType{}
				klog.V(5).Infof("alicloud: batch add backend servers, else %v", target)
			}
			if _, err := s.c.AddBackendServers(ctx, lb.LoadBalancerId, target); err != nil {
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
				klog.Errorf("alicloud: node providerid=%s is not "+
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
		klog.V(5).Infof("alicloud: delete loadbalancer backend for [%s] %v", lb.LoadBalancerId, deletions)
		// only 20 backend servers is accepted per delete.
		for len(deletions) > 0 {
			var target []string
			if len(deletions) > MAX_LOADBALANCER_BACKEND {
				target = deletions[0:MAX_LOADBALANCER_BACKEND]
				deletions = deletions[MAX_LOADBALANCER_BACKEND:]
				klog.V(5).Infof("alicloud: batch delete backend servers, %s", target)
			} else {
				target = deletions
				deletions = []string{}
				klog.V(5).Infof("alicloud: batch delete backend servers else, %s", target)
			}
			var mdelete []slb.BackendServerType
			for _, del := range target {
				mdelete = append(mdelete, slb.BackendServerType{ServerId: del})
			}
			_, err := s.c.RemoveBackendServers(ctx, lb.LoadBalancerId, mdelete)
			if err != nil {
				return err
			}
		}
	}
	if len(additions) <= 0 && len(deletions) <= 0 {
		klog.V(5).Infof("alicloud: no backend servers need to be updated.[%s]", lb.LoadBalancerId)
		return nil
	}
	klog.V(5).Infof("alicloud: update loadbalancer`s backend servers finished! [%s]", lb.LoadBalancerId)
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
	if service == nil {
		utils.Logf(service, "service is nil")
		return false
	}

	if len(service.Status.LoadBalancer.Ingress) == 0 {
		utils.Logf(service, "service %v doesn't have ingresses", service.Name)
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
func keepResourceVersion(service *v1.Service) error {
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

func IsENIBackendType(svc *v1.Service) bool {
	if svc.Annotations[utils.BACKEND_TYPE_LABEL] != "" {
		return svc.Annotations[utils.BACKEND_TYPE_LABEL] == utils.BACKEND_TYPE_ENI
	}

	if os.Getenv("SERVICE_FORCE_BACKEND_ENI") != "" {
		return os.Getenv("SERVICE_FORCE_BACKEND_ENI") == "true"
	}

	return cfg.Global.ServiceBackendType == utils.BACKEND_TYPE_ENI
}
