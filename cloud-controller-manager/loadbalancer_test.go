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
	"github.com/denverdino/aliyungo/util"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"strings"
	"testing"
	"time"
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
	token := &TokenAuth{
		auth: metadata.RoleAuth{
			AccessKeyId:     "xxxxxxx",
			AccessKeySecret: "yyyyyyyyyyyyyyyyyyyyy",
		},
		active: false,
	}

	mgr := &ClientMgr{
		stop:  make(<-chan struct{}, 1),
		token: token,
		meta: metadata.NewMockMetaData(nil, func(resource string) (string, error) {
			if strings.Contains(resource, metadata.REGION) {
				return "region-test", nil
			}
			return "", errors.New("not found")
		}),
		loadbalancer: &LoadBalancerClient{
			c: client,
		},
	}
	return mgr, nil
}

func TestFindLoadBalancer(t *testing.T) {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   "default",
			Name:        "service-test",
			UID:         "abcdefghigklmnopqrstu",
			Annotations: map[string]string{
			//ServiceAnnotationLoadBalancerId: LOADBALANCER_ID,
			},
		},
		Spec: v1.ServiceSpec{
			Type: "LoadBalancer",
		},
	}

	base := newBaseLoadbalancer()
	mgr, _ := NewMockClientMgr(&mockClientSLB{
		describeLoadBalancers: func(args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error) {

			if args.LoadBalancerId != "" {
				base[0].LoadBalancerId = args.LoadBalancerId
				return base, nil
			}
			if args.LoadBalancerName != "" {
				base[0].LoadBalancerName = args.LoadBalancerName
				return base, nil
			}
			if len(args.Tags) > 0 {
				base[0].LoadBalancerName = cloudprovider.GetLoadBalancerName(service)
			} else {
				return nil, errors.New("loadbalancerid or loadbanancername must be specified.\n")
			}
			return base, nil
		},
		describeLoadBalancerAttribute: func(loadBalancerId string) (loadBalancer *slb.LoadBalancerType, err error) {
			t.Logf("findloadbalancer, [%s]", loadBalancerId)
			return loadbalancerAttrib(&base[0]), nil
		},
	})

	// 1.
	// user need to create new loadbalancer. did not specify any exist loadbalancer.
	// Expected fallback to use service UID to generate slb .
	exist, lb, err := mgr.loadbalancer.findLoadBalancer(service)
	if err != nil || !exist {
		t.Logf("1. user need to create new loadbalancer. did not specify any exist loadbalancer.")
		t.Fatal("Test findLoadBalancer fail.")
	}
	t.Logf("find loadbalancer: with name , [%s]", lb.LoadBalancerName)
	if lb.LoadBalancerName != cloudprovider.GetLoadBalancerName(service) {
		t.Fatal("find loadbalancer fail. suppose to find by name.")
	}

	// 2.
	// user need to use an exist loadbalancer through annotations
	service.Annotations[ServiceAnnotationLoadBalancerId] = LOADBALANCER_ID + "-new"
	exist, lb, err = mgr.loadbalancer.findLoadBalancer(service)
	if err != nil || !exist {
		t.Logf("2. user need to use an exist loadbalancer through annotations")
		t.Fatal("Test findLoadBalancer fail.")
	}
	if lb.LoadBalancerId != LOADBALANCER_ID+"-new" {
		t.Fatal("find loadbalancer fail. suppose to find by exist loadbalancerid.")
	}
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
		err = json.Indent(&prettyJSON, a, "", "    ")
		fmt.Printf(string(prettyJSON.Bytes()))
	}
	lbs, err := slbclient.DescribeLoadBalancerAttribute(LOADBALANCER_ID)
	if err == nil {
		a, _ := json.Marshal(lbs)
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, a, "", "    ")
		fmt.Printf(string(prettyJSON.Bytes()))
	}
	listener, err := slbclient.DescribeLoadBalancerTCPListenerAttribute(LOADBALANCER_ID, 80)
	if err == nil {
		a, _ := json.Marshal(listener)
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, a, "", "    ")
		fmt.Printf(string(prettyJSON.Bytes()))
	}
}

type mockClientSLB struct {
	describeLoadBalancers          func(args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error)
	createLoadBalancer             func(args *slb.CreateLoadBalancerArgs) (response *slb.CreateLoadBalancerResponse, err error)
	deleteLoadBalancer             func(loadBalancerId string) (err error)
	modifyLoadBalancerInternetSpec func(args *slb.ModifyLoadBalancerInternetSpecArgs) (err error)
	describeLoadBalancerAttribute  func(loadBalancerId string) (loadBalancer *slb.LoadBalancerType, err error)
	removeBackendServers           func(loadBalancerId string, backendServers []string) (result []slb.BackendServerType, err error)
	addBackendServers              func(loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error)

	stopLoadBalancerListener                   func(loadBalancerId string, port int) (err error)
	startLoadBalancerListener                  func(loadBalancerId string, port int) (err error)
	createLoadBalancerTCPListener              func(args *slb.CreateLoadBalancerTCPListenerArgs) (err error)
	createLoadBalancerUDPListener              func(args *slb.CreateLoadBalancerUDPListenerArgs) (err error)
	deleteLoadBalancerListener                 func(loadBalancerId string, port int) (err error)
	createLoadBalancerHTTPSListener            func(args *slb.CreateLoadBalancerHTTPSListenerArgs) (err error)
	createLoadBalancerHTTPListener             func(args *slb.CreateLoadBalancerHTTPListenerArgs) (err error)
	describeLoadBalancerHTTPSListenerAttribute func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error)
	describeLoadBalancerTCPListenerAttribute   func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error)
	describeLoadBalancerUDPListenerAttribute   func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerUDPListenerAttributeResponse, err error)
	describeLoadBalancerHTTPListenerAttribute  func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPListenerAttributeResponse, err error)

	setLoadBalancerHTTPListenerAttribute  func(args *slb.SetLoadBalancerHTTPListenerAttributeArgs) (err error)
	setLoadBalancerHTTPSListenerAttribute func(args *slb.SetLoadBalancerHTTPSListenerAttributeArgs) (err error)
	setLoadBalancerTCPListenerAttribute   func(args *slb.SetLoadBalancerTCPListenerAttributeArgs) (err error)
	setLoadBalancerUDPListenerAttribute   func(args *slb.SetLoadBalancerUDPListenerAttributeArgs) (err error)
	removeTags                            func(args *slb.RemoveTagsArgs) error
	describeTags                          func(args *slb.DescribeTagsArgs) (tags []slb.TagItemType, pagination *common.PaginationResult, err error)
	addTags                               func(args *slb.AddTagsArgs) error
}

var (
	LOADBALANCER_ID           = "lb-bp1ids9hmq5924m6uk5w1"
	LOADBALANCER_NAME         = "a594334ad276811e8a62700163e10c02"
	LOADBALANCER_ADDRESS      = "47.97.241.114"
	LOADBALANCER_NETWORK_TYPE = "classic"
)

func newBaseLoadbalancer() []slb.LoadBalancerType {
	return []slb.LoadBalancerType{
		{
			LoadBalancerId:     LOADBALANCER_ID,
			LoadBalancerName:   LOADBALANCER_NAME,
			LoadBalancerStatus: "active",
			Address:            LOADBALANCER_ADDRESS,
			RegionId:           "cn-hangzhou",
			RegionIdAlias:      "cn-hangzhou",
			AddressType:        "internet",
			VSwitchId:          "",
			VpcId:              "",
			NetworkType:        LOADBALANCER_NETWORK_TYPE,
			Bandwidth:          0,
			InternetChargeType: "4",
			CreateTime:         "2018-03-14T17:16Z",
			CreateTimeStamp:    util.NewISO6801Time(time.Now()),
		},
	}
}

func (c *mockClientSLB) DescribeLoadBalancers(args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error) {
	if c.describeLoadBalancers != nil {
		return c.describeLoadBalancers(args)
	}
	return newBaseLoadbalancer(), nil
}

func (c *mockClientSLB) StopLoadBalancerListener(loadBalancerId string, port int) (err error) {
	if c.stopLoadBalancerListener != nil {
		return c.stopLoadBalancerListener(loadBalancerId, port)
	}
	// return nil indicate no stop success
	return nil
}

func (c *mockClientSLB) CreateLoadBalancer(args *slb.CreateLoadBalancerArgs) (response *slb.CreateLoadBalancerResponse, err error) {
	if c.createLoadBalancer != nil {
		return c.createLoadBalancer(args)
	}
	return &slb.CreateLoadBalancerResponse{
		LoadBalancerId:   LOADBALANCER_ID,
		Address:          LOADBALANCER_ADDRESS,
		NetworkType:      LOADBALANCER_NETWORK_TYPE,
		LoadBalancerName: LOADBALANCER_NAME,
	}, nil
}
func (c *mockClientSLB) DeleteLoadBalancer(loadBalancerId string) (err error) {
	if c.deleteLoadBalancer != nil {
		return c.deleteLoadBalancer(loadBalancerId)
	}
	return nil
}
func (c *mockClientSLB) ModifyLoadBalancerInternetSpec(args *slb.ModifyLoadBalancerInternetSpecArgs) (err error) {
	if c.modifyLoadBalancerInternetSpec != nil {
		return c.modifyLoadBalancerInternetSpec(args)
	}
	return nil
}
func (c *mockClientSLB) DescribeLoadBalancerAttribute(loadBalancerId string) (loadBalancer *slb.LoadBalancerType, err error) {
	if c.describeLoadBalancerAttribute != nil {
		return c.describeLoadBalancerAttribute(loadBalancerId)
	}
	return nil, nil
}
func (c *mockClientSLB) RemoveBackendServers(loadBalancerId string, backendServers []string) (result []slb.BackendServerType, err error) {
	if c.removeBackendServers != nil {
		return c.removeBackendServers(loadBalancerId, backendServers)
	}
	return nil, nil
}
func (c *mockClientSLB) AddBackendServers(loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error) {
	if c.addBackendServers != nil {
		return c.addBackendServers(loadBalancerId, backendServers)
	}
	return nil, nil
}
func (c *mockClientSLB) StartLoadBalancerListener(loadBalancerId string, port int) (err error) {
	if c.startLoadBalancerListener != nil {
		return c.startLoadBalancerListener(loadBalancerId, port)
	}
	return nil
}
func (c *mockClientSLB) CreateLoadBalancerTCPListener(args *slb.CreateLoadBalancerTCPListenerArgs) (err error) {
	if c.createLoadBalancerTCPListener != nil {
		return c.createLoadBalancerTCPListener(args)
	}
	return nil
}
func (c *mockClientSLB) CreateLoadBalancerUDPListener(args *slb.CreateLoadBalancerUDPListenerArgs) (err error) {
	if c.createLoadBalancerUDPListener != nil {
		return c.createLoadBalancerUDPListener(args)
	}
	return nil
}
func (c *mockClientSLB) DeleteLoadBalancerListener(loadBalancerId string, port int) (err error) {
	if c.deleteLoadBalancerListener != nil {
		return c.deleteLoadBalancerListener(loadBalancerId, port)
	}
	return nil
}
func (c *mockClientSLB) CreateLoadBalancerHTTPSListener(args *slb.CreateLoadBalancerHTTPSListenerArgs) (err error) {
	if c.createLoadBalancerHTTPSListener != nil {
		return c.createLoadBalancerHTTPSListener(args)
	}
	return nil
}
func (c *mockClientSLB) CreateLoadBalancerHTTPListener(args *slb.CreateLoadBalancerHTTPListenerArgs) (err error) {
	if c.createLoadBalancerHTTPListener != nil {
		return c.createLoadBalancerHTTPListener(args)
	}
	return nil
}
func (c *mockClientSLB) DescribeLoadBalancerHTTPSListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error) {
	if c.describeLoadBalancerHTTPSListenerAttribute != nil {
		return c.describeLoadBalancerHTTPSListenerAttribute(loadBalancerId, port)
	}
	return nil, nil
}
func (c *mockClientSLB) DescribeLoadBalancerTCPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error) {
	if c.describeLoadBalancerTCPListenerAttribute != nil {
		return c.describeLoadBalancerTCPListenerAttribute(loadBalancerId, port)
	}
	return nil, nil
}
func (c *mockClientSLB) DescribeLoadBalancerUDPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerUDPListenerAttributeResponse, err error) {
	if c.describeLoadBalancerUDPListenerAttribute != nil {
		return c.describeLoadBalancerUDPListenerAttribute(loadBalancerId, port)
	}
	return nil, nil
}
func (c *mockClientSLB) DescribeLoadBalancerHTTPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPListenerAttributeResponse, err error) {
	if c.describeLoadBalancerHTTPListenerAttribute != nil {
		return c.describeLoadBalancerHTTPListenerAttribute(loadBalancerId, port)
	}
	return nil, nil
}

func (c *mockClientSLB) SetLoadBalancerHTTPListenerAttribute(args *slb.SetLoadBalancerHTTPListenerAttributeArgs) (err error) {
	if c.setLoadBalancerHTTPListenerAttribute != nil {
		return c.setLoadBalancerHTTPListenerAttribute(args)
	}
	return nil
}

func (c *mockClientSLB) SetLoadBalancerHTTPSListenerAttribute(args *slb.SetLoadBalancerHTTPSListenerAttributeArgs) (err error) {
	if c.setLoadBalancerHTTPSListenerAttribute != nil {
		return c.setLoadBalancerHTTPSListenerAttribute(args)
	}
	return nil
}

func (c *mockClientSLB) SetLoadBalancerTCPListenerAttribute(args *slb.SetLoadBalancerTCPListenerAttributeArgs) (err error) {
	if c.setLoadBalancerTCPListenerAttribute != nil {
		return c.setLoadBalancerTCPListenerAttribute(args)
	}
	return nil
}

func (c *mockClientSLB) SetLoadBalancerUDPListenerAttribute(args *slb.SetLoadBalancerUDPListenerAttributeArgs) (err error) {
	if c.setLoadBalancerUDPListenerAttribute != nil {
		return c.setLoadBalancerUDPListenerAttribute(args)
	}
	return nil
}

func (c *mockClientSLB) RemoveTags(args *slb.RemoveTagsArgs) error {
	if c.removeTags != nil {
		return c.removeTags(args)
	}
	return nil
}
func (c *mockClientSLB) DescribeTags(args *slb.DescribeTagsArgs) (tags []slb.TagItemType, pagination *common.PaginationResult, err error) {
	if c.describeTags != nil {
		return c.describeTags(args)
	}
	return []slb.TagItemType{}, nil, nil
}
func (c *mockClientSLB) AddTags(args *slb.AddTagsArgs) error {
	if c.addTags != nil {
		return c.addTags(args)
	}
	return nil
}
