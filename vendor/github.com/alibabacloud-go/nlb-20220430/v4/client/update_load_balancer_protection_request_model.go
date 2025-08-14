// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerProtectionRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *UpdateLoadBalancerProtectionRequest
	GetClientToken() *string
	SetDeletionProtectionEnabled(v bool) *UpdateLoadBalancerProtectionRequest
	GetDeletionProtectionEnabled() *bool
	SetDeletionProtectionReason(v string) *UpdateLoadBalancerProtectionRequest
	GetDeletionProtectionReason() *string
	SetDryRun(v bool) *UpdateLoadBalancerProtectionRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *UpdateLoadBalancerProtectionRequest
	GetLoadBalancerId() *string
	SetModificationProtectionReason(v string) *UpdateLoadBalancerProtectionRequest
	GetModificationProtectionReason() *string
	SetModificationProtectionStatus(v string) *UpdateLoadBalancerProtectionRequest
	GetModificationProtectionStatus() *string
	SetRegionId(v string) *UpdateLoadBalancerProtectionRequest
	GetRegionId() *string
}

type UpdateLoadBalancerProtectionRequest struct {
	// The client token used to ensure the idempotence of the request.
	//
	// You can use the client to generate this value. Ensure that the value is unique among all requests. Only ASCII characters are allowed.
	//
	// >  If you do not specify this parameter, the value of **RequestId*	- is used.***	- **RequestId*	- of each request is different.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// Specifies whether to enable deletion protection. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	DeletionProtectionEnabled *bool `json:"DeletionProtectionEnabled,omitempty" xml:"DeletionProtectionEnabled,omitempty"`
	// The reason why deletion protection is enabled. The reason must be 2 to 128 characters in length, can contain letters, digits, periods (.), underscores (_), and hyphens (-), and must start with a letter.
	//
	// >  This parameter takes effect only when **DeletionProtectionEnabled*	- is set to **true**.
	//
	// example:
	//
	// Instance_Is_Bound_By_Acceleration_Instance
	DeletionProtectionReason *string `json:"DeletionProtectionReason,omitempty" xml:"DeletionProtectionReason,omitempty"`
	// Specifies whether to perform a dry run, without sending the actual request. Valid values:
	//
	// 	- **true**: performs only a dry run. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the dry run, an error message is returned. If the request passes the dry run, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): sends a request. If the request passes the dry run, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// false
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The ID of the NLB instance.
	//
	// This parameter is required.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The reason why the configuration read-only mode is enabled. The reason must be 2 to 128 characters in length, can contain letters, digits, periods (.), underscores (_), and hyphens (-), and must start with a letter.
	//
	// >  This parameter takes effect only when **Status*	- is set to **ConsoleProtection**.
	//
	// example:
	//
	// ConsoleProtection
	ModificationProtectionReason *string `json:"ModificationProtectionReason,omitempty" xml:"ModificationProtectionReason,omitempty"`
	// Specifies whether to enable the configuration read-only mode. Valid values:
	//
	// 	- **NonProtection**: disables the configuration read-only mode. In this case, you cannot set the **ModificationProtectionReason*	- parameter. If you specify **ModificationProtectionReason**, the value is cleared.
	//
	// 	- **ConsoleProtection**: enables the configuration read-only mode. In this case, you can specify **ModificationProtectionReason**.
	//
	// >  If you set this parameter to **ConsoleProtection**, you cannot use the NLB console to modify configurations of the NLB instance. However, you can call API operations to modify the instance configurations.
	//
	// example:
	//
	// ConsoleProtection
	ModificationProtectionStatus *string `json:"ModificationProtectionStatus,omitempty" xml:"ModificationProtectionStatus,omitempty"`
	// The region ID of the NLB instance.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
}

func (s UpdateLoadBalancerProtectionRequest) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerProtectionRequest) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerProtectionRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *UpdateLoadBalancerProtectionRequest) GetDeletionProtectionEnabled() *bool {
	return s.DeletionProtectionEnabled
}

func (s *UpdateLoadBalancerProtectionRequest) GetDeletionProtectionReason() *string {
	return s.DeletionProtectionReason
}

func (s *UpdateLoadBalancerProtectionRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *UpdateLoadBalancerProtectionRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *UpdateLoadBalancerProtectionRequest) GetModificationProtectionReason() *string {
	return s.ModificationProtectionReason
}

func (s *UpdateLoadBalancerProtectionRequest) GetModificationProtectionStatus() *string {
	return s.ModificationProtectionStatus
}

func (s *UpdateLoadBalancerProtectionRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *UpdateLoadBalancerProtectionRequest) SetClientToken(v string) *UpdateLoadBalancerProtectionRequest {
	s.ClientToken = &v
	return s
}

func (s *UpdateLoadBalancerProtectionRequest) SetDeletionProtectionEnabled(v bool) *UpdateLoadBalancerProtectionRequest {
	s.DeletionProtectionEnabled = &v
	return s
}

func (s *UpdateLoadBalancerProtectionRequest) SetDeletionProtectionReason(v string) *UpdateLoadBalancerProtectionRequest {
	s.DeletionProtectionReason = &v
	return s
}

func (s *UpdateLoadBalancerProtectionRequest) SetDryRun(v bool) *UpdateLoadBalancerProtectionRequest {
	s.DryRun = &v
	return s
}

func (s *UpdateLoadBalancerProtectionRequest) SetLoadBalancerId(v string) *UpdateLoadBalancerProtectionRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *UpdateLoadBalancerProtectionRequest) SetModificationProtectionReason(v string) *UpdateLoadBalancerProtectionRequest {
	s.ModificationProtectionReason = &v
	return s
}

func (s *UpdateLoadBalancerProtectionRequest) SetModificationProtectionStatus(v string) *UpdateLoadBalancerProtectionRequest {
	s.ModificationProtectionStatus = &v
	return s
}

func (s *UpdateLoadBalancerProtectionRequest) SetRegionId(v string) *UpdateLoadBalancerProtectionRequest {
	s.RegionId = &v
	return s
}

func (s *UpdateLoadBalancerProtectionRequest) Validate() error {
	return dara.Validate(s)
}
