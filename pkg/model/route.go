package model

import v1 "k8s.io/api/core/v1"

const (
	RouteMaxQueryRouteEntry  = 500
	RouteNextHopTypeInstance = "Instance"
	RouteEntryTypeCustom     = "Custom"
)

// Route external route for node
type Route struct {
	Name            string
	DestinationCIDR string
	ProviderId      string
	NodeReference   *v1.Node
}
