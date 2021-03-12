package health

import "fmt"

var (
	CRDReady      bool
	CheckFuncList = []Checker{
		&DefaultHealthCheck{}, &CustomizeHealthCheck{},
	}
)

type Checker interface {
	Check() error
}

type DefaultHealthCheck struct {
}

type CustomizeHealthCheck struct {
}

func (dh *DefaultHealthCheck) Check() error {
	if CRDReady != true {
		return fmt.Errorf("crd ready check failed")
	}
	return nil
}
func (ch *CustomizeHealthCheck) Check() error {
	return nil
}
