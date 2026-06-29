package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// hhmmPattern ตรวจรูปแบบเวลา HH:MM (00:00–23:59)
var hhmmPattern = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// SchoolRepository contract ของชั้น DB
type SchoolRepository interface {
	Get(ctx context.Context, schoolID string) (*domain.School, error)
	Update(ctx context.Context, schoolID string, us domain.UpdateSchool, audit domain.AuditEntry) (bool, error)
}

// SchoolDTO ข้อมูลโรงเรียนสำหรับ response
type SchoolDTO struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Code                string     `json:"code"`
	Address             AddressDTO `json:"address"`
	Phone               string     `json:"phone"`
	Email               string     `json:"email"`
	Website             string     `json:"website"`
	DirectorName          string     `json:"director_name"`
	IsActive              bool       `json:"is_active"`
	AttendanceLateAfter   string     `json:"attendance_late_after"`
	AttendanceLatePenalty int        `json:"attendance_late_penalty"`
}

// UpdateSchoolInput ข้อมูลแก้ไขโรงเรียน
type UpdateSchoolInput struct {
	Name                  string
	Address               AddressDTO
	Phone                 string
	Email                 string
	Website               string
	DirectorName          string
	AttendanceLateAfter   string
	AttendanceLatePenalty int
}

// SchoolService — อ่านข้อมูลโรงเรียน (ทุกคนที่ล็อกอิน), แก้ไข (school admin)
type SchoolService struct {
	repo SchoolRepository
}

func NewSchoolService(repo SchoolRepository) *SchoolService {
	return &SchoolService{repo: repo}
}

func (s *SchoolService) Get(ctx context.Context) (*SchoolDTO, error) {
	sc, err := s.repo.Get(ctx, tenant.SchoolIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, domain.ErrSchoolNotFound
	}
	return &SchoolDTO{
		ID: sc.ID, Name: sc.Name, Code: sc.Code, Address: addressDTO(sc.Address),
		Phone: sc.Phone, Email: sc.Email, Website: sc.Website, DirectorName: sc.DirectorName,
		IsActive: sc.IsActive, AttendanceLateAfter: sc.AttendanceLateAfter, AttendanceLatePenalty: sc.AttendanceLatePenalty,
	}, nil
}

func (s *SchoolService) Update(ctx context.Context, in UpdateSchoolInput) error {
	if err := requireSchoolAdmin(ctx); err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return domain.ErrValidation
	}
	lateAfter := strings.TrimSpace(in.AttendanceLateAfter)
	if lateAfter == "" {
		lateAfter = "08:00"
	}
	if !hhmmPattern.MatchString(lateAfter) {
		return &domain.Error{Status: 400, Code: "INVALID_TIME", Message: "เวลาต้องอยู่ในรูปแบบ HH:MM (เช่น 08:00)"}
	}
	if in.AttendanceLatePenalty < 0 || in.AttendanceLatePenalty > 100 {
		return &domain.Error{Status: 400, Code: "INVALID_INPUT", Message: "คะแนนหักต้องอยู่ระหว่าง 0–100"}
	}
	audit := auditFor(ctx, domain.AuditUpdate, "school", tenant.SchoolIDFromContext(ctx), map[string]any{"fields": []string{"name", "address", "contact", "director", "attendance_late_after", "attendance_late_penalty"}})
	found, err := s.repo.Update(ctx, tenant.SchoolIDFromContext(ctx), domain.UpdateSchool{
		Name: name,
		Address: domain.Address{
			HouseNo: strings.TrimSpace(in.Address.HouseNo), Moo: strings.TrimSpace(in.Address.Moo), Road: strings.TrimSpace(in.Address.Road),
			Subdistrict: strings.TrimSpace(in.Address.Subdistrict), District: strings.TrimSpace(in.Address.District),
			Province: strings.TrimSpace(in.Address.Province), PostalCode: strings.TrimSpace(in.Address.PostalCode),
		},
		Phone: strings.TrimSpace(in.Phone), Email: strings.TrimSpace(in.Email), Website: strings.TrimSpace(in.Website),
		DirectorName: strings.TrimSpace(in.DirectorName), AttendanceLateAfter: lateAfter, AttendanceLatePenalty: in.AttendanceLatePenalty,
	}, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrSchoolNotFound
	}
	return nil
}
