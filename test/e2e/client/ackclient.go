package client

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
	"time"

	cs "github.com/alibabacloud-go/cs-20151215/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/klog/v2"
)

func NewACKClient() (*ACKClient, error) {
	ak, sk, err := base.LoadAK()
	if err != nil {
		return nil, fmt.Errorf("create ack client error: load ak error: %s", err.Error())
	}
	region := options.TestConfig.RegionId
	config := &openapi.Config{
		AccessKeyId:     &ak,
		AccessKeySecret: &sk,
		RegionId:        &region,
	}
	c, err := cs.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("create ack client error: %s", err.Error())
	}
	return &ACKClient{client: c}, nil
}

type ACKClient struct {
	client *cs.Client
}

func (e *ACKClient) DescribeClusterDetail(clusterId string) (*cs.DescribeClusterDetailResponseBody, error) {
	resp, err := e.client.DescribeClusterDetail(&clusterId)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("DescribeClusterDetail resp is nil")
	}
	return resp.Body, err
}

func (e *ACKClient) getDefaultNodePool(clusterId string) (*string, error) {
	detail, err := e.client.DescribeClusterNodePools(tea.String(clusterId))
	if err != nil {
		return nil, err
	}
	for _, np := range detail.Body.Nodepools {
		if np.NodepoolInfo == nil {
			continue
		}
		if np.NodepoolInfo.IsDefault != nil && *np.NodepoolInfo.IsDefault {
			return np.NodepoolInfo.NodepoolId, nil
		}
		if np.NodepoolInfo.Name != nil && *np.NodepoolInfo.Name == "default-nodepool" {
			return np.NodepoolInfo.NodepoolId, nil
		}
	}
	return nil, fmt.Errorf("get defalt NodepoolId fail")
}

func (e *ACKClient) ScaleOutCluster(clusterId string, nodeCount int64) error {
	nodePoolId, err := e.getDefaultNodePool(clusterId)
	if err != nil {
		return err
	}
	scaleOutNodePoolRequest := &cs.ScaleClusterNodePoolRequest{
		Count: &nodeCount,
	}
	resp, err := e.client.ScaleClusterNodePool(tea.String(clusterId), nodePoolId, scaleOutNodePoolRequest)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("DescribeClusterDetail resp is nil")
	}

	return wait.PollImmediate(30*time.Second, 5*time.Minute, func() (done bool, err error) {
		detail, err := e.DescribeClusterDetail(clusterId)
		if err != nil {
			return false, nil
		}
		if detail == nil || detail.State == nil {
			return false, nil
		}
		klog.Infof("waiting cluster state to be running, now %s", *detail.State)
		return *detail.State == "running", nil
	})
}

func (e *ACKClient) DeleteClusterNodes(clusterId, nodeName string) error {
	deleteClusterNodesRequest := &cs.DeleteClusterNodesRequest{
		DrainNode:   tea.Bool(true),
		ReleaseNode: tea.Bool(true),
		Nodes:       []*string{tea.String(nodeName)},
	}
	_, err := e.client.DeleteClusterNodes(tea.String(clusterId), deleteClusterNodesRequest)
	if err != nil {
		return err
	}
	return wait.PollImmediate(30*time.Second, 5*time.Minute, func() (done bool, err error) {
		detail, err := e.DescribeClusterDetail(clusterId)
		if err != nil {
			return false, err
		}
		if *detail.State == "running" {
			return true, nil
		} else {
			klog.Infof("waiting cluster state to be running, now is %s", *detail.State)
		}
		return false, err
	})
}

type AddonInfo struct {
	ComponentName string
	Version       string
}

func (e *ACKClient) DescribeClusterAddonsUpgradeStatus(clusterId string, componentId string) (*AddonInfo, error) {
	req := &cs.DescribeClusterAddonsUpgradeStatusRequest{
		ComponentIds: []*string{tea.String(componentId)},
	}
	resp, err := e.client.DescribeClusterAddonsUpgradeStatus(tea.String(clusterId), req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("DescribeClusterAddonsUpgradeStatus resp is nil")
	}
	for key, value := range resp.Body {
		if key == componentId {
			addonInfo := value.(map[string]interface{})["addon_info"].(map[string]interface{})
			return &AddonInfo{
				ComponentName: addonInfo["component_name"].(string),
				Version:       addonInfo["version"].(string),
			}, nil
		}
	}

	return nil, fmt.Errorf("DescribeClusterAddonsUpgradeStatus resp body %v do not contains %s", resp.Body, componentId)
}

func (e *ACKClient) ModifyClusterConfiguration(clusterId string, addonName string, configs map[string]string) error {
	var customizeConfigs []*cs.ModifyClusterConfigurationRequestCustomizeConfigConfigs
	for k, v := range configs {
		customizeConfigs = append(customizeConfigs, &cs.ModifyClusterConfigurationRequestCustomizeConfigConfigs{
			Key:   tea.String(k),
			Value: tea.String(v),
		})
	}

	req := &cs.ModifyClusterConfigurationRequest{
		CustomizeConfig: []*cs.ModifyClusterConfigurationRequestCustomizeConfig{
			{
				Name:    tea.String(addonName),
				Configs: customizeConfigs,
			},
		},
	}
	_, err := e.client.ModifyClusterConfiguration(tea.String(clusterId), req)
	return err
}
