package framework

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"math/rand"
	"strings"
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
	Logf("default action: %s", u.Description)
	newm, err := WaitServiceMutate(f.Client, u.Service, u.Mutator)
	if err != nil {
		return fmt.Errorf("mutator service: %s", err.Error())
	}
	u.Service = newm
	nodes, err := f.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("get node: %s", err.Error())
	}
	svc, err := f.
		Client.
		CoreV1().
		Services(u.Service.Namespace).
		Get(context.Background(), u.Service.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get service: %s", err.Error())
	}
	expect := NewExpection(u.TestUnit, f)
	expect.Nodes = nodes.Items
	expect.Svc = svc

	endps, err := f.
		Client.
		CoreV1().
		Endpoints(u.Service.Namespace).
		Get(context.Background(), u.Service.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get endpoints: %s", err.Error())
	}
	expect.Endpoint = endps
	return WaitTimeout(expect)
}

func NewDeleteAction(u *TestUnit) *DeleteAction {
	return &DeleteAction{TestUnit: u}
}

type DeleteAction struct{ *TestUnit }

func (u *DeleteAction) RunAction(f *FrameWorkE2E) error {
	Logf("delete action: %s", u.Description)
	svc, err := f.
		Client.
		CoreV1().
		Services(u.Service.Namespace).
		Get(context.Background(), u.Service.Name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			Logf("service not found, expected")
			return nil
		}
		return fmt.Errorf("unexpected error: %s", err.Error())
	}
	err = f.Client.
		CoreV1().
		Services(u.Service.Namespace).
		Delete(context.Background(), u.Service.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("delete service: %s", err.Error())
	}
	expect := NewExpection(u.TestUnit, f)
	expect.Svc = svc
	return WaitTimeout(expect)
}

func WaitTimeout(
	expect *Expectation,
) error {

	return wait.PollImmediate(
		4*time.Second,
		1*time.Minute,
		func() (done bool, err error) {
			err = expect.ExpectOK()
			if err != nil {
				Logf("WaitFor Expectation(%s) Timeout: %s, %s", expect.Case.Description, time.Now(), err.Error())
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
