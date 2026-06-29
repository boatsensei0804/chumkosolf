package domain

// StatusCount คือจำนวนตามสถานะ (ใช้สรุปการเช็คชื่อ)
type StatusCount struct {
	Status string // "" = ยังไม่เช็ค
	Count  int
}

// Advisee คือนักเรียนในห้องที่ปรึกษาของครู พร้อมสถานะเช็คชื่อเข้าเรียนของวันนี้
type Advisee struct {
	StudentID     string
	StudentCode   string
	Prefix        string
	FirstName     string
	LastName      string
	Phone         string
	NationalIDEnc []byte
	GradeLevel    string
	RoomName      string
	TodayStatus   string // "" = ยังไม่เช็ค
}
