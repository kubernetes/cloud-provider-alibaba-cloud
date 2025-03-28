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

// DescribeVpnGatewayAvailableZones invokes the vpc.DescribeVpnGatewayAvailableZones API synchronously
func (client *Client) DescribeVpnGatewayAvailableZones(request *DescribeVpnGatewayAvailableZonesRequest) (response *DescribeVpnGatewayAvailableZonesResponse, err error) {
	response = CreateDescribeVpnGatewayAvailableZonesResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeVpnGatewayAvailableZonesWithChan invokes the vpc.DescribeVpnGatewayAvailableZones API asynchronously
func (client *Client) DescribeVpnGatewayAvailableZonesWithChan(request *DescribeVpnGatewayAvailableZonesRequest) (<-chan *DescribeVpnGatewayAvailableZonesResponse, <-chan error) {
	responseChan := make(chan *DescribeVpnGatewayAvailableZonesResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeVpnGatewayAvailableZones(request)
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

// DescribeVpnGatewayAvailableZonesWithCallback invokes the vpc.DescribeVpnGatewayAvailableZones API asynchronously
func (client *Client) DescribeVpnGatewayAvailableZonesWithCallback(request *DescribeVpnGatewayAvailableZonesRequest, callback func(response *DescribeVpnGatewayAvailableZonesResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeVpnGatewayAvailableZonesResponse
		var err error
		defer close(result)
		response, err = client.DescribeVpnGatewayAvailableZones(request)
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

// DescribeVpnGatewayAvailableZonesRequest is the request struct for api DescribeVpnGatewayAvailableZones
type DescribeVpnGatewayAvailableZonesRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	Spec                 string           `position:"Query" name:"Spec"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	AcceptLanguage       string           `position:"Query" name:"AcceptLanguage"`
}

// DescribeVpnGatewayAvailableZonesResponse is the response struct for api DescribeVpnGatewayAvailableZones
type DescribeVpnGatewayAvailableZonesResponse struct {
	*responses.BaseResponse
	RegionId            string            `json:"RegionId" xml:"RegionId"`
	RequestId           string            `json:"RequestId" xml:"RequestId"`
	AvailableZoneIdList []AvailableZoneId `json:"AvailableZoneIdList" xml:"AvailableZoneIdList"`
}

// CreateDescribeVpnGatewayAvailableZonesRequest creates a request to invoke DescribeVpnGatewayAvailableZones API
func CreateDescribeVpnGatewayAvailableZonesRequest() (request *DescribeVpnGatewayAvailableZonesRequest) {
	request = &DescribeVpnGatewayAvailableZonesRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Vpc", "2016-04-28", "DescribeVpnGatewayAvailableZones", "vpc", "openAPI")
	request.Method = requests.GET
	return
}

// CreateDescribeVpnGatewayAvailableZonesResponse creates a response to parse from DescribeVpnGatewayAvailableZones response
func CreateDescribeVpnGatewayAvailableZonesResponse() (response *DescribeVpnGatewayAvailableZonesResponse) {
	response = &DescribeVpnGatewayAvailableZonesResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
