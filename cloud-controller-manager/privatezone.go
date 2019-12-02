package alicloud

import (
	"fmt"
	"github.com/denverdino/aliyungo/pvtz"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
)

// DEFAULT_LANG default lang
const DEFAULT_LANG = "en"

// ClientPVTZSDK private zone sdk interface
type ClientPVTZSDK interface {
	DescribeZones(args *pvtz.DescribeZonesArgs) (zones []pvtz.ZoneType, err error)
	AddZone(args *pvtz.AddZoneArgs) (response *pvtz.AddZoneResponse, err error)
	DeleteZone(args *pvtz.DeleteZoneArgs) (err error)
	CheckZoneName(args *pvtz.CheckZoneNameArgs) (bool, error)
	UpdateZoneRemark(args *pvtz.UpdateZoneRemarkArgs) error
	DescribeZoneInfo(args *pvtz.DescribeZoneInfoArgs) (response *pvtz.DescribeZoneInfoResponse, err error)
	BindZoneVpc(args *pvtz.BindZoneVpcArgs) (err error)
	DescribeRegions() (regions []pvtz.RegionType, err error)
	DescribeZoneRecords(args *pvtz.DescribeZoneRecordsArgs) (records []pvtz.ZoneRecordType, err error)
	DescribeZoneRecordsByRR(zoneId string, rr string) (records []pvtz.ZoneRecordType, err error)
	DeleteZoneRecordsByRR(zoneId string, rr string) error
	AddZoneRecord(args *pvtz.AddZoneRecordArgs) (response *pvtz.AddZoneRecordResponse, err error)
	UpdateZoneRecord(args *pvtz.UpdateZoneRecordArgs) (err error)
	DeleteZoneRecord(args *pvtz.DeleteZoneRecordArgs) (err error)
	SetZoneRecordStatus(args *pvtz.SetZoneRecordStatusArgs) (err error)
}

// PrivateZoneClient private zone client wrapper
type PrivateZoneClient struct {
	c ClientPVTZSDK
	// known service resource version
}

func (s *PrivateZoneClient) findPrivateZone(service *v1.Service) (bool, *pvtz.DescribeZoneInfoResponse, error) {
	def, _ := ExtractAnnotationRequest(service)

	// User assigned private zone id go first.
	if def.PrivateZoneId != "" {
		return s.findPrivateZoneById(def.PrivateZoneId)
	}

	// if not, find by private zone name
	if def.PrivateZoneName != "" {
		return s.findPrivateZoneByName(def.PrivateZoneName)
	}

	return false, nil, nil
}

func (s *PrivateZoneClient) findPrivateZoneById(id string) (bool, *pvtz.DescribeZoneInfoResponse, error) {
	zone, err := s.c.DescribeZoneInfo(
		&pvtz.DescribeZoneInfoArgs{
			Lang:   DEFAULT_LANG,
			ZoneId: id,
		},
	)
	if err != nil || zone == nil {
		return false, nil, err
	}

	return true, zone, nil
}

func (s *PrivateZoneClient) findPrivateZoneByName(name string) (bool, *pvtz.DescribeZoneInfoResponse, error) {
	zones, err := s.c.DescribeZones(
		&pvtz.DescribeZonesArgs{
			Lang:    DEFAULT_LANG,
			Keyword: name,
		},
	)
	if err != nil {
		return false, nil, err
	}

	if zones == nil || len(zones) == 0 {
		return false, nil, nil
	}

	var selectedZoneId string

	// recommend user to use the private zone that name can perfectly match to configured name
	// if we found one can't match to configured name perfectly, give user a warning
	if len(zones) > 1 {
		for _, zone := range zones {
			if zone.ZoneName == name {
				selectedZoneId = zone.ZoneId
				break
			}
		}

		if selectedZoneId == "" {
			glog.Warningf("multiple private zone returned with name [%s], "+
				"and we can't find one which matches to name perfectly,"+
				"using the first one with ID=%s", name, zones[0].ZoneId)
			selectedZoneId = zones[0].ZoneId
		}
	} else {
		if zones[0].ZoneName != name {
			glog.Warningf("just one private zone returned with name [%s], "+
				"but this private zone can't match to name perfectly,"+
				"found private zone ID=%s", name, zones[0].ZoneId)
		}
		selectedZoneId = zones[0].ZoneId
	}

	return s.findPrivateZoneById(selectedZoneId)
}

func (s *PrivateZoneClient) findRecordByRr(zone *pvtz.DescribeZoneInfoResponse, rr string) (*pvtz.ZoneRecordType, error) {
	records, err := s.c.DescribeZoneRecordsByRR(zone.ZoneId, rr)
	if err != nil {
		return nil, err
	}

	switch len(records) {
	case 0:
		return nil, nil
	case 1:
		return &records[0], nil
	default:
		return nil, fmt.Errorf("alicloud: multiple private zone record returned with rr [%s]", rr)
	}
}

func (s *PrivateZoneClient) findRecordByService(service *v1.Service) (*pvtz.DescribeZoneInfoResponse, *pvtz.ZoneRecordType, error) {
	_, request := ExtractAnnotationRequest(service)

	if request.PrivateZoneRecordName == "" {
		return nil, nil, nil
	}

	exists, zone, err := s.findPrivateZone(service)
	if err != nil {
		return nil, nil, err
	}

	if !exists {
		return nil, nil, err
	}

	record, err := s.findRecordByRr(zone, request.PrivateZoneRecordName)
	if err != nil {
		return nil, nil, err
	}

	if record == nil {
		return zone, nil, nil
	}

	return zone, record, nil
}

func (s *PrivateZoneClient) findExactRecordByService(service *v1.Service, ip string, ipVersion slb.AddressIPVersionType) (*pvtz.DescribeZoneInfoResponse, *pvtz.ZoneRecordType, bool, error) {
	zone, record, err := s.findRecordByService(service)
	if err != nil {
		return nil, nil, false, err
	}

	if record == nil {
		return nil, nil, false, nil
	}
	recordType := getRecordType(ipVersion)
	// check the ip is matched with ip address, if not it may be user managed
	if record.Type != recordType || record.Value != ip {
		return zone, record, false, nil
	}

	return zone, record, true, nil
}

func (s *PrivateZoneClient) updateRecordCache(service *v1.Service, zone *pvtz.DescribeZoneInfoResponse, record *pvtz.ZoneRecordType, err error) (*pvtz.DescribeZoneInfoResponse, *pvtz.ZoneRecordType, error) {
	if err != nil {
		return zone, record, err
	}

	kv := GetPrivateZoneRecordCache()

	var recordId int64 = -1
	previousId, found := kv.get(string(service.GetUID()))

	if record != nil {
		recordId = record.RecordId
	}

	// we will delete record which we created before
	if found && previousId != recordId {
		_ = s.c.DeleteZoneRecord(&pvtz.DeleteZoneRecordArgs{
			RecordId: previousId,
			Lang:     DEFAULT_LANG,
		})
	}

	// update new record id to cache or delete cache
	if record != nil {
		kv.set(string(service.GetUID()), recordId)
	} else {
		kv.remove(string(service.GetUID()))
	}

	return zone, record, err
}

// EnsurePrivateZoneRecord make sure private zone record is reconciled
func (s *PrivateZoneClient) EnsurePrivateZoneRecord(service *v1.Service, ip string, ipVersion slb.AddressIPVersionType) (zone *pvtz.DescribeZoneInfoResponse, record *pvtz.ZoneRecordType, err error) {
	glog.V(4).Infof("alicloud: ensure private zone record for ip(%s) with service details, \n%+v", ip, PrettyJson(service))

	// update record cache after ensure
	defer func() {
		zone, record, err = s.updateRecordCache(service, zone, record, err)
	}()

	recordType := getRecordType(ipVersion)
	_, request := ExtractAnnotationRequest(service)

	zone, record, err = s.findRecordByService(service)
	if err != nil {
		return nil, nil, err
	}

	// we will not create private zone, and user must to create it manually
	if zone == nil {
		utils.Logf(service, "config or private zone not found, "+
			"we will skip to configure private zone")
		return nil, nil, nil
	}

	utils.Logf(service, "find private zone with id %s", zone.ZoneId)

	if record == nil {
		utils.Logf(service, "create and bind new private zone record [%s.%s] to ip [%s]",
			request.PrivateZoneRecordName,
			zone.ZoneName,
			ip)

		_, err := s.c.AddZoneRecord(
			&pvtz.AddZoneRecordArgs{
				ZoneId: zone.ZoneId,
				Rr:     request.PrivateZoneRecordName,
				Type:   recordType,
				Value:  ip,
			})
		if err != nil {
			return nil, nil, err
		}

		// ensure the record has been created
		record, err = s.findRecordByRr(zone, request.PrivateZoneRecordName)
		if err != nil {
			return nil, nil, err
		}

		if record == nil {
			return nil, nil, fmt.Errorf("alicloud: unknown error on creating private zone record, it shouldn't be happened. ")
		}
	} else if record.Type != recordType || record.Value != ip {
		utils.Logf(service, "update private zone record [%s.%s] bind to ip [%s]",
			request.PrivateZoneRecordName,
			zone.ZoneName,
			ip)

		err = s.c.UpdateZoneRecord(
			&pvtz.UpdateZoneRecordArgs{
				RecordId: record.RecordId,
				Rr:       request.PrivateZoneRecordName,
				Type:     recordType,
				Value:    ip,
				Lang:     DEFAULT_LANG,
			})
		if err != nil {
			return nil, nil, err
		}
	}

	return zone, record, nil
}

// EnsurePrivateZoneRecordDeleted make sure private zone record is deleted.
func (s *PrivateZoneClient) EnsurePrivateZoneRecordDeleted(service *v1.Service, ip string, ipVersion slb.AddressIPVersionType) error {
	// need to save the resource version when deleted event
	err := keepResourceVersion(service)
	if err != nil {
		glog.Warningf("failed to save "+
			"deleted service resourceVersion, [%s] due to [%s] ", service.Name, err.Error())
	}

	zoneInfo, record, exactMatch, err := s.findExactRecordByService(service, ip, ipVersion)
	if err != nil {
		return err
	}

	// check the ip is matched with ip address, if not, it may be user managed record
	if !exactMatch {
		utils.Logf(service, "private zone record not created by cloudprovider, skip to delete it. "+
			"service [%s]", service.Name)
		return nil
	}

	if zoneInfo != nil && record != nil {
		utils.Logf(service, "private zone record deleted by cloudprovider. service [%s]", service.Name)
		return s.c.DeleteZoneRecordsByRR(zoneInfo.ZoneId, record.Rr)
	}

	return nil
}

func getHostName(pz *pvtz.DescribeZoneInfoResponse, pzr *pvtz.ZoneRecordType) string {
	var hostname string
	if pz != nil && pzr != nil {
		hostname = fmt.Sprintf("%s.%s", pzr.Rr, pz.ZoneName)
	}
	return hostname
}

func getRecordType(ipVersion slb.AddressIPVersionType) string {
	if ipVersion == slb.IPv6 {
		return "AAAA"
	}
	return "A"
}
