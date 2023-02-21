package ingress

import (
	"fmt"
	"reflect"

	"k8s.io/cloud-provider-alibaba-cloud/cmd/health"
	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
		NewAlbConfigCRD(client),
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

// AlbConfigCRD is the cluster crd .
type AlbConfigCRD struct {
	crdc crd.Interface
	//kino vcset.Interface
}

func NewAlbConfigCRD(
	//kinoClient vcset.Interface,
	crdClient crd.Interface,
) *AlbConfigCRD {
	return &AlbConfigCRD{
		crdc: crdClient,
		//kino: kinoClient,
	}
}

// podTerminatorCRD satisfies resource.crd interface.
func (p *AlbConfigCRD) Initialize() error {
	crd := crd.Conf{
		Kind:                    "AlbConfig",
		NamePlural:              "albconfigs",
		Group:                   "alibabacloud.com",
		Version:                 "v1",
		Scope:                   apiextv1.ClusterScoped,
		EnableStatusSubresource: true,
	}

	return p.crdc.EnsurePresent(crd)
}

// GetListerWatcher satisfies resource.crd interface (and retrieve.Retriever).
func (p *AlbConfigCRD) GetListerWatcher() cache.ListerWatcher {
	return nil
}

// GetObject satisfies resource.crd interface (and retrieve.Retriever).
func (p *AlbConfigCRD) GetObject() runtime.Object { return &v1.AlbConfig{} }
