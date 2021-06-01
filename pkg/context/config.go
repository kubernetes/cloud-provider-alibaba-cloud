package context

var CFG = &CloudConfig{}

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
		ServiceBackendType   string `json:"serviceBackendType"`

		ConfigureCloudRoutes      bool   `json:"configureCloudRoutes"`
		RouteReconciliationPeriod string `json:"routeReconciliationPeriod"`
		ClusterCidr               string `json:"clusterCidr"`

		DisablePublicSLB bool `json:"disablePublicSLB"`

		PrivateZoneID        string `json:"privateZoneId"`
		PrivateZoneRecordTTL int64  `json:"privateZoneRecordTTL"`

		AccessKeyID     string `json:"accessKeyID"`
		AccessKeySecret string `json:"accessKeySecret"`
	}
}

type Flags struct {
	LogLevel           string
	CloudConfig        string
	EnableLeaderSelect bool
	EnableControllers  []string
}

var GlobalFlag = &Flags{}
