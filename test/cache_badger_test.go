package test

import (
	"errors"
	"testing"
	"time"

	cache "github.com/dreamsxin/go-utils/cache/badger"
)

func TestCacheBadger(t *testing.T) {
	// 初始化配置
	cfg := cache.Config{
		InMemory:   true,
		GCInterval: 10 * time.Minute,
	}

	// 创建缓存实例
	c, err := cache.NewCache(cfg)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// 设置缓存
	err = c.Set("session:123", []byte("user_data"), 2*time.Second)
	if err != nil {
		t.Log("设置缓存失败:", err)
	}

	// 获取缓存
	val, err := c.Get("session:123")
	switch {
	case errors.Is(err, cache.ErrKeyNotFound):
		t.Log("键不存在")
	case err != nil:
		t.Log("获取错误:", err)
	default:
		t.Log("获取到值", val)
	}

	time.Sleep(2 * time.Second)

	// 检查存在性
	exists, err := c.Exists("session:123")
	if err != nil {
		t.Log("存在性检查错误:", err)
	}
	if exists {
		t.Log("键存在")
	} else {
		t.Log("键不存在")
	}

	// 删除键
	if err := c.Delete("session:123"); err != nil {
		t.Log("删除失败:", err)
	}
}
