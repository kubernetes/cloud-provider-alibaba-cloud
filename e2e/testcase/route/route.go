package route

import (
	"context"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/options"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"strings"
)

func RunRouteControllerTestCases(f *framework.Framework) {
	ginkgo.Describe("route controller", func() {

		ginkgo.Context("reconcile", func() {
			ginkgo.It("route-reconcile", func() {
				err := f.ExpectRouteEqual()
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("remove route entry", func() {
			ginkgo.It("remove route entry", func() {
				node, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(node).NotTo(gomega.BeNil())
				err = f.DeleteRouteEntry(node)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectRouteEqual()
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("multi vpc tables", func() {
			ginkgo.It("multi vpc tables", func() {
				// create route table
				resp, err := f.Client.CloudClient.CreateRouteTable(context.TODO(), options.TestConfig.VPCID,
					"usedForCCME2etest")
				gomega.Expect(err).To(gomega.BeNil())
				routeTables, err := f.Client.CloudClient.DescribeRouteTableList(context.TODO(), options.TestConfig.VPCID)
				gomega.Expect(err).To(gomega.BeNil())
				defer func() {
					routes, err := f.Client.CloudClient.ListRoute(context.TODO(), resp.RouteTableId)
					gomega.Expect(err).To(gomega.BeNil())
					for _, t := range routes {
						err = f.Client.CloudClient.DeleteRoute(context.TODO(), resp.RouteTableId,
							t.ProviderId, t.DestinationCIDR)
						gomega.Expect(err).To(gomega.BeNil())
					}
					_, err = f.Client.CloudClient.DeleteRouteTable(context.TODO(), resp.RouteTableId)
					gomega.Expect(err).To(gomega.BeNil())
				}()
				gomega.Expect(len(routeTables) > 1).To(gomega.BeTrue())
				// update ccm config
				if options.TestConfig.ClusterType == "ManagedKubernetes" {
					err = f.Client.ACKClient.ModifyClusterConfiguration(options.TestConfig.ClusterId,
						"cloud-controller-manager",
						map[string]string{"RouteTableIDS": strings.Join(routeTables, ",")},
					)
					gomega.Expect(err).To(gomega.BeNil())
					defer func() {
						err = f.Client.ACKClient.ModifyClusterConfiguration(options.TestConfig.ClusterId,
							"cloud-controller-manager",
							map[string]string{"RouteTableIDS": ""},
						)
						gomega.Expect(err).To(gomega.BeNil())
					}()

					raw := ctrlCfg.CloudCFG.Global.RouteTableIDS
					ctrlCfg.CloudCFG.Global.RouteTableIDS = strings.Join(routeTables, ",")
					defer func() {
						ctrlCfg.CloudCFG.Global.RouteTableIDS = raw
					}()
				}

				err = f.ExpectRouteEqual()
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("sync NetworkAvailable status", func() {
			ginkgo.It("sync NetworkAvailable status", func() {
				node, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(node).NotTo(gomega.BeNil())

				newNode := node.DeepCopy()
				var conditions []v1.NodeCondition
				for _, cond := range node.Status.Conditions {
					if cond.Type == v1.NodeNetworkUnavailable {
						gomega.Expect(cond.Status).To(gomega.Equal(v1.ConditionFalse))
						cond.Status = v1.ConditionTrue
						conditions = append(conditions, cond)
						continue
					}
					conditions = append(conditions, cond)
				}
				newNode.Status.Conditions = conditions
				_, err = f.Client.KubeClient.PatchNodeStatus(node, newNode)
				gomega.Expect(err).To(gomega.BeNil())

				err = f.ExpectNodeEqual()
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

	})
}
