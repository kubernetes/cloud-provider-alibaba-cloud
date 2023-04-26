package elb

type EdgeEipAttribute struct {
	NamedKey
	Name               string
	Description        string
	EnsRegionId        string
	AllocationId       string
	IpAddress          string
	InstanceId         string
	InstanceType       string
	Status             string
	InternetChargeType string
	InstanceChargeType string
	Bandwidth          int
	IsUserManaged      bool
}

const (
	//EIP
	EipDefaultBandwidth          = 10
	EipDefaultInstanceChargeType = "PostPaid"
	EipDefaultInternetChargeType = "95BandwidthByMonth"
)
