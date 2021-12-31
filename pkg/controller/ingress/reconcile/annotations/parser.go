package annotations

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ParseOptions struct {
	exact               bool
	alternativePrefixes []string
}

type ParseOption func(opts *ParseOptions)

func WithExact() ParseOption {
	return func(opts *ParseOptions) {
		opts.exact = true
	}
}

func WithAlternativePrefixes(prefixes ...string) ParseOption {
	return func(opts *ParseOptions) {
		opts.alternativePrefixes = append(opts.alternativePrefixes, prefixes...)
	}
}

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

var (
	// AnnotationsPrefix is the mutable attribute that the controller explicitly refers to
	AnnotationsPrefix = DefaultAnnotationsPrefix
)

// IngressAnnotation has a method to parse annotations located in Ingress
type IngressAnnotation interface {
	Parse(ing *networking.Ingress) (interface{}, error)
}

type ingAnnotations map[string]string

func (a ingAnnotations) parseBool(name string) (bool, error) {
	val, ok := a[name]
	if ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, NewInvalidAnnotationContent(name, val)
		}
		return b, nil
	}
	return false, ErrMissingAnnotations
}

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

func (a ingAnnotations) parseInt(name string) (int, error) {
	val, ok := a[name]
	if ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, NewInvalidAnnotationContent(name, val)
		}
		return i, nil
	}
	return 0, ErrMissingAnnotations
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

// GetBoolAnnotation extracts a boolean from an Ingress annotation
func GetBoolAnnotation(name string, ing *networking.Ingress) (bool, error) {
	v := GetAnnotationWith(name)
	err := checkAnnotation(v, ing)
	if err != nil {
		return false, err
	}
	return ingAnnotations(ing.GetAnnotations()).parseBool(v)
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

// GetIntAnnotation extracts an int from an Ingress annotation
func GetIntAnnotation(name string, ing *networking.Ingress) (int, error) {
	v := GetAnnotationWith(name)
	err := checkAnnotation(v, ing)
	if err != nil {
		return 0, err
	}
	return ingAnnotations(ing.GetAnnotations()).parseInt(v)
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

var configmapAnnotations = sets.NewString(
	"auth-proxy-set-header",
	"fastcgi-params-configmap",
)

// AnnotationsReferencesConfigmap checks if at least one annotation in the Ingress rule
// references a configmap.
func AnnotationsReferencesConfigmap(ing *networking.Ingress) bool {
	if ing == nil || len(ing.GetAnnotations()) == 0 {
		return false
	}

	for name := range ing.GetAnnotations() {
		if configmapAnnotations.Has(name) {
			return true
		}
	}

	return false
}

// StringToURL parses the provided string into URL and returns error
// message in case of failure
func StringToURL(input string) (*url.URL, error) {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("%v is not a valid URL: %v", input, err)
	}

	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("url scheme is empty")
	} else if parsedURL.Host == "" {
		return nil, fmt.Errorf("url host is empty")
	} else if strings.Contains(parsedURL.Host, "..") {
		return nil, fmt.Errorf("invalid url host")
	}

	return parsedURL, nil
}

var (
	// ErrMissingAnnotations the ingress rule does not contain annotations
	// This is an error only when annotations are being parsed
	ErrMissingAnnotations = errors.New("ingress rule without annotations")

	// ErrInvalidAnnotationName the ingress rule does contains an invalid
	// annotation name
	ErrInvalidAnnotationName = errors.New("invalid annotation name")
)

// NewInvalidAnnotationConfiguration returns a new InvalidConfiguration error for use when
// annotations are not correctly configured
func NewInvalidAnnotationConfiguration(name string, reason string) error {
	return InvalidConfiguration{
		Name: fmt.Sprintf("the annotation %v does not contain a valid configuration: %v", name, reason),
	}
}

// NewInvalidAnnotationContent returns a new InvalidContent error
func NewInvalidAnnotationContent(name string, val interface{}) error {
	return InvalidContent{
		Name: fmt.Sprintf("the annotation %v does not contain a valid value (%v)", name, val),
	}
}

// NewLocationDenied returns a new LocationDenied error
func NewLocationDenied(reason string) error {
	return LocationDenied{
		Reason: errors.Errorf("Location denied, reason: %v", reason),
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

// IsLocationDenied checks if the err is an error which
// indicates a location should return HTTP code 503
func IsLocationDenied(e error) bool {
	_, ok := e.(LocationDenied)
	return ok
}

// IsMissingAnnotations checks if the err is an error which
// indicates the ingress does not contain annotations
func IsMissingAnnotations(e error) bool {
	return e == ErrMissingAnnotations
}

// IsInvalidContent checks if the err is an error which
// indicates an annotations value is not valid
func IsInvalidContent(e error) bool {
	_, ok := e.(InvalidContent)
	return ok
}

// New returns a new error
func New(m string) error {
	return errors.New(m)
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}
