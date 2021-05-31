package route

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"net"
)

func getIPv4RouteForNode(node *v1.Node) (string, error) {
	var ipv4CIDR string
	for _, podCidr := range append(node.Spec.PodCIDRs, node.Spec.PodCIDR) {
		if podCidr != "" {
			_, cidr, err := net.ParseCIDR(podCidr)
			if err != nil {
				return "", fmt.Errorf("invalid pod cidr on node spec: %v", podCidr)
			}
			ipv4CIDR = cidr.String()
			if len(cidr.Mask) == net.IPv4len {
				ipv4CIDR = cidr.String()
				break
			}
		}
	}
	return ipv4CIDR, nil
}
