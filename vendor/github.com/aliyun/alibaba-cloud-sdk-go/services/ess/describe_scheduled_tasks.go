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

// DescribeScheduledTasks invokes the ess.DescribeScheduledTasks API synchronously
func (client *Client) DescribeScheduledTasks(request *DescribeScheduledTasksRequest) (response *DescribeScheduledTasksResponse, err error) {
	response = CreateDescribeScheduledTasksResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeScheduledTasksWithChan invokes the ess.DescribeScheduledTasks API asynchronously
func (client *Client) DescribeScheduledTasksWithChan(request *DescribeScheduledTasksRequest) (<-chan *DescribeScheduledTasksResponse, <-chan error) {
	responseChan := make(chan *DescribeScheduledTasksResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeScheduledTasks(request)
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

// DescribeScheduledTasksWithCallback invokes the ess.DescribeScheduledTasks API asynchronously
func (client *Client) DescribeScheduledTasksWithCallback(request *DescribeScheduledTasksRequest, callback func(response *DescribeScheduledTasksResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeScheduledTasksResponse
		var err error
		defer close(result)
		response, err = client.DescribeScheduledTasks(request)
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

// DescribeScheduledTasksRequest is the request struct for api DescribeScheduledTasks
type DescribeScheduledTasksRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ScheduledAction      *[]string        `position:"Query" name:"ScheduledAction"  type:"Repeated"`
	ScalingGroupId       string           `position:"Query" name:"ScalingGroupId"`
	TaskName             string           `position:"Query" name:"TaskName"`
	PageNumber           requests.Integer `position:"Query" name:"PageNumber"`
	PageSize             requests.Integer `position:"Query" name:"PageSize"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	RecurrenceValue      string           `position:"Query" name:"RecurrenceValue"`
	ScheduledTaskName    *[]string        `position:"Query" name:"ScheduledTaskName"  type:"Repeated"`
	TaskEnabled          requests.Boolean `position:"Query" name:"TaskEnabled"`
	ScheduledTaskId      *[]string        `position:"Query" name:"ScheduledTaskId"  type:"Repeated"`
	RecurrenceType       string           `position:"Query" name:"RecurrenceType"`
}

// DescribeScheduledTasksResponse is the response struct for api DescribeScheduledTasks
type DescribeScheduledTasksResponse struct {
	*responses.BaseResponse
	RequestId      string         `json:"RequestId" xml:"RequestId"`
	PageNumber     int            `json:"PageNumber" xml:"PageNumber"`
	PageSize       int            `json:"PageSize" xml:"PageSize"`
	TotalCount     int            `json:"TotalCount" xml:"TotalCount"`
	ScheduledTasks ScheduledTasks `json:"ScheduledTasks" xml:"ScheduledTasks"`
}

// CreateDescribeScheduledTasksRequest creates a request to invoke DescribeScheduledTasks API
func CreateDescribeScheduledTasksRequest() (request *DescribeScheduledTasksRequest) {
	request = &DescribeScheduledTasksRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ess", "2014-08-28", "DescribeScheduledTasks", "ess", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDescribeScheduledTasksResponse creates a response to parse from DescribeScheduledTasks response
func CreateDescribeScheduledTasksResponse() (response *DescribeScheduledTasksResponse) {
	response = &DescribeScheduledTasksResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
