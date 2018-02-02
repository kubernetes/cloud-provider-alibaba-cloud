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
