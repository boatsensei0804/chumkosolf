// Package domain เก็บ entity และ business rule ที่ไม่พึ่ง framework
package domain

// Role คือบทบาทของผู้ใช้ (ตรงกับ CHECK constraint ในตาราง users)
const (
	RoleSuperAdmin = "super_admin"
	RoleTeacher    = "teacher"
	RoleExecutive  = "executive"
	RoleStudent    = "student"
)

// User คือบัญชีผู้ใช้ที่ใช้ login (ครู/ผู้บริหาร/นักเรียน ใช้ตารางเดียวกัน)
// PasswordHash เป็น bcrypt — ห้ามส่งออกนอก service ชั้นใน
type User struct {
	ID            string
	SchoolID      string
	Username      string
	PasswordHash  string
	Role          string
	IsSchoolAdmin bool
	IsActive      bool
}

// WorkGroupMembership คือกลุ่มงานที่ผู้ใช้สังกัด พร้อมสถานะ admin ของกลุ่ม
type WorkGroupMembership struct {
	WorkGroupID  string `json:"work_group_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	IsGroupAdmin bool   `json:"is_group_admin"`
}
