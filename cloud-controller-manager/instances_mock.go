package alicloud

import (
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/golang/glog"
	"reflect"
	"strings"
	"sync"
)

var INSTANCE = InstanceStore{}

type InstanceStore struct {
	instance sync.Map
}

func WithNewInstanceStore() InitialSet {
	return func() {
		INSTANCE = InstanceStore{}
	}
}

func WithInstance() InitialSet {
	return func() {
		INSTANCE.instance.Store(
			INSTANCEID,
			ecs.InstanceAttributesType{
				InstanceId:          INSTANCEID,
				ImageId:             "centos_7_04_64_20G_alibase_201701015.vhd",
				RegionId:            REGION,
				ZoneId:              REGION_A,
				InstanceType:        "ecs.sn1ne.large",
				InstanceTypeFamily:  "ecs.sn1ne",
				Status:              "running",
				InstanceNetworkType: "vpc",
				VpcAttributes: ecs.VpcAttributesType{
					VpcId:     VPCID,
					VSwitchId: VSWITCH_ID,
					PrivateIpAddress: ecs.IpAddressSetType{
						IpAddress: []string{"192.168.211.130"},
					},
				},
				InstanceChargeType: common.PostPaid,
			},
		)
	}
}

type mockClientInstanceSDK struct {
	describeInstances func(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error)
}

func (m *mockClientInstanceSDK) DescribeInstances(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error) {
	if m.describeInstances != nil {
		return m.describeInstances(args)
	}
	var results []ecs.InstanceAttributesType
	INSTANCE.instance.Range(
		func(key, value interface{}) bool {
			v, ok := value.(ecs.InstanceAttributesType)
			if !ok {
				glog.Info("API: DescribeInstances, "+
					"unexpected type %s, not slb.InstanceAttributesType", reflect.TypeOf(value))
				return true
			}
			if args.InstanceIds != "" &&
				!strings.Contains(args.InstanceIds, v.InstanceId) {
				// continue next
				return true
			}
			if args.RegionId != "" &&
				args.RegionId != v.RegionId {
				// continue next
				return true
			}
			results = append(results, v)
			return true
		},
	)
	return results, nil, nil
}
