package domain

import "time"

// DefaultBehaviorScore คือคะแนนความประพฤติตั้งต้นของนักเรียน (ธรรมเนียมไทย = 100)
// คะแนนปัจจุบัน = DefaultBehaviorScore + SUM(points) ของเทอมนั้น (CLAUDE.md ข้อ 4.8)
const DefaultBehaviorScore = 100

// BehaviorRecord คือรายการหัก/เพิ่มคะแนนความประพฤติ 1 ครั้ง (รายเทอม)
type BehaviorRecord struct {
	ID         string
	SchoolID   string
	SemesterID string
	StudentID  string
	Points     int // บวก = เพิ่ม, ลบ = หัก
	Reason     string
	RecordedBy string
	OccurredAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewBehaviorRecord payload สร้างรายการคะแนน
type NewBehaviorRecord struct {
	Points     int
	Reason     string
	OccurredAt *time.Time
}
