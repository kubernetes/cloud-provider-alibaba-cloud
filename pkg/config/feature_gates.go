package config

import (
	"fmt"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/component-base/featuregate"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
)

const (
	IPv6DualStack featuregate.Feature = "IPv6DualStack"
	EndpointSlice featuregate.Feature = "EndpointSlice"
)

var CloudProviderFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	IPv6DualStack: {Default: false, PreRelease: featuregate.Alpha},
	EndpointSlice: {Default: true, PreRelease: featuregate.GA},
}

func init() {
	runtime.Must(utilfeature.DefaultMutableFeatureGate.Add(CloudProviderFeatureGates))
}

func BindFeatureGates(client *apiext.Clientset, features string) error {
	m := make(map[string]bool)
	for _, s := range strings.Split(features, ",") {
		if len(s) == 0 {
			continue
		}
		arr := strings.SplitN(s, "=", 2)
		k := strings.TrimSpace(arr[0])
		if len(arr) != 2 {
			return fmt.Errorf("missing bool value for %s", k)
		}
		v := strings.TrimSpace(arr[1])
		boolValue, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("invalid value of %s=%s, err: %v", k, v, err)
		}
		m[k] = boolValue
	}

	v20, err := util.ClusterVersionAtLeast(client, "v1.20.0")
	if err != nil {
		return err
	}

	if !v20 {
		if _, ok := m[string(EndpointSlice)]; ok {
			m[string(EndpointSlice)] = false
			klog.Error("kubernetes version should greater than v1.20.0 to use EndpointSlice")
		}

		if _, ok := m[string(IPv6DualStack)]; ok {
			m[string(IPv6DualStack)] = false
			klog.Error("kubernetes version should greater than v1.20.0 to use IPv6DualStack")
		}
	}

	return utilfeature.DefaultMutableFeatureGate.SetFromMap(m)
}
