package lock

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

type RedisMutex struct {
	ctx         context.Context
	db          *redis.Client
	LockPath    string
	ChannelPath string
	ch          <-chan *redis.Message
	LockTime    time.Duration
}

func NewRedisMutex(ctx context.Context, db *redis.Client, lockName string, lockTime time.Duration) (*RedisMutex, error) {
	_, err := db.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	if lockTime < 0 {
		lockTime = time.Duration(0)
	}
	channelPath := "RedisMutex:Channel:" + lockName
	ps := db.Subscribe(ctx, channelPath)
	return &RedisMutex{
		ctx:         ctx,
		db:          db,
		LockPath:    "RedisMutex:EXIST:" + lockName,
		ChannelPath: channelPath,
		ch:          ps.Channel(),
		LockTime:    lockTime,
	}, err
}

func (m *RedisMutex) Lock() {
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

func (m *RedisMutex) TryLock() bool {
	for {
		created, err := m.db.SetNX(m.ctx, m.LockPath, "lock", m.LockTime).Result()
		if err != nil {
			panic(err)
		}
		return created
	}
}

func (m *RedisMutex) Unlock() {
	m.db.Del(m.ctx, m.LockPath)
	m.db.Publish(m.ctx, m.ChannelPath, "unlock")
}
