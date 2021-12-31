/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package store

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"k8s.io/utils/pointer"

	"github.com/eapache/channels"

	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

// IngressFilterFunc decides if an Ingress should be omitted or not
type IngressFilterFunc func(*Ingress) bool

// Storer is the interface that wraps the required methods to gather information
// about ingresses, services, secrets and ingress annotations.
type Storer interface {

	// GetService returns the Service matching key.
	GetService(key string) (*corev1.Service, error)

	// GetServiceEndpoints returns the Endpoints of a Service matching key.
	GetServiceEndpoints(key string) (*corev1.Endpoints, error)

	GetPod(key string) (*corev1.Pod, error)

	// ListIngresses returns a list of all Ingresses in the store.
	ListIngresses() []*Ingress

	// Run initiates the synchronization of the controllers
	Run(stopCh chan struct{})
}

// Informer defines the required SharedIndexInformers that interact with the API server.
type Informer struct {
	Ingress  cache.SharedIndexInformer
	Endpoint cache.SharedIndexInformer
	Service  cache.SharedIndexInformer
	Node     cache.SharedIndexInformer
	Pod      cache.SharedIndexInformer
}

// Lister contains object listers (stores).
type Lister struct {
	Ingress               IngressLister
	Service               ServiceLister
	Endpoint              EndpointLister
	Pod                   PodLister
	Node                  NodeLister
	IngressWithAnnotation IngressWithAnnotationsLister
}

// NotExistsError is returned when an object does not exist in a local store.
type NotExistsError string

// Error implements the error interface.
func (e NotExistsError) Error() string {
	return fmt.Sprintf("no object matching key %q in local store", string(e))
}

// Run initiates the synchronization of the informers against the API server.
func (i *Informer) Run(stopCh chan struct{}) {
	go i.Pod.Run(stopCh)
	go i.Endpoint.Run(stopCh)
	go i.Service.Run(stopCh)
	go i.Node.Run(stopCh)
	// wait for all involved caches to be synced before processing items
	// from the queue
	if !cache.WaitForCacheSync(stopCh,
		i.Pod.HasSynced,
		i.Endpoint.HasSynced,
		i.Service.HasSynced,
		i.Node.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}

	// in big clusters, deltas can keep arriving even after HasSynced
	// functions have returned 'true'
	time.Sleep(1 * time.Second)

	// we can start syncing ingress objects only after other caches are
	// ready, because ingress rules require content from other listers, and
	// 'add' events get triggered in the handlers during caches population.
	go i.Ingress.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh,
		i.Ingress.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}
}

// k8sStore internal Storer implementation using informers and thread safe stores
type k8sStore struct {

	// informer contains the cache Informers
	informers *Informer

	// listers contains the cache.Store interfaces used in the ingress controller
	listers *Lister

	// secretIngressMap contains information about which ingress references a
	// secret in the annotations.
	secretIngressMap ObjectRefMap

	// updateCh
	updateCh *channels.RingChannel

	// syncSecretMu protects against simultaneous invocations of syncSecret
	syncSecretMu *sync.Mutex

	// backendConfigMu protects against simultaneous read/write of backendConfig
	backendConfigMu *sync.RWMutex
}

// New creates a new object store to be used in the ingress controller
func New(
	namespace string,
	resyncPeriod time.Duration,
	client clientset.Interface,
	updateCh *channels.RingChannel,
	disableCatchAll bool) Storer {

	store := &k8sStore{
		informers:        &Informer{},
		listers:          &Lister{},
		updateCh:         updateCh,
		syncSecretMu:     &sync.Mutex{},
		backendConfigMu:  &sync.RWMutex{},
		secretIngressMap: NewObjectRefMap(),
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&clientcorev1.EventSinkImpl{
		Interface: client.CoreV1().Events(namespace),
	})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{
		Component: "alb-ingress-controller",
	})

	store.listers.IngressWithAnnotation.Store = cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	// create informers factory, enable and assign required informers
	infFactory := informers.NewSharedInformerFactoryWithOptions(client, resyncPeriod,
		informers.WithNamespace(namespace),
	)

	store.informers.Ingress = infFactory.Networking().V1().Ingresses().Informer()
	store.listers.Ingress.Store = store.informers.Ingress.GetStore()

	store.informers.Endpoint = infFactory.Core().V1().Endpoints().Informer()
	store.listers.Endpoint.Store = store.informers.Endpoint.GetStore()

	store.informers.Service = infFactory.Core().V1().Services().Informer()
	store.listers.Service.Store = store.informers.Service.GetStore()

	store.informers.Node = infFactory.Core().V1().Nodes().Informer()
	store.listers.Node.Store = store.informers.Node.GetStore()

	store.informers.Pod = infFactory.Core().V1().Pods().Informer()
	store.listers.Pod.Store = store.informers.Pod.GetStore()

	ingDeleteHandler := func(obj interface{}) {
		ing, ok := toIngress(obj)
		if !ok {
			// If we reached here it means the ingress was deleted but its final state is unrecorded.
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				klog.ErrorS(nil, "Error obtaining object from tombstone", "key", obj)
				return
			}
			ing, ok = tombstone.Obj.(*networking.Ingress)
			if !ok {
				klog.Errorf("Tombstone contained object that is not an Ingress: %#v", obj)
				return
			}
		}

		if !IsValid(ing) {
			return
		}

		if isCatchAllIngress(ing.Spec) && disableCatchAll {
			klog.InfoS("Ignoring delete for catch-all because of --disable-catch-all", "ingress", klog.KObj(ing))
			return
		}

		//store.listers.IngressWithAnnotation.Delete(ing)

		key := MetaNamespaceKey(ing)
		store.secretIngressMap.Delete(key)

		updateCh.In() <- helper.Event{
			Type: helper.IngressDeleteEvent,
			Obj:  obj,
		}
	}

	ingEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ing, ok := toIngress(obj)
			if !ok {
				return
			}
			if !IsValid(ing) {
				ingressClass, _ := annotations.GetStringAnnotation(IngressKey, ing)
				klog.InfoS("Ignoring ingress", "ingress", klog.KObj(ing), "kubernetes.io/ingress.class", ingressClass, "ingressClassName", pointer.StringPtrDerefOr(ing.Spec.IngressClassName, ""))
				return
			}

			if isCatchAllIngress(ing.Spec) && disableCatchAll {
				klog.InfoS("Ignoring add for catch-all ingress because of --disable-catch-all", "ingress", klog.KObj(ing))
				return
			}

			recorder.Eventf(ing, corev1.EventTypeNormal, "Sync", "Scheduled for sync")

			store.syncIngress(ing)

			updateCh.In() <- helper.Event{
				Type: helper.CreateEvent,
				Obj:  obj,
			}
		},
		DeleteFunc: ingDeleteHandler,
		UpdateFunc: func(old, cur interface{}) {
			oldIng, ok := toIngress(old)
			if !ok {
				return
			}
			curIng, ok := toIngress(cur)
			if !ok {
				return
			}
			validOld := IsValid(oldIng)
			validCur := IsValid(curIng)
			if !validOld && validCur {
				if isCatchAllIngress(curIng.Spec) && disableCatchAll {
					klog.InfoS("ignoring update for catch-all ingress because of --disable-catch-all", "ingress", klog.KObj(curIng))
					return
				}

				klog.InfoS("creating ingress", "ingress", klog.KObj(curIng), "class", IngressKey)
				recorder.Eventf(curIng, corev1.EventTypeNormal, "Sync", "Scheduled for sync")
			} else if validOld && !validCur {
				klog.InfoS("removing ingress", "ingress", klog.KObj(curIng), "class", IngressKey)
				ingDeleteHandler(old)
				return
			} else if validCur && !reflect.DeepEqual(old, cur) {
				if isCatchAllIngress(curIng.Spec) && disableCatchAll {
					klog.InfoS("ignoring update for catch-all ingress and delete old one because of --disable-catch-all", "ingress", klog.KObj(curIng))
					ingDeleteHandler(old)
					return
				}

				recorder.Eventf(curIng, corev1.EventTypeNormal, "Sync", "Scheduled for sync")
			} else {
				klog.V(3).InfoS("No changes on ingress. Skipping update", "ingress", klog.KObj(curIng))
				return
			}

			store.syncIngress(curIng)

			updateCh.In() <- helper.Event{
				Type: helper.UpdateEvent,
				Obj:  cur,
			}
		},
	}

	epEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ep1 := obj.(*corev1.Endpoints)
			key := MetaNamespaceKey(ep1)
			svc, exist, err := store.listers.Service.GetByKey(key)
			if err != nil {
				klog.Error(err, "get service GetByKey by endpoint failed", "endpoint", ep1)
				return
			}
			if !exist {
				klog.Warningf("epEventHandler %s", key)
				return
			}
			s := svc.(*corev1.Service)

			klog.Info("controller: endpoint add event",
				util.NamespacedName(ep1).String())
			updateCh.In() <- helper.Event{
				Type: helper.EndPointEvent,
				Obj:  s,
			}
		},
		DeleteFunc: func(obj interface{}) {
			ep1 := obj.(*corev1.Endpoints)
			key := MetaNamespaceKey(ep1)
			svc, exist, err := store.listers.Service.GetByKey(key)
			if err != nil {
				klog.Error(err, "DeleteFunc get service GetByKey by endpoint failed", "endpoint", ep1)
				return
			}
			if !exist {
				klog.Warningf("DeleteFunc epEventHandler %s", key)
				return
			}

			s := svc.(*corev1.Service)

			klog.Info("controller: endpoint delete event",
				util.NamespacedName(ep1).String())
			updateCh.In() <- helper.Event{
				Type: helper.EndPointEvent,
				Obj:  s,
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			ep1 := old.(*corev1.Endpoints)
			ep2 := cur.(*corev1.Endpoints)
			if !reflect.DeepEqual(ep1.Subsets, ep2.Subsets) {
				key := MetaNamespaceKey(ep1)
				svc, exist, err := store.listers.Service.GetByKey(key)
				if err != nil {
					klog.Error(err, "UpdateFunc get service GetByKey by endpoint failed", "endpoint", ep1)
					return
				}
				if !exist {
					klog.Warningf("UpdateFunc epEventHandler %s", key)
					return
				}
				s := svc.(*corev1.Service)

				klog.Info("controller: endpoint update event",
					util.NamespacedName(ep1).String())
				updateCh.In() <- helper.Event{
					Type: helper.EndPointEvent,
					Obj:  s,
				}
			}
		},
	}
	podEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			err := store.listers.Pod.Add(obj)
			if err != nil {
				klog.Error(err, "Pod Add failed")
				return
			}
		},
		DeleteFunc: func(obj interface{}) {
			store.listers.Pod.Delete(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
		},
	}
	nodeEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			serviceList := store.listers.Service.List()
			for _, v := range serviceList {
				svc := v.(*corev1.Service)
				klog.Info("node change: enqueue service", util.Key(svc))
				updateCh.In() <- helper.Event{
					Type: helper.NodeEvent,
					Obj:  svc,
				}
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			nodeOld := old.(*corev1.Node)
			nodeNew := cur.(*corev1.Node)

			if !reflect.DeepEqual(nodeOld.Labels, nodeNew.Labels) {
				serviceList := store.listers.Service.List()
				for _, v := range serviceList {
					svc := v.(*corev1.Service)
					klog.Info("node change: enqueue service", util.Key(svc))
					updateCh.In() <- helper.Event{
						Type: helper.NodeEvent,
						Obj:  svc,
					}
				}
			}
		},

		DeleteFunc: func(obj interface{}) {
			serviceList := store.listers.Service.List()
			for _, v := range serviceList {
				svc := v.(*corev1.Service)
				klog.Info("node change: enqueue service", util.Key(svc))
				updateCh.In() <- helper.Event{
					Type: helper.NodeEvent,
					Obj:  svc,
				}
			}

		},
	}

	serviceHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			curSvc := obj.(*corev1.Service)
			store.enqueueImpactedIngresses(updateCh, curSvc)
		},
		UpdateFunc: func(old, cur interface{}) {
			// update the server group
			oldSvc := old.(*corev1.Service)
			curSvc := cur.(*corev1.Service)

			if reflect.DeepEqual(oldSvc, curSvc) {
				return
			}

			updateCh.In() <- helper.Event{
				Type: helper.ServiceEvent,
				Obj:  cur,
			}
		},
		DeleteFunc: func(obj interface{}) {
			// ingress refer service to delete
			curSvc := obj.(*corev1.Service)
			store.enqueueImpactedIngresses(updateCh, curSvc)
		},
	}

	store.informers.Ingress.AddEventHandler(ingEventHandler)
	store.informers.Endpoint.AddEventHandler(epEventHandler)
	store.informers.Node.AddEventHandler(podEventHandler)
	store.informers.Service.AddEventHandler(serviceHandler)
	store.informers.Node.AddEventHandler(nodeEventHandler)
	return store
}

func (s *k8sStore) enqueueImpactedIngresses(updateCh *channels.RingChannel, svc *corev1.Service) {
	ingList := s.listers.Ingress.List()

	for _, t := range ingList {
		ing := t.(*networking.Ingress)

		if !IsIngressAlbClass(*ing) {
			continue
		}

		updateCh.In() <- helper.Event{
			Type: helper.IngressEvent,
			Obj:  ing,
		}
	}
}

func IsIngressAlbClass(ing networking.Ingress) bool {
	if ingClassAnnotation, exists := ing.Annotations[util.IngressClass]; exists {
		if ingClassAnnotation == IngressClassName {
			return true
		}
	}
	return false
}

// isCatchAllIngress returns whether or not an ingress produces a
// catch-all server, and so should be ignored when --disable-catch-all is set
func isCatchAllIngress(spec networking.IngressSpec) bool {
	return spec.DefaultBackend != nil && len(spec.Rules) == 0
}

// syncIngress parses ingress annotations converting the value of the
// annotation to a go struct
func (s *k8sStore) syncIngress(ing *networking.Ingress) {
	key := MetaNamespaceKey(ing)
	klog.V(3).Infof("updating annotations information for ingress %v", key)
	if !IsValid(ing) {
		return
	}
	copyIng := &networking.Ingress{}
	ing.ObjectMeta.DeepCopyInto(&copyIng.ObjectMeta)
	ing.Spec.DeepCopyInto(&copyIng.Spec)
	ing.Status.DeepCopyInto(&copyIng.Status)

	for ri, rule := range copyIng.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		for pi, path := range rule.HTTP.Paths {
			if path.Path == "" {
				copyIng.Spec.Rules[ri].HTTP.Paths[pi].Path = "/"
			}
		}
	}

	SetDefaultALBPathType(copyIng)

	err := s.listers.IngressWithAnnotation.Update(&Ingress{
		Ingress: *copyIng,
	})
	if err != nil {
		klog.Error(err)
	}
}

// GetService returns the Service matching key.
func (s *k8sStore) GetService(key string) (*corev1.Service, error) {
	return s.listers.Service.ByKey(key)
}

// getIngress returns the Ingress matching key.
func (s *k8sStore) getIngress(key string) (*networking.Ingress, error) {
	ing, err := s.listers.IngressWithAnnotation.ByKey(key)
	if err != nil {
		return nil, err
	}

	return &ing.Ingress, nil
}

func sortIngressSlice(ingresses []*Ingress) {
	// sort Ingresses using the CreationTimestamp field
	sort.SliceStable(ingresses, func(i, j int) bool {
		ir := ingresses[i].CreationTimestamp
		jr := ingresses[j].CreationTimestamp
		if ir.Equal(&jr) {
			in := fmt.Sprintf("%v/%v", ingresses[i].Namespace, ingresses[i].Name)
			jn := fmt.Sprintf("%v/%v", ingresses[j].Namespace, ingresses[j].Name)
			klog.V(3).Infof("Ingress %v and %v have identical CreationTimestamp", in, jn)
			return in > jn
		}
		return ir.Before(&jr)
	})
}

// ListIngresses returns the list of Ingresses
func (s *k8sStore) ListIngresses() []*Ingress {
	// filter ingress rules
	ingresses := make([]*Ingress, 0)
	for _, item := range s.listers.IngressWithAnnotation.List() {
		ing := item.(*Ingress)
		if IsValid(&ing.Ingress) {
			ingresses = append(ingresses, ing)
		}

	}

	sortIngressSlice(ingresses)

	return ingresses
}

// GetServiceEndpoints returns the Endpoints of a Service matching key.
func (s *k8sStore) GetServiceEndpoints(key string) (*corev1.Endpoints, error) {
	return s.listers.Endpoint.ByKey(key)
}

func (s *k8sStore) GetPod(key string) (*corev1.Pod, error) {
	return s.listers.Pod.ByKey(key)
}

// Run initiates the synchronization of the informers and the initial
// synchronization of the secrets.
func (s *k8sStore) Run(stopCh chan struct{}) {
	// start informers
	s.informers.Run(stopCh)
}

var runtimeScheme = k8sruntime.NewScheme()

func init() {
	utilruntime.Must(networking.AddToScheme(runtimeScheme))
}

func toIngress(obj interface{}) (*networking.Ingress, bool) {
	if ing, ok := obj.(*networking.Ingress); ok {
		SetDefaultALBPathType(ing)
		return ing, true
	}

	return nil, false
}

func isLoadBalancerOrNodePortService(svc *corev1.Service) bool {
	return isLoadBalancerService(svc) || isNodePortService(svc)
}
func isLoadBalancerService(svc *corev1.Service) bool {
	return svc.Spec.Type == corev1.ServiceTypeLoadBalancer
}

func isNodePortService(svc *corev1.Service) bool {
	return svc.Spec.Type == corev1.ServiceTypeNodePort
}
func isLocalModeService(svc *corev1.Service) bool {
	return svc.Spec.ExternalTrafficPolicy == corev1.ServiceExternalTrafficPolicyTypeLocal
}
