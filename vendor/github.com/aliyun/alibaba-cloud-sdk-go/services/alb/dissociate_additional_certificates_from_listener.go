package alb

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

// DissociateAdditionalCertificatesFromListener invokes the alb.DissociateAdditionalCertificatesFromListener API synchronously
func (client *Client) DissociateAdditionalCertificatesFromListener(request *DissociateAdditionalCertificatesFromListenerRequest) (response *DissociateAdditionalCertificatesFromListenerResponse, err error) {
	response = CreateDissociateAdditionalCertificatesFromListenerResponse()
	err = client.DoAction(request, response)
	return
}

// DissociateAdditionalCertificatesFromListenerWithChan invokes the alb.DissociateAdditionalCertificatesFromListener API asynchronously
func (client *Client) DissociateAdditionalCertificatesFromListenerWithChan(request *DissociateAdditionalCertificatesFromListenerRequest) (<-chan *DissociateAdditionalCertificatesFromListenerResponse, <-chan error) {
	responseChan := make(chan *DissociateAdditionalCertificatesFromListenerResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DissociateAdditionalCertificatesFromListener(request)
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

// DissociateAdditionalCertificatesFromListenerWithCallback invokes the alb.DissociateAdditionalCertificatesFromListener API asynchronously
func (client *Client) DissociateAdditionalCertificatesFromListenerWithCallback(request *DissociateAdditionalCertificatesFromListenerRequest, callback func(response *DissociateAdditionalCertificatesFromListenerResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DissociateAdditionalCertificatesFromListenerResponse
		var err error
		defer close(result)
		response, err = client.DissociateAdditionalCertificatesFromListener(request)
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

// DissociateAdditionalCertificatesFromListenerRequest is the request struct for api DissociateAdditionalCertificatesFromListener
type DissociateAdditionalCertificatesFromListenerRequest struct {
	*requests.RpcRequest
	ClientToken  string                                                      `position:"Query" name:"ClientToken"`
	ListenerId   string                                                      `position:"Query" name:"ListenerId"`
	DryRun       requests.Boolean                                            `position:"Query" name:"DryRun"`
	Certificates *[]DissociateAdditionalCertificatesFromListenerCertificates `position:"Query" name:"Certificates"  type:"Repeated"`
}

// DissociateAdditionalCertificatesFromListenerCertificates is a repeated param struct in DissociateAdditionalCertificatesFromListenerRequest
type DissociateAdditionalCertificatesFromListenerCertificates struct {
	CertificateId string `name:"CertificateId"`
}

// DissociateAdditionalCertificatesFromListenerResponse is the response struct for api DissociateAdditionalCertificatesFromListener
type DissociateAdditionalCertificatesFromListenerResponse struct {
	*responses.BaseResponse
	JobId     string `json:"JobId" xml:"JobId"`
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateDissociateAdditionalCertificatesFromListenerRequest creates a request to invoke DissociateAdditionalCertificatesFromListener API
func CreateDissociateAdditionalCertificatesFromListenerRequest() (request *DissociateAdditionalCertificatesFromListenerRequest) {
	request = &DissociateAdditionalCertificatesFromListenerRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Alb", "2020-06-16", "DissociateAdditionalCertificatesFromListener", "alb", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDissociateAdditionalCertificatesFromListenerResponse creates a response to parse from DissociateAdditionalCertificatesFromListener response
func CreateDissociateAdditionalCertificatesFromListenerResponse() (response *DissociateAdditionalCertificatesFromListenerResponse) {
	response = &DissociateAdditionalCertificatesFromListenerResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}