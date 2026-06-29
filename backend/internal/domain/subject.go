package domain

import "time"

// Subject คือรายวิชา (ถาวร ไม่ผูกเทอม)
type Subject struct {
	ID          string
	SchoolID    string
	SubjectCode string
	Name        string
	Credit      *float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewSubject payload สร้างรายวิชา
type NewSubject struct {
	SubjectCode string
	Name        string
	Credit      *float64
}

// UpdateSubject payload แก้ไขรายวิชา
type UpdateSubject struct {
	SubjectCode string
	Name        string
	Credit      *float64
}

// TeachingAssignment คือการมอบหมายการสอน (รายเทอม) — ครู+วิชา+ห้อง
// ฟิลด์ชื่อ (เติมตอน list ผ่าน join) ใช้แสดงผล
type TeachingAssignment struct {
	ID          string
	SchoolID    string
	SemesterID  string
	PersonnelID string
	SubjectID   string
	ClassID     string
	// joined สำหรับแสดงผล
	TeacherPrefix    string
	TeacherFirstName string
	TeacherLastName  string
	SubjectCode      string
	SubjectName      string
	GradeLevel       string
	RoomName         string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewTeachingAssignment payload สร้างการมอบหมายการสอน
type NewTeachingAssignment struct {
	PersonnelID string
	SubjectID   string
	ClassID     string
}
