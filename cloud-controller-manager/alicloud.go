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
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/route"
	"k8s.io/kubernetes/pkg/cloudprovider"
	ctrlclient "k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/version"
	"math/rand"
	"net"
	"strings"
	"time"
)

// ProviderName is the name of this cloud provider.
const ProviderName = "alicloud"

var CLUSTER_ID = "clusterid"

// KUBERNETES_ALICLOUD_IDENTITY is for statistic purpose.
var KUBERNETES_ALICLOUD_IDENTITY = fmt.Sprintf("Kubernetes.Alicloud/%s", version.Get().String())

// Cloud is an implementation of Interface, LoadBalancer and Instances for Alicloud Services.
type Cloud struct {
	climgr *ClientMgr

	cfg    *CloudConfig
	region common.Region
	vpcID  string
	cid    string
}

var (
	DEFAULT_CHARGE_TYPE  = common.PayByTraffic
	DEFAULT_BANDWIDTH    = 100
	DEFAULT_ADDRESS_TYPE = slb.InternetAddressType

	// DEFAULT_REGION should be override in cloud initialize.
	DEFAULT_REGION = common.Hangzhou
)

// CloudConfig wraps the settings for the Alicloud provider.
type CloudConfig struct {
	Global struct {
		KubernetesClusterTag string
		UID                  string `json:"uid"`
		VpcID                string `json:"vpcid"`
		Region               string `json:"region"`
		ZoneID               string `json:"zoneid"`
		VswitchID            string `json:"vswitchid"`
		ClusterID            string `json:"clusterID"`
		RouteTableIDS        string `json:"routeTableIDs"`

		AccessKeyID     string `json:"accessKeyID"`
		AccessKeySecret string `json:"accessKeySecret"`
	}
}

var cfg CloudConfig

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName,
		func(config io.Reader) (cloudprovider.Interface, error) {
			var (
				keyid     = ""
				keysecret = ""
				rtableids = ""
			)
			if config != nil {
				if err := json.NewDecoder(config).Decode(&cfg); err != nil {
					return nil, err
				}
				if cfg.Global.AccessKeyID != "" && cfg.Global.AccessKeySecret != "" {
					key, err := b64.StdEncoding.DecodeString(cfg.Global.AccessKeyID)
					if err != nil {
						return nil, err
					}
					keyid = string(key)
					secret, err := b64.StdEncoding.DecodeString(cfg.Global.AccessKeySecret)
					if err != nil {
						return nil, err
					}
					keysecret = string(secret)
					glog.V(2).Infof("Alicloud: Try Accesskey and AccessKeySecret from config file.")
				}
				if cfg.Global.ClusterID != "" {
					CLUSTER_ID = cfg.Global.ClusterID
					glog.Infof("use clusterid %s", CLUSTER_ID)
				}

				if cfg.Global.RouteTableIDS != "" {
					rtableids = cfg.Global.RouteTableIDS
				}
			}
			if keyid == "" || keysecret == "" {
				glog.V(2).Infof("cloud config does not have keyid and keysecret . try environment ACCESS_KEY_ID ACCESS_KEY_SECRET")
				keyid = os.Getenv("ACCESS_KEY_ID")
				keysecret = os.Getenv("ACCESS_KEY_SECRET")
			}
			mgr, err := NewClientMgr(keyid, keysecret)
			if err != nil {
				return nil, err
			}
			return newAliCloud(mgr, rtableids)
		})
}

func newAliCloud(mgr *ClientMgr, rtableids string) (*Cloud, error) {

	region, err := mgr.MetaData().Region()
	if err != nil {
		return nil, errors.New("please provide region in " +
			"Alicloud configuration file or make sure your ECS is under VPC network")
	}
	DEFAULT_REGION = common.Region(region)

	vpc, err := mgr.MetaData().VpcID()
	if err != nil {
		return nil, fmt.Errorf("Alicloud: error get vpcid. %s\n", err.Error())
	}
	glog.Infof("Using vpc region: region=%s, vpcid=%s", region, vpc)
	err = mgr.Routes().WithVPC(vpc, rtableids)
	if err != nil {
		return nil, fmt.Errorf("set vpc info error: %s", err.Error())
	}
	return &Cloud{climgr: mgr, region: common.Region(region), vpcID: vpc}, nil
}

// Initialize passes a Kubernetes clientBuilder interface to the cloud provider
func (c *Cloud) Initialize(builder ctrlclient.ControllerClientBuilder) {
	if !route.Options.ConfigCloudRoutes {
		return
	}

	cidr := route.Options.ClusterCIDR
	if len(strings.TrimSpace(cidr)) == 0 {
		panic(fmt.Sprintf("ivalid cluster CIDR %s", cidr))
	}
	_, cidrc, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(fmt.Sprintf("Unsuccessful parsing of cluster CIDR %v: %v", cidr, err))
	}

	shared := informers.NewSharedInformerFactory(
		builder.ClientOrDie("shared-informers"), syncPeriod())
	clusterid := ""
	if c.cfg != nil {
		clusterid = c.cfg.Global.KubernetesClusterTag
	}
	ctrl, err := route.New(c,
		builder.ClientOrDie(route.ROUTE_CONTROLLER),
		shared.Core().V1().Nodes(),
		clusterid, cidrc)
	if err != nil {
		panic(fmt.Sprintf("unable to initialize route controller, %s", err.Error()))
	}
	var stop <-chan struct{}
	go ctrl.Run(stop, route.Options.RouteReconciliationPeriod.Duration)
	time.Sleep(wait.Jitter(route.Options.ControllerStartInterval.Duration, 1.0))
	shared.Start(stop)
	glog.Infof("route controller started.")
}

func syncPeriod() time.Duration {
	return time.Duration(float64(route.Options.MinResyncPeriod.Nanoseconds()) * (rand.Float64() + 1))
}

// TODO: Break this up into different interfaces (LB, etc) when we have more than one type of service
// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (c *Cloud) GetLoadBalancer(clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {

	exists, lb, err := c.climgr.LoadBalancers().findLoadBalancer(service)

	if err != nil || !exists {
		return nil, exists, err
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{{IP: lb.Address}}}, true, nil
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (c *Cloud) EnsureLoadBalancer(clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {

	glog.V(2).Infof("Alicloud.EnsureLoadBalancer(%v, %s/%s, %v, %v)",
		clusterName, service.Namespace, service.Name, c.region, NodeList(nodes))
	ns, err := c.fileOutNode(nodes, service)
	if err != nil {
		return nil, err
	}
	glog.V(5).Infof("alicloud: ensure loadbalancer with final nodes list , %v\n", NodeList(ns))
	//if service.Spec.SessionAffinity != v1.ServiceAffinityNone {
	//	// Does not support SessionAffinity
	//	return nil, fmt.Errorf("unsupported load balancer affinity: %v", service.Spec.SessionAffinity)
	//}
	if len(service.Spec.Ports) == 0 {
		return nil, fmt.Errorf("requested load balancer with no ports")
	}
	if service.Spec.LoadBalancerIP != "" {
		return nil, fmt.Errorf("LoadBalancerIP cannot be specified for Alicloud SLB")
	}
	vswitchid := ""
	if len(ns) <= 0 {
		var err error
		vswitchid, err = c.climgr.MetaData().VswitchID()
		if err != nil {
			return nil, err
		}
		glog.V(2).Infof("alicloud: current vswitchid=%s\n", vswitchid)
		if vswitchid == "" {
			glog.Warningf("alicloud: can not find vswitch id, this will prevent you " +
				"from creating VPC intranet SLB. But classic LB is still available.")
		}
	} else {
		for _, v := range ns {
			i, err := c.climgr.Instances().findInstanceByProviderID(v.Spec.ProviderID)
			if err != nil {
				return nil, err
			}
			vswitchid = i.VpcAttributes.VSwitchId
			break
		}
	}

	lb, err := c.climgr.LoadBalancers().EnsureLoadBalancer(service, ns, vswitchid)
	if err != nil {
		return nil, err
	}
	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{
				IP: lb.Address,
			},
		},
	}, nil
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (c *Cloud) UpdateLoadBalancer(clusterName string, service *v1.Service, nodes []*v1.Node) error {
	glog.V(2).Infof("Alicloud.UpdateLoadBalancer(%v, %v, %v, %v, %v, %v, %v)",
		clusterName, service.Namespace, service.Name, c.region, service.Spec.LoadBalancerIP, service.Spec.Ports, NodeList(nodes))
	ns, err := c.fileOutNode(nodes, service)
	if err != nil {
		return err
	}
	return c.climgr.LoadBalancers().UpdateLoadBalancer(service, ns, true)
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (c *Cloud) EnsureLoadBalancerDeleted(clusterName string, service *v1.Service) error {
	glog.V(2).Infof("Alicloud.EnsureLoadBalancerDeleted(%v, %v, %v, %v, %v, %v)",
		clusterName, service.Namespace, service.Name, c.region, service.Spec.LoadBalancerIP, service.Spec.Ports)
	return c.climgr.LoadBalancers().EnsureLoadBalanceDeleted(service)
}

// NodeAddresses returns the addresses of the specified instance.
// TODO(roberthbailey): This currently is only used in such a way that it
// returns the address of the calling instance. We should do a rename to
// make this clearer.
func (c *Cloud) NodeAddresses(name types.NodeName) ([]v1.NodeAddress, error) {
	glog.V(2).Infof("Alicloud.NodeAddresses(\"%s\")", name)

	return c.climgr.Instances().findAddressByNodeName(name)
}

// InstanceTypeByProviderID returns the cloudprovider instance type of the node with the specified unique providerID
// This method will not be called from the node that is requesting this ID. i.e. metadata service
// and other local methods cannot be used here
func (c *Cloud) InstanceTypeByProviderID(providerID string) (string, error) {
	glog.V(5).Infof("Alicloud.InstanceTypeByProviderID(\"%s\")", providerID)
	ins, err := c.climgr.Instances().findInstanceByProviderID(providerID)
	if err == nil {
		return ins.InstanceType, nil
	}
	return "", err
}

// NodeAddressesByProviderID returns the node addresses of an instances with the specified unique providerID
// This method will not be called from the node that is requesting this ID. i.e. metadata service
// and other local methods cannot be used here
func (c *Cloud) NodeAddressesByProviderID(providerID string) ([]v1.NodeAddress, error) {
	glog.V(5).Infof("Alicloud.NodeAddressesByProviderID(\"%s\")", providerID)
	return c.climgr.Instances().findAddressByProviderID(providerID)
}

// ExternalID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (c *Cloud) ExternalID(nodeName types.NodeName) (string, error) {
	glog.V(5).Infof("Alicloud.ExternalID(\"%s\")", nodeName)
	instance, err := c.climgr.Instances().findInstanceByNodeName(nodeName)
	if err != nil {
		return "", err
	}
	return instance.InstanceId, nil
}

// InstanceID returns the cloud provider ID of the node with the specified NodeName.
func (c *Cloud) InstanceID(nodeName types.NodeName) (string, error) {
	glog.V(5).Infof("Alicloud.InstanceID(\"%s\")", nodeName)
	instance, err := c.climgr.Instances().findInstanceByNodeName(nodeName)
	if err != nil {
		return "", err
	}
	return instance.InstanceId, nil
}

// InstanceType returns the type of the specified instance.
func (c *Cloud) InstanceType(name types.NodeName) (string, error) {
	glog.V(5).Infof("Alicloud.InstanceType(\"%s\")", name)
	instance, err := c.climgr.Instances().findInstanceByNodeName(name)
	if err != nil {
		return "", err
	}
	return instance.InstanceType, nil
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (c *Cloud) AddSSHKeyToAllInstances(user string, keyData []byte) error {
	return errors.New("Alicloud.AddSSHKeyToAllInstances() is not implemented")
}

// CurrentNodeName returns the name of the node we are currently running on
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (c *Cloud) CurrentNodeName(hostname string) (types.NodeName, error) {
	nodeName, err := c.climgr.MetaData().InstanceID()
	if err != nil {
		return "", err
	}
	region, err := c.climgr.MetaData().Region()
	if err != nil {
		return "", err
	}
	glog.V(2).Infof("Alicloud.CurrentNodeName(\"%s\")", nodeName)
	return types.NodeName(fmt.Sprintf("%s.%s", region, nodeName)), nil
}

// InstanceExistsByProviderID returns true if the instance for the given provider id still is running.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
func (c *Cloud) InstanceExistsByProviderID(providerID string) (bool, error) {
	_, err := c.climgr.Instances().findInstanceByProviderID(providerID)
	if err == cloudprovider.InstanceNotFound {
		glog.V(2).Infof("Alicloud.InstanceExistsByProviderID(\"%s\") message=[%s]", providerID, err.Error())
		return false, err
	}
	return true, err
}
func (c *Cloud) RouteTables(clusterName string) ([]string, error) {
	return c.climgr.Routes().RouteTables()
}

// ListRoutes lists all managed routes that belong to the specified clusterName
func (c *Cloud) ListRoutes(clusterName string, tableid string) ([]*cloudprovider.Route, error) {
	glog.V(5).Infof("alicloud: ListRoutes \n")

	return c.climgr.Routes().ListRoutes(tableid)
}

// CreateRoute creates the described managed route
// route.Name will be ignored, although the cloud-provider may use nameHint
// to create a more user-meaningful name.
func (c *Cloud) CreateRoute(clusterName string, nameHint string, tableid string, route *cloudprovider.Route) error {
	glog.V(2).Infof("Alicloud.CreateRoute(\"%s, %+v\")", clusterName, route)
	ins, err := c.climgr.Instances().findInstanceByProviderID(string(route.TargetNode))
	if err != nil {
		return err
	}
	cRoute := &cloudprovider.Route{
		Name:            fmt.Sprintf("%s.%s", ins.RegionId, ins.InstanceId),
		DestinationCIDR: route.DestinationCIDR,
		TargetNode:      types.NodeName(ins.InstanceId),
	}
	return c.climgr.Routes().CreateRoute(tableid, cRoute, ins.RegionId, ins.VpcAttributes.VpcId)
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
func (c *Cloud) DeleteRoute(clusterName string, tableid string, route *cloudprovider.Route) error {
	glog.V(2).Infof("Alicloud.DeleteRoute(\"%s, %+v\")", clusterName, route)
	ins, err := c.climgr.Instances().findInstanceByProviderID(string(route.TargetNode))
	if err != nil {
		return err
	}
	cRoute := &cloudprovider.Route{
		Name:            route.Name,
		DestinationCIDR: route.DestinationCIDR,
		TargetNode:      types.NodeName(ins.InstanceId),
	}
	return c.climgr.Routes().DeleteRoute(tableid, cRoute, ins.RegionId, ins.VpcAttributes.VpcId)
}

// GetZone returns the Zone containing the current failure zone and locality region that the program is running in
func (c *Cloud) GetZone() (cloudprovider.Zone, error) {
	if cfg.Global.ZoneID != "" && cfg.Global.Region != "" {
		return cloudprovider.Zone{
			Region:        cfg.Global.Region,
			FailureDomain: cfg.Global.ZoneID,
		}, nil
	}
	host, err := c.climgr.MetaData().InstanceID()
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Alicloud.GetZone(): error execute c.meta.InstanceID(). message=[%s]", err.Error())
	}
	region, err := c.climgr.MetaData().Region()
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Alicloud.GetZone(): error execute c.meta.Region(). message=[%s]", err.Error())
	}
	i, err := c.climgr.Instances().findInstanceByProviderID(fmt.Sprintf("%s.%s", region, host))
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Alicloud.GetZone(): error execute findInstanceByProviderID(). message=[%s]", err.Error())
	}
	return cloudprovider.Zone{
		Region:        string(c.region),
		FailureDomain: i.ZoneId,
	}, nil
}

// GetZoneByNodeName returns the Zone containing the current zone and locality region of the node specified by node name
// This method is particularly used in the context of external cloud providers where node initialization must be down
// outside the kubelets.
func (c *Cloud) GetZoneByNodeName(nodeName types.NodeName) (cloudprovider.Zone, error) {

	i, err := c.climgr.Instances().findInstanceByNodeName(nodeName)
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Alicloud.GetZoneByNodeName(): error execute findInstanceByNode(). message=[%s]", err.Error())
	}
	return cloudprovider.Zone{
		Region:        string(c.region),
		FailureDomain: i.ZoneId,
	}, nil
}

// GetZoneByProviderID returns the Zone containing the current zone and locality region of the node specified by providerId
// This method is particularly used in the context of external cloud providers where node initialization must be down
// outside the kubelets.
func (c *Cloud) GetZoneByProviderID(providerID string) (cloudprovider.Zone, error) {
	i, err := c.climgr.Instances().findInstanceByProviderID(providerID)
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Alicloud.GetZoneByProviderID(), error execute findInstanceByNode(). message=[%s]", err.Error())
	}
	return cloudprovider.Zone{
		Region:        string(c.region),
		FailureDomain: i.ZoneId,
	}, nil
}

// ListClusters lists the names of the available clusters.
func (c *Cloud) ListClusters() ([]string, error) {
	return nil, errors.New("Alicloud.ListClusters() is not implemented")
}

// Master gets back the address (either DNS name or IP address) of the master node for the cluster.
func (c *Cloud) Master(clusterName string) (string, error) {
	return "", errors.New("Alicloud.ListClusters is not implemented")
}

// Clusters returns the list of clusters.
func (c *Cloud) Clusters() (cloudprovider.Clusters, bool) { return nil, false }

// ProviderName returns the cloud provider ID.
func (c *Cloud) ProviderName() string { return ProviderName }

// ScrubDNS filters DNS settings for pods.
func (c *Cloud) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nameservers, searches
}

// LoadBalancer returns an implementation of LoadBalancer for Alicloud Services.
func (c *Cloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) { return c, true }

// Instances returns an implementation of Instances for Alicloud Services.
func (c *Cloud) Instances() (cloudprovider.Instances, bool) { return c, true }

// Zones returns an implementation of Zones for Alicloud Services.
func (c *Cloud) Zones() (cloudprovider.Zones, bool) { return c, true }

// Routes returns an implementation of Routes for Alicloud Services.
func (c *Cloud) Routes() (cloudprovider.Routes, bool) { return nil, false }

// HasClusterID returns true if a ClusterID is required and set
func (c *Cloud) HasClusterID() bool { return CLUSTER_ID != "clusterid" }

//
func (c *Cloud) fileOutNode(nodes []*v1.Node, service *v1.Service) ([]*v1.Node, error) {
	ar, _ := ExtractAnnotationRequest(service)
	return c.climgr.Instances().filterOutByLabel(nodes, ar.BackendLabel)
}
