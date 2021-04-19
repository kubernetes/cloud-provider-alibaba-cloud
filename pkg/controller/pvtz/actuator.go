package pvtz

import (
	"context"
	"fmt"
	"net"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Actuator struct {
	client   client.Client
	provider prvd.Provider
}

func NewActuator(c client.Client, p prvd.Provider) *Actuator {
	a := &Actuator{
		client:   c,
		provider: p,
	}
	return a
}

func (a *Actuator) Update(svc *corev1.Service) error {
	ep, err := a.desiredEndpoints(svc)
	if err != nil {
		return err
	}
	return a.provider.UpdatePVTZ(context.TODO(), ep)
}

func (a *Actuator) UpdatePtr(svc *corev1.Service) error {
	ep, err := a.desiredPtr(svc)
	if err != nil {
		return err
	}
	return a.provider.UpdatePVTZ(context.TODO(), ep)
}

func (a *Actuator) UpdateSrv(svc *corev1.Service) error {
	ep, err := a.desiredSrv(svc)
	if err != nil {
		return err
	}
	return a.provider.UpdatePVTZ(context.TODO(), ep)
}

func (a *Actuator) Delete(svcName types.NamespacedName) error {
	ep := &prvd.PvtzEndpoint{
		Rr:  serviceRrByName(svcName),
		Ttl: ctx2.CFG.Global.PrivateZoneRecordTTL,
	}
	return a.provider.DeletePVTZ(context.TODO(), ep)
}

func (a *Actuator) getEndpoints(epName types.NamespacedName) (*corev1.Endpoints, error) {
	eps := &corev1.Endpoints{}
	err := a.client.Get(context.TODO(), epName, eps)
	if err != nil {
		return nil, err
	}
	return eps, nil
}

// desiredEndpoints should applies to Kubernetes DNS Spec
// https://github.com/kubernetes/dns/blob/master/docs/specification.md
func (a *Actuator) desiredEndpoints(svc *corev1.Service) (*prvd.PvtzEndpoint, error) {
	epb := prvd.NewPvtzEndpointBuilder()
	epb.WithRr(serviceRr(svc))
	epb.WithTtl(ctx2.CFG.Global.PrivateZoneRecordTTL)

	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			epb.WithValueData(ingress.IP)
			if IsIPv4(ingress.IP) {
				epb.WithType(prvd.RecordTypeA)
			} else if IsIPv6(ingress.IP) {
				epb.WithType(prvd.RecordTypeAAAA)
			} else {
				return nil, fmt.Errorf("ingress ip %s is invalid", ingress.IP)
			}
		}
	case corev1.ServiceTypeClusterIP:
		if svc.Spec.ClusterIP == corev1.ClusterIPNone {
			rawEps, err := a.getEndpoints(types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name})
			if err != nil {
				return nil, fmt.Errorf("getting endpoints error: %s", err)
			}
			for _, rawSubnet := range rawEps.Subsets {
				for _, addr := range rawSubnet.Addresses {
					epb.WithValueData(addr.IP)
					if IsIPv4(addr.IP) {
						epb.WithType(prvd.RecordTypeA)
					} else if IsIPv6(addr.IP) {
						epb.WithType(prvd.RecordTypeAAAA)
					} else {
						return nil, fmt.Errorf("pod ip %s is invalid", addr.IP)
					}
				}
			}
		} else {
			ips := append(svc.Spec.ClusterIPs, svc.Spec.ClusterIP)
			for _, ip := range removeDuplicateElement(ips) {
				epb.WithValueData(ip)
				if IsIPv4(ip) {
					epb.WithType(prvd.RecordTypeA)
				} else if IsIPv6(ip) {
					epb.WithType(prvd.RecordTypeAAAA)
				} else {
					return nil, fmt.Errorf("cluster ip %s is invalid", ip)
				}
			}
		}
	case corev1.ServiceTypeNodePort:
		ips := append(svc.Spec.ClusterIPs, svc.Spec.ClusterIP)
		for _, ip := range removeDuplicateElement(ips) {
			epb.WithValueData(ip)
			if IsIPv4(ip) {
				epb.WithType(prvd.RecordTypeA)
			} else if IsIPv6(ip) {
				epb.WithType(prvd.RecordTypeAAAA)
			} else {
				return nil, fmt.Errorf("cluster ip %s is invalid", ip)
			}
		}
	case corev1.ServiceTypeExternalName:
		epb.WithValueData(svc.Spec.ExternalName)
		if ip := net.ParseIP(svc.Spec.ExternalName); ip != nil {
			if IsIPv4(ip.String()) {
				epb.WithType(prvd.RecordTypeA)
			} else if IsIPv6(ip.String()) {
				epb.WithType(prvd.RecordTypeAAAA)
			} else {
				return nil, fmt.Errorf("external ip %s is invalid", ip.String())
			}
		} else {
			epb.WithType(prvd.RecordTypeCNAME)
		}
	}
	return epb.Build(), nil
}

func (a *Actuator) desiredSrv(svc *corev1.Service) (*prvd.PvtzEndpoint, error) {
	epb := prvd.NewPvtzEndpointBuilder()
	epb.WithTtl(ctx2.CFG.Global.PrivateZoneRecordTTL)
	epb.WithType(prvd.RecordTypeSRV)
	svcName := svc.Name
	ns := svc.Namespace
	for _, servicePort := range svc.Spec.Ports {
		srvRr := fmt.Sprintf("_%s._%s.%s.%s.svc", servicePort.Name, servicePort.Protocol, svcName, ns)
		epb.WithRr(srvRr)
		v := fmt.Sprintf("0 100 %s %s.%s.svc", servicePort.TargetPort.String(), svcName, ns)
		epb.WithValueData(v)
	}
	return epb.Build(), nil
}

func (a *Actuator) desiredPtr(svc *corev1.Service) (*prvd.PvtzEndpoint, error) {
	epb := prvd.NewPvtzEndpointBuilder()
	epb.WithTtl(ctx2.CFG.Global.PrivateZoneRecordTTL)
	epb.WithType(prvd.RecordTypePTR)
	epb.WithValueData(serviceRr(svc))

	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if IsIPv4(ingress.IP) {
				epb.WithRr(reverseIPv4(ingress.IP))
			} else if IsIPv6(ingress.IP) {
				epb.WithRr(reverseIPv6(ingress.IP))
			} else {
				return nil, fmt.Errorf("ingress ip %s is invalid", ingress.IP)
			}
		}
	case corev1.ServiceTypeClusterIP:
		if svc.Spec.ClusterIP == corev1.ClusterIPNone {
			rawEps, err := a.getEndpoints(types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name})
			if err != nil {
				return nil, fmt.Errorf("getting endpoints error: %s", err)
			}
			for _, rawSubnet := range rawEps.Subsets {
				for _, addr := range rawSubnet.Addresses {
					if IsIPv4(addr.IP) {
						epb.WithRr(reverseIPv4(addr.IP))
					} else if IsIPv6(addr.IP) {
						epb.WithRr(reverseIPv6(addr.IP))
					} else {
						return nil, fmt.Errorf("pod ip %s is invalid", addr.IP)
					}
				}
			}
		} else {
			ips := append(svc.Spec.ClusterIPs, svc.Spec.ClusterIP)
			for _, ip := range removeDuplicateElement(ips) {
				if IsIPv4(ip) {
					epb.WithRr(reverseIPv4(ip))
				} else if IsIPv6(ip) {
					epb.WithRr(reverseIPv6(ip))
				} else {
					return nil, fmt.Errorf("cluster ip %s is invalid", ip)
				}
			}
		}
	case corev1.ServiceTypeNodePort:
		ips := append(svc.Spec.ClusterIPs, svc.Spec.ClusterIP)
		for _, ip := range removeDuplicateElement(ips) {
			if IsIPv4(ip) {
				epb.WithRr(reverseIPv4(ip))
			} else if IsIPv6(ip) {
				epb.WithRr(reverseIPv6(ip))
			} else {
				return nil, fmt.Errorf("cluster ip %s is invalid", ip)
			}
		}
	case corev1.ServiceTypeExternalName:
		epb.WithValueData(svc.Spec.ExternalName)
		if ip := net.ParseIP(svc.Spec.ExternalName); ip != nil {
			if IsIPv4(ip.String()) {
				epb.WithRr(reverseIPv4(ip.String()))
			} else if IsIPv6(ip.String()) {
				epb.WithRr(reverseIPv6(ip.String()))
			} else {
				return nil, fmt.Errorf("external ip %s is invalid", ip.String())
			}
		}
	}
	return epb.Build(), nil
}
