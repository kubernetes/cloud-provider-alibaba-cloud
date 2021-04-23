package pvtz

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

func reverseIPv4(ip string) string {
	ipS := strings.Split(ip, ".")
	l := len(ipS)
	for i := 0; i < l/2; i++ {
		t := ipS[l-1-i]
		ipS[l-1-i] = ipS[i]
		ipS[i] = t
	}
	return strings.Join(ipS, ".")
}

func reverseIPv6(ip string) string {
	tIp := strings.Replace(ip, ":", "", -1)
	ipS := strings.Split(tIp, "")
	l := len(ipS)
	for i := 0; i < l/2; i++ {
		t := ipS[l-1-i]
		ipS[l-1-i] = ipS[i]
		ipS[i] = t
	}
	return strings.Join(ipS, ".")
}

type NamedPortMap map[string]map[string]int32

func NewNamedPortMap(ep *corev1.Endpoints) NamedPortMap {
	namedPortMap := make(map[string]map[string]int32)
	for _, rawSubnet := range ep.Subsets {
		for _, rawPort := range rawSubnet.Ports {
			if _, ok := namedPortMap[string(rawPort.Protocol)]; !ok {
				namedPortMap[string(rawPort.Protocol)] = make(map[string]int32)
			}
			namedPortMap[string(rawPort.Protocol)][rawPort.Name] = rawPort.Port
		}
	}
	return namedPortMap
}

func (m NamedPortMap) GetByProtocolAndPortName(protocol, portName string) (int32, bool) {
	if portMap, ok := m[protocol]; ok {
		p, o := portMap[portName]
		return p, o
	}
	return 0, false
}
