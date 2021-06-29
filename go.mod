module k8s.io/cloud-provider-alibaba-cloud

go 1.16

require (
	github.com/denverdino/aliyungo v0.0.0-20201222091910-a47aa053adf7
	github.com/docker/distribution v2.7.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-cmd/cmd v1.2.0
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	k8s.io/api v0.18.1
	k8s.io/apimachinery v0.18.1
	k8s.io/apiserver v0.0.0
	k8s.io/client-go v0.18.1
	k8s.io/cloud-provider v0.0.0
	k8s.io/component-base v0.18.1
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.0.0
	k8s.io/kubernetes v0.0.0
)

replace (
	k8s.io/api v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/api v0.0.0-20200325144952-9e991415386e
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20200325144952-9e991415386e
	k8s.io/apimachinery v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20200325144952-9e991415386e
	k8s.io/apiserver v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20200325144952-9e991415386e
	k8s.io/cli-runtime v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20200325144952-9e991415386e
	k8s.io/client-go v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/client-go v0.0.0-20200325144952-9e991415386e
	k8s.io/cloud-provider v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20200325144952-9e991415386e
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20200325144952-9e991415386e
	k8s.io/code-generator v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20200325144952-9e991415386e
	k8s.io/component-base v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/component-base v0.0.0-20200325144952-9e991415386e
	k8s.io/cri-api v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20200325144952-9e991415386e
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20200325144952-9e991415386e
	k8s.io/gengo => k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/heapster => k8s.io/heapster v1.2.0-beta.1
	k8s.io/klog => k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20200325144952-9e991415386e
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20200325144952-9e991415386e
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	k8s.io/kube-proxy v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20200325144952-9e991415386e
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20200325144952-9e991415386e
	k8s.io/kubectl v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20200325144952-9e991415386e
	k8s.io/kubelet v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20200325144952-9e991415386e
	k8s.io/kubernetes => k8s.io/kubernetes v0.0.0-20200325144952-9e991415386e
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20200325144952-9e991415386e
	k8s.io/metrics v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/metrics v0.0.0-20200325144952-9e991415386e
	k8s.io/repo-infra => k8s.io/repo-infra v0.0.1-alpha.1
	k8s.io/sample-apiserver v0.0.0 => k8s.io/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20200325144952-9e991415386e
	k8s.io/system-validators => k8s.io/system-validators v1.0.4
	k8s.io/utils => k8s.io/utils v0.0.0-20200117235808-5f6fbceb4c31
	modernc.org/cc => modernc.org/cc v1.0.0
)
