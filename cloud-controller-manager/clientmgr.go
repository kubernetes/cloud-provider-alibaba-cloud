/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alicloud

import (
	"encoding/json"
	"k8s.io/klog"
	"path/filepath"
	"time"

	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/metadata"
	"github.com/denverdino/aliyungo/pvtz"
	"github.com/denverdino/aliyungo/slb"
	"github.com/go-cmd/cmd"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"
)

// ROLE_NAME default kubernetes master role name
var ROLE_NAME = "KubernetesMasterRole"

// ASSUME_ROLE_NAME managed kubernetes role name
var ASSUME_ROLE_NAME = "AliyunCSManagedKubernetesRole"

// TOKEN_RESYNC_PERIOD default token sync period
var TOKEN_RESYNC_PERIOD = 10 * time.Minute

// ClientMgr client manager for aliyun sdk
type ClientMgr struct {
	stop <-chan struct{}

	token TokenAuth

	meta         IMetaData
	routes       *RoutesClient
	loadbalancer *LoadBalancerClient
	privateZone  *PrivateZoneClient
	instance     *InstanceClient
}

// NewClientMgr return a new client manager
func NewClientMgr(key, secret string) (*ClientMgr, error) {
	m := NewMetaData()
	region, err := m.Region()
	if err != nil {
		return nil, fmt.Errorf("can not determin region: %s", err.Error())
	}
	vpcid, err := m.VpcID()
	if err != nil {
		return nil, fmt.Errorf("can not determin vpcid: %s", err.Error())
	}
	ecsclient := ecs.NewECSClientWithSecurityToken4RegionalDomain(key, secret, "", common.Region(region))
	mgr := &ClientMgr{
		stop: make(<-chan struct{}, 1),
		meta: m,
		instance: &InstanceClient{
			c: ecsclient,
		},
		loadbalancer: &LoadBalancerClient{
			vpcid: vpcid,
			ins:   ecsclient,
			c:     slb.NewSLBClientWithSecurityToken4RegionalDomain(key, secret, "", common.Region(region)),
		},
		privateZone: &PrivateZoneClient{
			c: pvtz.NewPVTZClientWithSecurityToken4RegionalDomain(key, secret, "", common.Region("cn-hangzhou")),
		},
		routes: &RoutesClient{
			client: ecs.NewVPCClientWithSecurityToken4RegionalDomain(key, secret, "", common.Region(region)),
			region: region,
		},
	}

	if key == "" || secret == "" {
		klog.Infof("alicloud: use ramrole token mode without ak.")
		mgr.token = &RamRoleToken{meta: m}
	} else {
		inittoken := &Token{
			AccessKey:    key,
			AccessSecret: secret,
			UID:          cfg.Global.UID,
		}
		if inittoken.UID == "" {
			klog.Infof("alicloud: ak mode to authenticate user. without token and role assume")
			mgr.token = &AkAuthToken{ak: inittoken}
		} else {
			klog.Infof("alicloud: service account auth mode")
			mgr.token = &ServiceToken{svcak: inittoken}
		}
	}
	return mgr, nil
}

func (mgr *ClientMgr) Start(settoken func(mgr *ClientMgr, token *Token) error) error {
	initialized := false
	tokenfunc := func() {
		// refresh client token periodically
		token, err := mgr.token.NextToken()
		if err != nil {
			klog.Errorf("token retrieve: %s", err.Error())
			return
		}
		err = settoken(mgr, token)
		if err != nil {
			klog.Errorf("set token: %s", err.Error())
			return
		}
		initialized = true
	}

	go wait.Until(
		tokenfunc,
		time.Duration(TOKEN_RESYNC_PERIOD),
		mgr.stop,
	)
	return wait.ExponentialBackoff(
		wait.Backoff{
			Steps:    7,
			Duration: 1 * time.Second,
			Jitter:   1,
			Factor:   2,
		}, func() (done bool, err error) {
			tokenfunc()
			klog.Infof("wait for token ready")
			return initialized, nil
		},
	)
}

func RefreshToken(mgr *ClientMgr, token *Token) error {
	ecsclient := mgr.instance.c.(*ecs.Client)
	slbclient := mgr.loadbalancer.c.(*slb.Client)
	pvtzclient := mgr.privateZone.c.(*pvtz.Client)
	vpcclient := mgr.routes.client.(*ecs.Client)
	ecsclient.WithSecurityToken(token.Token).
		WithAccessKeyId(token.AccessKey).
		WithAccessKeySecret(token.AccessSecret)
	slbclient.WithSecurityToken(token.Token).
		WithAccessKeyId(token.AccessKey).
		WithAccessKeySecret(token.AccessSecret)
	pvtzclient.WithSecurityToken(token.Token).
		WithAccessKeyId(token.AccessKey).
		WithAccessKeySecret(token.AccessSecret)
	vpcclient.WithSecurityToken(token.Token).
		WithAccessKeyId(token.AccessKey).
		WithAccessKeySecret(token.AccessSecret)

	ecsclient.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	slbclient.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	pvtzclient.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	vpcclient.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	return nil
}

// Instances return instance client
func (mgr *ClientMgr) Instances() *InstanceClient { return mgr.instance }

// Routes return routes client
func (mgr *ClientMgr) Routes() *RoutesClient { return mgr.routes }

// LoadBalancers return loadbalancer client
func (mgr *ClientMgr) LoadBalancers() *LoadBalancerClient { return mgr.loadbalancer }

// PrivateZones return PrivateZones client
func (mgr *ClientMgr) PrivateZones() *PrivateZoneClient { return mgr.privateZone }

// MetaData return MetaData client
func (mgr *ClientMgr) MetaData() IMetaData { return mgr.meta }

// Token base token info
type Token struct {
	AccessSecret string `json:"accessSecret,omitempty"`
	UID          string `json:"uid,omitempty"`
	Token        string `json:"token,omitempty"`
	AccessKey    string `json:"accesskey,omitempty"`
}

// TokenAuth is an interface of token auth method
type TokenAuth interface {
	NextToken() (*Token, error)
}

// AkAuthToken implement ak auth
type AkAuthToken struct{ ak *Token }

func (f *AkAuthToken) NextToken() (*Token, error) { return f.ak, nil }

type RamRoleToken struct {
	meta IMetaData
}

func (f *RamRoleToken) NextToken() (*Token, error) {
	roleName, err := f.meta.RoleName()
	if err != nil {
		return nil, fmt.Errorf("role name: %s", err.Error())
	}
	// use instance ram file way.
	role, err := f.meta.RamRoleToken(roleName)
	if err != nil {
		return nil, fmt.Errorf("ramrole token retrieve: %s", err.Error())
	}
	return &Token{
		AccessKey:    role.AccessKeyId,
		AccessSecret: role.AccessKeySecret,
		Token:        role.SecurityToken,
	}, nil
}

// ServiceToken is an implemention of service account auth
type ServiceToken struct {
	svcak    *Token
	execpath string
}

func (f *ServiceToken) NextToken() (*Token, error) {
	status := <-cmd.NewCmd(
		filepath.Join(f.execpath, "servicetoken"),
		fmt.Sprintf("--uid=%s", f.svcak.UID),
		fmt.Sprintf("--key=%s", f.svcak.AccessKey),
		fmt.Sprintf("--secret=%s", f.svcak.AccessSecret),
	).Start()
	if status.Error != nil {
		return nil, fmt.Errorf("invoke servicetoken: %s", status.Error.Error())
	}
	token := &Token{}
	err := json.Unmarshal(
		[]byte(strings.Join(status.Stdout, "")),
		token,
	)
	if err == nil {
		return token, nil
	}
	return nil, fmt.Errorf("unmarshal token: %s", err.Error())
}

// IMetaData metadata interface
type IMetaData interface {
	HostName() (string, error)
	ImageID() (string, error)
	InstanceID() (string, error)
	Mac() (string, error)
	NetworkType() (string, error)
	OwnerAccountID() (string, error)
	PrivateIPv4() (string, error)
	Region() (string, error)
	SerialNumber() (string, error)
	SourceAddress() (string, error)
	VpcCIDRBlock() (string, error)
	VpcID() (string, error)
	VswitchCIDRBlock() (string, error)
	Zone() (string, error)
	NTPConfigServers() ([]string, error)
	RoleName() (string, error)
	RamRoleToken(role string) (metadata.RoleAuth, error)
	VswitchID() (string, error)
}

// NewMetaData return new metadata
func NewMetaData() IMetaData {
	if cfg.Global.VpcID != "" &&
		cfg.Global.VswitchID != "" {
		klog.V(2).Infof("use mocked metadata server.")
		return &fakeMetaData{base: metadata.NewMetaData(nil)}
	}
	return metadata.NewMetaData(nil)
}

type fakeMetaData struct {
	base IMetaData
}

func (m *fakeMetaData) HostName() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) ImageID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) InstanceID() (string, error) {

	return "fakedInstanceid", nil
}

func (m *fakeMetaData) Mac() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) NetworkType() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) OwnerAccountID() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) PrivateIPv4() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) Region() (string, error) {
	if cfg.Global.Region != "" {
		return cfg.Global.Region, nil
	}
	return m.base.Region()
}

func (m *fakeMetaData) SerialNumber() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) SourceAddress() (string, error) {

	return "", fmt.Errorf("unimplemented")

}

func (m *fakeMetaData) VpcCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) VpcID() (string, error) {
	if cfg.Global.VpcID != "" {
		return cfg.Global.VpcID, nil
	}
	return m.base.VpcID()
}

func (m *fakeMetaData) VswitchCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

// zone1:vswitchid1,zone2:vswitch2
func (m *fakeMetaData) VswitchID() (string, error) {

	if cfg.Global.VswitchID == "" {
		// get vswitch id from meta server
		return m.base.VswitchID()
	}
	zlist := strings.Split(cfg.Global.VswitchID, ",")
	if len(zlist) == 1 {
		vSwitchs := strings.Split(cfg.Global.VswitchID, ":")
		if len(vSwitchs) == 2 {
			klog.Infof("only one vswitchid mode, %s", vSwitchs[1])
			return vSwitchs[1], nil
		}
		klog.Infof("simple vswitchid mode, %s", cfg.Global.VswitchID)
		return cfg.Global.VswitchID, nil
	}
	mzone, err := m.Zone()
	if err != nil {
		return "", fmt.Errorf("retrieve vswitchid error for %s", err.Error())
	}
	for _, zone := range zlist {
		vs := strings.Split(zone, ":")
		if len(vs) != 2 {
			return "", fmt.Errorf("cloud-config vswitch format error: %s", cfg.Global.VswitchID)
		}
		if vs[0] == "" || vs[0] == mzone {
			return vs[1], nil
		}
	}
	klog.Infof("zone[%s] match failed, fallback with simple vswitch id mode, [%s]", mzone, cfg.Global.VswitchID)
	return cfg.Global.VswitchID, nil
}

func (m *fakeMetaData) EIPv4() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) DNSNameServers() ([]string, error) {

	return []string{""}, fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) NTPConfigServers() ([]string, error) {

	return []string{""}, fmt.Errorf("unimplemented")
}

func (m *fakeMetaData) Zone() (string, error) {
	if cfg.Global.ZoneID != "" {
		return cfg.Global.ZoneID, nil
	}
	return m.base.Zone()
}

func (m *fakeMetaData) RoleName() (string, error) {

	return m.base.RoleName()
}

func (m *fakeMetaData) RamRoleToken(role string) (metadata.RoleAuth, error) {

	return m.base.RamRoleToken(role)
}
