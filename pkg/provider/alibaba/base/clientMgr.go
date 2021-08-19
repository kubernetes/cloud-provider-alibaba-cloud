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

package base

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog/klogr"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/pvtz"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/ghodss/yaml"
	"github.com/go-cmd/cmd"

	"k8s.io/apimachinery/pkg/util/wait"
	ctrlCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/version"
)

const (
	KubernetesCloudControllerManager = "ack.ccm"
	TokenSyncPeriod                  = 10 * time.Minute
)

type AuthMode string

const (
	AKMode      = AuthMode("ak")      //get token by accessKeyId and accessKeySecretId
	SAMode      = AuthMode("service") //get token by assuming role
	RamRoleMode = AuthMode("ramrole") //get token by ecs ram role
)

var log = klogr.New().WithName("clientMgr")

// ClientMgr client manager for aliyun sdk
type ClientMgr struct {
	stop   <-chan struct{}
	Region string

	Meta prvd.IMetaData
	ECS  *ecs.Client
	VPC  *vpc.Client
	SLB  *slb.Client
	PVTZ *pvtz.Client
}

// NewClientMgr return a new client manager
func NewClientMgr() (*ClientMgr, error) {
	if err := loadCloudCFG(); err != nil {
		return nil, fmt.Errorf("load cloud config %s error: %s", ctrlCtx.ControllerCFG.CloudConfig, err.Error())
	}

	meta := NewMetaData()
	CLUSTER_ID = meta.ClusterID()
	region, err := meta.Region()
	if err != nil {
		return nil, fmt.Errorf("can not determin region: %s", err.Error())
	}

	ecli, err := ecs.NewClientWithStsToken(
		region, "key", "secret", "",
	)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba ecs client: %s", err.Error())
	}
	ecli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)

	vpcli, err := vpc.NewClientWithStsToken(
		region, "key", "secret", "")
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba vpc client: %s", err.Error())
	}
	vpcli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)

	slbcli, err := slb.NewClientWithStsToken(
		region, "key", "secret", "")
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba slb client: %s", err.Error())
	}
	slbcli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)

	pvtzcli, err := pvtz.NewClientWithStsToken(
		region, "key", "secret", "",
	)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba pvtz client: %s", err.Error())
	}
	pvtzcli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)

	auth := &ClientMgr{
		Meta:   meta,
		ECS:    ecli,
		VPC:    vpcli,
		SLB:    slbcli,
		PVTZ:   pvtzcli,
		Region: region,
		stop:   make(<-chan struct{}, 1),
	}
	return auth, nil
}

func (mgr *ClientMgr) Start(
	settoken func(mgr *ClientMgr, token *Token) error,
) error {
	initialized := false
	authMode := mgr.GetAuthMode()

	tokenfunc := func(authMode AuthMode) {
		var err error
		token := &Token{
			Region: mgr.Region,
		}
		switch authMode {
		case AKMode:
			akToken := &AkAuthToken{ak: token}
			token, err = akToken.NextToken()
		case SAMode:
			saToken := &ServiceToken{svcak: token}
			token, err = saToken.NextToken()
		case RamRoleMode:
			ramRoleToken := &RamRoleToken{meta: mgr.Meta}
			token, err = ramRoleToken.NextToken()
		}
		if err != nil {
			log.Error(err, "fail to get next token")
			return
		}
		err = settoken(mgr, token)
		if err != nil {
			log.Error(err, "fail to set token")
			return
		}
		initialized = true
	}
	go wait.Until(
		func() { tokenfunc(authMode) },
		TokenSyncPeriod,
		mgr.stop,
	)
	return wait.ExponentialBackoff(
		wait.Backoff{
			Steps:    7,
			Duration: 1 * time.Second,
			Jitter:   1,
			Factor:   2,
		}, func() (done bool, err error) {
			tokenfunc(authMode)
			log.Info("wait for Token ready")
			return initialized, nil
		},
	)
}

func (mgr *ClientMgr) GetAuthMode() AuthMode {
	if ctrlCtx.CloudCFG.Global.AccessKeyID != "" &&
		ctrlCtx.CloudCFG.Global.AccessKeySecret != "" {
		if ctrlCtx.CloudCFG.Global.UID != "" {
			log.Info("use assume role mode to get token")
			return SAMode
		} else {
			log.Info("use ak mode to get token")
			return AKMode
		}
	}

	if os.Getenv("ACCESS_KEY_ID") != "" &&
		os.Getenv("ACCESS_KEY_SECRET") != "" {
		log.Info("use ak mode to get token")
		return AKMode
	}
	log.Info("use ram role mode to get token")
	return RamRoleMode
}

func RefreshToken(mgr *ClientMgr, token *Token) error {
	log.V(5).Info("refresh token", "region", token.Region)
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

	if ctrlCtx.ControllerCFG.NetWork == "vpc" {
		setVPCEndpoint(mgr)
	}

	return nil
}

func setVPCEndpoint(mgr *ClientMgr) {
	mgr.ECS.Network = "vpc"
	mgr.VPC.Network = "vpc"
	mgr.SLB.Network = "vpc"
	mgr.PVTZ.Network = "vpc"
}

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

func (f *AkAuthToken) NextToken() (*Token, error) {
	key, secret, err := loadAK()
	if err != nil {
		return f.ak, err
	}
	f.ak = &Token{
		AccessSecret: secret,
		AccessKey:    key,
		Region:       f.ak.Region,
	}
	return f.ak, nil
}

type RamRoleToken struct {
	meta prvd.IMetaData
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
	region, err := f.meta.Region()
	if err != nil {
		return nil, fmt.Errorf("read region error: %s", err.Error())
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
	key, secret, err := loadAK()
	if err != nil {
		return nil, err
	}
	status := <-cmd.NewCmd(
		filepath.Join(f.execpath, "servicetoken"),
		fmt.Sprintf("--uid=%s", ctrlCtx.CloudCFG.Global.UID),
		fmt.Sprintf("--key=%s", key),
		fmt.Sprintf("--secret=%s", secret),
		fmt.Sprintf("--region=%s", f.svcak.Region),
	).Start()
	if status.Error != nil {
		return nil, fmt.Errorf("invoke servicetoken: %s", status.Error.Error())
	}
	token := &Token{Region: f.svcak.Region}
	if err = json.Unmarshal([]byte(strings.Join(status.Stdout, "")), token); err != nil {
		return nil, fmt.Errorf("unmarshal Token error: %s, %s, %s", err.Error(), status.Stdout, status.Stderr)
	}

	return token, nil
}

func loadCloudCFG() error {
	content, err := ioutil.ReadFile(ctrlCtx.ControllerCFG.CloudConfig)
	if err != nil {
		return fmt.Errorf("read cloud config error: %s ", err.Error())
	}
	return yaml.Unmarshal(content, ctrlCtx.CloudCFG)
}

func loadAK() (string, string, error) {
	var keyId, keySecret string
	log.V(5).Info(fmt.Sprintf("load cfg from file: %s", ctrlCtx.ControllerCFG.CloudConfig))
	if err := loadCloudCFG(); err != nil {
		return "", "", fmt.Errorf("load cloud config %s error: %v",
			ctrlCtx.ControllerCFG.CloudConfig, err.Error())
	}

	if ctrlCtx.CloudCFG.Global.AccessKeyID != "" && ctrlCtx.CloudCFG.Global.AccessKeySecret != "" {
		key, err := base64.StdEncoding.DecodeString(ctrlCtx.CloudCFG.Global.AccessKeyID)
		if err != nil {
			return "", "", err
		}
		keyId = string(key)
		secret, err := base64.StdEncoding.DecodeString(ctrlCtx.CloudCFG.Global.AccessKeySecret)
		if err != nil {
			return "", "", err
		}
		keySecret = string(secret)
	}

	if keyId == "" || keySecret == "" {
		log.V(5).Info("LoadAK: cloud config does not have keyId or keySecret. " +
			"try environment ACCESS_KEY_ID ACCESS_KEY_SECRET")
		keyId = os.Getenv("ACCESS_KEY_ID")
		keySecret = os.Getenv("ACCESS_KEY_SECRET")
		if keyId == "" || keySecret == "" {
			return "", "", fmt.Errorf("cloud config and env do not have keyId or keySecret, load AK failed")
		}
	}

	return keyId, keySecret, nil
}
