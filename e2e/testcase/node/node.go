package node

import (
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/options"
)

func RunNodeControllerTestCases(f *framework.Framework) {
	ginkgo.Describe("node controller", func() {

		ginkgo.Context("reconcile", func() {
			ginkgo.It("node-reconcile", func() {
				err := f.ExpectNodeEqual()
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("node-address-changed", func() {
				oldNode, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())

				newNode := oldNode.DeepCopy()
				for index, value := range newNode.Status.Addresses {
					if value.Type == v1.NodeInternalIP {
						newNode.Status.Addresses[index].Address = "123.123.123.123"
					}
				}
				_, err = f.Client.KubeClient.PatchNodeStatus(oldNode, newNode)
				gomega.Expect(err).To(gomega.BeNil())

				time.Sleep(5 * time.Minute)
				err = f.ExpectNodeEqual()
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("node-label-changed", func() {
				oldNode, err := f.Client.KubeClient.GetLatestNode()
				gomega.Expect(err).To(gomega.BeNil())

				err = f.Client.KubeClient.LabelNode(oldNode.Name, v1.LabelInstanceType, "test-type")
				gomega.Expect(err).To(gomega.BeNil())

				time.Sleep(5 * time.Minute)
				err = f.ExpectNodeEqual()
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("add-node", func() {
			ginkgo.It("add-node", func() {
				if options.TestConfig.ClusterId != "" {
					// created svc
					svc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
					gomega.Expect(err).To(gomega.BeNil())
					defer func() {
						_ = f.Client.KubeClient.DeleteService()
					}()
					// add node
					err = f.Client.ACKClient.ScaleOutCluster(options.TestConfig.ClusterId, 1)
					gomega.Expect(err).To(gomega.BeNil())
					// check node whether equal
					err = f.ExpectNodeEqual()
					gomega.Expect(err).To(gomega.BeNil())
					// check route whether equal
					if options.TestConfig.Network == options.Flannel {
						err = f.ExpectRouteEqual()
						gomega.Expect(err).To(gomega.BeNil())
					}
					// check service whether equal
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})

		ginkgo.Context("remove-node", func() {
			ginkgo.It("remove-node", func() {
				if options.TestConfig.ClusterId != "" {
					// svc created
					svc, err := f.Client.KubeClient.CreateServiceByAnno(nil)
					gomega.Expect(err).To(gomega.BeNil())
					defer func() {
						_ = f.Client.KubeClient.DeleteService()
					}()
					// delete node
					node, err := f.Client.KubeClient.GetLatestNode()
					gomega.Expect(err).To(gomega.BeNil())
					gomega.Expect(node).NotTo(gomega.BeNil())
					err = f.Client.ACKClient.DeleteClusterNodes(options.TestConfig.ClusterId, node.Name)
					gomega.Expect(err).To(gomega.BeNil())
					// check node whether equal
					err = f.ExpectNodeEqual()
					gomega.Expect(err).To(gomega.BeNil())
					// check route whether equal
					if options.TestConfig.Network == options.Flannel {
						err = f.ExpectRouteEqual()
						gomega.Expect(err).To(gomega.BeNil())
					}
					// check service whether equal
					err = f.ExpectLoadBalancerEqual(svc)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})
	})
}
