package framework

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/client"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/options"
	"k8s.io/klog/v2"
)

const (
	SLBResource = "SLB"
	NLBResource = "NLB"
	ACLResource = "ACL"
	EIPResource = "EIP"
)

const (
	TestAnnotationIgnoreBackends = "test-ignore-backends"
)

func scopedResourceName(role string) string {
	return fmt.Sprintf("%s-%s-%s", options.TestConfig.ClusterId, options.TestConfig.WorkerScope(), role)
}

type Framework struct {
	Client                          *client.E2EClient
	CreatedResource                 map[string]string
	secondaryIntranetLoadBalancerID string
	pendingServiceCleanup           *v1.Service
	namespaceCreated                bool
}

type cloudResource struct {
	id, resourceType string
}

func NewFrameWork(c *client.E2EClient) *Framework {
	return &Framework{
		Client:          c,
		CreatedResource: make(map[string]string),
	}
}

func (f *Framework) BeforeSuit() error {
	err := f.Client.KubeClient.CreateNamespace()
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("test namespace %s already exists; another run may be active or manual cleanup is required", client.Namespace)
		} else {
			return err
		}
	}
	f.namespaceCreated = true
	if err := f.Client.KubeClient.CreateDeployment(); err != nil {
		return err
	}

	if options.TestConfig.EnableVK {
		if err := f.Client.KubeClient.CreateVKDeployment(); err != nil {
			return err
		}
	}

	return nil
}

func (f *Framework) AfterSuit() error {
	var errs []error
	serviceCleaned := true
	if f.namespaceCreated {
		// Stop a failed spec's Service from referencing or recreating a fixture.
		// Keep the namespace itself until cloud cleanup finishes: it is the lock
		// that prevents a new run from reusing the same fixed worker names.
		if err := f.AfterEach(); err != nil {
			errs = append(errs, fmt.Errorf("clean residual service: %w", err))
			serviceCleaned = false
		}
	}
	if f.namespaceCreated && options.TestConfig.AllowCreateCloudResource {
		// Lazy fixtures are created from specs. If the cloud accepted a create
		// request but its response was lost, the ID was never recorded locally.
		// Fixed worker-scoped names let teardown rediscover and remove it.
		if err := f.FindWorkerScopedCloudResources(); err != nil {
			errs = append(errs, fmt.Errorf("find worker-scoped cloud resources: %w", err))
		}
	}
	if serviceCleaned {
		if err := f.CleanCloudResources(); err != nil {
			errs = append(errs, fmt.Errorf("clean cloud resources: %w", err))
		}
	}
	if f.namespaceCreated && len(errs) == 0 {
		if err := f.Client.KubeClient.DeleteNamespace(); err != nil {
			errs = append(errs, fmt.Errorf("delete test namespace: %w", err))
		} else {
			f.namespaceCreated = false
		}
	}
	return utilerrors.NewAggregate(errs)
}

func (f *Framework) AfterEach() error {
	svc, err := f.Client.KubeClient.GetService()
	if err != nil {
		if apierrors.IsNotFound(err) {
			if f.pendingServiceCleanup == nil {
				return nil
			}
			if err := f.deleteManualEndpoints(f.pendingServiceCleanup); err != nil {
				return err
			}
			return f.waitForServiceCleanup(f.pendingServiceCleanup)
		}
		return err
	}
	// Record the Service before sending DELETE. The API server may accept the
	// request even if the client loses the response; keep enough state to verify
	// the corresponding cloud cleanup on the next attempt.
	f.pendingServiceCleanup = svc.DeepCopy()
	if err = f.Client.KubeClient.DeleteService(); err != nil {
		return err
	}
	if err = f.deleteManualEndpoints(svc); err != nil {
		return err
	}
	return f.waitForServiceCleanup(f.pendingServiceCleanup)
}

func (f *Framework) deleteManualEndpoints(svc *v1.Service) error {
	if svc == nil || len(svc.Spec.Selector) != 0 {
		return nil
	}
	return f.Client.KubeClient.DeleteEndpoints()
}

func (f *Framework) waitForServiceCleanup(svc *v1.Service) error {
	var cleanupErr error
	pollErr := wait.PollImmediate(2*time.Second, 2*time.Minute, func() (bool, error) {
		_, cleanupErr = f.Client.KubeClient.GetService()
		if cleanupErr == nil {
			cleanupErr = fmt.Errorf("service %s/%s still exists", svc.Namespace, svc.Name)
			return false, nil
		}
		if !apierrors.IsNotFound(cleanupErr) {
			return false, nil
		}

		// The Service finalizer is the authoritative signal that CCM finished
		// processing deletion. Only inspect the cloud state after the Service is
		// gone; polling the full NLB model while the finalizer is still present
		// creates avoidable API pressure across parallel workers.
		cleanupErr = f.expectServiceCleaned(svc)
		if cleanupErr != nil {
			return false, nil
		}
		return true, nil
	})
	if pollErr == nil {
		f.pendingServiceCleanup = nil
		return nil
	}
	if pollErr != nil && cleanupErr != nil {
		return cleanupErr
	}
	return pollErr
}

func (f *Framework) expectServiceCleaned(svc *v1.Service) error {
	if isNLBService(svc) {
		return f.expectNLBServiceCleaned(svc)
	}

	remote, err := buildRemoteModel(f, svc)
	if err != nil {
		if strings.Contains(err.Error(), "The specified LoadBalancerId does not exist") {
			//  reuse SLB failed, ignoring cleanup
			return nil
		}
		return err
	}
	if svc.Annotations[annotation.Annotation(annotation.LoadBalancerId)] == "" {
		if remote.LoadBalancerAttribute.LoadBalancerId == "" {
			return nil
		} else {
			return fmt.Errorf("slb %s is not deleted", remote.LoadBalancerAttribute.LoadBalancerId)
		}
	} else {
		return f.ExpectLoadBalancerClean(svc, remote)
	}
}

func isNLBService(svc *v1.Service) bool {
	return svc != nil && svc.Spec.LoadBalancerClass != nil && *svc.Spec.LoadBalancerClass == helper.NLBClass
}

func (f *Framework) expectNLBServiceCleaned(svc *v1.Service) error {
	remote, err := buildNLBRemoteModel(f, svc)
	if err != nil {
		if isNetworkLoadBalancerNotFound(err) {
			return nil
		}
		return err
	}
	if svc.Annotations[annotation.Annotation(annotation.LoadBalancerId)] == "" {
		if remote.LoadBalancerAttribute == nil || remote.LoadBalancerAttribute.LoadBalancerId == "" {
			return nil
		}
		return fmt.Errorf("nlb %s is not deleted", remote.LoadBalancerAttribute.LoadBalancerId)
	}
	return f.ExpectNetworkLoadBalancerClean(svc, remote)
}

func (f *Framework) CreateCloudResource() error {
	f.CreatedResource = make(map[string]string)
	var region string
	if options.TestConfig.NeedsCloudResource("clb") {
		var err error
		region, err = f.Client.CloudClient.Region()
		if err != nil {
			return err
		}
	}

	var zoneMappings []nlbmodel.ZoneMapping
	if options.TestConfig.NeedsCloudResource("nlb") {
		var err error
		zoneMappings, err = ParseNLBZoneMappings(options.TestConfig.NLBZoneMaps)
		if err != nil {
			return err
		}
	}

	manageIntranetCLBFixture := options.TestConfig.NeedsCloudResource("clb") &&
		options.TestConfig.IntranetLoadBalancerID == ""
	if manageIntranetCLBFixture {
		vswId, err := f.Client.CloudClient.VswitchID()
		if err != nil {
			return fmt.Errorf("get vsw id error: %s", err.Error())
		}
		slbM := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				AddressType:      model.IntranetAddressType,
				LoadBalancerSpec: model.S1Small,
				RegionId:         region,
				VSwitchId:        vswId,
				LoadBalancerName: scopedResourceName("private-clb"),
			},
		}
		if err := f.Client.CloudClient.FindLoadBalancerByName(slbM); err != nil {
			return err
		}
		if slbM.LoadBalancerAttribute.LoadBalancerId == "" {
			if err := f.Client.CloudClient.CreateLoadBalancer(context.TODO(), slbM, ""); err != nil {
				return fmt.Errorf("create intranet slb error: %s", err.Error())
			}
		}
		options.TestConfig.IntranetLoadBalancerID = slbM.LoadBalancerAttribute.LoadBalancerId
		f.CreatedResource[options.TestConfig.IntranetLoadBalancerID] = SLBResource
	}

	if manageIntranetCLBFixture && options.TestConfig.VServerGroupID == "" {
		if err := f.ensureCLBVServerGroups(options.TestConfig.IntranetLoadBalancerID); err != nil {
			return err
		}
	}

	if options.TestConfig.NeedsCloudResource("nlb") && options.TestConfig.IntranetNetworkLoadBalancerID == "" {
		slbM := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				AddressType:  nlbmodel.IntranetAddressType,
				ZoneMappings: zoneMappings,
				VpcId:        options.TestConfig.VPCID,
				Name:         scopedResourceName("private-nlb"),
			},
		}

		if err := f.Client.CloudClient.FindNLBByName(context.TODO(), slbM); err != nil {
			return err
		}
		if slbM.LoadBalancerAttribute.LoadBalancerId == "" {
			if err := f.Client.CloudClient.CreateNLB(context.TODO(), slbM, ""); err != nil {
				return fmt.Errorf("create intranet nlb error: %s", err.Error())
			}
		}
		options.TestConfig.IntranetNetworkLoadBalancerID = slbM.LoadBalancerAttribute.LoadBalancerId
		f.CreatedResource[options.TestConfig.IntranetNetworkLoadBalancerID] = NLBResource
	}

	if options.TestConfig.NeedsCloudResource("clb") && options.TestConfig.AclID == "" {
		aclName := scopedResourceName("acl-a")
		aclId, err := f.Client.CloudClient.DescribeAccessControlList(context.TODO(), aclName)
		if err != nil {
			return fmt.Errorf("DescribeAccessControlList error: %s", err.Error())
		}
		if aclId == "" {
			aclId, err = f.Client.CloudClient.CreateAccessControlList(context.TODO(), aclName)
			if err != nil {
				return fmt.Errorf("CreateAccessControlList error: %s", err.Error())
			}
		}
		options.TestConfig.AclID = aclId
		f.CreatedResource[aclId] = ACLResource
	}

	if options.TestConfig.NeedsCloudResource("clb") && options.TestConfig.AclID2 == "" {
		aclName := scopedResourceName("acl-b")
		aclId, err := f.Client.CloudClient.DescribeAccessControlList(context.TODO(), aclName)
		if err != nil {
			return fmt.Errorf("DescribeAccessControlList error: %s", err.Error())
		}
		if aclId == "" {
			aclId, err = f.Client.CloudClient.CreateAccessControlList(context.TODO(), aclName)
			if err != nil {
				return fmt.Errorf("CreateAccessControlList error: %s", err.Error())
			}
		}
		options.TestConfig.AclID2 = aclId
		f.CreatedResource[aclId] = ACLResource
	}

	klog.Infof("created resource: %s", util.PrettyJson(f.CreatedResource))
	return nil
}

// EnsureInternetLoadBalancer lazily creates the public CLB fixture. Most E2E
// specs use the intranet fixture; only specs that explicitly verify public-LB
// behavior should call this method.
func (f *Framework) EnsureInternetLoadBalancer() error {
	if options.TestConfig.InternetLoadBalancerID != "" {
		return nil
	}
	region, err := f.Client.CloudClient.Region()
	if err != nil {
		return err
	}
	lb := &model.LoadBalancer{LoadBalancerAttribute: model.LoadBalancerAttribute{
		AddressType:      model.InternetAddressType,
		LoadBalancerSpec: model.S1Small,
		RegionId:         region,
		LoadBalancerName: scopedResourceName("public-clb"),
	}}
	if err := f.Client.CloudClient.FindLoadBalancerByName(lb); err != nil {
		return err
	}
	if lb.LoadBalancerAttribute.LoadBalancerId == "" {
		if err := f.Client.CloudClient.CreateLoadBalancer(context.TODO(), lb, ""); err != nil {
			return fmt.Errorf("create internet slb: %w", err)
		}
	}
	options.TestConfig.InternetLoadBalancerID = lb.LoadBalancerAttribute.LoadBalancerId
	f.CreatedResource[options.TestConfig.InternetLoadBalancerID] = SLBResource
	return nil
}

// EnsureSecondaryIntranetLoadBalancer lazily creates a second private CLB for
// cases that must verify a resource belongs to a different load balancer.
func (f *Framework) EnsureSecondaryIntranetLoadBalancer() (string, error) {
	if f.secondaryIntranetLoadBalancerID != "" {
		return f.secondaryIntranetLoadBalancerID, nil
	}
	region, err := f.Client.CloudClient.Region()
	if err != nil {
		return "", err
	}
	vswitchID, err := f.Client.CloudClient.VswitchID()
	if err != nil {
		return "", err
	}
	lb := &model.LoadBalancer{LoadBalancerAttribute: model.LoadBalancerAttribute{
		AddressType:      model.IntranetAddressType,
		LoadBalancerSpec: model.S1Small,
		RegionId:         region,
		VSwitchId:        vswitchID,
		LoadBalancerName: scopedResourceName("secondary-private-clb"),
	}}
	if err := f.Client.CloudClient.FindLoadBalancerByName(lb); err != nil {
		return "", err
	}
	if lb.LoadBalancerAttribute.LoadBalancerId == "" {
		if err := f.Client.CloudClient.CreateLoadBalancer(context.TODO(), lb, ""); err != nil {
			return "", fmt.Errorf("create secondary intranet slb: %w", err)
		}
	}
	f.secondaryIntranetLoadBalancerID = lb.LoadBalancerAttribute.LoadBalancerId
	f.CreatedResource[f.secondaryIntranetLoadBalancerID] = SLBResource
	return f.secondaryIntranetLoadBalancerID, nil
}

// EnsureInternetNetworkLoadBalancer lazily creates the public NLB fixture for
// the small set of specs that explicitly exercise public-NLB behavior.
func (f *Framework) EnsureInternetNetworkLoadBalancer() error {
	if options.TestConfig.InternetNetworkLoadBalancerID != "" {
		return nil
	}
	zoneMappings, err := ParseNLBZoneMappings(options.TestConfig.NLBZoneMaps)
	if err != nil {
		return err
	}
	lb := &nlbmodel.NetworkLoadBalancer{LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
		AddressType:  nlbmodel.InternetAddressType,
		ZoneMappings: zoneMappings,
		VpcId:        options.TestConfig.VPCID,
		Name:         scopedResourceName("public-nlb"),
	}}
	if err := f.Client.CloudClient.FindNLBByName(context.TODO(), lb); err != nil {
		return err
	}
	if lb.LoadBalancerAttribute.LoadBalancerId == "" {
		if err := f.Client.CloudClient.CreateNLB(context.TODO(), lb, ""); err != nil {
			return fmt.Errorf("create internet nlb: %w", err)
		}
	}
	options.TestConfig.InternetNetworkLoadBalancerID = lb.LoadBalancerAttribute.LoadBalancerId
	f.CreatedResource[options.TestConfig.InternetNetworkLoadBalancerID] = NLBResource
	return nil
}

// EnsureEIPAddress lazily allocates the EIP fixture for the zone-mapping specs
// that explicitly exercise public addresses.
func (f *Framework) EnsureEIPAddress() error {
	if options.TestConfig.EIPID != "" {
		return nil
	}
	name := scopedResourceName("eip")
	id, err := f.Client.CloudClient.DescribeEipIdByName(context.TODO(), name)
	if err != nil {
		return fmt.Errorf("describe eip by name: %w", err)
	}
	if id == "" {
		id, err = f.Client.CloudClient.AllocateEipAddressWithChargeType(context.TODO(), name, "PayByTraffic")
		if err != nil {
			return fmt.Errorf("allocate eip address: %w", err)
		}
	}
	options.TestConfig.EIPID = id
	f.CreatedResource[id] = EIPResource
	return nil
}

func (f *Framework) ensureCLBVServerGroups(loadBalancerID string) error {
	vsg := []model.VServerGroup{{VGroupName: scopedResourceName("vsg-a")}, {VGroupName: scopedResourceName("vsg-b")}}
	remote, err := f.Client.CloudClient.DescribeVServerGroups(context.TODO(), loadBalancerID)
	if err != nil {
		return err
	}
	for i := range vsg {
		for _, existing := range remote {
			if vsg[i].VGroupName == existing.VGroupName {
				vsg[i].VGroupId = existing.VGroupId
				break
			}
		}
		if vsg[i].VGroupId == "" {
			if err := f.Client.CloudClient.CreateVServerGroup(context.TODO(), &vsg[i], loadBalancerID); err != nil {
				return fmt.Errorf("create vserver group: %w", err)
			}
		}
	}
	options.TestConfig.VServerGroupID = vsg[0].VGroupId
	options.TestConfig.VServerGroupID2 = vsg[1].VGroupId
	return nil
}

// FindWorkerScopedCloudResources discovers fixtures owned by the configured
// suite worker without creating anything. It recovers safely after interruption.
func (f *Framework) FindWorkerScopedCloudResources() error {
	if f.CreatedResource == nil {
		f.CreatedResource = make(map[string]string)
	}

	if options.TestConfig.NeedsCloudResource("clb") {
		region, err := f.Client.CloudClient.Region()
		if err != nil {
			return err
		}
		for _, role := range []string{"public-clb", "private-clb"} {
			lb := &model.LoadBalancer{LoadBalancerAttribute: model.LoadBalancerAttribute{
				RegionId:         region,
				LoadBalancerName: scopedResourceName(role),
			}}
			if err := f.Client.CloudClient.FindLoadBalancerByName(lb); err != nil {
				return err
			}
			if lb.LoadBalancerAttribute.LoadBalancerId != "" {
				f.CreatedResource[lb.LoadBalancerAttribute.LoadBalancerId] = SLBResource
			}
		}
		secondary := &model.LoadBalancer{LoadBalancerAttribute: model.LoadBalancerAttribute{
			RegionId:         region,
			LoadBalancerName: scopedResourceName("secondary-private-clb"),
		}}
		if err := f.Client.CloudClient.FindLoadBalancerByName(secondary); err != nil {
			return err
		}
		if secondary.LoadBalancerAttribute.LoadBalancerId != "" {
			f.CreatedResource[secondary.LoadBalancerAttribute.LoadBalancerId] = SLBResource
		}
		for _, suffix := range []string{"a", "b"} {
			name := scopedResourceName("acl-" + suffix)
			id, err := f.Client.CloudClient.DescribeAccessControlList(context.TODO(), name)
			if err != nil {
				return err
			}
			if id != "" {
				f.CreatedResource[id] = ACLResource
			}
		}
	}

	if options.TestConfig.NeedsCloudResource("nlb") {
		for _, role := range []string{"public-nlb", "private-nlb"} {
			lb := &nlbmodel.NetworkLoadBalancer{LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				Name: scopedResourceName(role),
			}}
			if err := f.Client.CloudClient.FindNLBByName(context.TODO(), lb); err != nil {
				return err
			}
			if lb.LoadBalancerAttribute.LoadBalancerId != "" {
				f.CreatedResource[lb.LoadBalancerAttribute.LoadBalancerId] = NLBResource
			}
		}
		name := scopedResourceName("eip")
		id, err := f.Client.CloudClient.DescribeEipIdByName(context.TODO(), name)
		if err != nil {
			return err
		}
		if id != "" {
			f.CreatedResource[id] = EIPResource
		}
	}

	klog.Infof("found worker-scoped resources: %s", util.PrettyJson(f.CreatedResource))
	return nil
}

func (f *Framework) DeleteLoadBalancer(lbid string) error {
	region, err := f.Client.CloudClient.Region()
	if err != nil {
		return err
	}
	slbM := &model.LoadBalancer{
		LoadBalancerAttribute: model.LoadBalancerAttribute{
			LoadBalancerId: lbid,
			RegionId:       region,
		},
	}
	err = f.Client.CloudClient.SetLoadBalancerDeleteProtection(context.TODO(), lbid, string(model.OffFlag))
	if err != nil {
		return err
	}

	err = f.Client.CloudClient.DeleteLoadBalancer(context.TODO(), slbM)
	if err != nil {
		return err
	}
	return nil
}

func (f *Framework) DeleteLoadBalancerAndWait(lbid string) error {
	return f.deleteCloudResourceAndWait(lbid, SLBResource)
}

func (f *Framework) DeleteNetworkLoadBalancerAndWait(lbid string) error {
	return f.deleteCloudResourceAndWait(lbid, NLBResource)
}

func (f *Framework) DeleteSecurityGroupAndWait(securityGroupID string) error {
	if err := retryCloudCleanup(func() error {
		return f.Client.CloudClient.DeleteSecurityGroup(context.TODO(), securityGroupID)
	}); err != nil && !isCloudResourceGone(err) {
		return err
	}

	var observeErr error
	pollErr := wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
		groups, err := f.Client.CloudClient.DescribeSecurityGroups(context.TODO(), nil)
		observeErr = err
		if observeErr != nil {
			if isRetryableCloudCleanupError(observeErr) {
				return false, nil
			}
			return false, observeErr
		}
		for _, group := range groups {
			if group.ID == securityGroupID {
				return false, nil
			}
		}
		return true, nil
	})
	if pollErr != nil && observeErr != nil {
		return observeErr
	}
	if pollErr != nil {
		return fmt.Errorf("timed out waiting for security group %s to disappear: %w", securityGroupID, pollErr)
	}
	return nil
}

func (f *Framework) CleanCloudResources() error {
	klog.Infof("try to clean cloud resources: %+v", f.CreatedResource)
	var errs []error
	for _, resource := range orderedCloudResources(f.CreatedResource) {
		if err := f.deleteCloudResourceAndWait(resource.id, resource.resourceType); err != nil {
			errs = append(errs, fmt.Errorf("delete %s %s: %w", resource.resourceType, resource.id, err))
			continue
		}
		delete(f.CreatedResource, resource.id)
	}
	return utilerrors.NewAggregate(errs)
}

func (f *Framework) deleteCloudResourceAndWait(id, resourceType string) error {
	if err := retryCloudCleanup(func() error {
		return f.deleteCloudResource(id, resourceType)
	}); err != nil {
		return err
	}

	var observeErr error
	pollErr := wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
		var gone bool
		gone, observeErr = f.cloudResourceGone(id, resourceType)
		if observeErr != nil {
			if isCloudResourceGone(observeErr) {
				return true, nil
			}
			if isRetryableCloudCleanupError(observeErr) {
				klog.Warningf("retry cloud cleanup verification for %s %s: %s", resourceType, id, observeErr)
				return false, nil
			}
			return false, observeErr
		}
		return gone, nil
	})
	if pollErr != nil && observeErr != nil {
		return observeErr
	}
	if pollErr != nil {
		return fmt.Errorf("timed out waiting for %s %s to disappear: %w", resourceType, id, pollErr)
	}
	return nil
}

func orderedCloudResources(created map[string]string) []cloudResource {
	resources := make([]cloudResource, 0, len(created))
	for id, resourceType := range created {
		resources = append(resources, cloudResource{id: id, resourceType: resourceType})
	}
	// Delete load balancers before resources they may still reference.
	priority := func(resourceType string) int {
		switch resourceType {
		case SLBResource, NLBResource:
			return 0
		case ACLResource:
			return 1
		case EIPResource:
			return 2
		default:
			return 3
		}
	}
	sort.Slice(resources, func(i, j int) bool {
		left, right := priority(resources[i].resourceType), priority(resources[j].resourceType)
		if left != right {
			return left < right
		}
		if resources[i].resourceType != resources[j].resourceType {
			return resources[i].resourceType < resources[j].resourceType
		}
		return resources[i].id < resources[j].id
	})
	return resources
}

func (f *Framework) deleteCloudResource(id, resourceType string) error {
	switch resourceType {
	case SLBResource:
		return f.DeleteLoadBalancer(id)
	case NLBResource:
		return f.DeleteNetworkLoadBalancer(id)
	case ACLResource:
		return f.Client.CloudClient.DeleteAccessControlList(context.TODO(), id)
	case EIPResource:
		return f.Client.CloudClient.ReleaseEipAddress(context.TODO(), id)
	default:
		return fmt.Errorf("unknown cloud resource type %q", resourceType)
	}
}

func (f *Framework) cloudResourceGone(id, resourceType string) (bool, error) {
	switch resourceType {
	case SLBResource:
		lb := &model.LoadBalancer{LoadBalancerAttribute: model.LoadBalancerAttribute{LoadBalancerId: id}}
		err := f.Client.CloudClient.DescribeLoadBalancer(context.TODO(), lb)
		return isCloudResourceGone(err), ignoreGoneError(err)
	case NLBResource:
		lb := &nlbmodel.NetworkLoadBalancer{LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{LoadBalancerId: id}}
		err := f.Client.CloudClient.DescribeNLB(context.TODO(), lb)
		return isCloudResourceGone(err), ignoreGoneError(err)
	case ACLResource:
		for _, role := range []string{"acl-a", "acl-b"} {
			foundID, err := f.Client.CloudClient.DescribeAccessControlList(context.TODO(), scopedResourceName(role))
			if err != nil {
				return false, err
			}
			if foundID == id {
				return false, nil
			}
		}
		return true, nil
	case EIPResource:
		foundID, err := f.Client.CloudClient.DescribeEipIdByName(context.TODO(), scopedResourceName("eip"))
		return foundID != id, err
	default:
		return false, fmt.Errorf("unknown cloud resource type %q", resourceType)
	}
}

func ignoreGoneError(err error) error {
	if isCloudResourceGone(err) {
		return nil
	}
	return err
}

func retryCloudCleanup(fn func() error) error {
	var lastErr error
	pollErr := wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		lastErr = fn()
		if lastErr == nil || isCloudResourceGone(lastErr) {
			return true, nil
		}
		if !isRetryableCloudCleanupError(lastErr) {
			return false, lastErr
		}
		klog.Warningf("retry transient cloud cleanup error: %s", lastErr.Error())
		return false, nil
	})
	if pollErr != nil && lastErr != nil {
		return lastErr
	}
	return pollErr
}

func isRetryableCloudCleanupError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	for _, fragment := range []string{
		"ResourceInConfiguring.loadbalancer",
		"connection reset by peer",
		"i/o timeout",
		"no such host",
		"Client.Timeout",
		"context deadline exceeded",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

func isCloudResourceGone(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	for _, fragment := range []string{
		"ResourceNotFound.loadBalancer",
		"InvalidLoadBalancerId.NotFound",
		"The specified LoadBalancerId does not exist",
		"InvalidAllocationId.NotFound",
		"InvalidAcl.NotFound",
		"InvalidSecurityGroup.NotFound",
		"InvalidSecurityGroupId.NotFound",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

func (f *Framework) DeleteNetworkLoadBalancer(lbid string) error {
	slbM := &nlbmodel.NetworkLoadBalancer{
		LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
			LoadBalancerId: lbid,
		},
	}

	delCfg := &nlbmodel.DeletionProtectionConfig{
		Enabled: false,
	}

	err := f.Client.CloudClient.UpdateLoadBalancerProtection(context.TODO(), lbid, delCfg, nil)
	if err != nil {
		return err
	}

	err = f.Client.CloudClient.DeleteNLB(context.TODO(), slbM)
	if err != nil {
		return err
	}
	return nil
}
