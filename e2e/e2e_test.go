package e2e

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	_ "k8s.io/cloud-provider-alibaba-cloud/e2e/testcase"
	"testing"
)

func init() {
	framework.ViperizeFlags()
}

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(framework.Fail)
	ginkgo.RunSpecs(t, "run ccm e2e test")
}
