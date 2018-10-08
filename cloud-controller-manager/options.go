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

package alicloud

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"bytes"
	"encoding/json"
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
	ServiceAnnotationLoadBalancerMasterZoneID                  = ServiceAnnotationLoadBalancerPrefix + "master-zoneid"
	ServiceAnnotationLoadBalancerSlaveZoneID                   = ServiceAnnotationLoadBalancerPrefix + "slave-zoneid"
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
	ServiceAnnotationLoadBalancerHealthCheckDomain             = ServiceAnnotationLoadBalancerPrefix + "health-check-domain"
	ServiceAnnotationLoadBalancerHealthCheckHTTPCode           = ServiceAnnotationLoadBalancerPrefix + "health-check-httpcode"

	ServiceAnnotationLoadBalancerOverrideListener = ServiceAnnotationLoadBalancerPrefix + "force-override-listeners"

	ServiceAnnotationLoadBalancerSpec               = ServiceAnnotationLoadBalancerPrefix + "spec"
	ServiceAnnotationLoadBalancerSessionStick       = ServiceAnnotationLoadBalancerPrefix + "sticky-session"
	ServiceAnnotationLoadBalancerSessionStickType   = ServiceAnnotationLoadBalancerPrefix + "sticky-session-type"
	ServiceAnnotationLoadBalancerCookieTimeout      = ServiceAnnotationLoadBalancerPrefix + "cookie-timeout"
	ServiceAnnotationLoadBalancerCookie             = ServiceAnnotationLoadBalancerPrefix + "cookie"
	ServiceAnnotationLoadBalancerPersistenceTimeout = ServiceAnnotationLoadBalancerPrefix + "persistence-timeout"
	MagicHealthCheckConnectPort                     = -520

	MAX_LOADBALANCER_BACKEND = 18
)

// defaulted is the parameters which set by programe.
// request represent user defined parameters.
func ExtractAnnotationRequest(service *v1.Service) (*AnnotationRequest, *AnnotationRequest) {
	defaulted, request := &AnnotationRequest{}, &AnnotationRequest{}
	annotation := make(map[string]string)
	for k, v := range service.Annotations {
		annotation[replaceCamel(k)] = v
	}
	bandwidth, ok := annotation[ServiceAnnotationLoadBalancerBandwidth]
	if ok {
		if i, err := strconv.Atoi(bandwidth); err == nil {
			request.Bandwidth = i
			defaulted.Bandwidth = i
		} else {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth"+
				" must be integer, but got [%s], default with no limit. message=[%s]\n",
				bandwidth, err.Error())
			defaulted.Bandwidth = DEFAULT_BANDWIDTH
		}
	} else {
		defaulted.Bandwidth = DEFAULT_BANDWIDTH
	}

	addtype, ok := annotation[ServiceAnnotationLoadBalancerAddressType]
	if ok {
		defaulted.AddressType = slb.AddressType(addtype)
		request.AddressType = defaulted.AddressType
	} else {
		defaulted.AddressType = slb.InternetAddressType
	}

	networkType, ok := annotation[ServiceAnnotationLoadBalancerSLBNetworkType]
	if ok {
		defaulted.SLBNetworkType = networkType
		request.SLBNetworkType = defaulted.SLBNetworkType
	}

	chargtype, ok := annotation[ServiceAnnotationLoadBalancerChargeType]
	if ok {
		defaulted.ChargeType = slb.InternetChargeType(chargtype)
		request.ChargeType = defaulted.ChargeType
	} else {
		defaulted.ChargeType = slb.PayByTraffic
	}

	mzoneid, ok := annotation[ServiceAnnotationLoadBalancerMasterZoneID]
	if ok && mzoneid != "" {
		defaulted.MasterZoneID = mzoneid
		request.MasterZoneID = mzoneid
	}

	szoneid, ok := annotation[ServiceAnnotationLoadBalancerSlaveZoneID]
	if ok && szoneid != "" {
		defaulted.SlaveZoneID = szoneid
		request.SlaveZoneID = szoneid
	}

	lbid, ok := annotation[ServiceAnnotationLoadBalancerId]
	if ok {
		defaulted.Loadbalancerid = lbid
		request.Loadbalancerid = defaulted.Loadbalancerid
	}

	blabel, ok := annotation[ServiceAnnotationLoadBalancerBackendLabel]
	if ok {
		defaulted.BackendLabel = blabel
		request.BackendLabel = defaulted.BackendLabel
	}

	certid, ok := annotation[ServiceAnnotationLoadBalancerCertID]
	if ok {
		defaulted.CertID = certid
		request.CertID = defaulted.CertID
	}

	hcFlag, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckFlag]
	if ok {
		defaulted.HealthCheck = slb.FlagType(hcFlag)
		request.HealthCheck = defaulted.HealthCheck
	} else {
		defaulted.HealthCheck = slb.OffFlag
	}

	hcType, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckType]
	if ok {
		defaulted.HealthCheckType = slb.HealthCheckType(hcType)
		request.HealthCheckType = defaulted.HealthCheckType
	} else {
		defaulted.HealthCheckType = slb.TCPHealthCheckType
	}

	override, ok := annotation[ServiceAnnotationLoadBalancerOverrideListener]
	if ok {
		defaulted.OverrideListeners = override
		request.OverrideListeners = defaulted.OverrideListeners
	} else {
		defaulted.OverrideListeners = "false"
	}

	hcUri, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckURI]
	if ok {
		defaulted.HealthCheckURI = hcUri
		request.HealthCheckURI = defaulted.HealthCheckURI
	}

	healthCheckConnectPort, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckConnectPort]
	if ok {
		port, err := strconv.Atoi(healthCheckConnectPort)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-"+
				"loadbalancer-health-check-connect-port must be integer, but got [%s]. message=[%s]\n",
				healthCheckConnectPort, err.Error())
			//defaulted.HealthCheckConnectPort = MagicHealthCheckConnectPort
		} else {
			defaulted.HealthCheckConnectPort = port
			request.HealthCheckConnectPort = defaulted.HealthCheckConnectPort
		}
	}

	healthCheckHealthyThreshold, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold]
	if ok {
		thresh, err := strconv.Atoi(healthCheckHealthyThreshold)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-"+
				"health-check-healthy-threshold must be integer, but got [%s], use default number 3. message=[%s]\n",
				healthCheckHealthyThreshold, err.Error())
			//defaulted.HealthyThreshold = 3
		} else {
			defaulted.HealthyThreshold = thresh
			request.HealthyThreshold = defaulted.HealthyThreshold
		}
	}

	healthCheckUnhealthyThreshold, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold]
	if ok {
		unThresh, err := strconv.Atoi(healthCheckUnhealthyThreshold)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-"+
				"health-check-unhealthy-threshold must be integer, but got [%s], use default number 3. message=[%s]\n",
				healthCheckUnhealthyThreshold, err.Error())
			//defaulted.UnhealthyThreshold = 3
		} else {
			defaulted.UnhealthyThreshold = unThresh
			request.UnhealthyThreshold = defaulted.UnhealthyThreshold
		}
	}

	healthCheckInterval, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckInterval]
	if ok {
		interval, err := strconv.Atoi(healthCheckInterval)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-"+
				"health-check-interval must be integer, but got [%s], use default number 2. message=[%s]\n",
				healthCheckInterval, err.Error())
			//defaulted.HealthCheckInterval = 2
		} else {
			defaulted.HealthCheckInterval = interval
			request.HealthCheckInterval = defaulted.HealthCheckInterval
		}
	}

	healthCheckConnectTimeout, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckConnectTimeout]
	if ok {
		connout, err := strconv.Atoi(healthCheckConnectTimeout)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-"+
				"health-check-connect-timeout must be integer, but got [%s], use default number 5. message=[%s]\n",
				healthCheckConnectTimeout, err.Error())
			//defaulted.HealthCheckConnectTimeout = 5
		} else {
			defaulted.HealthCheckConnectTimeout = connout
			request.HealthCheckConnectTimeout = defaulted.HealthCheckConnectTimeout
		}
	}

	healthCheckTimeout, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckTimeout]
	if ok {
		hout, err := strconv.Atoi(healthCheckTimeout)
		if err != nil {
			glog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-"+
				"check-timeout must be integer, but got [%s], use default number 5. message=[%s]\n",
				healthCheckConnectTimeout, err.Error())
			//defaulted.HealthCheckTimeout = 5
		} else {
			defaulted.HealthCheckTimeout = hout
			request.HealthCheckTimeout = defaulted.HealthCheckTimeout
		}
	}

	hcDomain, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckDomain]
	if ok {
		defaulted.HealthCheckDomain = hcDomain
		request.HealthCheckDomain = defaulted.HealthCheckDomain
	}

	httpCode, ok := annotation[ServiceAnnotationLoadBalancerHealthCheckHTTPCode]
	if ok {
		defaulted.HealthCheckHttpCode = slb.HealthCheckHttpCodeType(httpCode)
		request.HealthCheckHttpCode = defaulted.HealthCheckHttpCode
	}

	loadbalancerSpec, ok := annotation[ServiceAnnotationLoadBalancerSpec]
	if ok {
		defaulted.LoadBalancerSpec = slb.LoadBalancerSpecType(loadbalancerSpec)
		request.LoadBalancerSpec = defaulted.LoadBalancerSpec
	}

	// stick session
	stickSession, ok := annotation[ServiceAnnotationLoadBalancerSessionStick]
	if ok {
		request.StickySession = slb.FlagType(stickSession)
		defaulted.StickySession = request.StickySession
	} else {
		request.StickySession = slb.FlagType(stickSession)
		defaulted.StickySession = slb.OffFlag
	}

	// stick session type
	stickSessionType, ok := annotation[ServiceAnnotationLoadBalancerSessionStickType]
	if ok {
		defaulted.StickySessionType = slb.StickySessionType(stickSessionType)
		request.StickySessionType = defaulted.StickySessionType
	}

	persistenceTimeout, ok := annotation[ServiceAnnotationLoadBalancerPersistenceTimeout]

	if ok {
		timeout, err := strconv.Atoi(persistenceTimeout)
		if err != nil {
			glog.Warningf("annotation persistence timeout must be integer, but got [%s]. message=[%s]\n",
				persistenceTimeout, err.Error())
			//defaulted.PersistenceTimeout = 0
		} else {
			defaulted.PersistenceTimeout = timeout
			request.PersistenceTimeout = defaulted.PersistenceTimeout
		}
	}
	cookieTimeout, ok := annotation[ServiceAnnotationLoadBalancerCookieTimeout]
	if ok {
		timeout, err := strconv.Atoi(cookieTimeout)
		if err != nil {
			glog.Warningf("annotation persistence timeout must be integer, but got [%s]. message=[%s]\n",
				cookieTimeout, err.Error())
			//defaulted.CookieTimeout = 0
		} else {
			defaulted.CookieTimeout = timeout
			request.CookieTimeout = defaulted.CookieTimeout
		}
	}

	cookie, ok := annotation[ServiceAnnotationLoadBalancerCookie]
	if ok {
		request.Cookie = cookie
		defaulted.Cookie = request.Cookie
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
