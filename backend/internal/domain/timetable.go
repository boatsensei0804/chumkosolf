package domain

import "time"

// TimetableSettings คือค่าตั้งตารางสอนรายเทอม (ขนาดตาราง)
type TimetableSettings struct {
	ID            string
	SchoolID      string
	SemesterID    string
	DaysPerWeek   int
	PeriodsPerDay int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PeriodDefinition คือนิยามคาบเรียน (เวลาเริ่ม-จบ; เก็บเวลาเป็น "HH:MM")
type PeriodDefinition struct {
	ID         string
	SchoolID   string
	SemesterID string
	PeriodNo   int
	Label      string
	StartTime  string
	EndTime    string
	IsBreak    bool
}

// TeacherBrief คือข้อมูลย่อของบุคลากร (ใช้แสดงรายชื่อครูว่าง)
type TeacherBrief struct {
	ID        string
	Prefix    string
	FirstName string
	LastName  string
}

// TeacherPeriod คือคู่ (ครู, คาบ) ที่ติดสอน — ใช้คำนวณครูที่ว่าง
type TeacherPeriod struct {
	PersonnelID string
	PeriodNo    int
}

// NewPeriodDefinition payload บันทึกนิยามคาบ
type NewPeriodDefinition struct {
	PeriodNo  int
	Label     string
	StartTime string
	EndTime   string
	IsBreak   bool
}

// TimetableSlot คือช่องตารางสอน (ห้อง×วัน×คาบ → มอบหมายการสอน)
type TimetableSlot struct {
	ID                   string
	SchoolID             string
	SemesterID           string
	ClassID              string
	DayOfWeek            int
	PeriodNo             int
	TeachingAssignmentID string
	// joined สำหรับแสดงผล
	SubjectCode      string
	SubjectName      string
	TeacherPrefix    string
	TeacherFirstName string
	TeacherLastName  string
}

// NewTimetableSlot payload ตั้งค่าช่องตารางสอน
type NewTimetableSlot struct {
	DayOfWeek            int
	PeriodNo             int
	TeachingAssignmentID string
}

// TeacherCheckinSlot คือคาบที่ครู (user) สอน — ใช้แสดงกริดเช็คชื่อรายวิชา
type TeacherCheckinSlot struct {
	SlotID      string
	DayOfWeek   int
	PeriodNo    int
	SubjectCode string
	SubjectName string
	GradeLevel  string
	RoomName    string
}

// SlotDate คือคู่ (คาบ, วันที่) ที่มีการเช็คชื่อแล้ว
type SlotDate struct {
	SlotID string
	Date   time.Time
}
