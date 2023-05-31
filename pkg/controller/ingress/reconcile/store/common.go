package store

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	// IngressKey picks a specific "class" for the Ingress.
	// The controller only processes Ingresses with this annotation either
	// unset, or set to either the configured value or the empty string.
	IngressKey = "kubernetes.io/ingress.class"
)

var (
	// DefaultClass defines the default class used in the alb ingress controller
	DefaultClass = "alb"

	// IngressClass sets the runtime ingress class to use
	// An empty string means accept all ingresses without
	// annotation and the ones configured with class alb
	IngressClassName = "alb"
)

// Ingress holds the definition of an Ingress plus its annotations
type Ingress struct {
	networking.Ingress `json:"-"`
}

// IsValid returns true if the given Ingress specify the ingress.class
// annotation or IngressClassName resource for Kubernetes >= v1.18
func IsValid(ing *networking.Ingress) bool {
	// 1. with annotation or IngressClass
	ingress, ok := ing.GetAnnotations()[IngressKey]
	if !ok && ing.Spec.IngressClassName != nil {
		ingress = *ing.Spec.IngressClassName
	}

	// k8s > v1.18.
	// Processing may be redundant because k8s.IngressClass is obtained by IngressClass
	// 2. without annotation and IngressClass. Check IngressClass
	if IngressClass != nil {
		return ingress == IngressClass.Name
	}

	// 3. with IngressClass
	return ingress == IngressClassName
}

// ConfigMapLister makes a Store that lists Configmaps.
type ConfigMapLister struct {
	cache.Store
}

// ByKey returns the ConfigMap matching key in the local ConfigMap Store.
func (cml *ConfigMapLister) ByKey(key string) (*apiv1.ConfigMap, error) {
	s, exists, err := cml.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return s.(*apiv1.ConfigMap), nil
}

// EndpointLister makes a Store that lists Endpoints.
type EndpointLister struct {
	cache.Store
}

// ByKey returns the Endpoints of the Service matching key in the local Endpoint Store.
func (s *EndpointLister) ByKey(key string) (*apiv1.Endpoints, error) {
	eps, exists, err := s.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return eps.(*apiv1.Endpoints), nil
}

// IngressLister makes a Store that lists Ingress.
type IngressLister struct {
	cache.Store
}

// ByKey returns the Ingress matching key in the local Ingress Store.
func (il IngressLister) ByKey(key string) (*networking.Ingress, error) {
	i, exists, err := il.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return i.(*networking.Ingress), nil
}

// FilterIngresses returns the list of Ingresses
func FilterIngresses(ingresses []*Ingress, filterFunc IngressFilterFunc) []*Ingress {
	afterFilter := make([]*Ingress, 0)
	for _, ingress := range ingresses {
		if !filterFunc(ingress) {
			afterFilter = append(afterFilter, ingress)
		}
	}

	sortIngressSlice(afterFilter)
	return afterFilter
}

// NodeLister makes a Store that lists Nodes.
type NodeLister struct {
	cache.Store
}

// ByKey returns the Node matching key in the local Node Store.
func (sl *NodeLister) ByKey(key string) (*apiv1.Node, error) {
	s, exists, err := sl.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return s.(*apiv1.Node), nil
}

// IngressWithAnnotationsLister makes a Store that lists Ingress rules with annotations already parsed
type IngressWithAnnotationsLister struct {
	cache.Store
}

// ByKey returns the Ingress with annotations matching key in the local store or an error
func (il IngressWithAnnotationsLister) ByKey(key string) (*Ingress, error) {
	i, exists, err := il.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return i.(*Ingress), nil
}

// ParseNameNS parses a string searching a namespace and name
func ParseNameNS(input string) (string, string, error) {
	nsName := strings.Split(input, "/")
	if len(nsName) != 2 {
		return "", "", fmt.Errorf("invalid format (namespace/name) found in '%v'", input)
	}

	return nsName[0], nsName[1], nil
}

// GetNodeIPOrName returns the IP address or the name of a node in the cluster
func GetNodeIPOrName(kubeClient clientset.Interface, name string, useInternalIP bool) string {
	node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Error getting node", "name", name)
		return ""
	}

	defaultOrInternalIP := ""
	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeInternalIP {
			if address.Address != "" {
				defaultOrInternalIP = address.Address
				break
			}
		}
	}

	if useInternalIP {
		return defaultOrInternalIP
	}

	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeExternalIP {
			if address.Address != "" {
				return address.Address
			}
		}
	}

	return defaultOrInternalIP
}

var (
	// IngressPodDetails hold information about the ingress-nginx pod
	IngressPodDetails *PodInfo
)

// PodInfo contains runtime information about the pod running the Ingres controller
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PodInfo struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

// GetIngressPod load the ingress-nginx pod
func GetIngressPod(kubeClient clientset.Interface) error {
	podName := os.Getenv("POD_NAME")
	podNs := os.Getenv("POD_NAMESPACE")

	if podName == "" || podNs == "" {
		return fmt.Errorf("unable to get POD information (missing POD_NAME or POD_NAMESPACE environment variable")
	}

	pod, err := kubeClient.CoreV1().Pods(podNs).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get POD information: %v", err)
	}

	IngressPodDetails = &PodInfo{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
	}

	pod.ObjectMeta.DeepCopyInto(&IngressPodDetails.ObjectMeta)
	IngressPodDetails.SetLabels(pod.GetLabels())

	return nil
}

// MetaNamespaceKey knows how to make keys for API objects which implement meta.Interface.
func MetaNamespaceKey(obj interface{}) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Warning(err)
	}

	return key
}

// IsIngressV1Beta1Ready indicates if the running Kubernetes version is at least v1.18.0
var IsIngressV1Beta1Ready bool

// IsIngressV1Ready indicates if the running Kubernetes version is at least v1.19.0
var IsIngressV1Ready bool

// IngressClass indicates the class of the Ingress to use as filter
var IngressClass *networking.IngressClass

// IngressNGINXController defines the valid value of IngressClass
// Controller field for ingress-nginx
const IngressNGINXController = "k8s.io/ingress-nginx"

// NetworkingIngressAvailable checks if the package "k8s.io/api/networking/v1beta1"
// is available or not and if Ingress V1 is supported (k8s >= v1.18.0)
func NetworkingIngressAvailable(client clientset.Interface) (bool, bool, bool) {
	// check kubernetes version to use new ingress package or not
	version114, _ := version.ParseGeneric("v1.14.0")
	version118, _ := version.ParseGeneric("v1.18.0")
	version119, _ := version.ParseGeneric("v1.19.0")

	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return false, false, false
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		klog.ErrorS(err, "unexpected error parsing running Kubernetes version")
		return false, false, false
	}

	return runningVersion.AtLeast(version114), runningVersion.AtLeast(version118), runningVersion.AtLeast(version119)
}

// default path type is Prefix to not break existing definitions
var defaultPathType = networking.PathTypePrefix

// SetDefaultNGINXPathType sets a default PathType when is not defined.
func SetDefaultALBPathType(ing *networking.Ingress) {
	for _, rule := range ing.Spec.Rules {
		if rule.IngressRuleValue.HTTP == nil {
			continue
		}

		for idx := range rule.IngressRuleValue.HTTP.Paths {
			p := &rule.IngressRuleValue.HTTP.Paths[idx]
			if p.PathType == nil {
				p.PathType = &defaultPathType
			}

			if *p.PathType == networking.PathTypeImplementationSpecific {
				p.PathType = &defaultPathType
			}
		}
	}
}

// ObjectRefMap is a map of references from object(s) to object (1:n). It is
// used to keep track of which data objects (Secrets) are used within Ingress
// objects.
type ObjectRefMap interface {
	Insert(consumer string, ref ...string)
	Delete(consumer string)
	Len() int
	Has(ref string) bool
	HasConsumer(consumer string) bool
	Reference(ref string) []string
	ReferencedBy(consumer string) []string
}

type objectRefMap struct {
	sync.Mutex
	v map[string]sets.Set[string]
}

// NewObjectRefMap returns a new ObjectRefMap.
func NewObjectRefMap() ObjectRefMap {
	return &objectRefMap{
		v: make(map[string]sets.Set[string]),
	}
}

// Insert adds a consumer to one or more referenced objects.
func (o *objectRefMap) Insert(consumer string, ref ...string) {
	o.Lock()
	defer o.Unlock()

	for _, r := range ref {
		if _, ok := o.v[r]; !ok {
			o.v[r] = sets.New(consumer)
			continue
		}
		o.v[r].Insert(consumer)
	}
}

// Delete deletes a consumer from all referenced objects.
func (o *objectRefMap) Delete(consumer string) {
	o.Lock()
	defer o.Unlock()

	for ref, consumers := range o.v {
		consumers.Delete(consumer)
		if consumers.Len() == 0 {
			delete(o.v, ref)
		}
	}
}

// Len returns the count of referenced objects.
func (o *objectRefMap) Len() int {
	return len(o.v)
}

// Has returns whether the given object is referenced by any other object.
func (o *objectRefMap) Has(ref string) bool {
	o.Lock()
	defer o.Unlock()

	if _, ok := o.v[ref]; ok {
		return true
	}
	return false
}

// HasConsumer returns whether the store contains the given consumer.
func (o *objectRefMap) HasConsumer(consumer string) bool {
	o.Lock()
	defer o.Unlock()

	for _, consumers := range o.v {
		if consumers.Has(consumer) {
			return true
		}
	}
	return false
}

// Reference returns all objects referencing the given object.
func (o *objectRefMap) Reference(ref string) []string {
	o.Lock()
	defer o.Unlock()

	consumers, ok := o.v[ref]
	if !ok {
		return make([]string, 0)
	}
	return consumers.UnsortedList()
}

// ReferencedBy returns all objects referenced by the given object.
func (o *objectRefMap) ReferencedBy(consumer string) []string {
	o.Lock()
	defer o.Unlock()

	refs := make([]string, 0)
	for ref, consumers := range o.v {
		if consumers.Has(consumer) {
			refs = append(refs, ref)
		}
	}
	return refs
}

// PodLister makes a Store that lists Pods.
type PodLister struct {
	cache.Store
}

// ByKey returns the Endpoints of the Service matching key in the local Endpoint Store.
func (s *PodLister) ByKey(key string) (*apiv1.Pod, error) {
	pod, exists, err := s.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return pod.(*apiv1.Pod), nil
}

// SecretLister makes a Store that lists Secrets.
type SecretLister struct {
	cache.Store
}

// ByKey returns the Secret matching key in the local Secret Store.
func (sl *SecretLister) ByKey(key string) (*apiv1.Secret, error) {
	s, exists, err := sl.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return s.(*apiv1.Secret), nil
}

// ServiceLister makes a Store that lists Services.
type ServiceLister struct {
	cache.Store
}

// ByKey returns the Service matching key in the local Service Store.
func (sl *ServiceLister) ByKey(key string) (*apiv1.Service, error) {
	s, exists, err := sl.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return s.(*apiv1.Service), nil
}
