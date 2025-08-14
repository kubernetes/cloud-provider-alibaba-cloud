// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListListenerCertificatesResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetCertificateIds(v []*string) *ListListenerCertificatesResponseBody
	GetCertificateIds() []*string
	SetCertificates(v []*ListListenerCertificatesResponseBodyCertificates) *ListListenerCertificatesResponseBody
	GetCertificates() []*ListListenerCertificatesResponseBodyCertificates
	SetMaxResults(v int32) *ListListenerCertificatesResponseBody
	GetMaxResults() *int32
	SetNextToken(v string) *ListListenerCertificatesResponseBody
	GetNextToken() *string
	SetRequestId(v string) *ListListenerCertificatesResponseBody
	GetRequestId() *string
	SetTotalCount(v int32) *ListListenerCertificatesResponseBody
	GetTotalCount() *int32
}

type ListListenerCertificatesResponseBody struct {
	// The server certificates.
	CertificateIds []*string `json:"CertificateIds,omitempty" xml:"CertificateIds,omitempty" type:"Repeated"`
	// The certificates.
	Certificates []*ListListenerCertificatesResponseBodyCertificates `json:"Certificates,omitempty" xml:"Certificates,omitempty" type:"Repeated"`
	// The number of entries returned per page. Valid values: **1*	- to **50**. Default value: **20**.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The returned value of NextToken is a pagination token, which can be used in the next request to retrieve a new page of results. Valid values:
	//
	// 	- You do not need to specify this parameter for the first request.
	//
	// 	- You must specify the token that is obtained from the previous query as the value of NextToken.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The request ID.
	//
	// example:
	//
	// 2198BD6D-9EBB-5E1C-9C48-E0ABB79CF831
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The total number of entries returned.
	//
	// example:
	//
	// 1
	TotalCount *int32 `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListListenerCertificatesResponseBody) String() string {
	return dara.Prettify(s)
}

func (s ListListenerCertificatesResponseBody) GoString() string {
	return s.String()
}

func (s *ListListenerCertificatesResponseBody) GetCertificateIds() []*string {
	return s.CertificateIds
}

func (s *ListListenerCertificatesResponseBody) GetCertificates() []*ListListenerCertificatesResponseBodyCertificates {
	return s.Certificates
}

func (s *ListListenerCertificatesResponseBody) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListListenerCertificatesResponseBody) GetNextToken() *string {
	return s.NextToken
}

func (s *ListListenerCertificatesResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *ListListenerCertificatesResponseBody) GetTotalCount() *int32 {
	return s.TotalCount
}

func (s *ListListenerCertificatesResponseBody) SetCertificateIds(v []*string) *ListListenerCertificatesResponseBody {
	s.CertificateIds = v
	return s
}

func (s *ListListenerCertificatesResponseBody) SetCertificates(v []*ListListenerCertificatesResponseBodyCertificates) *ListListenerCertificatesResponseBody {
	s.Certificates = v
	return s
}

func (s *ListListenerCertificatesResponseBody) SetMaxResults(v int32) *ListListenerCertificatesResponseBody {
	s.MaxResults = &v
	return s
}

func (s *ListListenerCertificatesResponseBody) SetNextToken(v string) *ListListenerCertificatesResponseBody {
	s.NextToken = &v
	return s
}

func (s *ListListenerCertificatesResponseBody) SetRequestId(v string) *ListListenerCertificatesResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListListenerCertificatesResponseBody) SetTotalCount(v int32) *ListListenerCertificatesResponseBody {
	s.TotalCount = &v
	return s
}

func (s *ListListenerCertificatesResponseBody) Validate() error {
	return dara.Validate(s)
}

type ListListenerCertificatesResponseBodyCertificates struct {
	// The ID of the certificate. Only one server certificate is supported.
	//
	// example:
	//
	// 12315790343_166f8204689_1714763408_70998****
	CertificateId *string `json:"CertificateId,omitempty" xml:"CertificateId,omitempty"`
	// The type of the certificate.
	//
	// example:
	//
	// Server
	CertificateType *string `json:"CertificateType,omitempty" xml:"CertificateType,omitempty"`
	// Indicates whether the certificate is the default certificate of the listener. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// true
	IsDefault *bool `json:"IsDefault,omitempty" xml:"IsDefault,omitempty"`
	// Indicates whether the certificate is associated with the listener. Valid values:
	//
	// 	- **Associating**
	//
	// 	- **Associated**
	//
	// 	- **Diassociating**
	//
	// example:
	//
	// Associating
	Status *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s ListListenerCertificatesResponseBodyCertificates) String() string {
	return dara.Prettify(s)
}

func (s ListListenerCertificatesResponseBodyCertificates) GoString() string {
	return s.String()
}

func (s *ListListenerCertificatesResponseBodyCertificates) GetCertificateId() *string {
	return s.CertificateId
}

func (s *ListListenerCertificatesResponseBodyCertificates) GetCertificateType() *string {
	return s.CertificateType
}

func (s *ListListenerCertificatesResponseBodyCertificates) GetIsDefault() *bool {
	return s.IsDefault
}

func (s *ListListenerCertificatesResponseBodyCertificates) GetStatus() *string {
	return s.Status
}

func (s *ListListenerCertificatesResponseBodyCertificates) SetCertificateId(v string) *ListListenerCertificatesResponseBodyCertificates {
	s.CertificateId = &v
	return s
}

func (s *ListListenerCertificatesResponseBodyCertificates) SetCertificateType(v string) *ListListenerCertificatesResponseBodyCertificates {
	s.CertificateType = &v
	return s
}

func (s *ListListenerCertificatesResponseBodyCertificates) SetIsDefault(v bool) *ListListenerCertificatesResponseBodyCertificates {
	s.IsDefault = &v
	return s
}

func (s *ListListenerCertificatesResponseBodyCertificates) SetStatus(v string) *ListListenerCertificatesResponseBodyCertificates {
	s.Status = &v
	return s
}

func (s *ListListenerCertificatesResponseBodyCertificates) Validate() error {
	return dara.Validate(s)
}
