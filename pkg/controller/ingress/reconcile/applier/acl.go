package applier

import (
	"context"
	"fmt"
	"sync"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
)

func NewAclApplier(albProvider prvd.Provider, trackingProvider tracking.TrackingProvider, stack core.Manager, logger logr.Logger) *aclApplier {
	return &aclApplier{
		albProvider:      albProvider,
		trackingProvider: trackingProvider,
		stack:            stack,
		logger:           logger,
	}
}

type aclApplier struct {
	albProvider      prvd.Provider
	trackingProvider tracking.TrackingProvider
	stack            core.Manager
	logger           logr.Logger
}

func (s *aclApplier) Apply(ctx context.Context) error {
	var resAcls []*alb.Acl
	s.stack.ListResources(&resAcls)

	resAclsByLsID, err := mapResAclByListenerID(ctx, resAcls)
	if err != nil {
		return err
	}

	var resLSs []*alb.Listener
	s.stack.ListResources(&resLSs)

	resLSsByLsID, err := mapResListenerByListenerID(ctx, resLSs)
	if err != nil {
		return err
	}

	var (
		errSynthesize error
		wgSynthesize  sync.WaitGroup
		chSynthesize  = make(chan struct{}, util.ListenerConcurrentNum)
	)

	for lsID := range resLSsByLsID {
		chSynthesize <- struct{}{}
		wgSynthesize.Add(1)

		go func(listenerID string) {
			util.RandomSleepFunc(util.ConcurrentMaxSleepMillisecondTime)

			defer func() {
				wgSynthesize.Done()
				<-chSynthesize
			}()

			acls := resAclsByLsID[listenerID]
			if errOnce := s.synthesizeAclsOnListener(ctx, resLSsByLsID[listenerID], acls); errSynthesize == nil && errOnce != nil {
				s.logger.Error(errOnce, "synthesize acl failed", "listener", listenerID)
				errSynthesize = errOnce
				return
			}
		}(lsID)
	}
	wgSynthesize.Wait()
	if errSynthesize != nil {
		return errSynthesize
	}

	return nil
}
func (s *aclApplier) PostApply(ctx context.Context) error {
	return nil
}

func (s *aclApplier) synthesizeAclsOnListener(ctx context.Context, listener *alb.Listener, resAcls []*alb.Acl) error {
	if listener == nil {
		return fmt.Errorf("empty listenerwhen synthesize acls error")
	}
	traceID := ctx.Value(util.TraceID)
	lsId, err := listener.ListenerID().Resolve(ctx)
	if err != nil {
		return err
	}
	var oldAclType string
	if len(resAcls) > 0 {
		oldAclType = resAcls[0].Spec.AclType
	}

	aclId, aclType, err := s.findListenerAclConfig(ctx, lsId)
	if err != nil {
		return err
	}
	sdkAcls := make([]albsdk.Acl, 0)
	// get acl list on listener
	if aclId != "" {
		sdkAcls, err = s.findSDKAclsOnLS(ctx, listener, aclId)
		if err != nil {
			return err
		}
	}

	matchedResAndSDKAcls, unmatchedResAcls, unmatchedSDKAcls := matchResAndSDKAcls(resAcls, sdkAcls)
	// acl type change need re-related operation
	if aclType != oldAclType {
		unmatchedSDKAcls = sdkAcls
		unmatchedResAcls = resAcls
		matchedResAndSDKAcls = make([]alb.ResAndSDKAclPair, 0)
	}

	if len(matchedResAndSDKAcls) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize acls",
			"matchedResAndSDKAcls", matchedResAndSDKAcls,
			"traceID", traceID)
	}
	if len(unmatchedResAcls) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize acls",
			"unmatchedResAcls", unmatchedResAcls,
			"traceID", traceID)
	}
	if len(unmatchedSDKAcls) != 0 {
		s.logger.V(util.SynLogLevel).Info("synthesize acls",
			"unmatchedSDKAcls", unmatchedSDKAcls,
			"traceID", traceID)
	}
	for _, sdkAcl := range unmatchedSDKAcls {
		// if err := s.albProvider.DeleteAcl(ctx, lsId, sdkAcl.AclId); err != nil {
		// 	return err
		// }
		// TODO: cannt remove acl, sync alb ACL list to aclconfig, wirte to stack
		sdkAclEntries, err := s.albProvider.ListAclEntriesByID(traceID, sdkAcl.AclId)
		if err != nil {
			return err
		}
		entries := make([]alb.AclEntry, 0)
		for _, entry := range sdkAclEntries {
			entries = append(entries, alb.AclEntry{Entry: entry.Entry})
		}
		aclSpec := &alb.AclSpec{
			ListenerID: listener.ListenerID(),
			AclName:    sdkAcl.AclName,
			AclType:    aclType,
			AclEntries: entries,
		}
		aclResID := fmt.Sprintf("%v", listener.Spec.ListenerPort)
		alb.NewAcl(s.stack, aclResID, *aclSpec)

	}
	for _, resAcl := range unmatchedResAcls {
		aclStatus, err := s.albProvider.CreateAcl(ctx, resAcl)
		if err != nil {
			return err
		}
		resAcl.SetStatus(aclStatus)
	}
	for _, resAndSDKAcl := range matchedResAndSDKAcls {
		aclStatus, err := s.albProvider.UpdateAcl(ctx, lsId, resAndSDKAcl)
		if err != nil {
			return err
		}
		resAndSDKAcl.ResAcl.SetStatus(aclStatus)
	}
	return nil
}

func (s *aclApplier) findListenerAclConfig(ctx context.Context, lsId string) (string, string, error) {
	lsAttrResponse, err := s.albProvider.GetALBListenerAttribute(ctx, lsId)
	if err != nil {
		return "", "", err
	}
	if len(lsAttrResponse.AclConfig.AclRelations) == 0 {
		return "", "", nil
	}
	aclConfig := lsAttrResponse.AclConfig
	return aclConfig.AclRelations[0].AclId, aclConfig.AclType, nil
}

func (s *aclApplier) findSDKAclsOnLS(ctx context.Context, listener *alb.Listener, aclId string) ([]albsdk.Acl, error) {
	acls, err := s.albProvider.ListAcl(ctx, listener, aclId)
	if err != nil {
		return nil, err
	}
	return acls, nil
}

func matchResAndSDKAcls(resAcls []*alb.Acl, sdkAcls []albsdk.Acl) ([]alb.ResAndSDKAclPair, []*alb.Acl, []albsdk.Acl) {
	var matchedResAndSDKAcls []alb.ResAndSDKAclPair
	var unmatchedResAcls []*alb.Acl
	var unmatchedSDKAcls []albsdk.Acl

	resAclsMap := mapResAclByName(resAcls)
	sdkAclsMap := mapSdkAclByName(sdkAcls)

	resAclsSet := sets.StringKeySet(resAclsMap)
	sdkAclsSet := sets.StringKeySet(sdkAclsMap)

	for _, aclName := range resAclsSet.Intersection(sdkAclsSet).List() {
		resAcl := resAclsMap[aclName]
		sdkAcl := sdkAclsMap[aclName]
		matchedResAndSDKAcls = append(matchedResAndSDKAcls, alb.ResAndSDKAclPair{
			ResAcl: resAcl,
			SdkAcl: &sdkAcl,
		})
	}
	for _, aclName := range resAclsSet.Difference(sdkAclsSet).List() {
		unmatchedResAcls = append(unmatchedResAcls, resAclsMap[aclName])
	}
	for _, aclName := range sdkAclsSet.Difference(resAclsSet).List() {
		unmatchedSDKAcls = append(unmatchedSDKAcls, sdkAclsMap[aclName])
	}
	return matchedResAndSDKAcls, unmatchedResAcls, unmatchedSDKAcls
}

func mapResAclByName(resAcls []*alb.Acl) map[string]*alb.Acl {
	resAclByName := make(map[string]*alb.Acl, 0)
	for _, resAcl := range resAcls {
		resAclByName[resAcl.Spec.AclName] = resAcl
	}
	return resAclByName
}
func mapSdkAclByName(sdkAcls []albsdk.Acl) map[string]albsdk.Acl {
	sdkAclByName := make(map[string]albsdk.Acl, 0)
	for _, sdkAcl := range sdkAcls {
		sdkAclByName[sdkAcl.AclName] = sdkAcl
	}
	return sdkAclByName
}

func mapResAclByListenerID(ctx context.Context, resAcls []*alb.Acl) (map[string][]*alb.Acl, error) {
	resAclByLsID := make(map[string][]*alb.Acl, 0)
	for _, acl := range resAcls {
		lsID, err := acl.Spec.ListenerID.Resolve(ctx)
		if err != nil {
			return nil, err
		}
		resAclByLsID[lsID] = append(resAclByLsID[lsID], acl)
	}
	return resAclByLsID, nil
}
