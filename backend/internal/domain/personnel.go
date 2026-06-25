package domain

import "time"

// Address คือที่อยู่แบบแยกฟิลด์ตามมาตรฐานไทย (ใช้ร่วมกับ students/guardians ภายหลังได้)
type Address struct {
	HouseNo     string
	Moo         string
	Road        string
	Subdistrict string
	District    string
	Province    string
	PostalCode  string
}

// PersonnelProfile คือข้อมูลโปรไฟล์ที่ไม่อ่อนไหวสูง (แก้ไขได้)
type PersonnelProfile struct {
	Prefix    string
	FirstName string
	LastName  string
	BirthDate *time.Time
	Phone     string
	Email     string
	Address   Address
}

// Personnel คือ entity ระดับ DB ของบุคลากร — national_id เก็บเป็น ciphertext เท่านั้น
// (service เป็นผู้ถอดรหัส/mask; ห้ามให้ repository รู้ plaintext)
type Personnel struct {
	ID                 string
	SchoolID           string
	UserID             string
	NationalIDEnc      []byte
	NationalIDHash     string
	CivilServantIDEnc  []byte // nil = ไม่มี
	CivilServantIDHash string
	Profile            PersonnelProfile
	PhotoPath          string

	// fields ที่ join มาจากตาราง users
	Username string
	Role     string
	IsActive bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPersonnel คือ payload สำหรับสร้างบุคลากร (พร้อมบัญชี user) — ค่าอ่อนไหวถูกเข้ารหัสมาแล้วโดย service
type NewPersonnel struct {
	// บัญชี user (สำหรับ login)
	Username      string
	PasswordHash  string
	Role          string
	IsSchoolAdmin bool

	// ข้อมูลอ่อนไหว (เข้ารหัส + hash มาแล้ว)
	NationalIDEnc      []byte
	NationalIDHash     string
	CivilServantIDEnc  []byte
	CivilServantIDHash string

	Profile   PersonnelProfile
	PhotoPath string
}

// UpdatePersonnel คือ payload แก้ไขบุคลากร — เปลี่ยนเลขบัตรเฉพาะเมื่อ ChangeNationalID/ChangeCivilID = true
type UpdatePersonnel struct {
	Profile PersonnelProfile

	ChangeNationalID bool
	NationalIDEnc    []byte
	NationalIDHash   string

	ChangeCivilID      bool
	CivilServantIDEnc  []byte
	CivilServantIDHash string
}

// AuditEntry คือรายการ audit สำหรับการเข้าถึงข้อมูลส่วนบุคคล (PDPA)
// Detail เก็บเฉพาะชื่อ field ที่แตะ ไม่เก็บค่าจริงของข้อมูลอ่อนไหว
type AuditEntry struct {
	SchoolID    string
	ActorUserID string
	Action      string // view, create, update, delete, export
	TargetType  string // personnel, student, ...
	TargetID    string
	Detail      map[string]any
	IPAddress   string
}

// Audit actions
const (
	AuditView   = "view"
	AuditCreate = "create"
	AuditUpdate = "update"
	AuditDelete = "delete"
	AuditExport = "export"
)
