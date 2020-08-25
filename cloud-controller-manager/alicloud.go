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
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"
	"io"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/cloud-provider"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/node"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/route"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"k8s.io/klog"
	controller "k8s.io/kube-aggregator/pkg/controllers"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

// ProviderName is the name of this cloud provider.
const (
	ProviderName = "alicloud"
)

// CLUSTER_ID default cluster id if it is not specified.
var CLUSTER_ID = "clusterid"

// KUBERNETES_ALICLOUD_IDENTITY is for statistic purpose.
var KUBERNETES_ALICLOUD_IDENTITY = fmt.Sprintf("Kubernetes.Alicloud/%s", Version)

// cloud is an implementation of Interface, LoadBalancer and Instances for Alicloud Services.
type Cloud struct {
	climgr *ClientMgr

	cfg    *CloudConfig
	region common.Region
	vpcID  string
	//cid      string
	ifactory informers.SharedInformerFactory
}

var (
	// DEFAULT_CHARGE_TYPE default charge type
	DEFAULT_CHARGE_TYPE = common.PayByTraffic

	// DEFAULT_BANDWIDTH default bandwidth
	DEFAULT_BANDWIDTH = 100

	// DEFAULT_ADDRESS_TYPE default address type
	DEFAULT_ADDRESS_TYPE = slb.InternetAddressType

	DEFAULT_NODE_MONITOR_PERIOD = 120 * time.Second

	DEFAULT_NODE_ADDR_SYNC_PERIOD = 240 * time.Second

	// DEFAULT_REGION should be override in cloud initialize.
	DEFAULT_REGION = common.Hangzhou
)

// CloudConfig wraps the settings for the Alicloud provider.
type CloudConfig struct {
	Global struct {
		KubernetesClusterTag string `json:"kubernetesClusterTag"`
		NodeMonitorPeriod    int64  `json:"nodeMonitorPeriod"`
		NodeAddrSyncPeriod   int64  `json:"nodeAddrSyncPeriod"`
		UID                  string `json:"uid"`
		VpcID                string `json:"vpcid"`
		Region               string `json:"region"`
		ZoneID               string `json:"zoneid"`
		VswitchID            string `json:"vswitchid"`
		ClusterID            string `json:"clusterID"`
		RouteTableIDS        string `json:"routeTableIDs"`

		DisablePublicSLB bool `json:"disablePublicSLB"`

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
					klog.V(2).Infof("Alicloud: Try Accesskey and AccessKeySecret from config file.")
				}
				if cfg.Global.ClusterID != "" {
					CLUSTER_ID = cfg.Global.ClusterID
					klog.Infof("use clusterid %s", CLUSTER_ID)
				}

				if cfg.Global.RouteTableIDS != "" {
					rtableids = cfg.Global.RouteTableIDS
				}
			}
			if keyid == "" || keysecret == "" {
				klog.V(2).Infof("cloud config does not have keyid and keysecret . try environment ACCESS_KEY_ID ACCESS_KEY_SECRET")
				keyid = os.Getenv("ACCESS_KEY_ID")
				keysecret = os.Getenv("ACCESS_KEY_SECRET")
			}
			mgr, err := NewClientMgr(keyid, keysecret)
			if err != nil {
				return nil, err
			}
			// wait for client initialized
			err = mgr.Start(RefreshToken)
			if err != nil {
				panic(fmt.Sprintf("token not ready %s", err.Error()))
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
	klog.Infof("Using vpc region: region=%s, vpcid=%s", region, vpc)
	err = mgr.Routes().WithVPC(context.Background(), vpc, rtableids)
	if err != nil {
		return nil, fmt.Errorf("set vpc info error: %s", err.Error())
	}
	return &Cloud{
		climgr: mgr,
		region: common.Region(region),
		vpcID:  vpc,
		cfg:    &cfg,
	}, nil
}

// Initialize passes a Kubernetes clientBuilder interface to the cloud provider
func (c *Cloud) Initialize(builder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
	shared := informers.NewSharedInformerFactory(
		builder.ClientOrDie("shared-informers"), syncPeriod())
	if route.Options.ConfigCloudRoutes {
		cidr := route.Options.ClusterCIDR
		if len(strings.TrimSpace(cidr)) == 0 {
			panic(fmt.Sprintf("ivalid cluster CIDR %s", cidr))
		}
		_, cidrc, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("Unsuccessful parsing of cluster CIDR %v: %v", cidr, err))
		}

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
		go ctrl.Run(stop, route.Options.RouteReconciliationPeriod.Duration)
		klog.Infof("route controller started.")
	}

	func() {
		nodeMonitorPeriod := DEFAULT_NODE_MONITOR_PERIOD
		nodeAddrSyncPeriod := DEFAULT_NODE_ADDR_SYNC_PERIOD
		if c.cfg != nil &&
			c.cfg.Global.NodeMonitorPeriod != 0 {
			nodeMonitorPeriod = time.Duration(cfg.Global.NodeMonitorPeriod * int64(time.Second))
		}
		if c.cfg != nil &&
			c.cfg.Global.NodeAddrSyncPeriod != 0 {
			nodeAddrSyncPeriod = time.Duration(cfg.Global.NodeAddrSyncPeriod * int64(time.Second))
		}
		// run node controller
		nctrl := node.NewCloudNodeController(
			shared.Core().V1().Nodes(),
			builder.ClientOrDie(node.NODE_CONTROLLER),
			c,
			// delete node monitor
			nodeMonitorPeriod,
			// node address updator
			nodeAddrSyncPeriod,
		)
		go nctrl.Run(stop)
	}()
	inform := shared.Core().V1().Endpoints().Informer()
	shared.Start(stop)
	if !controller.WaitForCacheSync(
		"service", nil, inform.HasSynced,
	) {
		klog.Error("endpoints cache has not been syncd")
		return
	}
	c.ifactory = shared
}

func syncPeriod() time.Duration {
	return time.Duration(float64(route.Options.MinResyncPeriod.Nanoseconds()) * (rand.Float64() + 1))
}

// GetLoadBalancerName returns the name of the load balancer. Implementations must treat the
// *v1.Service parameter as read-only and not modify it.
func (c *Cloud) GetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) string {
	return ""
}

// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
// Implementations must treat the *v1.svc parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
// TODO: Break this up into different interfaces (LB, etc) when we have more than one type of service
func (c *Cloud) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {

	exists, lb, err := c.climgr.LoadBalancers().FindLoadBalancer(ctx, service)

	if err != nil || !exists {
		return nil, exists, err
	}

	zone, record, exists, err := c.climgr.PrivateZones().findExactRecordByService(ctx, service, lb.Address, lb.AddressIPVersion)
	if err != nil || !exists {
		return nil, true, err
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{{
			IP:       lb.Address,
			Hostname: getHostName(zone, record),
		}}}, true, nil
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.svc and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (c *Cloud) EnsureLoadBalancer(
	ctx context.Context,
	clusterName string,
	service *v1.Service,
	nodes []*v1.Node,
) (*v1.LoadBalancerStatus, error) {

	klog.V(2).Infof("Alicloud.EnsureLoadBalancer(%v, %s/%s, %v, %v)",
		clusterName, service.Namespace, service.Name, c.region, NodeList(nodes))
	defaulted, _ := ExtractAnnotationRequest(service)
	if defaulted.AddressType == slb.InternetAddressType {
		if c.cfg != nil && c.cfg.Global.DisablePublicSLB {
			return nil, fmt.Errorf("PublicAddress SLB is Not allowed")
		}
	}

	ns, err := c.fileOutNode(nodes, service)
	if err != nil {
		return nil, err
	}

	if len(service.Spec.Ports) == 0 {
		return nil, fmt.Errorf("requested load balancer with no ports")
	}
	vswitchid := defaulted.VswitchID
	if vswitchid == "" {
		var err error
		vswitchid, err = c.climgr.MetaData().VswitchID()
		if err != nil {
			return nil, fmt.Errorf("can not obtain vswitchid %s", err)
		}
		if vswitchid == "" {
			klog.Warningf("vswitch id not found, vpc intranet slb creation would fail")
		}
	}
	// set up endpoints
	eps, err := c.ifactory.
		Core().V1().
		Endpoints().
		Lister().
		Endpoints(
			service.Namespace,
		).Get(service.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// compatible with nil endpoint
			klog.Warningf("get available endpoints when EnsureLoadBalancer: %s", err.Error())
		} else {
			// avoid removing existing backends of SLB when getting endpoint error
			return nil, fmt.Errorf("get available endpoints when EnsureLoadBalancer: %s", err.Error())
		}
	}
	LogSubsetInfo(eps, "api")

	backends := &EndpointWithENI{
		LocalMode:      ServiceModeLocal(service),
		Endpoints:      eps,
		Nodes:          ns,
		BackendTypeENI: utils.IsENIBackendType(service),
	}

	utils.Logf(service, "using vswitch id=%s", vswitchid)

	// EnsureLoadBalancer with EndpointWithENI
	lb, err := c.climgr.
		LoadBalancers().
		EnsureLoadBalancer(
			ctx, service, backends, vswitchid,
		)
	if err != nil {
		return nil, err
	}

	status := &v1.LoadBalancerStatus{}

	// EIP ExternalIPType, display the slb associated elastic ip as service external ip
	if defaulted.ExternalIPType == string(EIPExternalIPType) {
		status.Ingress, err = c.setEIPAsExternalIP(ctx, lb.LoadBalancerId)
	}

	// SLB ExternalIPType, display the slb ip as service external ip
	// If the length of elastic ip is 0, display the slb ip
	if len(status.Ingress) == 0 {
		pz, pzr, err := c.climgr.
			PrivateZones().
			EnsurePrivateZoneRecord(
				ctx, service, lb.Address, defaulted.AddressIPVersion,
			)
		if err != nil {
			return nil, err
		}
		status.Ingress = append(status.Ingress,
			v1.LoadBalancerIngress{
				IP:       lb.Address,
				Hostname: getHostName(pz, pzr),
			})

	}
	return status, err
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.svc and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (c *Cloud) UpdateLoadBalancer(
	ctx context.Context,
	clusterName string,
	service *v1.Service,
	nodes []*v1.Node,
) error {
	klog.V(2).Infof("Alicloud.UpdateLoadBalancer(%v, %v, %v, %v, %v, %v, %v)",
		clusterName, service.Namespace, service.Name, c.region, service.Spec.LoadBalancerIP, service.Spec.Ports, NodeList(nodes))
	ns, err := c.fileOutNode(nodes, service)
	if err != nil {
		return err
	}
	// set up endpoints
	eps, err := c.ifactory.
		Core().V1().
		Endpoints().
		Lister().
		Endpoints(
			service.Namespace,
		).Get(service.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// compatible with nil endpoint
			klog.Warningf("get available endpoints when UpdateLoadBalancer: %s", err.Error())
		} else {
			// avoid removing existing backends of SLB when getting endpoint error
			return fmt.Errorf("get available endpoints when UpdateLoadBalancer: %s", err.Error())
		}
	}
	backends := &EndpointWithENI{
		LocalMode:      ServiceModeLocal(service),
		Endpoints:      eps,
		Nodes:          ns,
		BackendTypeENI: utils.IsENIBackendType(service),
	}
	return c.climgr.LoadBalancers().UpdateLoadBalancer(ctx, service, backends, true)
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
// Implementations must treat the *v1.svc parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (c *Cloud) EnsureLoadBalancerDeleted(
	ctx context.Context,
	clusterName string,
	service *v1.Service,
) error {
	klog.V(2).Infof("Alicloud.EnsureLoadBalancerDeleted(%v, %v, %v, %v, %v, %v)",
		clusterName, service.Namespace, service.Name, c.region, service.Spec.LoadBalancerIP, service.Spec.Ports)

	defaulted, _ := ExtractAnnotationRequest(service)

	if len(service.Status.LoadBalancer.Ingress) > 0 {
		err := c.climgr.PrivateZones().EnsurePrivateZoneRecordDeleted(ctx, service, service.Status.LoadBalancer.Ingress[0].IP, defaulted.AddressIPVersion)
		if err != nil {
			return err
		}
	}

	return c.climgr.LoadBalancers().EnsureLoadBalanceDeleted(ctx, service)
}

// NodeAddresses returns the addresses of the specified instance.
// TODO(roberthbailey): This currently is only used in such a way that it
// returns the address of the calling instance. We should do a rename to
// make this clearer.
func (c *Cloud) NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	klog.V(2).Infof("Alicloud.NodeAddresses(\"%s\")", name)

	return c.climgr.Instances().findAddressByNodeName(ctx, name)
}

func (c *Cloud) ListInstances(ctx context.Context, ids []string) (map[string]*node.CloudNodeAttribute, error) {
	start := time.Now()
	defer func() {
		klog.V(5).Infof("ListInstance take %s to return", time.Since(start))
	}()
	return c.climgr.Instances().ListInstances(ctx, ids)
}

func (c *Cloud) SetInstanceTags(ctx context.Context, insid string, tags map[string]string) error {
	return c.climgr.Instances().AddCloudTags(ctx, insid, tags, c.region)
}

// InstanceTypeByProviderID returns the cloudprovider instance type of the node with the specified unique providerID
// This method will not be called from the node that is requesting this ID. i.e. metadata service
// and other local methods cannot be used here
func (c *Cloud) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	klog.V(5).Infof("Alicloud.InstanceTypeByProviderID(\"%s\")", providerID)
	ins, err := c.climgr.Instances().findInstanceByProviderID(ctx, providerID)
	if err == nil {
		return ins.InstanceType, nil
	}
	return "", err
}

// NodeAddressesByProviderID returns the node addresses of an instances with the specified unique providerID
// This method will not be called from the node that is requesting this ID. i.e. metadata service
// and other local methods cannot be used here
func (c *Cloud) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	klog.V(5).Infof("Alicloud.NodeAddressesByProviderID(\"%s\")", providerID)
	return c.climgr.Instances().findAddressByProviderID(ctx, providerID)
}

// ExternalID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (c *Cloud) ExternalID(ctx context.Context, nodeName types.NodeName) (string, error) {
	klog.V(5).Infof("Alicloud.ExternalID(\"%s\")", nodeName)
	instance, err := c.climgr.Instances().findInstanceByNodeName(ctx, nodeName)
	if err != nil {
		return "", err
	}
	return instance.InstanceId, nil
}

// InstanceID returns the cloud provider ID of the node with the specified NodeName.
func (c *Cloud) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	klog.V(5).Infof("Alicloud.InstanceID(\"%s\")", nodeName)
	instance, err := c.climgr.Instances().findInstanceByNodeName(ctx, nodeName)
	if err != nil {
		return "", err
	}
	return instance.InstanceId, nil
}

// InstanceType returns the type of the specified instance.
func (c *Cloud) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	klog.V(5).Infof("Alicloud.InstanceType(\"%s\")", name)
	instance, err := c.climgr.Instances().findInstanceByNodeName(ctx, name)
	if err != nil {
		return "", err
	}
	return instance.InstanceType, nil
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (c *Cloud) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return errors.New("Alicloud.AddSSHKeyToAllInstances() is not implemented")
}

// CurrentNodeName returns the name of the node we are currently running on
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (c *Cloud) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	nodeName, err := c.climgr.MetaData().InstanceID()
	if err != nil {
		return "", err
	}
	region, err := c.climgr.MetaData().Region()
	if err != nil {
		return "", err
	}
	klog.V(2).Infof("Alicloud.CurrentNodeName(\"%s\")", nodeName)
	return types.NodeName(fmt.Sprintf("%s.%s", region, nodeName)), nil
}

// InstanceExistsByProviderID returns true if the instance for the given provider id still is running.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
func (c *Cloud) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	_, err := c.climgr.Instances().findInstanceByProviderID(ctx, providerID)
	if err == cloudprovider.InstanceNotFound {
		klog.V(2).Infof("Alicloud.InstanceExistsByProviderID(\"%s\") message=[%s]", providerID, err.Error())
		return false, err
	}
	return true, err
}

// InstanceShutdownByProviderID returns true if the instance is shutdown in cloudprovider
func (c *Cloud) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {

	return false, fmt.Errorf("unimplemented")
}

// RouteTables return route table list
func (c *Cloud) RouteTables(ctx context.Context, clusterName string) ([]string, error) {
	return c.climgr.Routes().RouteTables(ctx)
}

// ListRoutes lists all managed routes that belong to the specified clusterName
func (c *Cloud) ListRoutes(ctx context.Context, clusterName string, tableid string) ([]*cloudprovider.Route, error) {
	klog.V(5).Infof("alicloud: ListRoutes \n")

	return c.climgr.Routes().ListRoutes(ctx, tableid)
}

// CreateRoute creates the described managed route
// route.Name will be ignored, although the cloud-provider may use nameHint
// to create a more user-meaningful name.
func (c *Cloud) CreateRoute(ctx context.Context, clusterName string, nameHint string, tableid string, route *cloudprovider.Route) error {
	klog.V(2).Infof("Alicloud.CreateRoute(\"%s, %+v\")", clusterName, route)
	ins, err := c.climgr.Instances().findInstanceByProviderID(ctx, string(route.TargetNode))
	if err != nil {
		return err
	}
	cRoute := &cloudprovider.Route{
		Name:            fmt.Sprintf("%s.%s", ins.RegionId, ins.InstanceId),
		DestinationCIDR: route.DestinationCIDR,
		TargetNode:      types.NodeName(ins.InstanceId),
	}
	return c.climgr.Routes().CreateRoute(ctx, tableid, cRoute, ins.RegionId, ins.VpcAttributes.VpcId)
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
func (c *Cloud) DeleteRoute(ctx context.Context, clusterName string, tableid string, route *cloudprovider.Route) error {
	klog.V(2).Infof("Alicloud.DeleteRoute(\"%s, %+v\")", clusterName, route)

	region, instid, err := nodeFromProviderID(string(route.TargetNode))
	if err != nil {
		return fmt.Errorf("route TargetNode[%s] error: %s", route.TargetNode, err.Error())
	}
	cRoute := &cloudprovider.Route{
		Name:            route.Name,
		DestinationCIDR: route.DestinationCIDR,
		TargetNode:      types.NodeName(instid),
	}
	return c.climgr.Routes().DeleteRoute(ctx, tableid, cRoute, region)
}

// GetZone returns the Zone containing the current failure zone and locality region that the program is running in
func (c *Cloud) GetZone(ctx context.Context) (cloudprovider.Zone, error) {
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
	i, err := c.climgr.Instances().findInstanceByProviderID(ctx, fmt.Sprintf("%s.%s", region, host))
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Alicloud.GetZone(): error execute findInstanceByProviderID(). message=[%s]", err.Error())
	}
	return cloudprovider.Zone{
		Region:        string(i.RegionId),
		FailureDomain: i.ZoneId,
	}, nil
}

// GetZoneByNodeName returns the Zone containing the current zone and locality region of the node specified by node name
// This method is particularly used in the context of external cloud providers where node initialization must be down
// outside the kubelets.
func (c *Cloud) GetZoneByNodeName(ctx context.Context, nodeName types.NodeName) (cloudprovider.Zone, error) {

	i, err := c.climgr.Instances().findInstanceByNodeName(ctx, nodeName)
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Alicloud.GetZoneByNodeName(): error execute findInstanceByNode(). message=[%s]", err.Error())
	}
	return cloudprovider.Zone{
		Region:        string(i.RegionId),
		FailureDomain: i.ZoneId,
	}, nil
}

// GetZoneByProviderID returns the Zone containing the current zone and locality region of the node specified by providerId
// This method is particularly used in the context of external cloud providers where node initialization must be down
// outside the kubelets.
func (c *Cloud) GetZoneByProviderID(ctx context.Context, providerID string) (cloudprovider.Zone, error) {
	i, err := c.climgr.Instances().findInstanceByProviderID(ctx, providerID)
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Alicloud.GetZoneByProviderID(), error execute findInstanceByNode(). message=[%s]", err.Error())
	}
	return cloudprovider.Zone{
		Region:        string(i.RegionId),
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

func (c *Cloud) setEIPAsExternalIP(ctx context.Context, lbId string) ([]v1.LoadBalancerIngress, error) {
	var (
		ingress    []v1.LoadBalancerIngress
		pagination common.Pagination
		eipAddr    []ecs.EipAddressSetType
	)

	for {
		ret, paginationResult, err := c.climgr.instance.DescribeEipAddresses(ctx,
			&ecs.DescribeEipAddressesArgs{
				RegionId:               c.region,
				AssociatedInstanceType: ecs.AssociatedInstanceTypeSlbInstance,
				AssociatedInstanceId:   lbId,
				Pagination:             pagination,
			})
		if err != nil {
			return nil, fmt.Errorf("alicloud: failed to get eip by slb id[%s], DescribeEipAddresses error: %s", lbId,
				err.Error())
		}

		eipAddr = append(eipAddr, ret...)

		next := paginationResult.NextPage()
		if next == nil {
			break
		} else {
			pagination = *next
		}
	}

	if len(eipAddr) == 0 {
		// actually return is warning, and fmt.Errorf() is used to reveal the warning event
		return nil, fmt.Errorf("warning: slb %s has no eip, service.beta.kubernetes.io/alibaba-cloud-loadbalancer-external-ip-type annotation does not work", lbId)
	} else {
		for _, eip := range eipAddr {
			ingress = append(ingress,
				v1.LoadBalancerIngress{
					IP: eip.IpAddress,
				})
		}
		// actually return is warning, and fmt.Errorf() is used to reveal the warning event
		if len(eipAddr) > 1 {
			return ingress, fmt.Errorf("warning: slb %s has multiple eips, eip len %d", lbId, len(eipAddr))
		}
	}
	return ingress, nil
}
