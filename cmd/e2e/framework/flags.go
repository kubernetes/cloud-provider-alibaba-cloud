package framework

import (
	"flag"
	"testing"
)

func RegisterCommonFlags() {
	flag.StringVar(&TestContext.KubeConfig, "kubeconfig", "", "kubernetes config path")
	flag.StringVar(&TestContext.CloudConfig, "cloud-config", "", "cloud config")
	flag.StringVar(&TestContext.LoadBalancerID, "lbid", "", "reused slb id")
	flag.StringVar(&TestContext.MasterZoneID, "MasterZoneID", "", "master zone id")
	flag.StringVar(&TestContext.SlaveZoneID, "SlaveZoneID", "", "slave zone id")
	flag.StringVar(&TestContext.BackendLabel, "BackendLabel", "", "backend label: k1=v1,k2=v2")
	flag.StringVar(&TestContext.AclID, "aclid", "", "acl id")
	flag.StringVar(&TestContext.VSwitchID, "vswitchid", "", "vswitch id")
	flag.StringVar(&TestContext.CertID, "certid", "", "cert id")
	flag.StringVar(&TestContext.PrivateZoneID, "PrivateZoneID", "", "private zone id")
	flag.StringVar(&TestContext.PrivateZoneName, "PrivateZoneName", "", "private zone name")
	flag.StringVar(&TestContext.PrivateZoneRecordName, "PrivateZoneRecordName", "", "private zone record id")
	flag.StringVar(&TestContext.PrivateZoneRecordTTL, "PrivateZoneRecordTTL", "", "private zone record ttl")
	flag.StringVar(&TestContext.ResourceGroupID, "ResourceGroupID", "", "resource group id")
}

// ViperizeFlags sets up all flag and config processing. Future configuration info should be added to viper, not to flags.
func ViperizeFlags() {

	testing.Init()
	// Part 1: Set regular flags.
	// TODO: Future, lets eliminate e2e 'flag' deps entirely in favor of viper only,
	// since go test 'flag's are sort of incompatible w/ flag, klog, etc.
	RegisterCommonFlags()
	flag.Parse()
}
