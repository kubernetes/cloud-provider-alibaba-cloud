package testcase

import (
	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	req "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	"strconv"
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
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Create load balancing of HTTP type finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-004", func() {
		ginkgo.By("Create load balancing of HTTPS type")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ProtocolPort): "https:443",
			req.Annotation(req.CertID):       framework.CONF.YOUR_CERT_ID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(spec),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Specify load balancing specifications finished%v", spec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-006", func() {
		ginkgo.By("Using existing load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.LoadBalancerId): framework.CONF.YOUR_LB_ID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Using existing load balancing finished%v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-007", func() {
		ginkgo.By("Using existing load balancing - overlay listening")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.LoadBalancerId):   framework.CONF.YOUR_LB_ID,
			req.Annotation(req.OverrideListener): "true",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("test Using existing load balancing - overlay listening finished%v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-008", func() {
		ginkgo.By("Temporarily add additional tags with existing load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.LoadBalancerId): framework.CONF.YOUR_LB_ID,
			req.Annotation(req.AdditionalTags): "test tags",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Temporarily add additional tags with existing load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-009", func() {
		ginkgo.By("When creating load balancing, specify the active and standby zones")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.MasterZoneID): framework.CONF.YOUR_MASTER_ZONE_ID,
			req.Annotation(req.SlaveZoneID):  framework.CONF.YOUR_SLAVE_ZONE_ID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Create a load balance for pay by traffic public network type load balancing instance finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0012", func() {
		ginkgo.By("Create a pay per bandwidth load balancing private network type load balancing instance")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AddressType): "intranet",
			req.Annotation(req.ChargeType):  "paybybandwidth",
			req.Annotation(req.Bandwidth):   "2",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}


		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Create a pay per bandwidth load balancing private network type load balancing instance finished %v", dspec)
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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


		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Create a HTTP type health check load balancing finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0015", func() {
		ginkgo.By("Setting scheduling algorithm for load balancing-rr")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.Scheduler): "rr",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Setting scheduling algorithm for load balancing-wrr finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0017", func() {
		ginkgo.By("Setting scheduling algorithm for load balancing-wlc")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.Scheduler): "wlc",
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Setting scheduling algorithm for load balancing-wlc finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0018", func() {
		ginkgo.By("Specifying virtual switches for load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AddressType): "intranet",
			req.Annotation(req.VswitchId):   framework.CONF.YOUR_VSWITCH_ID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(spec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(spec),
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

		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDefaultAction(spec),
			framework.NewDeleteAction(spec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Specify the load balancing name finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-SLB-0025", func() {
		ginkgo.By("Specifies the resource group to which load balancing belongs")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ResourceGroupId): framework.CONF.YOUR_RG_ID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
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
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Configure session retention for load balancing for HTTPS protocols finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-004", func() {
		ginkgo.By("Configure access control policy groups for load balancing - acl-type: white")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AclType):   "white",
			req.Annotation(req.AclStatus): "on",
			req.Annotation(req.AclID):     framework.CONF.YOUR_ALC_ID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Configure access control policy groups for load balancing - acl-type: white finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-005", func() {
		ginkgo.By("Configure access control policy groups for load balancing - acl-type: black")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.AclType):   "black",
			req.Annotation(req.AclStatus): "on",
			req.Annotation(req.AclID):     framework.CONF.YOUR_ALC_ID,
		})

		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Configure access control policy groups for load balancing - acl-type: white finished %v", dspec)
	})
	ginkgo.It("CCM-Service-IT-Annotation-Monitor-006", func() {
		ginkgo.By("Specify forwarding port for load balancing")
		base := framework.NewBaseSVC(map[string]string{
			req.Annotation(req.ProtocolPort): "https:443,http:80",
			req.Annotation(req.CertID):       framework.CONF.YOUR_CERT_ID,
			req.Annotation(req.ForwardPort):  "80:443",
		})
		dspec := &framework.TestUnit{
			Service: base, Description: "Initializes an SVC of the specified type", ExpectOK: ExpectSLBExistAndEqual,
		}
		err := framework.RunActions(
			f,
			framework.NewDefaultAction(dspec),
			framework.NewDeleteAction(dspec),
		)
		framework.ExpectNoError(err)
		framework.Logf("Test Specify forwarding port for load balancing finished %v", dspec)
	})

})

func ExpectSLBExistAndEqual(m *framework.Expectation) (bool, error) {
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(m.Case.Service),
	}

	//0. expect slb must exist


	//err := m.Case.ReqCtx.BuildLoadBalancerAttributeFromCloud(lbMdl)
	err := m.E2E.ModelBuilder.LoadBalancerMgr.Find(m.Case.ReqCtx, lbMdl)


	if err != nil {
		// TODO if err, need retry
		return false, framework.NewErrorRetry(err)
	}
	if lbMdl.LoadBalancerAttribute.LoadBalancerId == "" {
		return false, framework.NewErrorRetry(err)
	}

	//1. spec equal

	if spec := m.Case.ReqCtx.Anno.Get(req.Spec); spec != "" {

		klog.Infof("expect spec type ok: %s", spec)
		if string(lbMdl.LoadBalancerAttribute.LoadBalancerSpec) != spec {
			klog.Info("expected: waiting slb spec change: ", lbMdl.LoadBalancerAttribute.LoadBalancerSpec)
			return false, framework.NewErrorRetry(err)
		}
	}
	//2.network type equal
	if AddressType := m.Case.ReqCtx.Anno.Get(req.AddressType); AddressType != "" {
		klog.Infof("expect AddressType  ok: %s", AddressType)
		if string(lbMdl.LoadBalancerAttribute.AddressType) != AddressType {
			klog.Info("expected: waiting slb AddressType change: ", lbMdl.LoadBalancerAttribute.AddressType)
			return false, framework.NewErrorRetry(err)
		}
	}
	//3.payment type equal
	if paymentType := m.Case.ReqCtx.Anno.Get(req.ChargeType); paymentType != "" {
		klog.Infof("expect payment type ok:", paymentType)
		if string(lbMdl.LoadBalancerAttribute.InternetChargeType) != paymentType {
			klog.Info("expected: waiting slb payment change: ", lbMdl.LoadBalancerAttribute.InternetChargeType)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 4. LoadBalancerName equal
	if LoadBalancerName := m.Case.ReqCtx.Anno.Get(req.LoadBalancerName); LoadBalancerName != "" {
		klog.Infof("expect LoadBalancerName  ok:", LoadBalancerName)
		if string(lbMdl.LoadBalancerAttribute.LoadBalancerName) != LoadBalancerName {
			klog.Info("expected: waiting slb LoadBalancerName change: ", lbMdl.LoadBalancerAttribute.LoadBalancerName)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 5. VSwitchId equal
	if VSwitchId := m.Case.ReqCtx.Anno.Get(req.VswitchId); VSwitchId != "" {
		klog.Infof("expect VSwitchId  ok:", VSwitchId)
		if string(lbMdl.LoadBalancerAttribute.VSwitchId) != VSwitchId {
			klog.Info("expected: waiting slb LoadBalancerName change: ", lbMdl.LoadBalancerAttribute.VSwitchId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 6. MasterZoneId equal
	if MasterZoneId := m.Case.ReqCtx.Anno.Get(req.MasterZoneID); MasterZoneId != "" {
		klog.Infof("expect MasterZoneId  ok:", MasterZoneId)
		if string(lbMdl.LoadBalancerAttribute.MasterZoneId) != MasterZoneId {
			klog.Info("expected: waiting slb MasterZoneId change: ", lbMdl.LoadBalancerAttribute.MasterZoneId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 7. SlaveZoneId equal
	if SlaveZoneId := m.Case.ReqCtx.Anno.Get(req.SlaveZoneID); SlaveZoneId != "" {
		klog.Infof("expect SlaveZoneId  ok:", SlaveZoneId)
		if string(lbMdl.LoadBalancerAttribute.SlaveZoneId) != SlaveZoneId {
			klog.Info("expected: waiting slb SlaveZoneId change: ", lbMdl.LoadBalancerAttribute.SlaveZoneId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 8. DeleteProtection equal
	if DeleteProtection := m.Case.ReqCtx.Anno.Get(req.DeleteProtection); DeleteProtection != "" {
		klog.Infof("expect DeleteProtection  ok:", DeleteProtection)
		if string(lbMdl.LoadBalancerAttribute.DeleteProtection) != DeleteProtection {
			klog.Info("expected: waiting slb DeleteProtectionStatus change: ", lbMdl.LoadBalancerAttribute.SlaveZoneId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 9. ModificationProtectionStatus equal
	if ModificationProtectionStatus := m.Case.ReqCtx.Anno.Get(req.ModificationProtection); ModificationProtectionStatus != "" {
		klog.Infof("expect ModificationProtectionStatus  ok:", ModificationProtectionStatus)
		if string(lbMdl.LoadBalancerAttribute.ModificationProtectionStatus) != ModificationProtectionStatus {
			klog.Info("expected: waiting slb ModificationProtectionStatus change: ", lbMdl.LoadBalancerAttribute.ModificationProtectionStatus)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 10. ResourceGroupId equal
	if ResourceGroupId := m.Case.ReqCtx.Anno.Get(req.ResourceGroupId); ResourceGroupId != "" {
		klog.Infof("expect ResourceGroupId  ok:", ResourceGroupId)
		if string(lbMdl.LoadBalancerAttribute.ResourceGroupId) != ResourceGroupId {
			klog.Info("expected: waiting slb ResourceGroupId change: ", lbMdl.LoadBalancerAttribute.ResourceGroupId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 11. Bandwidth equal
	if Bandwidth := m.Case.ReqCtx.Anno.Get(req.Bandwidth); Bandwidth != "" {
		klog.Infof("expect Bandwidth  ok:%s", Bandwidth)
		if strconv.Itoa(lbMdl.LoadBalancerAttribute.Bandwidth) != Bandwidth {
			klog.Info("expected: waiting slb Bandwidth change: ", lbMdl.LoadBalancerAttribute.Bandwidth)
			return false, framework.NewErrorRetry(err)
		}
	}
	err = m.E2E.ModelBuilder.ListenerMgr.BuildLocalModel(m.Case.ReqCtx,lbMdl)
	if err != nil {
		// TODO if err, need retry
		return false, framework.NewErrorRetry(err)
	}
	//12 .port & proto equal
	if HealthCheckURI := m.Case.ReqCtx.Anno.Get(req.HealthCheckURI); HealthCheckURI != "" {
		klog.Infof("expect HealthCheckURI  ok:%s", HealthCheckURI)
		klog.Infof("expect lbMdl.Listeners  ok:%+v", lbMdl.Listeners)
		klog.Infof("expect Check  ok:%+v", lbMdl.Listeners[0].HealthCheckURI)
		if lbMdl.Listeners[0].HealthCheckURI == ""{
			return false, framework.NewErrorRetry(err)
		}
	}
	return true, nil
}
