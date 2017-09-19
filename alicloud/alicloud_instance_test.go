package alicloud

import (
	"github.com/denverdino/aliyungo/common"
	"testing"
)

func TestInstanceRefeshInstance(t *testing.T) {
	ins := NewSDKClientINS(keyid, keysecret)
	_, err := ins.refreshInstance("xxxx", common.Zhangjiakou)
	if err != nil {
		t.Errorf("TestInstanceRefeshInstance error: %s\n", err.Error())
	}
}
