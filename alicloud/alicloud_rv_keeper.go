package alicloud

import "sync"

// This keeper is try to resource version
var svcRVKeeper *svcResourceVersionKeeper
var once sync.Once

type svcResourceVersionKeeper struct {
	maxSVCResourceVersionMap map[string]uint64
	versionMapLock           sync.RWMutex
}

func GetSvcResourceVersionKeeper() *svcResourceVersionKeeper {
	once.Do(func() {
		svcRVKeeper = &svcResourceVersionKeeper{
			maxSVCResourceVersionMap: map[string]uint64{},
		}
	})
	return svcRVKeeper
}

func (s *svcResourceVersionKeeper) set(serviceUID string, resourceVersion uint64) {
	s.versionMapLock.Lock()
	// s.versionMapLock.Lock()
	defer s.versionMapLock.Unlock()
	s.maxSVCResourceVersionMap[serviceUID] = resourceVersion
}

func (s *svcResourceVersionKeeper) get(serviceUID string) (resourceVersion uint64, found bool) {
	s.versionMapLock.RLock()
	defer s.versionMapLock.RUnlock()
	resourceVersion, found = s.maxSVCResourceVersionMap[serviceUID]
	return
}
