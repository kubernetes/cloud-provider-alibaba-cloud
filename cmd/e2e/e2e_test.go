package e2e

import (
	"bytes"
	"encoding/json"
	"github.com/golang/glog"
	"k8s.io/cloud-provider-alibaba-cloud/cmd/e2e/framework"
	_ "k8s.io/cloud-provider-alibaba-cloud/cmd/e2e/test"
	"testing"
)

func init() {
	framework.ViperizeFlags()
}

func TestE2E(t *testing.T) {
	if framework.TestContext.KubeConfig == "" {
		t.Logf("empty kubeconfig, assume running in cluster")
	}
	if framework.TestContext.CloudConfig == "" {
		t.Logf("empty cloud-config")
	}
	if framework.TestContext.LoadBalancerID == "" {
		framework.TestContext.LoadBalancerID = "lb-xxxxx"
	}
	if framework.TestContext.MasterZoneID == "" {
		framework.TestContext.MasterZoneID = "cn-beijing-g"
	}
	if framework.TestContext.SlaveZoneID == "" {
		framework.TestContext.SlaveZoneID = "cn-beijing-f"
	}
	if framework.TestContext.BackendLabel == "" {
		framework.TestContext.BackendLabel = "failure-domain.beta.kubernetes.io/region=cn-beijing,cs-worker=3"
	}
	if framework.TestContext.AclID == "" {
		framework.TestContext.AclID = "acl-xxxxxxx"
	}
	if framework.TestContext.VSwitchID == "" {
		framework.TestContext.VSwitchID = "vsw-xxxxx"
	}
	if framework.TestContext.CertID == "" {
		framework.TestContext.CertID = "xxxxxx_xxxxxx_xxxxxx_xxxxxx"
	}
	for i := range framework.Frames {
		fram := framework.Frames[i]
		t.Logf("run action: %d", i)
		err := fram(t)
		if err != nil {
			t.Logf("action fail: %s", err.Error())
			t.Fail()
		}
	}
}

// PrettyJson  pretty json output
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
