package domain

import "time"

// AcademicYear คือปีการศึกษา (พ.ศ.)
type AcademicYear struct {
	ID        string
	SchoolID  string
	Year      int
	IsCurrent bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Semester คือภาคเรียน (1/2) ของปีการศึกษา
type Semester struct {
	ID             string
	SchoolID       string
	AcademicYearID string
	Year           int // ปีการศึกษาของเทอมนี้ (join)
	Term           int
	StartDate      *time.Time
	EndDate        *time.Time
	IsCurrent      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewSemester payload สร้างภาคเรียน
type NewSemester struct {
	AcademicYearID string
	Term           int
	StartDate      *time.Time
	EndDate        *time.Time
}
