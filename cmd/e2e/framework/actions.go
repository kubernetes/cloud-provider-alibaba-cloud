package framework

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	cloud "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func NewDefaultAction(u *TestUnit) Action {
	return &DefaultAction{TestUnit: u}
}

type Action interface {
	RunAction(f *FrameWorkE2E) error
}

type DefaultAction struct{ *TestUnit }

func (u *DefaultAction) RunAction(f *FrameWorkE2E) error {
	f.Logf("default action: %s", u.Description)
	newm, err := WaitServiceMutate(f.Client, f.InitService, u.Mutator)
	if err != nil {
		return fmt.Errorf("mutator service: %s", err.Error())
	}
	f.InitService = newm
	fa := cloud.NewFrameWorkWithOptions(
		func(fa *cloud.FrameWork) {
			fa.SVC = f.InitService
			fa.Cloud = NewAlibabaCloudOrDie(TestContext.CloudConfig)
		},
	)
	nodes, err := f.Client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("get node: %s", err.Error())
	}
	svc, err := f.
		Client.
		CoreV1().
		Services(f.InitService.Namespace).
		Get(f.InitService.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get service: %s", err.Error())
	}
	fa.Nodes = ToPTR(nodes.Items)
	fa.SVC = svc

	endps, err := f.
		Client.
		CoreV1().
		Endpoints(f.InitService.Namespace).
		Get(f.InitService.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get endpoints: %s", err.Error())
	}
	fa.Endpoint = endps
	ExpectOK := u.ExpectOK
	if ExpectOK == nil {
		ExpectOK = cloud.ExpectExistAndEqual
	}
	return WaitTimeout(f.Test, fa, ExpectOK)
}

func NewDeleteAction(u *TestUnit) *DeleteAction {
	return &DeleteAction{TestUnit: u}
}

type DeleteAction struct{ *TestUnit }

func (u *DeleteAction) RunAction(f *FrameWorkE2E) error {
	f.Logf("delete action: %s", u.Description)
	svc, err := f.
		Client.
		CoreV1().
		Services(f.InitService.Namespace).
		Get(f.InitService.Name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			f.Logf("service not found, expected")
			return nil
		}
		return fmt.Errorf("unexpected error: %s", err.Error())
	}
	err = f.
		Client.
		CoreV1().
		Services(f.InitService.Namespace).
		Delete(f.InitService.Name, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("delete service: %s", err.Error())
	}
	fa := cloud.NewFrameWorkWithOptions(
		func(fa *cloud.FrameWork) {
			fa.SVC = svc
			fa.Cloud = NewAlibabaCloudOrDie(TestContext.CloudConfig)
		},
	)
	ExpectOK := u.ExpectOK
	if ExpectOK == nil {
		ExpectOK = cloud.ExpectNotExist
	}
	return WaitTimeout(f.Test, fa, ExpectOK)
}

func WaitTimeout(
	test *testing.T,
	f *cloud.FrameWork,
	ExpectExistAndEqual func(f *cloud.FrameWork) error,
) error {

	return wait.PollImmediate(
		4*time.Second,
		1*time.Minute,
		func() (done bool, err error) {
			err = ExpectExistAndEqual(f)
			if err != nil {
				test.Logf("WaitForTimeout: %s, %s\n", time.Now(), err.Error())
				return false, nil
			}
			return true, nil
		},
	)
}

func NewRandomAction(rand []Action) Action {
	return &RandomAction{random: rand}
}

type RandomAction struct{ random []Action }

func (u *RandomAction) RunAction(f *FrameWorkE2E) error {
	rand.Shuffle(
		len(u.random),
		func(i, j int) {
			u.random[i], u.random[j] = u.random[j], u.random[i]
		},
	)
	for _, action := range u.random {
		if err := action.RunAction(f); err != nil {
			return fmt.Errorf("run random action: %s", err.Error())
		}
	}
	return nil
}
