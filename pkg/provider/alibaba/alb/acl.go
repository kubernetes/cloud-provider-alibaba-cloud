package alb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/pkg/errors"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
)

func (m *ALBProvider) CreateAcl(ctx context.Context, resAcl *alb.Acl) (alb.AclStatus, error) {
	traceID := ctx.Value(util.TraceID)

	// 创建Acl实例
	aclResp, err := m.createAcl(traceID, resAcl)
	if err != nil {
		return alb.AclStatus{}, err
	}
	// 新增acl状态等待，需要等到状态转为Available
	err = m.waitAclStatus(traceID, aclResp.AclId)
	if err != nil {
		return alb.AclStatus{}, err
	}

	// 添加cidr实体
	err = m.addEntriesToAcl(traceID, resAcl.Spec.AclEntries, resAcl, aclResp.AclId)
	if err != nil {
		m.deleteAcl(traceID, aclResp.AclId)
		return alb.AclStatus{}, err
	}
	err = m.waitAclStatus(traceID, aclResp.AclId)
	if err != nil {
		m.deleteAcl(traceID, aclResp.AclId)
		return alb.AclStatus{}, err
	}
	// 关联Acl和Listener
	err = m.associateAclWithListener(ctx, traceID, resAcl, aclResp.AclId)
	if err != nil {
		m.deleteAcl(traceID, aclResp.AclId)
		return alb.AclStatus{}, err
	}

	return buildResAclStatus(aclResp.AclId), nil
}
func (m *ALBProvider) UpdateAcl(ctx context.Context, listenerID string, resAndSdkAclPair alb.ResAndSDKAclPair) (alb.AclStatus, error) {
	traceID := ctx.Value(util.TraceID)

	resAcl := resAndSdkAclPair.ResAcl
	sdkAcl := resAndSdkAclPair.SdkAcl
	m.logger.V(util.MgrLogLevel).Info("update acls",
		"resAcl", resAcl,
		"sdkAcl", sdkAcl,
		"traceID", traceID,
		util.Action, "UpdateAcl")
	// 获取sdkEntries列表
	sdkAclEntries, err := m.listAclEntries(traceID, sdkAcl.AclId)
	if err != nil {
		return alb.AclStatus{}, err
	}
	m.logger.V(util.MgrLogLevel).Info("update acls",
		"sdkAclEntries", sdkAclEntries,
		"traceID", traceID,
		util.Action, "UpdateAcl")

	// 对比resEntries和sdkEntries比较差异
	unmatchResAclEntries, unmatchSDKAclEntries := m.matchResAndSDKAclEntries(resAcl.Spec.AclEntries, sdkAclEntries)

	if len(unmatchResAclEntries) != 0 {
		m.logger.V(util.SynLogLevel).Info("update acls",
			"unmatchedResAcls", unmatchResAclEntries,
			"traceID", traceID)
	}
	if len(unmatchSDKAclEntries) != 0 {
		m.logger.V(util.SynLogLevel).Info("update acls",
			"unmatchedSDKAcls", unmatchSDKAclEntries,
			"traceID", traceID)
	}

	// 如果有差异进行更新，sdk多了的删除，少了的增加
	if len(unmatchResAclEntries) > 0 {
		// 添加Entries
		if err := m.addEntriesToAcl(traceID, unmatchResAclEntries, resAcl, sdkAcl.AclId); err != nil {
			return alb.AclStatus{}, err
		}
	}
	if len(unmatchSDKAclEntries) > 0 {
		// 删除Entries
		if err := m.removeEntriesFromAcl(traceID, unmatchSDKAclEntries, resAcl, sdkAcl.AclId); err != nil {
			return alb.AclStatus{}, err
		}
	}
	isAssociated, err := m.isAssociateAclWithListener(ctx, traceID, resAcl, listenerID, sdkAcl.AclId)
	if err != nil {
		return alb.AclStatus{}, err
	}
	if !isAssociated {
		err = m.associateAclWithListener(ctx, traceID, resAcl, sdkAcl.AclId)
		if err != nil {
			m.deleteAcl(traceID, sdkAcl.AclId)
			return alb.AclStatus{}, err
		}
	}

	return buildResAclStatus(sdkAcl.AclId), nil
}

func (m *ALBProvider) DeleteAcl(ctx context.Context, listenerID, sdkAclID string) error {
	traceID := ctx.Value(util.TraceID)

	// 解除关联listener
	if err := m.disassociateAclWithListener(traceID, listenerID, sdkAclID); err != nil {
		return err
	}

	// 删除Acl实例
	if err := m.deleteAcl(traceID, sdkAclID); err != nil {
		return err
	}

	return nil
}
func (m *ALBProvider) ListAcl(ctx context.Context, listener *alb.Listener, aclId string) ([]albsdk.Acl, error) {
	traceID := ctx.Value(util.TraceID)

	if listener == nil {
		return nil, fmt.Errorf("invalid listener for listing acls")
	}

	var (
		nextToken string
		acls      []albsdk.Acl
	)

	// TODO 根据listener查询Acl列表
	lsID := listener.ListenerID()
	listAclsReq := albsdk.CreateListAclsRequest()
	listAclsReq.AclIds = &[]string{aclId}

	for {
		listAclsReq.NextToken = nextToken

		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("listing acls",
			"listenerID", lsID,
			"aclId", aclId,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.ListAcl)
		listAclResp, err := m.auth.ALB.ListAcls(listAclsReq)
		if err != nil {
			return nil, err
		}
		m.logger.V(util.MgrLogLevel).Info("listed acls",
			"listenerID", lsID,
			"aclId", aclId,
			"traceID", traceID,
			"requestID", listAclResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListAcl)

		acls = append(acls, listAclResp.Acls...)

		if listAclResp.NextToken == "" {
			break
		} else {
			nextToken = listAclResp.NextToken
		}
	}
	return acls, nil
}
func (m *ALBProvider) ListAclEntriesByID(traceID interface{}, sdkAclID string) ([]albsdk.AclEntry, error) {
	return m.listAclEntries(traceID, sdkAclID)
}

func (m *ALBProvider) deleteAcl(traceID interface{}, aclID string) error {
	deleteAclReq := albsdk.CreateDeleteAclRequest()
	deleteAclReq.AclId = aclID
	if err := util.RetryImmediateOnError(m.waitAclExistencePollInterval, m.waitAclExistenceTimeout, func(err error) bool { return true }, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("deleting acl",
			"aclID", aclID,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.DeleteAcl)
		deleteAclResp, err := m.auth.ALB.DeleteAcl(deleteAclReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("deleting acl",
				"aclID", aclID,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.DeleteAcl)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("deleted acl",
			"aclID", aclID,
			"traceID", traceID,
			"requestID", deleteAclResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.DeleteAcl)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to delete acl")
	}
	return nil
}

func (m *ALBProvider) disassociateAclWithListener(traceID interface{}, listenerID, aclID string) error {
	disassociateAclWithListenerReq := albsdk.CreateDissociateAclsFromListenerRequest()
	disassociateAclWithListenerReq.ListenerId = listenerID
	disassociateAclWithListenerReq.AclIds = &[]string{aclID}
	if err := util.RetryImmediateOnError(m.waitAclExistencePollInterval, m.waitAclExistenceTimeout, isQuotaExceededError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("disassociate acl with listener",
			"traceID", traceID,
			"listenerID", listenerID,
			"aclID", aclID,
			"startTime", startTime,
			util.Action, util.DissociateAclsFromListener)
		disassociateAclWithListenerResp, err := m.auth.ALB.DissociateAclsFromListener(disassociateAclWithListenerReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("disassociate acl with listener",
				"traceID", traceID,
				"listenerID", listenerID,
				"aclID", aclID,
				"requestID", disassociateAclWithListenerResp.RequestId,
				"elapsedTime", time.Since(startTime).Milliseconds(),
				"error", err.Error(),
				util.Action, util.DissociateAclsFromListener)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("disassociate acl with listener",
			"traceID", traceID,
			"listenerID", listenerID,
			"aclID", aclID,
			"requestID", disassociateAclWithListenerResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.DissociateAclsFromListener)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to disassociate acl with listener")
	}
	return nil
}

func (m *ALBProvider) isAssociateAclWithListener(ctx context.Context, traceID interface{}, resAcl *alb.Acl, listenerID, aclID string) (bool, error) {
	listAclRelationsReq := albsdk.CreateListAclRelationsRequest()
	listAclRelationsReq.AclIds = &[]string{aclID}
	var listAclRelationsResp *albsdk.ListAclRelationsResponse
	if err := util.RetryImmediateOnError(m.waitAclExistencePollInterval, m.waitAclExistenceTimeout, isQuotaExceededError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("list acl associates",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"listenerID", listenerID,
			"aclID", aclID,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.ListAclRelations)
		var err error
		listAclRelationsResp, err = m.auth.ALB.ListAclRelations(listAclRelationsReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("list acl associates",
				"stackID", resAcl.Stack().StackID(),
				"resourceID", resAcl.ID(),
				"listenerID", listenerID,
				"aclID", aclID,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.ListAclRelations)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("list acl associates",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"traceID", traceID,
			"listenerID", listenerID,
			"aclID", aclID,
			"listenerIDS", listAclRelationsResp.AclRelations,
			"requestID", listAclRelationsResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListAclRelations)
		return nil
	}); err != nil {
		return false, errors.Wrap(err, "failed to list acl associates")
	}
	res := false
	for _, listener := range listAclRelationsResp.AclRelations[0].RelatedListeners {
		if listener.ListenerId == listenerID {
			res = true
			break
		}
	}
	return res, nil
}

func (m *ALBProvider) removeEntriesFromAcl(traceID interface{}, entries []albsdk.AclEntry, resAcl *alb.Acl, aclID string) error {
	cnt := util.BatchRemoveEntriesToAclMaxNum
	total := len(entries)

	for total > cnt {
		if err := m.removeEntriesFromAclSingle(traceID, entries[0:cnt], resAcl, aclID); err != nil {
			return err
		}
		entries = entries[cnt:]
		total = len(entries)
	}
	if total <= 0 {
		return nil
	}

	return m.removeEntriesFromAclSingle(traceID, entries, resAcl, aclID)
}

func (m *ALBProvider) removeEntriesFromAclSingle(traceID interface{}, entries []albsdk.AclEntry, resAcl *alb.Acl, aclID string) error {
	removeEntriesFromAclReq := albsdk.CreateRemoveEntriesFromAclRequest()
	removeEntriesFromAclReq.AclId = aclID
	removeAclEntries := make([]string, 0)
	for _, entry := range entries {
		removeAclEntries = append(removeAclEntries, entry.Entry)
	}
	removeEntriesFromAclReq.Entries = &removeAclEntries
	if err := util.RetryImmediateOnError(m.waitAclExistencePollInterval, m.waitAclExistenceTimeout, isQuotaExceededError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("remove entries from acl",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.RemoveEntriesFromAcl)
		removeEntriesFromAclResp, err := m.auth.ALB.RemoveEntriesFromAcl(removeEntriesFromAclReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("remove entries from acl",
				"stackID", resAcl.Stack().StackID(),
				"resourceID", resAcl.ID(),
				"traceID", traceID,
				"aclID", resAcl.Spec.AclId,
				"error", err.Error(),
				util.Action, util.RemoveEntriesFromAcl)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("remove entries from acl",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"traceID", traceID,
			"aclID", resAcl.Spec.AclId,
			"requestID", removeEntriesFromAclResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.RemoveEntriesFromAcl)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to remove entries from acl")
	}
	return nil
}

func (m *ALBProvider) matchResAndSDKAclEntries(resAclEntries []alb.AclEntry, sdkAclEntries []albsdk.AclEntry) ([]alb.AclEntry, []albsdk.AclEntry) {
	resAclEntriesMap := make(map[string]alb.AclEntry)
	for _, entry := range resAclEntries {
		resAclEntriesMap[entry.Entry] = entry
	}
	sdkAclEntriesMap := make(map[string]albsdk.AclEntry)
	for _, entry := range sdkAclEntries {
		sdkAclEntriesMap[entry.Entry] = entry
	}

	resAclEntriesSet := sets.StringKeySet(resAclEntriesMap)
	sdkAclEntriesSet := sets.StringKeySet(sdkAclEntriesMap)
	var unmatchedResAclEntries []alb.AclEntry
	var unmatchedSDKAclEntries []albsdk.AclEntry
	for _, entry := range resAclEntriesSet.Difference(sdkAclEntriesSet).List() {
		unmatchedResAclEntries = append(unmatchedResAclEntries, resAclEntriesMap[entry])
	}
	for _, entry := range sdkAclEntriesSet.Difference(resAclEntriesSet).List() {
		unmatchedSDKAclEntries = append(unmatchedSDKAclEntries, sdkAclEntriesMap[entry])
	}
	return unmatchedResAclEntries, unmatchedSDKAclEntries
}

func (m *ALBProvider) listAclEntries(traceID interface{}, sdkAclID string) ([]albsdk.AclEntry, error) {
	createListAclEntriesReq := albsdk.CreateListAclEntriesRequest()
	createListAclEntriesReq.AclId = sdkAclID

	var aclList []albsdk.AclEntry
	var nextToken string

	for {
		createListAclEntriesReq.NextToken = nextToken
		var createListAclEntriesResp *albsdk.ListAclEntriesResponse

		if err := util.RetryImmediateOnError(m.waitAclExistencePollInterval, m.waitAclExistenceTimeout, isQuotaExceededError, func() error {
			startTime := time.Now()
			m.logger.V(util.MgrLogLevel).Info("list entries",
				"aclID", sdkAclID,
				"traceID", traceID,
				"startTime", startTime,
				util.Action, util.ListAclEntries)
			var err error
			createListAclEntriesResp, err = m.auth.ALB.ListAclEntries(createListAclEntriesReq)
			if err != nil {
				m.logger.V(util.MgrLogLevel).Info("list entries",
					"aclID", sdkAclID,
					"traceID", traceID,
					"error", err.Error(),
					util.Action, util.ListAclEntries)
				return err
			}
			m.logger.V(util.MgrLogLevel).Info("list entries",
				"traceID", traceID,
				"aclID", sdkAclID,
				"listAclEntries", createListAclEntriesResp.AclEntries,
				"requestID", createListAclEntriesResp.RequestId,
				"elapsedTime", time.Since(startTime).Milliseconds(),
				util.Action, util.ListAclEntries)
			return nil
		}); err != nil {
			return []albsdk.AclEntry{}, errors.Wrap(err, "failed to list entries")
		}
		aclList = append(aclList, createListAclEntriesResp.AclEntries...)

		if createListAclEntriesResp.NextToken == "" {
			break
		} else {
			nextToken = createListAclEntriesResp.NextToken
		}
	}

	return aclList, nil
}

func (m *ALBProvider) associateAclWithListener(ctx context.Context, traceID interface{}, resAcl *alb.Acl, aclID string) error {
	associateAclsWithListenerReq := albsdk.CreateAssociateAclsWithListenerRequest()
	listenerID, err := resAcl.Spec.ListenerID.Resolve(ctx)
	associateAclsWithListenerReq.ListenerId = listenerID
	associateAclsWithListenerReq.AclType = resAcl.Spec.AclType
	associateAclsWithListenerReq.AclIds = &[]string{aclID}
	if err != nil {
		return err
	}
	if err := util.RetryImmediateOnError(m.waitAclExistencePollInterval, m.waitAclExistenceTimeout, isQuotaExceededError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("associate acl with listener",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"listenerID", listenerID,
			"aclID", aclID,
			"aclType", associateAclsWithListenerReq.AclType,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.AssociateAclsWithListener)
		associateAclsWithListenerResp, err := m.auth.ALB.AssociateAclsWithListener(associateAclsWithListenerReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("associate acl with listener",
				"stackID", resAcl.Stack().StackID(),
				"resourceID", resAcl.ID(),
				"listenerID", listenerID,
				"aclID", aclID,
				"aclType", associateAclsWithListenerReq.AclType,
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.AssociateAclsWithListener)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("associate acl with listener",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"traceID", traceID,
			"listenerID", listenerID,
			"aclID", aclID,
			"aclType", associateAclsWithListenerReq.AclType,
			"requestID", associateAclsWithListenerResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.AssociateAclsWithListener)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to associate acl with listener")
	}
	return nil
}

func (m *ALBProvider) addEntriesToAcl(traceID interface{}, resAclEntries []alb.AclEntry, resAcl *alb.Acl, aclID string) error {
	if len(resAclEntries) == 0 {
		return nil
	}
	cnt := util.BatchAddEntriesToAclMaxNum
	total := len(resAclEntries)

	for total > cnt {
		if err := m.addEntriesToAclSingle(traceID, resAclEntries[0:cnt], resAcl, aclID); err != nil {
			return err
		}
		resAclEntries = resAclEntries[cnt:]
		total = len(resAclEntries)
	}
	if total <= 0 {
		return nil
	}

	return m.addEntriesToAclSingle(traceID, resAclEntries, resAcl, aclID)
}

func (m *ALBProvider) addEntriesToAclSingle(traceID interface{}, resAclEntries []alb.AclEntry, resAcl *alb.Acl, aclID string) error {
	if len(resAclEntries) == 0 {
		return nil
	}
	addEntriesToAclReq := albsdk.CreateAddEntriesToAclRequest()
	addEntriesToAclReq.AclId = aclID
	addEntries := make([]albsdk.AddEntriesToAclAclEntries, 0)
	for _, entry := range resAclEntries {
		addEntries = append(addEntries, albsdk.AddEntriesToAclAclEntries{
			Entry: entry.Entry,
		})
	}
	addEntriesToAclReq.AclEntries = &addEntries
	if err := util.RetryImmediateOnError(m.waitAclExistencePollInterval, m.waitAclExistenceTimeout, isQuotaExceededError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("add entries to acl",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.AddEntriesToAclALBAcl)
		addEntriesToAclResp, err := m.auth.ALB.AddEntriesToAcl(addEntriesToAclReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("add entries to acl",
				"stackID", resAcl.Stack().StackID(),
				"resourceID", resAcl.ID(),
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.AddEntriesToAclALBAcl)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("add entries to acl",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"traceID", traceID,
			"aclID", addEntriesToAclReq.AclId,
			"requestID", addEntriesToAclResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.AddEntriesToAclALBAcl)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to add entries acl")
	}

	return nil
}

func (m *ALBProvider) createAcl(traceID interface{}, resAcl *alb.Acl) (*albsdk.CreateAclResponse, error) {
	createAclReq := albsdk.CreateCreateAclRequest()
	createAclReq.AclName = resAcl.Spec.AclName
	var createAclResp *albsdk.CreateAclResponse
	if err := util.RetryImmediateOnError(m.waitAclExistencePollInterval, m.waitAclExistenceTimeout, isQuotaExceededError, func() error {
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("creating acl",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.CreateAcl)
		var err error
		createAclResp, err = m.auth.ALB.CreateAcl(createAclReq)
		if err != nil {
			m.logger.V(util.MgrLogLevel).Info("creating acl",
				"stackID", resAcl.Stack().StackID(),
				"resourceID", resAcl.ID(),
				"traceID", traceID,
				"error", err.Error(),
				util.Action, util.CreateAcl)
			return err
		}
		m.logger.V(util.MgrLogLevel).Info("created acl",
			"stackID", resAcl.Stack().StackID(),
			"resourceID", resAcl.ID(),
			"traceID", traceID,
			"aclID", createAclResp.AclId,
			"requestID", createAclResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.CreateAcl)
		return nil
	}); err != nil {
		return createAclResp, errors.Wrap(err, "failed to create acl")
	}
	return createAclResp, nil
}

func (m *ALBProvider) waitAclStatus(traceID interface{}, aclId string) error {
	stopCh := make(chan struct{})
	var err error
	util.WaitUntilStop(m.waitAclExistencePollInterval, func() (bool, error) {
		var sdkacl albsdk.Acl
		sdkacl, err = m.getAclById(traceID, aclId)
		if err != nil {
			return true, err
		}
		if sdkacl.AclStatus != util.AclStatusAvailable {
			return false, nil
		}
		return true, nil
	}, stopCh)
	return err
}

func (m *ALBProvider) getAclById(traceID interface{}, aclId string) (albsdk.Acl, error) {
	listAclsReq := albsdk.CreateListAclsRequest()
	aclIds := aclId
	listAclsReq.AclIds = &[]string{aclIds}
	var (
		nextToken string
		acls      []albsdk.Acl
	)
	for {
		listAclsReq.NextToken = nextToken
		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("listing acls",
			"aclID", aclId,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.ListAcl)
		listAclResp, err := m.auth.ALB.ListAcls(listAclsReq)
		if err != nil {
			return albsdk.Acl{}, err
		}
		m.logger.V(util.MgrLogLevel).Info("listed acls",
			"aclID", aclId,
			"traceID", traceID,
			"requestID", listAclResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListAcl)

		acls = append(acls, listAclResp.Acls...)

		if listAclResp.NextToken == "" {
			break
		} else {
			nextToken = listAclResp.NextToken
		}
	}
	if len(acls) == 0 {
		return albsdk.Acl{}, fmt.Errorf("list acls failed with %s", aclId)
	}

	return acls[0], nil
}

func isQuotaExceededError(err error) bool {
	return strings.Contains(err.Error(), "QuotaExceeded.AclsNum")
}

func buildResAclStatus(aclID string) alb.AclStatus {
	return alb.AclStatus{
		AclID: aclID,
	}
}
