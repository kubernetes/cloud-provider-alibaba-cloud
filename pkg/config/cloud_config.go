package config

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
	}
}
