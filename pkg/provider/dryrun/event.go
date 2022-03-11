package dryrun

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

const (
	FAIL    = "Fail"
	SUCCESS = "Success"
	SLB     = "ccmSLB"
	VPC     = "ccmVPC"
	ECS     = "ccmECS"
	PVTZ    = "ccmPVTZ"
)

type MessageLevel string

const (
	NORMAL  = MessageLevel("normal")
	ERROR   = MessageLevel("error")
	WARN    = MessageLevel("warn")
	UNKNOWN = MessageLevel("unknown")
)

type ContextKey string

const (
	ContextService = ContextKey("ctx.service")
	ContextMessage = ContextKey("ctx.msg")
	ContextSLB     = ContextKey("ctx.slb")
)

const BATCHSIZE = 20

var Message = cache{
	Data: make(map[string]CheckResult),
}

type cache struct {
	Lock sync.RWMutex
	Data map[string]CheckResult
}

func AddEvent(checkName string, name string, id string, msgCode string, msgLevel MessageLevel, msg string) {
	Message.Lock.Lock()
	defer Message.Lock.Unlock()

	checkResult, ok := Message.Data[checkName]
	if !ok {
		checkResult = CheckResult{
			Name:        checkName,
			ItemResults: make(map[string]ItemResult),
		}
	}
	checkResult.ItemResults[name] = ItemResult{
		Name:         name,
		Id:           id,
		MessageLevel: msgLevel,
		MessageCode:  msgCode,
		AdviseCode:   checkName,
		AffectCode:   msgCode,
		Message:      msg,
	}
	switch msgLevel {
	case ERROR:
		checkResult.FailedCount += 1
	case WARN:
		checkResult.WarnCount += 1
	case UNKNOWN:
		checkResult.UnknownCount += 1
	default:
		checkResult.PassCount += 1
	}
	Message.Data[checkName] = checkResult
}

func ResultEvent(client client.Client, status string, reason string) error {
	ens := "kube-system"
	result := Result{
		State:  status,
		Reason: reason,
	}
	Message.Lock.RLock()
	for _, checkResult := range Message.Data {
		checkResult.ItemResults = getBatchItemResult(checkResult.ItemResults)
		result.CheckResults = append(result.CheckResults, checkResult)
	}
	Message.Lock.RUnlock()

	jresult, err := json.Marshal(result)
	if err != nil {
		panic(fmt.Sprintf("should be json: %+v", result))
	}
	event := &v1.Event{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "CloudControllerManagerPrecheckResult",
			Namespace: ens,
		},
		Reason:  status,
		Message: string(jresult),
		Type:    v1.EventTypeNormal,
		InvolvedObject: v1.ObjectReference{
			Kind:       "Pod",
			APIVersion: "v1",
			Name:       "ccm-precheck-pod",
			Namespace:  ens,
		},
	}

	return wait.PollImmediate(
		3*time.Second,
		1*time.Minute,
		func() (done bool, err error) {

			err = createOrUpdate(client, event)
			if err != nil {
				klog.Infof("fire CloudControllerManagerPrecheckResult event fail: %s", err.Error())
				return false, nil
			}
			klog.Infof("create event finished: %s, message: %s", event.Reason, event.Message)
			return true, nil
		},
	)
}

func createOrUpdate(client client.Client, evt *v1.Event) error {
	err := client.Create(context.Background(), evt)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		return client.Update(context.Background(), evt)
	}
	return nil
}

type Result struct {
	State        string
	Reason       string
	CheckResults []CheckResult
}

type CheckResult struct {
	Name         string
	ItemResults  map[string]ItemResult `json:"instanceResults,omitempty"`
	PassCount    int
	FailedCount  int
	WarnCount    int
	UnknownCount int
}

type ItemResult struct {
	Name         string
	Id           string
	MessageLevel MessageLevel
	MessageCode  string
	AffectCode   string
	AdviseCode   string
	Message      string
}

func getBatchItemResult(itemResults map[string]ItemResult) map[string]ItemResult {
	if len(itemResults) <= BATCHSIZE {
		return itemResults
	} else {
		// Separate normal and abnormal results and return abnormal results first
		normalResult := make(map[string]ItemResult)
		abnormalResults := make(map[string]ItemResult)
		for key, value := range itemResults {
			if value.MessageLevel != NORMAL {
				abnormalResults[key] = value
			} else {
				normalResult[key] = value
			}
			if len(abnormalResults) >= BATCHSIZE {
				return abnormalResults
			}
		}

		normalSize := BATCHSIZE - len(abnormalResults)
		for key, value := range normalResult {
			normalSize -= 1
			abnormalResults[key] = value
			if normalSize <= 0 {
				break
			}
		}

		return abnormalResults
	}
}
