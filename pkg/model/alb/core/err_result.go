package core

import (
	"github.com/pkg/errors"
)

type ErrInfo struct {
	ErrMsgs []error
}

// ErrResult
//The key of the ErrResultMap is an int value represents listenerPort,
//and the value of the ErrResultMap is an error slice
type ErrResult struct {
	ErrResultMap map[int]ErrInfo
}

func NewErrMessages(errMsgs []error) ErrInfo {
	return ErrInfo{
		ErrMsgs: errMsgs,
	}
}

func NewDefaultErrResult() ErrResult {
	return ErrResult{
		ErrResultMap: make(map[int]ErrInfo),
	}
}

func isErrMsgEmpty(errMsgs []error) bool {
	return len(errMsgs) == 0
}

func (e *ErrResult) CheckErrMsgsByListenerPort(ListenerPort int) error {
	errResMap := e.ErrResultMap
	errInfo, ok := errResMap[ListenerPort]
	if ok && !isErrMsgEmpty(errInfo.ErrMsgs) {
		return errors.Errorf("ListenerPort exists in ErrResult, ListenerPort: %v", ListenerPort)
	}
	return nil
}

func (e *ErrResult) FindErrMsgsByListenerPort(ListenerPort int) ([]error, error) {
	if e.CheckErrMsgsByListenerPort(ListenerPort) == nil {
		return nil, nil
	}
	errResMap := e.ErrResultMap
	errInfo, ok := errResMap[ListenerPort]
	if ok && !isErrMsgEmpty(errInfo.ErrMsgs) {
		errMsgs := errInfo.ErrMsgs
		return errMsgs, errors.Errorf("ListenerPort exists in ErrResult, ListenerPort: %v", ListenerPort)
	}
	return nil, nil
}

func (e *ErrResult) AddErrMsgsWithListenerPort(ListenerPort int, errMsg error) error {
	if errMsgs, err := e.FindErrMsgsByListenerPort(ListenerPort); err != nil {
		errMsgs = append(errMsgs, errMsg)
		e.ErrResultMap[ListenerPort] = NewErrMessages(errMsgs)
		return nil
	}
	var newErrMsgs []error
	newErrMsgs = append(newErrMsgs, errMsg)
	e.ErrResultMap[ListenerPort] = NewErrMessages(newErrMsgs)
	return nil
}
