package health

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

func (dh *DefaultHealthCheck) Check() error {
	return nil
}

type CustomizeHealthCheck struct {
}

func (ch *CustomizeHealthCheck) Check() error {
	return nil
}
