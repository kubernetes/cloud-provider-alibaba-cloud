package model

// CertificateInfo is a nested struct in cas response
type CertificateInfo struct {
	CommonName      string `json:"CommonName" xml:"CommonName"`
	CertName        string `json:"CertName" xml:"CertName"`
	Issuer          string `json:"Issuer" xml:"Issuer"`
	Algorithm       string `json:"Algorithm" xml:"Algorithm"`
	CertIdentifier  string `json:"CertIdentifier" xml:"CertIdentifier"`
	KeySize         int    `json:"KeySize" xml:"KeySize"`
	BeforeDate      int64  `json:"BeforeDate" xml:"BeforeDate"`
	Sha2            string `json:"Sha2" xml:"Sha2"`
	SignAlgorithm   string `json:"SignAlgorithm" xml:"SignAlgorithm"`
	AfterDate       int64  `json:"AfterDate" xml:"AfterDate"`
	DomainMatchCert bool   `json:"DomainMatchCert" xml:"DomainMatchCert"`
	Md5             string `json:"Md5" xml:"Md5"`
	SerialNo        string `json:"SerialNo" xml:"SerialNo"`
	Sans            string `json:"Sans" xml:"Sans"`
}
