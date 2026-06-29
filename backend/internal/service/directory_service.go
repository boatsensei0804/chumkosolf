package service

import (
	"context"
	"strings"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// DirectoryService ให้ครู "ดู/ค้นหา" ห้องเรียนและนักเรียนแบบ read-only (ข้อมูลพื้นฐานเท่านั้น)
// PDPA: ไม่ส่งเลขบัตร/เบอร์/ที่อยู่ — แค่ ชื่อ/รหัส/ห้อง; ปิดไม่ให้ role student เข้าถึง
type DirectoryClassRepository interface {
	ListBySemester(ctx context.Context, schoolID, semesterID string) ([]domain.Class, error)
}

type DirectoryEnrollmentRepository interface {
	ListByClass(ctx context.Context, schoolID, classID string) ([]domain.ClassEnrollment, error)
	SearchByName(ctx context.Context, schoolID, semesterID, term string) ([]domain.StudentClassBrief, error)
}

type DirectoryService struct {
	classes     DirectoryClassRepository
	enrollments DirectoryEnrollmentRepository
}

func NewDirectoryService(classes DirectoryClassRepository, enrollments DirectoryEnrollmentRepository) *DirectoryService {
	return &DirectoryService{classes: classes, enrollments: enrollments}
}

// DTOs (ข้อมูลพื้นฐาน — ไม่มี PII)
type DirectoryClassDTO struct {
	ID           string `json:"id"`
	GradeLevel   string `json:"grade_level"`
	RoomName     string `json:"room_name"`
	StudentCount int    `json:"student_count"`
}

type DirectoryStudentDTO struct {
	StudentID   string `json:"student_id"`
	StudentCode string `json:"student_code"`
	Prefix      string `json:"prefix"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

type DirectoryStudentClassDTO struct {
	StudentID   string `json:"student_id"`
	StudentCode string `json:"student_code"`
	Prefix      string `json:"prefix"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	ClassLabel  string `json:"class_label"`
}

// authorize: นักเรียน (role student) เข้าไม่ได้ — ที่เหลือ (ครู/ผู้บริหาร/วิชาการ/admin) ดูได้
func (s *DirectoryService) authorize(ctx context.Context) error {
	if tenant.IsSchoolAdminFromContext(ctx) {
		return nil
	}
	if tenant.RoleFromContext(ctx) == domain.RoleStudent {
		return domain.ErrForbidden
	}
	return nil
}

// Classes คืนรายการห้องเรียนของเทอมปัจจุบัน (พร้อมจำนวนนักเรียน)
func (s *DirectoryService) Classes(ctx context.Context) ([]DirectoryClassDTO, error) {
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.classes.ListBySemester(ctx, tenant.SchoolIDFromContext(ctx), sem)
	if err != nil {
		return nil, err
	}
	out := make([]DirectoryClassDTO, 0, len(rows))
	for i := range rows {
		c := &rows[i]
		out = append(out, DirectoryClassDTO{
			ID: c.ID, GradeLevel: c.GradeLevel, RoomName: c.RoomName, StudentCount: c.StudentCount,
		})
	}
	return out, nil
}

// ClassStudents คืนรายชื่อนักเรียนในห้อง (ข้อมูลพื้นฐาน)
func (s *DirectoryService) ClassStudents(ctx context.Context, classID string) ([]DirectoryStudentDTO, error) {
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}
	rows, err := s.enrollments.ListByClass(ctx, tenant.SchoolIDFromContext(ctx), classID)
	if err != nil {
		return nil, err
	}
	out := make([]DirectoryStudentDTO, 0, len(rows))
	for i := range rows {
		e := &rows[i]
		out = append(out, DirectoryStudentDTO{
			StudentID: e.StudentID, StudentCode: e.StudentCode,
			Prefix: e.Prefix, FirstName: e.FirstName, LastName: e.LastName,
		})
	}
	return out, nil
}

// SearchStudents ค้นหานักเรียนว่าอยู่ห้องไหน (ตามชื่อ/รหัส)
func (s *DirectoryService) SearchStudents(ctx context.Context, term string) ([]DirectoryStudentClassDTO, error) {
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}
	term = strings.TrimSpace(term)
	if term == "" {
		return []DirectoryStudentClassDTO{}, nil
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.enrollments.SearchByName(ctx, tenant.SchoolIDFromContext(ctx), sem, term)
	if err != nil {
		return nil, err
	}
	out := make([]DirectoryStudentClassDTO, 0, len(rows))
	for i := range rows {
		b := &rows[i]
		out = append(out, DirectoryStudentClassDTO{
			StudentID: b.StudentID, StudentCode: b.StudentCode,
			Prefix: b.Prefix, FirstName: b.FirstName, LastName: b.LastName,
			ClassLabel: strings.TrimSpace(b.GradeLevel + " " + b.RoomName),
		})
	}
	return out, nil
}
