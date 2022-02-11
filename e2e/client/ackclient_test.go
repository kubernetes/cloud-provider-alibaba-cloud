package client

import (
	cs "github.com/alibabacloud-go/cs-20151215/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"testing"
)

func CreateACKClient() (*ACKClient, error) {
	ak, sk, region := "<your-access-key>", "<your-access-secret>", "<region>"
	config := &openapi.Config{
		AccessKeyId:     &ak,
		AccessKeySecret: &sk,
		RegionId:        &region,
	}
	c, err := cs.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ACKClient{client: c}, nil

}

func TestACKClient_DescribeClusterDetail(t *testing.T) {
	client, err := CreateACKClient()
	if err != nil {
		t.Fatalf(err.Error())
	}

	resp, err := client.DescribeClusterDetail("<your-cluster-id>")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf(*resp.State)

}
func TestACKClient_ScaleOutCluster(t *testing.T) {
	client, err := CreateACKClient()
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = client.ScaleOutCluster("<your-cluster-id>", 1)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestACKClient_DeleteClusterNodes(t *testing.T) {
	client, err := CreateACKClient()
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = client.DeleteClusterNodes("<your-cluster-id>", "<your-node-name>")
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestACKClient_DescribeClusterAddonsVersion(t *testing.T) {
	client, err := CreateACKClient()
	if err != nil {
		t.Fatalf(err.Error())
	}

	addon, err := client.DescribeClusterAddonsUpgradeStatus("<your-cluster-id>", "terway-eniip")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%v", addon)
}

func TestNewACKClient_ModifyClusterConfiguration(t *testing.T) {
	client, err := CreateACKClient()
	if err != nil {
		t.Fatalf(err.Error())
	}
	configs := make(map[string]string)
	configs["RouteTableIDS"] = ""

	err = client.ModifyClusterConfiguration("<your-cluster-id>", "cloud-controller-manager",
		configs)
	if err != nil {
		t.Fatalf(err.Error())
	}

}
