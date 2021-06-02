package route

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"net"
)

func getIPv4RouteForNode(node *v1.Node) (*net.IPNet, string, error) {
	var (
		ipv4CIDR    *net.IPNet
		ipv4CIDRStr string
		err         error
	)
	for _, podCidr := range append(node.Spec.PodCIDRs, node.Spec.PodCIDR) {
		if podCidr != "" {
			_, ipv4CIDR, err = net.ParseCIDR(podCidr)
			if err != nil {
				return nil, "", fmt.Errorf("invalid pod cidr on node spec: %v", podCidr)
			}
			ipv4CIDRStr = ipv4CIDR.String()
			if len(ipv4CIDR.Mask) == net.IPv4len {
				ipv4CIDRStr = ipv4CIDR.String()
				break
			}
		}
	}
	return ipv4CIDR, ipv4CIDRStr, nil
}
