package pvtz

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

// BindResolverRuleVpc invokes the pvtz.BindResolverRuleVpc API synchronously
func (client *Client) BindResolverRuleVpc(request *BindResolverRuleVpcRequest) (response *BindResolverRuleVpcResponse, err error) {
	response = CreateBindResolverRuleVpcResponse()
	err = client.DoAction(request, response)
	return
}

// BindResolverRuleVpcWithChan invokes the pvtz.BindResolverRuleVpc API asynchronously
func (client *Client) BindResolverRuleVpcWithChan(request *BindResolverRuleVpcRequest) (<-chan *BindResolverRuleVpcResponse, <-chan error) {
	responseChan := make(chan *BindResolverRuleVpcResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.BindResolverRuleVpc(request)
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

// BindResolverRuleVpcWithCallback invokes the pvtz.BindResolverRuleVpc API asynchronously
func (client *Client) BindResolverRuleVpcWithCallback(request *BindResolverRuleVpcRequest, callback func(response *BindResolverRuleVpcResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *BindResolverRuleVpcResponse
		var err error
		defer close(result)
		response, err = client.BindResolverRuleVpc(request)
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

// BindResolverRuleVpcRequest is the request struct for api BindResolverRuleVpc
type BindResolverRuleVpcRequest struct {
	*requests.RpcRequest
	Vpc          *[]BindResolverRuleVpcVpc `position:"Query" name:"Vpc"  type:"Repeated"`
	UserClientIp string                    `position:"Query" name:"UserClientIp"`
	Lang         string                    `position:"Query" name:"Lang"`
	RuleId       string                    `position:"Query" name:"RuleId"`
}

// BindResolverRuleVpcVpc is a repeated param struct in BindResolverRuleVpcRequest
type BindResolverRuleVpcVpc struct {
	VpcType  string `name:"VpcType"`
	RegionId string `name:"RegionId"`
	VpcId    string `name:"VpcId"`
}

// BindResolverRuleVpcResponse is the response struct for api BindResolverRuleVpc
type BindResolverRuleVpcResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateBindResolverRuleVpcRequest creates a request to invoke BindResolverRuleVpc API
func CreateBindResolverRuleVpcRequest() (request *BindResolverRuleVpcRequest) {
	request = &BindResolverRuleVpcRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("pvtz", "2018-01-01", "BindResolverRuleVpc", "pvtz", "openAPI")
	request.Method = requests.POST
	return
}

// CreateBindResolverRuleVpcResponse creates a response to parse from BindResolverRuleVpc response
func CreateBindResolverRuleVpcResponse() (response *BindResolverRuleVpcResponse) {
	response = &BindResolverRuleVpcResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
