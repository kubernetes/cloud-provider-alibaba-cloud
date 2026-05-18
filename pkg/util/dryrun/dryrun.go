package dryrun

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Instance struct {
	lock       sync.Mutex
	checkNames sets.Set[string]
	done       chan struct{}
}

var global = &Instance{
	checkNames: sets.New[string](),
	done:       make(chan struct{}),
}

func (i *Instance) RegisterDryRun(checkName string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	global.checkNames.Insert(checkName)
}

func (i *Instance) Finish(checkName string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	global.checkNames.Delete(checkName)
	if global.checkNames.Len() == 0 {
		close(global.done)
	}
}

func (i *Instance) Done() <-chan struct{} {
	return global.done
}

func RegisterDryRun(checkName string) {
	global.RegisterDryRun(checkName)
}

func Finish(checkName string) {
	global.Finish(checkName)
}

func Done() <-chan struct{} {
	return global.Done()
}
