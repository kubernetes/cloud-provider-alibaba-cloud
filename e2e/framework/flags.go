package framework

import (
	"flag"
	"testing"
)

func RegisterCommonFlags() {
	flag.StringVar(&TestConfig.KubeConfig, "kubeconfig", "", "kubernetes config path")
	flag.StringVar(&TestConfig.CloudConfig, "cloud-config", "", "cloud config path")
	flag.StringVar(&TestConfig.LoadBalancerID, "lb-id", "", "reused slb id")
	flag.StringVar(&TestConfig.MasterZoneID, "master-zone-id", "", "master zone id")
	flag.StringVar(&TestConfig.SlaveZoneID, "slave-zone-id", "", "slave zone id")
	flag.StringVar(&TestConfig.BackendLabel, "backend-label", "", "backend label: k1=v1,k2=v2")
	flag.StringVar(&TestConfig.AclID, "acl-id", "", "acl id")
	flag.StringVar(&TestConfig.VSwitchID, "vswitch-id", "", "vswitch id")
	flag.StringVar(&TestConfig.CertID, "cert-id", "", "cert id")
	flag.StringVar(&TestConfig.PrivateZoneID, "private-zone-id", "", "private zone id")
	flag.StringVar(&TestConfig.PrivateZoneName, "private-zone-mame", "", "private zone name")
	flag.StringVar(&TestConfig.PrivateZoneRecordName, "private-zone-record-name", "", "private zone record id")
	flag.StringVar(&TestConfig.PrivateZoneRecordTTL, "private-zone-record-ttl", "", "private zone record ttl")
	flag.StringVar(&TestConfig.ResourceGroupID, "resource-group-id", "", "resource group id")
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
