package config

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testCloudConfig = `
global:
  accessKeyId: dGVzdC1hY2Nlc3Mta2V5
  accessKeySecret: dGVzdC1rZXktc2VjcmV0
  clusterID: test-cluster-id
  resourceGroupID: rg-123456
  region: cn-hangzhou
  routeTableIDS: vtb-123456
  serviceBackendType: ecs
  uid: 123456
  vpcID: vpc-123456
  vswitchID: cn-hangzhou-a:vsw-123456,cn-wulanchabu-c:vsw-234567
  serviceMaxConcurrentReconciles: 5
  nodeMaxConcurrentReconciles: 2
`
)

func TestCloudConfig_SetDefaultValue(t *testing.T) {
	cfg := &CloudConfig{}
	cfg.SetDefaultValue()

	assert.Equal(t, DefaultServiceMaxConcurrentReconciles, cfg.Global.ServiceMaxConcurrentReconciles)
	assert.Equal(t, DefaultNodeMaxConcurrentReconciles, cfg.Global.NodeMaxConcurrentReconciles)
	assert.Equal(t, DefaultRouteMaxConcurrentReconciles, cfg.Global.RouteMaxConcurrentReconciles)

	cfg = &CloudConfig{}
	cfg.Global.ServiceMaxConcurrentReconciles = 10
	cfg.Global.NodeMaxConcurrentReconciles = 15
	cfg.Global.RouteMaxConcurrentReconciles = 20
	cfg.SetDefaultValue()

	assert.Equal(t, 10, cfg.Global.ServiceMaxConcurrentReconciles)
	assert.Equal(t, 15, cfg.Global.NodeMaxConcurrentReconciles)
	assert.Equal(t, 20, cfg.Global.RouteMaxConcurrentReconciles)
}

func TestCloudConfig_GetKubernetesClusterTag(t *testing.T) {
	cfg := &CloudConfig{}
	tag := cfg.GetKubernetesClusterTag()
	assert.Equal(t, util.ClusterTagKey, tag) // util.ClusterTagKey 的值

	cfg = &CloudConfig{}
	customTag := "my-custom-cluster-tag"
	cfg.Global.KubernetesClusterTag = customTag

	tag = cfg.GetKubernetesClusterTag()
	assert.Equal(t, customTag, tag)
}

func TestCloudConfig_LoadCloudCFG_YAML(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "cloud-config")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(testCloudConfig))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	oldPath := ControllerCFG.CloudConfigPath
	ControllerCFG.CloudConfigPath = tmpfile.Name()
	defer func() {
		ControllerCFG.CloudConfigPath = oldPath
	}()
	cfg := &CloudConfig{}
	err = cfg.LoadCloudCFG()
	assert.NoError(t, err)

	assert.Equal(t, "dGVzdC1hY2Nlc3Mta2V5", cfg.Global.AccessKeyID)
	assert.Equal(t, "dGVzdC1rZXktc2VjcmV0", cfg.Global.AccessKeySecret)
	assert.Equal(t, "test-cluster-id", cfg.Global.ClusterID)
	assert.Equal(t, "cn-hangzhou", cfg.Global.Region)
	assert.Equal(t, "vpc-123456", cfg.Global.VpcID)
	assert.Equal(t, "cn-hangzhou-a:vsw-123456,cn-wulanchabu-c:vsw-234567", cfg.Global.VswitchID)
	assert.Equal(t, "123456", cfg.Global.UID)
	assert.Equal(t, 5, cfg.Global.ServiceMaxConcurrentReconciles)
	assert.Equal(t, 2, cfg.Global.NodeMaxConcurrentReconciles)
}

func TestCloudConfig_LoadCloudCFG_JSON(t *testing.T) {
	config := `
{
  "Global": {
    "AccessKeyID": "dGVzdC1hY2Nlc3Mta2V5",
    "AccessKeySecret": "dGVzdC1rZXktc2VjcmV0",
    "ClusterID": "test-cluster-id",
    "ResourceGroupID": "rg-123456",
    "Region": "cn-hangzhou",
    "RouteTableIDS": "vtb-123456",
    "ServiceBackendType": "ecs",
    "UID": "123456",
    "VpcID": "vpc-123456",
    "VswitchID": "cn-hangzhou-a:vsw-123456,cn-wulanchabu-c:vsw-234567",
    "ServiceMaxConcurrentReconciles": 5,
    "NodeMaxConcurrentReconciles": 2
  }
}
`
	tmpfile, err := os.CreateTemp("", "cloud-config")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(config))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	oldPath := ControllerCFG.CloudConfigPath
	ControllerCFG.CloudConfigPath = tmpfile.Name()
	defer func() {
		ControllerCFG.CloudConfigPath = oldPath
	}()
	cfg := &CloudConfig{}
	err = cfg.LoadCloudCFG()
	assert.NoError(t, err)

	assert.Equal(t, "dGVzdC1hY2Nlc3Mta2V5", cfg.Global.AccessKeyID)
	assert.Equal(t, "dGVzdC1rZXktc2VjcmV0", cfg.Global.AccessKeySecret)
	assert.Equal(t, "test-cluster-id", cfg.Global.ClusterID)
	assert.Equal(t, "cn-hangzhou", cfg.Global.Region)
	assert.Equal(t, "vpc-123456", cfg.Global.VpcID)
	assert.Equal(t, "cn-hangzhou-a:vsw-123456,cn-wulanchabu-c:vsw-234567", cfg.Global.VswitchID)
	assert.Equal(t, "123456", cfg.Global.UID)
	assert.Equal(t, 5, cfg.Global.ServiceMaxConcurrentReconciles)
	assert.Equal(t, 2, cfg.Global.NodeMaxConcurrentReconciles)
}

func TestCloudConfig_LoadCloudCFG_FormatError(t *testing.T) {
	config := `
{
  "Global": {
        "AccessKeyID": ""
}
`
	tmpfile, err := os.CreateTemp("", "cloud-config")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(config))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	oldPath := ControllerCFG.CloudConfigPath
	ControllerCFG.CloudConfigPath = tmpfile.Name()
	defer func() {
		ControllerCFG.CloudConfigPath = oldPath
	}()
	cfg := &CloudConfig{}
	err = cfg.LoadCloudCFG()
	assert.Error(t, err)
}

func TestCloudConfig_LoadCloudCFG_FileNotFound(t *testing.T) {
	oldPath := ControllerCFG.CloudConfigPath
	ControllerCFG.CloudConfigPath = "/non/existent/file.yaml"
	defer func() {
		ControllerCFG.CloudConfigPath = oldPath
	}()

	cfg := &CloudConfig{}
	err := cfg.LoadCloudCFG()
	assert.Error(t, err)
}

func TestCloudConfig_PrintInfo(t *testing.T) {
	cfg := &CloudConfig{}
	cfg.Global.RouteTableIDS = "route-table-1,route-table-2"
	cfg.Global.ResourceGroupID = "rg-12345"
	cfg.Global.FeatureGates = "Feature1=true,Feature2=false"
	cfg.Global.NodeMaxConcurrentReconciles = 3
	cfg.Global.ServiceMaxConcurrentReconciles = 5
	cfg.Global.RouteMaxConcurrentReconciles = 2

	assert.NotPanics(t, func() {
		cfg.PrintInfo()
	})
}
