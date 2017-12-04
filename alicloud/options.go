package alicloud

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

const (
	ServiceAnnotationLoadBalancerPrefix                        = "service.beta.kubernetes.io/alicloud-loadbalancer-"
	ServiceAnnotationLoadBalancerProtocolPort                  = ServiceAnnotationLoadBalancerPrefix + "protocol-port"
	ServiceAnnotationLoadBalancerAddressType                   = ServiceAnnotationLoadBalancerPrefix + "address-type"
	ServiceAnnotationLoadBalancerSLBNetworkType                = ServiceAnnotationLoadBalancerPrefix + "slb-network-type"
	ServiceAnnotationLoadBalancerChargeType                    = ServiceAnnotationLoadBalancerPrefix + "charge-type"
	ServiceAnnotationLoadBalancerId                            = ServiceAnnotationLoadBalancerPrefix + "id"
	ServiceAnnotationLoadBalancerBackendLabel                  = ServiceAnnotationLoadBalancerPrefix + "backend-label"
	ServiceAnnotationLoadBalancerRegion                        = ServiceAnnotationLoadBalancerPrefix + "region"
	ServiceAnnotationLoadBalancerBandwidth                     = ServiceAnnotationLoadBalancerPrefix + "bandwidth"
	ServiceAnnotationLoadBalancerCertID                        = ServiceAnnotationLoadBalancerPrefix + "cert-id"
	ServiceAnnotationLoadBalancerHealthCheckFlag               = ServiceAnnotationLoadBalancerPrefix + "health-check-flag"
	ServiceAnnotationLoadBalancerHealthCheckType               = ServiceAnnotationLoadBalancerPrefix + "health-check-type"
	ServiceAnnotationLoadBalancerHealthCheckURI                = ServiceAnnotationLoadBalancerPrefix + "health-check-uri"
	ServiceAnnotationLoadBalancerHealthCheckConnectPort        = ServiceAnnotationLoadBalancerPrefix + "health-check-connect-port"
	ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold   = ServiceAnnotationLoadBalancerPrefix + "healthy-threshold"
	ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold = ServiceAnnotationLoadBalancerPrefix + "unhealthy-threshold"
	ServiceAnnotationLoadBalancerHealthCheckInterval           = ServiceAnnotationLoadBalancerPrefix + "health-check-interval"
	ServiceAnnotationLoadBalancerHealthCheckConnectTimeout     = ServiceAnnotationLoadBalancerPrefix + "health-check-connect-timeout"
	ServiceAnnotationLoadBalancerHealthCheckTimeout            = ServiceAnnotationLoadBalancerPrefix + "health-check-timeout"

	MagicHealthCheckConnectPort = -520

	MAX_LOADBALANCER_BACKEND = 20
)

func ExtractAnnotationRequest(service *v1.Service) *AnnotationRequest {
	ar := &AnnotationRequest{}
	annotation := make(map[string]string)
	for k, v := range service.Annotations {
		annotation[replaceCamel(k)] = v
	}
	bandwith := annotation[ServiceAnnotationLoadBalancerBandwidth]
	if bandwith != "" {
		if i, err := strconv.Atoi(bandwith); err == nil {
			ar.Bandwidth = i
		} else {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth must be integer, but got [%s], use default number 50. message=[%s]\n",
				bandwith, err.Error())
			ar.Bandwidth = DEFAULT_BANDWIDTH
		}
	} else {
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

	ar.BackendLabel = annotation[ServiceAnnotationLoadBalancerBackendLabel]

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

	healthCheckConnectPort := annotation[ServiceAnnotationLoadBalancerHealthCheckConnectPort]
	if healthCheckConnectPort != "" {
		port, err := strconv.Atoi(healthCheckConnectPort)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-port must be integer, but got [%s]. message=[%s]\n",
				healthCheckConnectPort, err.Error())
			ar.HealthCheckConnectPort = MagicHealthCheckConnectPort
		} else {
			ar.HealthCheckConnectPort = port
		}
	}

	healthCheckHealthyThreshold := annotation[ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold]
	if healthCheckHealthyThreshold != "" {
		thresh, err := strconv.Atoi(healthCheckHealthyThreshold)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-healthy-threshold must be integer, but got [%s], use default number 3. message=[%s]\n",
				healthCheckHealthyThreshold, err.Error())
			ar.HealthyThreshold = 3
		} else {
			ar.HealthyThreshold = thresh
		}
	}

	healthCheckUnhealthyThreshold := annotation[ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold]
	if healthCheckUnhealthyThreshold != "" {
		unThresh, err := strconv.Atoi(healthCheckUnhealthyThreshold)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-unhealthy-threshold must be integer, but got [%s], use default number 3. message=[%s]\n",
				healthCheckUnhealthyThreshold, err.Error())
			ar.UnhealthyThreshold = 3
		} else {
			ar.UnhealthyThreshold = unThresh
		}
	}

	healthCheckInterval := annotation[ServiceAnnotationLoadBalancerHealthCheckInterval]
	if healthCheckInterval != "" {
		interval, err := strconv.Atoi(healthCheckInterval)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval must be integer, but got [%s], use default number 2. message=[%s]\n",
				healthCheckInterval, err.Error())
			ar.HealthCheckInterval = 2
		} else {
			ar.HealthCheckInterval = interval
		}
	}

	healthCheckConnectTimeout := annotation[ServiceAnnotationLoadBalancerHealthCheckConnectTimeout]
	if healthCheckConnectTimeout != "" {
		connout, err := strconv.Atoi(healthCheckConnectTimeout)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout must be integer, but got [%s], use default number 5. message=[%s]\n",
				healthCheckConnectTimeout, err.Error())
			ar.HealthCheckConnectTimeout = 5
		} else {
			ar.HealthCheckConnectTimeout = connout
		}
	}

	healthCheckTimeout := annotation[ServiceAnnotationLoadBalancerHealthCheckTimeout]
	if healthCheckTimeout != "" {
		hout, err := strconv.Atoi(annotation[healthCheckTimeout])
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout must be integer, but got [%s], use default number 5. message=[%s]\n",
				healthCheckConnectTimeout, err.Error())
			ar.HealthCheckTimeout = 5
		} else {
			ar.HealthCheckTimeout = hout
		}
	}

	return ar
}

func splitCamel(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "SLBN", "etwork" -> "SLB", "Network"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, strings.ToLower(string(s)))
		}
	}
	return
}

func serviceAnnotation(service *v1.Service, annotate string) string {
	for k, v := range service.Annotations {
		if annotate == replaceCamel(k) {
			return v
		}
	}
	return ""
}

func replaceCamel(str string) string {
	if !strings.HasPrefix(str, ServiceAnnotationLoadBalancerPrefix) {
		// If it is not start with ServiceAnnotationLoadBalancerPrefix, just skip.
		return str
	}
	target := str[len(ServiceAnnotationLoadBalancerPrefix):]
	res := splitCamel(target)

	return ServiceAnnotationLoadBalancerPrefix + strings.Join(res, "-")
}
