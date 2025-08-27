package lock

import (
	"sync"
	"sync/atomic"
)

type refCounter struct {
	waitGroup sync.WaitGroup
	lock      *sync.RWMutex
	counter   int32
}

// MultipleLock is the main interface for lock based on key
type MultipleLock interface {
	// Lock based on the key
	Lock(key interface{})

	// TryLock tries to lock based on the key, returns true if successful
	TryLock(key interface{}) bool

	// RLock lock the rw for reading
	RLock(key interface{})

	// Unlock the key
	Unlock(key interface{})

	// RUnlock the read lock
	RUnlock(key interface{})

	// Wait for all operations on the key to complete
	Wait(key interface{})
}

type lock struct {
	inUse sync.Map
	pool  *sync.Pool
}

func (l *lock) Lock(key interface{}) {
	m := l.getLocker(key)
	atomic.AddInt32(&m.counter, 1)
	m.waitGroup.Add(1)
	m.lock.Lock()
}

func (l *lock) TryLock(key interface{}) bool {
	m := l.getLocker(key)
	if !m.lock.TryLock() {
		// 如果获取锁失败，需要减少计数器
		if atomic.AddInt32(&m.counter, -1) == 0 {
			l.pool.Put(m.lock)
			l.inUse.Delete(key)
		}
		return false
	}
	atomic.AddInt32(&m.counter, 1)
	m.waitGroup.Add(1)
	return true
}

func (l *lock) RLock(key interface{}) {
	m := l.getLocker(key)
	atomic.AddInt32(&m.counter, 1)
	m.waitGroup.Add(1)
	m.lock.RLock()
}

func (l *lock) Unlock(key interface{}) {
	if m, ok := l.inUse.Load(key); ok {
		ref := m.(*refCounter)
		ref.lock.Unlock()
		ref.waitGroup.Done()
		if atomic.AddInt32(&ref.counter, -1) == 0 {
			l.pool.Put(ref.lock)
			l.inUse.Delete(key)
		}
	}
}

func (l *lock) RUnlock(key interface{}) {
	if m, ok := l.inUse.Load(key); ok {
		ref := m.(*refCounter)
		ref.lock.RUnlock()
		ref.waitGroup.Done()
		if atomic.AddInt32(&ref.counter, -1) == 0 {
			l.pool.Put(ref.lock)
			l.inUse.Delete(key)
		}
	}
}

func (l *lock) Wait(key interface{}) {
	if m, ok := l.inUse.Load(key); ok {
		m.(*refCounter).waitGroup.Wait()
	}
}

func (l *lock) getLocker(key interface{}) *refCounter {
	actual, loaded := l.inUse.LoadOrStore(key, &refCounter{
		counter: 0,
		lock:    l.pool.Get().(*sync.RWMutex),
	})

	if !loaded {
		return actual.(*refCounter)
	}

	// 如果已存在，增加计数器
	ref := actual.(*refCounter)
	atomic.AddInt32(&ref.counter, 1)
	return ref
}

// NewMultipleLock creates a new multiple lock
func NewMultipleLock() MultipleLock {
	return &lock{
		pool: &sync.Pool{
			New: func() interface{} {
				return &sync.RWMutex{}
			},
		},
	}
}
