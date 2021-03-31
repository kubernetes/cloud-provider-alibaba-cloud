package provider

import "context"

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

func (v *PvtzValue) InVals(vals []PvtzValue) bool {
	var found bool
	for _, v := range vals {
		if v.Data == v.Data {
			found = true
			break
		}
	}
	return found
}
