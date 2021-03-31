package pvtz

import (
	"context"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Actuator struct {
	client client.Client
	provider provider.Provider
}

func NewActuator(c client.Client, p provider.Provider) *Actuator {
	a := &Actuator{
		client: c,
		provider: p,
	}
	return a
}

func (a *Actuator) getEndpoints(epName types.NamespacedName) (*corev1.Endpoints, error) {
	eps := &corev1.Endpoints{}
	err := a.client.Get(context.TODO(), epName, eps)
	if err != nil {
		return nil, err
	}
	return eps, nil
}

func (a *Actuator) DesiredEndpoints(svc *corev1.Service) (*provider.PvtzEndpoint, error) {
	ep := &provider.PvtzEndpoint{
		Rr: fmt.Sprintf("%s.%s.svc", svc.Name, svc.Namespace),
		// TODO customized ttl
		Ttl: int64(60),
	}
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		// FIXME Short circuit to ClusterIP?
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			lbIP := svc.Status.LoadBalancer.Ingress[0].IP
			if lbIP == "" {
				return nil, fmt.Errorf("not lbIP found for loadbalancer service %s/%s", svc.Namespace, svc.Name)
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
				return nil, fmt.Errorf("getting endpoints for %s/%s error: %s", svc.Namespace, svc.Name, err)
			}
			ep.Values = make([]provider.PvtzValue, 0)
			for _, rawSubnet := range rawEps.Subsets {
				for _, addr := range rawSubnet.Addresses {
					ep.Values = append(ep.Values, provider.PvtzValue{Data: addr.IP})
				}
			}
			log.Infof("headless service have %d endpoints %s", len(ep.Values), rawEps.Subsets)
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