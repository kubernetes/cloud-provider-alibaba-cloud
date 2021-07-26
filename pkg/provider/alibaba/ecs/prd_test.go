package ecs

import (
	"encoding/base64"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func modifyUserData(client *ecs.Client, id, data string) error {
	req := ecs.CreateModifyInstanceAttributeRequest()
	req.InstanceId = id
	req.UserData = data
	_, err := client.ModifyInstanceAttribute(req)
	return err
}

func GetUserData(client *ecs.Client, id string) (string, error) {
	req := ecs.CreateDescribeUserDataRequest()
	req.InstanceId = id
	data, err := client.DescribeUserData(req)
	if err != nil {
		return "", fmt.Errorf("[ReplaceSystemDisk] find userdata %s: %s", id, err.Error())
	}
	return data.UserData, nil
}

func TestUserData(t *testing.T) {
	client, err := ecs.NewClientWithAccessKey(
		"", "", "",
	)
	if err != nil {
		t.Fatal(err)
		return
	}
	id := "i-wz9faoi1xyw09w9r994i"
	user, err := GetUserData(client, id)
	if err != nil {
		t.Fatal(err)
		return
	}
	if user != "" {
		data, err := base64.StdEncoding.DecodeString(user)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Logf("Old UserData: [%s][%s]", user, data)
	} else {
		t.Logf("empty userdata")
	}

	err = modifyUserData(client, id, base64.StdEncoding.EncodeToString([]byte("exit 0;")))
	if err != nil {
		t.Fatal(err)
		return
	}
	user, err = GetUserData(client, id)
	if err != nil {
		t.Fatal(err)
		return
	}
	if user != "" {
		data, err := base64.StdEncoding.DecodeString(user)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Logf("New UserData: [%s][%s]", user, data)
	} else {
		t.Logf("empty userdata after modified")
	}
}

func TestEcsProvider_ListInstances(t *testing.T) {
	client, err := ecs.NewClientWithAccessKey(
		"", "", "",
	)
	if err != nil {
		t.Fatal(err)
		return
	}
	ids := []string{
		"cn-hangzhou.i-bp10p8lib5ep9nu6chg",
		"cn-shanghai.i-uf606juqu3a9tpk2fzj",
		"cn-shanghai.i-uf606juqu3a9tpk2fzj",
		"cn-shanghai.i-uf606juqu3a9tpk2fzj",
		"cn-shanghai.i-uf658ihjgkfvrf6axff",
		"cn-shanghai.i-uf61cwrwp6rs3tplf32",
		"cn-shanghai.i-uf6jf9j3ltkgq49p1jy",
	}

	ecsProvider := NewEcsProvider(&base.ClientAuth{
		Meta: nil,
		ECS:  client,
		VPC:  nil,
		SLB:  nil,
		PVTZ: nil,
	})

	cloudNodes, err := ecsProvider.ListInstances(&node.NodeContext{}, ids)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	for _, id := range ids {
		c, ok := cloudNodes[id]
		if !ok || c == nil {
			t.Errorf("cannot find %s ecs from cloud", id)
		}
		t.Logf("%s, instance address: %s", id, c.Addresses)
	}

	t.Logf("ListInstances test successfully")
}
