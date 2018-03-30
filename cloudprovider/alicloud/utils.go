package alicloud

import "sync"

// localService is a local cache try to record the max resource version of each service.
// this is a workaround of BUG #https://github.com/kubernetes/kubernetes/issues/59084
var (
	cache 	*localService
	once 	sync.Once
)

type localService struct {
	maxResourceVersion map[string]bool
	lock               sync.RWMutex
}

func GetLocalService() *localService {
	once.Do(func() {
		cache = &localService{
			maxResourceVersion: map[string]bool{},
		}
	})
	return cache
}

func (s *localService) set(serviceUID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.maxResourceVersion[serviceUID] = true
}

func (s *localService) get(serviceUID string) (found bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, found = s.maxResourceVersion[serviceUID]
	return
}
