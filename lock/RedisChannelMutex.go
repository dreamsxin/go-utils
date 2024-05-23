package lock

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// 默认锁超时时间
const lockTime = 5 * time.Second

type Option func(lock *RedisChannelMutex)

// WithTimeout 设置锁过期时间
func WithTimeout(timeout time.Duration) Option {
	return func(lock *RedisChannelMutex) {
		lock.lockTime = timeout
	}
}

// WithAutoRenew 是否开启自动续期
func WithAutoRenew() Option {
	return func(lock *RedisChannelMutex) {
		lock.isAutoRenew = true
	}
}

// WithToken 设置锁的Token
func WithToken(token string) Option {
	return func(lock *RedisChannelMutex) {
		lock.token = token
	}
}

type RedisChannelMutex struct {
	ctx             context.Context
	db              *redis.Client
	lockKey         string
	token           string
	lockPath        string
	channelPath     string
	ch              <-chan *redis.Message
	lockTime        time.Duration
	isAutoRenew     bool
	autoRenewCtx    context.Context
	autoRenewCancel context.CancelFunc
}

func NewRedisChannelMutex(ctx context.Context, db *redis.Client, lockKey string, options ...Option) (*RedisChannelMutex, error) {
	_, err := db.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	lock := &RedisChannelMutex{
		ctx:      ctx,
		db:       db,
		lockKey:  lockKey,
		lockTime: lockTime,
	}

	for _, f := range options {
		f(lock)
	}

	if lock.token == "" {
		lock.token = fmt.Sprintf("token:%d", time.Now().UnixNano())
	}

	lock.lockPath = "RedisMutex:key:" + lock.lockKey
	lock.channelPath = "RedisMutex:Channel:" + lockKey
	ps := db.Subscribe(ctx, lock.channelPath)
	lock.ch = ps.Channel()

	return lock, nil
}

func (m *RedisChannelMutex) Lock() {
	for {
		created, err := m.db.SetNX(m.ctx, m.lockPath, m.token, m.lockTime).Result()
		if err != nil {
			panic(err)
		}
		if created {
			if m.autoRenewCancel != nil {
				m.autoRenewCancel()
			}
			if m.isAutoRenew {
				m.autoRenewCtx, m.autoRenewCancel = context.WithCancel(m.ctx)
				go m.autoRenew()
			}
			break
		}
		<-m.ch
	}
}

func (m *RedisChannelMutex) TryLock() bool {

	created, err := m.db.SetNX(m.ctx, m.lockPath, m.token, m.lockTime).Result()
	if err != nil {
		panic(err)
	}
	if created {
		if m.autoRenewCancel != nil {
			m.autoRenewCancel()
		}
		if m.isAutoRenew {
			m.autoRenewCtx, m.autoRenewCancel = context.WithCancel(m.ctx)
			go m.autoRenew()
		}
	}
	return created
}

func (m *RedisChannelMutex) Unlock() {
	if m.autoRenewCancel != nil {
		m.autoRenewCancel()
	}
	m.db.Del(m.ctx, m.lockPath)
	m.db.Publish(m.ctx, m.channelPath, "unlock")
}

func (m *RedisChannelMutex) Renew() (bool, error) {
	return m.db.Expire(m.ctx, m.lockPath, m.lockTime).Result()
	//return m.db.ExpireNX(m.ctx, m.lockPath, m.lockTime).Result()
}

func (m *RedisChannelMutex) autoRenew() {
	ticker := time.NewTicker(m.lockTime / 2)
	defer ticker.Stop()

	for {
		select {
		case <-m.autoRenewCtx.Done():
			m.autoRenewCancel = nil
			log.Println("autoRenew cancel")
			return
		case <-ticker.C:
			ret, err := m.Renew()
			if err != nil || !ret {
				m.autoRenewCancel = nil
				log.Println("autoRenew failed:", err)
				return
			}

			log.Println("autoRenew success")
		}
	}
}
