package prvd

import (
	"context"
	"sort"
	"strings"
)

const (
	RecordTypeA     = "A"
	RecordTypeAAAA  = "AAAA"
	RecordTypeCNAME = "CNAME"
	RecordTypeTXT   = "TXT"
	RecordTypePTR   = "PTR"
	RecordTypeSRV   = "SRV"
)

type PrivateZone interface {
	ListPVTZ(ctx context.Context) ([]*PvtzEndpoint, error)
	SearchPVTZ(ctx context.Context, ep *PvtzEndpoint, exact bool) ([]*PvtzEndpoint, error)
	UpdatePVTZ(ctx context.Context, ep *PvtzEndpoint) error
	DeletePVTZ(ctx context.Context, ep *PvtzEndpoint) error
}

type PvtzValue struct {
	Data     string
	RecordId int64
}

type PvtzEndpoint struct {
	Rr     string      `json:"Rr,omitempty"`
	Values []PvtzValue `json:"values,omitempty"`
	Type   string      `json:"recordType,omitempty"`
	Ttl    int64       `json:"recordTTL,omitempty"`
}

func (e *PvtzEndpoint) ValueString() string {
	vals := make([]string, 0)
	for _, val := range e.Values {
		vals = append(vals, val.Data)
	}
	sort.Strings(vals)
	return strings.Join(vals, ",")
}

func (e *PvtzEndpoint) ValueEqual(ep *PvtzEndpoint) bool {
	if e.Rr != ep.Rr {
		return false
	}
	if e.Type != ep.Type {
		return false
	}
	if e.ValueString() != ep.ValueString() {
		return false
	}
	return true
}

func (v *PvtzValue) InVals(vals []PvtzValue) bool {
	var found bool
	for _, val := range vals {
		if v.Data == val.Data {
			found = true
			break
		}
	}
	return found
}

type PvtzEndpointBuilder struct {
	PvtzEndpoint
}

func NewPvtzEndpointBuilder() PvtzEndpointBuilder {
	return PvtzEndpointBuilder{
		PvtzEndpoint{
			Values: make([]PvtzValue, 0),
		},
	}
}

func (peb *PvtzEndpointBuilder) WithValueData(data string) {
	peb.Values = append(peb.Values, PvtzValue{
		Data: data,
	})
}

func (peb *PvtzEndpointBuilder) WithRr(rr string) {
	peb.Rr = rr
}

func (peb *PvtzEndpointBuilder) WithType(recordType string) {
	peb.Type = recordType
}

func (peb *PvtzEndpointBuilder) DeepCopy() *PvtzEndpointBuilder {
	return &PvtzEndpointBuilder{
		PvtzEndpoint: peb.PvtzEndpoint,
	}
}

func (peb *PvtzEndpointBuilder) WithTtl(ttl int64) {
	peb.Ttl = ttl
}

func (peb *PvtzEndpointBuilder) Build() *PvtzEndpoint {
	ret := &peb.PvtzEndpoint
	ret.Values = peb.unifyValues(peb.Values)
	return ret
}

func (peb *PvtzEndpointBuilder) unifyValues(vals []PvtzValue) []PvtzValue {
	valMap := make(map[string]PvtzValue)
	for _, val := range vals {
		valMap[val.Data] = val
	}
	retVals := make([]PvtzValue, 0)
	for _, val := range valMap {
		retVals = append(retVals, val)
	}
	return retVals
}
