package common

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/test/e2e/framework"
	"k8s.io/klog/v2"
)

type aEvent struct {
	FirstTimeStamp time.Time
	LastTimeStamp  time.Time
	Reason         string
	Message        string
	Type           string
	InvokeObject   v1.ObjectReference
}

func PrintEventsWhenError(f *framework.Framework) string {
	klog.Info("Looks like something went wrong~")
	klog.Info("`kubectl -n e2e-test get events`")
	results, err := f.Client.KubeClient.CoreV1().Events("e2e-test").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Info("list e2e-test events failed", err)
	}
	followEvts := make([]aEvent, 0)
	for _, evt := range results.Items {
		if evt.InvolvedObject.Kind != "Ingress" {
			continue
		}
		if strings.HasPrefix("Successfully", evt.Message) || strings.HasPrefix("Scheduled", evt.Message) {
			continue
		}
		followEvts = append(followEvts, aEvent{
			FirstTimeStamp: evt.FirstTimestamp.Time,
			LastTimeStamp:  evt.LastTimestamp.Time,
			Reason:         evt.Reason,
			Message:        evt.Message,
			Type:           evt.Type,
			InvokeObject:   evt.InvolvedObject,
		})
	}

	eventByte, _ := json.MarshalIndent(followEvts, "", "  ")
	eventString := string(eventByte)
	return eventString
}
