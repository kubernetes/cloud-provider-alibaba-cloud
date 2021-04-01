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
	provider provider.Provider
}

func NewActuator(c client.Client, p provider.Provider) *Actuator {
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
	ep := &provider.PvtzEndpoint{
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

func (a *Actuator) desiredEndpoints(svc *corev1.Service) (*provider.PvtzEndpoint, error) {
	ep := &provider.PvtzEndpoint{
		Rr:  serviceRr(svc),
		Ttl: ctx2.CFG.Global.PrivateZoneRecordTTL,
	}
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		// FIXME Short circuit to ClusterIP?
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			lbIP := svc.Status.LoadBalancer.Ingress[0].IP
			if lbIP == "" {
				return nil, fmt.Errorf("no lb IP found")
			}
			ep.Values = []provider.PvtzValue{{Data: lbIP}}
			ep.Type = provider.RecordTypeA
		}
	case corev1.ServiceTypeClusterIP:
		// Headless
		if svc.Spec.ClusterIP == corev1.ClusterIPNone {
			// TODO
			rawEps, err := a.getEndpoints(types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name})
			if err != nil {
				return nil, fmt.Errorf("getting endpoints error: %s", err)
			}
			ep.Values = make([]provider.PvtzValue, 0)
			for _, rawSubnet := range rawEps.Subsets {
				for _, addr := range rawSubnet.Addresses {
					ep.Values = append(ep.Values, provider.PvtzValue{Data: addr.IP})
				}
			}
			ep.Type = provider.RecordTypeA
		} else {
			ep.Values = []provider.PvtzValue{{Data: svc.Spec.ClusterIP}}
			ep.Type = provider.RecordTypeA
		}
	case corev1.ServiceTypeNodePort:
		ep.Values = []provider.PvtzValue{{Data: svc.Spec.ClusterIP}}
		ep.Type = provider.RecordTypeA
	case corev1.ServiceTypeExternalName:
		ep.Values = []provider.PvtzValue{{Data: svc.Spec.ExternalName}}
		if ip := net.ParseIP(svc.Spec.ExternalName); ip != nil {
			ep.Type = provider.RecordTypeA
		} else {
			ep.Type = provider.RecordTypeCNAME
		}
	}
	return ep, nil
}
