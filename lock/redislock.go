package lock

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMutex struct {
	ctx             context.Context
	db              *redis.Client
	LockPath        string
	LockTime        time.Duration
	autoRenewCtx    context.Context
	autoRenewCancel context.CancelFunc
}

func NewRedisMutex(ctx context.Context, db *redis.Client, lockTime time.Duration) (*RedisMutex, error) {
	_, err := db.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	if lockTime < 0 {
		lockTime = time.Duration(0)
	}
	return &RedisMutex{
		ctx:      ctx,
		db:       db,
		LockPath: "RedisMutex:EXIST:",
		LockTime: lockTime,
	}, err
}

func (m *RedisMutex) TryLock(lockKey string) bool {

	created, err := m.db.SetNX(m.ctx, m.LockPath+lockKey, "lock", m.LockTime).Result()
	if err != nil {
		panic(err)
	}
	if created {
		if m.autoRenewCancel != nil {
			m.autoRenewCancel()
		}
	}
	return created
}

func (m *RedisMutex) Unlock(lockKey string) {
	if m.autoRenewCancel != nil {
		m.autoRenewCancel()
	}
	m.db.Del(m.ctx, m.LockPath+lockKey)
}

func (m *RedisMutex) Renew(lockKey string) (bool, error) {
	return m.db.ExpireNX(m.ctx, m.LockPath+lockKey, m.LockTime).Result()
}

func (m *RedisMutex) AutoRenew(lockKey string) {
	m.autoRenewCtx, m.autoRenewCancel = context.WithCancel(m.ctx)
	ticker := time.NewTicker(m.LockTime / 2)
	defer ticker.Stop()

	for {
		select {
		case <-m.autoRenewCtx.Done():
			m.autoRenewCancel = nil
			log.Println("autoRenew cancel")
			return
		case <-ticker.C:
			ret, err := m.Renew(lockKey)
			if err != nil || !ret {
				m.autoRenewCancel = nil
				log.Println("autoRenew failed:", err)
				return
			}
		}
	}
}
