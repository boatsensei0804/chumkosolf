// Package config โหลดค่าตั้งจาก environment variable (ตาม 12-factor)
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config คือค่าตั้งทั้งหมดของ backend
type Config struct {
	AppEnv   string
	HTTPPort string

	DatabaseURL string
	RedisURL    string

	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration

	// EncryptionKey ใช้เข้ารหัสข้อมูลอ่อนไหว (เลขบัตรประชาชน ฯลฯ) แบบ AES-256-GCM
	// ต้องเป็น hex ขนาด 32 ไบต์ (64 ตัวอักษร)
	EncryptionKey string
}

// Load อ่าน config จาก env; คืน error ถ้าค่าจำเป็นขาด
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisURL:      os.Getenv("REDIS_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
	}

	cfg.AccessTokenExpiry = getEnvDuration("ACCESS_TOKEN_EXPIRY", 15*time.Minute)
	cfg.RefreshTokenExpiry = getEnvDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour)

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL จำเป็นต้องตั้งค่า")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET จำเป็นต้องตั้งค่า")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	// รองรับทั้งรูปแบบวินาที (ตัวเลขล้วน) และ duration string เช่น "15m"
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	return fallback
}
