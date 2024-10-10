package pond

import (
	"sync/atomic"
	"testing"
	"time"
)

func assertEqual(t *testing.T, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Helper()
		t.Errorf("Expected %T(%v) but was %T(%v)", expected, expected, actual, actual)
	}
}

func TestNew(t *testing.T) {

	pool := New(17, 10, MinWorkers(2), IdleTimeout(1*time.Second))
	assertEqual(t, 17, pool.maxWorkers)
	assertEqual(t, 10, pool.maxCapacity)
	assertEqual(t, 2, pool.minWorkers)
	assertEqual(t, 1*time.Second, pool.idleTimeout)
}

func TestNewWithInconsistentOptions(t *testing.T) {

	pool := New(-10, -5, MinWorkers(20), IdleTimeout(-1*time.Second))
	assertEqual(t, 1, pool.maxWorkers)
	assertEqual(t, 0, pool.maxCapacity)
	assertEqual(t, 1, pool.minWorkers)
	assertEqual(t, defaultIdleTimeout, pool.idleTimeout)
}

func TestPurgeAfterPoolStopped(t *testing.T) {

	pool := New(1, 1)

	var doneCount int32
	pool.SubmitAndWait(func() {
		atomic.AddInt32(&doneCount, 1)
	})

	time.Sleep(10 * time.Millisecond)

	assertEqual(t, int32(1), atomic.LoadInt32(&doneCount))
	assertEqual(t, 1, pool.RunningWorkers())

	// Simulate purger goroutine attempting to stop a worker after tasks channel is closed
	atomic.StoreInt32(&pool.stopped, 1)
	pool.maybeStopIdleWorker()
}

// See: https://github.com/alitto/pond/issues/33
func TestPurgeDuringSubmit(t *testing.T) {

	pool := New(1, 1)

	var doneCount int32

	// Submit a task to ensure at least 1 worker is started
	pool.SubmitAndWait(func() {
		atomic.AddInt32(&doneCount, 1)
	})

	time.Sleep(10 * time.Millisecond)

	assertEqual(t, 1, pool.IdleWorkers())

	// Stop an idle worker right before submitting another task
	pool.maybeStopIdleWorker()
	pool.Submit(func() {
		atomic.AddInt32(&doneCount, 1)
	})

	pool.StopAndWait()

	assertEqual(t, int32(2), atomic.LoadInt32(&doneCount))
	assertEqual(t, 0, pool.RunningWorkers())

}
