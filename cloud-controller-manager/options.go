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
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"k8s.io/klog"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"bytes"
	"encoding/json"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/api/core/v1"
)

const (
	// ServiceAnnotationPrefix prefix of service annotation
	ServiceAnnotationPrefix = "service.beta.kubernetes.io/alibaba-cloud-"

	// ServiceAnnotationLegacyPrefix legacy prefix of service annotation
	ServiceAnnotationLegacyPrefix = "service.beta.kubernetes.io/alicloud-"

	// ServiceAnnotationLoadBalancerPrefix loadbalancer prefix
	ServiceAnnotationLoadBalancerPrefix = ServiceAnnotationPrefix + "loadbalancer-"

	// ServiceAnnotationPrivateZonePrefix private zone prefix
	ServiceAnnotationPrivateZonePrefix = ServiceAnnotationPrefix + "private-zone-"

	// ServiceAnnotationLoadBalancerAclStatus enable or disable acl on all listener
	ServiceAnnotationLoadBalancerAclStatus = ServiceAnnotationLoadBalancerPrefix + "acl-status"

	// ServiceAnnotationLoadBalancerAclID acl id
	ServiceAnnotationLoadBalancerAclID = ServiceAnnotationLoadBalancerPrefix + "acl-id"

	// ServiceAnnotationLoadBalancerAclType acl type, black or white
	ServiceAnnotationLoadBalancerAclType = ServiceAnnotationLoadBalancerPrefix + "acl-type"

	// ServiceAnnotationLoadBalancerProtocolPort protocol port
	ServiceAnnotationLoadBalancerProtocolPort = ServiceAnnotationLoadBalancerPrefix + "protocol-port"

	// ServiceAnnotationLoadBalancerAddressType loadbalancer address type
	ServiceAnnotationLoadBalancerAddressType = ServiceAnnotationLoadBalancerPrefix + "address-type"

	// ServiceAnnotationLoadBalancerVswitch loadbalancer vswitch id
	ServiceAnnotationLoadBalancerVswitch = ServiceAnnotationLoadBalancerPrefix + "vswitch-id"

	// ServiceAnnotationLoadBalancerForwardPort loadbalancer forward port
	ServiceAnnotationLoadBalancerForwardPort = ServiceAnnotationLoadBalancerPrefix + "forward-port"

	// ServiceAnnotationLoadBalancerSLBNetworkType loadbalancer network type
	ServiceAnnotationLoadBalancerSLBNetworkType = ServiceAnnotationLoadBalancerPrefix + "slb-network-type"
	// ServiceAnnotationLoadBalancerChargeType lb charge type
	ServiceAnnotationLoadBalancerChargeType = ServiceAnnotationLoadBalancerPrefix + "charge-type"

	// ServiceAnnotationLoadBalancerId lb id
	ServiceAnnotationLoadBalancerId = ServiceAnnotationLoadBalancerPrefix + "id"

	//ServiceAnnotationLoadBalancerName slb name
	ServiceAnnotationLoadBalancerName = ServiceAnnotationLoadBalancerPrefix + "name"

	// ServiceAnnotationLoadBalancerBackendLabel backend labels
	ServiceAnnotationLoadBalancerBackendLabel = ServiceAnnotationLoadBalancerPrefix + "backend-label"

	// ServiceAnnotationLoadBalancerRegion region
	ServiceAnnotationLoadBalancerRegion = ServiceAnnotationLoadBalancerPrefix + "region"

	// ServiceAnnotationLoadBalancerMasterZoneID master zone id
	ServiceAnnotationLoadBalancerMasterZoneID = ServiceAnnotationLoadBalancerPrefix + "master-zoneid"

	// ServiceAnnotationLoadBalancerSlaveZoneID slave zone id
	ServiceAnnotationLoadBalancerSlaveZoneID = ServiceAnnotationLoadBalancerPrefix + "slave-zoneid"

	// ServiceAnnotationLoadBalancerBandwidth bandwidth
	ServiceAnnotationLoadBalancerBandwidth = ServiceAnnotationLoadBalancerPrefix + "bandwidth"

	// ServiceAnnotationLoadBalancerCertID cert id
	ServiceAnnotationLoadBalancerCertID = ServiceAnnotationLoadBalancerPrefix + "cert-id"

	// ServiceAnnotationLoadBalancerHealthCheckFlag health check flag
	ServiceAnnotationLoadBalancerHealthCheckFlag = ServiceAnnotationLoadBalancerPrefix + "health-check-flag"

	// ServiceAnnotationLoadBalancerHealthCheckType health check type
	ServiceAnnotationLoadBalancerHealthCheckType = ServiceAnnotationLoadBalancerPrefix + "health-check-type"

	// ServiceAnnotationLoadBalancerHealthCheckURI health check uri
	ServiceAnnotationLoadBalancerHealthCheckURI = ServiceAnnotationLoadBalancerPrefix + "health-check-uri"

	// ServiceAnnotationLoadBalancerHealthCheckConnectPort health check connect port
	ServiceAnnotationLoadBalancerHealthCheckConnectPort = ServiceAnnotationLoadBalancerPrefix + "health-check-connect-port"

	// ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold health check healthy thresh hold
	ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold = ServiceAnnotationLoadBalancerPrefix + "healthy-threshold"

	// ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold health check unhealthy thresh hold
	ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold = ServiceAnnotationLoadBalancerPrefix + "unhealthy-threshold"

	// ServiceAnnotationLoadBalancerHealthCheckInterval health check interval
	ServiceAnnotationLoadBalancerHealthCheckInterval = ServiceAnnotationLoadBalancerPrefix + "health-check-interval"

	// ServiceAnnotationLoadBalancerHealthCheckConnectTimeout health check connect timeout
	ServiceAnnotationLoadBalancerHealthCheckConnectTimeout = ServiceAnnotationLoadBalancerPrefix + "health-check-connect-timeout"

	// ServiceAnnotationLoadBalancerHealthCheckTimeout health check timeout
	ServiceAnnotationLoadBalancerHealthCheckTimeout = ServiceAnnotationLoadBalancerPrefix + "health-check-timeout"

	// ServiceAnnotationLoadBalancerHealthCheckDomain health check domain
	ServiceAnnotationLoadBalancerHealthCheckDomain = ServiceAnnotationLoadBalancerPrefix + "health-check-domain"

	// ServiceAnnotationLoadBalancerHealthCheckHTTPCode health check http code
	ServiceAnnotationLoadBalancerHealthCheckHTTPCode = ServiceAnnotationLoadBalancerPrefix + "health-check-httpcode"

	// ServiceAnnotationLoadBalancerAdditionalTags For example: "Key1=Val1,Key2=Val2,KeyNoVal1=,KeyNoVal2",same with aws
	ServiceAnnotationLoadBalancerAdditionalTags = ServiceAnnotationLoadBalancerPrefix + "additional-resource-tags"

	// ServiceAnnotationLoadBalancerOverrideListener force override listeners
	ServiceAnnotationLoadBalancerOverrideListener = ServiceAnnotationLoadBalancerPrefix + "force-override-listeners"

	// ServiceAnnotationLoadBalancerSpec slb spec
	ServiceAnnotationLoadBalancerSpec = ServiceAnnotationLoadBalancerPrefix + "spec"

	// ServiceAnnotationLoadBalancerScheduler slb scheduler
	ServiceAnnotationLoadBalancerScheduler = ServiceAnnotationLoadBalancerPrefix + "scheduler"

	// ServiceAnnotationLoadBalancerSessionStick sticky session
	ServiceAnnotationLoadBalancerSessionStick = ServiceAnnotationLoadBalancerPrefix + "sticky-session"

	// ServiceAnnotationLoadBalancerSessionStickType session sticky type
	ServiceAnnotationLoadBalancerSessionStickType = ServiceAnnotationLoadBalancerPrefix + "sticky-session-type"

	// ServiceAnnotationLoadBalancerCookieTimeout cookie timeout
	ServiceAnnotationLoadBalancerCookieTimeout = ServiceAnnotationLoadBalancerPrefix + "cookie-timeout"

	//ServiceAnnotationLoadBalancerCookie lb cookie
	ServiceAnnotationLoadBalancerCookie = ServiceAnnotationLoadBalancerPrefix + "cookie"

	// ServiceAnnotationLoadBalancerPersistenceTimeout persistence timeout
	ServiceAnnotationLoadBalancerPersistenceTimeout = ServiceAnnotationLoadBalancerPrefix + "persistence-timeout"
	//MagicHealthCheckConnectPort                     = -520

	//ServiceAnnotationLoadBalancerIPVersion ip version
	ServiceAnnotationLoadBalancerIPVersion = ServiceAnnotationLoadBalancerPrefix + "ip-version"

	// MAX_LOADBALANCER_BACKEND max default lb backend count.
	MAX_LOADBALANCER_BACKEND = 18

	// ServiceAnnotationLoadBalancerPrivateZoneName private zone name
	ServiceAnnotationLoadBalancerPrivateZoneName = ServiceAnnotationPrivateZonePrefix + "name"

	// ServiceAnnotationLoadBalancerPrivateZoneId private zone id
	ServiceAnnotationLoadBalancerPrivateZoneId = ServiceAnnotationPrivateZonePrefix + "id"

	// ServiceAnnotationLoadBalancerPrivateZoneRecordName private zone record name
	ServiceAnnotationLoadBalancerPrivateZoneRecordName = ServiceAnnotationPrivateZonePrefix + "record-name"

	// ServiceAnnotationLoadBalancerPrivateZoneRecordTTL private zone record ttl
	ServiceAnnotationLoadBalancerPrivateZoneRecordTTL = ServiceAnnotationPrivateZonePrefix + "record-ttl"

	// ServiceAnnotationLoadBalancerBackendType backend type
	ServiceAnnotationLoadBalancerBackendType = utils.BACKEND_TYPE_LABEL

	// ServiceAnnotationLoadBalancerResourceGroupId resource group id
	ServiceAnnotationLoadBalancerResourceGroupId = ServiceAnnotationLoadBalancerPrefix + "resource-group-id"

	// ServiceAnnotationLoadBalancerDeleteProtection delete protection
	ServiceAnnotationLoadBalancerDeleteProtection = ServiceAnnotationLoadBalancerPrefix + "delete-protection"

	// ServiceAnnotationLoadBalancerModificationProtection modification type
	ServiceAnnotationLoadBalancerModificationProtection = ServiceAnnotationLoadBalancerPrefix + "modification-protection"

	// ServiceAnnotationLoadBalancerBackendType external ip type
	ServiceAnnotationLoadBalancerExternalIPType = ServiceAnnotationLoadBalancerPrefix + "external-ip-type"
)

type ExternalIPType string

const (
	EIPExternalIPType = ExternalIPType("eip")
)

//compatible to old camel annotation
func getBackwardsCompatibleAnnotation(annotations map[string]string) map[string]string {
	newAnnotation := make(map[string]string)
	for k, v := range annotations {
		newAnnotation[replaceCamel(normalizePrefix(k))] = v
	}
	return newAnnotation
}

// ExtractAnnotationRequest  extract annotations from service labels
// defaulted is the parameters which set by programe.
// request represent user defined parameters.
func ExtractAnnotationRequest(service *v1.Service) (*AnnotationRequest, *AnnotationRequest) {
	defaulted, request := &AnnotationRequest{}, &AnnotationRequest{}
	annotation := getBackwardsCompatibleAnnotation(service.Annotations)
	bandwidth, ok := annotation[ServiceAnnotationLoadBalancerBandwidth]
	if ok {
		if i, err := strconv.Atoi(bandwidth); err == nil {
			request.Bandwidth = i
			defaulted.Bandwidth = i
		} else {
			klog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth"+
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

	vswid, ok := annotation[ServiceAnnotationLoadBalancerVswitch]
	if ok {
		defaulted.VswitchID = vswid
		request.VswitchID = defaulted.VswitchID
	}

	status, ok := annotation[ServiceAnnotationLoadBalancerAclStatus]
	if ok {
		defaulted.AclStatus = status
		request.AclStatus = defaulted.AclStatus
	} else {
		defaulted.AclStatus = "off"
	}

	aclid, ok := annotation[ServiceAnnotationLoadBalancerAclID]
	if ok {
		defaulted.AclID = aclid
		request.AclID = defaulted.AclID
	}
	acltype, ok := annotation[ServiceAnnotationLoadBalancerAclType]
	if ok {
		defaulted.AclType = acltype
		request.AclType = defaulted.AclType
	}

	forward, ok := annotation[ServiceAnnotationLoadBalancerForwardPort]
	if ok {
		defaulted.ForwardPort = forward
		request.ForwardPort = defaulted.ForwardPort
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

	lbName, ok := annotation[ServiceAnnotationLoadBalancerName]
	if ok {
		defaulted.LoadBalancerName = lbName
		request.LoadBalancerName = defaulted.LoadBalancerName
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
			klog.Warningf("annotation service.beta.kubernetes.io/alicloud-"+
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
			klog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-"+
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
			klog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-"+
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
			klog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-"+
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
			klog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-"+
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
			klog.Warningf("annotation service.beta.kubernetes.io/alicloud-loadbalancer-health-"+
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
	} else {
		defaulted.LoadBalancerSpec = "slb.s1.small"
	}

	scheduler, ok := annotation[ServiceAnnotationLoadBalancerScheduler]
	if ok {
		defaulted.Scheduler = scheduler
		request.Scheduler = defaulted.Scheduler
	} else {
		defaulted.Scheduler = "rr"
	}

	// stick session
	stickSession, ok := annotation[ServiceAnnotationLoadBalancerSessionStick]
	if ok {
		request.StickySession = slb.FlagType(stickSession)
		defaulted.StickySession = request.StickySession
	} else {
		request.StickySession = slb.OffFlag
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
			klog.Warningf("annotation persistence timeout must be integer, but got [%s]. message=[%s]\n",
				persistenceTimeout, err.Error())
			//defaulted.PersistenceTimeout = 0
		} else {
			defaulted.PersistenceTimeout = &timeout
			request.PersistenceTimeout = defaulted.PersistenceTimeout
		}
	}
	cookieTimeout, ok := annotation[ServiceAnnotationLoadBalancerCookieTimeout]
	if ok {
		timeout, err := strconv.Atoi(cookieTimeout)
		if err != nil {
			klog.Warningf("annotation persistence timeout must be integer, but got [%s]. message=[%s]\n",
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

	ipVersion, ok := annotation[ServiceAnnotationLoadBalancerIPVersion]
	if ok {
		request.AddressIPVersion = slb.AddressIPVersionType(ipVersion)
		defaulted.AddressIPVersion = request.AddressIPVersion
	}

	privateZoneName, ok := annotation[ServiceAnnotationLoadBalancerPrivateZoneName]
	if ok {
		request.PrivateZoneName = privateZoneName
		defaulted.PrivateZoneName = request.PrivateZoneName
	}

	privateZoneId, ok := annotation[ServiceAnnotationLoadBalancerPrivateZoneId]
	if ok {
		request.PrivateZoneId = privateZoneId
		defaulted.PrivateZoneId = request.PrivateZoneId
	}

	privateZoneRecordName, ok := annotation[ServiceAnnotationLoadBalancerPrivateZoneRecordName]
	if ok {
		request.PrivateZoneRecordName = privateZoneRecordName
		defaulted.PrivateZoneRecordName = request.PrivateZoneRecordName
	}

	privateZoneRecordTTL, ok := annotation[ServiceAnnotationLoadBalancerPrivateZoneRecordTTL]
	if ok {
		ttl, err := strconv.Atoi(privateZoneRecordTTL)
		if err != nil {
			klog.Warningf("annotation "+ServiceAnnotationLoadBalancerPrivateZoneRecordTTL+
				" must be integer, but got [%s], use default number 60. message=[%s]\n",
				privateZoneRecordTTL, err.Error())
			defaulted.PrivateZoneRecordTTL = 60
		} else {
			defaulted.PrivateZoneRecordTTL = ttl
			request.PrivateZoneRecordTTL = defaulted.PrivateZoneRecordTTL
		}
	}

	backendType, ok := annotation[ServiceAnnotationLoadBalancerBackendType]
	if ok {
		request.BackendType = backendType
		defaulted.BackendType = request.BackendType
	} else {
		defaulted.BackendType = utils.BACKEND_TYPE_ECS
		request.BackendType = defaulted.BackendType
	}

	removeUnscheduledBackend, ok := annotation[utils.ServiceAnnotationLoadBalancerRemoveUnscheduledBackend]
	if ok {
		request.RemoveUnscheduledBackend = removeUnscheduledBackend
		defaulted.RemoveUnscheduledBackend = request.RemoveUnscheduledBackend
	} else {
		defaulted.RemoveUnscheduledBackend = "off"
		request.RemoveUnscheduledBackend = defaulted.RemoveUnscheduledBackend
	}

	resourceGroupId, ok := annotation[ServiceAnnotationLoadBalancerResourceGroupId]
	if ok {
		request.ResourceGroupId = resourceGroupId
		defaulted.ResourceGroupId = request.ResourceGroupId
	}

	delProtection, ok := annotation[ServiceAnnotationLoadBalancerDeleteProtection]
	if ok {
		defaulted.DeleteProtection = slb.FlagType(delProtection)
		request.DeleteProtection = defaulted.DeleteProtection
	} else {
		defaulted.DeleteProtection = slb.OnFlag
	}

	modificationProtection, ok := annotation[ServiceAnnotationLoadBalancerModificationProtection]
	if ok {
		request.ModificationProtectionStatus = slb.ModificationProtectionType(modificationProtection)
		defaulted.ModificationProtectionStatus = request.ModificationProtectionStatus
	} else {
		defaulted.ModificationProtectionStatus = slb.ConsoleProtection
	}

	externalIpType, ok := annotation[ServiceAnnotationLoadBalancerExternalIPType]
	if ok {
		request.ExternalIPType = externalIpType
		defaulted.ExternalIPType = request.ExternalIPType
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
		if annotate == replaceCamel(normalizePrefix(k)) {
			return v
		}
	}
	return ""
}

// Change the legacy prefix 'alicloud' to 'alibaba-cloud'
func normalizePrefix(str string) string {
	if !strings.HasPrefix(str, ServiceAnnotationLegacyPrefix) {
		// If it is not start with ServiceAnnotationLegacyPrefix, just skip
		return str
	}

	return ServiceAnnotationPrefix + str[len(ServiceAnnotationLegacyPrefix):]
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

// PrettyJson  pretty json output
func PrettyJson(obj interface{}) string {
	pretty := bytes.Buffer{}
	data, err := json.Marshal(obj)
	if err != nil {
		klog.Errorf("PrettyJson, mashal error: %s\n", err.Error())
		return ""
	}
	err = json.Indent(&pretty, data, "", "    ")

	if err != nil {
		klog.Errorf("PrettyJson, indent error: %s\n", err.Error())
		return ""
	}
	return pretty.String()
}
