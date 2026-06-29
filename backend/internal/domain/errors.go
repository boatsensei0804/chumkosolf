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

// Classes / advisors / enrollments errors
var (
	ErrNoActiveSemester = &Error{
		Status:  http.StatusBadRequest,
		Code:    "NO_ACTIVE_SEMESTER",
		Message: "ยังไม่ได้กำหนดเทอมปัจจุบัน กรุณาเข้าสู่ระบบใหม่หรือตั้งค่าเทอม",
	}
	ErrClassNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "CLASS_NOT_FOUND",
		Message: "ไม่พบห้องเรียน",
	}
	ErrDuplicateClass = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_CLASS",
		Message: "มีห้องเรียนระดับชั้น/ห้องนี้ในเทอมนี้แล้ว",
	}
	ErrAdvisorNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "ADVISOR_NOT_FOUND",
		Message: "ไม่พบครูที่ปรึกษานี้ในห้อง",
	}
	ErrEnrollmentNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "ENROLLMENT_NOT_FOUND",
		Message: "ไม่พบนักเรียนในห้องนี้",
	}
)

// Personnel works (ผลงานครู) + ไฟล์แนบ errors
var (
	ErrWorkNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "WORK_NOT_FOUND",
		Message: "ไม่พบผลงาน",
	}
	ErrWorkFileNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "WORK_FILE_NOT_FOUND",
		Message: "ไม่พบไฟล์แนบ",
	}
	ErrInvalidFileType = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_FILE_TYPE",
		Message: "ประเภทไฟล์ต้องเป็น รูปภาพ/เอกสาร/เกียรติบัตร",
	}
	ErrFileTooLarge = &Error{
		Status:  http.StatusRequestEntityTooLarge,
		Code:    "FILE_TOO_LARGE",
		Message: "ไฟล์มีขนาดใหญ่เกินกำหนด (สูงสุด 10 MB)",
	}
	ErrFileRequired = &Error{
		Status:  http.StatusBadRequest,
		Code:    "FILE_REQUIRED",
		Message: "กรุณาเลือกไฟล์ที่จะอัปโหลด",
	}
	ErrStorageUnavailable = &Error{
		Status:  http.StatusServiceUnavailable,
		Code:    "STORAGE_UNAVAILABLE",
		Message: "ระบบจัดเก็บไฟล์ยังไม่พร้อมใช้งาน",
	}
	ErrInvalidImageType = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_IMAGE_TYPE",
		Message: "รูปภาพต้องเป็นไฟล์ JPG, PNG หรือ WEBP",
	}
	ErrStudentPhotoNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "STUDENT_PHOTO_NOT_FOUND",
		Message: "ไม่พบรูปนักเรียนนี้",
	}
	ErrPhotoLimitReached = &Error{
		Status:  http.StatusBadRequest,
		Code:    "PHOTO_LIMIT_REACHED",
		Message: "เก็บรูปได้สูงสุด 10 รูปต่อนักเรียน",
	}
	ErrFaceServiceUnavailable = &Error{
		Status:  http.StatusServiceUnavailable,
		Code:    "FACE_SERVICE_UNAVAILABLE",
		Message: "ระบบสแกนหน้ายังไม่พร้อมใช้งาน",
	}
	ErrNoFaceDetected = &Error{
		Status:  http.StatusUnprocessableEntity,
		Code:    "NO_FACE_DETECTED",
		Message: "ไม่พบใบหน้าในรูป",
	}
)

// เช็คชื่อเข้าเรียน (attendances) + คะแนนความประพฤติ (behavior_records) errors
var (
	ErrInvalidAttendanceStatus = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_ATTENDANCE_STATUS",
		Message: "สถานะการเช็คชื่อไม่ถูกต้อง",
	}
	ErrInvalidDate = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_DATE",
		Message: "กรุณาระบุวันที่ให้ถูกต้อง",
	}
	ErrStudentNotInClass = &Error{
		Status:  http.StatusBadRequest,
		Code:    "STUDENT_NOT_IN_CLASS",
		Message: "มีนักเรียนที่ไม่ได้อยู่ในห้องนี้",
	}
	ErrInvalidPoints = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_POINTS",
		Message: "คะแนนต้องไม่เป็นศูนย์ (ระบุเป็นบวกเพื่อเพิ่ม หรือลบเพื่อหัก)",
	}
	ErrReasonRequired = &Error{
		Status:  http.StatusBadRequest,
		Code:    "REASON_REQUIRED",
		Message: "กรุณาระบุเหตุผล",
	}
	ErrBehaviorNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "BEHAVIOR_NOT_FOUND",
		Message: "ไม่พบรายการคะแนนความประพฤติ",
	}
)

// ตารางสอน (Phase 3) errors
var (
	ErrSubjectNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "SUBJECT_NOT_FOUND",
		Message: "ไม่พบรายวิชา",
	}
	ErrDuplicateSubjectCode = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_SUBJECT_CODE",
		Message: "รหัสวิชานี้มีอยู่ในระบบแล้ว",
	}
	ErrTeachingAssignmentNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "TEACHING_ASSIGNMENT_NOT_FOUND",
		Message: "ไม่พบการมอบหมายการสอน",
	}
	ErrDuplicateTeachingAssignment = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_TEACHING_ASSIGNMENT",
		Message: "มีการมอบหมายการสอนนี้อยู่แล้ว (ครู/วิชา/ห้องซ้ำ)",
	}
	ErrTimetableSettingsNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "TIMETABLE_SETTINGS_NOT_FOUND",
		Message: "ยังไม่ได้ตั้งค่าตารางสอนของเทอมนี้",
	}
	ErrPeriodNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "PERIOD_NOT_FOUND",
		Message: "ไม่พบคาบเรียน",
	}
	ErrTimetableSlotNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "TIMETABLE_SLOT_NOT_FOUND",
		Message: "ไม่พบช่องตารางสอนนี้",
	}
	ErrInvalidTimetableSlot = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_TIMETABLE_SLOT",
		Message: "วัน/คาบของตารางสอนไม่ถูกต้อง",
	}
	ErrTeacherTimeConflict = &Error{
		Status:  http.StatusConflict,
		Code:    "TEACHER_TIME_CONFLICT",
		Message: "ครูมีคาบสอนห้องอื่นในวันและคาบเดียวกันนี้แล้ว",
	}
	ErrSubjectAttendanceNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "SUBJECT_ATTENDANCE_NOT_FOUND",
		Message: "ไม่พบการเช็คชื่อรายวิชา",
	}
)

// ปีการศึกษา/ภาคเรียน (จัดการเทอม) errors
var (
	ErrYearNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "YEAR_NOT_FOUND",
		Message: "ไม่พบปีการศึกษา",
	}
	ErrDuplicateYear = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_YEAR",
		Message: "ปีการศึกษานี้มีอยู่แล้ว",
	}
	ErrInvalidYear = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_YEAR",
		Message: "ปีการศึกษาไม่ถูกต้อง (ระบุเป็น พ.ศ.)",
	}
	ErrSemesterNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "SEMESTER_NOT_FOUND",
		Message: "ไม่พบภาคเรียน",
	}
	ErrDuplicateSemester = &Error{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_SEMESTER",
		Message: "ภาคเรียนนี้ของปีการศึกษานี้มีอยู่แล้ว",
	}
	ErrInvalidTerm = &Error{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_TERM",
		Message: "ภาคเรียนต้องเป็น 1 หรือ 2",
	}
	ErrSchoolNotFound = &Error{
		Status:  http.StatusNotFound,
		Code:    "SCHOOL_NOT_FOUND",
		Message: "ไม่พบข้อมูลโรงเรียน",
	}
)
