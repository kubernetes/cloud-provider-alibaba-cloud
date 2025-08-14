// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDescribeZonesResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetRequestId(v string) *DescribeZonesResponseBody
	GetRequestId() *string
	SetZones(v []*DescribeZonesResponseBodyZones) *DescribeZonesResponseBody
	GetZones() []*DescribeZonesResponseBodyZones
}

type DescribeZonesResponseBody struct {
	// The ID of the request.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// A list of zones.
	Zones []*DescribeZonesResponseBodyZones `json:"Zones,omitempty" xml:"Zones,omitempty" type:"Repeated"`
}

func (s DescribeZonesResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DescribeZonesResponseBody) GoString() string {
	return s.String()
}

func (s *DescribeZonesResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DescribeZonesResponseBody) GetZones() []*DescribeZonesResponseBodyZones {
	return s.Zones
}

func (s *DescribeZonesResponseBody) SetRequestId(v string) *DescribeZonesResponseBody {
	s.RequestId = &v
	return s
}

func (s *DescribeZonesResponseBody) SetZones(v []*DescribeZonesResponseBodyZones) *DescribeZonesResponseBody {
	s.Zones = v
	return s
}

func (s *DescribeZonesResponseBody) Validate() error {
	return dara.Validate(s)
}

type DescribeZonesResponseBodyZones struct {
	// The zone name.
	//
	// example:
	//
	// ap_southeast_2
	LocalName *string `json:"LocalName,omitempty" xml:"LocalName,omitempty"`
	// The zone ID.
	//
	// example:
	//
	// cn-hangzhou-g
	ZoneId *string `json:"ZoneId,omitempty" xml:"ZoneId,omitempty"`
}

func (s DescribeZonesResponseBodyZones) String() string {
	return dara.Prettify(s)
}

func (s DescribeZonesResponseBodyZones) GoString() string {
	return s.String()
}

func (s *DescribeZonesResponseBodyZones) GetLocalName() *string {
	return s.LocalName
}

func (s *DescribeZonesResponseBodyZones) GetZoneId() *string {
	return s.ZoneId
}

func (s *DescribeZonesResponseBodyZones) SetLocalName(v string) *DescribeZonesResponseBodyZones {
	s.LocalName = &v
	return s
}

func (s *DescribeZonesResponseBodyZones) SetZoneId(v string) *DescribeZonesResponseBodyZones {
	s.ZoneId = &v
	return s
}

func (s *DescribeZonesResponseBodyZones) Validate() error {
	return dara.Validate(s)
}
