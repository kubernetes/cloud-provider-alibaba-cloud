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

		DisablePublicSLB bool `json:"disablePublicSLB"`

		PrivateZoneID string `json:"privateZoneId"`

		AccessKeyID     string `json:"accessKeyID"`
		AccessKeySecret string `json:"accessKeySecret"`
	}
}

type Flags struct {
	LogLevel           string
	CloudConfig        string
	EnableLeaderSelect bool
}

var GlobalFlag = &Flags{}
