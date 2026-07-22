package nlbv2

import v1 "k8s.io/api/core/v1"

func loadBalancerStatusEqual(l, r *v1.LoadBalancerStatus) bool {
	if len(l.Ingress) != len(r.Ingress) {
		return false
	}
	for i := range l.Ingress {
		if l.Ingress[i].IP != r.Ingress[i].IP || l.Ingress[i].Hostname != r.Ingress[i].Hostname {
			return false
		}
	}
	return true
}
