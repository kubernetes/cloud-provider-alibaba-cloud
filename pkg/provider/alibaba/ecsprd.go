package alibaba

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	log "github.com/sirupsen/logrus"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

/*
	Provider needs permission
	"alibaba:DeleteInstance", "alibaba:RunCommand",
*/
func NewEcsProvider(
	auth *metadata.ClientAuth,
) *EcsProvider {
	return &EcsProvider{auth: auth}
}

var _ prvd.Instance = &EcsProvider{}

type EcsProvider struct {
	auth *metadata.ClientAuth
}

func (e *EcsProvider) ListInstances(ctx *node.NodeContext, ids []string) (map[string]*prvd.NodeAttribute, error) {
	nodeRegionMap := make(map[string][]string)
	for _, id := range ids {
		regionID, nodeID, err := nodeFromProviderID(id)
		if err != nil {
			return nil, err
		}
		nodeRegionMap[regionID] = append(nodeRegionMap[regionID], nodeID)
	}

	var insList []ecs.Instance
	for region, nodes := range nodeRegionMap {
		ins, err := e.getInstances(nodes, region)
		if err != nil {
			return nil, err
		}
		insList = append(insList, ins...)
	}
	mins := make(map[string]*prvd.NodeAttribute)
	for _, id := range ids {
		mins[id] = nil
		for _, n := range insList {
			if strings.Contains(id, n.InstanceId) {
				mins[id] = &prvd.NodeAttribute{
					InstanceID:   n.InstanceId,
					InstanceType: n.InstanceType,
					Addresses:    findAddress(&n),
					Zone:         n.ZoneId,
					Region:       n.RegionId,
				}
				break
			}
		}
	}
	return mins, nil
}

func (e *EcsProvider) getInstances(ids []string, region string) ([]ecs.Instance, error) {
	bids, err := json.Marshal(ids)
	if err != nil {
		return nil, fmt.Errorf("get instances error: %s", err.Error())
	}
	req := ecs.CreateDescribeInstancesRequest()
	req.InstanceIds = string(bids)
	req.PageSize = requests.NewInteger(50)

	// TODO: pagination
	instances, err := e.auth.ECS.DescribeInstances(req)
	if err != nil {
		log.Errorf("calling DescribeInstances: region=%s, "+
			"instancename=%s, message=[%s].", req.RegionId, req.InstanceName, err.Error())
		return nil, err
	}
	return instances.Instances.Instance, nil
}

func (e *EcsProvider) SetInstanceTags(
	ctx *node.NodeContext, id string, tags map[string]string,
) error {
	var mtag []ecs.AddTagsTag
	for k, v := range tags {
		mtag = append(mtag, ecs.AddTagsTag{Key: k, Value: v})
	}
	req := ecs.CreateAddTagsRequest()
	req.ResourceId = id
	req.Tag = &mtag
	req.ResourceType = "instance"

	_, err := e.auth.ECS.AddTags(req)
	return err
}

func (e *EcsProvider) DetailECS(ctx *node.NodeContext) (*prvd.DetailECS, error) {
	return nil, nil
}

func (e *EcsProvider) ReplaceSystemDisk(ctx *node.NodeContext) error {
	return nil
}

func (e *EcsProvider) DestroyECS(ctx *node.NodeContext) error {
	return nil
}
func (e *EcsProvider) RestartECS(ctx *node.NodeContext) error { return nil }
func (e *EcsProvider) RunCommand(
	ctx *node.NodeContext, command string,
) (*ecs.Invocation, error) {

	return nil, nil
}

const (
	// 240s
	StopECSTimeout  = 240
	StartECSTimeout = 300

	// Default timeout value for WaitForInstance method
	InstanceDefaultTimeout = 120
	DefaultWaitForInterval = 5
)

func Wait4Instance(
	client *ecs.Client,
	id, status string,
	timeout int,
) error {
	if timeout <= 0 {
		timeout = InstanceDefaultTimeout
	}
	req := ecs.CreateDescribeInstanceAttributeRequest()
	req.InstanceId = id
	for {
		instance, err := client.DescribeInstanceAttribute(req)
		if err != nil {
			return err
		}
		if instance.Status == status {
			//TODO
			//Sleep one more time for timing issues
			time.Sleep(DefaultWaitForInterval * time.Second)
			break
		}
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return fmt.Errorf("timeout waiting %s %s", id, status)
		}
		time.Sleep(DefaultWaitForInterval * time.Second)

	}
	return nil
}

// providerID
// 1) the id of the instance in the alicloud API. Use '.' to separate providerID which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7'. The format of "REGION.NODEID"
// 2) the id for an instance in the kubernetes API, which has 'alicloud://' prefix. e.g. alicloud://cn-hangzhou.i-v98dklsmnxkkgiiil7
func nodeFromProviderID(providerID string) (string, string, error) {
	if strings.HasPrefix(providerID, "alicloud://") {
		k8sName := strings.Split(providerID, "://")
		if len(k8sName) < 2 {
			return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
		} else {
			providerID = k8sName[1]
		}
	}

	name := strings.Split(providerID, ".")
	if len(name) < 2 {
		return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
	}
	return name[0], name[1], nil
}

func findAddress(instance *ecs.Instance) []v1.NodeAddress {
	var addrs []v1.NodeAddress

	if len(instance.PublicIpAddress.IpAddress) > 0 {
		for _, ipaddr := range instance.PublicIpAddress.IpAddress {
			addrs = append(addrs, v1.NodeAddress{Type: v1.NodeExternalIP, Address: ipaddr})
		}
	}

	if instance.EipAddress.IpAddress != "" {
		addrs = append(addrs, v1.NodeAddress{Type: v1.NodeExternalIP, Address: instance.EipAddress.IpAddress})
	}

	if len(instance.InnerIpAddress.IpAddress) > 0 {
		for _, ipaddr := range instance.InnerIpAddress.IpAddress {
			addrs = append(addrs, v1.NodeAddress{Type: v1.NodeInternalIP, Address: ipaddr})
		}
	}

	if len(instance.VpcAttributes.PrivateIpAddress.IpAddress) > 0 {
		for _, ipaddr := range instance.VpcAttributes.PrivateIpAddress.IpAddress {
			addrs = append(addrs, v1.NodeAddress{Type: v1.NodeInternalIP, Address: ipaddr})
		}
	}

	return addrs
}
