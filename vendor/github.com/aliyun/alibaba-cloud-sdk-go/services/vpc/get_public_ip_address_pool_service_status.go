package vpc

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

// GetPublicIpAddressPoolServiceStatus invokes the vpc.GetPublicIpAddressPoolServiceStatus API synchronously
func (client *Client) GetPublicIpAddressPoolServiceStatus(request *GetPublicIpAddressPoolServiceStatusRequest) (response *GetPublicIpAddressPoolServiceStatusResponse, err error) {
	response = CreateGetPublicIpAddressPoolServiceStatusResponse()
	err = client.DoAction(request, response)
	return
}

// GetPublicIpAddressPoolServiceStatusWithChan invokes the vpc.GetPublicIpAddressPoolServiceStatus API asynchronously
func (client *Client) GetPublicIpAddressPoolServiceStatusWithChan(request *GetPublicIpAddressPoolServiceStatusRequest) (<-chan *GetPublicIpAddressPoolServiceStatusResponse, <-chan error) {
	responseChan := make(chan *GetPublicIpAddressPoolServiceStatusResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.GetPublicIpAddressPoolServiceStatus(request)
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

// GetPublicIpAddressPoolServiceStatusWithCallback invokes the vpc.GetPublicIpAddressPoolServiceStatus API asynchronously
func (client *Client) GetPublicIpAddressPoolServiceStatusWithCallback(request *GetPublicIpAddressPoolServiceStatusRequest, callback func(response *GetPublicIpAddressPoolServiceStatusResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *GetPublicIpAddressPoolServiceStatusResponse
		var err error
		defer close(result)
		response, err = client.GetPublicIpAddressPoolServiceStatus(request)
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

// GetPublicIpAddressPoolServiceStatusRequest is the request struct for api GetPublicIpAddressPoolServiceStatus
type GetPublicIpAddressPoolServiceStatusRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ClientToken          string           `position:"Query" name:"ClientToken"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
}

// GetPublicIpAddressPoolServiceStatusResponse is the response struct for api GetPublicIpAddressPoolServiceStatus
type GetPublicIpAddressPoolServiceStatusResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	Enabled   bool   `json:"Enabled" xml:"Enabled"`
}

// CreateGetPublicIpAddressPoolServiceStatusRequest creates a request to invoke GetPublicIpAddressPoolServiceStatus API
func CreateGetPublicIpAddressPoolServiceStatusRequest() (request *GetPublicIpAddressPoolServiceStatusRequest) {
	request = &GetPublicIpAddressPoolServiceStatusRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Vpc", "2016-04-28", "GetPublicIpAddressPoolServiceStatus", "vpc", "openAPI")
	request.Method = requests.POST
	return
}

// CreateGetPublicIpAddressPoolServiceStatusResponse creates a response to parse from GetPublicIpAddressPoolServiceStatus response
func CreateGetPublicIpAddressPoolServiceStatusResponse() (response *GetPublicIpAddressPoolServiceStatusResponse) {
	response = &GetPublicIpAddressPoolServiceStatusResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
