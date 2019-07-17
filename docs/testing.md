# Testing

### UnitTest

Alibaba Cloud Controller use mocked cloud SDK to implement code unit test.

```go
// Test Http configuration.
func TestEnsureLoadBalancerVswitchID(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "https-service",
				UID:  types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerVswitch:     VSWITCH_ID,
					ServiceAnnotationLoadBalancerAddressType: string(slb.IntranetAddressType),
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
		// initial node based on your definition.
		// backend of the created loadbalaner
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
		},
		nil,
		nil,
	)

	f.RunDefault(t, "Create Loadbalancer With VswitchID")
}
```

Here is an example of self defined test point.
```go
func TestEnsureLoadbalancerDeleted(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "https-service",
				UID:         types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
		// initial node based on your definition.
		// backend of the created loadbalaner
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
		},
		nil,
		nil,
	)

	f.Run(
		t,
		"Delete Loadbalancer", "ecs",
		func() {
			_, err := f.Cloud.EnsureLoadBalancer(CLUSTER_ID, f.SVC, f.Nodes)
			if err != nil {
				t.Fatalf("delete loadbalancer error: create %s", err.Error())
			}
			err = f.Cloud.EnsureLoadBalancerDeleted(CLUSTER_ID, f.SVC)
			if err != nil {
				t.Fatalf("ensure loadbalancer delete error, %s", err.Error())
			}
			exist, _, err := f.LoadBalancer().findLoadBalancer(f.SVC)
			if err != nil || exist {
				t.Fatalf("Delete LoadBalancer error: %v, %t", err, exist)
			}
		},
	)
}
```

Faked SDK made unit test easier.

Use ```make test``` to run unit test.

### Integration Test

Alibaba Cloud Controller Manager integration test is expected to follow the kubernetes testgrid rule addressed in https://github.com/kubernetes/community/pull/2224#issuecomment-395410751 for consistency.

For detailed e2etest see `cmd/e2e/README.md`, [Testing E2E](https://github.com/kubernetes/cloud-provider-alibaba-cloud/tree/master/cmd/e2e/README.md)

