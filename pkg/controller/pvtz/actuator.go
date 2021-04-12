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
		epb.WithType(prvd.RecordTypeA)
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			epb.WithValueData(ingress.IP)
		}
	case corev1.ServiceTypeClusterIP:
		epb.WithType(prvd.RecordTypeA)
		if svc.Spec.ClusterIP == corev1.ClusterIPNone {
			rawEps, err := a.getEndpoints(types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name})
			if err != nil {
				return nil, fmt.Errorf("getting endpoints error: %s", err)
			}
			for _, rawSubnet := range rawEps.Subsets {
				for _, addr := range rawSubnet.Addresses {
					epb.WithValueData(addr.IP)
				}
			}
		} else {
			epb.WithValueData(svc.Spec.ClusterIP)
			for _, ip := range svc.Spec.ClusterIPs {
				epb.WithValueData(ip)
			}
		}
	case corev1.ServiceTypeNodePort:
		epb.WithType(prvd.RecordTypeA)
		epb.WithValueData(svc.Spec.ClusterIP)
		for _, ip := range svc.Spec.ClusterIPs {
			epb.WithValueData(ip)
		}
	case corev1.ServiceTypeExternalName:
		epb.WithValueData(svc.Spec.ExternalName)
		if ip := net.ParseIP(svc.Spec.ExternalName); ip != nil {
			epb.WithType(prvd.RecordTypeA)
		} else {
			epb.WithType(prvd.RecordTypeCNAME)
		}
	}
	return epb.Build(), nil
}
