package domain

import "time"

// สถานะการเช็คชื่อเข้าเรียน (ตรงกับ CHECK constraint และ frontend zod)
const (
	AttendancePresent       = "present"
	AttendanceAbsent        = "absent"
	AttendanceLate          = "late"
	AttendanceSickLeave     = "sick_leave"
	AttendancePersonalLeave = "personal_leave"
)

// ValidAttendanceStatus ตรวจว่าสถานะอยู่ในชุดที่อนุญาต
func ValidAttendanceStatus(s string) bool {
	switch s {
	case AttendancePresent, AttendanceAbsent, AttendanceLate, AttendanceSickLeave, AttendancePersonalLeave:
		return true
	default:
		return false
	}
}

// AttendanceRosterEntry คือ 1 แถวของรายชื่อนักเรียนในห้อง พร้อมผลเช็คชื่อของวันนั้น
// Status = "" หมายถึงยังไม่ได้เช็คในวันนั้น
type AttendanceRosterEntry struct {
	StudentID    string
	StudentNo    *int
	StudentCode  string
	Prefix       string
	FirstName    string
	LastName     string
	AttendanceID string
	Status       string
	Note         string
	// DailyStatus = สถานะเช็คชื่อเข้าเรียนรายวันของวันนั้น (โชว์ "มาสาย" ในเช็คชื่อรายวิชา); "" = ยังไม่เช็ค
	DailyStatus string
}

// AttendanceMark คือผลเช็คชื่อของนักเรียน 1 คนที่จะบันทึก
type AttendanceMark struct {
	StudentID string
	Status    string
	Note      string
}

// Attendance คือผลเช็คชื่อรายวันของนักเรียน (รายเทอม)
type Attendance struct {
	ID         string
	SchoolID   string
	SemesterID string
	ClassID    string
	StudentID  string
	Date       time.Time
	Status     string
	Note       string
	CheckedBy  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
