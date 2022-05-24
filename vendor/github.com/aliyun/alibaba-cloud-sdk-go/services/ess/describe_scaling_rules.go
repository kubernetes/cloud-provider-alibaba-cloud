package ess

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

// DescribeScalingRules invokes the ess.DescribeScalingRules API synchronously
func (client *Client) DescribeScalingRules(request *DescribeScalingRulesRequest) (response *DescribeScalingRulesResponse, err error) {
	response = CreateDescribeScalingRulesResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeScalingRulesWithChan invokes the ess.DescribeScalingRules API asynchronously
func (client *Client) DescribeScalingRulesWithChan(request *DescribeScalingRulesRequest) (<-chan *DescribeScalingRulesResponse, <-chan error) {
	responseChan := make(chan *DescribeScalingRulesResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeScalingRules(request)
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

// DescribeScalingRulesWithCallback invokes the ess.DescribeScalingRules API asynchronously
func (client *Client) DescribeScalingRulesWithCallback(request *DescribeScalingRulesRequest, callback func(response *DescribeScalingRulesResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeScalingRulesResponse
		var err error
		defer close(result)
		response, err = client.DescribeScalingRules(request)
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

// DescribeScalingRulesRequest is the request struct for api DescribeScalingRules
type DescribeScalingRulesRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ScalingGroupId       string           `position:"Query" name:"ScalingGroupId"`
	ScalingRuleId        *[]string        `position:"Query" name:"ScalingRuleId"  type:"Repeated"`
	PageNumber           requests.Integer `position:"Query" name:"PageNumber"`
	ScalingRuleName      *[]string        `position:"Query" name:"ScalingRuleName"  type:"Repeated"`
	PageSize             requests.Integer `position:"Query" name:"PageSize"`
	ScalingRuleType      string           `position:"Query" name:"ScalingRuleType"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	ScalingRuleAri       *[]string        `position:"Query" name:"ScalingRuleAri"  type:"Repeated"`
	ShowAlarmRules       requests.Boolean `position:"Query" name:"ShowAlarmRules"`
}

// DescribeScalingRulesResponse is the response struct for api DescribeScalingRules
type DescribeScalingRulesResponse struct {
	*responses.BaseResponse
	RequestId    string       `json:"RequestId" xml:"RequestId"`
	PageNumber   int          `json:"PageNumber" xml:"PageNumber"`
	PageSize     int          `json:"PageSize" xml:"PageSize"`
	TotalCount   int          `json:"TotalCount" xml:"TotalCount"`
	ScalingRules ScalingRules `json:"ScalingRules" xml:"ScalingRules"`
}

// CreateDescribeScalingRulesRequest creates a request to invoke DescribeScalingRules API
func CreateDescribeScalingRulesRequest() (request *DescribeScalingRulesRequest) {
	request = &DescribeScalingRulesRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ess", "2014-08-28", "DescribeScalingRules", "ess", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDescribeScalingRulesResponse creates a response to parse from DescribeScalingRules response
func CreateDescribeScalingRulesResponse() (response *DescribeScalingRulesResponse) {
	response = &DescribeScalingRulesResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
