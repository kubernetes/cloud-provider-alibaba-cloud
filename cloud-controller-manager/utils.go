/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alicloud

import (
	"fmt"
	"github.com/docker/distribution/uuid"
	"k8s.io/api/core/v1"
	"strings"
	"sync"
)

// localService is a local cache try to record the max resource version of each service.
// this is a workaround of BUG #https://github.com/kubernetes/kubernetes/issues/59084
var (
	versionCache *localService
	serviceCache *kvstore
	once         sync.Once
)

type localService struct {
	maxResourceVersion map[string]bool
	lock               sync.RWMutex
}

type kvstore struct {
	store map[string]int64
	lock  sync.RWMutex
}

func init() {
	once.Do(func() {
		versionCache = &localService{
			maxResourceVersion: map[string]bool{},
		}
		serviceCache = &kvstore{
			store: map[string]int64{},
		}
	})
}

// GetPrivateZoneRecordCache return record cache
func GetPrivateZoneRecordCache() *kvstore {
	return serviceCache
}

// GetLocalService return service cache
func GetLocalService() *localService {
	return versionCache
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

func (kv *kvstore) get(key string) (int64, bool) {
	kv.lock.RLock()
	defer kv.lock.RUnlock()

	value, found := kv.store[key]
	return value, found
}

func (kv *kvstore) set(key string, value int64) {
	kv.lock.Lock()
	defer kv.lock.Unlock()
	kv.store[key] = value
}

func (kv *kvstore) remove(key string) {
	kv.lock.Lock()
	defer kv.lock.Unlock()
	delete(kv.store, key)
}

// NodeList return nodes list in string
func NodeList(nodes []*v1.Node) []string {
	ns := []string{}
	for _, node := range nodes {
		ns = append(ns, node.Name)
	}
	return ns
}

// Contains contains in
func Contains(list []int, x int) bool {
	for _, item := range list {
		if item == x {
			return true
		}
	}
	return false
}

func newid() string {
	return fmt.Sprintf("lb-%s", strings.ToLower(uuid.Generate().String())[0:15])
}
