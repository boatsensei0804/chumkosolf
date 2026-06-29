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

	// Storage (MinIO/S3) สำหรับไฟล์แนบส่วนบุคคล (ผลงานครู) — เข้าถึงผ่าน signed URL
	// ถ้า Endpoint ว่าง = ไม่เปิดใช้งาน storage (endpoint ไฟล์จะตอบ error)
	Storage StorageConfig

	// FaceServiceURL = base URL ของ face-svc (สแกนหน้า) — ว่าง = ปิดฟีเจอร์สแกนหน้า
	FaceServiceURL string

	// AttendanceLateAfter = เวลาตัด "มา/สาย" รูปแบบ HH:MM (โซนเวลาไทย); หลังเวลานี้ = สาย
	AttendanceLateAfter string

	// FaceLiveness = เปิดการตรวจ liveness (ต้องขยับหน้า) ตอนสแกน; ปิด = สแกนเฟรมเดียวทันที
	FaceLiveness bool
}

// StorageConfig คือค่าตั้งของ object storage (S3-compatible)
type StorageConfig struct {
	Endpoint string // endpoint ที่ backend ใช้ต่อ (เช่น minio:9000)
	// PublicEndpoint ใช้สร้าง signed URL ให้เบราว์เซอร์ (host ที่ภายนอก resolve ได้)
	// ถ้าว่าง ใช้ Endpoint แทน
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
	// Region ตั้งให้ชัดเพื่อให้ minio-go ไม่ต้องเรียก GetBucketLocation ตอน presign
	// (มิฉะนั้น presign client จะ dial public endpoint จากใน container ไม่ได้)
	Region string
}

// Enabled คืน true เมื่อ storage ถูกตั้งค่าครบพอใช้งาน
func (s StorageConfig) Enabled() bool {
	return s.Endpoint != "" && s.AccessKey != "" && s.SecretKey != "" && s.Bucket != ""
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
		Storage: StorageConfig{
			Endpoint:       os.Getenv("STORAGE_ENDPOINT"),
			PublicEndpoint: os.Getenv("STORAGE_PUBLIC_ENDPOINT"),
			AccessKey:      os.Getenv("STORAGE_ACCESS_KEY"),
			SecretKey:      os.Getenv("STORAGE_SECRET_KEY"),
			Bucket:         getEnv("STORAGE_BUCKET", "chumko-files"),
			UseSSL:         os.Getenv("STORAGE_USE_SSL") == "true",
			Region:         getEnv("STORAGE_REGION", "us-east-1"),
		},
		FaceServiceURL:      os.Getenv("FACE_SERVICE_URL"),
		AttendanceLateAfter: getEnv("ATTENDANCE_LATE_AFTER", "08:00"),
		FaceLiveness:        os.Getenv("FACE_LIVENESS") == "on",
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
