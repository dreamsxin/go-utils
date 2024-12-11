package file

import (
	"sync"

	"github.com/gofrs/flock"
)

type Filelock struct {
	l     sync.Mutex
	locks sync.Map
}

func New() *Filelock {
	filelock := Filelock{}
	return &filelock
}

func (lock *Filelock) GetFileLock(path string) *flock.Flock {
	filelock := flock.New(path)
	if v, loaded := lock.locks.LoadOrStore(path, filelock); loaded {
		return v.(*flock.Flock)
	}
	return filelock
}

func (lock *Filelock) DelFileLock(path string) {
	lock.l.Lock()
	defer lock.l.Unlock()

	v, loaded := lock.locks.LoadAndDelete(path)
	if loaded {
		v.(*flock.Flock).Unlock()
	}
}

func (lock *Filelock) Lock(path string) error {
	fileLock := lock.GetFileLock(path)

	err := fileLock.Lock()
	if err != nil {
		return err
	}
	return nil
}

func (lock *Filelock) TryLock(path string) bool {
	fileLock := lock.GetFileLock(path)

	locked, err := fileLock.TryLock()
	if err != nil {
		return false
	}
	return locked
}

func (lock *Filelock) Unlock(path string) {
	lock.DelFileLock(path)
}

var defaultFilelock *Filelock

func init() {
	defaultFilelock = New()
}

func Lock(path string) {
	defaultFilelock.Lock(path)
}

func TryLock(path string) bool {
	return defaultFilelock.TryLock(path)
}

func Unlock(path string) {
	defaultFilelock.Unlock(path)
}
