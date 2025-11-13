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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ess"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/pvtz"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sls"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"k8s.io/klog/v2/klogr"

	efloController "github.com/alibabacloud-go/eflo-controller-20221215/v2/client"
	nlb "github.com/alibabacloud-go/nlb-20220430/v4/client"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/version"
)

const (
	KubernetesCloudControllerManager = "ack.ccm"
	AgentClusterId                   = "ClusterId"
	TokenSyncPeriod                  = 10 * time.Minute

	AccessKeyID     = "ACCESS_KEY_ID"
	AccessKeySecret = "ACCESS_KEY_SECRET"
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
	ALB  *alb.Client
	NLB  *nlb.Client
	SLS  *sls.Client
	CAS  *cas.Client
	ESS  *ess.Client
	EFLO *efloController.Client
}

// NewClientMgr return a new client manager
func NewClientMgr() (*ClientMgr, error) {
	if err := ctrlCfg.CloudCFG.LoadCloudCFG(); err != nil {
		return nil, fmt.Errorf("load cloud config %s error: %s", ctrlCfg.ControllerCFG.CloudConfigPath, err.Error())
	}

	meta := NewMetaData()
	CLUSTER_ID = meta.ClusterID()
	region, err := meta.Region()
	if err != nil {
		return nil, fmt.Errorf("can not determin region: %s", err.Error())
	}

	credential := &credentials.StsTokenCredential{
		AccessKeyId:       "key",
		AccessKeySecret:   "secret",
		AccessKeyStsToken: "",
	}

	ecli, err := ecs.NewClientWithOptions(region, clientCfg(), credential)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba ecs client: %s", err.Error())
	}
	ecli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)
	ecli.AppendUserAgent(AgentClusterId, CLUSTER_ID)

	vpcli, err := vpc.NewClientWithOptions(region, clientCfg(), credential)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba vpc client: %s", err.Error())
	}
	vpcli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)
	vpcli.AppendUserAgent(AgentClusterId, CLUSTER_ID)

	slbcli, err := slb.NewClientWithOptions(region, clientCfg(), credential)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba slb client: %s", err.Error())
	}
	slbcli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)
	slbcli.AppendUserAgent(AgentClusterId, CLUSTER_ID)

	albcli, err := alb.NewClientWithOptions(region, clientCfg(), credential)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba alb client: %s", err.Error())
	}
	albcli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)
	albcli.AppendUserAgent(AgentClusterId, CLUSTER_ID)

	slscli, err := sls.NewClientWithOptions(region, clientCfg(), credential)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba sls client: %s", err.Error())
	}
	slscli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)
	slscli.AppendUserAgent(AgentClusterId, CLUSTER_ID)

	cascli, err := cas.NewClientWithOptions(region, clientCfg(), credential)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba cas client: %s", err.Error())
	}
	cascli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)
	cascli.AppendUserAgent(AgentClusterId, CLUSTER_ID)

	pvtzcli, err := pvtz.NewClientWithOptions(region, clientCfg(), credential)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba pvtz client: %s", err.Error())
	}
	pvtzcli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)
	pvtzcli.AppendUserAgent(AgentClusterId, CLUSTER_ID)

	esscli, err := ess.NewClientWithOptions(region, clientCfg(), credential)
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba pvtz client: %s", err.Error())
	}
	esscli.AppendUserAgent(KubernetesCloudControllerManager, version.Version)
	esscli.AppendUserAgent(AgentClusterId, CLUSTER_ID)

	// new sdk
	nlbcli, err := nlb.NewClient(openapiCfg(region, credential, ctrlCfg.ControllerCFG.NetWork))
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba nlb client: %s", err.Error())
	}

	eflocli, err := efloController.NewClient(openapiCfg(region, credential, ctrlCfg.ControllerCFG.NetWork))
	if err != nil {
		return nil, fmt.Errorf("initialize alibaba eflo client: %s", err.Error())
	}

	auth := &ClientMgr{
		Meta:   meta,
		ECS:    ecli,
		VPC:    vpcli,
		SLB:    slbcli,
		PVTZ:   pvtzcli,
		ALB:    albcli,
		NLB:    nlbcli,
		SLS:    slscli,
		CAS:    cascli,
		ESS:    esscli,
		EFLO:   eflocli,
		Region: region,
		stop:   make(<-chan struct{}, 1),
	}
	return auth, nil
}

func (mgr *ClientMgr) Start(
	settoken func(mgr *ClientMgr, token *DefaultToken) error,
) error {
	initialized := false
	tokenAuth := mgr.GetTokenAuth()

	tokenfunc := func() {
		token, err := tokenAuth.NextToken()
		if err != nil {
			log.Error(err, "fail to get next token")
		}
		err = settoken(mgr, token)
		if err != nil {
			log.Error(err, "fail to set token")
			return
		}
		initialized = true
	}

	go wait.Until(
		func() { tokenfunc() },
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
			tokenfunc()
			log.Info("wait for Token ready")
			return initialized, nil
		},
	)
}

func (mgr *ClientMgr) GetTokenAuth() TokenAuth {
	// priority: AddonToken > ServiceToken > AKMode > RamRoleToken
	if _, err := os.Stat(AddonTokenFilePath); err == nil {
		log.Info("use addon token mode to get token")
		return &AddonToken{Region: mgr.Region}
	}

	if ctrlCfg.CloudCFG.Global.AccessKeyID != "" && ctrlCfg.CloudCFG.Global.AccessKeySecret != "" {
		if ctrlCfg.CloudCFG.Global.UID != "" {
			log.Info("use assume role mode to get token")
			return &ServiceToken{Region: mgr.Region}
		} else {
			log.Info("use ak mode to get token")
			return &AkAuthToken{Region: mgr.Region}
		}
	}

	if os.Getenv(AccessKeyID) != "" && os.Getenv(AccessKeySecret) != "" {
		log.Info("use ak mode to get token")
		return &AkAuthToken{Region: mgr.Region}
	}

	log.Info("use ram role mode to get token")
	return &RamRoleToken{meta: mgr.Meta}
}

func RefreshToken(mgr *ClientMgr, token *DefaultToken) error {
	log.V(5).Info("refresh token", "region", token.Region)
	credential := &credentials.StsTokenCredential{
		AccessKeyId:       token.AccessKeyId,
		AccessKeySecret:   token.AccessKeySecret,
		AccessKeyStsToken: token.SecurityToken,
	}

	err := mgr.ECS.InitWithOptions(token.Region, clientCfg(), credential)
	if err != nil {
		return fmt.Errorf("init ecs sts token config: %s", err.Error())
	}

	err = mgr.VPC.InitWithOptions(token.Region, clientCfg(), credential)
	if err != nil {
		return fmt.Errorf("init vpc sts token config: %s", err.Error())
	}

	err = mgr.SLB.InitWithOptions(token.Region, clientCfg(), credential)
	if err != nil {
		return fmt.Errorf("init slb sts token config: %s", err.Error())
	}

	err = mgr.ALB.InitWithOptions(token.Region, clientCfg(), credential)
	if err != nil {
		return fmt.Errorf("init alb sts token config: %s", err.Error())
	}

	err = mgr.SLS.InitWithOptions(token.Region, clientCfg(), credential)
	if err != nil {
		return fmt.Errorf("init sls sts token config: %s", err.Error())
	}

	err = mgr.CAS.InitWithOptions(token.Region, clientCfg(), credential)
	if err != nil {
		return fmt.Errorf("init cas sts token config: %s", err.Error())
	}

	err = mgr.PVTZ.InitWithOptions(token.Region, clientCfg(), credential)
	if err != nil {
		return fmt.Errorf("init pvtz sts token config: %s", err.Error())
	}

	err = mgr.NLB.Init(openapiCfg(token.Region, credential, ctrlCfg.ControllerCFG.NetWork))
	if err != nil {
		return fmt.Errorf("init nlb sts token config: %s", err.Error())
	}

	err = mgr.EFLO.Init(openapiCfg(token.Region, credential, ctrlCfg.ControllerCFG.NetWork))
	if err != nil {
		return fmt.Errorf("init eflo controller sts token config: %s", err.Error())
	}
	// fix eflo endpoint
	ep, err := mgr.EFLO.GetEndpoint(dara.String("eflo-controller"), dara.String(token.Region),
		dara.String("regional"), dara.String(ctrlCfg.ControllerCFG.NetWork), nil, nil, nil)
	if err != nil {
		return fmt.Errorf("init eflo controller endpint config: %s", err.Error())
	}
	mgr.EFLO.Endpoint = ep

	if ctrlCfg.ControllerCFG.NetWork == "vpc" {
		setVPCEndpoint(mgr)
	}

	setCustomizedEndpoint(mgr)

	return nil
}

func setVPCEndpoint(mgr *ClientMgr) {
	mgr.ECS.Network = "vpc"
	mgr.VPC.Network = "vpc"
	mgr.SLB.Network = "vpc"
	mgr.PVTZ.Network = "vpc"
	mgr.ALB.Network = "vpc"
	mgr.SLS.Network = "vpc"
	mgr.CAS.Network = "vpc"
}

func setCustomizedEndpoint(mgr *ClientMgr) {
	if ecsEndpoint, err := parseURL(os.Getenv("ECS_ENDPOINT")); err == nil && ecsEndpoint != "" {
		mgr.ECS.Domain = ecsEndpoint
	}
	if vpcEndpoint, err := parseURL(os.Getenv("VPC_ENDPOINT")); err == nil && vpcEndpoint != "" {
		mgr.VPC.Domain = vpcEndpoint
	}
	if slbEndpoint, err := parseURL(os.Getenv("SLB_ENDPOINT")); err == nil && slbEndpoint != "" {
		mgr.SLB.Domain = slbEndpoint
	}
	if nlbEndpoint, err := parseURL(os.Getenv("NLB_ENDPOINT")); err == nil && nlbEndpoint != "" {
		mgr.NLB.Endpoint = tea.String(nlbEndpoint)
	}
	if efloControllerEndpoint, err := parseURL(os.Getenv("EFLO_CONTROLLER_ENDPOINT")); err == nil && efloControllerEndpoint != "" {
		mgr.EFLO.Endpoint = tea.String(efloControllerEndpoint)
	}
}

func parseURL(str string) (string, error) {
	if str == "" {
		return "", nil
	}

	if !strings.HasPrefix(str, "http") {
		str = "http://" + str
	}
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}

func clientCfg() *sdk.Config {
	scheme := "HTTPS"
	if os.Getenv("ALICLOUD_CLIENT_SCHEME") == "HTTP" {
		scheme = "HTTP"
	}
	return &sdk.Config{
		Timeout:   20 * time.Second,
		Transport: http.DefaultTransport,
		Scheme:    scheme,
	}
}

func openapiCfg(region string, credential *credentials.StsTokenCredential, network string) *openapi.Config {
	scheme := "HTTPS"
	if os.Getenv("ALICLOUD_CLIENT_SCHEME") == "HTTP" {
		scheme = "HTTP"
	}
	return &openapi.Config{
		UserAgent:       tea.String(getUserAgent()),
		Protocol:        tea.String(scheme),
		RegionId:        tea.String(region),
		Network:         &network,
		ConnectTimeout:  tea.Int(20000),
		ReadTimeout:     tea.Int(20000),
		AccessKeyId:     tea.String(credential.AccessKeyId),
		AccessKeySecret: tea.String(credential.AccessKeySecret),
		SecurityToken:   tea.String(credential.AccessKeyStsToken),
	}
}

func getUserAgent() string {
	agents := map[string]string{
		KubernetesCloudControllerManager: version.Version,
		AgentClusterId:                   CLUSTER_ID,
	}
	ret := ""
	for k, v := range agents {
		ret += fmt.Sprintf(" %s/%s", k, v)
	}
	return ret
}
