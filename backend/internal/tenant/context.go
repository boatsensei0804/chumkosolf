// Package tenant จัดการ scope ของ tenant (school_id) และเทอม (semester_id)
// ผ่าน context.Context — เป็นแหล่งความจริงเดียวของ scope ห้ามรับค่าพวกนี้จาก client
package tenant

import "context"

type contextKey string

const (
	schoolIDKey      contextKey = "school_id"
	semesterIDKey    contextKey = "semester_id"
	userIDKey        contextKey = "user_id"
	roleKey          contextKey = "role"
	isSchoolAdminKey contextKey = "is_school_admin"
	ipAddressKey     contextKey = "ip_address"
)

// Identity คือข้อมูล scope/ตัวตนที่ middleware ดึงจาก JWT แล้วฝังลง context
type Identity struct {
	UserID        string
	SchoolID      string
	SemesterID    string
	Role          string
	IsSchoolAdmin bool
	// IPAddress ของผู้เรียก (สำหรับ audit log) — มาจาก request ไม่ใช่ JWT
	IPAddress string
}

// WithIdentity ฝัง identity ลง context
func WithIdentity(ctx context.Context, id Identity) context.Context {
	ctx = context.WithValue(ctx, userIDKey, id.UserID)
	ctx = context.WithValue(ctx, schoolIDKey, id.SchoolID)
	ctx = context.WithValue(ctx, semesterIDKey, id.SemesterID)
	ctx = context.WithValue(ctx, roleKey, id.Role)
	ctx = context.WithValue(ctx, isSchoolAdminKey, id.IsSchoolAdmin)
	ctx = context.WithValue(ctx, ipAddressKey, id.IPAddress)
	return ctx
}

// SchoolIDFromContext คืน school_id ที่ middleware ตั้งไว้ (string ว่างถ้าไม่มี)
func SchoolIDFromContext(ctx context.Context) string { return stringValue(ctx, schoolIDKey) }

// SemesterIDFromContext คืน semester_id ปัจจุบัน
func SemesterIDFromContext(ctx context.Context) string { return stringValue(ctx, semesterIDKey) }

// UserIDFromContext คืน user_id ของผู้เรียก
func UserIDFromContext(ctx context.Context) string { return stringValue(ctx, userIDKey) }

// RoleFromContext คืน role ของผู้เรียก
func RoleFromContext(ctx context.Context) string { return stringValue(ctx, roleKey) }

// IsSchoolAdminFromContext คืนว่าผู้เรียกเป็น school admin หรือไม่
func IsSchoolAdminFromContext(ctx context.Context) bool {
	if v, ok := ctx.Value(isSchoolAdminKey).(bool); ok {
		return v
	}
	return false
}

// IPAddressFromContext คืน IP ของผู้เรียก (สำหรับ audit)
func IPAddressFromContext(ctx context.Context) string { return stringValue(ctx, ipAddressKey) }

func stringValue(ctx context.Context, key contextKey) string {
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}
