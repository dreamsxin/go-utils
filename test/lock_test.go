package test

import (
	"time"
	"github.com/dreamsxin/goutils/lock"
	"testing"
)

var ml lock.Multiplelock

//go test -v -count=1 --run TestHello .
func TestHello(t *testing.T) {
	t.Log("hello world")

}

//终止当前测试用例
func TestFailNow(t *testing.T) {
    t.Log("before fail")
    t.FailNow()
    t.Log("after fail")
}

//go test -v -count=1 --run TestFail .
func TestFail(t *testing.T) {
    t.Log("before fail")
    t.Fail()
    t.Log("after fail")
}

//go test -v -count=1 --bench BenchmarkAdd .
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

//go test -v -count=1 --run TestMultiplelock .
func TestMultiplelock(t *testing.T) {
	go func() {
		ml.Lock(1)
		t.Log("lock success")
		time.Sleep(time.Second)
		ml.Unlock(1)
	}()
	go func() {
		ml.Lock(1)
		t.Log("lock success")
		time.Sleep(time.Second)
		ml.Unlock(1)
	}()

}