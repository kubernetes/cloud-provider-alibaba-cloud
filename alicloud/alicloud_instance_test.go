package alicloud

import (
	"github.com/denverdino/aliyungo/common"
	"testing"
	"fmt"
)

func TestInstanceRefeshInstance(t *testing.T) {
	ins := NewSDKClientINS(keyid, keysecret)
	_, err := ins.refreshInstance("xxxx", common.Zhangjiakou)
	if err != nil {
		t.Errorf("TestInstanceRefeshInstance error: %s\n", err.Error())
	}
}

func TestReplaceCaml(t *testing.T) {
	fmt.Println(ServiceAnnotationLoadBalancerBackendLabel)
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerBackendLabel))
}