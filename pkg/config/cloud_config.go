package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"k8s.io/klog"
)

var CloudCFG = &CloudConfig{}

// CloudConfig wraps the settings for the Alibaba Cloud provider.
type CloudConfig struct {
	Global struct {
		UID             string `json:"uid"`
		AccessKeyID     string `json:"accessKeyID"`
		AccessKeySecret string `json:"accessKeySecret"`

		// cluster related
		ClusterID            string `json:"clusterID"`
		KubernetesClusterTag string `json:"kubernetesClusterTag"`
		Region               string `json:"region"`
		VpcID                string `json:"vpcid"`
		ZoneID               string `json:"zoneid"`
		VswitchID            string `json:"vswitchid"`

		// service controller
		ServiceBackendType string `json:"serviceBackendType"`
		DisablePublicSLB   bool   `json:"disablePublicSLB"`

		// node controller
		NodeMonitorPeriod  int64 `json:"nodeMonitorPeriod"`
		NodeAddrSyncPeriod int64 `json:"nodeAddrSyncPeriod"`

		// route controller
		RouteTableIDS string `json:"routeTableIDs"`

		// pvtz controller
		PrivateZoneID        string `json:"privateZoneId"`
		PrivateZoneRecordTTL int64  `json:"privateZoneRecordTTL"`

		FeatureGates string `json:"featureGates"`
	}
}

func (cc *CloudConfig) LoadCloudCFG() error {
	content, err := ioutil.ReadFile(ControllerCFG.CloudConfig)
	if err != nil {
		return fmt.Errorf("read cloud config error: %s ", err.Error())
	}
	return yaml.Unmarshal(content, CloudCFG)
}

func (cc *CloudConfig) PrintInfo() {
	if cc.Global.RouteTableIDS != "" {
		klog.Infof("using user customized route table ids [%s]", cc.Global.RouteTableIDS)
	}

	if cc.Global.FeatureGates != "" {
		klog.Infof("using feature gate: %s", cc.Global.FeatureGates)
	}
}
