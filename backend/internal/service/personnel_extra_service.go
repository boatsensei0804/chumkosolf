package service

import (
	"context"
	"strings"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

// personnelGuard: ตรวจสิทธิ์กลุ่มงานบุคคล + ยืนยันว่ามี personnel จริงในโรงเรียน
type personnelGuard interface {
	IsUserInWorkGroup(ctx context.Context, schoolID, userID, groupCode string) (bool, error)
	GetByID(ctx context.Context, schoolID, id string) (*domain.Personnel, error)
}

// personnelAccess รวม logic การอนุญาตเข้าถึง sub-resource ของบุคลากร (ใช้ร่วมหลาย service)
type personnelAccess struct {
	guard personnelGuard
}

// authorize: ต้องเป็น school admin หรือสมาชิกกลุ่มบุคคล และ personnel ต้องมีอยู่ในโรงเรียน (scope)
func (a personnelAccess) authorize(ctx context.Context, personnelID string) error {
	schoolID := tenant.SchoolIDFromContext(ctx)
	if !tenant.IsSchoolAdminFromContext(ctx) {
		ok, err := a.guard.IsUserInWorkGroup(ctx, schoolID, tenant.UserIDFromContext(ctx), personnelWorkGroupCode)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}
	}
	p, err := a.guard.GetByID(ctx, schoolID, personnelID)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrPersonnelNotFound
	}
	return nil
}

func auditFor(ctx context.Context, action, targetType, targetID string, detail map[string]any) domain.AuditEntry {
	return domain.AuditEntry{
		SchoolID:    tenant.SchoolIDFromContext(ctx),
		ActorUserID: tenant.UserIDFromContext(ctx),
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Detail:      detail,
		IPAddress:   tenant.IPAddressFromContext(ctx),
	}
}

func dateStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// ================= Admin positions =================

// AdminPositionRepository contract ของชั้น DB
type AdminPositionRepository interface {
	ListByPersonnel(ctx context.Context, schoolID, personnelID string) ([]domain.AdminPosition, error)
	Create(ctx context.Context, schoolID, personnelID string, np domain.NewAdminPosition, audit domain.AuditEntry) (string, error)
	SoftDelete(ctx context.Context, schoolID, personnelID, id string, audit domain.AuditEntry) (bool, error)
}

// AdminPositionDTO ข้อมูลตำแหน่งบริหารสำหรับ response
type AdminPositionDTO struct {
	ID          string `json:"id"`
	Position    string `json:"position"`
	IsActive    bool   `json:"is_active"`
	AppointedAt string `json:"appointed_at"`
	CreatedAt   string `json:"created_at"`
}

// CreateAdminPositionInput ข้อมูลสร้างตำแหน่งบริหาร
type CreateAdminPositionInput struct {
	Position    string
	IsActive    bool
	AppointedAt *time.Time
}

// AdminPositionService จัดการตำแหน่งบริหารของบุคลากร
type AdminPositionService struct {
	access personnelAccess
	repo   AdminPositionRepository
}

func NewAdminPositionService(repo AdminPositionRepository, guard personnelGuard) *AdminPositionService {
	return &AdminPositionService{access: personnelAccess{guard: guard}, repo: repo}
}

func (s *AdminPositionService) List(ctx context.Context, personnelID string) ([]AdminPositionDTO, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByPersonnel(ctx, tenant.SchoolIDFromContext(ctx), personnelID)
	if err != nil {
		return nil, err
	}
	out := make([]AdminPositionDTO, 0, len(rows))
	for i := range rows {
		out = append(out, AdminPositionDTO{
			ID:          rows[i].ID,
			Position:    rows[i].Position,
			IsActive:    rows[i].IsActive,
			AppointedAt: dateStr(rows[i].AppointedAt),
			CreatedAt:   rows[i].CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *AdminPositionService) Create(ctx context.Context, personnelID string, in CreateAdminPositionInput) (string, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return "", err
	}
	if in.Position != domain.PositionDirector && in.Position != domain.PositionDeputyDirector {
		return "", &domain.Error{Status: 400, Code: "INVALID_POSITION", Message: "ตำแหน่งต้องเป็นผู้อำนวยการหรือรองผู้อำนวยการ"}
	}
	audit := auditFor(ctx, domain.AuditCreate, "admin_position", "", map[string]any{"position": in.Position})
	return s.repo.Create(ctx, tenant.SchoolIDFromContext(ctx), personnelID, domain.NewAdminPosition{
		Position:    in.Position,
		IsActive:    in.IsActive,
		AppointedAt: in.AppointedAt,
	}, audit)
}

func (s *AdminPositionService) Delete(ctx context.Context, personnelID, id string) error {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "admin_position", id, map[string]any{"action": "soft_delete"})
	found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), personnelID, id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrAdminPositionNotFound
	}
	return nil
}

// ================= Academic standings =================

// AcademicStandingRepository contract ของชั้น DB
type AcademicStandingRepository interface {
	ListByPersonnel(ctx context.Context, schoolID, personnelID string) ([]domain.AcademicStanding, error)
	Create(ctx context.Context, schoolID, personnelID string, ns domain.NewAcademicStanding, audit domain.AuditEntry) (string, error)
	Update(ctx context.Context, schoolID, personnelID, id string, us domain.UpdateAcademicStanding, audit domain.AuditEntry) (bool, error)
	SoftDelete(ctx context.Context, schoolID, personnelID, id string, audit domain.AuditEntry) (bool, error)
}

// AcademicStandingDTO ข้อมูลวิทยฐานะสำหรับ response
type AcademicStandingDTO struct {
	ID            string `json:"id"`
	Standing      string `json:"standing"`
	EffectiveDate string `json:"effective_date"`
	IsCurrent     bool   `json:"is_current"`
	CreatedAt     string `json:"created_at"`
}

// StandingInput ข้อมูลสร้าง/แก้วิทยฐานะ
type StandingInput struct {
	Standing      string
	EffectiveDate *time.Time
	IsCurrent     bool
}

// AcademicStandingService จัดการวิทยฐานะ (ประวัติ)
type AcademicStandingService struct {
	access personnelAccess
	repo   AcademicStandingRepository
}

func NewAcademicStandingService(repo AcademicStandingRepository, guard personnelGuard) *AcademicStandingService {
	return &AcademicStandingService{access: personnelAccess{guard: guard}, repo: repo}
}

func (s *AcademicStandingService) List(ctx context.Context, personnelID string) ([]AcademicStandingDTO, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByPersonnel(ctx, tenant.SchoolIDFromContext(ctx), personnelID)
	if err != nil {
		return nil, err
	}
	out := make([]AcademicStandingDTO, 0, len(rows))
	for i := range rows {
		out = append(out, AcademicStandingDTO{
			ID:            rows[i].ID,
			Standing:      rows[i].Standing,
			EffectiveDate: dateStr(rows[i].EffectiveDate),
			IsCurrent:     rows[i].IsCurrent,
			CreatedAt:     rows[i].CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *AcademicStandingService) Create(ctx context.Context, personnelID string, in StandingInput) (string, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return "", err
	}
	if strings.TrimSpace(in.Standing) == "" {
		return "", domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditCreate, "academic_standing", "", map[string]any{"is_current": in.IsCurrent})
	return s.repo.Create(ctx, tenant.SchoolIDFromContext(ctx), personnelID, domain.NewAcademicStanding{
		Standing:      strings.TrimSpace(in.Standing),
		EffectiveDate: in.EffectiveDate,
		IsCurrent:     in.IsCurrent,
	}, audit)
}

func (s *AcademicStandingService) Update(ctx context.Context, personnelID, id string, in StandingInput) error {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return err
	}
	if strings.TrimSpace(in.Standing) == "" {
		return domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditUpdate, "academic_standing", id, map[string]any{"is_current": in.IsCurrent})
	found, err := s.repo.Update(ctx, tenant.SchoolIDFromContext(ctx), personnelID, id, domain.UpdateAcademicStanding{
		Standing:      strings.TrimSpace(in.Standing),
		EffectiveDate: in.EffectiveDate,
		IsCurrent:     in.IsCurrent,
	}, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrStandingNotFound
	}
	return nil
}

func (s *AcademicStandingService) Delete(ctx context.Context, personnelID, id string) error {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "academic_standing", id, map[string]any{"action": "soft_delete"})
	found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), personnelID, id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrStandingNotFound
	}
	return nil
}
