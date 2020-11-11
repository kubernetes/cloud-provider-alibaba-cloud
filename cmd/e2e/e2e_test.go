package e2e

import (
	cloud "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
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
	} else {
		cloud.CloudConfigFile = framework.TestContext.CloudConfig
	}
	if framework.TestContext.LoadBalancerID == "" {
		framework.TestContext.LoadBalancerID = "lb-xxxx"
	}
	if framework.TestContext.MasterZoneID == "" {
		framework.TestContext.MasterZoneID = "x-xxx-x"
	}
	if framework.TestContext.SlaveZoneID == "" {
		framework.TestContext.SlaveZoneID = "x-xxx-x"
	}
	if framework.TestContext.BackendLabel == "" {
		framework.TestContext.BackendLabel = "failure-domain.beta.kubernetes.io/region=cn-beijing,cs-worker=3"
	}
	if framework.TestContext.AclID == "" {
		framework.TestContext.AclID = "acl-xxxxxx"
	}
	if framework.TestContext.VSwitchID == "" {
		framework.TestContext.VSwitchID = "vsw-xxxxxx"
	}
	if framework.TestContext.CertID == "" {
		framework.TestContext.CertID = "xxxxxx_xxxxxx_xxxxxx_xxxxxx"
	}
	if framework.TestContext.PrivateZoneID == "" {
		framework.TestContext.PrivateZoneID = "xxxxx"
	}
	if framework.TestContext.PrivateZoneName == "" {
		framework.TestContext.PrivateZoneName = "xxxxx"
	}
	if framework.TestContext.PrivateZoneRecordName == "" {
		framework.TestContext.PrivateZoneRecordName = "xxxxx"
	}
	if framework.TestContext.PrivateZoneRecordTTL == "" {
		framework.TestContext.PrivateZoneRecordTTL = "0"
	}
	if framework.TestContext.ResourceGroupID == "" {
		framework.TestContext.ResourceGroupID = "rg-xxx"
	}

	if framework.TestContext.TestLabel == "quick" {
		for i := range framework.QuickFrames {
			fram := framework.QuickFrames[i]
			t.Logf("[QuickTest]run action: %d", i)
			err := fram(t)
			if err != nil {
				t.Logf("[QuickTest]action fail: %s", err.Error())
				t.Fail()
			}
		}
	} else {
		for i := range framework.AllFrames {
			fram := framework.AllFrames[i]
			t.Logf("[AllTest]run action: %d", i)
			err := fram(t)
			if err != nil {
				t.Logf("[AllTest]action fail: %s", err.Error())
				t.Fail()
			}
		}
	}

}
