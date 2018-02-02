package alicloud

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

type AnnotationRequest struct {
	Loadbalancerid string
	BackendLabel   string

	SSLPorts       string
	AddressType    slb.AddressType
	SLBNetworkType string

	ChargeType slb.InternetChargeType
	Region     common.Region
	Bandwidth  int
	CertID     string

	HealthCheck            slb.FlagType
	HealthCheckURI         string
	HealthCheckConnectPort int
	HealthyThreshold       int
	UnhealthyThreshold     int
	HealthCheckInterval    int

	HealthCheckConnectTimeout int                 // for tcp
	HealthCheckType           slb.HealthCheckType // for tcp, Type could be http tcp
	HealthCheckTimeout        int                 // for https and http
}

type LoadBalancerClient struct {
	c *slb.Client
	// known service resource version
	knownSVCResourceVersionMap map[string]uint64
}

func (s *LoadBalancerClient) findLoadBalancer(service *v1.Service) (bool, *slb.LoadBalancerType, error) {
	ar := ExtractAnnotationRequest(service)
	if ar.Loadbalancerid != "" {
		return s.findLoadBalancerByID(ar.Loadbalancerid, ar.Region)
	}
	return s.findLoadBalancerByName(cloudprovider.GetLoadBalancerName(service), ar.Region)
}

func (s *LoadBalancerClient) findLoadBalancerByID(lbid string, region common.Region) (bool, *slb.LoadBalancerType, error) {
	glog.V(2).Infof("Alicloud.findLoadBalancerByID(): lbid=%+v", lbid)
	lbs, err := s.c.DescribeLoadBalancers(
		&slb.DescribeLoadBalancersArgs{
			RegionId:       region,
			LoadBalancerId: lbid,
		},
	)
	glog.V(2).Infof("Alicloud.findLoadBalancerByID(): lbs=%+v, check err=%v", lbs, err)
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		return false, nil, fmt.Errorf("The specified Load Balancer ID %s can't be found", lbid)
	}
	if len(lbs) > 1 {
		glog.Warningf("Alicloud.GetLoadBalancerByID(): multiple loadbalancer returned with id=%s, "+
			"using the first one with IP=%s", lbid, lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(lbs[0].LoadBalancerId)
	glog.V(2).Infof("Alicloud.findLoadBalancerByID(): lb=%+v, check err=%v", lb, err)
	return err == nil, lb, err
}

func (s *LoadBalancerClient) findLoadBalancerByName(lbn string, region common.Region) (bool, *slb.LoadBalancerType, error) {
	glog.V(2).Infof("Alicloud.findLoadBalancerByName(): lbn=%s, region=%v", lbn, region)
	lbs, err := s.c.DescribeLoadBalancers(
		&slb.DescribeLoadBalancersArgs{
			RegionId:         region,
			LoadBalancerName: lbn,
		},
	)
	glog.V(2).Infof("Alicloud.findLoadBalancerByName(): lbs=%+v, check err=%v", lbs, err)
	if err != nil {
		return false, nil, err
	}

	if lbs == nil || len(lbs) == 0 {
		return false, nil, nil
	}
	if len(lbs) > 1 {
		glog.Warningf("Alicloud.GetLoadBalancerByName(): multiple loadbalancer returned with name=%s, "+
			"using the first one with IP=%s", lbn, lbs[0].Address)
	}
	lb, err := s.c.DescribeLoadBalancerAttribute(lbs[0].LoadBalancerId)
	glog.V(2).Infof("Alicloud.findLoadBalancerByName(): lb=%+v, check err=%v", lb, err)
	return err == nil, lb, err
}

func (s *LoadBalancerClient) EnsureLoadBalancer(service *v1.Service, nodes []*v1.Node, vswitchid string) (*slb.LoadBalancerType, error) {
	glog.V(2).Infof("Alicloud.EnsureLoadBalancer(): service details=%+v", service)

	exists, lb, err := s.findLoadBalancer(service)
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("Alicloud.EnsureLoadBalancer(): sync loadbalancer=%v, exist=%v\n", lb, exists)
	if !exists && checkIfSLBExistInService(service) {
		glog.V(2).Infof("Not able to find SLB named %s in aliyun openapi, but it's defined in service.loaderbalancer.ingress.\n", service.Name)
		return nil, fmt.Errorf("Not able to find SLB named %s in aliyun openapi, but it's defined in service.loaderbalancer.ingress.\n", service.Name)
	}

	// If need created, double check the resource version
	if !exists {
		err = s.EnsureSVCResourceVersion(service)
		if err != nil {
			glog.V(2).Infof("Alicloud.EnsureLoadBalancer(): EnsureSVCResourceVersion func can't work properly due to %+v, exit", err)
			os.Exit(1)
		}
	}

	ar := ExtractAnnotationRequest(service)
	opts := s.getLoadBalancerOpts(ar)
	if strings.Compare(string(opts.AddressType),
		string(slb.IntranetAddressType)) == 0 && ar.SLBNetworkType != "classic" {

		glog.Infof("Alicloud.EnsureLoadBalancer(): intranet VPC loadbalancer will be created. "+
			"addressType=%s, switch=%s\n", opts.AddressType, vswitchid)
		opts.VSwitchId = vswitchid
	}
	opts.LoadBalancerName = cloudprovider.GetLoadBalancerName(service)
	if !exists {
		glog.V(2).Infof("Alicloud.EnsureLoadBalancer(): loadbalancer dose not exist, "+
			"create new loadbalancer with option=%+v\n", opts)
		lbr, err := s.c.CreateLoadBalancer(opts)
		if err != nil {
			return nil, err
		}
		lb, err = s.c.DescribeLoadBalancerAttribute(lbr.LoadBalancerId)
		if err != nil {
			return nil, err
		}
	} else {
		// Todo: here we need to compare loadbalance
		if opts.InternetChargeType != lb.InternetChargeType {
			glog.Infof("Alicloud.EnsureLoadBalancer(): diff loadbalancer changes. "+
				"[InternetChargeType] changed from [%s] to [%s] , need to update loadbalancer:[%+v]\n",
				string(lb.InternetChargeType), string(opts.InternetChargeType), opts)
			if err := s.c.ModifyLoadBalancerInternetSpec(
				&slb.ModifyLoadBalancerInternetSpecArgs{
					LoadBalancerId:     lb.LoadBalancerId,
					InternetChargeType: opts.InternetChargeType,
					//Bandwidth:          opts.Bandwidth,
				}); err != nil {
				return nil, err
			}
		}
		if opts.AddressType != lb.AddressType {
			glog.Infof("Alicloud.EnsureLoadBalance(): diff loadbalancer changes. "+
				"[AddressType] changed from [%s] to [%s]. need to recreate loadbalancer!",
				lb.AddressType, opts.AddressType, opts.LoadBalancerName)
			// Can not modify AddressType.  We can only recreate it.
			if err := s.c.DeleteLoadBalancer(lb.LoadBalancerId); err != nil {
				return nil, err
			}
			lbc, err := s.c.CreateLoadBalancer(opts)
			if err != nil {
				return nil, err
			}
			lb, err = s.c.DescribeLoadBalancerAttribute(lbc.LoadBalancerId)
			if err != nil {
				return nil, err
			}
		}
	}

	if _, err := s.EnsureLoadBalancerListener(service, lb); err != nil {

		return nil, err
	}

	if !exists {
		err = s.KeepSVCResourceVersion(service)
		glog.V(2).Infof("Failed to KeepSVCResourceVersion service %s due to %v ", service.Name, err)
	}

	return s.EnsureBackendServer(service, nodes, lb)
}

func (s *LoadBalancerClient) UpdateLoadBalancer(service *v1.Service, nodes []*v1.Node) error {

	exists, lb, err := s.findLoadBalancer(service)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(fmt.Sprintf("the loadbalance you specified by name [%s] does not exist!", service.Name))
	}
	_, err = s.EnsureBackendServer(service, nodes, lb)
	return err
}

func (s *LoadBalancerClient) EnsureLoadBalancerListener(service *v1.Service, lb *slb.LoadBalancerType) (*slb.LoadBalancerType, error) {
	//ssl := service.Annotations["sec_ports"]
	additions, deletions, err := s.diffListeners(service, lb)
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("Alicloud.EnsureLoadBalancerListener(): verify addition [LoadBalancerListerners]=%+v.  deletion [LoadBalancerListerners]=%+v", additions, deletions)
	if len(deletions) > 0 {

		glog.Infof("Alicloud.EnsureLoadBalancerListener(): adding additional [LoadBalancerListerners]=%+v.", additions)
		for _, p := range deletions {
			// stop first
			// todo: here should retry for none runing status
			if err := s.c.StopLoadBalancerListener(lb.LoadBalancerId, p.Port); err != nil {
				return nil, err
			}
			// deal with port delete
			if err := s.c.DeleteLoadBalancerListener(lb.LoadBalancerId, p.Port); err != nil {
				return nil, err
			}
		}
	}
	if len(additions) > 0 {

		glog.Infof("Alicloud.EnsureLoadBalancerListener(): deleting [LoadBalancerListerners]=%+v", deletions)
		// deal with port add
		for _, p := range additions {
			if err := s.createListener(lb, p); err != nil {
				return nil, err
			}
			// todo : here should retry
			if err := s.c.StartLoadBalancerListener(lb.LoadBalancerId, p.Port); err != nil {
				return nil, err
			}
		}
	}
	return lb, nil
}

type PortListener struct {
	Port     int
	NodePort int
	Protocol string

	Bandwidth int

	Scheduler     slb.SchedulerType
	StickySession slb.FlagType
	CertID        string

	HealthCheck            slb.FlagType
	HealthCheckType        slb.HealthCheckType
	HealthCheckURI         string
	HealthCheckConnectPort int

	HealthyThreshold    int
	UnhealthyThreshold  int
	HealthCheckInterval int

	HealthCheckConnectTimeout int // for tcp
	HealthCheckTimeout        int // for https and http
}

// 1. Modify ListenPort would cause listener to be recreated
// 2. Modify NodePort would cause listener to be recreated
// 3. Modify Protocol would cause listener to be recreated
//
func (s *LoadBalancerClient) diffListeners(service *v1.Service, lb *slb.LoadBalancerType) (
	[]PortListener, []PortListener, error) {
	lp := lb.ListenerPortsAndProtocol.ListenerPortAndProtocol
	additions, deletions := []PortListener{}, []PortListener{}

	ar := ExtractAnnotationRequest(service)
	stickSession := slb.OffFlag
	// find additions
	for _, v1 := range service.Spec.Ports {
		found := false
		proto, err := transProtocol(serviceAnnotation(service, ServiceAnnotationLoadBalancerProtocolPort), &v1)
		if err != nil {
			return nil, nil, err
		}
		new := PortListener{
			Port:                   int(v1.Port),
			Protocol:               proto,
			NodePort:               int(v1.NodePort),
			Bandwidth:              ar.Bandwidth,
			HealthCheck:            ar.HealthCheck,
			StickySession:          stickSession,
			CertID:                 ar.CertID,
			HealthCheckType:        ar.HealthCheckType,
			HealthCheckConnectPort: ar.HealthCheckConnectPort,
			HealthCheckURI:         ar.HealthCheckURI,

			HealthyThreshold:          ar.HealthyThreshold,
			UnhealthyThreshold:        ar.UnhealthyThreshold,
			HealthCheckInterval:       ar.HealthCheckInterval,
			HealthCheckConnectTimeout: ar.HealthCheckConnectTimeout,
			HealthCheckTimeout:        ar.HealthCheckTimeout,
		}
		for _, v2 := range lp {
			if int64(v1.Port) == int64(v2.ListenerPort) {
				old, err := s.findPortListener(lb, v2.ListenerPort, v2.ListenerProtocol)
				if err != nil {
					return nil, nil, err
				}
				update := false
				if proto != v2.ListenerProtocol {
					update = true
					glog.V(2).Infof("Alicloud.diffListeners(): [%s], protocol changed from [%s] to [%s]\n", lb.LoadBalancerId, v2.ListenerProtocol, proto)
				}
				if int(v1.NodePort) != old.NodePort {
					update = true
					glog.V(2).Infof("Alicloud.diffListeners(): [%s], nodeport changed from [%d] to [%d]\n", lb.LoadBalancerId, old.NodePort, v1.NodePort)
				}

				if old.Bandwidth != ar.Bandwidth {
					update = true
					glog.V(2).Infof("Alicloud.diffListeners(): [%s], bandwidth changed from [%d] to [%d]\n", lb.LoadBalancerId, old.Bandwidth, ar.Bandwidth)
				}
				if old.CertID != ar.CertID && proto == "https" {
					update = true
					glog.V(2).Infof("Alicloud.diffListeners(): [%s], certid changed  from [%s] to [%s]\n", lb.LoadBalancerId, old.CertID, ar.CertID)
				}
				//if old.HealthCheck != ar.HealthCheck ||
				//	old.HealthCheckType != ar.HealthCheckType ||
				//	old.HealthCheckURI != ar.HealthCheckURI ||
				//	old.HealthCheckConnectPort != ar.HealthCheckConnectPort {
				//	update = true
				//	glog.V(2).Infof("Alicloud.diffListeners(): [%s] healthcheck changed \n", lb.LoadBalancerId)
				//}
				if update {
					deletions = append(deletions, old)
					additions = append(additions, new)
				}
				found = true
			}
		}
		if !found {
			additions = append(additions, new)
		}
	}

	// Find deletions
	for _, v1 := range lp {
		found := false
		for _, v2 := range service.Spec.Ports {
			if int64(v1.ListenerPort) == int64(v2.Port) {
				found = true
			}
		}
		if !found {
			deletions = append(deletions, PortListener{Port: v1.ListenerPort})
		}
	}

	return additions, deletions, nil
}

func (s *LoadBalancerClient) findPortListener(lb *slb.LoadBalancerType, port int, proto string) (PortListener, error) {
	switch proto {
	case "http":
		p, err := s.c.DescribeLoadBalancerHTTPListenerAttribute(lb.LoadBalancerId, port)
		if err != nil {
			return PortListener{}, err
		}
		return PortListener{
			Port:                   p.ListenerPort,
			NodePort:               p.BackendServerPort,
			Protocol:               proto,
			Bandwidth:              p.Bandwidth,
			HealthCheck:            p.HealthCheck,
			Scheduler:              p.Scheduler,
			StickySession:          p.StickySession,
			HealthCheckURI:         p.HealthCheckURI,
			HealthCheckConnectPort: p.HealthCheckConnectPort,

			HealthyThreshold:    p.HealthyThreshold,
			UnhealthyThreshold:  p.UnhealthyThreshold,
			HealthCheckInterval: p.HealthCheckInterval,
			HealthCheckTimeout:  p.HealthCheckTimeout,
		}, nil
	case "tcp":
		p, err := s.c.DescribeLoadBalancerTCPListenerAttribute(lb.LoadBalancerId, port)
		if err != nil {
			return PortListener{}, err
		}
		return PortListener{
			Port:      p.ListenerPort,
			NodePort:  p.BackendServerPort,
			Protocol:  proto,
			Bandwidth: p.Bandwidth,
			Scheduler: p.Scheduler,

			HealthyThreshold:          p.HealthyThreshold,
			UnhealthyThreshold:        p.UnhealthyThreshold,
			HealthCheckInterval:       p.HealthCheckInterval,
			HealthCheckConnectTimeout: p.HealthCheckConnectTimeout,
			HealthCheckTimeout:        p.HealthCheckConnectTimeout,
		}, nil
	case "https":
		p, err := s.c.DescribeLoadBalancerHTTPSListenerAttribute(lb.LoadBalancerId, port)
		if err != nil {
			return PortListener{}, err
		}
		return PortListener{
			Port:          p.ListenerPort,
			NodePort:      p.BackendServerPort,
			Protocol:      proto,
			Bandwidth:     p.Bandwidth,
			HealthCheck:   p.HealthCheck,
			Scheduler:     p.Scheduler,
			StickySession: p.StickySession,
			CertID:        p.ServerCertificateId,

			HealthCheckURI:         p.HealthCheckURI,
			HealthCheckConnectPort: p.HealthCheckConnectPort,

			HealthyThreshold:    p.HealthyThreshold,
			UnhealthyThreshold:  p.UnhealthyThreshold,
			HealthCheckInterval: p.HealthCheckInterval,
			HealthCheckTimeout:  p.HealthCheckTimeout,
		}, nil
	case "udp":
		p, err := s.c.DescribeLoadBalancerUDPListenerAttribute(lb.LoadBalancerId, port)
		if err != nil {
			return PortListener{}, err
		}
		return PortListener{
			Port:      p.ListenerPort,
			NodePort:  p.BackendServerPort,
			Protocol:  proto,
			Bandwidth: p.Bandwidth,
			Scheduler: p.Scheduler,

			HealthCheckConnectPort: p.HealthCheckConnectPort,

			HealthyThreshold:    p.HealthyThreshold,
			UnhealthyThreshold:  p.UnhealthyThreshold,
			HealthCheckInterval: p.HealthCheckInterval,
			HealthCheckTimeout:  p.HealthCheckConnectTimeout,
		}, nil
	}
	return PortListener{}, errors.New(fmt.Sprintf("protocol not match: %s", proto))
}

func (s *LoadBalancerClient) EnsureBackendServer(service *v1.Service, nodes []*v1.Node, lb *slb.LoadBalancerType) (*slb.LoadBalancerType, error) {

	additions, deletions := s.diffServers(nodes, lb)
	glog.V(2).Infof("Alicloud.EnsureBackendServer(): verify additional-backend-servers=[%+v],   removing-backend-servers=[%+v]", additions, deletions)
	if len(additions) > 0 {
		glog.Infof("Alicloud.EnsureBackendServer(): adding additional backend servers=[%+v]", additions)
		// deal with server add
		if _, err := s.c.AddBackendServers(lb.LoadBalancerId, additions); err != nil {

			return lb, err
		}
	}
	if len(deletions) > 0 {
		glog.Infof("Alicloud.EnsureBackendServer():  deleting removed backend servers=[%+v]", deletions)
		servers := []string{}
		for _, v := range deletions {
			servers = append(servers, v.ServerId)
		}
		if _, err := s.c.RemoveBackendServers(lb.LoadBalancerId, servers); err != nil {
			return lb, err
		}
	}
	return lb, nil
}

func (s *LoadBalancerClient) EnsureLoadBalanceDeleted(service *v1.Service) error {
	ar := ExtractAnnotationRequest(service)
	if ar.Loadbalancerid != "" {
		glog.V(2).Infof("It will not delete slb %s's load balancer id %s, because it's created by the user.", service.Name, ar.Loadbalancerid)
		err := s.KeepSVCResourceVersion(service)
		if err != nil {
			glog.V(2).Infof("Failed to KeepSVCResourceVersion service %s due to %v ", service.Name, err)
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
	err = s.KeepSVCResourceVersion(service)
	glog.V(2).Infof("Failed to KeepSVCResourceVersion service %s due to %v ", service.Name, err)
	return s.c.DeleteLoadBalancer(lb.LoadBalancerId)
}

func (s *LoadBalancerClient) EnsureHealthCheck(service *v1.Service, old *PortListener, new *PortListener) (*slb.LoadBalancerType, error) {

	return nil, nil
}

func (s *LoadBalancerClient) createListener(lb *slb.LoadBalancerType, pp PortListener) error {
	protocol := pp.Protocol

	if pp.HealthCheckConnectPort == MagicHealthCheckConnectPort {
		pp.HealthCheckConnectPort = 0
	}

	if protocol == "https" {
		lis := slb.CreateLoadBalancerHTTPSListenerArgs(
			slb.HTTPSListenerType{
				HTTPListenerType: slb.HTTPListenerType{
					LoadBalancerId:    lb.LoadBalancerId,
					ListenerPort:      pp.Port,
					BackendServerPort: pp.NodePort,
					//Health Check
					HealthCheck:   pp.HealthCheck,
					Bandwidth:     pp.Bandwidth,
					StickySession: pp.StickySession,

					HealthCheckURI:         pp.HealthCheckURI,
					HealthCheckConnectPort: pp.HealthCheckConnectPort,
					HealthyThreshold:       pp.HealthyThreshold,
					UnhealthyThreshold:     pp.UnhealthyThreshold,
					HealthCheckTimeout:     pp.HealthCheckTimeout,
					HealthCheckInterval:    pp.HealthCheckInterval,
				},
				ServerCertificateId: pp.CertID,
			},
		)
		if err := s.c.CreateLoadBalancerHTTPSListener(&lis); err != nil {
			return err
		}
	}
	if protocol == "http" {
		lis := slb.CreateLoadBalancerHTTPListenerArgs(
			slb.HTTPListenerType{
				LoadBalancerId:    lb.LoadBalancerId,
				ListenerPort:      pp.Port,
				BackendServerPort: pp.NodePort,
				//Health Check
				HealthCheck: pp.HealthCheck,

				Bandwidth:     pp.Bandwidth,
				StickySession: pp.StickySession,

				HealthCheckURI:         pp.HealthCheckURI,
				HealthCheckConnectPort: pp.HealthCheckConnectPort,
				HealthyThreshold:       pp.HealthyThreshold,
				UnhealthyThreshold:     pp.UnhealthyThreshold,
				HealthCheckTimeout:     pp.HealthCheckTimeout,
				HealthCheckInterval:    pp.HealthCheckInterval,
			})
		if err := s.c.CreateLoadBalancerHTTPListener(&lis); err != nil {

			return err
		}
	}
	if protocol == strings.ToLower(string(api.ProtocolTCP)) {
		if err := s.c.CreateLoadBalancerTCPListener(
			&slb.CreateLoadBalancerTCPListenerArgs{
				LoadBalancerId:    lb.LoadBalancerId,
				ListenerPort:      pp.Port,
				BackendServerPort: pp.NodePort,
				//Health Check
				Bandwidth: pp.Bandwidth,

				HealthCheckType:           pp.HealthCheckType,
				HealthCheckURI:            pp.HealthCheckURI,
				HealthCheckConnectPort:    pp.HealthCheckConnectPort,
				HealthyThreshold:          pp.HealthyThreshold,
				UnhealthyThreshold:        pp.UnhealthyThreshold,
				HealthCheckConnectTimeout: pp.HealthCheckConnectTimeout,
				HealthCheckInterval:       pp.HealthCheckInterval,
			}); err != nil {
			return err
		}
	}
	if protocol == strings.ToLower(string(api.ProtocolUDP)) {
		if err := s.c.CreateLoadBalancerUDPListener(
			&slb.CreateLoadBalancerUDPListenerArgs{
				LoadBalancerId:    lb.LoadBalancerId,
				ListenerPort:      pp.Port,
				BackendServerPort: pp.NodePort,
				//Health Check
				Bandwidth: pp.Bandwidth,

				HealthCheckConnectPort:    pp.HealthCheckConnectPort,
				HealthyThreshold:          pp.HealthyThreshold,
				UnhealthyThreshold:        pp.UnhealthyThreshold,
				HealthCheckConnectTimeout: pp.HealthCheckTimeout,
				HealthCheckInterval:       pp.HealthCheckInterval,
			}); err != nil {
			return err
		}
	}

	return nil
}

func (s *LoadBalancerClient) getLoadBalancerOpts(ar *AnnotationRequest) *slb.CreateLoadBalancerArgs {

	return &slb.CreateLoadBalancerArgs{
		AddressType:        ar.AddressType,
		InternetChargeType: ar.ChargeType,
		Bandwidth:          ar.Bandwidth,
		RegionId:           ar.Region,
	}
}

const DEFAULT_SERVER_WEIGHT = 100

func (s *LoadBalancerClient) diffServers(nodes []*v1.Node, lb *slb.LoadBalancerType) ([]slb.BackendServerType, []slb.BackendServerType) {
	additions, deletions := []slb.BackendServerType{}, []slb.BackendServerType{}
	for _, n1 := range nodes {
		found := false
		_, id, err := nodeid(types.NodeName(n1.Spec.ExternalID))
		for _, n2 := range lb.BackendServers.BackendServer {
			if err != nil {
				glog.Errorf("Alicloud.diffServers(): node externalid=%s is not in the correct form, expect regionid.instanceid. skip add op", n1.Spec.ExternalID)
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
	for _, n1 := range lb.BackendServers.BackendServer {
		found := false
		for _, n2 := range nodes {
			_, id, err := nodeid(types.NodeName(n2.Spec.ExternalID))
			if err != nil {
				glog.Errorf("Alicloud.diffServers(): node externalid=%s is not in the correct form, expect regionid.instanceid.. skip delete op", n2.Spec.ExternalID)
				continue
			}
			if n1.ServerId == string(id) {
				found = true
				break
			}
		}
		if !found {
			deletions = append(deletions, n1)
		}
	}
	return additions, deletions
}

func transProtocol(annotation string, port *v1.ServicePort) (string, error) {
	if annotation != "" {
		for _, v := range strings.Split(annotation, ",") {
			pp := strings.Split(v, ":")
			if len(pp) < 2 {
				return "", errors.New(fmt.Sprintf("port and protocol format must be like 'https:443' with colon separated. got=[%+v]", pp))
			}

			if pp[0] != "http" &&
				pp[0] != "tcp" &&
				pp[0] != "https" &&
				pp[0] != "udp" {
				return "", errors.New(fmt.Sprintf("port protocol format must be either [http|https|tcp|udp], protocol not supported wit [%s]\n", pp[0]))
			}

			if pp[1] == fmt.Sprintf("%d", port.Port) {
				return pp[0], nil
			}
		}
		return strings.ToLower(string(port.Protocol)), nil
	}

	return strings.ToLower(string(port.Protocol)), nil
}

// check if the service exists in service definition
func checkIfSLBExistInService(service *v1.Service) (exists bool) {
	if service == nil ||
		len(service.Status.LoadBalancer.Ingress) == 0 {
		glog.V(2).Infof("checkIfSLBExistInService(): service %s doesn't have ingresses\n", service.Name)
		exists = false
	} else {
		glog.V(2).Infof("checkIfSLBExistInService(): service %s has ingresses=%v\n", service.Name, service.Status.LoadBalancer.Ingress)
		exists = true
	}

	return exists

}

// ensure service resource version properly and update last known resource version to the largest one, for now only keep create and delete behavior
func (s *LoadBalancerClient) EnsureSVCResourceVersion(service *v1.Service) error {
	if service == nil {
		return fmt.Errorf("EnsureSVCResourceVersion:service is nil")
	}
	if s.knownSVCResourceVersionMap == nil {
		s.knownSVCResourceVersionMap = map[string]uint64{}
	}

	serviceUID := string(service.GetUID())
	currentVersion, err := strconv.ParseUint(service.ResourceVersion, 10, 64)
	if err != nil {
		glog.V(2).Infof("EnsureSVCResourceVersion(): service %s's ResourceVersion %v has been parsed failed due to %v", service.Name, service.ResourceVersion, err)
		return err
	}

	if lastKnownVersion, ok := s.knownSVCResourceVersionMap[serviceUID]; ok {
		if currentVersion > lastKnownVersion {
			glog.V(2).Infof("EnsureSVCResourceVersion(): service %s's uid %v and ResourceVersion %v > lastKnownResourceVersion %v, as expected.\n", service.Name,
				serviceUID,
				service.ResourceVersion,
				lastKnownVersion)
		} else {
			glog.V(2).Infof("EnsureSVCResourceVersion(): service %s's uid %v and ResourceVersion %v < lastKnownResourceVersion %v, it means the resourceVersion has been processed, not as expected.\n", service.Name,
				serviceUID,
				service.ResourceVersion,
				lastKnownVersion)
			return fmt.Errorf("EnsureSVCResourceVersion(): service %s's uid %v and ResourceVersion %v < lastKnownResourceVersion %v, it means the resourceVersion has been processed, not as expected.\n", service.Name,
				serviceUID,
				service.ResourceVersion,
				lastKnownVersion)
		}
	} else {
		glog.V(2).Infof("EnsureSVCResourceVersion(): service %s's uid %v is not kept, first time to process, as expected.\n", service.Name, serviceUID)
	}

	return nil
}

func (s *LoadBalancerClient) KeepSVCResourceVersion(service *v1.Service) error {
	if service == nil {
		return fmt.Errorf("KeepSVCResourceVersion:service is nil")
	}
	if s.knownSVCResourceVersionMap == nil {
		s.knownSVCResourceVersionMap = map[string]uint64{}
	}

	serviceUID := string(service.GetUID())
	currentVersion, err := strconv.ParseUint(service.ResourceVersion, 10, 64)
	if err != nil {
		glog.V(2).Infof("KeepSVCResourceVersion(): service %s's ResourceVersion %v has been parsed failed due to %v", service.Name, service.ResourceVersion, err)
		return err
	}

	glog.V(2).Infof("KeepSVCResourceVersion(): service %s's uid %s and ResourceVersion %v has been kept.", service.Name, serviceUID, service.ResourceVersion)
	s.knownSVCResourceVersionMap[serviceUID] = currentVersion
	return nil
}
