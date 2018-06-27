package ecs

import (
	"errors"
	"encoding/base64"
	"time"

	"github.com/denverdino/aliyungo/common"
)


const (
	BazaarEndpoint = "https://ecs-cn-hangzhou.aliyuncs.com"
	BazaarAPIVersion = "2016-03-14"
	BazaarServiceCode = "EcsInc"
)

// ---------------------------------------
// ---------------------------------------
func NewBazaarClientWithRegion(accessKeyId string, accessKeySecret string, regionID common.Region) *Client {
	client := &Client{}
	client.NewInit(BazaarEndpoint, BazaarAPIVersion, accessKeyId, accessKeySecret, BazaarServiceCode, regionID)
	return client
}

// -----------------------------------------------------
// LaunchBazaarInstance
// -----------------------------------------------------
type LaunchBazaarInstanceArgs struct {
	RegionId                    common.Region
	ZoneId                      string
	ImageId                     string
	LinkedSecurityGroupId       string
	LinkedVSwitchId             string
	InstanceName                string
	HostName                    string
	Password                    string
	VSwitchId                   string
	KeyPairName                 string
	RamRoleName                 string
	Cpu			    float64
	Mem			    float64
	AssumeRoleAccessKeyId	    string
	AssumeRoleAccessKeySecret   string
	AssumeRoleSecurityToken     string
	LaunchData                  string
}

type LaunchBazaarInstanceResponse struct {
	common.Response
	InstanceId string
}

// CreateInstance creates instance
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&createinstance
func (client *Client) LaunchBazaarInstance(args *LaunchBazaarInstanceArgs) (instanceId string, err error) {
	if args.LaunchData != "" {
		// Encode to base64 string
		args.LaunchData = base64.StdEncoding.EncodeToString([]byte(args.LaunchData))
	}
	response := LaunchBazaarInstanceResponse{}
	err = client.Invoke("LaunchBazaarInstance", args, &response)
	if err != nil {
		return "", err
	}
	return response.InstanceId, err
}

type DescribeBazaarInstancesArgs struct {
	RegionId            common.Region
	LinkedVSwitchId     string
	InstanceIds         []string
	common.Pagination
}

type BazaarInstanceAttributesType struct {
	InstanceId         string
	Cpu		   float64
	Memory		   float64
	InstanceName       string
	RegionId           common.Region
	LinkedVSwitchId    string
	NetworkInterfaces  struct { NetworkInterface []NetworkInterfaceType }
	Status             InstanceStatus
	//Tags                    struct {
	//	Tag []TagItemType
	//}
}

type DescribeBazaarInstancesResponse struct {
	common.Response
	common.PaginationResult
	InstanceSet struct {
		InstanceSet []BazaarInstanceAttributesType
	}
}

// DescribeBazaarInstances describes instances
func (client *Client) DescribeBazaarInstances(args *DescribeBazaarInstancesArgs) (instances []BazaarInstanceAttributesType, pagination *common.PaginationResult,
					      err error) {
	args.Validate()
	response := &DescribeBazaarInstancesResponse{}

	err = client.Invoke("DescribeBazaarInstances", args, &response)
	if err != nil {
		return nil, nil, err
	}

	return response.InstanceSet.InstanceSet, &response.PaginationResult, nil
}

// WaitForBazaarInstance waits for instance to given status
func (client *Client) WaitForBazaarInstance(instanceId string, status InstanceStatus, timeout int) error {
	if timeout <= 0 {
		timeout = InstanceDefaultTimeout
	}
	for {
		describeBazaarInstancesArgs := DescribeBazaarInstancesArgs {
			InstanceIds: []string{instanceId},
		}
		instances, _, err := client.DescribeBazaarInstances(&describeBazaarInstancesArgs)
		if err != nil {
			return err
		}

		if len(instances) == 0 {
			return errors.New("WaitForBazaarInstance failed. Cannot found instanceId\n")
		}

		if instances[0].Status == status {
			//Sleep one more time for timing issues
			time.Sleep(DefaultWaitForInterval * time.Second)
			break
		}
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return common.GetClientErrorFromString("Timeout")
		}
		time.Sleep(DefaultWaitForInterval * time.Second)

	}
	return nil
}

type TerminateBazaarInstanceArgs struct {
	InstanceId string
}

type TerminateBazaarInstanceResponse struct {
	common.Response
}

// DeleteInstance deletes instance
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&deleteinstance
func (client *Client) TerminateBazaarInstance(instanceId string) error {
	args := TerminateBazaarInstanceArgs{InstanceId: instanceId}
	response := TerminateBazaarInstanceResponse{}
	err := client.Invoke("TerminateBazaarInstance", &args, &response)
	return err
}

