// Package storage จัดการ object storage (S3-compatible / MinIO) สำหรับไฟล์แนบส่วนบุคคล
// ไฟล์อ่อนไหวตาม PDPA เข้าถึงผ่าน signed URL ที่หมดอายุเท่านั้น ไม่ใช่ public bucket
package storage

import (
	"context"
	"io"
	"time"
)

// Storage คือ contract ของชั้นเก็บไฟล์ (service ขึ้นกับ interface นี้เพื่อทดสอบง่าย)
type Storage interface {
	// Put อัปโหลดไฟล์ไปยัง objectPath (path ภายใน bucket)
	Put(ctx context.Context, objectPath string, r io.Reader, size int64, contentType string) error
	// Get อ่านไฟล์กลับมาเป็น bytes (ใช้ภายใน backend เช่น ส่งรูปให้ face-svc — ไม่ผ่าน signed URL)
	Get(ctx context.Context, objectPath string) ([]byte, error)
	// PresignGet คืน signed URL สำหรับดาวน์โหลดไฟล์ (หมดอายุตาม expiry)
	// downloadName ตั้งชื่อไฟล์ตอนดาวน์โหลด (Content-Disposition) — ค่าว่างได้
	PresignGet(ctx context.Context, objectPath, downloadName string, expiry time.Duration) (string, error)
	// Remove ลบไฟล์ออกจาก storage
	Remove(ctx context.Context, objectPath string) error
}
