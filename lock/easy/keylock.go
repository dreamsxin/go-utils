package easy

import (
	"hash/crc32"
	"sync"
)

type EasyKeylock struct {
	lock_count uint32
	locks      []sync.Mutex
	table      *crc32.Table
}

func New(lock_count uint32) *EasyKeylock {
	table := crc32.MakeTable(crc32.Koopman)
	keylock := EasyKeylock{locks: make([]sync.Mutex, lock_count), table: table}
	keylock.lock_count = lock_count
	return &keylock
}

func (lock *EasyKeylock) Lock(key string) {
	lock.locks[lock.KeyToIndex(key)].Lock()
}

func (lock *EasyKeylock) TryLock(key string) bool {
	return lock.locks[lock.KeyToIndex(key)].TryLock()
}

func (lock *EasyKeylock) Unlock(key string) {
	lock.locks[lock.KeyToIndex(key)].Unlock()
}

func (lock *EasyKeylock) KeyToIndex(key string) uint32 {
	return crc32.Checksum([]byte(key), lock.table) % lock.lock_count
}

var defaultEasyKeylock *EasyKeylock

func init() {
	defaultEasyKeylock = New(4096)
}

func Lock(key string) {
	defaultEasyKeylock.locks[defaultEasyKeylock.KeyToIndex(key)].Lock()
}

func TryLock(key string) bool {
	return defaultEasyKeylock.locks[defaultEasyKeylock.KeyToIndex(key)].TryLock()
}

func Unlock(key string) {
	defaultEasyKeylock.locks[defaultEasyKeylock.KeyToIndex(key)].Unlock()
}

func KeyToIndex(key string) uint32 {
	return defaultEasyKeylock.KeyToIndex(key)
}
