package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
)

const (
	ReconcileHashLable = "alibabacloud.com/reconcile.hash"
)

// HashObject
// Entrance for object computeHash
func HashObject(o interface{}) string {
	data, _ := json.Marshal(o)
	var a interface{}
	err := json.Unmarshal(data, &a)
	if err != nil {
		log.Errorf("unmarshal: %s", err.Error())
	}
	remove(&a)
	return computeHash(PrettyYaml(a))
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
	return computeHash(hashStr), nil
}

// TODO change to remove
func RemoveEmptyValues(m map[string]interface{}) {
	for k, v := range m {
		if subM, ok := v.(map[string]interface{}); ok {
			RemoveEmptyValues(subM)
		}

		if isUnderlyingTypeZero(v) {
			delete(m, k)
		}
	}
}

func HashString(o interface{}) string {
	data, _ := json.Marshal(o)
	var a interface{}
	err := json.Unmarshal(data, &a)
	if err != nil {
		log.Errorf("unmarshal: %s", err.Error())
	}
	remove(&a)
	return PrettyYaml(a)
}

func remove(v *interface{}) {
	o := *v
	switch o.(type) {
	case []interface{}:
		under := o.([]interface{})
		// remove empty object

		for _, m := range under {
			remove(&m)
		}
		var emit []interface{}
		for _, m := range under {
			// remove empty under object
			if isUnderlyingTypeZero(m) {
				continue
			}
			emit = append(emit, m)
		}
		*v = emit
	case map[string]interface{}:
		me := o.(map[string]interface{})
		for k, v := range me {
			if isHashLabel(k) {
				delete(me, k)
				continue
			}
			if isUnderlyingTypeZero(v) {
				delete(me, k)
			} else {
				// continue on next value
				remove(&v)
			}
		}
		*v = o
	default:
	}
}

func isUnderlyingTypeZero(x interface{}) bool {
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

func isHashLabel(k string) bool {
	return k == ReconcileHashLable
}

func PrettyYaml(obj interface{}) string {
	bs, err := yaml.Marshal(obj)
	if err != nil {
		glog.Errorf("failed to parse yaml, ' %s'", err.Error())
	}
	return string(bs)
}

func computeHash(target string) string {
	hash := sha256.Sum224([]byte(target))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}
