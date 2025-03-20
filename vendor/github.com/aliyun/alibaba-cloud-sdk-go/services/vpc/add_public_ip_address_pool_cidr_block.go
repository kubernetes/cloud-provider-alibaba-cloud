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

// AddPublicIpAddressPoolCidrBlock invokes the vpc.AddPublicIpAddressPoolCidrBlock API synchronously
func (client *Client) AddPublicIpAddressPoolCidrBlock(request *AddPublicIpAddressPoolCidrBlockRequest) (response *AddPublicIpAddressPoolCidrBlockResponse, err error) {
	response = CreateAddPublicIpAddressPoolCidrBlockResponse()
	err = client.DoAction(request, response)
	return
}

// AddPublicIpAddressPoolCidrBlockWithChan invokes the vpc.AddPublicIpAddressPoolCidrBlock API asynchronously
func (client *Client) AddPublicIpAddressPoolCidrBlockWithChan(request *AddPublicIpAddressPoolCidrBlockRequest) (<-chan *AddPublicIpAddressPoolCidrBlockResponse, <-chan error) {
	responseChan := make(chan *AddPublicIpAddressPoolCidrBlockResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.AddPublicIpAddressPoolCidrBlock(request)
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

// AddPublicIpAddressPoolCidrBlockWithCallback invokes the vpc.AddPublicIpAddressPoolCidrBlock API asynchronously
func (client *Client) AddPublicIpAddressPoolCidrBlockWithCallback(request *AddPublicIpAddressPoolCidrBlockRequest, callback func(response *AddPublicIpAddressPoolCidrBlockResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *AddPublicIpAddressPoolCidrBlockResponse
		var err error
		defer close(result)
		response, err = client.AddPublicIpAddressPoolCidrBlock(request)
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

// AddPublicIpAddressPoolCidrBlockRequest is the request struct for api AddPublicIpAddressPoolCidrBlock
type AddPublicIpAddressPoolCidrBlockRequest struct {
	*requests.RpcRequest
	CidrMask              requests.Integer `position:"Query" name:"CidrMask"`
	PublicIpAddressPoolId string           `position:"Query" name:"PublicIpAddressPoolId"`
	ResourceOwnerId       requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ClientToken           string           `position:"Query" name:"ClientToken"`
	DryRun                requests.Boolean `position:"Query" name:"DryRun"`
	ResourceOwnerAccount  string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount          string           `position:"Query" name:"OwnerAccount"`
	OwnerId               requests.Integer `position:"Query" name:"OwnerId"`
	CidrBlock             string           `position:"Query" name:"CidrBlock"`
}

// AddPublicIpAddressPoolCidrBlockResponse is the response struct for api AddPublicIpAddressPoolCidrBlock
type AddPublicIpAddressPoolCidrBlockResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	CidrBlock string `json:"CidrBlock" xml:"CidrBlock"`
}

// CreateAddPublicIpAddressPoolCidrBlockRequest creates a request to invoke AddPublicIpAddressPoolCidrBlock API
func CreateAddPublicIpAddressPoolCidrBlockRequest() (request *AddPublicIpAddressPoolCidrBlockRequest) {
	request = &AddPublicIpAddressPoolCidrBlockRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Vpc", "2016-04-28", "AddPublicIpAddressPoolCidrBlock", "vpc", "openAPI")
	request.Method = requests.POST
	return
}

// CreateAddPublicIpAddressPoolCidrBlockResponse creates a response to parse from AddPublicIpAddressPoolCidrBlock response
func CreateAddPublicIpAddressPoolCidrBlockResponse() (response *AddPublicIpAddressPoolCidrBlockResponse) {
	response = &AddPublicIpAddressPoolCidrBlockResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
