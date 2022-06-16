package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
)

const (
	ENDPOINT = "http://100.100.100.200"

	META_VERSION_LATEST = "latest"

	RS_TYPE_META_DATA = "meta-data"
	RS_TYPE_USER_DATA = "user-data"

	DNS_NAMESERVERS    = "dns-conf/nameservers"
	EIPV4              = "eipv4"
	HOSTNAME           = "hostname"
	IMAGE_ID           = "image-id"
	INSTANCE_ID        = "instance-id"
	MAC                = "mac"
	NETWORK_TYPE       = "network-type"
	NTP_CONF_SERVERS   = "ntp-conf/ntp-servers"
	OWNER_ACCOUNT_ID   = "owner-account-id"
	PRIVATE_IPV4       = "private-ipv4"
	REGION             = "region-id"
	SERIAL_NUMBER      = "serial-number"
	SOURCE_ADDRESS     = "source-address"
	VPC_CIDR_BLOCK     = "vpc-cidr-block"
	VPC_ID             = "vpc-id"
	VSWITCH_CIDR_BLOCK = "vswitch-cidr-block"
	VSWITCH_ID         = "vswitch-id"
	ZONE               = "zone-id"
	RAM_SECURITY       = "ram/security-credentials"
)

type IMetaDataRequest interface {
	Version(version string) IMetaDataRequest
	ResourceType(rtype string) IMetaDataRequest
	Resource(resource string) IMetaDataRequest
	SubResource(sub string) IMetaDataRequest
	Url() (string, error)
	Do(api interface{}) error
}

// NewMetaData return new metadata
func NewMetaData() prvd.IMetaData {
	if ctrlCfg.CloudCFG.Global.VpcID != "" &&
		ctrlCfg.CloudCFG.Global.VswitchID != "" {
		klog.Infof("read metadata from cfg")
		return &CfgMetaData{base: NewBaseMetaData(nil)}
	}
	return NewBaseMetaData(nil)
}

var _ prvd.IMetaData = &BaseMetaData{}

type BaseMetaData struct {
	// mock for unit test.
	mock requestMock

	client *http.Client
}

func NewBaseMetaData(client *http.Client) *BaseMetaData {
	if client == nil {
		client = &http.Client{}
	}
	return &BaseMetaData{
		client: client,
	}
}

func (m *BaseMetaData) New() *MetaDataRequest {
	return &MetaDataRequest{
		client:      m.client,
		sendRequest: m.mock,
	}
}

func (m *BaseMetaData) HostName() (string, error) {
	var hostname ResultList
	err := m.New().Resource(HOSTNAME).Do(&hostname)
	if err != nil {
		return "", err
	}
	return hostname.result[0], nil
}

func (m *BaseMetaData) ImageID() (string, error) {
	var image ResultList
	err := m.New().Resource(IMAGE_ID).Do(&image)
	if err != nil {
		return "", err
	}
	return image.result[0], err
}

func (m *BaseMetaData) InstanceID() (string, error) {
	var instanceid ResultList
	err := m.New().Resource(INSTANCE_ID).Do(&instanceid)
	if err != nil {
		return "", err
	}
	return instanceid.result[0], err
}

func (m *BaseMetaData) Mac() (string, error) {
	var mac ResultList
	err := m.New().Resource(MAC).Do(&mac)
	if err != nil {
		return "", err
	}
	return mac.result[0], nil
}

func (m *BaseMetaData) NetworkType() (string, error) {
	var network ResultList
	err := m.New().Resource(NETWORK_TYPE).Do(&network)
	if err != nil {
		return "", err
	}
	return network.result[0], nil
}

func (m *BaseMetaData) OwnerAccountID() (string, error) {
	var owner ResultList
	err := m.New().Resource(OWNER_ACCOUNT_ID).Do(&owner)
	if err != nil {
		return "", err
	}
	return owner.result[0], nil
}

func (m *BaseMetaData) PrivateIPv4() (string, error) {
	var private ResultList
	err := m.New().Resource(PRIVATE_IPV4).Do(&private)
	if err != nil {
		return "", err
	}
	return private.result[0], nil
}

func (m *BaseMetaData) Region() (string, error) {
	var region ResultList
	err := m.New().Resource(REGION).Do(&region)
	if err != nil {
		return "", err
	}
	return region.result[0], nil
}

func (m *BaseMetaData) SerialNumber() (string, error) {
	var serial ResultList
	err := m.New().Resource(SERIAL_NUMBER).Do(&serial)
	if err != nil {
		return "", err
	}
	return serial.result[0], nil
}

func (m *BaseMetaData) SourceAddress() (string, error) {
	var source ResultList
	err := m.New().Resource(SOURCE_ADDRESS).Do(&source)
	if err != nil {
		return "", err
	}
	return source.result[0], nil

}

func (m *BaseMetaData) VpcCIDRBlock() (string, error) {
	var vpcCIDR ResultList
	err := m.New().Resource(VPC_CIDR_BLOCK).Do(&vpcCIDR)
	if err != nil {
		return "", err
	}
	return vpcCIDR.result[0], err
}

func (m *BaseMetaData) VpcID() (string, error) {
	var vpcId ResultList
	err := m.New().Resource(VPC_ID).Do(&vpcId)
	if err != nil {
		return "", err
	}
	return vpcId.result[0], err
}

func (m *BaseMetaData) VswitchCIDRBlock() (string, error) {
	var cidr ResultList
	err := m.New().Resource(VSWITCH_CIDR_BLOCK).Do(&cidr)
	if err != nil {
		return "", err
	}
	return cidr.result[0], err
}

func (m *BaseMetaData) VswitchID() (string, error) {
	var vswithcid ResultList
	err := m.New().Resource(VSWITCH_ID).Do(&vswithcid)
	if err != nil {
		return "", err
	}
	return vswithcid.result[0], err
}

func (m *BaseMetaData) EIPv4() (string, error) {
	var eip ResultList
	err := m.New().Resource(EIPV4).Do(&eip)
	if err != nil {
		return "", err
	}
	return eip.result[0], nil
}

func (m *BaseMetaData) DNSNameServers() ([]string, error) {
	var data ResultList
	err := m.New().Resource(DNS_NAMESERVERS).Do(&data)
	if err != nil {
		return []string{}, err
	}
	return data.result, nil
}

func (m *BaseMetaData) NTPConfigServers() ([]string, error) {
	var data ResultList
	err := m.New().Resource(NTP_CONF_SERVERS).Do(&data)
	if err != nil {
		return []string{}, err
	}
	return data.result, nil
}

func (m *BaseMetaData) Zone() (string, error) {
	var zone ResultList
	err := m.New().Resource(ZONE).Do(&zone)
	if err != nil {
		return "", err
	}
	return zone.result[0], nil
}

func (m *BaseMetaData) RoleName() (string, error) {
	var roleName ResultList
	err := m.New().Resource("ram/security-credentials/").Do(&roleName)
	if err != nil {
		return "", err
	}
	return roleName.result[0], nil
}

func (m *BaseMetaData) RamRoleToken(role string) (prvd.RoleAuth, error) {
	var roleauth prvd.RoleAuth
	err := m.New().Resource(RAM_SECURITY).SubResource(role).Do(&roleauth)
	if err != nil {
		return prvd.RoleAuth{}, err
	}
	return roleauth, nil
}

func (m *BaseMetaData) ClusterID() string {
	return ctrlCfg.CloudCFG.Global.ClusterID
}

type requestMock func(resource string) (string, error)

type MetaDataRequest struct {
	version      string
	resourceType string
	resource     string
	subResource  string
	client       *http.Client

	sendRequest requestMock
}

func (vpc *MetaDataRequest) Version(version string) IMetaDataRequest {
	vpc.version = version
	return vpc
}

func (vpc *MetaDataRequest) ResourceType(rtype string) IMetaDataRequest {
	vpc.resourceType = rtype
	return vpc
}

func (vpc *MetaDataRequest) Resource(resource string) IMetaDataRequest {
	vpc.resource = resource
	return vpc
}

func (vpc *MetaDataRequest) SubResource(sub string) IMetaDataRequest {
	vpc.subResource = sub
	return vpc
}

var retry = util.AttemptStrategy{
	Min:   5,
	Total: 5 * time.Second,
	Delay: 200 * time.Millisecond,
}

func (vpc *MetaDataRequest) Url() (string, error) {
	if vpc.version == "" {
		vpc.version = "latest"
	}
	if vpc.resourceType == "" {
		vpc.resourceType = "meta-data"
	}
	if vpc.resource == "" {
		return "", errors.New("the resource you want to visit must not be nil!")
	}
	endpoint := os.Getenv("METADATA_ENDPOINT")
	if endpoint == "" {
		endpoint = ENDPOINT
	}
	r := fmt.Sprintf("%s/%s/%s/%s", endpoint, vpc.version, vpc.resourceType, vpc.resource)
	if vpc.subResource == "" {
		return r, nil
	}
	return fmt.Sprintf("%s/%s", r, vpc.subResource), nil
}

func (vpc *MetaDataRequest) Do(api interface{}) (err error) {
	var res = ""
	for r := retry.Start(); r.Next(); {
		if vpc.sendRequest != nil {
			res, err = vpc.sendRequest(vpc.resource)
		} else {
			res, err = vpc.send()
		}
		if !shouldRetry(err) {
			break
		}
	}
	if err != nil {
		return err
	}
	return vpc.Decode(res, api)
}

func (vpc *MetaDataRequest) Decode(data string, api interface{}) error {
	if data == "" {
		url, _ := vpc.Url()
		return fmt.Errorf("metadata: alivpc decode data must not be nil. url=[%s]\n", url)
	}
	switch api := api.(type) {
	case *ResultList:
		api.result = strings.Split(data, "\n")
		return nil
	case *prvd.RoleAuth:
		return json.Unmarshal([]byte(data), api)
	default:
		return fmt.Errorf("metadata: unknow type to decode, type=%s\n", reflect.TypeOf(api))
	}
}

func (vpc *MetaDataRequest) send() (string, error) {
	url, err := vpc.Url()
	if err != nil {
		return "", err
	}
	requ, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return "", err
	}
	resp, err := vpc.client.Do(requ)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Aliyun Metadata API Error: Status Code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type TimeoutError interface {
	error
	Timeout() bool // Is the error a timeout?
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(TimeoutError)
	if ok {
		return true
	}

	switch err {
	case io.ErrUnexpectedEOF, io.EOF:
		return true
	}
	switch e := err.(type) {
	case *net.DNSError:
		return true
	case *net.OpError:
		switch e.Op {
		case "read", "write":
			return true
		}
	case *url.Error:
		// url.Error can be returned either by net/url if a URL cannot be
		// parsed, or by net/http if the response is closed before the headers
		// are received or parsed correctly. In that later case, e.Op is set to
		// the HTTP method name with the first letter uppercased. We don't want
		// to retry on POST operations, since those are not idempotent, all the
		// other ones should be safe to retry.
		switch e.Op {
		case "Get", "Put", "Delete", "Head":
			return shouldRetry(e.Err)
		default:
			return false
		}
	}
	return false
}

type ResultList struct {
	result []string
}

// render meta data from cloud config file
var _ prvd.IMetaData = &CfgMetaData{}

type CfgMetaData struct {
	base prvd.IMetaData
}

func (m *CfgMetaData) HostName() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) ImageID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) InstanceID() (string, error) {

	return "fakedInstanceid", nil
}

func (m *CfgMetaData) Mac() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) NetworkType() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) OwnerAccountID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) PrivateIPv4() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) Region() (string, error) {
	if ctrlCfg.CloudCFG.Global.Region != "" {
		return ctrlCfg.CloudCFG.Global.Region, nil
	}
	return m.base.Region()
}

func (m *CfgMetaData) SerialNumber() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) SourceAddress() (string, error) {

	return "", fmt.Errorf("unimplemented")

}

func (m *CfgMetaData) VpcCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) VpcID() (string, error) {
	if ctrlCfg.CloudCFG.Global.VpcID != "" {
		return ctrlCfg.CloudCFG.Global.VpcID, nil
	}
	return m.base.VpcID()
}

func (m *CfgMetaData) VswitchCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

// zone1:vswitchid1,zone2:vswitch2
func (m *CfgMetaData) VswitchID() (string, error) {
	if ctrlCfg.CloudCFG.Global.VswitchID == "" {
		// get vswitch id from meta server
		return m.base.VswitchID()
	}

	vsws := strings.Split(ctrlCfg.CloudCFG.Global.VswitchID, ",")
	if len(vsws) == 0 {
		klog.Infof("parse vswitchid error, use vsw %s as default", ctrlCfg.CloudCFG.Global.VswitchID)
		return ctrlCfg.CloudCFG.Global.VswitchID, nil
	}

	attr := strings.Split(vsws[0], ":")
	if len(attr) == 2 {
		klog.Infof("use vsw %s as default from %s", attr[1], ctrlCfg.CloudCFG.Global.VswitchID)
		return attr[1], nil
	}

	klog.Infof("use vsw %s as default", ctrlCfg.CloudCFG.Global.VswitchID)
	return ctrlCfg.CloudCFG.Global.VswitchID, nil
}

func (m *CfgMetaData) EIPv4() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) DNSNameServers() ([]string, error) {

	return []string{""}, fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) NTPConfigServers() ([]string, error) {

	return []string{""}, fmt.Errorf("unimplemented")
}

func (m *CfgMetaData) Zone() (string, error) {
	if ctrlCfg.CloudCFG.Global.ZoneID != "" {
		return ctrlCfg.CloudCFG.Global.ZoneID, nil
	}
	return m.base.Zone()
}

func (m *CfgMetaData) RoleName() (string, error) {

	return m.base.RoleName()
}

func (m *CfgMetaData) RamRoleToken(role string) (prvd.RoleAuth, error) {

	return m.base.RamRoleToken(role)
}

func (m *CfgMetaData) ClusterID() string {
	return ctrlCfg.CloudCFG.Global.ClusterID
}
