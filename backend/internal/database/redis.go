package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient สร้าง client ไปยัง Redis และ ping ตรวจสอบ
// คืน nil โดยไม่ error ถ้า redisURL ว่าง (Redis เป็น optional ในบาง env)
func NewRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	if redisURL == "" {
		return nil, nil
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("database: parse redis url: %w", err)
	}

	client := redis.NewClient(opt)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("database: ping redis: %w", err)
	}

	return client, nil
}
