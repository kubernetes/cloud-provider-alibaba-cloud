/*
Copyright 2018 The Kubernetes Authors.

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

package alicloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/metadata"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	climgr, err := NewMockClientMgr(&mockClientSLB{})
	if climgr == nil || err != nil {
		t.Logf("create climgr error!")
		t.Fail()
	}
	//realSlbClient(keyid,keysecret)
}

func NewMockClientMgr(client ClientSLBSDK) (*ClientMgr, error) {
	mgr := &ClientMgr{
		stop: make(<-chan struct{}, 1),
		meta: metadata.NewMockMetaData(
			nil,
			func(resource string) (string, error) {
				if strings.Contains(resource, metadata.REGION) {
					return string(REGION), nil
				}
				return "", errors.New("not found")
			},
		),
		loadbalancer: &LoadBalancerClient{c: client},
	}
	return mgr, nil
}

func TestFindLoadBalancer(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   "default",
				Name:        "service-test",
				UID:         types.UID(serviceUIDExist),
				Annotations: map[string]string{},
			},
			Spec: v1.ServiceSpec{
				Type: "LoadBalancer",
			},
		},
		// initial node based on your definition.
		// backend of the created loadbalancer
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
		"Create Loadbalancer With SPEC", "ecs",
		func() {

			// ==================================================================================
			// 1. No LOADBALANCER_ID specified, Exist.
			//    user need to create new loadbalancer. did not specify any exist loadbalancer.
			//	  Expected fallback to use service UID to generate slb .
			t.Logf("1. findLoadBalancer: No LOADBALANCER_ID specified, Expect:Exist")
			exist, lb, err := f.LoadBalancer().findLoadBalancer(f.SVC)
			if err != nil || !exist {
				t.Fatal("Test findLoadBalancer fail. user need to create new loadbalancer. did not specify any exist loadbalancer.")
			}
			if lb.LoadBalancerName != cloudprovider.GetLoadBalancerName(f.SVC) {
				t.Fatal("find loadbalancer fail. suppose to find by name.")
			}

			// ==================================================================================
			// 2. No LOADBALANCER_ID specified, None-Exist.
			//    user need to create new loadbalancer. did not specify any exist loadbalancer.
			//	  Expected fallback to use service UID to generate slb .
			t.Logf("2. findLoadBalancer: No LOADBALANCER_ID specified, Expect:Non-Exist")
			f.SVC.UID = types.UID("xxxxxxxxxxxxxxxxxxx-new")
			exist, _, err = f.LoadBalancer().findLoadBalancer(f.SVC)
			if err != nil {
				t.Fatal("Test findLoadBalancer fail.", err.Error())
			}
			if exist {
				t.Fatal(fmt.Sprintf("loadbalancer should not exist, %s", f.SVC.UID))
			}

			// ==================================================================================
			// 3. Loadbalancer id was specified, Expect: Exist
			// user need to use an exist loadbalancer through annotations
			t.Logf("3. findLoadBalancer: Loadbalancer id was specified, Expect:Exist")
			f.SVC.Annotations[ServiceAnnotationLoadBalancerId] = LOADBALANCER_ID
			exist, lb, err = f.LoadBalancer().findLoadBalancer(f.SVC)
			if err != nil || !exist {
				t.Fatal("3. findLoadBalancer: Loadbalancer id was specified, Expect:Exist")
			}
			if lb.LoadBalancerId != LOADBALANCER_ID {
				t.Fatal("find loadbalancer fail. suppose to find by exist loadbalancerid.")
			}

			// ==================================================================================
			// 4. Loadbalancer id was specified, Expect: NonExist
			// user need to use an exist loadbalancer through annotations
			t.Logf("4. findLoadBalancer: Loadbalancer id was specified, Expect:NonExist")
			f.SVC.Annotations[ServiceAnnotationLoadBalancerId] = LOADBALANCER_ID + "-new"
			exist, lb, err = f.LoadBalancer().findLoadBalancer(f.SVC)
			if err != nil {
				t.Logf("4. error: %s", err.Error())
				t.Fail()
			}
			if exist {
				t.Logf("4. findLoadBalancer: Loadbalancer id was specified, Expect:NonExist")
				t.Logf("   user need to use an exist loadbalancer through annotations")
				t.Fail()
			}
		},
	)
}

func realSlbClient(keyid, keysec string) {

	slbclient := slb.NewClient(keyid, keysec)
	slbclient.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	lb, err := slbclient.DescribeLoadBalancers(&slb.DescribeLoadBalancersArgs{
		RegionId:       common.Hangzhou,
		LoadBalancerId: "lb-bp1ids9hmq5924m6uk5w1",
	})
	if err == nil {
		a, _ := json.Marshal(lb)
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, a, "", "    ")
		fmt.Printf(string(prettyJSON.Bytes()))
	}
	lbs, err := slbclient.DescribeLoadBalancerAttribute(LOADBALANCER_ID)
	if err == nil {
		a, _ := json.Marshal(lbs)
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, a, "", "    ")
		fmt.Printf(string(prettyJSON.Bytes()))
	}
	listener, err := slbclient.DescribeLoadBalancerTCPListenerAttribute(LOADBALANCER_ID, 80)
	if err == nil {
		a, _ := json.Marshal(listener)
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, a, "", "    ")
		fmt.Printf(string(prettyJSON.Bytes()))
	}
}

func TestGetLoadBalancerAdditionalTags(t *testing.T) {
	tagTests := []struct {
		Annotations map[string]string
		Tags        map[string]string
	}{
		{
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAdditionalTags: "",
			},
			Tags: map[string]string{},
		},
		{
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAdditionalTags: "Key=Val",
			},
			Tags: map[string]string{
				"Key": "Val",
			},
		},
		{
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAdditionalTags: "Key1=Val1, Key2=Val2",
			},
			Tags: map[string]string{
				"Key1": "Val1",
				"Key2": "Val2",
			},
		},
		{
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAdditionalTags: "Key1=, Key2=Val2",
				"anotherKey": "anotherValue",
			},
			Tags: map[string]string{
				"Key1": "",
				"Key2": "Val2",
			},
		},
		{
			Annotations: map[string]string{
				"Nothing": "Key1=, Key2=Val2, Key3",
			},
			Tags: map[string]string{},
		},
		{
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAdditionalTags: "K=V K1=V2,Key1========, =====, ======Val, =Val, , 234,",
			},
			Tags: map[string]string{
				"K":    "V K1",
				"Key1": "",
				"234":  "",
			},
		},
	}

	for _, tagTest := range tagTests {
		result := getLoadBalancerAdditionalTags(tagTest.Annotations)
		for k, v := range result {
			if len(result) != len(tagTest.Tags) {
				t.Errorf("incorrect expected length: %v != %v", result, tagTest.Tags)
				continue
			}
			if tagTest.Tags[k] != v {
				t.Errorf("%s != %s", tagTest.Tags[k], v)
				continue
			}
		}
	}
}

// anomaly test case
func TestUpdateLoadBalancerWhenStartLoadBalancerFailed(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   "default",
				Name:        "service-test",
				UID:         types.UID(serviceUIDExist),
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
	// create service
	f.RunDefault(t, "create test service")

	// Simulate the failure of StartLoadBalancerListener()
	slbClient := f.Cloud.climgr.loadbalancer.c
	err := slbClient.StopLoadBalancerListener(LOADBALANCER_ID, int(listenPort1))
	if err != nil {
		t.Fatal(err.Error())
	}

	// update service
	f.RunDefault(t,"update test service")

	// check result
	res, err := slbClient.DescribeLoadBalancerTCPListenerAttribute(LOADBALANCER_ID, int(listenPort1))
	if err != nil {
		t.Fatalf("DescribeLoadBalancerTCPListenerAttribute error: %s", err.Error())
	}

	if res.Status != slb.Running {
		t.Fatalf("listener stop error.")
	}

}
