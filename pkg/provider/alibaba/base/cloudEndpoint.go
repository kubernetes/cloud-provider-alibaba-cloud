package base

import (
	"fmt"
)

// Region represents ECS region
type Region string

// Constants of region definition
const (
	Hangzhou    = Region("cn-hangzhou")
	Shenzhen    = Region("cn-shenzhen")
	Zhangjiakou = Region("cn-zhangjiakou")
	Huhehaote   = Region("cn-huhehaote")
	Chengdu     = Region("cn-chengdu")

	APSouthEast1 = Region("ap-southeast-1")
	APNorthEast1 = Region("ap-northeast-1")
	APSouthEast2 = Region("ap-southeast-2")
	APSouthEast3 = Region("ap-southeast-3")
	APSouthEast5 = Region("ap-southeast-5")

	APSouth1 = Region("ap-south-1")

	USWest1 = Region("us-west-1")
	USEast1 = Region("us-east-1")

	EUCentral1 = Region("eu-central-1")
	EUWest1    = Region("eu-west-1")

	HangZhouFinance = Region("cn-hangzhou-finance-1")

	CNNorth2Gov1 = Region("cn-north-2-gov-1")
)

var CentralDomainServices = map[string]string{
	"pvtz": "pvtz.vpc-proxy.aliyuncs.com",
}

var RegionalDomainServices = []string{
	"ecs",
	"vpc",
	"slb",
}

// Unit-Domain of central product
var UnitRegions = map[Region]interface{}{
	Hangzhou:        Hangzhou,
	Shenzhen:        Shenzhen,
	Zhangjiakou:     Zhangjiakou,
	Chengdu:         Chengdu,
	Huhehaote:       Huhehaote,
	APNorthEast1:    APNorthEast1,
	APSouthEast1:    APSouthEast1,
	APSouthEast2:    APSouthEast2,
	APSouthEast3:    APSouthEast3,
	APSouthEast5:    APSouthEast5,
	APSouth1:        APSouth1,
	USWest1:         USWest1,
	USEast1:         USEast1,
	EUCentral1:      EUCentral1,
	EUWest1:         EUWest1,
	CNNorth2Gov1:    CNNorth2Gov1,
	HangZhouFinance: Hangzhou,
}

// Get openapi endpoint accessed by ecs instance.
// For some UnitRegions, the endpoint pattern is https://[product].[regionid].aliyuncs.com
// For some CentralRegions, the endpoint pattern is https://[product].vpc-proxy.aliyuncs.com
// The other region, the endpoint pattern is https://[product]-vpc.[regionid].aliyuncs.com
func SetEndpoint4RegionalDomain(region Region, serviceCode string) string {
	if endpoint, ok := CentralDomainServices[serviceCode]; ok {
		return fmt.Sprintf("https://%s", endpoint)
	}
	for _, service := range RegionalDomainServices {
		if service == serviceCode {
			if ep, ok := UnitRegions[region]; ok {
				return fmt.Sprintf("https://%s.%s.aliyuncs.com", serviceCode, ep)
			}

			return fmt.Sprintf("https://%s%s.%s.aliyuncs.com", serviceCode, "-vpc", region)
		}
	}
	return ""
}
