package test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dreamsxin/go-utils/lock"
	"github.com/redis/go-redis/v9"
)

var ml lock.MultipleLock
var rl *lock.RedisChannelMutex

func init() {
	ml = lock.NewMultipleLock()
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456", // no password set
		DB:       0,        // use default DB
	})
	rl, _ = lock.NewRedisChannelMutex(ctx, rdb, "lock.test", time.Duration(10*time.Second))
}

// go test -v -count=1 --run TestHello .
func TestHello(t *testing.T) {
	t.Log("hello world")

}

// 终止当前测试用例
func TestFailNow(t *testing.T) {
	t.Log("before fail")
	t.FailNow()
	t.Log("after fail")
}

// go test -v -count=1 --run TestFail .
func TestFail(t *testing.T) {
	t.Log("before fail")
	t.Fail()
	t.Log("after fail")
}

// go test -v -count=1 --bench BenchmarkAdd .
func BenchmarkAdd(b *testing.B) {
	// 重置计时器
	b.ResetTimer()
	// 停止计时器
	b.StopTimer()
	var n int
	for i := 0; i < b.N; i++ {
		if n > 0 {
			// 开始计时器
			b.StartTimer()
		}
		n++
	}
}

// go test -v -count=1 --run TestMultiplelock .
func TestMultiplelock(t *testing.T) {
	go func() {
		ml.Lock(1)
		t.Log("lock success")
		time.Sleep(2 * time.Second)
		ml.Unlock(1)
	}()

	ml.Lock(1)
	t.Log("lock success")
	time.Sleep(1 * time.Second)
	ml.Unlock(1)

	ml.Wait(1)
}

// go test -v -count=1 --run TestRedislock .
func TestRedislock(t *testing.T) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		t.Log("start lock2")
		rl.Lock()
		t.Log("lock2 success")
		time.Sleep(2 * time.Second)
		rl.Unlock()
		t.Log("lock2 unlock")
	}()

	rl.Lock()
	t.Log("lock success")
	time.Sleep(1 * time.Second)
	rl.Unlock()
	t.Log("lock unlock")
	waitGroup.Wait()
}
