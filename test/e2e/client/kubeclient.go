package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/klog/v2"
)

var (
	Namespace           = "e2e-test"
	Service             = "basic-service"
	Deployment          = "nginx"
	VKDeployment        = "nginx-vk"
	NodeLabel           = "e2etest"
	ExcludeNodeLabel    = "service.beta.kubernetes.io/exclude-node"
	FixtureReadyTimeout = 10 * time.Minute
	namespaceOwner      = string(uuid.NewUUID())
)

// ConfigureTestResources isolates namespaced fixtures and worker-owned node
// labels.  Ginkgo parallel workers are separate processes, so these package
// values remain process-local and do not introduce data races.
func ConfigureTestResources(workerScope string, fixtureReadyTimeout time.Duration) {
	Namespace = "e2e-test"
	Service = "basic-service"
	Deployment = "nginx"
	VKDeployment = "nginx-vk"
	NodeLabel = "e2etest"
	FixtureReadyTimeout = fixtureReadyTimeout
	if workerScope == "" {
		return
	}

	Namespace = workerScope
	NodeLabel = workerScope
}

type KubeClient struct {
	kubernetes.Interface
}

func NewKubeClient(client kubernetes.Interface) *KubeClient {
	return &KubeClient{client}
}

// service
func (client *KubeClient) DefaultService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Service,
			Namespace: Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(443),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
			Selector:        map[string]string{"run": "nginx"},
		},
	}
}

func (client *KubeClient) DefaultNLBService() *v1.Service {
	ipFamilyPolicy := v1.IPFamilyPolicyPreferDualStack
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Service,
			Namespace: Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(443),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:              v1.ServiceTypeLoadBalancer,
			SessionAffinity:   v1.ServiceAffinityNone,
			Selector:          map[string]string{"run": "nginx"},
			LoadBalancerClass: tea.String(helper.NLBClass),
			IPFamilyPolicy:    &ipFamilyPolicy,
		},
	}
}

func (client *KubeClient) CreateServiceByAnno(anno map[string]string) (*v1.Service, error) {
	svc := client.DefaultService()
	svc.Annotations = anno
	return client.CreateService(svc)
}

func (client *KubeClient) CreateNLBServiceByAnno(anno map[string]string) (*v1.Service, error) {
	svc := client.DefaultNLBService()
	svc.Annotations = anno

	return client.CreateService(svc)
}

func (client *KubeClient) CreateServiceWithStringTargetPort(anno map[string]string) (*v1.Service, error) {
	svc := client.DefaultService()
	svc.Annotations = anno
	svc.Spec.Ports = []v1.ServicePort{
		{
			Name:       "http",
			Port:       80,
			TargetPort: intstr.FromString("http"),
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "https",
			Port:       443,
			TargetPort: intstr.FromString("https"),
			Protocol:   v1.ProtocolTCP,
		},
	}
	return client.CreateService(svc)
}

func (client *KubeClient) CreateNLBServiceWithStringTargetPort(anno map[string]string) (*v1.Service, error) {
	lbClass := helper.NLBClass
	svc := client.DefaultService()
	svc.Annotations = anno
	svc.Spec.LoadBalancerClass = &lbClass
	svc.Spec.Ports = []v1.ServicePort{
		{
			Name:       "http",
			Port:       80,
			TargetPort: intstr.FromString("http"),
			Protocol:   v1.ProtocolTCP,
		},
		{
			Name:       "https",
			Port:       443,
			TargetPort: intstr.FromString("https"),
			Protocol:   v1.ProtocolTCP,
		},
	}
	return client.CreateService(svc)
}

func (client *KubeClient) CreateService(svc *v1.Service) (*v1.Service, error) {
	if svc == nil {
		return nil, fmt.Errorf("svc is nil")
	}
	defaultTestServiceToIntranet(svc)
	return client.CoreV1().Services(Namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
}

// defaultTestServiceToIntranet keeps ordinary E2E cases off public load
// balancers. Cases that explicitly exercise public behavior set AddressType to
// Internet and are left unchanged.
func defaultTestServiceToIntranet(svc *v1.Service) {
	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}
	key := annotation.Annotation(annotation.AddressType)
	if _, ok := svc.Annotations[key]; !ok {
		svc.Annotations[key] = string(model.IntranetAddressType)
	}
}

func (client *KubeClient) PatchService(oldSvc, newSvc *v1.Service) (*v1.Service, error) {
	if newSvc == nil {
		return nil, fmt.Errorf("new svc is nil")
	}
	preserveOrDefaultTestServiceAddressType(oldSvc, newSvc)
	oldStr, _ := json.Marshal(oldSvc)
	newStr, _ := json.Marshal(newSvc)
	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldStr, newStr, &v1.Service{})
	if patchErr != nil {
		return nil, fmt.Errorf("create merge patch: %s", patchErr.Error())
	}
	return client.CoreV1().Services(Namespace).Patch(context.TODO(), Service, types.StrategicMergePatchType,
		patchBytes, metav1.PatchOptions{})
}

func preserveOrDefaultTestServiceAddressType(oldSvc, newSvc *v1.Service) {
	key := annotation.Annotation(annotation.AddressType)
	if _, ok := newSvc.Annotations[key]; ok {
		return
	}
	if oldSvc != nil && oldSvc.Annotations != nil {
		if addressType, ok := oldSvc.Annotations[key]; ok {
			if newSvc.Annotations == nil {
				newSvc.Annotations = make(map[string]string)
			}
			newSvc.Annotations[key] = addressType
			return
		}
	}
	defaultTestServiceToIntranet(newSvc)
}

func (client *KubeClient) CreateServiceWithoutSelector(anno map[string]string) (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        Service,
			Namespace:   Namespace,
			Annotations: anno,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}

	return client.CreateService(svc)
}

func (client *KubeClient) CreateNLBServiceWithoutSelector(anno map[string]string) (*v1.Service, error) {
	lbClass := helper.NLBClass
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        Service,
			Namespace:   Namespace,
			Annotations: anno,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:              v1.ServiceTypeLoadBalancer,
			LoadBalancerClass: &lbClass,
		},
	}

	return client.CreateService(svc)
}

func (client *KubeClient) DeleteService() error {
	return wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
		err = client.CoreV1().Services(Namespace).Delete(context.TODO(), Service, metav1.DeleteOptions{})
		if err == nil || apierrors.IsNotFound(err) {
			return true, nil
		}
		if isRetryableKubeAPIError(err) {
			return false, nil
		}
		return false, err
	})
}

func (client *KubeClient) DeleteServiceByName(name string) error {
	return wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
		err = client.CoreV1().Services(Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err == nil || apierrors.IsNotFound(err) {
			return true, nil
		}
		if isRetryableKubeAPIError(err) {
			return false, nil
		}
		return false, err
	})
}

func (client *KubeClient) DeleteEndpoints() error {
	err := client.CoreV1().Endpoints(Namespace).Delete(context.TODO(), Service, metav1.DeleteOptions{})
	if err == nil || apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (client *KubeClient) GetService() (*v1.Service, error) {
	return client.CoreV1().Services(Namespace).Get(context.TODO(), Service, metav1.GetOptions{})
}

// endpoints

func (client *KubeClient) GetEndpoint() (*v1.Endpoints, error) {
	return client.CoreV1().Endpoints(Namespace).Get(context.TODO(), Service, metav1.GetOptions{})
}

func (client *KubeClient) GetEndpointSlices() ([]discovery.EndpointSlice, error) {
	list, err := client.DiscoveryV1().EndpointSlices(Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubernetes.io/service-name=" + Service,
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (client *KubeClient) CreateEndpointsWithoutNodeName() (*v1.Endpoints, error) {
	ep := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Service,
			Namespace: Namespace,
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "123.123.123.123",
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port:     80,
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	}
	return client.CoreV1().Endpoints(Namespace).Create(context.TODO(), ep, metav1.CreateOptions{})
}

func (client *KubeClient) CreateEndpointsWithNotExistNode() (*v1.Endpoints, error) {
	nodeName := "cn-hangzhou.123.123.123.123"
	ep := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Service,
			Namespace: Namespace,
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "123.123.123.123",
						NodeName: &nodeName,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port:     80,
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	}
	return client.CoreV1().Endpoints(Namespace).Create(context.TODO(), ep, metav1.CreateOptions{})
}

func (client *KubeClient) CreateEndpointsWithIPs(svc *v1.Service, ips []string) (*v1.Endpoints, error) {
	subset := v1.EndpointSubset{}
	for _, p := range svc.Spec.Ports {
		subset.Ports = append(subset.Ports, v1.EndpointPort{
			Name:     p.Name,
			Port:     p.Port,
			Protocol: p.Protocol,
		})
	}
	for _, i := range ips {
		subset.Addresses = append(subset.Addresses, v1.EndpointAddress{
			IP: i,
		})
	}

	ep := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
		Subsets: []v1.EndpointSubset{subset},
	}

	return client.CoreV1().Endpoints(svc.Namespace).Create(context.TODO(), ep, metav1.CreateOptions{})
}

// deployment
func (client *KubeClient) CreateDeployment() error {
	var replica int32 = 3
	nginx := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Deployment,
			Namespace: Namespace,
			Labels: map[string]string{
				"run": "nginx",
			},
		},
		Spec: appv1.DeploymentSpec{
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
							Image:           "registry.cn-hangzhou.aliyuncs.com/acs-sample/nginx:latest",
							ImagePullPolicy: "Always",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      v1.ProtocolTCP,
								},
								{
									Name:          "https",
									ContainerPort: 443,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := client.AppsV1().Deployments(Namespace).Create(context.Background(), nginx, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return fmt.Errorf("create nginx error: %s", err.Error())
		}
	}
	return wait.Poll(5*time.Second, FixtureReadyTimeout, func() (done bool, err error) {
		pods, err := client.CoreV1().Pods(nginx.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "run=nginx"})
		if err != nil {
			klog.Infof("wait for nginx pod ready: %s", err.Error())
			return false, nil
		}
		if len(pods.Items) != int(*nginx.Spec.Replicas) {
			klog.Infof("wait for nginx pod replicas: %d", len(pods.Items))
			return false, nil
		}
		for i := range pods.Items {
			if !isPodReady(&pods.Items[i]) {
				klog.Infof("wait for nginx pod Ready: %s", pods.Items[i].Name)
				return false, nil
			}
		}
		return true, nil
	},
	)
}

func (client *KubeClient) CreateSecondaryDeployment() error {
	var replica int32 = 3
	nginx := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-2", Deployment),
			Namespace: Namespace,
			Labels: map[string]string{
				"run": "nginx",
			},
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run": "nginx",
					"dep": "secondary",
				},
			},
			Replicas: &replica,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"run": "nginx",
						"dep": "secondary",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "nginx",
							Image:           "registry.cn-hangzhou.aliyuncs.com/acs-sample/nginx:latest",
							ImagePullPolicy: "Always",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      v1.ProtocolTCP,
								},
								{
									Name:          "https-1",
									ContainerPort: 443,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := client.AppsV1().Deployments(Namespace).Create(context.Background(), nginx, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return fmt.Errorf("create nginx error: %s", err.Error())
		}
	}
	return wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		pods, err := client.CoreV1().Pods(nginx.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "dep=secondary"})
		if err != nil {
			klog.Infof("wait for nginx pod ready: %s", err.Error())
			return false, nil
		}
		if len(pods.Items) != int(*nginx.Spec.Replicas) {
			klog.Infof("wait for nginx pod replicas: %d", len(pods.Items))
			return false, nil
		}
		for i := range pods.Items {
			if !isPodReady(&pods.Items[i]) {
				klog.Infof("wait for nginx pod Ready: %s", pods.Items[i].Name)
				return false, nil
			}
		}
		return true, nil
	},
	)
}

func (client *KubeClient) DeleteSecondaryDeployment() error {
	name := fmt.Sprintf("%s-2", Deployment)
	err := client.AppsV1().Deployments(Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err == nil || apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (client *KubeClient) ScaleDeployment(replica int32) error {
	deploy, err := client.AppsV1().Deployments(Namespace).Get(context.TODO(), Deployment, metav1.GetOptions{})
	if err != nil {
		return err
	}
	deploy.Spec.Replicas = &replica
	_, err = client.AppsV1().Deployments(Namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return wait.PollImmediate(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		pods, err := client.CoreV1().Pods(Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "run=nginx"})
		if err != nil {
			klog.Infof("wait for nginx pod ready: %s", err.Error())
			return false, nil
		}
		if len(pods.Items) != int(replica) {
			klog.Infof("wait for nginx pod replicas: %d", len(pods.Items))
			return false, nil
		}
		for i := range pods.Items {
			if !isPodReady(&pods.Items[i]) {
				klog.Infof("wait for nginx pod Ready: %s", pods.Items[i].Name)
				return false, nil
			}
		}
		return true, nil
	})
}

func (client *KubeClient) CreateVKDeployment() error {
	var replica int32 = 2
	nginx := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VKDeployment,
			Namespace: Namespace,
			Labels: map[string]string{
				"run": "nginx",
				"app": "nginx-vk",
			},
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run": "nginx",
					"app": "nginx-vk",
				},
			},
			Replicas: &replica,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"run": "nginx",
						"app": "nginx-vk",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "nginx",
							Image:           "nginx:1.9.7",
							ImagePullPolicy: "Always",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      v1.ProtocolTCP,
								},
								{
									Name:          "https",
									ContainerPort: 443,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
					},
					NodeSelector: map[string]string{
						"type": "virtual-kubelet",
					},
					Tolerations: []v1.Toleration{
						{
							Operator: v1.TolerationOpExists,
						},
					},
				},
			},
		},
	}

	_, err := client.AppsV1().Deployments(Namespace).Create(context.Background(), nginx, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return fmt.Errorf("create nginx error: %s", err.Error())
		}
	}
	return wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		pods, err := client.CoreV1().Pods(nginx.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=nginx-vk"})
		if err != nil {
			klog.Infof("wait for nginx pod ready: %s", err.Error())
			return false, nil
		}
		if len(pods.Items) != int(*nginx.Spec.Replicas) {
			klog.Infof("wait for nginx pod replicas: %d", len(pods.Items))
			return false, nil
		}
		for i := range pods.Items {
			if !isPodReady(&pods.Items[i]) {
				klog.Infof("wait for nginx pod Ready: %s", pods.Items[i].Name)
				return false, nil
			}
		}
		return true, nil
	},
	)

}

// namespace
func (client *KubeClient) CreateNamespace() error {
	const ownerLabel = "e2e.cloud-provider-alibaba/owner"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Namespace,
			Namespace: Namespace,
			Labels:    map[string]string{ownerLabel: namespaceOwner},
		},
	}
	var lastErr error
	err := wait.PollImmediate(2*time.Second, 3*time.Minute, func() (bool, error) {
		_, lastErr = client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
		if lastErr == nil {
			return true, nil
		}
		if apierrors.IsAlreadyExists(lastErr) {
			existing, getErr := client.CoreV1().Namespaces().Get(context.TODO(), Namespace, metav1.GetOptions{})
			if getErr != nil {
				lastErr = getErr
				if isRetryableKubeAPIError(getErr) {
					return false, nil
				}
				return false, getErr
			}
			if existing.Labels[ownerLabel] == namespaceOwner {
				return true, nil
			}
			return false, lastErr
		}
		if isRetryableKubeAPIError(lastErr) {
			klog.Warningf("retry creating test namespace %s: %s", Namespace, lastErr)
			return false, nil
		}
		return false, lastErr
	})
	if err != nil && lastErr != nil {
		return lastErr
	}
	return err
}

func isRetryableKubeAPIError(err error) bool {
	return apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) ||
		apierrors.IsServiceUnavailable(err) || apierrors.IsTooManyRequests(err) ||
		utilnet.IsTimeout(err) || utilnet.IsConnectionReset(err) || utilnet.IsConnectionRefused(err)
}

func (client *KubeClient) DeleteNamespace() error {
	return wait.PollImmediate(5*time.Second, 3*time.Minute,
		func() (done bool, err error) {
			err = client.CoreV1().Namespaces().Delete(context.TODO(), Namespace, metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			if err == nil {
				return false, nil
			}
			if isRetryableKubeAPIError(err) {
				return false, nil
			}
			return false, err
		})
}

// node
func (client *KubeClient) LabelNode(nodeName string, key string, value string) error {
	return wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		n, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil || n == nil {
			return false, nil
		}
		if n.Labels == nil {
			n.Labels = make(map[string]string)
		}
		n.ObjectMeta.Labels[key] = value
		_, err = client.CoreV1().Nodes().Update(context.TODO(), n, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil
	})
}

func (client *KubeClient) UnLabelNode(nodeName string, key string) error {
	return wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		n, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil || n == nil {
			return false, nil
		}
		delete(n.ObjectMeta.Labels, key)
		_, err = client.CoreV1().Nodes().Update(context.TODO(), n, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil
	})
}

func (client *KubeClient) UnscheduledNode(nodeName string) error {
	return wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		n, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil || n == nil {
			return false, nil
		}
		n.Spec.Unschedulable = true
		_, err = client.CoreV1().Nodes().Update(context.TODO(), n, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil
	})

}

func (client *KubeClient) ScheduledNode(nodeName string) error {
	return wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		n, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil || n == nil {
			return false, nil
		}
		n.Spec.Unschedulable = false
		_, err = client.CoreV1().Nodes().Update(context.TODO(), n, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil
	})

}

func (client *KubeClient) AddTaint(nodeName string, taint v1.Taint) (bool, error) {
	added := false
	err := wait.PollImmediate(2*time.Second, 30*time.Second, func() (done bool, err error) {
		n, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		for _, existing := range n.Spec.Taints {
			if existing.MatchTaint(&taint) {
				if added {
					added = existing.Value == taint.Value
				}
				return true, nil
			}
		}
		n.Spec.Taints = append(n.Spec.Taints, taint)
		// Claim ownership before Update: the API server can accept the write
		// even when the client loses its response. Exact-value removal below
		// makes cleanup harmless if the write was not persisted.
		added = true
		_, err = client.CoreV1().Nodes().Update(context.TODO(), n, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	return added, err
}

func (client *KubeClient) RemoveTaint(nodeName string, taint v1.Taint) error {
	return wait.PollImmediate(2*time.Second, 30*time.Second, func() (done bool, err error) {
		n, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		var updateTaints []v1.Taint
		found := false
		for _, t := range n.Spec.Taints {
			if t.MatchTaint(&taint) && t.Value == taint.Value {
				found = true
				continue
			}
			updateTaints = append(updateTaints, t)
		}
		if !found {
			return true, nil
		}
		n.Spec.Taints = updateTaints
		_, err = client.CoreV1().Nodes().Update(context.TODO(), n, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil
	})
}

func (client *KubeClient) ListNodes() ([]v1.Node, error) {
	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

func (client *KubeClient) GetLatestNode() (*v1.Node, error) {
	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(nodeList.Items) == 0 {
		return nil, nil
	}

	var ret v1.Node
	for _, node := range nodeList.Items {
		if helper.IsExcludedNode(&node) {
			continue
		}
		if _, exclude := node.Labels[helper.LabelNodeExcludeBalancer]; exclude {
			continue
		}
		if _, exclude := node.Labels[helper.LabelNodeExcludeBalancerDeprecated]; exclude {
			continue
		}
		if _, isVK := node.Labels[helper.LabelNodeTypeVK]; isVK {
			continue
		}
		if ret.Name == "" {
			ret = node
		} else if ret.CreationTimestamp.Before(&node.CreationTimestamp) {
			ret = node
		}
	}
	if ret.Name == "" {
		return nil, nil
	}
	klog.Infof("return node:%s", ret.Name)
	return &ret, nil
}

func (client *KubeClient) PatchNodeStatus(oldNode, newNode *v1.Node) (*v1.Node, error) {
	oldStr, _ := json.Marshal(oldNode)
	newStr, _ := json.Marshal(newNode)
	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldStr, newStr, &v1.Node{})
	if patchErr != nil {
		return nil, fmt.Errorf("create merge patch: %s", patchErr.Error())
	}
	return client.CoreV1().Nodes().PatchStatus(context.TODO(), oldNode.Name, patchBytes)
}

func (client *KubeClient) PatchNode(oldNode, newNode *v1.Node) (*v1.Node, error) {
	oldStr, _ := json.Marshal(oldNode)
	newStr, _ := json.Marshal(newNode)
	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldStr, newStr, &v1.Node{})
	if patchErr != nil {
		return nil, fmt.Errorf("create merge patch: %s", patchErr.Error())
	}
	return client.CoreV1().Nodes().Patch(context.TODO(), oldNode.Name,
		types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
}

func (client *KubeClient) GetDeploymentPods() ([]v1.Pod, error) {
	podList, err := client.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "run=nginx",
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (client *KubeClient) DeletePod(name string) error {
	return client.CoreV1().Pods(Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (client *KubeClient) ForceDeletePod(name string) error {
	gracePeriod := int64(0)
	return client.CoreV1().Pods(Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})
}

func (client *KubeClient) CreateGracefulShutdownDeployment() error {
	var replica int32 = 3
	var gracePeriod int64 = 60
	name := "nginx-graceful"
	deploy := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: Namespace,
			Labels:    map[string]string{"run": name},
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"run": name},
			},
			Replicas: &replica,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"run": name},
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &gracePeriod,
					Containers: []v1.Container{
						{
							Name:            "nginx",
							Image:           "registry.cn-hangzhou.aliyuncs.com/acs-sample/nginx:latest",
							ImagePullPolicy: "Always",
							Ports: []v1.ContainerPort{
								{Name: "http", ContainerPort: 80, Protocol: v1.ProtocolTCP},
								{Name: "https", ContainerPort: 443, Protocol: v1.ProtocolTCP},
							},
							Lifecycle: &v1.Lifecycle{
								PreStop: &v1.LifecycleHandler{
									Exec: &v1.ExecAction{Command: []string{"sleep", "30"}},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := client.AppsV1().Deployments(Namespace).Create(context.Background(), deploy, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create %s error: %s", name, err.Error())
		}
		if err := client.DeleteGracefulShutdownDeployment(); err != nil {
			return fmt.Errorf("delete existing %s error: %s", name, err.Error())
		}
		_, err = client.AppsV1().Deployments(Namespace).Create(context.Background(), deploy, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("create %s error: %s", name, err.Error())
		}
	}
	return wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		pods, err := client.CoreV1().Pods(Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "run=" + name})
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != int(replica) {
			return false, nil
		}
		for _, pod := range pods.Items {
			if !isPodReady(&pod) {
				return false, nil
			}
		}
		return true, nil
	})
}

func (client *KubeClient) DeleteGracefulShutdownDeployment() error {
	name := "nginx-graceful"
	propagation := metav1.DeletePropagationForeground
	err := client.AppsV1().Deployments(Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	// Start foreground owner deletion first so it cannot replace Pods. Forcing
	// the deliberately slow Pods down then lets garbage collection finish fast
	// without leaving an old ReplicaSet that can race a same-name recreation.
	pods, listErr := client.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "run=" + name,
	})
	if listErr != nil {
		return listErr
	}
	for i := range pods.Items {
		if err := client.ForceDeletePod(pods.Items[i].Name); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return wait.PollImmediate(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		_, err = client.AppsV1().Deployments(Namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}

func (client *KubeClient) GetGracefulShutdownPods() ([]v1.Pod, error) {
	podList, err := client.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "run=nginx-graceful",
	})
	if err != nil {
		return nil, err
	}
	readyPods := make([]v1.Pod, 0, len(podList.Items))
	for i := range podList.Items {
		if podList.Items[i].DeletionTimestamp == nil && isPodReady(&podList.Items[i]) {
			readyPods = append(readyPods, podList.Items[i])
		}
	}
	return readyPods, nil
}

func isPodReady(pod *v1.Pod) bool {
	if pod == nil || pod.Status.Phase != v1.PodRunning {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
}

func (client *KubeClient) WaitForTerminatingEndpoint(serviceName, podName, podIP string) error {
	return wait.PollImmediate(time.Second, 30*time.Second, func() (bool, error) {
		slices, err := client.DiscoveryV1().EndpointSlices(Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: discovery.LabelServiceName + "=" + serviceName,
		})
		if err != nil {
			return false, nil
		}
		for i := range slices.Items {
			for j := range slices.Items[i].Endpoints {
				ep := &slices.Items[i].Endpoints[j]
				if ep.TargetRef == nil || ep.TargetRef.Name != podName || !containsString(ep.Addresses, podIP) {
					continue
				}
				return ep.Conditions.Terminating != nil && *ep.Conditions.Terminating &&
					ep.Conditions.Serving != nil && *ep.Conditions.Serving, nil
			}
		}
		return false, nil
	})
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func (client *KubeClient) ListServiceEvents(svc *v1.Service, reason string) ([]v1.Event, error) {
	selector := fmt.Sprintf("involvedObject.name=%s", svc.Name)
	if reason != "" {
		selector += fmt.Sprintf(",reason=%s", reason)
	}
	eventList, err := client.CoreV1().Events(Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	return eventList.Items, nil
}
