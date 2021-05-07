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

package metadata

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/pvtz"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/ghodss/yaml"
	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"

	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/version"
)

var KUBERNETES_CLOUD_CONTROLLER_MANAGER = "ack.ccm"

// TOKEN_RESYNC_PERIOD default Token sync period
var TOKEN_RESYNC_PERIOD = 10 * time.Minute

// ClientAuth client manager for aliyun sdk
type ClientAuth struct {
	stop <-chan struct{}

	Meta IMetaData
	ECS  *ecs.Client
	VPC  *vpc.Client
	SLB  *slb.Client
	PVTZ *pvtz.Client
}

// NewClientMgr return a new client manager
func NewClientAuth() (*ClientAuth, error) {
	meta := NewMetaData()
	// ctx2.CFG.Region might be nil,
	// but it is ok to have empty value.
	// because region will be filled in the next token refresh

	// TODO FIX ME
	region := ctx2.CFG.Global.Region
	log.Infof("new provider in region: [%s]", region)
	ecli, err := ecs.NewClientWithStsToken(
		region, "key", "secret", "",
	)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba ecs client: %s", err.Error())
	}
	ecli.AppendUserAgent(KUBERNETES_CLOUD_CONTROLLER_MANAGER, version.Version)

	vpcli, err := vpc.NewClientWithStsToken(
		region, "key", "secret", "")
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba vpc client: %s", err.Error())
	}
	vpcli.AppendUserAgent(KUBERNETES_CLOUD_CONTROLLER_MANAGER, version.Version)

	slbcli, err := slb.NewClientWithStsToken(
		region, "key", "secret", "")
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba slb client: %s", err.Error())
	}
	slbcli.AppendUserAgent(KUBERNETES_CLOUD_CONTROLLER_MANAGER, version.Version)

	pvtzcli, err := pvtz.NewClientWithStsToken(
		region, "key", "secret", "",
	)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba pvtz client: %s", err.Error())
	}
	pvtzcli.AppendUserAgent(KUBERNETES_CLOUD_CONTROLLER_MANAGER, version.Version)

	auth := &ClientAuth{
		Meta: meta,
		ECS:  ecli,
		VPC:  vpcli,
		SLB:  slbcli,
		PVTZ: pvtzcli,
		stop: make(<-chan struct{}, 1),
	}
	return auth, nil
}

func (mgr *ClientAuth) Start(
	settoken func(mgr *ClientAuth, token *Token) error,
) error {
	initialized := false
	tokenfunc := func() {
		log.Infof("load cfg from file: %s", ctx2.GlobalFlag.CloudConfig)
		// reload config while token refresh
		err := LoadCfg(ctx2.GlobalFlag.CloudConfig)
		if err != nil {
			log.Warnf("load config fail: %s", err.Error())
			return
		}
		// refresh client Token periodically
		token, err := mgr.Token().NextToken()
		if err != nil {
			log.Errorf("return next token: %s", err.Error())
			return
		}
		err = settoken(mgr, token)
		if err != nil {
			log.Errorf("set Token: %s", err.Error())
			return
		}
		initialized = true
	}
	go wait.Until(
		tokenfunc,
		TOKEN_RESYNC_PERIOD,
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
			log.Infof("wait for Token ready")
			return initialized, nil
		},
	)
}

func LoadCfg(cfg string) error {
	content, err := ioutil.ReadFile(cfg)
	if err != nil {
		return fmt.Errorf("read config file: %s", content)
	}
	return yaml.Unmarshal(content, ctx2.CFG)
}

func (mgr *ClientAuth) Token() TokenAuth {
	key, err := b64.StdEncoding.DecodeString(ctx2.CFG.Global.AccessKeyID)
	if err != nil {
		panic(fmt.Sprintf("ak key must be base64 encoded: %s", err.Error()))
	}
	secret, err := b64.StdEncoding.DecodeString(ctx2.CFG.Global.AccessKeySecret)
	if err != nil {
		panic(fmt.Sprintf("ak secret must be base64 encoded: %s", err.Error()))
	}
	if len(key) == 0 ||
		len(secret) == 0 {
		log.Infof("ccm: use ramrole Token mode without ak.")
		return &RamRoleToken{meta: mgr.Meta}
	}
	region := ctx2.CFG.Global.Region
	if region == "" {
		region, err = mgr.Meta.Region()
		if err != nil {
			panic(fmt.Sprintf("region not specified in config, detect region failed: %s", err.Error()))
		}
	}
	inittoken := &Token{
		AccessKey:    string(key),
		AccessSecret: string(secret),
		UID:          ctx2.CFG.Global.UID,
		Region:       region,
	}
	if inittoken.UID == "" {
		log.Infof("ccm: ak mode to authenticate user. without Token and role assume")
		return &AkAuthToken{ak: inittoken}
	}
	log.Infof("ccm: service account auth mode")
	return &ServiceToken{svcak: inittoken}
}

func RefreshToken(mgr *ClientAuth, token *Token) error {
	log.Infof("refresh token: [region=%s]", token.Region)
	err := mgr.ECS.InitWithStsToken(
		token.Region, token.AccessKey, token.AccessSecret, token.Token,
	)
	if err != nil {
		return fmt.Errorf("init ecs sts token config: %s", err.Error())
	}

	err = mgr.VPC.InitWithStsToken(
		token.Region, token.AccessKey, token.AccessSecret, token.Token,
	)
	if err != nil {
		return fmt.Errorf("init vpc sts token config: %s", err.Error())
	}

	err = mgr.SLB.InitWithStsToken(
		token.Region, token.AccessKey, token.AccessSecret, token.Token,
	)
	if err != nil {
		return fmt.Errorf("init slb sts token config: %s", err.Error())
	}

	err = mgr.PVTZ.InitWithStsToken(
		token.Region, token.AccessKey, token.AccessSecret, token.Token,
	)
	if err != nil {
		return fmt.Errorf("init pvtz sts token config: %s", err.Error())
	}

	// TODO revert me
	//setVPCEndpoint(mgr)
	return nil
}

func setVPCEndpoint(mgr *ClientAuth) {
	mgr.ECS.Network = "vpc"
	mgr.VPC.Network = "vpc"
	mgr.SLB.Network = "vpc"
	mgr.PVTZ.Network = "vpc"
}

// MetaData return MetaData client
func (mgr *ClientAuth) MetaData() IMetaData { return mgr.Meta }

// Token base Token info
type Token struct {
	Region       string `json:"region,omitempty"`
	AccessSecret string `json:"accessSecret,omitempty"`
	UID          string `json:"uid,omitempty"`
	Token        string `json:"token,omitempty"`
	AccessKey    string `json:"accesskey,omitempty"`
}

// TokenAuth is an interface of Token auth method
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
		return nil, fmt.Errorf("ramrole Token retrieve: %s", err.Error())
	}
	region := ctx2.CFG.Global.Region
	if region == "" {
		region, err = f.meta.Region()
		if err != nil {
			return nil, fmt.Errorf("read region error: %s", err.Error())
		}
	}
	return &Token{
		Token:        role.SecurityToken,
		Region:       region,
		AccessKey:    role.AccessKeyId,
		AccessSecret: role.AccessKeySecret,
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
	token := &Token{Region: f.svcak.Region}
	err := json.Unmarshal(
		[]byte(strings.Join(status.Stdout, "")), token,
	)
	if err == nil {
		return token, nil
	}
	return nil, fmt.Errorf("unmarshal Token: %s, %s, %s", err.Error(), status.Stdout, status.Stderr)
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
	RamRoleToken(role string) (RoleAuth, error)
	VswitchID() (string, error)
}

// NewMetaData return new metadata
func NewMetaData() IMetaData {
	if false {
		// use mocked Meta depend
		log.Infof("use mocked metadata server.")
		return &fakeMetaData{base: NewBaseMetaData(nil)}
	}
	return NewBaseMetaData(nil)
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
	return m.base.VpcID()
}

func (m *fakeMetaData) VswitchCIDRBlock() (string, error) {

	return "", fmt.Errorf("unimplemented")
}

// zone1:vswitchid1,zone2:vswitch2
func (m *fakeMetaData) VswitchID() (string, error) {

	return "", fmt.Errorf("unimplemented")
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
	return m.base.Zone()
}

func (m *fakeMetaData) RoleName() (string, error) {

	return m.base.RoleName()
}

func (m *fakeMetaData) RamRoleToken(role string) (RoleAuth, error) {

	return m.base.RamRoleToken(role)
}
