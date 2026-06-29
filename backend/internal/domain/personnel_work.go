package domain

import "time"

// ประเภทไฟล์แนบของผลงานครู
const (
	WorkFileImage       = "image"
	WorkFileDocument    = "document"
	WorkFileCertificate = "certificate"
)

// ValidWorkFileType ตรวจว่าเป็นประเภทไฟล์ที่อนุญาต
func ValidWorkFileType(t string) bool {
	switch t {
	case WorkFileImage, WorkFileDocument, WorkFileCertificate:
		return true
	default:
		return false
	}
}

// PersonnelWork คือผลงานครู — ผูกเทอม (semester_id) ตาม CLAUDE.md ข้อ 4.3
type PersonnelWork struct {
	ID          string
	SchoolID    string
	SemesterID  string
	PersonnelID string
	Title       string
	Description string
	WorkDate    *time.Time
	FileCount   int // จำนวนไฟล์แนบ (เติมตอน list)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewPersonnelWork payload สร้างผลงาน
type NewPersonnelWork struct {
	Title       string
	Description string
	WorkDate    *time.Time
}

// UpdatePersonnelWork payload แก้ไขผลงาน
type UpdatePersonnelWork struct {
	Title       string
	Description string
	WorkDate    *time.Time
}

// PersonnelWorkFile คือไฟล์แนบของผลงาน — เข้าถึงผ่าน signed URL เท่านั้น (PDPA)
type PersonnelWorkFile struct {
	ID              string
	SchoolID        string
	PersonnelWorkID string
	FileType        string
	StoragePath     string
	OriginalName    string
	ContentType     string
	SizeBytes       int64
	CreatedAt       time.Time
}

// NewPersonnelWorkFile payload บันทึก metadata ไฟล์แนบ (หลังอัปโหลดขึ้น storage แล้ว)
type NewPersonnelWorkFile struct {
	FileType     string
	StoragePath  string
	OriginalName string
	ContentType  string
	SizeBytes    int64
}
