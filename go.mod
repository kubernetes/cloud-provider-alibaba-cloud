module k8s.io/cloud-provider-alibaba-cloud

go 1.14

require (
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1024
	github.com/ghodss/yaml v1.0.0
	github.com/go-cmd/cmd v1.2.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cloud-provider v0.20.5
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.13.0
	sigs.k8s.io/controller-runtime v0.8.3
)

//github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
replace k8s.io/client-go => k8s.io/client-go v0.20.4 // Required by prometheus-operator
