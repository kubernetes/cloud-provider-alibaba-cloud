package alibaba

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/pvtz"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	util_errors "k8s.io/apimachinery/pkg/util/errors"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

const (
	DescribeZoneRecordPageSize = 50
	// TODO add remark
	ZoneRecordRemark = "record.managed.by.ack.ccm"
)

type PVTZProvider struct {
	client *pvtz.Client
	zoneId string
}

func NewPVTZProvider(auth *ClientAuth) *PVTZProvider {
	return &PVTZProvider{
		client: auth.PVTZ,
		zoneId: ctx2.CFG.Global.PrivateZoneID,
	}
}
func (p *PVTZProvider) ListPVTZ(ctx context.Context) ([]*prvd.PvtzEndpoint, error) {
	return p.SearchPVTZ(ctx, &prvd.PvtzEndpoint{}, false)
}

func (p *PVTZProvider) SearchPVTZ(ctx context.Context, ep *prvd.PvtzEndpoint, exact bool) ([]*prvd.PvtzEndpoint, error) {
	req := pvtz.CreateDescribeZoneRecordsRequest()
	req.ZoneId = p.zoneId
	req.PageSize = requests.NewInteger(DescribeZoneRecordPageSize)
	if ep.Rr != "" {
		req.Keyword = ep.Rr
		if exact {
			req.SearchMode = "EXACT"
		} else {
			req.SearchMode = "LIKE"
		}
	}
	records := make([]pvtz.Record, 0)
	pageNumber := 1
	for {
		req.PageNumber = requests.NewInteger(pageNumber)
		resp, err := p.client.DescribeZoneRecords(req)
		if err != nil {
			return nil, err
		}
		for _, record := range resp.Records.Record {
			if p.filterUnsupportedDNSRecordTypes(record) {
				continue
			}
			if p.filterUnmanagedDNSRecord(record) {
				continue
			}
			records = append(records, record)
		}
		if pageNumber < resp.TotalPages {
			pageNumber++
		} else {
			break
		}
	}
	// transform raw zone records into endpoints
	typedEndpointsMap := make(map[string]map[string]*prvd.PvtzEndpoint)
	for _, record := range records {
		if endpointsMap := typedEndpointsMap[record.Type]; endpointsMap == nil {
			typedEndpointsMap[record.Type] = make(map[string]*prvd.PvtzEndpoint)
		}

		if rrMap := typedEndpointsMap[record.Type][record.Rr]; rrMap == nil {
			typedEndpointsMap[record.Type][record.Rr] = &prvd.PvtzEndpoint{
				Rr: record.Rr,
				Values: []prvd.PvtzValue{{
					Data:     record.Value,
					RecordId: record.RecordId,
				}},
				Ttl:  int64(record.Ttl),
				Type: record.Type,
			}
		} else {
			typedEndpointsMap[record.Type][record.Rr].Values = append(typedEndpointsMap[record.Type][record.Rr].Values, prvd.PvtzValue{
				Data:     record.Value,
				RecordId: record.RecordId,
			})
		}
	}
	totalEndpoints := make([]*prvd.PvtzEndpoint, 0)
	for _, endpointsMap := range typedEndpointsMap {
		for _, endpoint := range endpointsMap {
			totalEndpoints = append(totalEndpoints, endpoint)
		}
	}
	return totalEndpoints, nil
}

func (p *PVTZProvider) UpdatePVTZ(ctx context.Context, ep *prvd.PvtzEndpoint) error {
	rlog := log.WithFields(log.Fields{"endpointRr": ep.Rr, "endpointType": ep.Type})
	newValues := ep.Values
	oldValues := make([]prvd.PvtzValue, 0)
	err := p.record(context.TODO(), ep)
	if err != nil {
		return errors.Wrap(err, "UpdatePVTZ query old zone records error")
	} else {
		oldValues = ep.Values
	}
	rlog.Infof("old endpoints %v, new endpoints %v", oldValues, newValues)
	valueToAdd := make([]string, 0)
	for _, newVal := range newValues {
		if !newVal.InVals(oldValues) {
			valueToAdd = append(valueToAdd, newVal.Data)
		}
	}
	recordIdToDelete := make([]int64, 0)
	for _, oldVal := range oldValues {
		if !oldVal.InVals(newValues) {
			recordIdToDelete = append(recordIdToDelete, oldVal.RecordId)
		}
	}
	errs := make([]error, 0)
	for _, val := range valueToAdd {
		_, err = p.create(ep.Type, ep.Rr, val, int(ep.Ttl))
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, id := range recordIdToDelete {
		err = p.delete(id)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Wrap(util_errors.NewAggregate(errs), "UpdatePVTZ update zone records error")
}

func (p *PVTZProvider) DeletePVTZ(ctx context.Context, ep *prvd.PvtzEndpoint) error {
	err := p.record(context.TODO(), ep)
	if err != nil {
		return errors.Wrap(err, "DeletePVTZ query old zone records error")
	}
	oldValues := ep.Values
	errs := make([]error, 0)
	for _, val := range oldValues {
		err = p.delete(val.RecordId)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Wrap(util_errors.NewAggregate(errs), "DeletePVTZ deleting old endpoint error")
}

func (p *PVTZProvider) filterUnmanagedDNSRecord(record pvtz.Record) bool {
	if strings.Contains(record.Remark, ZoneRecordRemark) {
		return false
	}
	return true
}

func (p *PVTZProvider) filterUnsupportedDNSRecordTypes(record pvtz.Record) bool {
	switch record.Type {
	case prvd.RecordTypeA, prvd.RecordTypeCNAME, prvd.RecordTypePTR, prvd.RecordTypeSRV, prvd.RecordTypeTXT:
		return false
	default:
		return true
	}
}

func (p *PVTZProvider) record(ctx context.Context, ep *prvd.PvtzEndpoint) error {
	if ep.Rr == "" {
		return fmt.Errorf("endpoint %s %s not found", ep.Rr, ep.Type)
	}
	records, err := p.SearchPVTZ(ctx, ep, true)
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.Rr == ep.Rr &&
			(ep.Type == "" || record.Type == ep.Type) {
			ep.Values = record.Values
			ep.Ttl = record.Ttl
			return nil
		}
	}
	// not found, setting result ep to empty
	ep.Values = []prvd.PvtzValue{}
	return nil
}

func (p *PVTZProvider) create(recordType, rr, value string, ttl int) (*pvtz.AddZoneRecordResponse, error) {
	req := pvtz.CreateAddZoneRecordRequest()
	req.ZoneId = p.zoneId
	req.Type = recordType
	req.Rr = rr
	req.Ttl = requests.NewInteger(ttl)
	req.Remark = ZoneRecordRemark
	req.Value = value
	resp, err := p.client.AddZoneRecord(req)
	return resp, err
}

func (p *PVTZProvider) delete(recordId int64) error {
	req := pvtz.CreateDeleteZoneRecordRequest()
	req.RecordId = requests.NewInteger(int(recordId))
	_, err := p.client.DeleteZoneRecord(req)
	return err
}
