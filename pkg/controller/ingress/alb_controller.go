package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	sdkutils "github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/eapache/channels"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/annotations"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/applier"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/backend"
	albconfigmanager "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/builder/albconfig_manager"
	servicemanager "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/builder/service_manager"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/store"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"sync"
	"time"
)

const (
	defaultMaxConcurrentReconciles = 3
	albIngressControllerName       = "alb-ingress-controller"
)

func NewAlbConfigReconciler(mgr manager.Manager, ctx *shared.SharedContext) (*albconfigReconciler, error) {
	config := Configuration{}
	logger := ctrl.Log.WithName("controllers").WithName(albIngressControllerName)
	logger.Info("start to register crds")
	err := RegisterCRD(mgr.GetConfig())
	if err != nil {
		logger.Error(err, "register crd: %s", err.Error())
		return nil, err
	}
	client, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}
	n := &albconfigReconciler{
		cloud:            ctx.Provider(),
		k8sClient:        mgr.GetClient(),
		groupLoader:      albconfigmanager.NewDefaultGroupLoader(mgr.GetClient(), annotations.NewSuffixAnnotationParser(annotations.DefaultAnnotationsPrefix)),
		referenceIndexer: helper.NewDefaultReferenceIndexer(),
		eventRecorder:    mgr.GetEventRecorderFor("ingress"),
		stackMarshaller:  NewDefaultStackMarshaller(),
		logger:           logger,
		updateCh:         channels.NewRingChannel(1024),
		albconfigBuilder: albconfigmanager.NewDefaultAlbConfigManagerBuilder(mgr.GetClient(), ctx.Provider(), logger),

		serverApplier: applier.NewServiceManagerApplier(
			mgr.GetClient(),
			ctx.Provider(),
			logger),
		stopLock:              &sync.Mutex{},
		groupFinalizerManager: albconfigmanager.NewDefaultFinalizerManager(helper.NewDefaultFinalizerManager(mgr.GetClient())),
		k8sFinalizerManager:   helper.NewDefaultFinalizerManager(mgr.GetClient()),

		maxConcurrentReconciles: defaultMaxConcurrentReconciles,
	}
	n.store = store.New(
		config.Namespace,
		config.ResyncPeriod,
		client,
		n.updateCh,
		config.DisableCatchAll)
	n.serverBuilder = servicemanager.NewDefaultServiceStackBuilder(backend.NewBackendManager(n.store, mgr.GetClient(), ctx.Provider(), logger))
	n.albconfigApplier = applier.NewAlbConfigManagerApplier(n.store, mgr.GetClient(), ctx.Provider(), util.IngressTagKeyPrefix, logger)
	n.syncQueue = helper.NewTaskQueue(n.syncIngress)
	n.syncServersQueue = helper.NewTaskQueue(n.syncServers)
	go n.Start()
	return n, nil
}

// Configuration contains all the settings required by an Ingress controller
type Configuration struct {
	Client          clientset.Interface
	ResyncPeriod    time.Duration
	ConfigMapName   string
	DefaultService  string
	Namespace       string
	DisableCatchAll bool
}

type albconfigReconciler struct {
	cloud            prvd.Provider
	k8sClient        client.Client
	groupLoader      albconfigmanager.GroupLoader
	referenceIndexer helper.ReferenceIndexer
	eventRecorder    record.EventRecorder
	stackMarshaller  StackMarshaller
	logger           logr.Logger
	store            store.Storer
	albconfigBuilder albconfigmanager.Builder
	albconfigApplier applier.AlbConfigManagerApplier
	serverBuilder    servicemanager.Builder
	serverApplier    applier.ServiceManagerApplier
	isShuttingDown   bool
	stopCh           chan struct{}
	updateCh         *channels.RingChannel
	acEventChan      chan event.GenericEvent
	// ngxErrCh is used to detect errors with the NGINX processes
	ngxErrCh                chan error
	stopLock                *sync.Mutex
	groupFinalizerManager   albconfigmanager.FinalizerManager
	k8sFinalizerManager     helper.FinalizerManager
	syncQueue               *helper.Queue
	syncServersQueue        *helper.Queue
	maxConcurrentReconciles int
}

func (g *albconfigReconciler) setupWatches(_ context.Context, c controller.Controller) error {
	g.acEventChan = make(chan event.GenericEvent)
	acEventHandler := NewEnqueueRequestsForAlbconfigEvent(g.k8sClient, g.eventRecorder, g.logger)
	if err := c.Watch(&source.Channel{Source: g.acEventChan}, acEventHandler); err != nil {
		return err
	}

	if err := c.Watch(&source.Kind{Type: &v1.AlbConfig{}}, acEventHandler); err != nil {
		return err
	}

	return nil
}

func (g *albconfigReconciler) SetupWithManager(ctx context.Context, mgr manager.Manager) error {
	c, err := controller.New(albIngressControllerName, mgr, controller.Options{
		MaxConcurrentReconciles: g.maxConcurrentReconciles,
		Reconciler:              g,
	})
	if err != nil {
		return err
	}

	if err := g.setupWatches(ctx, c); err != nil {
		return err
	}

	return nil
}

func (g *albconfigReconciler) syncIngress(obj interface{}) error {
	g.logger.Info("start syncIngress")
	traceID := sdkutils.GetUUID()
	ctx := context.WithValue(context.Background(), util.TraceID, traceID)
	e := obj.(helper.Element)
	evt := e.Event
	ings := g.store.ListIngresses()

	if len(ings) == 0 {
		g.logger.Info("ingress length: 0, skip")
		return nil
	}
	if evt.Type == helper.EndPointEvent || evt.Type == helper.NodeEvent || evt.Type == helper.ServiceEvent {
		g.syncServersQueue.EnqueueSkippableTask(evt)
		return nil
	}
	for _, ing := range ings {
		g.logger.Info("start ingress: %s", ing.Name)
		var (
			groupIDNew *albconfigmanager.GroupID
		)
		groupIDNew, _ = g.groupLoader.LoadGroupID(ctx, &ing.Ingress)
		var groupInChan = func(groupID *albconfigmanager.GroupID) {
			albconfig := &v1.AlbConfig{}
			if err := g.k8sClient.Get(ctx, types.NamespacedName{
				Namespace: groupID.Namespace,
				Name:      groupID.Name,
			}, albconfig); err != nil {
				if errors.IsNotFound(err) {
					err = g.k8sClient.Create(ctx, g.makeAlbConfig(ctx, groupID.Name, &ing.Ingress), &client.CreateOptions{})
					if err != nil {
						g.logger.Error(err, "Create albconfig failed", "albconfig", ing)
						return
					}
					return
				}
				g.logger.Error(err, "get albconfig failed", "albconfig", ing)
				return
			}
			lss := make([]*v1.ListenerSpec, 0)
			ingListByPort := make(map[int32]albconfigmanager.Protocol)
			ingGroup, _ := g.groupLoader.Load(ctx, *groupID, ings)
			if ingGroup.Members != nil && len(ingGroup.Members) > 0 {
				for _, ingm := range ingGroup.Members {
					portAndProtocol, _ := albconfigmanager.ComputeIngressListenPorts(ingm)
					for port, pro := range portAndProtocol {
						ingListByPort[port] = pro
					}
				}
			}
			for k, v := range ingListByPort {
				ls := &v1.ListenerSpec{
					Port:     intstr.FromInt(int(k)),
					Protocol: string(v),
				}
				lss = append(lss, ls)
			}
			albconfig.Spec.Listeners = lss
			err := g.k8sClient.Update(ctx, albconfig, &client.UpdateOptions{})
			if err != nil {
				g.logger.Error(err, "Update albconfig")
				return
			}
			g.acEventChan <- event.GenericEvent{
				Object: albconfig,
			}
		}

		groupInChan(groupIDNew)

	}

	return nil
}

func (g *albconfigReconciler) syncServers(obj interface{}) error {
	traceID := sdkutils.GetUUID()
	ctx := context.WithValue(context.Background(), util.TraceID, traceID)

	e := obj.(helper.Element)
	evt := e.Event

	startTime := time.Now()
	g.logger.Info("start syncServers",
		"request", e.Key,
		"traceID", traceID,
		"startTime", startTime)
	defer func() {
		g.logger.Info("finish syncServers",
			"request", e.Key,
			"traceID", traceID,
			"elapsedTime", time.Since(startTime).Milliseconds())
	}()

	ings := g.store.ListIngresses()
	if len(ings) == 0 {
		g.logger.Info("service not used by ingress, skip", "key", e.Key)
		return nil
	}
	request := reconcile.Request{}
	svc := evt.Obj.(*corev1.Service)
	request.Namespace = svc.Namespace
	request.Name = svc.Name

	servicePortToIngressNames := g.getServicePortToIngressNames(request, ings)
	if len(servicePortToIngressNames) > 0 {
		svcStackContext, err := g.buildServiceStackContext(ctx, request, servicePortToIngressNames)
		if err != nil {
			return err
		}

		if err = g.buildAndApplyServers(ctx, svcStackContext); err != nil {
			return err
		}
	}

	return nil
}

func (g *albconfigReconciler) getServicePortToIngressNames(request reconcile.Request, ingList []*store.Ingress) map[int32][]string {

	var servicePortToIngressNames = make(map[int32]map[string]struct{})

	var processIngressBackend = func(b networking.IngressBackend, ingName string) {
		servicePort := b.Service.Port.Number
		if _, ok := servicePortToIngressNames[servicePort]; !ok {
			servicePortToIngressNames[servicePort] = make(map[string]struct{})
		}
		servicePortToIngressNames[servicePort][ingName] = struct{}{}
	}

	for _, ing := range ingList {
		if ing.Spec.DefaultBackend != nil {
			if ing.Spec.DefaultBackend.Service.Name == request.Name {
				processIngressBackend(*ing.Spec.DefaultBackend, ing.Name)
			}
		}

		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service.Name == request.Name {
					g.logger.Info("processIngressBackend", "ServiceName", path.Backend.Service.Name, request.Name)
					processIngressBackend(path.Backend, ing.Name)
				}
			}
		}
	}

	var servicePortToIngressNameList = make(map[int32][]string)
	for servicePort, ingressNames := range servicePortToIngressNames {
		for ingressName := range ingressNames {
			servicePortToIngressNameList[servicePort] = append(servicePortToIngressNameList[servicePort], ingressName)
		}
	}

	return servicePortToIngressNameList
}

func (g *albconfigReconciler) buildServiceStackContext(ctx context.Context, request reconcile.Request, serverPortToIngressNames map[int32][]string) (*albmodel.ServiceStackContext, error) {
	var svcStackContext = &albmodel.ServiceStackContext{
		ClusterID:                 g.cloud.ClusterID(),
		ServiceNamespace:          request.Namespace,
		ServiceName:               request.Name,
		ServicePortToIngressNames: serverPortToIngressNames,
	}

	svc := &corev1.Service{}
	if err := g.k8sClient.Get(ctx, request.NamespacedName, svc); err != nil {
		// if not found service, need to delete
		if errors.IsNotFound(err) {
			g.logger.Info("service not found", "ServiceNotFound", request.NamespacedName.Namespace+"/"+request.NamespacedName.Name)
			svcStackContext.IsServiceNotFound = true
		} else {
			return nil, err
		}
	} else {
		svcStackContext.Service = svc
	}

	return svcStackContext, nil
}

func (s *albconfigReconciler) buildAndApplyServers(ctx context.Context, svcStackCtx *albmodel.ServiceStackContext) error {
	traceID := ctx.Value(util.TraceID)

	buildStartTime := time.Now()
	s.logger.Info("start build and apply stack",
		"service", svcStackCtx.ServiceNamespace+"/"+svcStackCtx.ServiceName,
		"traceID", traceID,
		"startTime", buildStartTime)

	serverStack, err := s.serverBuilder.Build(ctx, svcStackCtx)
	if err != nil {
		return fmt.Errorf("build service stack model error: %v", err)
	}

	serviceStackJson, err := json.Marshal(serverStack)
	if err != nil {
		return err
	}

	s.logger.Info("successfully built service stack",
		"service", svcStackCtx.ServiceNamespace+"/"+svcStackCtx.ServiceName,
		"traceID", traceID,
		"stack", string(serviceStackJson),
		"buildElapsedTime", time.Since(buildStartTime).Milliseconds())

	applyStartTime := time.Now()
	err = s.serverApplier.Apply(ctx, s.cloud, serverStack)
	if err != nil {
		return err
	}

	if serverStack.ContainsPotentialReadyEndpoints {
		return fmt.Errorf("retry potential ready endpoints")
	}
	s.logger.Info("successfully applied service stack",
		"service", svcStackCtx.ServiceNamespace+"/"+svcStackCtx.ServiceName,
		"traceID", traceID,
		"applyElapsedTime", time.Since(applyStartTime).Milliseconds())

	return nil
}

func (g *albconfigReconciler) makeAlbConfig(ctx context.Context, groupName string, ing *networking.Ingress) *v1.AlbConfig {
	id, _ := annotations.GetStringAnnotation(annotations.LoadBalancerId, ing)
	overrideListener := false
	if override, err := annotations.GetStringAnnotation(annotations.OverrideListener, ing); err == nil {
		if override == "true" {
			overrideListener = true
		}
	}

	name, _ := annotations.GetStringAnnotation(annotations.LoadBalancerName, ing)
	addressType, err := annotations.GetStringAnnotation(annotations.AddressType, ing)
	if err != nil {
		addressType = util.LoadBalancerAddressTypeInternet
	}
	addressAllocatedMode, err := annotations.GetStringAnnotation(annotations.AddressAllocatedMode, ing)
	if err != nil {
		addressAllocatedMode = util.LoadBalancerAddressAllocatedModeDynamic
	}
	chargeType, err := annotations.GetStringAnnotation(annotations.ChargeType, ing)
	if err != nil {
		chargeType = util.LoadBalancerPayTypePostPay
	}
	loadBalancerEdition, err := annotations.GetStringAnnotation(annotations.LoadBalancerEdition, ing)
	if err != nil {
		loadBalancerEdition = util.LoadBalancerEditionStandard
	}
	vswitchIds, _ := annotations.GetStringAnnotation(annotations.VswitchIds, ing)
	var al albmodel.AccessLog
	if accessLog, err := annotations.GetStringAnnotation(annotations.AccessLog, ing); err == nil {
		if err := json.Unmarshal([]byte(accessLog), &al); err != nil {
			g.logger.Error(err, "Unmarshal: %s", annotations.AccessLog)
		}
	}

	albconfig := &v1.AlbConfig{}
	albconfig.Name = groupName
	albconfig.Namespace = albconfigmanager.ALBConfigNamespace
	deletionProtectionEnabled := true
	albconfig.Spec.LoadBalancer = &v1.LoadBalancerSpec{
		Id:                        id,
		ForceOverride:             &overrideListener,
		Name:                      name,
		AddressAllocatedMode:      addressAllocatedMode,
		AddressType:               addressType,
		DeletionProtectionEnabled: &deletionProtectionEnabled,
		BillingConfig: v1.BillingConfig{
			PayType: chargeType,
		},
		Edition: loadBalancerEdition,
		AccessLogConfig: v1.AccessLogConfig{
			LogStore:   al.LogStore,
			LogProject: al.LogProject,
		},
	}
	vSwitchIdss := strings.Split(vswitchIds, ",")
	zoneMappings := make([]v1.ZoneMapping, 0)
	for _, vSwitchId := range vSwitchIdss {
		zoneMappings = append(zoneMappings, v1.ZoneMapping{
			VSwitchId: vSwitchId,
		})
	}
	albconfig.Spec.LoadBalancer.ZoneMappings = zoneMappings
	portAndProtocol, _ := albconfigmanager.ComputeIngressListenPorts(ing)
	lss := make([]*v1.ListenerSpec, 0)
	for port, pro := range portAndProtocol {
		ls := &v1.ListenerSpec{
			Port:     intstr.FromInt(int(port)),
			Protocol: string(pro),
		}
		lss = append(lss, ls)
	}
	albconfig.Spec.Listeners = lss

	return albconfig
}

func (g *albconfigReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	traceID := sdkutils.GetUUID()
	ctx = context.WithValue(ctx, util.TraceID, traceID)

	var err error
	startTime := time.Now()
	g.logger.Info("start reconcile",
		"request", req.String(),
		"traceID", traceID,
		"startTime", startTime)
	defer func() {
		if err != nil {
			g.logger.Error(err, "finish reconcile",
				"request", req.String(),
				"traceID", traceID,
				"elapsedTime", time.Since(startTime).Milliseconds())
			return
		}
		g.logger.Info("finish reconcile",
			"request", req.String(),
			"traceID", traceID,
			"elapsedTime", time.Since(startTime).Milliseconds())
	}()

	err = g.reconcile(ctx, req)
	return reconcile.Result{}, err
}

func (g *albconfigReconciler) reconcile(ctx context.Context, request reconcile.Request) error {
	albconfig := &v1.AlbConfig{}
	if err := g.k8sClient.Get(ctx, request.NamespacedName, albconfig); err != nil {
		return client.IgnoreNotFound(err)
	}
	ings := g.store.ListIngresses()
	ingGroup, err := g.groupLoader.Load(ctx, albconfigmanager.GroupID(request.NamespacedName), ings)
	if err != nil {
		return err
	}

	if !albconfig.DeletionTimestamp.IsZero() {
		if err := g.cleanupAlbLoadBalancerResources(ctx, albconfig, ingGroup); err != nil {
			return err
		}
	} else {
		if err := g.reconcileAlbLoadBalancerResources(ctx, albconfig, ingGroup); err != nil {
			return err
		}
	}

	if len(ingGroup.InactiveMembers) != 0 {
		if err := g.groupFinalizerManager.RemoveGroupFinalizer(ctx, ingGroup.InactiveMembers); err != nil {
			g.recordIngressGroupEvent(ctx, ingGroup, corev1.EventTypeWarning, helper.IngressEventReasonFailedRemoveFinalizer, helper.GetLogMessage(err))
			return err
		}
	}

	g.recordIngressGroupEvent(ctx, ingGroup, corev1.EventTypeNormal, helper.IngressEventReasonSuccessfullyReconciled, "Successfully reconciled")

	return nil
}

func (g *albconfigReconciler) cleanupAlbLoadBalancerResources(ctx context.Context, albconfig *v1.AlbConfig, ingGroup *albconfigmanager.Group) error {
	acFinalizer := albconfigmanager.GetIngressFinalizer()
	if helper.HasFinalizer(albconfig, acFinalizer) {
		_, _, err := g.buildAndApply(ctx, albconfig, ingGroup)
		if err != nil {
			return err
		}
		if err := g.k8sFinalizerManager.RemoveFinalizers(ctx, albconfig, acFinalizer); err != nil {
			g.eventRecorder.Event(albconfig, corev1.EventTypeWarning, helper.IngressEventReasonFailedRemoveFinalizer, fmt.Sprintf("Failed remove finalizer due to %v", err))
			return err
		}
		if len(ingGroup.Members) != 0 {
			if err := g.groupFinalizerManager.RemoveGroupFinalizer(ctx, ingGroup.Members); err != nil {
				g.recordIngressGroupEvent(ctx, ingGroup, corev1.EventTypeWarning, helper.IngressEventReasonFailedRemoveFinalizer, fmt.Sprintf("Failed remove finalizer due to %v", err))
				return err
			}
		}
	}
	return nil
}

func (g *albconfigReconciler) reconcileAlbLoadBalancerResources(ctx context.Context, albconfig *v1.AlbConfig, ingGroup *albconfigmanager.Group) error {
	acFinalizer := albconfigmanager.GetIngressFinalizer()
	if err := g.k8sFinalizerManager.AddFinalizers(ctx, albconfig, acFinalizer); err != nil {
		g.eventRecorder.Event(albconfig, corev1.EventTypeWarning, helper.IngressEventReasonFailedRemoveFinalizer, helper.GetLogMessage(err))
		return err
	}
	if err := g.groupFinalizerManager.AddGroupFinalizer(ctx, ingGroup.Members); err != nil {
		g.recordIngressGroupEvent(ctx, ingGroup, corev1.EventTypeWarning, helper.IngressEventReasonFailedAddFinalizer, fmt.Sprintf("Failed add finalizer due to %v", err))
		return err
	}

	_, lb, err := g.buildAndApply(ctx, albconfig, ingGroup)
	if err != nil {
		return err
	}
	if lb.Status == nil || lb.Status.DNSName == "" {
		return nil
	}
	for _, ing := range ingGroup.Members {
		if ing.Status.LoadBalancer.Ingress != nil && len(ing.Status.LoadBalancer.Ingress) > 0 && ing.Status.LoadBalancer.Ingress[0].Hostname == lb.Status.DNSName {
			continue
		}
		lbi := netv1.IngressLoadBalancerIngress{
			Hostname: lb.Status.DNSName,
		}
		ing.Status.LoadBalancer.Ingress = []netv1.IngressLoadBalancerIngress{lbi}
		err = g.k8sClient.Status().Update(ctx, ing, &client.SubResourceUpdateOptions{})
		if err != nil {
			g.logger.Error(err, "Ingress Status Update %s, error: %s", ing.Name)
			continue
		}
	}
	if albconfig.Status.LoadBalancer.DNSName == lb.Status.DNSName {
		return nil
	}
	albconfig.Status.LoadBalancer.Id = lb.Status.LoadBalancerID
	albconfig.Status.LoadBalancer.DNSName = lb.Status.DNSName

	err = g.k8sClient.Status().Update(ctx, albconfig, &client.SubResourceUpdateOptions{})
	if err != nil {
		g.logger.Error(err, "LB Status Update %s, error: %s", albconfig.Name)
		return err
	}
	return nil
}

func (g *albconfigReconciler) buildAndApply(ctx context.Context, albconfig *v1.AlbConfig, ingGroup *albconfigmanager.Group) (core.Manager, *albmodel.AlbLoadBalancer, error) {
	traceID := ctx.Value(util.TraceID)

	buildStartTime := time.Now()
	g.logger.Info("start build and apply stack",
		"albconfig", util.NamespacedName(albconfig).String(),
		"traceID", traceID,
		"startTime", buildStartTime)

	stack, lb, err := g.albconfigBuilder.Build(ctx, albconfig, ingGroup)
	if err != nil {
		g.recordIngressGroupEvent(ctx, ingGroup, corev1.EventTypeWarning, helper.IngressEventReasonFailedBuildModel, helper.GetLogMessage(err))
		return nil, nil, err
	}

	stackJSON, err := g.stackMarshaller.Marshal(stack)
	if err != nil {
		g.recordIngressGroupEvent(ctx, ingGroup, corev1.EventTypeWarning, helper.IngressEventReasonFailedBuildModel, helper.GetLogMessage(err))
		return nil, nil, err
	}

	g.logger.Info("successfully built albconfig stack",
		"albconfig", util.NamespacedName(albconfig).String(),
		"traceID", &traceID,
		"stack", stackJSON,
		"buildElapsedTime", time.Since(buildStartTime).Milliseconds())

	applyStartTime := time.Now()
	if err := g.albconfigApplier.Apply(ctx, stack); err != nil {
		g.recordIngressGroupEvent(ctx, ingGroup, corev1.EventTypeWarning, helper.IngressEventReasonFailedApplyModel, helper.GetLogMessage(err))
		return nil, nil, err
	}
	g.logger.Info("successfully applied albconfig stack",
		"albconfig", util.NamespacedName(albconfig).String(),
		"traceID", traceID,
		"applyElapsedTime", time.Since(applyStartTime).Milliseconds())

	return stack, lb, nil
}

func (g *albconfigReconciler) recordIngressGroupEvent(_ context.Context, ingGroup *albconfigmanager.Group, eventType string, reason string, message string) {
	for _, member := range ingGroup.Members {
		g.eventRecorder.Event(member, eventType, reason, message)
	}
}

// Start starts a new ALB master process running in the foreground.
func (n *albconfigReconciler) Start() {
	n.logger.Info("Starting ALB Ingress controller")
	n.store.Run(n.stopCh)
	go n.syncQueue.Run(1, time.Second, n.stopCh)
	go n.syncServersQueue.Run(3, time.Second, n.stopCh)
	for {
		select {
		case err := <-n.ngxErrCh:
			if n.isShuttingDown {
				return
			}
			n.logger.Error(err, "ErrCh received")

		case event := <-n.updateCh.Out():
			if n.isShuttingDown {
				break
			}

			if evt, ok := event.(helper.Event); ok {
				n.logger.Info("Event received", "type", evt.Type, "object", evt.Obj)

				n.syncQueue.EnqueueSkippableTask(evt)
			} else {
				n.logger.Info("Unexpected event type received %T", event)
			}
		case <-n.stopCh:
			return
		}
	}
}

// Stop gracefully stops the alb master process.
func (n *albconfigReconciler) Stop() error {
	n.isShuttingDown = true

	n.stopLock.Lock()
	defer n.stopLock.Unlock()

	if n.syncQueue.IsShuttingDown() {
		return fmt.Errorf("shutdown already in progress")
	}
	n.logger.Info("Shutting down controller queues")
	close(n.stopCh)
	go n.syncQueue.Shutdown()

	return nil
}

type StackMarshaller interface {
	Marshal(stack core.Manager) (string, error)
}

func NewDefaultStackMarshaller() *defaultStackMarshaller {
	return &defaultStackMarshaller{}
}

var _ StackMarshaller = &defaultStackMarshaller{}

type defaultStackMarshaller struct{}

func (m *defaultStackMarshaller) Marshal(stack core.Manager) (string, error) {
	builder := albconfigmanager.NewStackSchemaBuilder(stack.StackID())
	if err := stack.TopologicalTraversal(builder); err != nil {
		return "", err
	}
	stackSchema := builder.Build()
	payload, err := json.Marshal(stackSchema)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
