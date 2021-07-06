package framework

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"math/rand"
	"time"
)

func NewErrorRetry(err error) *ErrorRetry {
	return &ErrorRetry{Err: err}
}

type ErrorRetry struct {
	Err error
}

func (m *ErrorRetry) Error() string {
	return fmt.Sprintf("NeedRetry: [%v]", m.Err)
}

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
	// TODO: need rewrite?
	u.Service = newm
	u.NewReqContext(f.Cloud)
	return WaitExpection(f, u.TestUnit)
}

func WaitExpection(f *FrameWorkE2E, u *TestUnit) error {
	return wait.PollImmediate(
		4*time.Second,
		1*time.Minute,
		func() (done bool, err error) {
			expect := NewExpection(u, f)
			ok, err := expect.ExpectOK()
			if err != nil {
				merr, eok := err.(*ErrorRetry)
				if eok {
					Logf("retry on expectation: %s", merr.Error())
					return false, nil
				}
				return false, errors.Wrap(err, expect.Case.Description)
			}

			ExpectEqual(ok, true, "expect failed: %s", expect.Case.Description)
			return true, nil
		},
	)
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
		if apierrors.IsNotFound(err){
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
	u.Service = svc
	u.NewReqContext(f.Cloud)
	return WaitExpection(f, u.TestUnit)
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
