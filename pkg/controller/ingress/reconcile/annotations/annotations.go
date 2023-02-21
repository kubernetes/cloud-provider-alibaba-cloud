package annotations

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	networking "k8s.io/api/networking/v1"
)

// prefix
const (
	// DefaultAnnotationsPrefix defines the common prefix used in the nginx ingress controller
	DefaultAnnotationsPrefix = "alb.ingress.kubernetes.io"
)

// load balancer annotations
const (
	// AnnotationLoadBalancerPrefix loadbalancer prefix
	AnnotationLoadBalancerPrefix = "alb.ingress.kubernetes.io/"
	// config the rule conditions
	INGRESS_ALB_CONDITIONS_ANNOTATIONS = "alb.ingress.kubernetes.io/conditions.%s"
	// config the rule actions
	INGRESS_ALB_ACTIONS_ANNOTATIONS = "alb.ingress.kubernetes.io/actions.%s"
	// config the rule direction
	INGRESS_ALB_RULE_DIRECTION = "alb.ingress.kubernetes.io/rule-direction.%s"
	// Load Balancer Attribute
	AddressType            = AnnotationLoadBalancerPrefix + "address-type"             // AddressType loadbalancer address type
	VswitchIds             = AnnotationLoadBalancerPrefix + "vswitch-ids"              // VswitchId loadbalancer vswitch id
	ChargeType             = AnnotationLoadBalancerPrefix + "charge-type"              // ChargeType lb charge type
	LoadBalancerId         = AnnotationLoadBalancerPrefix + "id"                       // LoadBalancerId lb id
	OverrideListener       = AnnotationLoadBalancerPrefix + "force-override-listeners" // OverrideListener force override listeners
	LoadBalancerName       = AnnotationLoadBalancerPrefix + "name"                     // LoadBalancerName slb name
	AddressAllocatedMode   = AnnotationLoadBalancerPrefix + "address-allocated-mode"
	LoadBalancerEdition    = AnnotationLoadBalancerPrefix + "edition"
	ListenPorts            = AnnotationLoadBalancerPrefix + "listen-ports"
	AccessLog              = AnnotationLoadBalancerPrefix + "access-log"
	HealthCheckEnabled     = AnnotationLoadBalancerPrefix + "healthcheck-enabled"          // HealthCheckFlag health check flag
	HealthCheckPath        = AnnotationLoadBalancerPrefix + "healthcheck-path"             // HealthCheckType health check type
	HealthCheckProtocol    = AnnotationLoadBalancerPrefix + "healthcheck-protocol"         // HealthCheckConnectPort health check connect port
	HealthCheckConnectPort = AnnotationLoadBalancerPrefix + "healthcheck-connect-port"     // HealthCheckConnectPort health check connect port
	HealthCheckMethod      = AnnotationLoadBalancerPrefix + "healthcheck-method"           // HealthyThreshold health check healthy thresh hold
	HealthCheckInterval    = AnnotationLoadBalancerPrefix + "healthcheck-interval-seconds" // HealthCheckInterval health check interval
	HealthCheckTimeout     = AnnotationLoadBalancerPrefix + "healthcheck-timeout-seconds"  // HealthCheckTimeout health check timeout
	HealthThreshold        = AnnotationLoadBalancerPrefix + "healthy-threshold-count"      // HealthCheckDomain health check domain
	UnHealthThreshold      = AnnotationLoadBalancerPrefix + "unhealthy-threshold-count"    // HealthCheckHTTPCode health check http code
	HealthCheckHTTPCode    = AnnotationLoadBalancerPrefix + "healthcheck-httpcode"
	Order                  = AnnotationLoadBalancerPrefix + "order"
	// VServerBackend Attribute
	BackendLabel      = AnnotationLoadBalancerPrefix + "backend-label"              // BackendLabel backend labels
	BackendType       = "service.beta.kubernetes.io/backend-type"                   // BackendType backend type
	RemoveUnscheduled = AnnotationLoadBalancerPrefix + "remove-unscheduled-backend" // RemoveUnscheduled remove unscheduled node from backends
)

const (
	AnnotationNginxPrefix    = "nginx.ingress.kubernetes.io/"
	NginxCanary              = AnnotationNginxPrefix + "canary"
	NginxCanaryByHeader      = AnnotationNginxPrefix + "canary-by-header"
	NginxCanaryByHeaderValue = AnnotationNginxPrefix + "canary-by-header-value"
	NginxCanaryByCookie      = AnnotationNginxPrefix + "canary-by-cookie"
	NginxCanaryWeight        = AnnotationNginxPrefix + "canary-weight"
	NginxSslRedirect         = AnnotationNginxPrefix + "ssl-redirect"

	AnnotationAlbPrefix = "alb.ingress.kubernetes.io/"
	SessionStick        = AnnotationLoadBalancerPrefix + "sticky-session"      // SessionStick sticky session
	SessionStickType    = AnnotationLoadBalancerPrefix + "sticky-session-type" // SessionStickType session sticky type
	CookieTimeout       = AnnotationLoadBalancerPrefix + "cookie-timeout"      // CookieTimeout cookie timeout
	Cookie              = AnnotationLoadBalancerPrefix + "cookie"              // Cookie lb cookie

	AlbCanary               = AnnotationAlbPrefix + "canary"
	AlbCanaryByHeader       = AnnotationAlbPrefix + "canary-by-header"
	AlbCanaryByHeaderValue  = AnnotationAlbPrefix + "canary-by-header-value"
	AlbCanaryByCookie       = AnnotationAlbPrefix + "canary-by-cookie"
	AlbCanaryWeight         = AnnotationAlbPrefix + "canary-weight"
	AlbSslRedirect          = AnnotationAlbPrefix + "ssl-redirect"
	AlbRewriteTarget        = AnnotationAlbPrefix + "rewrite-target"
	AlbBackendProtocol      = AnnotationAlbPrefix + "backend-protocol"
	AlbBackendScheduler     = AnnotationAlbPrefix + "backend-scheduler"
	AlbTrafficLimitQps      = AnnotationAlbPrefix + "traffic-limit-qps"
	AlbEnableCors           = AnnotationAlbPrefix + "enable-cors"
	AlbCorsAllowOrigin      = AnnotationAlbPrefix + "cors-allow-origin"
	AlbCorsAllowMethods     = AnnotationAlbPrefix + "cors-allow-methods"
	AlbCorsAllowHeaders     = AnnotationAlbPrefix + "cors-allow-headers"
	AlbCorsExposeHeaders    = AnnotationAlbPrefix + "cors-expose-headers"
	AlbCorsAllowCredentials = AnnotationAlbPrefix + "cors-allow-credentials"
	AlbCorsMaxAge           = AnnotationAlbPrefix + "cors-max-age"
)

type ParseOptions struct {
	exact               bool
	alternativePrefixes []string
}

type ParseOption func(opts *ParseOptions)

type Parser interface {
	ParseStringAnnotation(annotation string, value *string, annotations map[string]string, opts ...ParseOption) bool

	ParseBoolAnnotation(annotation string, value *bool, annotations map[string]string, opts ...ParseOption) (bool, error)

	ParseInt64Annotation(annotation string, value *int64, annotations map[string]string, opts ...ParseOption) (bool, error)

	ParseStringSliceAnnotation(annotation string, value *[]string, annotations map[string]string, opts ...ParseOption) bool

	ParseJSONAnnotation(annotation string, value interface{}, annotations map[string]string, opts ...ParseOption) (bool, error)

	ParseStringMapAnnotation(annotation string, value *map[string]string, annotations map[string]string, opts ...ParseOption) (bool, error)
}

func NewSuffixAnnotationParser(annotationPrefix string) *suffixAnnotationParser {
	return &suffixAnnotationParser{
		annotationPrefix: annotationPrefix,
	}
}

var _ Parser = (*suffixAnnotationParser)(nil)

type suffixAnnotationParser struct {
	annotationPrefix string
}

func (p *suffixAnnotationParser) ParseStringAnnotation(annotation string, value *string, annotations map[string]string, opts ...ParseOption) bool {
	ret, _ := p.parseStringAnnotation(annotation, value, annotations, opts...)
	return ret
}

func (p *suffixAnnotationParser) ParseBoolAnnotation(annotation string, value *bool, annotations map[string]string, opts ...ParseOption) (bool, error) {
	raw := ""
	exists, matchedKey := p.parseStringAnnotation(annotation, &raw, annotations, opts...)
	if !exists {
		return false, nil
	}
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return true, errors.Wrapf(err, "failed to parse bool annotation, %v: %v", matchedKey, raw)
	}
	*value = val
	return true, nil
}

func (p *suffixAnnotationParser) ParseInt64Annotation(annotation string, value *int64, annotations map[string]string, opts ...ParseOption) (bool, error) {
	raw := ""
	exists, matchedKey := p.parseStringAnnotation(annotation, &raw, annotations, opts...)
	if !exists {
		return false, nil
	}
	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return true, errors.Wrapf(err, "failed to parse int64 annotation, %v: %v", matchedKey, raw)
	}
	*value = i
	return true, nil
}

func (p *suffixAnnotationParser) ParseStringSliceAnnotation(annotation string, value *[]string, annotations map[string]string, opts ...ParseOption) bool {
	raw := ""
	if exists, _ := p.parseStringAnnotation(annotation, &raw, annotations, opts...); !exists {
		return false
	}
	*value = splitCommaSeparatedString(raw)
	return true
}

func (p *suffixAnnotationParser) ParseJSONAnnotation(annotation string, value interface{}, annotations map[string]string, opts ...ParseOption) (bool, error) {
	raw := ""
	exists, matchedKey := p.parseStringAnnotation(annotation, &raw, annotations, opts...)
	if !exists {
		return false, nil
	}
	if err := json.Unmarshal([]byte(raw), value); err != nil {
		return true, errors.Wrapf(err, "failed to parse json annotation, %v: %v", matchedKey, raw)
	}
	return true, nil
}

func (p *suffixAnnotationParser) ParseStringMapAnnotation(annotation string, value *map[string]string, annotations map[string]string, opts ...ParseOption) (bool, error) {
	raw := ""
	exists, matchedKey := p.parseStringAnnotation(annotation, &raw, annotations, opts...)
	if !exists {
		return false, nil
	}
	rawKVPairs := splitCommaSeparatedString(raw)
	keyValues := make(map[string]string)
	for _, kvPair := range rawKVPairs {
		parts := strings.SplitN(kvPair, "=", 2)
		if len(parts) != 2 {
			return false, errors.Errorf("failed to parse stringMap annotation, %v: %v", matchedKey, raw)
		}
		key := parts[0]
		value := parts[1]
		if len(key) == 0 {
			return false, errors.Errorf("failed to parse stringMap annotation, %v: %v", matchedKey, raw)
		}
		keyValues[key] = value
	}
	if value != nil {
		*value = keyValues
	}
	return true, nil
}

func (p *suffixAnnotationParser) parseStringAnnotation(annotation string, value *string, annotations map[string]string, opts ...ParseOption) (bool, string) {
	keys := p.buildAnnotationKeys(annotation, opts...)
	for _, key := range keys {
		if raw, ok := annotations[key]; ok {
			*value = raw
			return true, key
		}
	}
	return false, ""
}

// buildAnnotationKey returns list of full annotation keys based on suffix and parse options
func (p *suffixAnnotationParser) buildAnnotationKeys(suffix string, opts ...ParseOption) []string {
	keys := []string{}
	parseOpts := ParseOptions{}
	for _, opt := range opts {
		opt(&parseOpts)
	}
	if parseOpts.exact {
		keys = append(keys, suffix)
	} else {
		keys = append(keys, fmt.Sprintf("%v/%v", p.annotationPrefix, suffix))
		for _, pfx := range parseOpts.alternativePrefixes {
			keys = append(keys, fmt.Sprintf("%v/%v", pfx, suffix))
		}
	}
	return keys
}

func splitCommaSeparatedString(commaSeparatedString string) []string {
	var result []string
	parts := strings.Split(commaSeparatedString, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		result = append(result, part)
	}
	return result
}

// IngressAnnotation has a method to parse annotations located in Ingress
type IngressAnnotation interface {
	Parse(ing *networking.Ingress) (interface{}, error)
}

type ingAnnotations map[string]string

func (a ingAnnotations) parseString(name string) (string, error) {
	val, ok := a[name]
	if ok {
		s := normalizeString(val)
		if len(s) == 0 {
			return "", NewInvalidAnnotationContent(name, val)
		}

		return s, nil
	}
	return "", ErrMissingAnnotations
}

func checkAnnotation(name string, ing *networking.Ingress) error {
	if ing == nil || len(ing.GetAnnotations()) == 0 {
		return ErrMissingAnnotations
	}
	if name == "" {
		return ErrInvalidAnnotationName
	}

	return nil
}

// GetStringAnnotation extracts a string from an Ingress annotation
func GetStringAnnotation(name string, ing *networking.Ingress) (string, error) {
	v := GetAnnotationWith(name)
	err := checkAnnotation(v, ing)
	if err != nil {
		return "", err
	}

	return ingAnnotations(ing.GetAnnotations()).parseString(v)
}

// GetStringAnnotationMutil extracts a string from an Ingress annotation
func GetStringAnnotationMutil(name, name1 string, ing *networking.Ingress) string {
	if val, ok := ing.Annotations[name]; ok {
		return val
	}
	if val, ok := ing.Annotations[name1]; ok {
		return val
	}
	return ""
}

// GetAnnotationWith returns the ingress annotations
func GetAnnotationWith(ann string) string {
	return fmt.Sprintf("%v", ann)
}

func normalizeString(input string) string {
	trimmedContent := []string{}
	for _, line := range strings.Split(input, "\n") {
		trimmedContent = append(trimmedContent, strings.TrimSpace(line))
	}

	return strings.Join(trimmedContent, "\n")
}

var (
	// ErrMissingAnnotations the ingress rule does not contain annotations
	// This is an error only when annotations are being parsed
	ErrMissingAnnotations = errors.New("ingress rule without annotations")

	// ErrInvalidAnnotationName the ingress rule does contains an invalid
	// annotation name
	ErrInvalidAnnotationName = errors.New("invalid annotation name")
)

// NewInvalidAnnotationContent returns a new InvalidContent error
func NewInvalidAnnotationContent(name string, val interface{}) error {
	return InvalidContent{
		Name: fmt.Sprintf("the annotation %v does not contain a valid value (%v)", name, val),
	}
}

// InvalidConfiguration Error
type InvalidConfiguration struct {
	Name string
}

func (e InvalidConfiguration) Error() string {
	return e.Name
}

// InvalidContent error
type InvalidContent struct {
	Name string
}

func (e InvalidContent) Error() string {
	return e.Name
}

// LocationDenied error
type LocationDenied struct {
	Reason error
}

func (e LocationDenied) Error() string {
	return e.Reason.Error()
}
