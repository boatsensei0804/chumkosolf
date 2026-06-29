package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/chumko-platform/backend/internal/config"
)

// MinioStorage คือ implementation ของ Storage บน MinIO/S3
// แยก 2 client: opClient ใช้ทำ operation จริง (ต่อ endpoint ภายใน),
// presignClient ใช้สร้าง signed URL ที่ host สาธารณะให้เบราว์เซอร์เข้าถึงได้
// (การ presign เป็นการคำนวณ signature แบบ offline ไม่ต้องต่อ network)
type MinioStorage struct {
	opClient      *minio.Client
	presignClient *minio.Client
	bucket        string
}

// NewMinioStorage สร้าง client + ยืนยันว่ามี bucket (สร้างให้ถ้ายังไม่มี)
func NewMinioStorage(ctx context.Context, cfg config.StorageConfig) (*MinioStorage, error) {
	creds := credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, "")

	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	opClient, err := minio.New(cfg.Endpoint, &minio.Options{Creds: creds, Secure: cfg.UseSSL, Region: region})
	if err != nil {
		return nil, fmt.Errorf("storage: สร้าง minio client: %w", err)
	}

	publicEndpoint := cfg.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}
	// presign client ตั้ง Region ชัด เพื่อให้ presign แบบ offline (ไม่ dial public endpoint จากใน container)
	presignClient, err := minio.New(publicEndpoint, &minio.Options{Creds: creds, Secure: cfg.UseSSL, Region: region})
	if err != nil {
		return nil, fmt.Errorf("storage: สร้าง minio presign client: %w", err)
	}

	s := &MinioStorage{opClient: opClient, presignClient: presignClient, bucket: cfg.Bucket}
	if err := s.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *MinioStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.opClient.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("storage: ตรวจ bucket: %w", err)
	}
	if exists {
		return nil
	}
	if err := s.opClient.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("storage: สร้าง bucket: %w", err)
	}
	return nil
}

// Put อัปโหลดไฟล์
func (s *MinioStorage) Put(ctx context.Context, objectPath string, r io.Reader, size int64, contentType string) error {
	_, err := s.opClient.PutObject(ctx, s.bucket, objectPath, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("storage: อัปโหลดไฟล์: %w", err)
	}
	return nil
}

// Get อ่านไฟล์เป็น bytes (ใช้ภายใน backend — เช่น ดึงรูปไปส่งให้ face-svc)
func (s *MinioStorage) Get(ctx context.Context, objectPath string) ([]byte, error) {
	obj, err := s.opClient.GetObject(ctx, s.bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("storage: เปิดไฟล์: %w", err)
	}
	defer func() { _ = obj.Close() }()
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("storage: อ่านไฟล์: %w", err)
	}
	return data, nil
}

// PresignGet คืน signed URL สำหรับดาวน์โหลด (หมดอายุตาม expiry)
func (s *MinioStorage) PresignGet(ctx context.Context, objectPath, downloadName string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	if downloadName != "" {
		reqParams.Set("response-content-disposition", "attachment; filename=\""+downloadName+"\"")
	}
	u, err := s.presignClient.PresignedGetObject(ctx, s.bucket, objectPath, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("storage: สร้าง signed URL: %w", err)
	}
	return u.String(), nil
}

// Remove ลบไฟล์
func (s *MinioStorage) Remove(ctx context.Context, objectPath string) error {
	if err := s.opClient.RemoveObject(ctx, s.bucket, objectPath, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("storage: ลบไฟล์: %w", err)
	}
	return nil
}
