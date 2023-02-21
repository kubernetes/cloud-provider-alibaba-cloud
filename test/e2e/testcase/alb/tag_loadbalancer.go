package alb

import (
	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/testcase/alb/common"
)

func transSDKTagListToMap(tagList []albsdk.Tag) map[string]string {
	tagMap := make(map[string]string)
	for _, tag := range tagList {
		tagMap[tag.Key] = tag.Value
	}
	return tagMap
}

func RunTagLoadBalancerTestCases(f *framework.Framework) {
	albConfig := common.ALB{}
	ingressClass := common.IngressClass{}
	ginkgo.Describe("alb-ingress-controller testcase: albconfig", func() {
		ingressClass.CreateIngressClass(f)
		ginkgo.It("[alb][p0] test tag common albconfig", func() {
			var err error
			err = albConfig.CleanAlbConfigByName(f, "default")
			gomega.Expect(err).To(gomega.BeNil())
			albConfigInfo := albConfig.GetDefaultAlbConfig()
			albConfigInfo.Spec.LoadBalancer.Tags = []v1.Tag{
				{
					Key:   "tagKey1",
					Value: "tagValue1",
				},
				{
					Key:   "tagKey2",
					Value: "tagValue2",
				},
			}
			albConfig.WaitCreateALBConfig(f, albConfigInfo, true)
			sdkALb, err := albConfig.GetCloudALBByName("default", f)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(len(sdkALb)).To(gomega.Equal(1))
			sdkMaps := transSDKTagListToMap(sdkALb[0].Tags)
			gomega.Expect(sdkMaps["tagKey1"]).To(gomega.Equal("tagValue1"))
			gomega.Expect(sdkMaps["tagKey2"]).To(gomega.Equal("tagValue2"))

		})

		ginkgo.It("[alb][p0] test tag special albconfig", func() {
			var err error
			err = albConfig.CleanAlbConfigByName(f, "default")
			gomega.Expect(err).To(gomega.BeNil())
			albConfigInfo := albConfig.GetDefaultAlbConfig()
			albConfigInfo.Spec.LoadBalancer.Tags = []v1.Tag{
				{
					Key:   "tag Key1",
					Value: " tagValue1 ",
				},
				{
					Key:   "tag-Key2",
					Value: "tag/Value2",
				},
			}
			albConfig.WaitCreateALBConfig(f, albConfigInfo, true)

			sdkALb, err := albConfig.GetCloudALBByName("default", f)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(len(sdkALb)).To(gomega.Equal(1))
			sdkMaps := transSDKTagListToMap(sdkALb[0].Tags)
			gomega.Expect(sdkMaps["tag Key1"]).To(gomega.Equal(" tagValue1 "))
			gomega.Expect(sdkMaps["tag-Key2"]).To(gomega.Equal("tag/Value2"))
		})
	})
}
