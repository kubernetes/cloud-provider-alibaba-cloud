package alibaba

import (
	"encoding/base64"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"testing"
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
