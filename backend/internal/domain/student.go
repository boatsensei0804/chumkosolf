package domain

import "time"

// ความสัมพันธ์ผู้ปกครอง
const (
	RelationshipFather = "father"
	RelationshipMother = "mother"
	RelationshipOther  = "other"
)

// สถานะนักเรียน
const (
	StudentStatusStudying  = "studying"  // กำลังศึกษา
	StudentStatusResigned  = "resigned"  // ลาออก
	StudentStatusSuspended = "suspended" // แขวนลอย
)

// ValidStudentStatus ตรวจว่าสถานะนักเรียนอยู่ในชุดที่อนุญาต
func ValidStudentStatus(s string) bool {
	switch s {
	case StudentStatusStudying, StudentStatusResigned, StudentStatusSuspended:
		return true
	default:
		return false
	}
}

// PersonProfile ข้อมูลโปรไฟล์พื้นฐานที่ใช้ร่วมกัน (นักเรียน/ผู้ปกครอง) — ไม่มี email
type PersonProfile struct {
	Prefix    string
	FirstName string
	LastName  string
	BirthDate *time.Time
	Phone     string
	Address   Address
}

// Student — entity ระดับ DB; national_id เก็บ ciphertext เท่านั้น (PDPA: ข้อมูลเด็ก)
type Student struct {
	ID             string
	SchoolID       string
	NationalIDEnc  []byte
	NationalIDHash string
	StudentCode    string
	Status         string
	Profile        PersonProfile
	PhotoPath      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewStudent payload สร้างนักเรียน (ค่าอ่อนไหวเข้ารหัสมาแล้ว) — ไม่สร้าง user account (user_id ตั้งภายหลังได้)
type NewStudent struct {
	NationalIDEnc  []byte
	NationalIDHash string
	StudentCode    string
	Status         string
	Profile        PersonProfile
}

// UpdateStudent payload แก้ไข — เปลี่ยนเลขบัตรเฉพาะเมื่อ ChangeNationalID = true
type UpdateStudent struct {
	StudentCode      string
	Status           string
	Profile          PersonProfile
	ChangeNationalID bool
	NationalIDEnc    []byte
	NationalIDHash   string
}

// StudentPhoto — รูปนักเรียน 1 รูป (หลายรูปต่อคนเพื่อความแม่นยำของระบบสแกนหน้า)
type StudentPhoto struct {
	ID          string
	StudentID   string
	StoragePath string
	ContentType string
	SizeBytes   int64
	IsPrimary   bool
	CreatedAt   time.Time
}

// NewStudentPhoto payload เพิ่มรูปนักเรียน (อัปขึ้น storage แล้ว)
type NewStudentPhoto struct {
	StoragePath string
	ContentType string
	SizeBytes   int64
}

// FaceEmbedding คือ embedding ใบหน้าของรูปหนึ่งใบ (ผูกกับนักเรียน) สำหรับ match สแกนหน้า
type FaceEmbedding struct {
	StudentID string
	PhotoID   string
	Vector    []float32
}

// StudentPhotoRow คือแถว flat (นักเรียน × รูป) สำหรับสร้าง dataset สแกนหน้า — group ในชั้น service
type StudentPhotoRow struct {
	StudentID   string
	StudentCode string
	Prefix      string
	FirstName   string
	LastName    string
	GradeLevel  string
	RoomName    string
	PhotoID     string
	StoragePath string
	IsPrimary   bool
}

// Guardian — entity ผู้ปกครอง
type Guardian struct {
	ID             string
	SchoolID       string
	NationalIDEnc  []byte
	NationalIDHash string
	Profile        PersonProfile
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewGuardian payload สร้างผู้ปกครอง
type NewGuardian struct {
	NationalIDEnc  []byte
	NationalIDHash string
	Profile        PersonProfile
}

// UpdateGuardian payload แก้ไขผู้ปกครอง
type UpdateGuardian struct {
	Profile          PersonProfile
	ChangeNationalID bool
	NationalIDEnc    []byte
	NationalIDHash   string
}

// StudentGuardian — ความเชื่อมโยงนักเรียน↔ผู้ปกครอง (M:N) พร้อมข้อมูลผู้ปกครองที่ join มา
type StudentGuardian struct {
	ID           string
	GuardianID   string
	Relationship string
	IsPrimary    bool
	// join จาก guardians (ชื่อแสดงผล + เลขบัตร ciphertext สำหรับ mask)
	Prefix         string
	FirstName      string
	LastName       string
	Phone          string
	NationalIDEnc  []byte
}

// NewStudentGuardian payload เชื่อมผู้ปกครองเข้านักเรียน
type NewStudentGuardian struct {
	GuardianID   string
	Relationship string
	IsPrimary    bool
}
