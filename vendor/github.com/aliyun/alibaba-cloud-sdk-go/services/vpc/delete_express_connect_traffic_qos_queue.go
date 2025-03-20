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

// DeleteExpressConnectTrafficQosQueue invokes the vpc.DeleteExpressConnectTrafficQosQueue API synchronously
func (client *Client) DeleteExpressConnectTrafficQosQueue(request *DeleteExpressConnectTrafficQosQueueRequest) (response *DeleteExpressConnectTrafficQosQueueResponse, err error) {
	response = CreateDeleteExpressConnectTrafficQosQueueResponse()
	err = client.DoAction(request, response)
	return
}

// DeleteExpressConnectTrafficQosQueueWithChan invokes the vpc.DeleteExpressConnectTrafficQosQueue API asynchronously
func (client *Client) DeleteExpressConnectTrafficQosQueueWithChan(request *DeleteExpressConnectTrafficQosQueueRequest) (<-chan *DeleteExpressConnectTrafficQosQueueResponse, <-chan error) {
	responseChan := make(chan *DeleteExpressConnectTrafficQosQueueResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DeleteExpressConnectTrafficQosQueue(request)
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

// DeleteExpressConnectTrafficQosQueueWithCallback invokes the vpc.DeleteExpressConnectTrafficQosQueue API asynchronously
func (client *Client) DeleteExpressConnectTrafficQosQueueWithCallback(request *DeleteExpressConnectTrafficQosQueueRequest, callback func(response *DeleteExpressConnectTrafficQosQueueResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DeleteExpressConnectTrafficQosQueueResponse
		var err error
		defer close(result)
		response, err = client.DeleteExpressConnectTrafficQosQueue(request)
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

// DeleteExpressConnectTrafficQosQueueRequest is the request struct for api DeleteExpressConnectTrafficQosQueue
type DeleteExpressConnectTrafficQosQueueRequest struct {
	*requests.RpcRequest
	ClientToken          string           `position:"Query" name:"ClientToken"`
	QosId                string           `position:"Query" name:"QosId"`
	QueueId              string           `position:"Query" name:"QueueId"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
}

// DeleteExpressConnectTrafficQosQueueResponse is the response struct for api DeleteExpressConnectTrafficQosQueue
type DeleteExpressConnectTrafficQosQueueResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateDeleteExpressConnectTrafficQosQueueRequest creates a request to invoke DeleteExpressConnectTrafficQosQueue API
func CreateDeleteExpressConnectTrafficQosQueueRequest() (request *DeleteExpressConnectTrafficQosQueueRequest) {
	request = &DeleteExpressConnectTrafficQosQueueRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Vpc", "2016-04-28", "DeleteExpressConnectTrafficQosQueue", "vpc", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDeleteExpressConnectTrafficQosQueueResponse creates a response to parse from DeleteExpressConnectTrafficQosQueue response
func CreateDeleteExpressConnectTrafficQosQueueResponse() (response *DeleteExpressConnectTrafficQosQueueResponse) {
	response = &DeleteExpressConnectTrafficQosQueueResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
