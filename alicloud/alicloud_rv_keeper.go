package alicloud

import "sync"

// This keeper is try to resource version
var svcRVKeeper *svcResourceVersionKeeper
var once sync.Once

type svcResourceVersionKeeper struct {
	knownSVCResourceVersionMap map[string]uint64
	versionMapLock             sync.RWMutex
}

func GetSvcResourceVersionKeeper() *svcResourceVersionKeeper {
	once.Do(func() {
		svcRVKeeper = &svcResourceVersionKeeper{
			knownSVCResourceVersionMap: map[string]uint64{},
		}
	})
	return svcRVKeeper
}
