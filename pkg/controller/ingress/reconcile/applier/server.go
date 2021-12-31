package applier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper/k8s"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/go-logr/logr"
)

func NewServerApplier(kubeClient client.Client, albProvider prvd.Provider, serverGroupID string, endpoints []albmodel.BackendItem, trafficPolicy string, logger logr.Logger) *serverApplier {
	return &serverApplier{
		kubeClient:    kubeClient,
		albProvider:   albProvider,
		serverGroupID: serverGroupID,
		endpoints:     endpoints,
		trafficPolicy: trafficPolicy,
		logger:        logger,
	}
}

type serverApplier struct {
	kubeClient    client.Client
	albProvider   prvd.Provider
	serverGroupID string
	endpoints     []albmodel.BackendItem
	trafficPolicy string
	logger        logr.Logger
}

func (s *serverApplier) Apply(ctx context.Context) error {
	traceID := ctx.Value(util.TraceID)

	servers, err := s.albProvider.ListALBServers(ctx, s.serverGroupID)
	if err != nil {
		return err
	}
	s.logger.V(util.SynLogLevel).Info("apply servers",
		"endpoints", s.endpoints,
		"traceID", traceID)
	// todo matched endpoints need check and update weight
	_, unmatchedResEndpoints, unmatchedSDKEndpoints := matchEndpointWithTargets(s.endpoints, servers, s.trafficPolicy)

	if len(unmatchedResEndpoints) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply servers",
			"unmatchedResEndpoints", unmatchedResEndpoints,
			"traceID", traceID)
	}
	if len(unmatchedSDKEndpoints) != 0 {
		s.logger.V(util.SynLogLevel).Info("apply servers",
			"unmatchedSDKEndpoints", unmatchedSDKEndpoints,
			"traceID", traceID)
	}

	if len(unmatchedResEndpoints) == 0 && len(unmatchedSDKEndpoints) == 0 {
		return nil
	}

	// If the number of servers to be added and deleted is less than 40, please call the replacement method, and the others are called separately
	if len(unmatchedResEndpoints) != 0 && len(unmatchedResEndpoints) < util.BatchReplaceServersMaxNum &&
		len(unmatchedSDKEndpoints) != 0 && len(unmatchedSDKEndpoints) < util.BatchReplaceServersMaxNum {
		if err := s.albProvider.ReplaceALBServers(ctx, s.serverGroupID, unmatchedResEndpoints, unmatchedSDKEndpoints); err != nil {
			return err
		}
		// TODO print err
		updateTargetHealthPodCondition(ctx, s.kubeClient, k8s.BuildReadinessGatePodConditionType(), s.endpoints)

		return nil
	}
	if len(unmatchedSDKEndpoints) != 0 {
		if err := s.albProvider.DeregisterALBServers(ctx, s.serverGroupID, unmatchedSDKEndpoints); err != nil {
			return err
		}
	}
	if len(unmatchedResEndpoints) != 0 {
		if err := s.albProvider.RegisterALBServers(ctx, s.serverGroupID, unmatchedResEndpoints); err != nil {
			return err
		}
		updateTargetHealthPodCondition(ctx, s.kubeClient, k8s.BuildReadinessGatePodConditionType(), s.endpoints)
	}

	return nil
}

func (s *serverApplier) PostApply(ctx context.Context) error {
	return nil
}

func matchEndpointWithTargets(endpoints []albmodel.BackendItem, targets []albsdk.BackendServer, trafficPolicy string) ([]endpointAndTargetPair, []albmodel.BackendItem, []albsdk.BackendServer) {
	var matchedEndpointAndTargets []endpointAndTargetPair
	var unmatchedEndpoints []albmodel.BackendItem
	var unmatchedTargets []albsdk.BackendServer

	if len(trafficPolicy) == 0 && len(endpoints) == 0 {
		trafficPolicy = albmodel.ENIBackendType
	}

	endpointsByUID := make(map[string]albmodel.BackendItem, len(endpoints))
	for _, endpoint := range endpoints {
		var endpointUID string
		if isEniTrafficPolicy(trafficPolicy) {
			endpointUID = fmt.Sprintf("%v:%v:%v", endpoint.ServerId, endpoint.ServerIp, endpoint.Port)
		} else {
			endpointUID = fmt.Sprintf("%v:%v", endpoint.ServerId, endpoint.Port)
		}
		klog.Infof("endpointUID: %s", endpointUID)
		endpointsByUID[endpointUID] = endpoint
	}
	targetsByUID := make(map[string]albsdk.BackendServer, len(targets))
	for _, target := range targets {
		var targetUID string
		if isEniTrafficPolicy(trafficPolicy) {
			targetUID = fmt.Sprintf("%v:%v:%v", target.ServerId, target.ServerIp, target.Port)
		} else {
			targetUID = fmt.Sprintf("%v:%v", target.ServerId, target.Port)
		}
		klog.Infof("targetUID: %s", targetUID)
		targetsByUID[targetUID] = target
	}
	endpointUIDs := sets.StringKeySet(endpointsByUID)
	targetUIDs := sets.StringKeySet(targetsByUID)

	for _, uid := range endpointUIDs.Intersection(targetUIDs).List() {
		endpoint := endpointsByUID[uid]
		target := targetsByUID[uid]
		matchedEndpointAndTargets = append(matchedEndpointAndTargets, endpointAndTargetPair{
			endpoint: endpoint,
			target:   target,
		})
	}
	for _, uid := range endpointUIDs.Difference(targetUIDs).List() {
		unmatchedEndpoints = append(unmatchedEndpoints, endpointsByUID[uid])
	}
	for _, uid := range targetUIDs.Difference(endpointUIDs).List() {
		target := targetsByUID[uid]
		if isServerStatusRemoving(target.Status) {
			continue
		}
		unmatchedTargets = append(unmatchedTargets, target)
	}
	return matchedEndpointAndTargets, unmatchedEndpoints, unmatchedTargets
}

func isServerStatusRemoving(status string) bool {
	return strings.EqualFold(status, util.ServerStatusRemoving)
}

type endpointAndTargetPair struct {
	endpoint albmodel.BackendItem
	target   albsdk.BackendServer
}

//TODO remove
func isTrafficPolicyValid(trafficPolicy string) bool {
	if !strings.EqualFold(trafficPolicy, util.TrafficPolicyEni) &&
		!strings.EqualFold(trafficPolicy, util.TrafficPolicyLocal) &&
		!strings.EqualFold(trafficPolicy, util.TrafficPolicyCluster) {
		return false
	}
	return true
}

func isEniTrafficPolicy(trafficPolicy string) bool {
	return strings.EqualFold(trafficPolicy, util.TrafficPolicyEni)
}

// updateTargetHealthPodCondition will updates pod's targetHealth condition for matchedEndpointAndTargets and unmatchedEndpoints.
// returns whether further probe is needed or not
func updateTargetHealthPodCondition(ctx context.Context, kubeClient client.Client, targetHealthCondType v1.PodConditionType,
	endpoints []albmodel.BackendItem) error {
	klog.Infof("start updateTargetHealthPodCondition")
	for _, endpointAndTarget := range endpoints {
		if len(endpointAndTarget.Pod.Spec.ReadinessGates) == 0 {
			continue
		}
		epsKey := types.NamespacedName{
			Namespace: endpointAndTarget.Pod.Namespace,
			Name:      endpointAndTarget.Pod.Name,
		}
		pod := &v1.Pod{}
		if err := kubeClient.Get(ctx, epsKey, pod); err != nil {
			if errors.IsNotFound(err) {
				klog.Errorf("updateTargetHealthPodCondition %v", err.Error())
				continue
			}
			continue
		}
		_, err := updateTargetHealthPodConditionForPod(ctx, kubeClient, pod, targetHealthCondType)
		if err != nil {
			klog.Errorf("updateTargetHealthPodConditionForPod %v", err.Error())
			continue
		}
	}
	return nil
}

// updateTargetHealthPodConditionForPod updates pod's targetHealth condition for a single pod and its matched target.
// returns whether further probe is needed or not.
func updateTargetHealthPodConditionForPod(ctx context.Context, kubeClient client.Client, pod *v1.Pod, targetHealthCondType v1.PodConditionType) (bool, error) {
	if !HasAnyOfReadinessGates(pod, []v1.PodConditionType{targetHealthCondType}) {
		return false, nil
	}

	targetHealthCondStatus := v1.ConditionTrue
	var reason, message string

	existingTargetHealthCond, exists := GetPodCondition(pod, targetHealthCondType)
	// we skip patch pod if it matches current computed status/reason/message.
	if exists &&
		existingTargetHealthCond.Status == targetHealthCondStatus &&
		existingTargetHealthCond.Reason == reason &&
		existingTargetHealthCond.Message == message {
		return true, nil
	}

	newTargetHealthCond := v1.PodCondition{
		Type:    targetHealthCondType,
		Status:  targetHealthCondStatus,
		Reason:  reason,
		Message: message,
	}
	if !exists || existingTargetHealthCond.Status != targetHealthCondStatus {
		newTargetHealthCond.LastTransitionTime = metav1.Now()
	}

	patch, err := buildPodConditionPatch(pod, newTargetHealthCond)
	if err != nil {
		return false, err
	}
	k8sPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			UID:       pod.UID,
		},
	}
	if err := kubeClient.Status().Patch(ctx, k8sPod, patch); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func buildPodConditionPatch(pod *v1.Pod, condition v1.PodCondition) (client.Patch, error) {
	oldData, err := json.Marshal(v1.Pod{
		Status: v1.PodStatus{
			Conditions: nil,
		},
	})
	if err != nil {
		return nil, err
	}
	newData, err := json.Marshal(v1.Pod{
		ObjectMeta: metav1.ObjectMeta{UID: pod.UID}, // only put the uid in the new object to ensure it appears in the patch as a precondition
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{condition},
		},
	})
	if err != nil {
		return nil, err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Pod{})
	if err != nil {
		return nil, err
	}
	return client.RawPatch(types.StrategicMergePatchType, patchBytes), nil
}

// HasAnyOfReadinessGates returns whether podInfo has any of these readinessGates
func HasAnyOfReadinessGates(pod *v1.Pod, conditionTypes []v1.PodConditionType) bool {
	for _, rg := range pod.Spec.ReadinessGates {
		for _, conditionType := range conditionTypes {
			if rg.ConditionType == conditionType {
				return true
			}
		}
	}
	return false
}

// GetPodCondition will get Pod's condition.
func GetPodCondition(pod *v1.Pod, conditionType v1.PodConditionType) (v1.PodCondition, bool) {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == conditionType {
			return cond, true
		}
	}

	return v1.PodCondition{}, false
}
