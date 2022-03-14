package options

import (
	"flag"
	"fmt"
)

const (
	Terway  = "terway"
	Flannel = "flannel"
)

var TestConfig = &E2EConfig{}

type E2EConfig struct {
	CloudConfig              string `json:"cloudConfig"`
	Controllers              string `json:"controller"`
	RegionId                 string `json:"regionId"`
	ClusterType              string `json:"clusterType"`
	ClusterId                string `json:"clusterId"`
	AllowCreateCloudResource bool   `json:"allowCreateCloudResource"` // whether to create cloud resources for test

	// need provided
	VPCLoadBalancerID string `json:"VPCLoadBalancerID"` // lb in other vpc
	EipLoadBalancerID string `json:"EipLoadBalancerID"` // intranet slb with eip
	ResourceGroupID   string `json:"ResourceGroupID"`

	// described by cluster info
	Network      string `json:"network"`
	EnableVK     bool   `json:"enableVK"`
	VPCID        string `json:"VPCID"`
	VSwitchID    string `json:"VSwitchID"`
	VSwitchID2   string `json:"VSwitchID2"`
	MasterZoneID string `json:"MasterZoneID"`
	SlaveZoneID  string `json:"SlaveZoneID"`
	CertID       string `json:"CertID"`
	CertID2      string `json:"CertID2"`

	// created automatically if AllowCreateCloudResource=true
	InternetLoadBalancerID string `json:"InternetLoadBalancerID"`
	IntranetLoadBalancerID string `json:"IntranetLoadBalancerID"`
	VServerGroupID         string `json:"VServerGroupID"`  // vServerGroupID of InternetLoadBalancerID
	VServerGroupID2        string `json:"VServerGroupID2"` // vServerGroupID of InternetLoadBalancerID
	AclID                  string `json:"AclID"`
	AclID2                 string `json:"AclID2"`
}

func (e *E2EConfig) BindFlags() {
	flag.StringVar(&e.CloudConfig, "cloud-config", "",
		"The path to the cloud provider configuration file. Empty string for no configuration file.")
	flag.StringVar(&e.Controllers, "controllers", "service,node,route", "controllers to run tests.")
	flag.StringVar(&e.RegionId, "region-id", "", "the region id of cluster")
	flag.StringVar(&e.ClusterId, "cluster-id", "", "the id of cluster which is used to run e2e test")
	flag.BoolVar(&e.AllowCreateCloudResource, "allow-create-cloud-resources", false, "whether allow to create cloud resources, including the Kubernetes Cluster, SLB, ECS, etc.")
	flag.StringVar(&TestConfig.EipLoadBalancerID, "eip-lb-id", "", "reused intranet slb id which has eip")
	flag.StringVar(&TestConfig.VPCLoadBalancerID, "vpc-lb-id", "", "reused intranet slb id which in other vpc")
	flag.StringVar(&TestConfig.MasterZoneID, "master-zone-id", "", "master zone id")
	flag.StringVar(&TestConfig.SlaveZoneID, "slave-zone-id", "", "slave zone id")
	flag.StringVar(&TestConfig.ResourceGroupID, "resource-group-id", "", "resource group id, do not use the resource group id of the ack cluster")
	flag.StringVar(&TestConfig.Network, "network", "", "the network type of kubernetes, values: flannel or terway")
	flag.StringVar(&TestConfig.IntranetLoadBalancerID, "intranet-lb-id", "", "reused intranet slb id")
	flag.StringVar(&TestConfig.InternetLoadBalancerID, "internet-lb-id", "", "reused internet slb id")
	flag.StringVar(&TestConfig.AclID, "acl-id", "", "acl id")
	flag.StringVar(&TestConfig.AclID2, "acl-id-2", "", "acl id")
	flag.StringVar(&TestConfig.VSwitchID, "vswitch-id", "", "vsw-id")
	flag.StringVar(&TestConfig.VSwitchID2, "vswitch-id-2", "", "vsw-id")
	flag.StringVar(&TestConfig.CertID, "cert-id", "", "cert id")
	flag.StringVar(&TestConfig.CertID2, "cert-id-2", "", "cert id")
	flag.StringVar(&TestConfig.VServerGroupID, "vserver-group-id", "", "vserver group id")
	flag.StringVar(&TestConfig.VServerGroupID2, "vserver-group-id-2", "", "vserver group id")
}

func (e *E2EConfig) Validate() error {
	if e.CloudConfig == "" {
		return fmt.Errorf("cloud config can not be empty")
	}
	if e.RegionId == "" {
		return fmt.Errorf("region id can not be empty")
	}
	if e.ClusterId == "" {
		return fmt.Errorf("cluster id can not be empty")
	}
	return nil
}
