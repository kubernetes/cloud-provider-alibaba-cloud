package model

const (
	RouteMaxQueryRouteEntry  = 100
	RouteNextHopTypeInstance = "Instance"
	RouteEntryTypeCustom     = "Custom"
)

// Route external route for node
type Route struct {
	Name            string
	DestinationCIDR string
	ProviderId      string
}
