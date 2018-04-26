package alicloud

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
// defaulted is the parameters which set by programe.
// request represent user defined parameters.
func ExtractAnnotationRequest(service *v1.Service) (*AnnotationRequest, *AnnotationRequest ) {
	defaulted, request := &AnnotationRequest{}, &AnnotationRequest{}
	annotation := make(map[string]string)
	for k, v := range service.Annotations {
		annotation[replaceCamel(k)] = v
	}
	bandwith := annotation[ServiceAnnotationLoadBalancerBandwidth]
	if bandwith != "" {
		if i, err := strconv.Atoi(bandwith); err == nil {
			request.Bandwidth = i
			defaulted.Bandwidth = i
		} else {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth must be integer, but got [%s], use default number 50. message=[%s]\n",
				bandwith, err.Error())
			defaulted.Bandwidth = DEFAULT_BANDWIDTH
		}
	} else {
		defaulted.Bandwidth = DEFAULT_BANDWIDTH
	}

	addtype := annotation[ServiceAnnotationLoadBalancerAddressType]
	if addtype != "" {
		defaulted.AddressType 	   = slb.AddressType(addtype)
		request.AddressType = defaulted.AddressType
	} else {
		defaulted.AddressType = slb.InternetAddressType
	}
	defaulted.SLBNetworkType 	= annotation[ServiceAnnotationLoadBalancerSLBNetworkType]
	request.SLBNetworkType 	= defaulted.SLBNetworkType

	chargtype := annotation[ServiceAnnotationLoadBalancerChargeType]
	if chargtype != "" {
		defaulted.ChargeType = slb.InternetChargeType(chargtype)
		request.ChargeType = defaulted.ChargeType
	} else {
		defaulted.ChargeType = slb.PayByTraffic
	}

	region := annotation[ServiceAnnotationLoadBalancerRegion]
	if region != "" {
		defaulted.Region = common.Region(region)
		request.Region = defaulted.Region
	} else {
		defaulted.Region = DEFAULT_REGION
	}

	defaulted.Loadbalancerid 	= annotation[ServiceAnnotationLoadBalancerId]
	request.Loadbalancerid 	= defaulted.Loadbalancerid

	defaulted.BackendLabel 	= annotation[ServiceAnnotationLoadBalancerBackendLabel]
	request.BackendLabel 	= defaulted.BackendLabel

	defaulted.CertID 	= annotation[ServiceAnnotationLoadBalancerCertID]
	request.CertID 	= defaulted.CertID

	hcFlag := annotation[ServiceAnnotationLoadBalancerHealthCheckFlag]
	if hcFlag != "" {
		defaulted.HealthCheck = slb.FlagType(hcFlag)
		request.HealthCheck = defaulted.HealthCheck
	} else {
		defaulted.HealthCheck = slb.OffFlag
	}

	hcType := annotation[ServiceAnnotationLoadBalancerHealthCheckType]
	if hcType != "" {
		defaulted.HealthCheckType = slb.HealthCheckType(hcType)
		request.HealthCheckType = defaulted.HealthCheckType
	} else {
		defaulted.HealthCheckType = slb.TCPHealthCheckType
	}

	hcUri := annotation[ServiceAnnotationLoadBalancerHealthCheckURI]
	if hcUri != "" {
		defaulted.HealthCheckURI = hcUri
		request.HealthCheckURI = defaulted.HealthCheckURI
	} else {
		defaulted.HealthCheckURI = "/"
	}

	healthCheckConnectPort := annotation[ServiceAnnotationLoadBalancerHealthCheckConnectPort]
	if healthCheckConnectPort != "" {
		port, err := strconv.Atoi(healthCheckConnectPort)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-port must be integer, but got [%s]. message=[%s]\n",
				healthCheckConnectPort, err.Error())
			defaulted.HealthCheckConnectPort = MagicHealthCheckConnectPort
		} else {
			defaulted.HealthCheckConnectPort = port
			request.HealthCheckConnectPort = defaulted.HealthCheckConnectPort
		}
	}

	healthCheckHealthyThreshold := annotation[ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold]
	if healthCheckHealthyThreshold != "" {
		thresh, err := strconv.Atoi(healthCheckHealthyThreshold)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-healthy-threshold must be integer, but got [%s], use default number 3. message=[%s]\n",
				healthCheckHealthyThreshold, err.Error())
			defaulted.HealthyThreshold = 3
		} else {
			defaulted.HealthyThreshold = thresh
			request.HealthyThreshold = defaulted.UnhealthyThreshold
		}
	}

	healthCheckUnhealthyThreshold := annotation[ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold]
	if healthCheckUnhealthyThreshold != "" {
		unThresh, err := strconv.Atoi(healthCheckUnhealthyThreshold)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-unhealthy-threshold must be integer, but got [%s], use default number 3. message=[%s]\n",
				healthCheckUnhealthyThreshold, err.Error())
			defaulted.UnhealthyThreshold = 3
		} else {
			defaulted.UnhealthyThreshold = unThresh
			request.UnhealthyThreshold = defaulted.UnhealthyThreshold
		}
	}

	healthCheckInterval := annotation[ServiceAnnotationLoadBalancerHealthCheckInterval]
	if healthCheckInterval != "" {
		interval, err := strconv.Atoi(healthCheckInterval)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval must be integer, but got [%s], use default number 2. message=[%s]\n",
				healthCheckInterval, err.Error())
			defaulted.HealthCheckInterval = 2
		} else {
			defaulted.HealthCheckInterval = interval
			request.HealthCheckInterval = defaulted.HealthCheckInterval
		}
	}

	healthCheckConnectTimeout := annotation[ServiceAnnotationLoadBalancerHealthCheckConnectTimeout]
	if healthCheckConnectTimeout != "" {
		connout, err := strconv.Atoi(healthCheckConnectTimeout)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout must be integer, but got [%s], use default number 5. message=[%s]\n",
				healthCheckConnectTimeout, err.Error())
			defaulted.HealthCheckConnectTimeout = 5
		} else {
			defaulted.HealthCheckConnectTimeout = connout
			request.HealthCheckConnectTimeout = defaulted.HealthCheckConnectTimeout
		}
	}

	healthCheckTimeout := annotation[ServiceAnnotationLoadBalancerHealthCheckTimeout]
	if healthCheckTimeout != "" {
		hout, err := strconv.Atoi(annotation[healthCheckTimeout])
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout must be integer, but got [%s], use default number 5. message=[%s]\n",
				healthCheckConnectTimeout, err.Error())
			defaulted.HealthCheckTimeout = 5
		} else {
			defaulted.HealthCheckTimeout = hout
			request.HealthCheckTimeout = defaulted.HealthCheckTimeout
		}
	}

	return defaulted, request
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
			lastClass = class
			if r == '-' {
				continue
			}
			runes = append(runes, []rune{r})
		}
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

func getProtocol(annotation string, port v1.ServicePort) (string, error) {

	if annotation == "" {
		return strings.ToLower(string(port.Protocol)), nil
	}
	for _, v := range strings.Split(annotation, ",") {
		pp := strings.Split(v, ":")
		if len(pp) < 2 {
			return "", errors.New(fmt.Sprintf("port and "+
				"protocol format must be like 'https:443' with colon separated. got=[%+v]", pp))
		}

		if pp[0] != "http" &&
			pp[0] != "tcp" &&
			pp[0] != "https" &&
			pp[0] != "udp" {
			return "", errors.New(fmt.Sprintf("port protocol"+
				" format must be either [http|https|tcp|udp], protocol not supported wit [%s]\n", pp[0]))
		}

		if pp[1] == fmt.Sprintf("%d", port.Port) {
			return pp[0], nil
		}
	}
	return strings.ToLower(string(port.Protocol)), nil
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

func PrettyJson(obj interface{}) string {
	pretty := bytes.Buffer{}
	data, err := json.Marshal(obj)
	if err != nil {
		glog.Errorf("PrettyJson, mashal error: %s\n", err.Error())
		return ""
	}
	err = json.Indent(&pretty, data, "", "    ")

	if err != nil {
		glog.Errorf("PrettyJson, indent error: %s\n", err.Error())
		return ""
	}
	return pretty.String()
}
