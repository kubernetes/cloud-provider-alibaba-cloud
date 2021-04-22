package pvtz

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	util_errors "k8s.io/apimachinery/pkg/util/errors"
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

func (a *Actuator) UpdateService(svc *corev1.Service) error {
	eps := make([]*prvd.PvtzEndpoint, 0)
	desiredFuncs := []func(svc *corev1.Service) ([]*prvd.PvtzEndpoint, error){
		a.desiredAandAAAA,
		a.desiredSRV,
		a.desiredCNAME,
		a.desiredPTR,
	}
	errs := make([]error, 0)
	for _, f := range desiredFuncs {
		ps, err := f(svc)
		if err != nil {
			errs = append(errs, err)
		}
		eps = append(eps, ps...)
	}
	for _, ep := range eps {
		err := a.provider.UpdatePVTZ(context.TODO(), ep)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Wrap(util_errors.NewAggregate(errs), "UpdateService error")
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
func (a *Actuator) desiredAandAAAA(svc *corev1.Service) ([]*prvd.PvtzEndpoint, error) {
	var ipsV4 []string
	var ipsV6 []string

	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if IsIPv4(ingress.IP) {
				ipsV4 = append(ipsV4, ingress.IP)
			} else if IsIPv6(ingress.IP) {
				ipsV6 = append(ipsV6, ingress.IP)
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
						ipsV4 = append(ipsV4, addr.IP)
					} else if IsIPv6(addr.IP) {
						ipsV6 = append(ipsV6, addr.IP)
					} else {
						return nil, fmt.Errorf("pod ip %s is invalid", addr.IP)
					}
				}
			}
		} else {
			ips := append(svc.Spec.ClusterIPs, svc.Spec.ClusterIP)
			for _, ip := range ips {
				if IsIPv4(ip) {
					ipsV4 = append(ipsV4, ip)
				} else if IsIPv6(ip) {
					ipsV6 = append(ipsV6, ip)
				} else {
					return nil, fmt.Errorf("cluster ip %s is invalid", ip)
				}
			}
		}
	case corev1.ServiceTypeNodePort:
		ips := append(svc.Spec.ClusterIPs, svc.Spec.ClusterIP)
		for _, ip := range ips {
			if IsIPv4(ip) {
				ipsV4 = append(ipsV4, ip)
			} else if IsIPv6(ip) {
				ipsV6 = append(ipsV6, ip)
			} else {
				return nil, fmt.Errorf("cluster ip %s is invalid", ip)
			}
		}
	case corev1.ServiceTypeExternalName:
		if ip := net.ParseIP(svc.Spec.ExternalName); ip != nil {
			if IsIPv4(ip.String()) {
				ipsV4 = append(ipsV4, svc.Spec.ExternalName)
			} else {
				ipsV6 = append(ipsV6, svc.Spec.ExternalName)
			}
		}
	}
	var eps []*prvd.PvtzEndpoint

	epTemplate := prvd.NewPvtzEndpointBuilder()
	epTemplate.WithRr(serviceRr(svc))
	epTemplate.WithTtl(ctx2.CFG.Global.PrivateZoneRecordTTL)

	if len(ipsV4) != 0 {
		epb := epTemplate.DeepCopy()
		epb.WithType(prvd.RecordTypeA)
		for _, ip := range ipsV4 {
			epb.WithValueData(ip)
		}
		eps = append(eps, epb.Build())
	}
	if len(ipsV6) != 0 {
		epb := epTemplate.DeepCopy()
		epb.WithType(prvd.RecordTypeAAAA)
		for _, ip := range ipsV6 {
			epb.WithValueData(ip)
		}
		eps = append(eps, epb.Build())
	}
	return eps, nil
}

func (a *Actuator) desiredSRV(svc *corev1.Service) ([]*prvd.PvtzEndpoint, error) {
	eps := make([]*prvd.PvtzEndpoint, 0)
	epTemplate := prvd.NewPvtzEndpointBuilder()
	epTemplate.WithTtl(ctx2.CFG.Global.PrivateZoneRecordTTL)
	epTemplate.WithType(prvd.RecordTypeSRV)
	svcName := svc.Name
	ns := svc.Namespace
	for _, servicePort := range svc.Spec.Ports {
		epb := epTemplate.DeepCopy()
		epb.WithRr(fmt.Sprintf("_%s._%s.%s.%s.svc", servicePort.Name, servicePort.Protocol, svcName, ns))
		epb.WithValueData(fmt.Sprintf("0 100 %s %s.%s.svc", servicePort.TargetPort.String(), svcName, ns))
		eps = append(eps, epb.Build())
	}
	return eps, nil
}

func (a *Actuator) desiredPTR(svc *corev1.Service) ([]*prvd.PvtzEndpoint, error) {
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
			for _, ip := range ips {
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
		for _, ip := range ips {
			if IsIPv4(ip) {
				epb.WithRr(reverseIPv4(ip))
			} else if IsIPv6(ip) {
				epb.WithRr(reverseIPv6(ip))
			} else {
				return nil, fmt.Errorf("cluster ip %s is invalid", ip)
			}
		}
	}
	return []*prvd.PvtzEndpoint{epb.Build()}, nil
}

func (a *Actuator) desiredCNAME(svc *corev1.Service) ([]*prvd.PvtzEndpoint, error) {
	epb := prvd.NewPvtzEndpointBuilder()
	epb.WithRr(serviceRr(svc))
	epb.WithTtl(ctx2.CFG.Global.PrivateZoneRecordTTL)
	epb.WithType(prvd.RecordTypeCNAME)
	if svc.Spec.Type == corev1.ServiceTypeExternalName {
		if ip := net.ParseIP(svc.Spec.ExternalName); ip == nil {
			epb.WithValueData(svc.Spec.ExternalName)
		}
	}
	return []*prvd.PvtzEndpoint{epb.Build()}, nil
}
