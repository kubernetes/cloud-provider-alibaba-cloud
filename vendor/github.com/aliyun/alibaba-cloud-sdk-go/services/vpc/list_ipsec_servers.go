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

// ListIpsecServers invokes the vpc.ListIpsecServers API synchronously
func (client *Client) ListIpsecServers(request *ListIpsecServersRequest) (response *ListIpsecServersResponse, err error) {
	response = CreateListIpsecServersResponse()
	err = client.DoAction(request, response)
	return
}

// ListIpsecServersWithChan invokes the vpc.ListIpsecServers API asynchronously
func (client *Client) ListIpsecServersWithChan(request *ListIpsecServersRequest) (<-chan *ListIpsecServersResponse, <-chan error) {
	responseChan := make(chan *ListIpsecServersResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ListIpsecServers(request)
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

// ListIpsecServersWithCallback invokes the vpc.ListIpsecServers API asynchronously
func (client *Client) ListIpsecServersWithCallback(request *ListIpsecServersRequest, callback func(response *ListIpsecServersResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ListIpsecServersResponse
		var err error
		defer close(result)
		response, err = client.ListIpsecServers(request)
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

// ListIpsecServersRequest is the request struct for api ListIpsecServers
type ListIpsecServersRequest struct {
	*requests.RpcRequest
	ResourceOwnerId requests.Integer `position:"Query" name:"ResourceOwnerId"`
	VpnGatewayId    string           `position:"Query" name:"VpnGatewayId"`
	CallerBid       string           `position:"Query" name:"callerBid"`
	ResourceGroupId string           `position:"Query" name:"ResourceGroupId"`
	NextToken       string           `position:"Query" name:"NextToken"`
	IpsecServerName string           `position:"Query" name:"IpsecServerName"`
	MaxResults      requests.Integer `position:"Query" name:"MaxResults"`
	IpsecServerId   *[]string        `position:"Query" name:"IpsecServerId"  type:"Repeated"`
}

// ListIpsecServersResponse is the response struct for api ListIpsecServers
type ListIpsecServersResponse struct {
	*responses.BaseResponse
	NextToken    string        `json:"NextToken" xml:"NextToken"`
	RequestId    string        `json:"RequestId" xml:"RequestId"`
	TotalCount   int           `json:"TotalCount" xml:"TotalCount"`
	MaxResults   int           `json:"MaxResults" xml:"MaxResults"`
	IpsecServers []IpsecServer `json:"IpsecServers" xml:"IpsecServers"`
}

// CreateListIpsecServersRequest creates a request to invoke ListIpsecServers API
func CreateListIpsecServersRequest() (request *ListIpsecServersRequest) {
	request = &ListIpsecServersRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Vpc", "2016-04-28", "ListIpsecServers", "vpc", "openAPI")
	request.Method = requests.POST
	return
}

// CreateListIpsecServersResponse creates a response to parse from ListIpsecServers response
func CreateListIpsecServersResponse() (response *ListIpsecServersResponse) {
	response = &ListIpsecServersResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
