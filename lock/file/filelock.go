package file

import (
	"sync"

	"github.com/gofrs/flock"
)

// Filelock 提供基于文件路径的分布式锁管理
// 使用 sync.Map 来管理不同路径的锁实例
type Filelock struct {
	locks sync.Map // key: string(path), value: *flock.Flock
}

// New 创建一个新的文件锁管理器
func New() *Filelock {
	return &Filelock{}
}

// GetFileLock 获取或创建指定路径的文件锁
// 如果路径已存在锁，则返回现有锁，否则创建新锁
func (lock *Filelock) GetFileLock(path string) *flock.Flock {
	// 使用 LoadOrStore 确保线程安全
	if v, loaded := lock.locks.LoadOrStore(path, flock.New(path)); loaded {
		return v.(*flock.Flock)
	}
	return lock.getLoadedLock(path)
}

// getLoadedLock 获取已加载的锁，用于内部处理
func (lock *Filelock) getLoadedLock(path string) *flock.Flock {
	v, _ := lock.locks.Load(path)
	return v.(*flock.Flock)
}

// DelFileLock 删除指定路径的文件锁并释放资源
// 如果锁存在且被锁定，则会先解锁
func (lock *Filelock) DelFileLock(path string) {
	if v, loaded := lock.locks.LoadAndDelete(path); loaded {
		fileLock := v.(*flock.Flock)
		// 尝试解锁，忽略错误（可能已经解锁）
		_ = fileLock.Unlock()
	}
}

// Lock 锁定指定路径的文件
// 如果锁已被其他进程持有，则会阻塞直到获取锁
func (lock *Filelock) Lock(path string) error {
	fileLock := lock.GetFileLock(path)
	return fileLock.Lock()
}

// TryLock 尝试锁定指定路径的文件
// 非阻塞操作，立即返回是否成功获取锁
func (lock *Filelock) TryLock(path string) (bool, error) {
	fileLock := lock.GetFileLock(path)
	return fileLock.TryLock()
}

// Unlock 解锁指定路径的文件
// 同时从管理器中移除该锁
func (lock *Filelock) Unlock(path string) error {
	if v, loaded := lock.locks.Load(path); loaded {
		fileLock := v.(*flock.Flock)
		err := fileLock.Unlock()
		if err != nil {
			return err
		}
		lock.locks.Delete(path)
	}
	return nil
}

// WithLock 使用闭包执行受锁保护的代码
// 自动处理锁的获取和释放，确保资源正确清理
func (lock *Filelock) WithLock(path string, fn func() error) error {
	if err := lock.Lock(path); err != nil {
		return err
	}
	defer lock.Unlock(path) // 使用命名返回值确保解锁
	return fn()
}

// 全局默认文件锁实例
var (
	defaultFilelock     *Filelock
	defaultFilelockOnce sync.Once
)

// init 初始化默认文件锁实例
func init() {
	defaultFilelockOnce.Do(func() {
		defaultFilelock = New()
	})
}

// Lock 使用默认实例锁定指定路径的文件
func Lock(path string) error {
	return defaultFilelock.Lock(path)
}

// TryLock 使用默认实例尝试锁定指定路径的文件
func TryLock(path string) (bool, error) {
	return defaultFilelock.TryLock(path)
}

// Unlock 使用默认实例解锁指定路径的文件
func Unlock(path string) error {
	return defaultFilelock.Unlock(path)
}

// WithLock 使用默认实例执行受锁保护的代码
func WithLock(path string, fn func() error) error {
	return defaultFilelock.WithLock(path, fn)
}
