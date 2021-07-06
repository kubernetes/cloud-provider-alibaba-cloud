package framework

import (
	"context"
	"encoding/json"
	"fmt"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

const NameSpace = "e2etest"

type OptionsFunc func(f *FrameWorkE2E)

//FrameWorkE2E e2e backup
type FrameWorkE2E struct {
	Description  string
	ModelBuilder service.ModelBuilder
	Cloud        prvd.Provider
	Client       kubernetes.Interface
}

var TestConfig Config

type Config struct {
	Host                  string
	CloudConfig           string
	KubeConfig            string
	LoadBalancerID        string
	MasterZoneID          string
	SlaveZoneID           string
	BackendLabel          string
	AclID                 string
	VSwitchID             string
	CertID                string
	PrivateZoneID         string
	PrivateZoneName       string
	PrivateZoneRecordName string
	PrivateZoneRecordTTL  string
	TestLabel             string
	ResourceGroupID       string
}

func NewBaseSVC(anno map[string]string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "basic-service",
			Namespace:   NameSpace,
			Annotations: anno,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{

					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
			Selector:        map[string]string{"run": "nginx"},
		},
	}
}

func NewNamedFrameWork(name string) *FrameWorkE2E {
	return NewFrameWork(func(f *FrameWorkE2E) { f.Description = name })
}

func NewFrameWork(
	option OptionsFunc,
) *FrameWorkE2E {
	frame := &FrameWorkE2E{}
	if option != nil {
		option(frame)
	}
	frame.Client = NewClientOrDie()
	frame.Cloud = alibaba.NewAlibabaCloud()
	runtimeClient := NewRuntimeClient()
	frame.ModelBuilder = service.ModelBuilder{
		LoadBalancerMgr: service.NewLoadBalancerManager(frame.Cloud),
		ListenerMgr:     service.NewListenerManager(frame.Cloud),
		VGroupMgr:       service.NewVGroupManager(runtimeClient, frame.Cloud),
	}
	return frame
}

type ServiceMutator func(service *v1.Service) error

func NewTestUnit(
	svc *v1.Service,
	mutate ServiceMutator,
	expect func(m *Expectation) (bool, error),
	description string,
) *TestUnit {
	if svc == nil {
		svc = NewBaseSVC(nil)
	}
	return &TestUnit{
		Service:     svc,
		Mutator:     mutate,
		ExpectOK:    expect,
		Description: description,
	}
}

type TestUnit struct {
	Service     *v1.Service
	ReqCtx      *service.RequestContext
	Description string
	Mutator     ServiceMutator
	ExpectOK    func(m *Expectation) (bool, error)
}

func (t *TestUnit) NewReqContext(cloud prvd.Provider) {
	t.ReqCtx = &service.RequestContext{
		Ctx:     context.Background(),
		Service: t.Service,
		Anno:    service.NewAnnotationRequest(t.Service),
	}
}

func NewExpection(mcase *TestUnit, e2e *FrameWorkE2E) *Expectation {
	return &Expectation{Case: mcase, E2E: e2e}
}

type Expectation struct {
	Case *TestUnit
	E2E  *FrameWorkE2E
}

func (m *Expectation) ExpectOK() (bool, error) {
	if m.Case == nil || m.Case.ExpectOK == nil {
		// no expectation found, default to succeed
		return true, nil
	}
	return m.Case.ExpectOK(m)
}

func (f *FrameWorkE2E) SetUp() { f.BeforeEach() }

func (f *FrameWorkE2E) CleanUp() { f.AfterEach() }

func (f *FrameWorkE2E) BeforeEach() {
	setup := func() (done bool, err error) {
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      NameSpace,
				Namespace: NameSpace,
			},
		}
		_, err = f.
			Client.
			CoreV1().
			Namespaces().
			Create(context.Background(), ns, metav1.CreateOptions{})
		if err != nil {
			if !strings.Contains(err.Error(), "exist") {
				Logf("retry initialize namespace: %s", err.Error())
				return false, nil
			}
			Logf("namespace %s exist", NameSpace)
		}
		err = RunNginxDeployment(f.Client)
		if err != nil {
			Logf("retry create nginx: %s", err.Error())
			return false, nil
		}
		return true, nil
	}
	ExpectNoError(wait.PollImmediate(2*time.Second, 30*time.Second, setup))
}

func (f *FrameWorkE2E) AfterEach() {
	destroy := func() (done bool, err error) {
		result, err := f.Client.
			CoreV1().
			Namespaces().
			Get(context.Background(), NameSpace, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			Logf("[namespace] %s deleted ", NameSpace)
			return true, nil
		}
		if err == nil {
			if result.Status.Phase == "Terminating" {
				Logf("[namespace] namespace still in [%s] state, %s", result.Status.Phase, time.Now())
				return false, nil
			}
			err := f.Client.
				CoreV1().
				Namespaces().
				Delete(context.Background(), NameSpace, metav1.DeleteOptions{})
			Logf("delete namespace, try again from error %v", err)
			return false, nil
		}
		Logf("delete namespace, poll error, namespace status unknown. %s", err.Error())
		return false, nil
	}
	ExpectNoError(wait.PollImmediate(6*time.Second, 2*time.Minute, destroy))
}

func RunActions(f *FrameWorkE2E, actions ...Action) error {
	Logf("RunActions (start)")
	Logf("===========================================================================")
	defer Logf("RunActions (finished)")

	for _, action := range actions {
		Logf("start to run test action: %s", reflect.TypeOf(action))
		if err := action.RunAction(f); err != nil {
			return fmt.Errorf("run action: %s", err.Error())
		}
		Logf("finished test action: %s\n\n\n", reflect.TypeOf(action))
	}
	return nil
}

func WaitServiceMutate(
	client kubernetes.Interface,
	svc *v1.Service,
	mutate ServiceMutator,
) (*v1.Service, error) {
	var newm *v1.Service
	err := wait.PollImmediate(
		3*time.Second,
		1*time.Minute,
		func() (done bool, err error) {
			m, err := CreateOrUpdate(client, svc, mutate)
			if err != nil {
				fmt.Printf("mutate service retry: %s", err.Error())
				return false, nil
			}
			newm = m
			return true, nil
		},
	)
	return newm, err
}

func CreateOrUpdate(
	client kubernetes.Interface,
	svc *v1.Service,
	mutate ServiceMutator,
) (*v1.Service, error) {
	o, err := client.CoreV1().Services(svc.Namespace).Get(context.Background(), svc.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			if mutate != nil {
				err := mutate(svc)
				if err != nil {
					return nil, fmt.Errorf("mutate: %s", err.Error())
				}
			}
			m, err := client.CoreV1().Services(svc.Namespace).Create(context.Background(), svc, metav1.CreateOptions{})
			if err != nil {
				return nil, fmt.Errorf("create service: NotFound %s", err.Error())
			}
			return m, nil
		} else {
			return nil, fmt.Errorf("create service: get %s", err.Error())
		}
	}
	od := o.DeepCopy()
	if mutate != nil {
		err := mutate(o)
		if err != nil {
			return nil, fmt.Errorf("mutate update: %s", err.Error())
		}
	}
	orig, err := json.Marshal(od)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %s", err.Error())
	}
	newm, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("json marshal: newobject %s", err.Error())
	}
	patch, err := strategicpatch.CreateTwoWayMergePatch(orig, newm, &v1.Service{})
	if err != nil {
		return nil, fmt.Errorf("create patch: %s", err.Error())
	}
	m, err := client.CoreV1().Services(svc.Namespace).Patch(context.Background(), svc.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return m, err
}

func ToPTR(a []v1.Node) []*v1.Node {
	var node []*v1.Node
	for i := range a {
		n := a[i]
		node = append(node, &n)
	}
	return node
}

func LoadConfig() (*restclient.Config, error) {
	c, err := clientcmd.LoadFromFile(TestConfig.KubeConfig)
	if err != nil {
		if TestConfig.KubeConfig == "" {
			return restclient.InClusterConfig()
		}
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(
		*c,
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: TestConfig.Host,
			},
		},
	).ClientConfig()
}

func NewClientOrDie() kubernetes.Interface {
	config, err := LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("error creating Client: %v", err.Error()))
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("new client : %s", err.Error()))
	}
	return client
}

func NewRuntimeClient() client.Client {
	config, err := LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("error creating Client: %v", err.Error()))
	}

	apiReader, err := client.New(config, client.Options{})
	if err != nil {
		panic(fmt.Sprintf("new runtime client error: %s", err.Error()))
	}
	return apiReader
}

func RunNginxDeployment(
	client kubernetes.Interface,
) error {
	var replica int32 = 3
	nginx := &v12.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: NameSpace,
			Labels: map[string]string{
				"run": "nginx",
			},
		},
		Spec: v12.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run": "nginx",
				},
			},
			Replicas: &replica,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"run": "nginx",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "nginx",
							Image:           "nginx:1.9.7",
							ImagePullPolicy: "Always",
						},
					},
				},
			},
		},
	}

	// TODO: create or update
	_, err := client.AppsV1().Deployments(nginx.Namespace).Create(context.Background(), nginx, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return fmt.Errorf("run nginx: %s", err.Error())
		}
		Logf("nginx already exist: %s", err.Error())
	}
	return wait.Poll(
		3*time.Second,
		1*time.Minute,
		func() (done bool, err error) {
			pods, err := client.CoreV1().Pods(nginx.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "run=nginx"})
			if err != nil {
				Logf("wait for nginx pod ready: %s", err.Error())
				return false, nil
			}
			if len(pods.Items) != int(*nginx.Spec.Replicas) {
				Logf("wait for nginx pod replicas: %d", len(pods.Items))
				return false, nil
			}
			for _, pod := range pods.Items {
				if pod.Status.Phase != "Running" {
					Logf("wait for nginx pod Running: %s", pod.Name)
					return false, nil
				}
			}
			return true, nil
		},
	)
}
func (f *FrameWorkE2E) EnsureDeleteSVC() {
	destroy := func() (done bool, err error) {
		result, err := f.Client.
			CoreV1().
			Namespaces().
			Get(context.Background(), NameSpace, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			Logf("[namespace] %s deleted ", NameSpace)
			return true, nil
		}
		if err == nil {
			if result.Status.Phase == "Terminating" {
				Logf("[namespace] namespace still in [%s] state, %s", result.Status.Phase, time.Now())
				return false, nil
			}
			err := f.Client.
				CoreV1().
				Namespaces().
				Delete(context.Background(), NameSpace, metav1.DeleteOptions{})
			Logf("delete namespace, try again from error %v", err)
			return false, nil
		}
		Logf("delete namespace, poll error, namespace status unknown. %s", err.Error())
		return false, nil
	}
	ExpectNoError(wait.PollImmediate(6*time.Second, 2*time.Minute, destroy))
}
