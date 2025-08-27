// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDisassociateAdditionalCertificatesWithListenerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *DisassociateAdditionalCertificatesWithListenerResponseBody
	GetJobId() *string
	SetRequestId(v string) *DisassociateAdditionalCertificatesWithListenerResponseBody
	GetRequestId() *string
}

type DisassociateAdditionalCertificatesWithListenerResponseBody struct {
	// The ID of the asynchronous task.
	//
	// example:
	//
	// 72dcd26b-f12d-4c27-b3af-18f6aed5****
	JobId *string `json:"JobId,omitempty" xml:"JobId,omitempty"`
	// The request ID.
	//
	// example:
	//
	// 365F4154-92F6-4AE4-92F8-7FF34B540710
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DisassociateAdditionalCertificatesWithListenerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DisassociateAdditionalCertificatesWithListenerResponseBody) GoString() string {
	return s.String()
}

func (s *DisassociateAdditionalCertificatesWithListenerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *DisassociateAdditionalCertificatesWithListenerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DisassociateAdditionalCertificatesWithListenerResponseBody) SetJobId(v string) *DisassociateAdditionalCertificatesWithListenerResponseBody {
	s.JobId = &v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerResponseBody) SetRequestId(v string) *DisassociateAdditionalCertificatesWithListenerResponseBody {
	s.RequestId = &v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerResponseBody) Validate() error {
	return dara.Validate(s)
}
