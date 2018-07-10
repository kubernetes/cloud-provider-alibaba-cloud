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
	"github.com/denverdino/aliyungo/common"
	"testing"
	"strings"
	"errors"
	"fmt"
	"github.com/denverdino/aliyungo/ecs"
	"k8s.io/apimachinery/pkg/types"
)


func NewMockClientInstanceMgr(client ClientInstanceSDK) (*ClientMgr, error) {

	mgr := &ClientMgr{
		instance: &InstanceClient{
			c: client,
		},
	}
	return mgr, nil
}

func TestInstanceRefeshInstance(t *testing.T) {
	instanceid := "i-2zecarjjmtkx3oru4233"
	mgr, err := NewMockClientInstanceMgr(&mockClientInstanceSDK{
		describeInstances: func(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error){
			if !strings.Contains(args.InstanceIds, instanceid) {
				return nil,nil, errors.New("not found")
			}
			instances = []ecs.InstanceAttributesType{
				{
					InstanceId: instanceid,
					ImageId:    "centos_7_04_64_20G_alibase_201701015.vhd",
					RegionId:   "cn-beijing",
					ZoneId:     "cn-beijing-f",
					InstanceType: "ecs.sn1ne.large",
					InstanceTypeFamily: "ecs.sn1ne",
					Status:     "running",
					InstanceNetworkType: "vpc",
					VpcAttributes: ecs.VpcAttributesType{
						VpcId: "vpc-2zeaybwqmvn6qgabfd3pe",
						VSwitchId: "vsw-2zeclpmxy66zzxj4cg4ls",
						PrivateIpAddress: ecs.IpAddressSetType{
							IpAddress: []string{"192.168.211.130"},
						},
					},
					InstanceChargeType: common.PostPaid,
				},
			}

			return instances, nil,nil
		},
	})
	id := "cn-hangzhou.i-2zecarjjmtkx3oru4233"
	if err != nil {
		t.Fatal(fmt.Sprintf("create client manager fail. [%s]\n", err.Error()))
	}
	ins, err := mgr.Instances().refreshInstance(types.NodeName(id), common.Beijing)
	if err != nil {
		t.Errorf("TestInstanceRefeshInstance error: %s\n", err.Error())
	}
	if ins.InstanceId != instanceid {
		t.Fatal("refresh instance error.")
	}
	ins, err = mgr.Instances().findInstanceByNode(types.NodeName(id))
	if err != nil {
		t.Fatal(fmt.Sprintf("findInstanceByNode error: %s\n", err.Error()))
	}
	if ins.InstanceId != instanceid {
		t.Fatal("find instance error.")
	}
}


type mockClientInstanceSDK struct {
	describeInstances func(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error)
}

func (m *mockClientInstanceSDK) DescribeInstances(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error){
	if m.describeInstances != nil {
		return m.describeInstances(args)
	}
	return nil,nil,errors.New("not implemented")
}

func TestNewMgr(t *testing.T) {
	_, err := NewMockClientInstanceMgr(&mockClientInstanceSDK{
		describeInstances: func(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error){
			return nil,nil,errors.New("not implemented")
		},
	})
	if err != nil {
		t.Fatal("error create new instance client")
	}
	realInsClient(keyid,keysecret)
}

func realInsClient(keyid, keysec string) {
	if keyid == "" || keysec == "" {
		return
	}
	nodeName := "i-2zecarjjmtkx3oru4233"
	cs := ecs.NewClient(keyid, keysec)
	cs.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	ins,_, _ := cs.DescribeInstances(&ecs.DescribeInstancesArgs{
		RegionId:    "cn-beijing",
		InstanceIds: fmt.Sprintf("[\"%s\"]", string(nodeName)),
	})
	fmt.Printf("%s\n", PrettyJson(ins))
}