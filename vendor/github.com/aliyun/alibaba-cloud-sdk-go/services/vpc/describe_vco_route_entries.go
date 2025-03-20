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

// DescribeVcoRouteEntries invokes the vpc.DescribeVcoRouteEntries API synchronously
func (client *Client) DescribeVcoRouteEntries(request *DescribeVcoRouteEntriesRequest) (response *DescribeVcoRouteEntriesResponse, err error) {
	response = CreateDescribeVcoRouteEntriesResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeVcoRouteEntriesWithChan invokes the vpc.DescribeVcoRouteEntries API asynchronously
func (client *Client) DescribeVcoRouteEntriesWithChan(request *DescribeVcoRouteEntriesRequest) (<-chan *DescribeVcoRouteEntriesResponse, <-chan error) {
	responseChan := make(chan *DescribeVcoRouteEntriesResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeVcoRouteEntries(request)
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

// DescribeVcoRouteEntriesWithCallback invokes the vpc.DescribeVcoRouteEntries API asynchronously
func (client *Client) DescribeVcoRouteEntriesWithCallback(request *DescribeVcoRouteEntriesRequest, callback func(response *DescribeVcoRouteEntriesResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeVcoRouteEntriesResponse
		var err error
		defer close(result)
		response, err = client.DescribeVcoRouteEntries(request)
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

// DescribeVcoRouteEntriesRequest is the request struct for api DescribeVcoRouteEntries
type DescribeVcoRouteEntriesRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ClientToken          string           `position:"Query" name:"ClientToken"`
	PageNumber           requests.Integer `position:"Query" name:"PageNumber"`
	PageSize             requests.Integer `position:"Query" name:"PageSize"`
	RouteEntryType       string           `position:"Query" name:"RouteEntryType"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	VpnConnectionId      string           `position:"Query" name:"VpnConnectionId"`
}

// DescribeVcoRouteEntriesResponse is the response struct for api DescribeVcoRouteEntries
type DescribeVcoRouteEntriesResponse struct {
	*responses.BaseResponse
	TotalCount      int              `json:"TotalCount" xml:"TotalCount"`
	PageNumber      int              `json:"PageNumber" xml:"PageNumber"`
	PageSize        int              `json:"PageSize" xml:"PageSize"`
	RequestId       string           `json:"RequestId" xml:"RequestId"`
	VcoRouteEntries []VcoRouteEntrie `json:"VcoRouteEntries" xml:"VcoRouteEntries"`
	VpnRouteCounts  []VpnRouteCount  `json:"VpnRouteCounts" xml:"VpnRouteCounts"`
}

// CreateDescribeVcoRouteEntriesRequest creates a request to invoke DescribeVcoRouteEntries API
func CreateDescribeVcoRouteEntriesRequest() (request *DescribeVcoRouteEntriesRequest) {
	request = &DescribeVcoRouteEntriesRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Vpc", "2016-04-28", "DescribeVcoRouteEntries", "vpc", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDescribeVcoRouteEntriesResponse creates a response to parse from DescribeVcoRouteEntries response
func CreateDescribeVcoRouteEntriesResponse() (response *DescribeVcoRouteEntriesResponse) {
	response = &DescribeVcoRouteEntriesResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
