package domain

import "net/http"

// Error คือ business error ที่พก code (machine-readable), ข้อความไทย, และ HTTP status
// handler แปลงเป็น response มาตรฐานได้โดยตรงผ่าน errors.As
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// Auth-related errors
// หมายเหตุด้านความปลอดภัย: school/username/password ที่ไม่ถูกต้องคืน ErrInvalidCredentials
// เหมือนกันหมด เพื่อกัน account/school enumeration
var (
	ErrInvalidCredentials = &Error{
		Status:  http.StatusUnauthorized,
		Code:    "INVALID_CREDENTIALS",
		Message: "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง",
	}
	ErrUserInactive = &Error{
		Status:  http.StatusForbidden,
		Code:    "USER_INACTIVE",
		Message: "บัญชีนี้ถูกระงับการใช้งาน",
	}
	ErrInvalidToken = &Error{
		Status:  http.StatusUnauthorized,
		Code:    "INVALID_TOKEN",
		Message: "โทเคนไม่ถูกต้องหรือหมดอายุ กรุณาเข้าสู่ระบบใหม่",
	}
	ErrUnauthorized = &Error{
		Status:  http.StatusUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: "กรุณาเข้าสู่ระบบก่อนใช้งาน",
	}
	ErrUserNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "USER_NOT_FOUND",
		Message: "ไม่พบบัญชีผู้ใช้",
	}
)

// Permission / resource errors
var (
	ErrForbidden = &Error{
		Status:  http.StatusForbidden,
		Code:    "FORBIDDEN",
		Message: "คุณไม่มีสิทธิ์เข้าถึงข้อมูลส่วนนี้",
	}
	ErrValidation = &Error{
		Status:  http.StatusBadRequest,
		Code:    "VALIDATION_ERROR",
		Message: "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน",
	}
	ErrPersonnelNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "PERSONNEL_NOT_FOUND",
		Message: "ไม่พบข้อมูลบุคลากร",
	}
	ErrDuplicateNationalID = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_NATIONAL_ID",
		Message: "เลขบัตรประชาชนนี้มีอยู่ในระบบแล้ว",
	}
	ErrDuplicateUsername = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_USERNAME",
		Message: "ชื่อผู้ใช้นี้ถูกใช้งานแล้ว",
	}
	ErrDuplicateDirector = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_DIRECTOR",
		Message: "มีผู้อำนวยการที่ดำรงตำแหน่งอยู่แล้ว (มีได้คนเดียว)",
	}
	ErrAdminPositionNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "ADMIN_POSITION_NOT_FOUND",
		Message: "ไม่พบตำแหน่งบริหาร",
	}
	ErrStandingNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "STANDING_NOT_FOUND",
		Message: "ไม่พบวิทยฐานะ",
	}
	ErrWorkGroupNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "WORK_GROUP_NOT_FOUND",
		Message: "ไม่พบกลุ่มงาน",
	}
	ErrWorkGroupAssignmentNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "ASSIGNMENT_NOT_FOUND",
		Message: "บุคลากรไม่ได้สังกัดกลุ่มงานนี้",
	}
)

// Students / guardians errors
var (
	ErrStudentNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "STUDENT_NOT_FOUND",
		Message: "ไม่พบข้อมูลนักเรียน",
	}
	ErrGuardianNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "GUARDIAN_NOT_FOUND",
		Message: "ไม่พบข้อมูลผู้ปกครอง",
	}
	ErrDuplicateStudentCode = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_STUDENT_CODE",
		Message: "รหัสนักเรียนนี้มีอยู่ในระบบแล้ว",
	}
	ErrDuplicateGuardianLink = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_GUARDIAN_LINK",
		Message: "ผู้ปกครองคนนี้ถูกเชื่อมกับนักเรียนแล้ว",
	}
	ErrGuardianLinkNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "GUARDIAN_LINK_NOT_FOUND",
		Message: "ไม่พบความเชื่อมโยงผู้ปกครองนี้",
	}
	ErrInvalidRelationship = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_RELATIONSHIP",
		Message: "ความสัมพันธ์ต้องเป็น บิดา/มารดา/อื่น ๆ",
	}
)
