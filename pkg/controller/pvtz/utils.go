package pvtz

import (
	"strings"
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

func removeDuplicateElement(elements []string) []string {
	result := make([]string, 0, len(elements))
	temp := map[string]struct{}{}
	for _, item := range elements {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}