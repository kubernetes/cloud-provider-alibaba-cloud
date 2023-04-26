package elb

type EdgeLoadBalancerAttribute struct {
	LoadBalancerId       string
	LoadBalancerName     string
	LoadBalancerStatus   string
	EnsRegionId          string
	NetworkId            string
	VSwitchId            string
	Address              string
	LoadBalancerSpec     string
	CreateTime           string
	AddressIPVersion     string
	PayType              string
	AssociatedEipId      string
	AssociatedEipName    string
	AssociatedEipAddress string
	IsUserManaged        bool
	IsReUsed             bool
}

// Default Config
const (
	ELBDefaultSpec    = "elb.s2.medium"
	ELBDefaultPayType = "PostPaid"
)
