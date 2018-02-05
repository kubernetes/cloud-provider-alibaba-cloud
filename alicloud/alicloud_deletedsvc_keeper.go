package alicloud

import "sync"

// This keeper is try to resource version
var svcRVKeeper *deletedSVCKeeper
var once sync.Once

type deletedSVCKeeper struct {
	maxSVCResourceVersionMap map[string]bool
	versionMapLock           sync.RWMutex
}

func GetDeletedSvcKeeper() *deletedSVCKeeper {
	once.Do(func() {
		svcRVKeeper = &deletedSVCKeeper{
			maxSVCResourceVersionMap: map[string]bool{},
		}
	})
	return svcRVKeeper
}

func (s *deletedSVCKeeper) set(serviceUID string) {
	s.versionMapLock.Lock()
	// s.versionMapLock.Lock()
	defer s.versionMapLock.Unlock()
	s.maxSVCResourceVersionMap[serviceUID] = true
}

func (s *deletedSVCKeeper) get(serviceUID string) (found bool) {
	s.versionMapLock.RLock()
	defer s.versionMapLock.RUnlock()
	_, found = s.maxSVCResourceVersionMap[serviceUID]
	return
}
