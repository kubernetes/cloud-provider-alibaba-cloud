package framework

import (
	"flag"
	"testing"
)

func RegisterCommonFlags() {
	flag.StringVar(&TestContext.KubeConfig, "kubeconfig", "", "kubernetes config path")
	flag.StringVar(&TestContext.CloudConfig, "framework-config", "", "framework config")
	flag.StringVar(&TestContext.LoadBalancerID, "lb-id", "", "reused slb id")
	flag.StringVar(&TestContext.MasterZoneID, "master-zone-id", "", "master zone id")
	flag.StringVar(&TestContext.SlaveZoneID, "slave-zone-id", "", "slave zone id")
	flag.StringVar(&TestContext.BackendLabel, "backend-label", "", "backend label: k1=v1,k2=v2")
	flag.StringVar(&TestContext.AclID, "acl-id", "", "acl id")
	flag.StringVar(&TestContext.VSwitchID, "vswitch-id", "", "vswitch id")
	flag.StringVar(&TestContext.CertID, "cert-id", "", "cert id")
	flag.StringVar(&TestContext.PrivateZoneID, "private-zone-id", "", "private zone id")
	flag.StringVar(&TestContext.PrivateZoneName, "private-zone-mame", "", "private zone name")
	flag.StringVar(&TestContext.PrivateZoneRecordName, "private-zone-record-name", "", "private zone record id")
	flag.StringVar(&TestContext.PrivateZoneRecordTTL, "private-zone-record-ttl", "", "private zone record ttl")
	flag.StringVar(&TestContext.ResourceGroupID, "resource-group-id", "", "resource group id")
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
