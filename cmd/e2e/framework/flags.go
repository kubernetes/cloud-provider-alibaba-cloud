package framework

import "flag"

func RegisterCommonFlags() {
	flag.StringVar(&TestContext.KubeConfig, "kubeconfig", "", "kubernetes config path")
	flag.StringVar(&TestContext.CloudConfig, "cloud-config", "", "cloud config")
	flag.StringVar(&TestContext.LoadBalancerID, "lbid", "", "reused slb id")
	flag.StringVar(&TestContext.MasterZoneID, "MasterZoneID", "", "master zone id")
	flag.StringVar(&TestContext.SlaveZoneID, "SlaveZoneID", "", "slave zone id")
	flag.StringVar(&TestContext.BackendLabel, "BackendLabel", "", "backend label")
	flag.StringVar(&TestContext.AclID, "aclid", "", "acl id")
	flag.StringVar(&TestContext.VSwitchID, "vswitchid", "", "vswitch id")
	flag.StringVar(&TestContext.CertID, "certid", "", "cert id")
}

// ViperizeFlags sets up all flag and config processing. Future configuration info should be added to viper, not to flags.
func ViperizeFlags() {

	// Part 1: Set regular flags.
	// TODO: Future, lets eliminate e2e 'flag' deps entirely in favor of viper only,
	// since go test 'flag's are sort of incompatible w/ flag, glog, etc.
	RegisterCommonFlags()
	flag.Parse()
}
