package testcase

import (
	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	req "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
)

var _ = ginkgo.Describe("Test CCM Annotation slb function", func() {

	f := framework.NewNamedFrameWork("Start initializing test framework")
	// setup nginx once for all test
	//f.SetUp()
	//defer f.CleanUp()

	// uncomment if u want to run initialize at each iteration
	ginkgo.AfterSuite(f.AfterEach)
	ginkgo.BeforeSuite(f.BeforeEach)

	ginkgo.By("slb test start")

	ginkgo.It("CCM-Service-IT-Annotation-SLB-001", func() {
		ginkgo.By("Create a public network type of load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AddressType): "internet",
		})
		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Create a public network type of load balancing finished%v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-002", func() {
		ginkgo.By("Create a private network type of load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AddressType): "intranet",
		})
		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Create a private network type of load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-003", func() {
		ginkgo.By("Create load balancing of HTTP type")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ProtocolPort): "http:80",
		})
		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Create load balancing of HTTP type finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-004", func() {
		ginkgo.By("Create load balancing of HTTPS type")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ProtocolPort): "https:443",
			req.Annotation(req.CertID):       framework.TestConfig.CertID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Create load balancing of HTTPS type finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-005", func() {
		ginkgo.By("Specify load balancing specifications")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.Spec): "slb.s1.small",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}

		spec := &framework.TestUnit{
			Service: base,
			Mutator: func(svc *v1.Service) error {
				svc.SetAnnotations(map[string]string{
					req.Annotation(req.Spec): "slb.s2.small",
				})
				return nil
			},
			ExpectOK:    ExpectSLBExistAndEqual,
			Description: "change spec type from slb.s1.small to slb.s2.small",
		}
		spec.NewReqContext(f.Cloud)

		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Specify load balancing specifications finished%v", spec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-006", func() {
		ginkgo.By("Using existing load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.LoadBalancerId): framework.TestConfig.LoadBalancerID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Using existing load balancing finished%v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-007", func() {
		ginkgo.By("Using existing load balancing - overlay listening")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.LoadBalancerId):   framework.TestConfig.LoadBalancerID,
			req.Annotation(req.OverrideListener): "true",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Using existing load balancing - overlay listening finished%v", dspec)
	})

	ginkgo.It("CCM-Service-IT-Annotation-SLB-009", func() {
		ginkgo.By("When creating load balancing, specify the active and standby zones")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.MasterZoneID): framework.TestConfig.MasterZoneID,
			req.Annotation(req.SlaveZoneID):  framework.TestConfig.SlaveZoneID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Temporarily add additional tags with existing load balancing finished %v", dspec)
	})

	ginkgo.It("CCM-Service-IT-Annotation-SLB-0010", func() {
		ginkgo.By("Create a pay per bandwidth load balancing public network type load balancing instance")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ChargeType): "paybybandwidth",
			req.Annotation(req.Bandwidth):  "2",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Create a pay per bandwidth load balancing public network type load balancing instance finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0011", func() {
		ginkgo.By("Create a load balance for pay by traffic public network type load balancing instance")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ChargeType): "paybytraffic",
			req.Annotation(req.Bandwidth):  "2",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Create a load balance for pay by traffic public network type load balancing instance finished %v", dspec)
	})

	ginkgo.It("CCM-Service-IT-Annotation-SLB-0013", func() {
		ginkgo.By("Create a TCP type health check load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.HealthCheckType):           "tcp",
			req.Annotation(req.HealthCheckConnectTimeout): "8",
			req.Annotation(req.HealthyThreshold):          "4",
			req.Annotation(req.UnhealthyThreshold):        "4",
			req.Annotation(req.HealthCheckInterval):       "3",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Create a TCP type health check load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0014", func() {
		ginkgo.By("Create a HTTP type health check load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.HealthCheckFlag):     "on",
			req.Annotation(req.HealthCheckType):     "http",
			req.Annotation(req.HealthCheckURI):      "/test/index.html",
			req.Annotation(req.HealthCheckTimeout):  "8",
			req.Annotation(req.HealthyThreshold):    "4",
			req.Annotation(req.UnhealthyThreshold):  "4",
			req.Annotation(req.HealthCheckInterval): "3",
			req.Annotation(req.ProtocolPort):        "http:80",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Create a HTTP type health check load balancing finished %v\n", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0015", func() {
		ginkgo.By("Setting scheduling algorithm for load balancing-rr")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.Scheduler): "rr",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Setting scheduling algorithm for load balancing-rr finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0016", func() {
		ginkgo.By("Setting scheduling algorithm for load balancing-wrr")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.Scheduler): "wrr",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Setting scheduling algorithm for load balancing-wrr finished %v", dspec)
	})
	ginkgo.FIt("CCM-Service-IT-Annotation-SLB-0017", func() {
		ginkgo.By("Setting scheduling algorithm for load balancing-wlc")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.Scheduler): "wlc",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Setting scheduling algorithm for load balancing-wlc finished %v", dspec)
	})
	ginkgo.FIt("CCM-Service-IT-Annotation-SLB-0018", func() {
		ginkgo.By("Specifying virtual switches for load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AddressType): "intranet",
			req.Annotation(req.VswitchId):   framework.TestConfig.VSwitchID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Specifying virtual switches for load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0019", func() {
		ginkgo.By("Add additional tags to load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AdditionalTags): "Key1=Value1,Key2=Value2",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Add additional tags to load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0020", func() {
		ginkgo.By("Creating IPv6 type load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.IPVersion): "ipv6",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Creating IPv6 type load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0021", func() {
		ginkgo.By("Creating IPv4 type load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.IPVersion): "ipv4",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Creating IPv4 type load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0022", func() {
		ginkgo.By("Config delete Protection for load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.DeleteProtection): "on",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		spec := &framework.TestUnit{
			Service: base,
			Mutator: func(svc *v1.Service) error {
				svc.SetAnnotations(map[string]string{
					req.Annotation(req.DeleteProtection): "off",
				})
				return nil
			},
			ExpectOK:    ExpectSLBExistAndEqual,
			Description: "change delete Protection configuration status from on to off ",
		}
		spec.NewReqContext(f.Cloud)

		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Config delete Protection for load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0023", func() {
		ginkgo.By("configuration modification protection for load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ModificationProtection): "on",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		spec := &framework.TestUnit{
			Service: base,
			Mutator: func(svc *v1.Service) error {
				svc.SetAnnotations(map[string]string{
					req.Annotation(req.ModificationProtection): "off",
				})
				return nil
			},
			ExpectOK:    ExpectSLBExistAndEqual,
			Description: "change modification Protection configuration status from on to off ",
		}
		spec.NewReqContext(f.Cloud)

		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test configuration modification protection for load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0024", func() {
		ginkgo.By("Specify the load balancing name")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.LoadBalancerName): "t",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		spec := &framework.TestUnit{
			Service: base,
			Mutator: func(svc *v1.Service) error {
				svc.SetAnnotations(map[string]string{
					req.Annotation(req.LoadBalancerName): "lb-8vbvuibh2uijsebl75njt",
				})
				return nil
			},
			ExpectOK:    ExpectSLBExistAndEqual,
			Description: "change load balancing name ",
		}
		spec.NewReqContext(f.Cloud)
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Specify the load balancing name finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0025", func() {
		ginkgo.By("Specifies the resource group to which load balancing belongs")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ResourceGroupId): framework.TestConfig.ResourceGroupID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Specifies the resource group to which load balancing belongs finished %v", dspec)
	})

	ginkgo.By("monitor test start")
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-001", func() {
		ginkgo.By("Configure session hold time for TCP type load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.PersistenceTimeout): "1800",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Configure session hold time for TCP type load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-002", func() {
		ginkgo.By("Configure session retention for load balancing for HTTP  protocols")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.SessionStick):     "on",
			req.Annotation(req.SessionStickType): "insert",
			req.Annotation(req.CookieTimeout):    "1800",
			req.Annotation(req.ProtocolPort):     "http:80",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Configure session retention for load balancing for HTTP  protocols finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-003", func() {
		ginkgo.By("Configure session retention for load balancing for HTTPS protocols")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.SessionStick):     "on",
			req.Annotation(req.SessionStickType): "insert",
			req.Annotation(req.CookieTimeout):    "1800",
			req.Annotation(req.ProtocolPort):     "https:443",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Configure session retention for load balancing for HTTPS protocols finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-004", func() {
		ginkgo.By("Configure access control policy groups for load balancing - acl-type: white")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AclType):   "white",
			req.Annotation(req.AclStatus): "on",
			req.Annotation(req.AclID):     framework.TestConfig.AclID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Configure access control policy groups for load balancing - acl-type: white finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-005", func() {
		ginkgo.By("Configure access control policy groups for load balancing - acl-type: black")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AclType):   "black",
			req.Annotation(req.AclStatus): "on",
			req.Annotation(req.AclID):     framework.TestConfig.AclID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Configure access control policy groups for load balancing - acl-type: white finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-006", func() {
		ginkgo.By("Specify forwarding port for load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ProtocolPort): "https:443,http:80",
			req.Annotation(req.CertID):       framework.TestConfig.CertID,
			req.Annotation(req.ForwardPort):  "80:443",
		})
		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type",
			ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Specify forwarding port for load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-007", func() {
		ginkgo.By("Specify forwarding port for load balancing")
		base := framework.NewBaseSVC(map[string]string{

		})
		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type",
			Mutator: func(service *v1.Service) error {
				service.SetAnnotations(map[string]string{
					req.Annotation(req.Scheduler): "wrr",
				})
				service.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
				return nil
			},
			ExpectOK: ExpectSLBExistAndEqual,
		}
		del := &framework.TestUnit{
			Service: base, Description: "Delete test svc", ExpectOK: EnsureDeleteSVC,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(del),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Specify forwarding port for load balancing finished %v", dspec)
	})

})



