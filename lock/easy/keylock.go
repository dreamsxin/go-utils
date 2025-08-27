package easy

import (
	"hash/crc32"
	"sync"
)

// EasyKeylock 提供了一个基于键的分片锁机制
// 通过将键哈希到不同的锁上来减少锁竞争
type EasyKeylock struct {
	lockCount uint32
	locks     []sync.RWMutex
	table     *crc32.Table
}

// New 创建一个新的 EasyKeylock 实例
// lockCount 指定要使用的锁数量，建议使用2的幂次方以获得最佳性能
func New(lockCount uint32) *EasyKeylock {
	if lockCount == 0 {
		lockCount = 1 // 确保至少有一个锁
	}

	// 使用更高效的IEEE表，它在大多数情况下比Koopman表更快
	table := crc32.IEEETable
	return &EasyKeylock{
		locks:     make([]sync.RWMutex, lockCount),
		table:     table,
		lockCount: lockCount,
	}
}

// Lock 基于键获取互斥锁
func (lock *EasyKeylock) Lock(key string) {
	lock.locks[lock.keyToIndex(key)].Lock()
}

// TryLock 尝试基于键获取互斥锁，成功返回true
func (lock *EasyKeylock) TryLock(key string) bool {
	return lock.locks[lock.keyToIndex(key)].TryLock()
}

// Unlock 基于键释放互斥锁
func (lock *EasyKeylock) Unlock(key string) {
	lock.locks[lock.keyToIndex(key)].Unlock()
}

// RLock 基于键获取读锁
func (lock *EasyKeylock) RLock(key string) {
	lock.locks[lock.keyToIndex(key)].RLock()
}

// RUnlock 基于键释放读锁
func (lock *EasyKeylock) RUnlock(key string) {
	lock.locks[lock.keyToIndex(key)].RUnlock()
}

// keyToIndex 将键转换为锁数组索引
// 使用CRC32哈希函数确保键的均匀分布
func (lock *EasyKeylock) keyToIndex(key string) uint32 {
	return crc32.Checksum([]byte(key), lock.table) % lock.lockCount
}

// 全局默认锁实例，使用4096个锁片以减少竞争
var (
	defaultEasyKeylock     *EasyKeylock
	defaultEasyKeylockOnce sync.Once
)

// init 初始化默认锁实例
func init() {
	// 使用sync.Once确保线程安全初始化
	defaultEasyKeylockOnce.Do(func() {
		defaultEasyKeylock = New(4096)
	})
}

// Lock 使用默认锁实例基于键获取互斥锁
func Lock(key string) {
	defaultEasyKeylock.Lock(key)
}

// TryLock 使用默认锁实例尝试基于键获取互斥锁
func TryLock(key string) bool {
	return defaultEasyKeylock.TryLock(key)
}

// Unlock 使用默认锁实例基于键释放互斥锁
func Unlock(key string) {
	defaultEasyKeylock.Unlock(key)
}

// RLock 使用默认锁实例基于键获取读锁
func RLock(key string) {
	defaultEasyKeylock.RLock(key)
}

// RUnlock 使用默认锁实例基于键释放读锁
func RUnlock(key string) {
	defaultEasyKeylock.RUnlock(key)
}

// KeyToIndex 使用默认锁实例将键转换为锁数组索引
// 主要用于调试和监控目的
func KeyToIndex(key string) uint32 {
	return defaultEasyKeylock.keyToIndex(key)
}
