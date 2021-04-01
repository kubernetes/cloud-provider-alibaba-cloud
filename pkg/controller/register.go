package controller

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/health"
	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"reflect"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/crd"
)

// RegisterFromInClusterCfg register crds from in cluster config file
func RegisterFromInClusterCfg() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("error create incluster config: %s", err.Error())
	}
	return RegisterCRD(cfg)
}

// RegisterFromKubeconfig register crds from kubeconfig file
func RegisterFromKubeconfig(name string) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", name)
	if err != nil {
		return fmt.Errorf("register crd: build rest.config, %s", err.Error())
	}
	return RegisterCRD(cfg)
}

func RegisterCRD(cfg *rest.Config) error {
	extc, err := apiext.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error create incluster client: %s", err.Error())
	}
	client := crd.NewClient(extc)
	for _, crd := range []CRD{
		NewGatewayCRD(client),
		NewAckIngressCRD(client),
	} {
		err := crd.Initialize()
		if err != nil {
			return fmt.Errorf("initialize crd: %s, %s", reflect.TypeOf(crd), err.Error())
		}
	}
	health.CRDReady = true
	return nil
}

type CRD interface {
	Initialize() error
	GetObject() runtime.Object
	GetListerWatcher() cache.ListerWatcher
}

// AckIngressCRD is the cluster crd .
type AckIngressCRD struct {
	crdc crd.Interface
	//kino vcset.Interface
}

func NewAckIngressCRD(
	//kinoClient vcset.Interface,
	crdClient crd.Interface,
) *AckIngressCRD {
	return &AckIngressCRD{
		crdc: crdClient,
		//kino: kinoClient,
	}
}

// podTerminatorCRD satisfies resource.crd interface.
func (p *AckIngressCRD) Initialize() error {
	crd := crd.Conf{
		Kind:                    "AckIngress",
		NamePlural:              "ackingresses",
		Group:                   "alibabacloud.com",
		Version:                 "v1",
		Scope:                   apiextv1beta1.NamespaceScoped,
		EnableStatusSubresource: true,
	}

	return p.crdc.EnsurePresent(crd)
}

// GetListerWatcher satisfies resource.crd interface (and retrieve.Retriever).
func (p *AckIngressCRD) GetListerWatcher() cache.ListerWatcher {
	//return &cache.ListWatch{
	//	ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
	//		return p.kino.KinoV1().Clusters("").List(context.TODO(), options)
	//	},
	//	WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
	//		return p.kino.KinoV1().Clusters("").Watch(context.TODO(),options)
	//	},
	//}
	return nil
}

// GetObject satisfies resource.crd interface (and retrieve.Retriever).
func (p *AckIngressCRD) GetObject() runtime.Object { return &v1.AckIngress{} }

// AckIngressCRD is the cluster crd .
type GatewayCRD struct {
	crdc crd.Interface
	//kino vcset.Interface
}

func NewGatewayCRD(
//kinoClient vcset.Interface,
	crdClient crd.Interface,
) *GatewayCRD {
	return &GatewayCRD{
		crdc: crdClient,
		//kino: kinoClient,
	}
}

// podTerminatorCRD satisfies resource.crd interface.
func (p *GatewayCRD) Initialize() error {
	crd := crd.Conf{
		Kind:                    "Gateway",
		NamePlural:              "gateways",
		Group:                   "alibabacloud.com",
		Version:                 "v1",
		Scope:                   apiextv1beta1.NamespaceScoped,
		EnableStatusSubresource: true,
	}

	return p.crdc.EnsurePresent(crd)
}

// GetListerWatcher satisfies resource.crd interface (and retrieve.Retriever).
func (p *GatewayCRD) GetListerWatcher() cache.ListerWatcher {
	//return &cache.ListWatch{
	//	ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
	//		return p.kino.KinoV1().Clusters("").List(context.TODO(), options)
	//	},
	//	WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
	//		return p.kino.KinoV1().Clusters("").Watch(context.TODO(),options)
	//	},
	//}
	return nil
}

// GetObject satisfies resource.crd interface (and retrieve.Retriever).
func (p *GatewayCRD) GetObject() runtime.Object { return &v1.Gateway{} }
