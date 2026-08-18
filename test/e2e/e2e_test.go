package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/client"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/node"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/route"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/service/clbv1"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/service/nlbv2"
	"k8s.io/klog/v2"
)

func init() {
	options.TestConfig.BindFlags()
}

func TestE2E(t *testing.T) {
	err := options.TestConfig.Validate()
	if err != nil {
		t.Fatalf("test config validate failed: %s", err.Error())
	}
	suiteConfig, _ := ginkgo.GinkgoConfiguration()
	parallelProcess, parallelTotal := suiteConfig.ParallelProcess, suiteConfig.ParallelTotal
	needsCLBResource, needsNLBResource := selectedCloudResourceFamilies(suiteConfig.LabelFilter)
	if err := options.TestConfig.ConfigureSuite(parallelProcess, parallelTotal, needsCLBResource, needsNLBResource); err != nil {
		t.Fatalf("configure parallel test resources: %s", err.Error())
	}
	client.ConfigureTestResources(options.TestConfig.WorkerScope(), options.TestConfig.FixtureReadyTimeout)

	c, err := client.NewClient()
	if err != nil {
		t.Fatalf("create client error: %s", err.Error())
	}
	f := framework.NewFrameWork(c)
	// Testcase registration expects a fully constructed client, but InitOptions
	// can fail on transient cloud API errors. Retry it here and defer reporting
	// the final error to BeforeSuite so the worker still joins the coordinator.
	initErr := initOptionsWithRetry(c)

	ginkgo.BeforeSuite(func() {
		setupErr := initErr
		if setupErr == nil {
			klog.Infof("test config: %s", util.PrettyJson(options.TestConfig))
			setupErr = f.BeforeSuit()
		}
		if setupErr == nil && options.TestConfig.AllowCreateCloudResource {
			setupErr = f.CreateCloudResource()
		}
		if setupErr != nil && f.Client != nil {
			if cleanupErr := f.AfterSuit(); cleanupErr != nil {
				klog.Errorf("cleanup partial setup: %s", cleanupErr)
			}
		}
		gomega.Expect(setupErr).To(gomega.BeNil())
	})

	ginkgo.AfterSuite(func() {
		err = f.AfterSuit()
		gomega.Expect(err).To(gomega.BeNil())
	})

	gomega.RegisterFailHandler(ginkgo.Fail)

	ginkgo.Describe("Run cloud controller manager e2e tests", func() {
		ginkgo.BeforeEach(func() {
			// A failed spec can leave a Service behind even after its scoped
			// cleanup runs. Remove it before the next spec so one cleanup race
			// cannot turn the rest of a worker's results into AlreadyExists noise.
			err := f.AfterEach()
			if err != nil {
				ginkgo.AbortSuite(fmt.Sprintf("cannot clean previous Service before running another case: %v", err))
			}
		})
		ginkgo.AfterEach(func() {
			ginkgo.By("delete service")
			err := f.AfterEach()
			if err != nil {
				// Keep the Service snapshot for one retry in the next BeforeEach
				// (or AfterSuite when this was the last case). This tolerates a
				// transient cloud configuring state without letting the worker
				// run additional cases on top of an unclean fixture.
				klog.Warningf("Service cleanup will be retried before the next case: %v", err)
			}
			gomega.Expect(err).To(gomega.BeNil(), "clean Service after case")
		})
		AddControllerTests(f)
	})

	ginkgo.RunSpecs(t, "run ccm e2e test")
}

func selectedCloudResourceFamilies(labelFilter string) (bool, bool) {
	matches := func(labels ...string) bool {
		return ginkgo.Label(labels...).MatchesLabelFilter(labelFilter)
	}
	needsCLBResource := matches("service", "clb") || matches("service", "clb", "cluster-serial")
	needsNLBResource := matches("service", "nlb") || matches("service", "nlb", "cluster-serial")
	return needsCLBResource, needsNLBResource
}

func initOptionsWithRetry(c *client.E2EClient) error {
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		if err = c.InitOptions(); err == nil {
			return nil
		}
		if attempt < 3 {
			klog.Warningf("init e2e options attempt %d failed, retrying: %s", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return err
}

func AddControllerTests(f *framework.Framework) {
	controllers := strings.Split(options.TestConfig.Controllers, ",")
	if len(controllers) == 0 {
		klog.Info("no controller tests need to run, finished")
		return
	}
	for _, c := range controllers {
		switch c {
		case "service":
			ginkgo.Describe("clb service controller tests", ginkgo.Label("service", "clb"), func() {
				clbv1.RunLoadBalancerTestCases(f)
				clbv1.RunListenerTestCases(f)
				clbv1.RunBackendTestCases(f)
				clbv1.RunGracefulShutdownTestCases(f)
			})

			if options.TestConfig.NLBZoneMaps != "" {
				ginkgo.Describe("nlb service controller tests", ginkgo.Label("service", "nlb"), func() {
					nlbv2.RunLoadBalancerTestCases(f)
					nlbv2.RunListenerTestCases(f)
					nlbv2.RunBackendTestCases(f)
					nlbv2.RunGracefulShutdownTestCases(f)
				})
			} else {
				klog.Warningf("NLBZoneMaps is empty, skip NLB service tests")
			}

		case "node":
			ginkgo.Describe("node controller tests", ginkgo.Label("node"), func() {
				node.RunNodeControllerTestCases(f)
			})
		case "route":
			if options.TestConfig.Network == options.Flannel {
				ginkgo.Describe("route controller tests", ginkgo.Label("route"), func() {
					route.RunRouteControllerTestCases(f)
				})
			}
		default:
			klog.Infof("%s controller is not supported", c)
		}

	}
}
