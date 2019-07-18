package framework

import (
	"encoding/json"
	"fmt"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	cloud "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
	"reflect"
	"strings"
	"testing"
	"time"
)

const NameSpace = "e2etest"

var Frames []EntryPoint

type EntryPoint func(t *testing.T) error

func Mark(
	f EntryPoint,
) error {
	Frames = append(Frames, f)
	return nil
}

type OptionsFunc func(f *FrameWorkE2E)

//FrameWorkE2E e2e framework
type FrameWorkE2E struct {
	InitService *v1.Service
	Test        *testing.T
	Desribe     string
	Log         Logger
	Client      kubernetes.Interface
}

type Logger struct {
}

func (m *FrameWorkE2E) Logf(format string, args ...interface{}) {
	m.Test.Logf("[%s] %s", m.Desribe, fmt.Sprintf(format, args...))
}

var TestContext Config

type Config struct {
	Host           string
	CloudConfig    string
	KubeConfig     string
	LoadBalancerID string
}

func NewFrameWork(
	option OptionsFunc,
) *FrameWorkE2E {
	frame := &FrameWorkE2E{}
	option(frame)
	if frame.InitService == nil {
		frame.InitService = &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "basic-service",
				Namespace: NameSpace,
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
				Selector: map[string]string{
					"run": "nginx",
				},
			},
		}
	}
	return frame
}

type ServiceMutator func(service *v1.Service) error

type TestUnit struct {
	Description string
	Mutator     ServiceMutator
	//Service  *v1.Service
	ExpectOK func(f *cloud.FrameWork) error
}

func (f *FrameWorkE2E) SetUp() error {

	return wait.PollImmediate(
		2*time.Second,
		30*time.Second,
		func() (done bool, err error) {
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
				Create(ns)
			if err != nil {
				if strings.Contains(err.Error(), "exist") {
					return true, RunNginxDeployment(f.Test, f.Client)
				}
				f.Logf("initialize namespace: %s", err.Error())
				return false, err
			}
			return false, nil
		},
	)
}

func (f *FrameWorkE2E) Destroy() error {
	return wait.PollImmediate(
		6*time.Second,
		30*time.Second,
		func() (done bool, err error) {
			result, err := f.Client.
				CoreV1().
				Namespaces().
				Get(NameSpace, metav1.GetOptions{})
			if err != nil && strings.Contains(err.Error(), "not found") {
				f.Logf("[namespace] %s deleted ", NameSpace)
				return true, nil
			}
			if err == nil {
				if result.Status.Phase == "Terminating" {
					f.Logf("[namespace] namespace still in [%s] state, %s", result.Status.Phase, time.Now())
					return false, nil
				}
				err := f.Client.
					CoreV1().
					Namespaces().
					Delete(NameSpace, &metav1.DeleteOptions{})
				f.Logf("delete namespace, try again from error %v", err)
				return false, nil
			}
			f.Logf("delete namespace, poll error, namespace status unknown. %s", err.Error())
			return false, nil
		},
	)
}

func (f *FrameWorkE2E) RunDefaultTest(actions ...Action) error {
	f.Logf("RunDefaultTest (start)")
	f.Logf("===========================================================================")
	defer f.Logf("RunDefaultTest (finished)")

	for _, action := range actions {
		f.Logf("start to run test action: %s", reflect.TypeOf(action))
		if err := action.RunAction(f); err != nil {
			return fmt.Errorf("run action: %s", err.Error())
		}
		f.Logf("finished test action: %s\n\n\n", reflect.TypeOf(action))
	}
	return nil
}

func WaitServiceMutate(
	client kubernetes.Interface,
	svc *v1.Service,
	mutate ServiceMutator,
) (*v1.Service, error) {
	var newm *v1.Service
	return newm, wait.PollImmediate(
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
}

func CreateOrUpdate(
	client kubernetes.Interface,
	svc *v1.Service,
	mutate ServiceMutator,
) (*v1.Service, error) {
	o, err := client.CoreV1().Services(svc.Namespace).Get(svc.Name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if mutate != nil {
				err := mutate(svc)
				if err != nil {
					return nil, fmt.Errorf("mutate: %s", err.Error())
				}
			}
			m, err := client.CoreV1().Services(svc.Namespace).Create(svc)
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
	m, err := client.CoreV1().Services(svc.Namespace).Patch(svc.Name, types.MergePatchType, patch)
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
	c, err := clientcmd.LoadFromFile(TestContext.KubeConfig)
	if err != nil {
		if TestContext.KubeConfig == "" {
			return restclient.InClusterConfig()
		}
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(
		*c,
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: TestContext.Host,
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

func RunNginxDeployment(
	t *testing.T,
	client kubernetes.Interface,
) error {
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

	_, err := client.AppsV1().Deployments(nginx.Namespace).Create(nginx)
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return fmt.Errorf("run nginx: %s", err.Error())
		}
		t.Logf("nginx already exist: %s", err.Error())
	}
	return nil
}
