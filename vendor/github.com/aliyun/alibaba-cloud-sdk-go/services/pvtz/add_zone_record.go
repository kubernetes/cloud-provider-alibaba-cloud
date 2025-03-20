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

// AddZoneRecord invokes the pvtz.AddZoneRecord API synchronously
func (client *Client) AddZoneRecord(request *AddZoneRecordRequest) (response *AddZoneRecordResponse, err error) {
	response = CreateAddZoneRecordResponse()
	err = client.DoAction(request, response)
	return
}

// AddZoneRecordWithChan invokes the pvtz.AddZoneRecord API asynchronously
func (client *Client) AddZoneRecordWithChan(request *AddZoneRecordRequest) (<-chan *AddZoneRecordResponse, <-chan error) {
	responseChan := make(chan *AddZoneRecordResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.AddZoneRecord(request)
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

// AddZoneRecordWithCallback invokes the pvtz.AddZoneRecord API asynchronously
func (client *Client) AddZoneRecordWithCallback(request *AddZoneRecordRequest, callback func(response *AddZoneRecordResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *AddZoneRecordResponse
		var err error
		defer close(result)
		response, err = client.AddZoneRecord(request)
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

// AddZoneRecordRequest is the request struct for api AddZoneRecord
type AddZoneRecordRequest struct {
	*requests.RpcRequest
	Rr           string           `position:"Query" name:"Rr"`
	ClientToken  string           `position:"Query" name:"ClientToken"`
	Line         string           `position:"Query" name:"Line"`
	Remark       string           `position:"Query" name:"Remark"`
	Type         string           `position:"Query" name:"Type"`
	Lang         string           `position:"Query" name:"Lang"`
	Value        string           `position:"Query" name:"Value"`
	Weight       requests.Integer `position:"Query" name:"Weight"`
	Priority     requests.Integer `position:"Query" name:"Priority"`
	Ttl          requests.Integer `position:"Query" name:"Ttl"`
	UserClientIp string           `position:"Query" name:"UserClientIp"`
	ZoneId       string           `position:"Query" name:"ZoneId"`
}

// AddZoneRecordResponse is the response struct for api AddZoneRecord
type AddZoneRecordResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	RecordId  int64  `json:"RecordId" xml:"RecordId"`
	Success   bool   `json:"Success" xml:"Success"`
}

// CreateAddZoneRecordRequest creates a request to invoke AddZoneRecord API
func CreateAddZoneRecordRequest() (request *AddZoneRecordRequest) {
	request = &AddZoneRecordRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("pvtz", "2018-01-01", "AddZoneRecord", "pvtz", "openAPI")
	request.Method = requests.POST
	return
}

// CreateAddZoneRecordResponse creates a response to parse from AddZoneRecord response
func CreateAddZoneRecordResponse() (response *AddZoneRecordResponse) {
	response = &AddZoneRecordResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
