// 包 cache 实现基于 BadgerDB 的带 TTL 的缓存功能
package badger

import (
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// 定义模块错误
var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrInvalidDataType = errors.New("invalid data type")
)

// Cache 结构体封装缓存实例
type Cache struct {
	db *badger.DB
}

// Config 配置参数
type Config struct {
	InMemory   bool          // 是否使用内存模式
	Dir        string        // 数据存储目录（磁盘模式需要）
	GCInterval time.Duration // 垃圾回收间隔
	Logger     badger.Logger // 自定义日志
}

// NewCache 创建新的缓存实例
func NewCache(cfg Config) (*Cache, error) {
	opts := badger.DefaultOptions(cfg.Dir)
	opts = opts.WithInMemory(cfg.InMemory)
	opts = opts.WithLogger(cfg.Logger)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	// 启动定期 GC 协程
	if cfg.GCInterval > 0 {
		go func() {
			ticker := time.NewTicker(cfg.GCInterval)
			defer ticker.Stop()
			for range ticker.C {
				for db.RunValueLogGC(0.5) == nil {
				}
			}
		}()
	}

	return &Cache{db: db}, nil
}

// Set 设置缓存值，ttl=0表示永不过期
func (c *Cache) Set(key string, value []byte, ttl time.Duration) error {
	return c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), value)
		if ttl > 0 {
			e = e.WithTTL(ttl)
		}
		return txn.SetEntry(e)
	})
}

// Get 获取缓存值
func (c *Cache) Get(key string) ([]byte, error) {
	var valCopy []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		valCopy = append([]byte{}, val...)
		return nil
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, ErrKeyNotFound
	}
	return valCopy, err
}

// Delete 删除指定键
func (c *Cache) Delete(key string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Exists 检查键是否存在
func (c *Cache) Exists(key string) (bool, error) {
	err := c.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})

	if err == nil {
		return true, nil
	}
	if errors.Is(err, badger.ErrKeyNotFound) {
		return false, nil
	}
	return false, err
}

// Clear 清空所有缓存数据（谨慎使用）
func (c *Cache) Clear() error {
	return c.db.DropAll()
}

// Close 关闭数据库连接
func (c *Cache) Close() error {
	return c.db.Close()
}
