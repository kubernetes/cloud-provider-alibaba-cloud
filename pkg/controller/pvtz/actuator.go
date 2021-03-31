package pvtz

import (
	"context"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	defaultTTL = ctx2.CFG.Global.PrivateZoneRecordTTL
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

func (a *Actuator) ReconcileService(svc *corev1.Service) error {
	rlog := log.WithFields(log.Fields{"service": fullServiceName(svc)})
	rlog.Println("reconciling service")
	ep, err := a.desiredEndpoints(svc)
	if err != nil {
		rlog.Errorf("getting desiredEndpoint error: %s", err)
	}
	err = a.provider.UpdatePVTZ(context.TODO(), ep)
	if err != nil {
		rlog.Errorf("adding service error: %s", err)
	}
	return nil
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
		Ttl: defaultTTL,
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
