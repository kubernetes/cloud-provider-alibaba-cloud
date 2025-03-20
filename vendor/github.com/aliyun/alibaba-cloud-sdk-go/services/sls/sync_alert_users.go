package sls

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

// SyncAlertUsers invokes the sls.SyncAlertUsers API synchronously
func (client *Client) SyncAlertUsers(request *SyncAlertUsersRequest) (response *SyncAlertUsersResponse, err error) {
	response = CreateSyncAlertUsersResponse()
	err = client.DoAction(request, response)
	return
}

// SyncAlertUsersWithChan invokes the sls.SyncAlertUsers API asynchronously
func (client *Client) SyncAlertUsersWithChan(request *SyncAlertUsersRequest) (<-chan *SyncAlertUsersResponse, <-chan error) {
	responseChan := make(chan *SyncAlertUsersResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.SyncAlertUsers(request)
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

// SyncAlertUsersWithCallback invokes the sls.SyncAlertUsers API asynchronously
func (client *Client) SyncAlertUsersWithCallback(request *SyncAlertUsersRequest, callback func(response *SyncAlertUsersResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *SyncAlertUsersResponse
		var err error
		defer close(result)
		response, err = client.SyncAlertUsers(request)
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

// SyncAlertUsersRequest is the request struct for api SyncAlertUsers
type SyncAlertUsersRequest struct {
	*requests.RpcRequest
	App            string `position:"Body" name:"App"`
	SlsAccessToken string `position:"Query" name:"SlsAccessToken"`
	Users          string `position:"Body" name:"Users"`
}

// SyncAlertUsersResponse is the response struct for api SyncAlertUsers
type SyncAlertUsersResponse struct {
	*responses.BaseResponse
	Code      string `json:"Code" xml:"Code"`
	Message   string `json:"Message" xml:"Message"`
	Data      string `json:"Data" xml:"Data"`
	RequestId string `json:"RequestId" xml:"RequestId"`
	Success   bool   `json:"Success" xml:"Success"`
}

// CreateSyncAlertUsersRequest creates a request to invoke SyncAlertUsers API
func CreateSyncAlertUsersRequest() (request *SyncAlertUsersRequest) {
	request = &SyncAlertUsersRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Sls", "2019-10-23", "SyncAlertUsers", "sls", "openAPI")
	request.Method = requests.POST
	return
}

// CreateSyncAlertUsersResponse creates a response to parse from SyncAlertUsers response
func CreateSyncAlertUsersResponse() (response *SyncAlertUsersResponse) {
	response = &SyncAlertUsersResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
