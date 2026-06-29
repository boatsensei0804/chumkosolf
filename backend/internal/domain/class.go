package domain

import "time"

// Class — ห้องที่ปรึกษา (รายเทอม)
type Class struct {
	ID         string
	SchoolID   string
	SemesterID string
	GradeLevel string
	RoomName   string
	// นับจากการ join (สำหรับรายการ)
	StudentCount int
	AdvisorCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type NewClass struct {
	GradeLevel string
	RoomName   string
}

type UpdateClass struct {
	GradeLevel string
	RoomName   string
}

// ClassAdvisor — ครูที่ปรึกษาของห้อง (M:N) พร้อมข้อมูลบุคลากรที่ join
type ClassAdvisor struct {
	ID          string
	PersonnelID string
	Prefix      string
	FirstName   string
	LastName    string
}

// ClassEnrollment — นักเรียนในห้อง (รายเทอม) พร้อมข้อมูลนักเรียนที่ join
type ClassEnrollment struct {
	ID          string
	StudentID   string
	StudentNo   *int
	StudentCode string
	Prefix      string
	FirstName   string
	LastName    string
}

// StudentClassBrief คือผลค้นหานักเรียน + ห้องที่สังกัด (ข้อมูลพื้นฐานเท่านั้น — PDPA)
type StudentClassBrief struct {
	StudentID   string
	StudentCode string
	Prefix      string
	FirstName   string
	LastName    string
	GradeLevel  string
	RoomName    string
}

type NewEnrollment struct {
	StudentID string
	StudentNo *int
}
