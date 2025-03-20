package ecs

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// CreateLaunchTemplateVersion invokes the ecs.CreateLaunchTemplateVersion API synchronously
func (client *Client) CreateLaunchTemplateVersion(request *CreateLaunchTemplateVersionRequest) (response *CreateLaunchTemplateVersionResponse, err error) {
	response = CreateCreateLaunchTemplateVersionResponse()
	err = client.DoAction(request, response)
	return
}

// CreateLaunchTemplateVersionWithChan invokes the ecs.CreateLaunchTemplateVersion API asynchronously
func (client *Client) CreateLaunchTemplateVersionWithChan(request *CreateLaunchTemplateVersionRequest) (<-chan *CreateLaunchTemplateVersionResponse, <-chan error) {
	responseChan := make(chan *CreateLaunchTemplateVersionResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.CreateLaunchTemplateVersion(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// CreateLaunchTemplateVersionWithCallback invokes the ecs.CreateLaunchTemplateVersion API asynchronously
func (client *Client) CreateLaunchTemplateVersionWithCallback(request *CreateLaunchTemplateVersionRequest, callback func(response *CreateLaunchTemplateVersionResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *CreateLaunchTemplateVersionResponse
		var err error
		defer close(result)
		response, err = client.CreateLaunchTemplateVersion(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// CreateLaunchTemplateVersionRequest is the request struct for api CreateLaunchTemplateVersion
type CreateLaunchTemplateVersionRequest struct {
	*requests.RpcRequest
	LaunchTemplateName             string                                         `position:"Query" name:"LaunchTemplateName"`
	ResourceOwnerId                requests.Integer                               `position:"Query" name:"ResourceOwnerId"`
	HttpPutResponseHopLimit        requests.Integer                               `position:"Query" name:"HttpPutResponseHopLimit"`
	SecurityEnhancementStrategy    string                                         `position:"Query" name:"SecurityEnhancementStrategy"`
	NetworkType                    string                                         `position:"Query" name:"NetworkType"`
	KeyPairName                    string                                         `position:"Query" name:"KeyPairName"`
	SpotPriceLimit                 requests.Float                                 `position:"Query" name:"SpotPriceLimit"`
	ImageOwnerAlias                string                                         `position:"Query" name:"ImageOwnerAlias"`
	DeletionProtection             requests.Boolean                               `position:"Query" name:"DeletionProtection"`
	ResourceGroupId                string                                         `position:"Query" name:"ResourceGroupId"`
	HostName                       string                                         `position:"Query" name:"HostName"`
	SystemDiskIops                 requests.Integer                               `position:"Query" name:"SystemDisk.Iops"`
	Tag                            *[]CreateLaunchTemplateVersionTag              `position:"Query" name:"Tag"  type:"Repeated"`
	SystemDiskAutoSnapshotPolicyId string                                         `position:"Query" name:"SystemDisk.AutoSnapshotPolicyId"`
	AutoRenewPeriod                requests.Integer                               `position:"Query" name:"AutoRenewPeriod"`
	Period                         requests.Integer                               `position:"Query" name:"Period"`
	LaunchTemplateId               string                                         `position:"Query" name:"LaunchTemplateId"`
	Ipv6AddressCount               requests.Integer                               `position:"Query" name:"Ipv6AddressCount"`
	OwnerId                        requests.Integer                               `position:"Query" name:"OwnerId"`
	VSwitchId                      string                                         `position:"Query" name:"VSwitchId"`
	SpotStrategy                   string                                         `position:"Query" name:"SpotStrategy"`
	PrivateIpAddress               string                                         `position:"Query" name:"PrivateIpAddress"`
	SystemDiskBurstingEnabled      requests.Boolean                               `position:"Query" name:"SystemDisk.BurstingEnabled"`
	PeriodUnit                     string                                         `position:"Query" name:"PeriodUnit"`
	InstanceName                   string                                         `position:"Query" name:"InstanceName"`
	AutoRenew                      requests.Boolean                               `position:"Query" name:"AutoRenew"`
	InternetChargeType             string                                         `position:"Query" name:"InternetChargeType"`
	ZoneId                         string                                         `position:"Query" name:"ZoneId"`
	InternetMaxBandwidthIn         requests.Integer                               `position:"Query" name:"InternetMaxBandwidthIn"`
	VersionDescription             string                                         `position:"Query" name:"VersionDescription"`
	SystemDiskDeleteWithInstance   requests.Boolean                               `position:"Query" name:"SystemDisk.DeleteWithInstance"`
	ImageId                        string                                         `position:"Query" name:"ImageId"`
	IoOptimized                    string                                         `position:"Query" name:"IoOptimized"`
	SecurityGroupId                string                                         `position:"Query" name:"SecurityGroupId"`
	InternetMaxBandwidthOut        requests.Integer                               `position:"Query" name:"InternetMaxBandwidthOut"`
	Description                    string                                         `position:"Query" name:"Description"`
	SystemDiskCategory             string                                         `position:"Query" name:"SystemDisk.Category"`
	SystemDiskPerformanceLevel     string                                         `position:"Query" name:"SystemDisk.PerformanceLevel"`
	UserData                       string                                         `position:"Query" name:"UserData"`
	PasswordInherit                requests.Boolean                               `position:"Query" name:"PasswordInherit"`
	HttpEndpoint                   string                                         `position:"Query" name:"HttpEndpoint"`
	InstanceType                   string                                         `position:"Query" name:"InstanceType"`
	InstanceChargeType             string                                         `position:"Query" name:"InstanceChargeType"`
	EnableVmOsConfig               requests.Boolean                               `position:"Query" name:"EnableVmOsConfig"`
	DeploymentSetId                string                                         `position:"Query" name:"DeploymentSetId"`
	NetworkInterface               *[]CreateLaunchTemplateVersionNetworkInterface `position:"Query" name:"NetworkInterface"  type:"Repeated"`
	ResourceOwnerAccount           string                                         `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount                   string                                         `position:"Query" name:"OwnerAccount"`
	SystemDiskDiskName             string                                         `position:"Query" name:"SystemDisk.DiskName"`
	RamRoleName                    string                                         `position:"Query" name:"RamRoleName"`
	AutoReleaseTime                string                                         `position:"Query" name:"AutoReleaseTime"`
	CreditSpecification            string                                         `position:"Query" name:"CreditSpecification"`
	SpotDuration                   requests.Integer                               `position:"Query" name:"SpotDuration"`
	SecurityGroupIds               *[]string                                      `position:"Query" name:"SecurityGroupIds"  type:"Repeated"`
	DataDisk                       *[]CreateLaunchTemplateVersionDataDisk         `position:"Query" name:"DataDisk"  type:"Repeated"`
	SystemDiskProvisionedIops      requests.Integer                               `position:"Query" name:"SystemDisk.ProvisionedIops"`
	SystemDiskSize                 requests.Integer                               `position:"Query" name:"SystemDisk.Size"`
	VpcId                          string                                         `position:"Query" name:"VpcId"`
	HttpTokens                     string                                         `position:"Query" name:"HttpTokens"`
	SystemDiskDescription          string                                         `position:"Query" name:"SystemDisk.Description"`
	SystemDiskEncrypted            string                                         `position:"Query" name:"SystemDisk.Encrypted"`
}

// CreateLaunchTemplateVersionTag is a repeated param struct in CreateLaunchTemplateVersionRequest
type CreateLaunchTemplateVersionTag struct {
	Key   string `name:"Key"`
	Value string `name:"Value"`
}

// CreateLaunchTemplateVersionNetworkInterface is a repeated param struct in CreateLaunchTemplateVersionRequest
type CreateLaunchTemplateVersionNetworkInterface struct {
	VSwitchId                   string    `name:"VSwitchId"`
	NetworkInterfaceName        string    `name:"NetworkInterfaceName"`
	Description                 string    `name:"Description"`
	SecurityGroupId             string    `name:"SecurityGroupId"`
	PrimaryIpAddress            string    `name:"PrimaryIpAddress"`
	SecurityGroupIds            *[]string `name:"SecurityGroupIds" type:"Repeated"`
	InstanceType                string    `name:"InstanceType"`
	NetworkInterfaceTrafficMode string    `name:"NetworkInterfaceTrafficMode"`
	DeleteOnRelease             string    `name:"DeleteOnRelease"`
}

// CreateLaunchTemplateVersionDataDisk is a repeated param struct in CreateLaunchTemplateVersionRequest
type CreateLaunchTemplateVersionDataDisk struct {
	PerformanceLevel     string `name:"PerformanceLevel"`
	Description          string `name:"Description"`
	SnapshotId           string `name:"SnapshotId"`
	Size                 string `name:"Size"`
	Device               string `name:"Device"`
	DiskName             string `name:"DiskName"`
	Category             string `name:"Category"`
	DeleteWithInstance   string `name:"DeleteWithInstance"`
	Encrypted            string `name:"Encrypted"`
	ProvisionedIops      string `name:"ProvisionedIops"`
	BurstingEnabled      string `name:"BurstingEnabled"`
	AutoSnapshotPolicyId string `name:"AutoSnapshotPolicyId"`
}

// CreateLaunchTemplateVersionResponse is the response struct for api CreateLaunchTemplateVersion
type CreateLaunchTemplateVersionResponse struct {
	*responses.BaseResponse
	LaunchTemplateVersionNumber int64  `json:"LaunchTemplateVersionNumber" xml:"LaunchTemplateVersionNumber"`
	RequestId                   string `json:"RequestId" xml:"RequestId"`
	LaunchTemplateId            string `json:"LaunchTemplateId" xml:"LaunchTemplateId"`
}

// CreateCreateLaunchTemplateVersionRequest creates a request to invoke CreateLaunchTemplateVersion API
func CreateCreateLaunchTemplateVersionRequest() (request *CreateLaunchTemplateVersionRequest) {
	request = &CreateLaunchTemplateVersionRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ecs", "2014-05-26", "CreateLaunchTemplateVersion", "ecs", "openAPI")
	request.Method = requests.POST
	return
}

// CreateCreateLaunchTemplateVersionResponse creates a response to parse from CreateLaunchTemplateVersion response
func CreateCreateLaunchTemplateVersionResponse() (response *CreateLaunchTemplateVersionResponse) {
	response = &CreateLaunchTemplateVersionResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
