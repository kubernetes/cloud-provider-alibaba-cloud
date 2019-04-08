package route

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
)

// RoutesOptions route controller options
type RoutesOptions struct {
	AllocateNodeCIDRs         bool
	ClusterCIDR               string
	ConfigCloudRoutes         bool
	MinResyncPeriod           metav1.Duration
	RouteReconciliationPeriod metav1.Duration
	ControllerStartInterval   metav1.Duration
}

// Options global options for route controller
var Options = RoutesOptions{}

// RealContainsCidr real contains cidr
func RealContainsCidr(outer string, inner string) (bool, error) {
	contains, err := ContainsCidr(outer, inner)
	if err != nil {
		return false, err
	}
	if outer == inner {
		return false, nil
	}
	return contains, nil
}

// ContainsCidr contains with equal.
func ContainsCidr(outer string, inner string) (bool, error) {
	_, outerCidr, err := net.ParseCIDR(outer)
	if err != nil {
		return false, fmt.Errorf("error parse outer cidr: %s, message: %s", outer, err.Error())
	}
	_, innerCidr, err := net.ParseCIDR(inner)
	if err != nil {
		return false, fmt.Errorf("error parse inner cidr: %s, message: %s", inner, err.Error())
	}

	lastIP := make([]byte, len(innerCidr.IP))
	for i := range lastIP {
		lastIP[i] = innerCidr.IP[i] | ^innerCidr.Mask[i]
	}
	if !outerCidr.Contains(innerCidr.IP) || !outerCidr.Contains(lastIP) {
		return false, nil
	}
	return true, nil
}
