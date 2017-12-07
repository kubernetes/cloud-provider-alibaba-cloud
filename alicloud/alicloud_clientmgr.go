package alicloud

import (
	"sync"
	"github.com/denverdino/aliyungo/metadata"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"
	"github.com/patrickmn/go-cache"
	"github.com/denverdino/aliyungo/common"
	"github.com/golang/glog"
)


var ROLE_NAME="CloudControllerManager"

var TOKEN_RESYNC_PERIOD=5 * time.Minute

type TokenAuth struct {
	lock   	sync.RWMutex
	auth   	metadata.RoleAuth
	active 	bool
}

func (token *TokenAuth) authid()(string, string,string) {
	token.lock.RLock()
	defer token.lock.RUnlock()

	return token.auth.AccessKeyId,
		token.auth.AccessKeySecret,
		token.auth.SecurityToken
}

type ClientMgr struct {
	stop   		<-chan struct{}

	token      	*TokenAuth

	meta       	*metadata.MetaData
	routes     	*RoutesClient
	loadbalancer 	*LoadBalancerClient
	instance   	*InstancerClient
}

func NewClientMgr(key, secret string) (*ClientMgr, error){
	token := &TokenAuth{
		auth :  metadata.RoleAuth{
			 AccessKeyId: 		key,
			 AccessKeySecret: 	secret,
		},
		active: false,
	}
	m := metadata.NewMetaData(nil)

	if key == "" || secret == "" {
		role, err := m.RamRoleToken(ROLE_NAME)
		if err != nil {
			return nil, err
		}
		token.auth   = role
		token.active = true
	}
	keyid, sec, tok := token.authid()
	ecsclient := ecs.NewClient(keyid, sec)
	ecsclient.SetSecurityToken(tok)
	ecsclient.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	ecsclient.SetRegionID(DEFAULT_REGION)

	slbclient := slb.NewClient(keyid, sec)
	slbclient.SetSecurityToken(tok)
	slbclient.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	slbclient.SetRegionID(DEFAULT_REGION)

	mgr := &ClientMgr{
		stop: 		make(<-chan struct{},1),
		token: 		token,
		meta:          	m,
		instance:       &InstancerClient{
			c: 	ecsclient,
		},
		loadbalancer:  	&LoadBalancerClient{
			c: 	slbclient,
		},
		routes:        	&RoutesClient{
			client: 	ecsclient,
			routers: 	cache.New(defaultCacheExpiration, defaultCacheExpiration),
			vpcs:    	cache.New(defaultCacheExpiration, defaultCacheExpiration),
		},
	}
	if ! token.active {
		// use key and secret
		glog.Infof("alicloud: clientmgr, use accesskeyid and accesskeysecret mode to authenticate user. without token")
		return mgr,nil
	}
	go wait.Until(func () {
		// refresh client token periodically
		token.lock.Lock()
		defer token.lock.Unlock()
		role, err := mgr.meta.RamRoleToken(ROLE_NAME)
		if err != nil {
			glog.Errorf("alicloud: clientmgr, error get ram role token [%s]\n",err.Error())
			return
		}
		token.auth = role
		ecsclient.WithSecurityToken(role.SecurityToken).
			WithAccessKeyId(role.AccessKeyId).
			WithAccessKeySecret(role.AccessKeySecret)
		slbclient.WithSecurityToken(role.SecurityToken).
			WithAccessKeyId(role.AccessKeyId).
			WithAccessKeySecret(role.AccessKeySecret)
	},time.Duration(TOKEN_RESYNC_PERIOD),mgr.stop)

	return mgr, nil
}



func (c * ClientMgr) Instances(region common.Region) *InstancerClient {
	return c.instance
}

func (c * ClientMgr) Routes(region common.Region) *RoutesClient {
	return c.routes
}

func (c * ClientMgr) LoadBalancers(region common.Region) *LoadBalancerClient {
	return c.loadbalancer
}

func (c * ClientMgr) MetaData() *metadata.MetaData {

	return c.meta
}