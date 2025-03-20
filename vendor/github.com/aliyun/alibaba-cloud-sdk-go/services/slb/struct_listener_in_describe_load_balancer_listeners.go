package slb

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

// ListenerInDescribeLoadBalancerListeners is a nested struct in slb response
type ListenerInDescribeLoadBalancerListeners struct {
	AclType             string              `json:"AclType" xml:"AclType"`
	Status              string              `json:"Status" xml:"Status"`
	VServerGroupId      string              `json:"VServerGroupId" xml:"VServerGroupId"`
	ListenerProtocol    string              `json:"ListenerProtocol" xml:"ListenerProtocol"`
	LoadBalancerId      string              `json:"LoadBalancerId" xml:"LoadBalancerId"`
	ListenerPort        int                 `json:"ListenerPort" xml:"ListenerPort"`
	ServiceManagedMode  string              `json:"ServiceManagedMode" xml:"ServiceManagedMode"`
	AclId               string              `json:"AclId" xml:"AclId"`
	Scheduler           string              `json:"Scheduler" xml:"Scheduler"`
	Bandwidth           int                 `json:"Bandwidth" xml:"Bandwidth"`
	Description         string              `json:"Description" xml:"Description"`
	AclStatus           string              `json:"AclStatus" xml:"AclStatus"`
	BackendServerPort   int                 `json:"BackendServerPort" xml:"BackendServerPort"`
	BackendProtocol     string              `json:"BackendProtocol" xml:"BackendProtocol"`
	AclIds              []string            `json:"AclIds" xml:"AclIds"`
	HTTPListenerConfig  HTTPListenerConfig  `json:"HTTPListenerConfig" xml:"HTTPListenerConfig"`
	HTTPSListenerConfig HTTPSListenerConfig `json:"HTTPSListenerConfig" xml:"HTTPSListenerConfig"`
	TCPListenerConfig   TCPListenerConfig   `json:"TCPListenerConfig" xml:"TCPListenerConfig"`
	TCPSListenerConfig  TCPSListenerConfig  `json:"TCPSListenerConfig" xml:"TCPSListenerConfig"`
	UDPListenerConfig   UDPListenerConfig   `json:"UDPListenerConfig" xml:"UDPListenerConfig"`
	Tags                []Tag               `json:"Tags" xml:"Tags"`
}
