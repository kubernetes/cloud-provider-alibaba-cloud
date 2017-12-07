package alicloud

import (
	"github.com/denverdino/aliyungo/common"
	"testing"
	"fmt"
)

func TestInstanceRefeshInstance(t *testing.T) {
	mgr,err := NewClientMgr("","")
	if err != nil {
		t.Errorf("create client manager fail. [%s]\n",err.Error())
		t.Fatal()
	}
	fmt.Printf("Acloud: [%+v][%+v]\n", mgr,err)
	_, e := mgr.Instances(DEFAULT_REGION).refreshInstance("xxxx", common.Zhangjiakou)
	if e != nil {
		t.Errorf("TestInstanceRefeshInstance error: %s\n", err.Error())
	}
}

func TestReplaceCaml(t *testing.T) {
	fmt.Println(ServiceAnnotationLoadBalancerBackendLabel)
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerBackendLabel))
}