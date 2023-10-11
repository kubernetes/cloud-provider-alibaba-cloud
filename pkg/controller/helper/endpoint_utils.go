package helper

import (
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"strings"
)

func LogEndpoints(eps *v1.Endpoints) string {
	if eps == nil {
		return "endpoints is nil"
	}
	var epAddrList []string
	for _, subSet := range eps.Subsets {
		for _, addr := range subSet.Addresses {
			epAddrList = append(epAddrList, addr.IP)
		}
	}
	return strings.Join(epAddrList, ",")
}

func LogEndpointSlice(es *discovery.EndpointSlice) string {
	if es == nil {
		return "endpointSlice is nil"
	}
	var epAddrList []string
	for _, ep := range es.Endpoints {
		epAddrList = append(epAddrList, ep.Addresses...)
	}

	return strings.Join(epAddrList, ",")
}

func LogEndpointSliceList(esList []discovery.EndpointSlice) string {
	if esList == nil {
		return "endpointSliceList is nil"
	}
	var epAddrList []string
	for _, es := range esList {
		for _, ep := range es.Endpoints {
			if ep.Conditions.Ready != nil && !*ep.Conditions.Ready {
				continue
			}
			epAddrList = append(epAddrList, ep.Addresses...)
		}
	}

	return strings.Join(epAddrList, ",")
}
