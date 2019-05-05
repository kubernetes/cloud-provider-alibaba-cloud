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
	"fmt"
	"testing"

	"github.com/denverdino/aliyungo/ecs"
)

func NewMockClientInstanceMgr() (*ClientMgr, error) {

	mgr := &ClientMgr{
		instance: &InstanceClient{
			c: &mockClientInstanceSDK{},
		},
	}
	return mgr, nil
}

func TestNewMgr(t *testing.T) {
	_, err := NewMockClientInstanceMgr()
	if err != nil {
		t.Fatal("error create new instance client")
	}
	realInsClient(keyid, keysecret)
}

func realInsClient(keyid, keysec string) {
	if keyid == "" || keysec == "" {
		return
	}
	nodeName := "i-2zecarjjmtkx3oru4233"
	cs := ecs.NewClient(keyid, keysec)
	cs.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	ins, _, _ := cs.DescribeInstances(&ecs.DescribeInstancesArgs{
		RegionId:    "cn-beijing",
		InstanceIds: fmt.Sprintf("[\"%s\"]", string(nodeName)),
	})
	fmt.Printf("%s\n", PrettyJson(ins))
}

// ======================================= This begins the TESTS ============================================

func TestInstanceRefeshInstance(t *testing.T) {

	mgr, err := NewMockClientInstanceMgr()
	if err != nil {
		t.Fatal(fmt.Sprintf("create client manager fail. [%s]\n", err.Error()))
	}

	PreSetCloudData(
		WithNewInstanceStore(),
		WithInstance(),
	)

	ins, err := mgr.Instances().refreshInstance(INSTANCEID, REGION)
	if err != nil {
		t.Errorf("TestInstanceRefeshInstance error: %s\n", err.Error())
	}
	if ins.InstanceId != INSTANCEID {
		t.Fatal("refresh instance error.")
	}
	ins, err = mgr.Instances().findInstanceByProviderID(fmt.Sprintf("%s.%s", REGION, INSTANCEID))
	if err != nil {
		t.Fatal(fmt.Sprintf("findInstanceByNode error: %s\n", err.Error()))
	}
	if ins.InstanceId != INSTANCEID {
		t.Fatal("find instance error.")
	}
}
