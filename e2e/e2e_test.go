package e2e

import (
	"github.com/onsi/ginkgo"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	_ "k8s.io/cloud-provider-alibaba-cloud/e2e/testcase"
	"testing"
)

func init() {
	framework.ViperizeFlags()
}

func TestE2E(t *testing.T) {
	ginkgo.RunSpecs(t, "run ccm e2e test")
}
