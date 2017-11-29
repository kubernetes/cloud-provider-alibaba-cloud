package alicloud

import (
	"strconv"
	"strings"
	"k8s.io/api/core/v1"
	"github.com/golang/glog"
	"github.com/denverdino/aliyungo/slb"
	"github.com/denverdino/aliyungo/common"
)

const ServiceAnnotationLoadBalancerProtocolPort 		= "service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port"

const ServiceAnnotationLoadBalancerAddressType 			= "service.beta.kubernetes.io/alicloud-loadbalancer-address-type"

const ServiceAnnotationLoadBalancerSLBNetworkType 		= "service.beta.kubernetes.io/alicloud-loadbalancer-slb-network-type"

const ServiceAnnotationLoadBalancerChargeType 			= "service.beta.kubernetes.io/alicloud-loadbalancer-charge-type"

const ServiceAnnotationLoadBalancerId 				= "service.beta.kubernetes.io/alicloud-loadbalancer-id"

const ServiceAnnotationLoadBalancerBackendLabel 		= "service.beta.kubernetes.io/alicloud-loadbalancer-backend-label"

const ServiceAnnotationLoadBalancerRegion 			= "service.beta.kubernetes.io/alicloud-loadbalancer-region"

const ServiceAnnotationLoadBalancerBandwidth 			= "service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth"

const ServiceAnnotationLoadBalancerCertID 			= "service.beta.kubernetes.io/alicloud-loadbalancer-cert-id"

const ServiceAnnotationLoadBalancerHealthCheckFlag 		= "service.beta.kubernetes.io/alicloud-loadbalancer-health-check-flag"

const ServiceAnnotationLoadBalancerHealthCheckType 		= "service.beta.kubernetes.io/alicloud-loadbalancer-health-check-type"

const ServiceAnnotationLoadBalancerHealthCheckURI 		= "service.beta.kubernetes.io/alicloud-loadbalancer-health-check-uri"

const ServiceAnnotationLoadBalancerHealthCheckConnectPort 	= "service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-port"

const ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold 	= "service.beta.kubernetes.io/alicloud-loadbalancer-healthy-threshold"

const ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold = "service.beta.kubernetes.io/alicloud-loadbalancer-unhealthy-threshold"

const ServiceAnnotationLoadBalancerHealthCheckInterval 		= "service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval"

const ServiceAnnotationLoadBalancerHealthCheckConnectTimeout 	= "service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout"

const ServiceAnnotationLoadBalancerHealthCheckTimeout 		= "service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout"

const MAX_LOADBALANCER_BACKEND=20

func ExtractAnnotationRequest(service *v1.Service) *AnnotationRequest {
	ar := &AnnotationRequest{}
	annotation := make(map[string]string)
	for k,v := range service.Annotations{
		annotation[replaceCamel(k)] = v
	}
	bandwith := annotation[ServiceAnnotationLoadBalancerBandwidth]
	if bandwith != "" {
		if i, err := strconv.Atoi(bandwith);err == nil {
			ar.Bandwidth = i
		} else {
			glog.Warningf("annotation bandwidth must be integer, but got [%s], use default number 50. message=[%s]\n",
				annotation[ServiceAnnotationLoadBalancerBandwidth], err.Error())
			ar.Bandwidth = DEFAULT_BANDWIDTH
		}
	}else {
		ar.Bandwidth = DEFAULT_BANDWIDTH
	}

	addtype := annotation[ServiceAnnotationLoadBalancerAddressType]
	if addtype != "" {
		ar.AddressType = slb.AddressType(addtype)
	} else {
		ar.AddressType = slb.InternetAddressType
	}
	ar.SLBNetworkType = annotation[ServiceAnnotationLoadBalancerSLBNetworkType]

	chargtype := annotation[ServiceAnnotationLoadBalancerChargeType]
	if chargtype != "" {
		ar.ChargeType = slb.InternetChargeType(chargtype)
	} else {
		ar.ChargeType = slb.PayByTraffic
	}

	region := annotation[ServiceAnnotationLoadBalancerRegion]
	if region != "" {
		ar.Region = common.Region(region)
	} else {
		ar.Region = DEFAULT_REGION
	}

	ar.Loadbalancerid = annotation[ServiceAnnotationLoadBalancerId]

	ar.BackendLabel   = annotation[ServiceAnnotationLoadBalancerBackendLabel]

	certid := annotation[ServiceAnnotationLoadBalancerCertID]
	if certid != "" {
		ar.CertID = certid
	}

	hcFlag := annotation[ServiceAnnotationLoadBalancerHealthCheckFlag]
	if hcFlag != "" {
		ar.HealthCheck = slb.FlagType(hcFlag)
	} else {
		ar.HealthCheck = slb.OffFlag
	}

	hcType := annotation[ServiceAnnotationLoadBalancerHealthCheckType]
	if hcType != "" {
		ar.HealthCheckType = slb.HealthCheckType(hcType)
	} else {
		ar.HealthCheckType = slb.TCPHealthCheckType
	}

	hcUri := annotation[ServiceAnnotationLoadBalancerHealthCheckURI]
	if hcUri != "" {
		ar.HealthCheckURI = hcUri
	} else {
		ar.HealthCheckURI = "/"
	}

	port, err := strconv.Atoi(annotation[ServiceAnnotationLoadBalancerHealthCheckConnectPort])
	if err != nil {
		ar.HealthCheckConnectPort = -520
	} else {
		ar.HealthCheckConnectPort = port
	}

	thresh, err := strconv.Atoi(annotation[ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold])
	if err != nil {
		ar.HealthyThreshold = 3
	} else {
		ar.HealthyThreshold = thresh
	}

	unThresh, err := strconv.Atoi(annotation[ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold])
	if err != nil {
		ar.UnhealthyThreshold = 3
	} else {
		ar.UnhealthyThreshold = unThresh
	}

	interval, err := strconv.Atoi(annotation[ServiceAnnotationLoadBalancerHealthCheckInterval])
	if err != nil {
		ar.HealthCheckInterval = 2
	} else {
		ar.HealthCheckInterval = interval
	}

	connout, err := strconv.Atoi(annotation[ServiceAnnotationLoadBalancerHealthCheckConnectTimeout])
	if err != nil {
		ar.HealthCheckConnectTimeout = 5
	} else {
		ar.HealthCheckConnectTimeout = connout
	}

	hout, err := strconv.Atoi(annotation[ServiceAnnotationLoadBalancerHealthCheckTimeout])
	if err != nil {
		ar.HealthCheckTimeout = 5
	} else {
		ar.HealthCheckConnectPort = hout
	}
	return ar
}

func replaceCamel(str string) string{

	if !strings.Contains(str,"alicloud-loadbalancer") {
		// If it is not `alicloud-loadbalancer` , just skip.
		return str
	}
	dot := strings.Split(str, "-")
	target := dot[len(dot) - 1]
	res := []byte{}
	for i,v := range []byte(target){
		if v <= 90 {
			if i>0 && target[i - 1] <=90 {
				// deal with all upper letter case, eg. URL -> url
				res = append(res,v+32)
			}else {
				// eg. A to -a
				res = append(res, 45, v + 32)
			}
			continue
		}
		res = append(res,v)
	}

	return strings.Join(dot[0:len(dot)-1],"-") + string(res)
}

