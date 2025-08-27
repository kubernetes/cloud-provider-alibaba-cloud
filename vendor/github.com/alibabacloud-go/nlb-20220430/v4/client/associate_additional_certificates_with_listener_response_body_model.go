// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iAssociateAdditionalCertificatesWithListenerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *AssociateAdditionalCertificatesWithListenerResponseBody
	GetJobId() *string
	SetRequestId(v string) *AssociateAdditionalCertificatesWithListenerResponseBody
	GetRequestId() *string
}

type AssociateAdditionalCertificatesWithListenerResponseBody struct {
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
	// 365F4154-92F6-4AE4-93F8-7FF34B540710
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s AssociateAdditionalCertificatesWithListenerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s AssociateAdditionalCertificatesWithListenerResponseBody) GoString() string {
	return s.String()
}

func (s *AssociateAdditionalCertificatesWithListenerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *AssociateAdditionalCertificatesWithListenerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *AssociateAdditionalCertificatesWithListenerResponseBody) SetJobId(v string) *AssociateAdditionalCertificatesWithListenerResponseBody {
	s.JobId = &v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerResponseBody) SetRequestId(v string) *AssociateAdditionalCertificatesWithListenerResponseBody {
	s.RequestId = &v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerResponseBody) Validate() error {
	return dara.Validate(s)
}
