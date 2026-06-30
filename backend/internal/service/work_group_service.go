package service

import (
	"context"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

// WorkGroupRepository contract ของชั้น DB สำหรับกลุ่มงาน + การมอบหมาย
type WorkGroupRepository interface {
	ListBySchool(ctx context.Context, schoolID string) ([]domain.WorkGroup, error)
	ExistsInSchool(ctx context.Context, schoolID, workGroupID string) (bool, error)
	ListForUser(ctx context.Context, schoolID, userID string) ([]domain.WorkGroupMembership, error)
	IsGroupAdmin(ctx context.Context, schoolID, userID, workGroupID string) (bool, error)
	Assign(ctx context.Context, schoolID, userID, workGroupID string, isGroupAdmin bool, audit domain.AuditEntry) error
	Unassign(ctx context.Context, schoolID, userID, workGroupID string, audit domain.AuditEntry) (bool, error)
}

// WorkGroupService จัดการการมอบหมายกลุ่มงานให้บุคลากร
type WorkGroupService struct {
	repo   WorkGroupRepository
	access personnelAccess // ใช้ guard.GetByID เพื่อ resolve personnel → user_id + ตรวจ scope
}

func NewWorkGroupService(repo WorkGroupRepository, guard personnelGuard) *WorkGroupService {
	return &WorkGroupService{repo: repo, access: personnelAccess{guard: guard}}
}

// ListGroups คืนกลุ่มงานทั้งหมดของโรงเรียน (ให้เลือกตอนมอบหมาย)
func (s *WorkGroupService) ListGroups(ctx context.Context) ([]domain.WorkGroup, error) {
	groups, err := s.repo.ListBySchool(ctx, tenant.SchoolIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	if groups == nil {
		groups = []domain.WorkGroup{} // คืน [] ไม่ใช่ null เพื่อให้ frontend parse array ได้
	}
	return groups, nil
}

// ListForPersonnel คืนกลุ่มงานที่บุคลากรสังกัด (ต้องมีสิทธิ์จัดการบุคลากร)
func (s *WorkGroupService) ListForPersonnel(ctx context.Context, personnelID string) ([]domain.WorkGroupMembership, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return nil, err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	p, err := s.access.guard.GetByID(ctx, schoolID, personnelID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrPersonnelNotFound
	}
	memberships, err := s.repo.ListForUser(ctx, schoolID, p.UserID)
	if err != nil {
		return nil, err
	}
	if memberships == nil {
		memberships = []domain.WorkGroupMembership{} // คืน [] ไม่ใช่ null
	}
	return memberships, nil
}

// Assign มอบหมายบุคลากรเข้ากลุ่มงาน (School Admin หรือหัวหน้ากลุ่มงานนั้น)
func (s *WorkGroupService) Assign(ctx context.Context, personnelID, workGroupID string, isGroupAdmin bool) error {
	schoolID := tenant.SchoolIDFromContext(ctx)

	p, err := s.access.guard.GetByID(ctx, schoolID, personnelID)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrPersonnelNotFound
	}

	exists, err := s.repo.ExistsInSchool(ctx, schoolID, workGroupID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrWorkGroupNotFound
	}

	if err := s.canAssignToGroup(ctx, workGroupID); err != nil {
		return err
	}

	audit := auditFor(ctx, domain.AuditCreate, "work_group_assignment", workGroupID,
		map[string]any{"personnel_id": personnelID, "is_group_admin": isGroupAdmin})
	return s.repo.Assign(ctx, schoolID, p.UserID, workGroupID, isGroupAdmin, audit)
}

// Unassign ถอดบุคลากรออกจากกลุ่มงาน
func (s *WorkGroupService) Unassign(ctx context.Context, personnelID, workGroupID string) error {
	schoolID := tenant.SchoolIDFromContext(ctx)

	p, err := s.access.guard.GetByID(ctx, schoolID, personnelID)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrPersonnelNotFound
	}
	if err := s.canAssignToGroup(ctx, workGroupID); err != nil {
		return err
	}

	audit := auditFor(ctx, domain.AuditDelete, "work_group_assignment", workGroupID,
		map[string]any{"personnel_id": personnelID})
	found, err := s.repo.Unassign(ctx, schoolID, p.UserID, workGroupID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrWorkGroupAssignmentNotFound
	}
	return nil
}

// canAssignToGroup: school admin มอบหมายได้ทุกกลุ่ม; ไม่งั้นต้องเป็นหัวหน้ากลุ่มงานนั้น
func (s *WorkGroupService) canAssignToGroup(ctx context.Context, workGroupID string) error {
	if tenant.IsSchoolAdminFromContext(ctx) {
		return nil
	}
	ok, err := s.repo.IsGroupAdmin(ctx, tenant.SchoolIDFromContext(ctx), tenant.UserIDFromContext(ctx), workGroupID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}
