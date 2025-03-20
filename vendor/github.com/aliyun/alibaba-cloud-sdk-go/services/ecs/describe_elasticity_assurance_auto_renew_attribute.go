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

// DescribeElasticityAssuranceAutoRenewAttribute invokes the ecs.DescribeElasticityAssuranceAutoRenewAttribute API synchronously
func (client *Client) DescribeElasticityAssuranceAutoRenewAttribute(request *DescribeElasticityAssuranceAutoRenewAttributeRequest) (response *DescribeElasticityAssuranceAutoRenewAttributeResponse, err error) {
	response = CreateDescribeElasticityAssuranceAutoRenewAttributeResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeElasticityAssuranceAutoRenewAttributeWithChan invokes the ecs.DescribeElasticityAssuranceAutoRenewAttribute API asynchronously
func (client *Client) DescribeElasticityAssuranceAutoRenewAttributeWithChan(request *DescribeElasticityAssuranceAutoRenewAttributeRequest) (<-chan *DescribeElasticityAssuranceAutoRenewAttributeResponse, <-chan error) {
	responseChan := make(chan *DescribeElasticityAssuranceAutoRenewAttributeResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeElasticityAssuranceAutoRenewAttribute(request)
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

// DescribeElasticityAssuranceAutoRenewAttributeWithCallback invokes the ecs.DescribeElasticityAssuranceAutoRenewAttribute API asynchronously
func (client *Client) DescribeElasticityAssuranceAutoRenewAttributeWithCallback(request *DescribeElasticityAssuranceAutoRenewAttributeRequest, callback func(response *DescribeElasticityAssuranceAutoRenewAttributeResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeElasticityAssuranceAutoRenewAttributeResponse
		var err error
		defer close(result)
		response, err = client.DescribeElasticityAssuranceAutoRenewAttribute(request)
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

// DescribeElasticityAssuranceAutoRenewAttributeRequest is the request struct for api DescribeElasticityAssuranceAutoRenewAttribute
type DescribeElasticityAssuranceAutoRenewAttributeRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	PrivatePoolOptionsId *[]string        `position:"Query" name:"PrivatePoolOptions.Id"  type:"Repeated"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
}

// DescribeElasticityAssuranceAutoRenewAttributeResponse is the response struct for api DescribeElasticityAssuranceAutoRenewAttribute
type DescribeElasticityAssuranceAutoRenewAttributeResponse struct {
	*responses.BaseResponse
	RequestId                          string                             `json:"RequestId" xml:"RequestId"`
	ElasticityAssuranceRenewAttributes ElasticityAssuranceRenewAttributes `json:"ElasticityAssuranceRenewAttributes" xml:"ElasticityAssuranceRenewAttributes"`
}

// CreateDescribeElasticityAssuranceAutoRenewAttributeRequest creates a request to invoke DescribeElasticityAssuranceAutoRenewAttribute API
func CreateDescribeElasticityAssuranceAutoRenewAttributeRequest() (request *DescribeElasticityAssuranceAutoRenewAttributeRequest) {
	request = &DescribeElasticityAssuranceAutoRenewAttributeRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ecs", "2014-05-26", "DescribeElasticityAssuranceAutoRenewAttribute", "ecs", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDescribeElasticityAssuranceAutoRenewAttributeResponse creates a response to parse from DescribeElasticityAssuranceAutoRenewAttribute response
func CreateDescribeElasticityAssuranceAutoRenewAttributeResponse() (response *DescribeElasticityAssuranceAutoRenewAttributeResponse) {
	response = &DescribeElasticityAssuranceAutoRenewAttributeResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
