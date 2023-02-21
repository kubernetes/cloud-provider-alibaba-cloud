package alb

import (
	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
)

type ServerGroupWithTags struct {
	albsdk.ServerGroup
	Servers []albsdk.BackendServer
	Tags    map[string]string
}
type AlbLoadBalancerWithTags struct {
	albsdk.LoadBalancer
	Tags map[string]string
}
