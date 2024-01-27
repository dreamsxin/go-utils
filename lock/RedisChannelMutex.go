package lock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisChannelMutex struct {
	ctx         context.Context
	db          *redis.Client
	LockPath    string
	ChannelPath string
	ch          <-chan *redis.Message
	LockTime    time.Duration
}

func NewRedisChannelMutex(ctx context.Context, db *redis.Client, LockKey string, lockTime time.Duration) (*RedisChannelMutex, error) {
	_, err := db.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	if lockTime < 0 {
		lockTime = time.Duration(0)
	}
	channelPath := "RedisMutex:Channel:" + LockKey
	ps := db.Subscribe(ctx, channelPath)
	return &RedisChannelMutex{
		ctx:         ctx,
		db:          db,
		LockPath:    "RedisMutex:" + LockKey,
		ChannelPath: channelPath,
		ch:          ps.Channel(),
		LockTime:    lockTime,
	}, err
}

func (m *RedisChannelMutex) Lock() {
	for {
		created, err := m.db.SetNX(m.ctx, m.LockPath, "lock", m.LockTime).Result()
		if err != nil {
			panic(err)
		}
		if created {
			break
		}
		<-m.ch
	}
}

func (m *RedisChannelMutex) TryLock() bool {
	for {
		created, err := m.db.SetNX(m.ctx, m.LockPath, "lock", m.LockTime).Result()
		if err != nil {
			panic(err)
		}
		return created
	}
}

func (m *RedisChannelMutex) Unlock() {
	m.db.Del(m.ctx, m.LockPath)
	m.db.Publish(m.ctx, m.ChannelPath, "unlock")
}
