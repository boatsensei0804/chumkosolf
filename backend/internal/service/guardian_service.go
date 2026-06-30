package service

import (
	"context"
	"strings"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

type GuardianRepository interface {
	List(ctx context.Context, schoolID string, limit, offset int) ([]domain.Guardian, int, error)
	GetByID(ctx context.Context, schoolID, id string) (*domain.Guardian, error)
	Create(ctx context.Context, schoolID string, ng domain.NewGuardian, audit domain.AuditEntry) (string, error)
	Upsert(ctx context.Context, schoolID string, ng domain.NewGuardian) (string, error)
	Update(ctx context.Context, schoolID, id string, ug domain.UpdateGuardian, audit domain.AuditEntry) (bool, error)
	SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error)
}

type GuardianListItem struct {
	ID               string `json:"id"`
	Prefix           string `json:"prefix"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	NationalIDMasked string `json:"national_id_masked"`
	Phone            string `json:"phone"`
	CreatedAt        string `json:"created_at"`
}

type GuardianDetail struct {
	ID               string     `json:"id"`
	Prefix           string     `json:"prefix"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	NationalIDMasked string     `json:"national_id_masked"`
	BirthDate        string     `json:"birth_date"`
	Phone            string     `json:"phone"`
	Address          AddressDTO `json:"address"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
}

type CreateGuardianInput struct {
	NationalID string
	Prefix     string
	FirstName  string
	LastName   string
	BirthDate  *time.Time
	Phone      string
	Address    AddressDTO
}

type UpdateGuardianInput = CreateGuardianInput // โครงเดียวกัน (national_id ""=ไม่เปลี่ยน ตอน update)

type GuardianService struct {
	guard  academicGuard
	repo   GuardianRepository
	cipher Cipher
}

func NewGuardianService(repo GuardianRepository, checker WorkGroupChecker, cipher Cipher) *GuardianService {
	return &GuardianService{guard: academicGuard{checker: checker}, repo: repo, cipher: cipher}
}

func (s *GuardianService) List(ctx context.Context, page, pageSize int) ([]GuardianListItem, int, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	rows, total, err := s.repo.List(ctx, tenant.SchoolIDFromContext(ctx), pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	items := make([]GuardianListItem, 0, len(rows))
	for i := range rows {
		g := &rows[i]
		items = append(items, GuardianListItem{
			ID: g.ID, Prefix: g.Profile.Prefix, FirstName: g.Profile.FirstName, LastName: g.Profile.LastName,
			NationalIDMasked: maskNID(s.cipher, g.NationalIDEnc), Phone: g.Profile.Phone,
			CreatedAt: g.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, total, nil
}

func (s *GuardianService) Get(ctx context.Context, id string) (*GuardianDetail, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	g, err := s.repo.GetByID(ctx, tenant.SchoolIDFromContext(ctx), id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.ErrGuardianNotFound
	}
	birth := ""
	if g.Profile.BirthDate != nil {
		birth = g.Profile.BirthDate.Format("2006-01-02")
	}
	return &GuardianDetail{
		ID: g.ID, Prefix: g.Profile.Prefix, FirstName: g.Profile.FirstName, LastName: g.Profile.LastName,
		NationalIDMasked: maskNID(s.cipher, g.NationalIDEnc), BirthDate: birth, Phone: g.Profile.Phone,
		Address: addressDTO(g.Profile.Address),
		CreatedAt: g.CreatedAt.Format(time.RFC3339), UpdatedAt: g.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *GuardianService) Create(ctx context.Context, in CreateGuardianInput) (string, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return "", err
	}
	if !nationalIDPattern.MatchString(in.NationalID) {
		return "", &domain.Error{Status: 400, Code: "INVALID_NATIONAL_ID", Message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"}
	}
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return "", domain.ErrValidation
	}
	enc, err := s.cipher.Encrypt(in.NationalID)
	if err != nil {
		return "", err
	}
	audit := auditFor(ctx, domain.AuditCreate, "guardian", "", map[string]any{"fields": []string{"national_id", "name"}})
	return s.repo.Create(ctx, tenant.SchoolIDFromContext(ctx), domain.NewGuardian{
		NationalIDEnc: enc, NationalIDHash: s.cipher.Hash(in.NationalID),
		Profile: toPersonProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Address),
	}, audit)
}

func (s *GuardianService) Update(ctx context.Context, id string, in UpdateGuardianInput) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return domain.ErrValidation
	}
	ug := domain.UpdateGuardian{Profile: toPersonProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Address)}
	if in.NationalID != "" {
		if !nationalIDPattern.MatchString(in.NationalID) {
			return &domain.Error{Status: 400, Code: "INVALID_NATIONAL_ID", Message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"}
		}
		enc, err := s.cipher.Encrypt(in.NationalID)
		if err != nil {
			return err
		}
		ug.ChangeNationalID = true
		ug.NationalIDEnc = enc
		ug.NationalIDHash = s.cipher.Hash(in.NationalID)
	}
	audit := auditFor(ctx, domain.AuditUpdate, "guardian", id, map[string]any{"action": "update"})
	found, err := s.repo.Update(ctx, tenant.SchoolIDFromContext(ctx), id, ug, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrGuardianNotFound
	}
	return nil
}

func (s *GuardianService) Delete(ctx context.Context, id string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "guardian", id, map[string]any{"action": "soft_delete"})
	found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrGuardianNotFound
	}
	return nil
}
