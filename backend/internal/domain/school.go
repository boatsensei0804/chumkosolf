package domain

// School คือข้อมูลโรงเรียน (สำหรับหน้าตั้งค่าระบบ)
type School struct {
	ID           string
	Name         string
	Code         string
	Address      Address // ที่อยู่แยกฟิลด์ (มาตรฐานไทย)
	Phone        string
	Email        string
	Website      string
	DirectorName string
	IsActive     bool
	// AttendanceLateAfter = เวลาตัดมา/สาย "HH:MM" (โซนไทย) — ตั้งค่าได้ต่อโรงเรียน
	AttendanceLateAfter string
	// AttendanceLatePenalty = คะแนนความประพฤติที่หักเมื่อมาสาย (0 = ไม่หัก)
	AttendanceLatePenalty int
}

// UpdateSchool payload แก้ไขข้อมูลโรงเรียน (code เป็นตัวระบุ ไม่ให้แก้)
type UpdateSchool struct {
	Name                  string
	Address               Address
	Phone                 string
	Email                 string
	Website               string
	DirectorName          string
	AttendanceLateAfter   string
	AttendanceLatePenalty int
}
