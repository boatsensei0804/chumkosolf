package domain

import "time"

// ตำแหน่งบริหาร
const (
	PositionDirector       = "director"
	PositionDeputyDirector = "deputy_director"
)

// AdminPosition คือตำแหน่งบริหารของบุคลากร (ผอ./รอง ผอ.)
type AdminPosition struct {
	ID          string
	SchoolID    string
	PersonnelID string
	Position    string
	IsActive    bool
	AppointedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewAdminPosition payload สร้างตำแหน่งบริหาร
type NewAdminPosition struct {
	Position    string
	IsActive    bool
	AppointedAt *time.Time
}

// AcademicStanding คือวิทยฐานะ (เก็บเป็นประวัติหลายรายการ; is_current ได้แค่ 1 ต่อคน)
type AcademicStanding struct {
	ID            string
	SchoolID      string
	PersonnelID   string
	Standing      string
	EffectiveDate *time.Time
	IsCurrent     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewAcademicStanding payload สร้างวิทยฐานะ
type NewAcademicStanding struct {
	Standing      string
	EffectiveDate *time.Time
	IsCurrent     bool
}

// UpdateAcademicStanding payload แก้ไขวิทยฐานะ
type UpdateAcademicStanding struct {
	Standing      string
	EffectiveDate *time.Time
	IsCurrent     bool
}
