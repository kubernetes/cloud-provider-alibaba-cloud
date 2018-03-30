package alicloud

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

type AnnotationRequest struct {
	Loadbalancerid 		string
	BackendLabel   		string

	SSLPorts       		string
	AddressType    		slb.AddressType
	SLBNetworkType 		string

	ChargeType 		slb.InternetChargeType
	Region     		common.Region
	Bandwidth  		int
	CertID     		string

	HealthCheck            	slb.FlagType
	HealthCheckURI         	string
	HealthCheckConnectPort 	int
	HealthyThreshold       	int
	UnhealthyThreshold     	int
	HealthCheckInterval    	int

	HealthCheckConnectTimeout int                 // for tcp
	HealthCheckType           slb.HealthCheckType // for tcp, Type could be http tcp
	HealthCheckTimeout        int                 // for https and http
}

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
}

type LoadBalancerClient struct {
	c ClientSLBSDK
	// known service resource version
}

func (s *LoadBalancerClient) findLoadBalancer(service *v1.Service) (bool, *slb.LoadBalancerType, error) {
	ar := ExtractAnnotationRequest(service)

	loadbalancer := service.Status.LoadBalancer
	if len(loadbalancer.Ingress) > 0 {
		lbid := loadbalancer.Ingress[0].Hostname
		if lbid != "" {
			if ar.Loadbalancerid != "" && ar.Loadbalancerid != lbid {
				glog.Warningf(fmt.Sprintf("alicloud: changing loadbalancer id was not allowed after loadbalancer"+
					" was already created ! please remove service annotation: [%s]\n", ServiceAnnotationLoadBalancerId))
			}
			// found loadbalancer id in service ingress status.
			// this id was set previously when loadbalancer was created.
			return s.findLoadBalancerByID(lbid, ar.Region)
		}
	}
	// if service ingress status was not initialized with loadbalancer, then
	// we check annotation to see if user had assigned a loadbalancer manually.
	// if so , we use user defined loadbalancer
	if ar.Loadbalancerid != "" {
		return s.findLoadBalancerByID(ar.Loadbalancerid, ar.Region)
	}

	// finally , fallback to find by name to compatible with old version
	return s.findLoadBalancerByName(cloudprovider.GetLoadBalancerName(service), ar.Region)
}

func (s *LoadBalancerClient) findLoadBalancerByID(lbid string, region common.Region) (bool, *slb.LoadBalancerType, error) {

	lbs, err := s.c.DescribeLoadBalancers(
		&slb.DescribeLoadBalancersArgs{
			RegionId:       region,
			LoadBalancerId: lbid,
		},
	)
	glog.V(4).Infof("alicloud: find loadbalancer with id [%s], %+v", lbid, lbs)
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		return false, nil, fmt.Errorf(" alicloud: can't find the specified loadbalancer id [%s]", lbid)
	}
	if len(lbs) > 1 {
		glog.Warningf("alicloud: "+
			"multiple loadbalancer returned with id [%s], using the first one with IP=%s", lbid, lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(lbs[0].LoadBalancerId)
	return err == nil, lb, err
}

func (s *LoadBalancerClient) findLoadBalancerByName(lbn string, region common.Region) (bool, *slb.LoadBalancerType, error) {
	lbs, err := s.c.DescribeLoadBalancers(
		&slb.DescribeLoadBalancersArgs{
			RegionId:         region,
			LoadBalancerName: lbn,
		},
	)
	glog.V(2).Infof("alicloud: fallback to find loadbalancer by name [%s]\n", lbs)
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		return false, nil, nil
	}
	if len(lbs) > 1 {
		glog.Warningf("alicloud: multiple loadbalancer returned with name [%s], "+
			"using the first one with IP=%s", lbn, lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(lbs[0].LoadBalancerId)
	return err == nil, lb, err
}

func (s *LoadBalancerClient) EnsureLoadBalancer(service *v1.Service, nodes []*v1.Node, vswitchid string) (*slb.LoadBalancerType, error) {
	glog.V(4).Infof("alicloud: ensure loadbalancer with service details, %+v", service)

	exists, lb, err := s.findLoadBalancer(service)
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("alicloud: find loadbalancer with result, exist=%v, %+v\n", exists, lb)

	// this is a workaround for issue: https://github.com/kubernetes/kubernetes/issues/59084
	if !exists {
		// If need created, double check if the resource id has been deleted
		err = s.EnsureSVCNotDeleted(service)
		if err != nil {
			glog.V(2).Infof("alicloud: EnsureSVCNotDeleted func can't work properly due to %+v, exit", err)
			os.Exit(1)
		}
	}

	if !exists && checkIfSLBExistInService(service) {
		return nil, fmt.Errorf("alicloud: not able to find loadbalancer "+
			"named [%s] in openapi, but it's defined in service.loaderbalancer.ingress.\n", service.Name)
	}

	opts := s.getLoadBalancerOpts(service, vswitchid)

	if !exists {
		lb, err = s.create(opts)
	} else {
		lb, err = s.update(lb, opts)
	}
	if err != nil {
		return nil, err
	}
	err = NewListenerManager(s.c, service, lb).ApplyUpdate()
	if err != nil {
		return nil, err
	}

	return lb, s.UpdateBackendServers(nodes, lb)
}

func (s *LoadBalancerClient) UpdateLoadBalancer(service *v1.Service, nodes []*v1.Node) error {

	exists, lb, err := s.findLoadBalancer(service)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(fmt.Sprintf("the loadbalance you specified by name [%s] does not exist!", service.Name))
	}

	return s.UpdateBackendServers(nodes, lb)
}

//create , play actual create.
func (s *LoadBalancerClient) create(loadbalancer *slb.CreateLoadBalancerArgs) (*slb.LoadBalancerType, error) {
	lbr, err := s.c.CreateLoadBalancer(loadbalancer)
	if err != nil {
		return nil, err
	}
	return s.c.DescribeLoadBalancerAttribute(lbr.LoadBalancerId)
}

//update, play loadbalancer update
func (s *LoadBalancerClient) update(old *slb.LoadBalancerType, new *slb.CreateLoadBalancerArgs) (*slb.LoadBalancerType, error) {
	// Todo: here we need to compare loadbalance
	if new.InternetChargeType != old.InternetChargeType {
		glog.Infof("alicloud: internet charge type changed. [%s] -> [%s], update loadbalancer [%s]\n",
			string(old.InternetChargeType), string(new.InternetChargeType), new.LoadBalancerName)

		if err := s.c.ModifyLoadBalancerInternetSpec(
			&slb.ModifyLoadBalancerInternetSpecArgs{
				LoadBalancerId:     old.LoadBalancerId,
				InternetChargeType: new.InternetChargeType,
				//Bandwidth:          opts.Bandwidth,
			}); err != nil {
			return nil, err
		}
	}
	if new.AddressType != old.AddressType {
		glog.Errorf("alicloud: warning! can not change "+
			"loadbalancer address type after it has been created! please "+
			"recreate the service.[%s]->[%s],[%s]\n", old.AddressType, new.AddressType, new.LoadBalancerName)
		return nil, errors.New("alicloud: change loadbalancer address type after created is not support.")
	}
	return s.c.DescribeLoadBalancerAttribute(old.LoadBalancerId)
}

func (s *LoadBalancerClient) EnsureLoadBalanceDeleted(service *v1.Service) error {

	// need to save the resource version when deleted event
	err := s.SaveDeletedSVCResourceVersion(service)
	if err != nil {
		glog.V(2).Infof("alicloud: "+
			"failed to save deleted service resourceVersion, [%s] due to [%s] ", service.Name, err.Error())
	}
	ar := ExtractAnnotationRequest(service)
	if ar.Loadbalancerid != "" {
		glog.V(2).Infof("alicloud: user created loadbalancer[%s] will not be deleted by cloudprovider. service [%s]", ar.Loadbalancerid, service.Name)
		err := s.SaveDeletedSVCResourceVersion(service)
		if err != nil {
			glog.V(2).Infof("Failed to SaveDeletedSVCResourceVersion service %s due to %v ", service.Name, err)
		}
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
	ar := ExtractAnnotationRequest(service)
	args = &slb.CreateLoadBalancerArgs{
		AddressType:        ar.AddressType,
		InternetChargeType: ar.ChargeType,
		Bandwidth:          ar.Bandwidth,
		RegionId:           ar.Region,
	}
	if ar.SLBNetworkType != "classic" &&
		strings.Compare(string(ar.AddressType), string(slb.IntranetAddressType)) == 0 {

		glog.Infof("alicloud: intranet vpc "+
			"loadbalancer will be created. address type=%s, switchid=%s\n", ar.AddressType, vswitchid)
		args.VSwitchId = vswitchid
	}
	args.LoadBalancerName = fmt.Sprintf("kubernetes-%s", cloudprovider.GetLoadBalancerName(service))[0:16]
	return
}

const DEFAULT_SERVER_WEIGHT = 100

func (s *LoadBalancerClient) UpdateBackendServers(nodes []*v1.Node, lb *slb.LoadBalancerType) error {
	additions, deletions := []slb.BackendServerType{}, []string{}

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
		_, err := s.c.AddBackendServers(lb.LoadBalancerId, additions)
		if err != nil {
			return err
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
	_, err := s.c.RemoveBackendServers(lb.LoadBalancerId, deletions)
	return err
}

// check if the service exists in service definition
func checkIfSLBExistInService(service *v1.Service) (exists bool) {
	if service == nil ||
		len(service.Status.LoadBalancer.Ingress) == 0 {
		glog.V(2).Infof("alicloud: service %s doesn't have ingresses\n", service.Name)
		exists = false
	} else {
		glog.V(2).Infof("alicloud: service %s has ingresses=%v\n", service.Name, service.Status.LoadBalancer.Ingress)
		exists = true
	}

	return exists

}

// ensure service resource version properly and update last known resource version to the largest one, for now only keep create and delete behavior
func (s *LoadBalancerClient) EnsureSVCNotDeleted(service *v1.Service) error {
	if service == nil {
		return fmt.Errorf("alicloud:service is nil")
	}

	serviceUID := string(service.GetUID())
	keeper := GetLocalService()
	deleted := keeper.get(serviceUID)
	if deleted {
		glog.V(2).Infof("alicloud: service "+
			"%s's uid %v has been deleted, shouldn't be created again.\n", service.Name, serviceUID)
		return fmt.Errorf("alicloud: service "+
			"%s's uid %v has been deleted, shouldn't be created again.\n", service.Name, serviceUID)
	} else {
		glog.V(2).Infof("alicloud: service %s's uid %v "+
			"hasn't been deleted, first time to process, as expected.\n", service.Name, serviceUID)
	}

	return nil
}

// save the deleted service's uid
func (s *LoadBalancerClient) SaveDeletedSVCResourceVersion(service *v1.Service) error {
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
