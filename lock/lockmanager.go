package lock

import (
	"sync"
	"sync/atomic"
)

// LockManager 管理基于键的锁
type LockManager interface {
	// Lock 基于键获取互斥锁
	Lock(key interface{})

	// TryLock 尝试基于键获取互斥锁，成功返回true
	TryLock(key interface{}) bool

	// RLock 基于键获取读锁
	RLock(key interface{})

	// Unlock 释放基于键的互斥锁
	Unlock(key interface{})

	// RUnlock 释放基于键的读锁
	RUnlock(key interface{})

	// Wait 等待键上的所有操作完成
	Wait(key interface{})
}

// lockEntry 表示一个锁条目
type lockEntry struct {
	mu       sync.RWMutex
	wg       sync.WaitGroup
	refCount int32
}

// lockManager 锁管理器的实现
type lockManager struct {
	entries sync.Map // key -> *lockEntry
	pool    *sync.Pool
}

// NewLockManager 创建一个新的锁管理器
func NewLockManager() LockManager {
	return &lockManager{
		pool: &sync.Pool{
			New: func() interface{} {
				return &lockEntry{}
			},
		},
	}
}

func (lm *lockManager) Lock(key interface{}) {
	entry := lm.getOrCreateEntry(key)
	atomic.AddInt32(&entry.refCount, 1)
	entry.wg.Add(1)
	entry.mu.Lock()
}

func (lm *lockManager) TryLock(key interface{}) bool {
	entry := lm.getOrCreateEntry(key)

	// 先尝试获取锁
	if !entry.mu.TryLock() {
		// 获取失败，减少引用计数
		if atomic.AddInt32(&entry.refCount, -1) == 0 {
			lm.cleanupEntry(key, entry)
		}
		return false
	}

	// 获取成功，增加引用计数和等待组
	atomic.AddInt32(&entry.refCount, 1)
	entry.wg.Add(1)
	return true
}

func (lm *lockManager) RLock(key interface{}) {
	entry := lm.getOrCreateEntry(key)
	atomic.AddInt32(&entry.refCount, 1)
	entry.wg.Add(1)
	entry.mu.RLock()
}

func (lm *lockManager) Unlock(key interface{}) {
	if entry, ok := lm.entries.Load(key); ok {
		lmEntry := entry.(*lockEntry)
		lmEntry.mu.Unlock()
		lmEntry.wg.Done()
		lm.decrementRefCount(key, lmEntry)
	}
}

func (lm *lockManager) RUnlock(key interface{}) {
	if entry, ok := lm.entries.Load(key); ok {
		lmEntry := entry.(*lockEntry)
		lmEntry.mu.RUnlock()
		lmEntry.wg.Done()
		lm.decrementRefCount(key, lmEntry)
	}
}

func (lm *lockManager) Wait(key interface{}) {
	if entry, ok := lm.entries.Load(key); ok {
		entry.(*lockEntry).wg.Wait()
	}
}

func (lm *lockManager) getOrCreateEntry(key interface{}) *lockEntry {
	entry, loaded := lm.entries.LoadOrStore(key, lm.pool.Get().(*lockEntry))
	if !loaded {
		// 新创建的条目，初始化引用计数
		atomic.StoreInt32(&entry.(*lockEntry).refCount, 1)
	}
	return entry.(*lockEntry)
}

func (lm *lockManager) decrementRefCount(key interface{}, entry *lockEntry) {
	if atomic.AddInt32(&entry.refCount, -1) == 0 {
		lm.cleanupEntry(key, entry)
	}
}

func (lm *lockManager) cleanupEntry(key interface{}, entry *lockEntry) {
	// 重置条目状态
	atomic.StoreInt32(&entry.refCount, 0)
	// 放回池中
	lm.pool.Put(entry)
	// 从映射中删除
	lm.entries.Delete(key)
}
