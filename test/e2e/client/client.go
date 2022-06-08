package client

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/ecs"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/pvtz"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/vpc"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
	"k8s.io/klog/v2"
	"os"
	runtime "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
)

type E2EClient struct {
	CloudClient   *alibaba.AlibabaCloud
	KubeClient    *KubeClient
	RuntimeClient runtime.Client
	ACKClient     *ACKClient
}

func NewClient() (*E2EClient, error) {
	ctrlCfg.ControllerCFG.CloudConfigPath = options.TestConfig.CloudConfig

	ackClient, err := NewACKClient()
	if err != nil {
		panic(fmt.Sprintf("initialize alibaba client: %s", err.Error()))
	}

	if err := InitCloudConfig(ackClient); err != nil {
		panic(fmt.Sprintf("init cloud config error: %s", err.Error()))
	}
	mgr, err := base.NewClientMgr()
	if err != nil || mgr == nil {
		return nil, fmt.Errorf("initialize alibaba cloud client auth error: %v", err)
	}
	err = mgr.Start(base.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %s", err.Error())
	}
	cc := &alibaba.AlibabaCloud{
		IMetaData:    mgr.Meta,
		ECSProvider:  ecs.NewECSProvider(mgr),
		SLBProvider:  slb.NewLBProvider(mgr),
		PVTZProvider: pvtz.NewPVTZProvider(mgr),
		VPCProvider:  vpc.NewVPCProvider(mgr),
	}

	cfg := config.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("new client : %s", err.Error()))
	}

	runtimeClient, err := runtime.New(cfg, runtime.Options{})
	if err != nil {
		panic(fmt.Sprintf("new runtime client error: %s", err.Error()))
	}

	return &E2EClient{
		CloudClient:   cc,
		KubeClient:    NewKubeClient(kubeClient),
		RuntimeClient: runtimeClient,
		ACKClient:     ackClient,
	}, nil
}

func InitCloudConfig(client *ACKClient) error {
	ack, err := client.DescribeClusterDetail(options.TestConfig.ClusterId)
	if err != nil {
		return err
	}
	if ctrlCfg.CloudCFG.Global.Region == "" {
		ctrlCfg.CloudCFG.Global.Region = *ack.RegionId
	}
	if ctrlCfg.CloudCFG.Global.ClusterID == "" {
		ctrlCfg.CloudCFG.Global.ClusterID = *ack.ClusterId
	}
	if ctrlCfg.CloudCFG.Global.VswitchID == "" {
		vswitchIds := strings.Split(*ack.VswitchId, ",")
		if len(vswitchIds) > 1 {
			ctrlCfg.CloudCFG.Global.VswitchID = vswitchIds[0]
		} else {
			ctrlCfg.CloudCFG.Global.VswitchID = *ack.VswitchId
		}
	}
	if ctrlCfg.CloudCFG.Global.VpcID == "" {
		ctrlCfg.CloudCFG.Global.VpcID = *ack.VpcId
	}

	return nil
}

func (client *E2EClient) InitOptions() error {
	ack, err := client.ACKClient.DescribeClusterDetail(options.TestConfig.ClusterId)
	if err != nil {
		return err
	}
	options.TestConfig.ClusterType = *ack.ClusterType
	if ack.SubnetCidr == nil || *ack.SubnetCidr == "" {
		options.TestConfig.Network = options.Terway
		if err := os.Setenv("SERVICE_FORCE_BACKEND_ENI", "true"); err != nil {
			return err
		}
	} else {
		options.TestConfig.Network = options.Flannel
		if err := os.Setenv("SERVICE_FORCE_BACKEND_ENI", "false"); err != nil {
			return err
		}
	}

	if options.TestConfig.VSwitchID == "" {
		vswId, err := client.CloudClient.VswitchID()
		if err != nil {
			return err
		}
		if vswId != "" {
			options.TestConfig.VSwitchID = vswId
		} else {
			vswitchIds := strings.Split(*ack.VswitchId, ",")
			if len(vswitchIds) > 1 {
				options.TestConfig.VSwitchID = vswitchIds[0]
			} else {
				options.TestConfig.VSwitchID = *ack.VswitchId
			}
		}
	}

	if options.TestConfig.VPCID == "" {
		vpcId, err := client.CloudClient.VpcID()
		if err != nil {
			return err
		}
		if vpcId != "" {
			options.TestConfig.VPCID = vpcId
		} else {
			options.TestConfig.VPCID = *ack.VpcId
		}
	}

	if options.TestConfig.VSwitchID2 == "" {
		vsws, err := client.CloudClient.DescribeVSwitches(context.TODO(), *ack.VpcId)
		if err != nil {
			return err
		}
		found := false
		for _, v := range vsws {
			if v.VSwitchId != options.TestConfig.VSwitchID {
				options.TestConfig.VSwitchID2 = v.VSwitchId
				found = true
				break
			}
		}
		if !found {
			klog.Warningf("vpc %s has no available vsws, VSwitchID2 is nil", *ack.VpcId)
		}
	}

	if options.TestConfig.MasterZoneID == "" || options.TestConfig.SlaveZoneID == "" {
		resources, err := client.CloudClient.DescribeAvailableResource(context.TODO(), "classic_internet", "ipv4")
		if err != nil {
			return fmt.Errorf("describe available slb resources error: %s", err.Error())
		}
		if len(resources) < 2 {
			return fmt.Errorf("no available slb resource, skip create internet slb")
		}
		options.TestConfig.MasterZoneID = resources[0].MasterZoneId
		options.TestConfig.SlaveZoneID = resources[0].SlaveZoneId
	}

	addon, err := client.ACKClient.DescribeClusterAddonsUpgradeStatus(options.TestConfig.ClusterId, "virtual-kubelet")
	if err != nil {
		return fmt.Errorf("DescribeClusterAddonsUpgradeStatus error: %s", err.Error())
	}
	if addon.Version == "" {
		options.TestConfig.EnableVK = false
	} else {
		options.TestConfig.EnableVK = true
	}

	if options.TestConfig.CertID == "" || options.TestConfig.CertID2 == "" {
		certs, err := client.CloudClient.DescribeServerCertificates(context.TODO())
		if err != nil {
			return fmt.Errorf("DescribeServerCertificates error: %s", err.Error())
		}
		if len(certs) != 0 {
			for _, cert := range certs {
				if options.TestConfig.CertID == "" {
					options.TestConfig.CertID = cert
					continue
				}
				if options.TestConfig.CertID2 == "" && cert != options.TestConfig.CertID {
					options.TestConfig.CertID2 = cert
				}
			}
		}
	}

	if options.TestConfig.CACertID == "" {
		cacerts, err := client.CloudClient.DescribeCACertificates(context.TODO())
		if err != nil {
			return fmt.Errorf("DescribeCACertificates error: %s", err.Error())
		}
		if len(cacerts) > 0 {
			options.TestConfig.CACertID = cacerts[0]
		}
	}
	return nil
}
