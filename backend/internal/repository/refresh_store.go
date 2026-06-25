package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const refreshKeyPrefix = "refresh:"

// RedisRefreshStore เก็บ refresh token ใน Redis (มี TTL, เพิกถอนได้)
type RedisRefreshStore struct {
	rdb *redis.Client
}

// NewRedisRefreshStore สร้าง store; rdb เป็น nil ได้ (จะคืน error เมื่อถูกเรียกใช้)
func NewRedisRefreshStore(rdb *redis.Client) *RedisRefreshStore {
	return &RedisRefreshStore{rdb: rdb}
}

var errRefreshUnavailable = errors.New("repository: refresh store ไม่พร้อมใช้งาน (ไม่มี Redis)")

// Save เก็บ token → value พร้อม TTL
func (s *RedisRefreshStore) Save(ctx context.Context, token, value string, ttl time.Duration) error {
	if s.rdb == nil {
		return errRefreshUnavailable
	}
	if err := s.rdb.Set(ctx, refreshKeyPrefix+token, value, ttl).Err(); err != nil {
		return fmt.Errorf("repository: save refresh: %w", err)
	}
	return nil
}

// Lookup คืน value ของ token; คืน ("", nil) ถ้าไม่พบ/หมดอายุ
func (s *RedisRefreshStore) Lookup(ctx context.Context, token string) (string, error) {
	if s.rdb == nil {
		return "", errRefreshUnavailable
	}
	value, err := s.rdb.Get(ctx, refreshKeyPrefix+token).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("repository: lookup refresh: %w", err)
	}
	return value, nil
}

// Delete ลบ token (idempotent)
func (s *RedisRefreshStore) Delete(ctx context.Context, token string) error {
	if s.rdb == nil {
		return errRefreshUnavailable
	}
	if err := s.rdb.Del(ctx, refreshKeyPrefix+token).Err(); err != nil {
		return fmt.Errorf("repository: delete refresh: %w", err)
	}
	return nil
}
