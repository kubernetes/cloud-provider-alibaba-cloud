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
	CreatePVTZ()
	Records(ctx context.Context) ([]*PvtzEndpoint, error)
	Create(ctx context.Context, ep *PvtzEndpoint) error
	Update(ctx context.Context, ep *PvtzEndpoint) error
	Delete(ctx context.Context, ep *PvtzEndpoint) error
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
