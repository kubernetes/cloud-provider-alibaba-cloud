package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
)

// HashObject
// Entrance for object hash
func HashObject(spec interface{}) (string, error) {
	m := make(map[string]interface{})
	s, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("hash marshal error: %s", err)
	}
	if err := json.Unmarshal(s, &m); err != nil {
		return "", fmt.Errorf("hash marshal error: %s", err)
	}
	RemoveEmptyValues(m)
	return hash(PrettyYaml(m)), nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	if x == nil {
		return true
	}
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	}

	zero := reflect.Zero(reflect.TypeOf(x)).Interface()
	return reflect.DeepEqual(x, zero)
}

func RemoveEmptyValues(m map[string]interface{}) {
	for k, v := range m {
		if subM, ok := v.(map[string]interface{}); ok {
			RemoveEmptyValues(subM)
		}

		if isZeroOfUnderlyingType(v) {
			delete(m, k)
		}
	}
}

func PrettyYaml(obj interface{}) string {
	bs, err := yaml.Marshal(obj)
	if err != nil {
		glog.Errorf("failed to parse yaml, ' %s'", err.Error())
	}
	return string(bs)
}

func hash(target string) string {
	hash := sha256.Sum224([]byte(target))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}
