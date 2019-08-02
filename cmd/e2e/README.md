## How to Run E2ETest

Using the following command to start an E2E test
```$bash
 go test -v \
    k8s.io/cloud-provider-alibaba-cloud/cmd/e2e \
    -test.run ^TestE2E$ \
    --kubeconfig /path/to/.kube/config.e2e \
    --cloud-config /path/to/.kube/config.cloud \
    --lbid loadbalancer-id \
    --MasterZoneID cn-beijing-a \
    --SlaveZoneID cn-beijing-b \
    --BackendLabel failure-domain.beta.kubernetes.io/region=cn-beijing \
    --aclid acl-id \
    --vswitchid vswitch-id \
    --certid cert-id
```

- lbid: When reusing an exist LoadBalancer, you need to set this parameter. You can get LoadBalancer id at https://slbnew.console.aliyun.com/slb/.  
- MasterZoneID & Slave ZoneID: If you want to create LoadBalancer with specific master zone and slave zone, you need to set this parameter. You can get region/zone information at https://help.aliyun.com/document_detail/40654.html?spm=a2c4g.11186623.2.15.458f41deOaYQvu.  
- BackendLabel: If you want to create LoadBalancer with specific backend labels, you can use this parameter.  
- aclid: If you want to create LoadBalancer with access control list, you need to set this parameter. You can get ACL ID at https://slbnew.console.aliyun.com/slb/.  
- vswitchid: If you want to create LoadBalancer with specific vswitch, you need to set this parameter. You can get vswitch id at https://vpc.console.aliyun.com/.  
- certid: If you want to create the HTTP protocol LoadBalancer, you need to add the cert id. You can create a certification follow instructions at https://slbnew.console.aliyun.com/slb/.  
- 
File `config.e2e` is the kubeconfig file used to connect to the target kubernetes cluster.
For Example:
```$yaml
apiVersion: v1
clusters:
- cluster:
    server: https://${APISERVER_ADDR}:6443
    certificate-authority-data: 
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: "kubernetes-admin"
  name: kubernetes-admin-c62df1ad0683947aa9556a09f20493964
current-context: kubernetes-admin-c62df1ad0683947aa9556a09f20493964
kind: Config
preferences: {}
users:
- name: "kubernetes-admin"
  user:
    client-certificate-data: 
    client-key-data: 

```

File `config.cloud` is the cloudprovider cloud-config which mainly used in initializing CloudOpenAPI client.

```$xslt
{
  "Global":
  {
    "vpcid":"vpc-2zeph6817y1poszhs5p7u",
    "vswitchid":"vsw-2zeb5vsg77722l96hast5",
    "region":"cn-beijing",
    "zoneid":"cn-beijing-a",
    "accessKeyID":"base64 key",
    "accessKeySecret":"base64 secret"
  }
}

```

## Write E2ETest Case

Use framework.Mark to register an e2e test case. Framework will test it automatically.

```go
var _ = framework.Mark(
	func(t *testing.T) error {
		f := framework.NewFrameWork(
			func(f *framework.FrameWorkE2E) {
				// This is the test description
				f.Desribe = "TestRandomCombinationCase"
				// set testing.T
				f.Test = t
				// initialize cloud client
				f.Client = framework.NewClientOrDie()
				
				// framework provider an InitService for basic test.
				// set an init service here if needed. Uncomment the template
                // 
				//f.InitService = &v1.Service{
                //	ObjectMeta: metav1.ObjectMeta{
                //		Name:      "basic-service",
                //		Namespace: framework.NameSpace,
                //		Annotations: map[string]string{
                //			alicloud.ServiceAnnotationLoadBalancerId:               "lb-2ze6jp9vemd1gvj9ku83e",
                //			alicloud.ServiceAnnotationLoadBalancerOverrideListener: "true",
                //		},
                //	},
                //	Spec: v1.ServiceSpec{
                //		Ports: []v1.ServicePort{
                //			{
                //				Port:       80,
                //				TargetPort: intstr.FromInt(80),
                //				Protocol:   v1.ProtocolTCP,
                //			},
                //		},
                //		Type:            v1.ServiceTypeLoadBalancer,
                //		SessionAffinity: v1.ServiceAffinityNone,
                //		Selector: map[string]string{
                //			"run": "nginx",
                //		},
                //	},
                //}
			},
		)
		
		// set up namespace and nginx deployment
		err := f.SetUp()
		if err != nil {
			return fmt.Errorf("setup error: %s", err.Error())
		}
		// destroy namespace after test finished
		defer f.Destroy()
		
		// Here begins the real test action.
		// 1) Every framework.Action contains a TestUnit composed of `service Mutator` and `ExpectOK function` 
		// 2) Mutator mutate the previous service and trigger an reconcile operation.
		// 3) ExpectOK test the reconciliation is process as the expected way.
		// see the Example below
		// 1) framework.NewDefaultAction is a default service test action suit for most cases.
		// 2) framework.NewDeleteAction test for service deletion
		// 3) framework.NewRandomAction run the actions it contains randomly
		rand := []framework.Action{
			// set spec
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerSpec: "slb.s1.small",
						}
						return nil
					},
					Description: "mutate lb spec to slb.s2.small. function not ready yet",
				},
			),
			// set health check
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerHealthCheckInterval:       "10",
							alicloud.ServiceAnnotationLoadBalancerHealthCheckConnectTimeout: "3",
						}
						return nil
					},
					Description: "mutate health check",
				},
			),
			framework.NewDefaultAction(
				&framework.TestUnit{
					Mutator: func(service *v1.Service) error {
						service.Annotations = map[string]string{
							alicloud.ServiceAnnotationLoadBalancerProtocolPort: "http:80",
						}
						return nil
					},
					Description: "mutate lb 80 listener to http",
				},
			),
		}
		return f.RunDefaultTest(
			framework.NewDefaultAction(&framework.TestUnit{Description: "default init"}),
			framework.NewRandomAction(rand),
			framework.NewDeleteAction(&framework.TestUnit{Description: "default delete"}),
		)
	},
)

```