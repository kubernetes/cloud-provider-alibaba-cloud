package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"reflect"
	"strings"
)

func PrettyJson(object interface{}) string {
	b, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		fmt.Printf("ERROR: PrettyJson, %v\n %s\n", err, b)
	}
	return string(b)
}

// HashObjects
func HashObjects(slices []interface{}) (string, error) {
	var hashStr string
	for _, item := range slices {
		m := make(map[string]interface{})
		s, err := json.Marshal(item)
		if err != nil {
			return "", fmt.Errorf("hash marshal error: %s", err)
		}
		if err := json.Unmarshal(s, &m); err != nil {
			return "", fmt.Errorf("hash marshal error: %s", err)
		}
		RemoveEmptyValues(m)
		hashStr += PrettyYaml(m)
	}
	return Hash(hashStr), nil
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

func Hash(target string) string {
	hash := sha256.Sum224([]byte(target))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

func PrettyYaml(obj interface{}) string {
	bs, err := yaml.Marshal(obj)
	if err != nil {
		klog.Errorf("failed to parse yaml, ' %s'", err.Error())
	}
	return string(bs)
}

func IsServiceHashChanged(service *v1.Service) (bool, error) {
	if oldHash, ok := service.Labels[LabelServiceHash]; ok {
		newHash, err := GetServiceHash(service)
		if err != nil {
			return true, err
		}
		if strings.Compare(newHash, oldHash) == 0 {
			klog.Infof("service %s/%s hash label not changed, skip", service.Namespace, service.Name)
			return false, nil
		}
	}
	return true, nil
}

func GetServiceHash(service *v1.Service) (string, error) {
	return HashObjects([]interface{}{service.Spec, service.Annotations})
}
