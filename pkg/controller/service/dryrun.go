package service

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

var initial = sync.Map{}

func initMap(client client.Client) error {
	svcs := v1.ServiceList{}
	err := client.List(context.TODO(), &svcs)
	if err != nil {
		return fmt.Errorf("init map fail: %s", err.Error())
	}

	length := 0
	for _, m := range svcs.Items {
		if !isProcessNeeded(&m) {
			continue
		}
		if !needAdd(&m) {
			continue
		}
		length++
		initial.Store(util.Key(&m), 0)
	}
	if length == 0 {
		util.ServiceLog.Info("ccm initial process finished.")
		err := dryrun.ResultEvent(client, dryrun.SUCCESS, "ccm initial process finished")
		if err != nil {
			util.ServiceLog.Error(err, "fail to write precheck event")
		}
		os.Exit(0)
	}
	return nil
}

func mapfull() bool {
	total, unsync := 0, 0
	initial.Range(
		func(key, value interface{}) bool {
			val, ok := value.(int)
			if !ok {
				// not supposed
				return true
			}
			if val != 1 {
				unsync += 1
			}
			total += 1
			return true
		},
	)
	util.ServiceLog.Info("Reconcile process", "total", total, "unsync", unsync)
	return unsync == 0
}
