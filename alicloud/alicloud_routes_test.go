package alicloud

import (
	"testing"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"time"
)

func TestRoute(t *testing.T) {
	rsdk,err := NewSDKClientRoutes("","")
	if err !=nil {
		t.Fail()
	}

	route,err := rsdk.ListRoutes(common.Hangzhou, []string{"",""})
	if err != nil {
		t.Fail()
	}
	for _,r := range route{
		t.Log(fmt.Sprintf("%+v\n",r))
	}
}


func TestRouteExpire(t *testing.T) {
	rsdk,err := NewSDKClientRoutes("","")
	if err !=nil {
		t.Fail()
	}

	for i:=0;i<20;i++{
		_,err := rsdk.ListRoutes(common.Hangzhou, []string{""})
		if err != nil {
			t.Fail()
		}
		time.Sleep(time.Duration(3*time.Second))
		fmt.Printf("Log 1 second: %d",i)
	}

}
