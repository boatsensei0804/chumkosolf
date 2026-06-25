package service

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/chumko-platform/backend/internal/crypto"
	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

const personnelWorkGroupCode = "personnel"

// PersonnelRepository คือ contract ของชั้น DB สำหรับบุคลากร (ทุก method scope school_id)
type PersonnelRepository interface {
	Create(ctx context.Context, schoolID string, np domain.NewPersonnel, audit domain.AuditEntry) (string, error)
	List(ctx context.Context, schoolID string, limit, offset int) ([]domain.Personnel, int, error)
	GetByID(ctx context.Context, schoolID, id string) (*domain.Personnel, error)
	Update(ctx context.Context, schoolID, id string, up domain.UpdatePersonnel, audit domain.AuditEntry) (bool, error)
	SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error)
	InsertAudit(ctx context.Context, audit domain.AuditEntry) error
	IsUserInWorkGroup(ctx context.Context, schoolID, userID, groupCode string) (bool, error)
}

// Cipher คือ contract สำหรับเข้ารหัส/ถอดรหัส/hash ข้อมูลอ่อนไหว (crypto.Cipher ทำ implement)
type Cipher interface {
	Encrypt(plaintext string) ([]byte, error)
	Decrypt(ciphertext []byte) (string, error)
	Hash(value string) string
}

// PersonnelService รวม business logic ของกลุ่มงานบุคคล
type PersonnelService struct {
	repo   PersonnelRepository
	cipher Cipher
}

// NewPersonnelService สร้าง service
func NewPersonnelService(repo PersonnelRepository, cipher Cipher) *PersonnelService {
	return &PersonnelService{repo: repo, cipher: cipher}
}

// --- DTOs ---

// AddressDTO ที่อยู่แบบแยกฟิลด์
type AddressDTO struct {
	HouseNo     string `json:"house_no"`
	Moo         string `json:"moo"`
	Road        string `json:"road"`
	Subdistrict string `json:"subdistrict"`
	District    string `json:"district"`
	Province    string `json:"province"`
	PostalCode  string `json:"postal_code"`
}

// PersonnelListItem ข้อมูลย่อสำหรับตาราง (เลขบัตร mask เสมอ)
type PersonnelListItem struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	Role             string `json:"role"`
	IsActive         bool   `json:"is_active"`
	Prefix           string `json:"prefix"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	NationalIDMasked string `json:"national_id_masked"`
	Phone            string `json:"phone"`
	CreatedAt        string `json:"created_at"`
}

// PersonnelDetail ข้อมูลเต็ม (ยัง mask เลขบัตร — เลขเต็มต้องผ่าน endpoint reveal แยกในอนาคต)
type PersonnelDetail struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	Username             string     `json:"username"`
	Role                 string     `json:"role"`
	IsActive             bool       `json:"is_active"`
	Prefix               string     `json:"prefix"`
	FirstName            string     `json:"first_name"`
	LastName             string     `json:"last_name"`
	NationalIDMasked     string     `json:"national_id_masked"`
	CivilServantIDMasked string     `json:"civil_servant_id_masked"`
	BirthDate            string     `json:"birth_date"`
	Phone                string     `json:"phone"`
	Email                string     `json:"email"`
	Address              AddressDTO `json:"address"`
	PhotoPath            string     `json:"photo_path"`
	CreatedAt            string     `json:"created_at"`
	UpdatedAt            string     `json:"updated_at"`
}

// CreatePersonnelInput ข้อมูลสำหรับสร้างบุคลากร (พร้อมบัญชี user)
type CreatePersonnelInput struct {
	Username       string
	Password       string
	Role           string
	NationalID     string
	CivilServantID string
	Prefix         string
	FirstName      string
	LastName       string
	BirthDate      *time.Time
	Phone          string
	Email          string
	Address        AddressDTO
}

// UpdatePersonnelInput ข้อมูลแก้ไข (NationalID/CivilServantID ว่าง = ไม่เปลี่ยน)
type UpdatePersonnelInput struct {
	NationalID     string
	CivilServantID string
	Prefix         string
	FirstName      string
	LastName       string
	BirthDate      *time.Time
	Phone          string
	Email          string
	Address        AddressDTO
}

var nationalIDPattern = regexp.MustCompile(`^[0-9]{13}$`)

// --- use cases ---

// List คืนรายการบุคลากร (แบ่งหน้า) — เฉพาะผู้มีสิทธิ์กลุ่มงานบุคคล
func (s *PersonnelService) List(ctx context.Context, page, pageSize int) ([]PersonnelListItem, int, error) {
	if err := s.ensureCanManage(ctx); err != nil {
		return nil, 0, err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, total, err := s.repo.List(ctx, schoolID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	items := make([]PersonnelListItem, 0, len(rows))
	for i := range rows {
		p := &rows[i]
		items = append(items, PersonnelListItem{
			ID:               p.ID,
			UserID:           p.UserID,
			Username:         p.Username,
			Role:             p.Role,
			IsActive:         p.IsActive,
			Prefix:           p.Profile.Prefix,
			FirstName:        p.Profile.FirstName,
			LastName:         p.Profile.LastName,
			NationalIDMasked: s.maskNationalID(p.NationalIDEnc),
			Phone:            p.Profile.Phone,
			CreatedAt:        p.CreatedAt.Format(time.RFC3339),
		})
	}

	// audit การดูรายการ (best-effort — ไม่ให้ list ล้มเพราะ audit)
	s.auditViewBestEffort(ctx, "", map[string]any{"action": "list", "count": len(items)})
	return items, total, nil
}

// Get คืนข้อมูลบุคลากร 1 ราย
func (s *PersonnelService) Get(ctx context.Context, id string) (*PersonnelDetail, error) {
	if err := s.ensureCanManage(ctx); err != nil {
		return nil, err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)

	p, err := s.repo.GetByID(ctx, schoolID, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrPersonnelNotFound
	}

	detail := s.toDetail(p)
	s.auditViewBestEffort(ctx, id, map[string]any{"action": "view"})
	return &detail, nil
}

// Create สร้างบุคลากรพร้อมบัญชี user
func (s *PersonnelService) Create(ctx context.Context, in CreatePersonnelInput) (string, error) {
	if err := s.ensureCanManage(ctx); err != nil {
		return "", err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)

	if err := validateRole(in.Role); err != nil {
		return "", err
	}
	if !nationalIDPattern.MatchString(in.NationalID) {
		return "", &domain.Error{Status: 400, Code: "INVALID_NATIONAL_ID", Message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"}
	}
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return "", domain.ErrValidation
	}

	natEnc, err := s.cipher.Encrypt(in.NationalID)
	if err != nil {
		return "", err
	}
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	np := domain.NewPersonnel{
		Username:       strings.TrimSpace(in.Username),
		PasswordHash:   string(pwdHash),
		Role:           in.Role,
		NationalIDEnc:  natEnc,
		NationalIDHash: s.cipher.Hash(in.NationalID),
		Profile:        in.profile(),
	}
	if in.CivilServantID != "" {
		civEnc, err := s.cipher.Encrypt(in.CivilServantID)
		if err != nil {
			return "", err
		}
		np.CivilServantIDEnc = civEnc
		np.CivilServantIDHash = s.cipher.Hash(in.CivilServantID)
	}

	audit := s.auditEntry(ctx, domain.AuditCreate, "", map[string]any{
		"fields": []string{"national_id", "first_name", "last_name", "user_account"},
	})
	return s.repo.Create(ctx, schoolID, np, audit)
}

// Update แก้ไขข้อมูลบุคลากร
func (s *PersonnelService) Update(ctx context.Context, id string, in UpdatePersonnelInput) error {
	if err := s.ensureCanManage(ctx); err != nil {
		return err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)

	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return domain.ErrValidation
	}

	up := domain.UpdatePersonnel{Profile: in.profile()}
	touched := []string{"first_name", "last_name", "address", "contact"}

	if in.NationalID != "" {
		if !nationalIDPattern.MatchString(in.NationalID) {
			return &domain.Error{Status: 400, Code: "INVALID_NATIONAL_ID", Message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"}
		}
		natEnc, err := s.cipher.Encrypt(in.NationalID)
		if err != nil {
			return err
		}
		up.ChangeNationalID = true
		up.NationalIDEnc = natEnc
		up.NationalIDHash = s.cipher.Hash(in.NationalID)
		touched = append(touched, "national_id")
	}
	if in.CivilServantID != "" {
		civEnc, err := s.cipher.Encrypt(in.CivilServantID)
		if err != nil {
			return err
		}
		up.ChangeCivilID = true
		up.CivilServantIDEnc = civEnc
		up.CivilServantIDHash = s.cipher.Hash(in.CivilServantID)
		touched = append(touched, "civil_servant_id")
	}

	audit := s.auditEntry(ctx, domain.AuditUpdate, id, map[string]any{"fields": touched})
	found, err := s.repo.Update(ctx, schoolID, id, up, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrPersonnelNotFound
	}
	return nil
}

// Delete ลบบุคลากร (soft delete + ปิดบัญชี user)
func (s *PersonnelService) Delete(ctx context.Context, id string) error {
	if err := s.ensureCanManage(ctx); err != nil {
		return err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)

	audit := s.auditEntry(ctx, domain.AuditDelete, id, map[string]any{"action": "soft_delete"})
	found, err := s.repo.SoftDelete(ctx, schoolID, id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrPersonnelNotFound
	}
	return nil
}

// --- helpers ---

// ensureCanManage: school admin หรือสมาชิกกลุ่มงานบุคคลเท่านั้น
func (s *PersonnelService) ensureCanManage(ctx context.Context) error {
	if tenant.IsSchoolAdminFromContext(ctx) {
		return nil
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	userID := tenant.UserIDFromContext(ctx)
	ok, err := s.repo.IsUserInWorkGroup(ctx, schoolID, userID, personnelWorkGroupCode)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *PersonnelService) toDetail(p *domain.Personnel) PersonnelDetail {
	birth := ""
	if p.Profile.BirthDate != nil {
		birth = p.Profile.BirthDate.Format("2006-01-02")
	}
	return PersonnelDetail{
		ID:                   p.ID,
		UserID:               p.UserID,
		Username:             p.Username,
		Role:                 p.Role,
		IsActive:             p.IsActive,
		Prefix:               p.Profile.Prefix,
		FirstName:            p.Profile.FirstName,
		LastName:             p.Profile.LastName,
		NationalIDMasked:     s.maskNationalID(p.NationalIDEnc),
		CivilServantIDMasked: s.maskCivilID(p.CivilServantIDEnc),
		BirthDate:            birth,
		Phone:                p.Profile.Phone,
		Email:                p.Profile.Email,
		Address: AddressDTO{
			HouseNo:     p.Profile.Address.HouseNo,
			Moo:         p.Profile.Address.Moo,
			Road:        p.Profile.Address.Road,
			Subdistrict: p.Profile.Address.Subdistrict,
			District:    p.Profile.Address.District,
			Province:    p.Profile.Address.Province,
			PostalCode:  p.Profile.Address.PostalCode,
		},
		PhotoPath: p.PhotoPath,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

// maskNationalID ถอดรหัสแล้ว mask (ไม่ส่งเลขเต็มออก); ถ้าถอดไม่ได้คืน mask ว่าง
func (s *PersonnelService) maskNationalID(enc []byte) string {
	if len(enc) == 0 {
		return ""
	}
	plain, err := s.cipher.Decrypt(enc)
	if err != nil {
		return crypto.Mask("")
	}
	return crypto.Mask(plain)
}

// maskCivilID เปิดเผยเฉพาะ 4 ตัวท้าย
func (s *PersonnelService) maskCivilID(enc []byte) string {
	if len(enc) == 0 {
		return ""
	}
	plain, err := s.cipher.Decrypt(enc)
	if err != nil || len(plain) < 4 {
		return "xxxx"
	}
	return strings.Repeat("x", len(plain)-4) + plain[len(plain)-4:]
}

func (s *PersonnelService) auditEntry(ctx context.Context, action, targetID string, detail map[string]any) domain.AuditEntry {
	return domain.AuditEntry{
		SchoolID:    tenant.SchoolIDFromContext(ctx),
		ActorUserID: tenant.UserIDFromContext(ctx),
		Action:      action,
		TargetType:  "personnel",
		TargetID:    targetID,
		Detail:      detail,
		IPAddress:   tenant.IPAddressFromContext(ctx),
	}
}

func (s *PersonnelService) auditViewBestEffort(ctx context.Context, targetID string, detail map[string]any) {
	if err := s.repo.InsertAudit(ctx, s.auditEntry(ctx, domain.AuditView, targetID, detail)); err != nil {
		log.Printf("personnel: บันทึก audit (view) ล้มเหลว: %v", err)
	}
}

func validateRole(role string) error {
	if role == domain.RoleTeacher || role == domain.RoleExecutive {
		return nil
	}
	return &domain.Error{
		Status:  400,
		Code:    "INVALID_ROLE",
		Message: "ตำแหน่งต้องเป็นครูหรือผู้บริหารเท่านั้น",
	}
}

func (in CreatePersonnelInput) profile() domain.PersonnelProfile { return toProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Email, in.Address) }
func (in UpdatePersonnelInput) profile() domain.PersonnelProfile { return toProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Email, in.Address) }

func toProfile(prefix, first, last string, birth *time.Time, phone, email string, addr AddressDTO) domain.PersonnelProfile {
	return domain.PersonnelProfile{
		Prefix:    strings.TrimSpace(prefix),
		FirstName: strings.TrimSpace(first),
		LastName:  strings.TrimSpace(last),
		BirthDate: birth,
		Phone:     strings.TrimSpace(phone),
		Email:     strings.TrimSpace(email),
		Address: domain.Address{
			HouseNo:     strings.TrimSpace(addr.HouseNo),
			Moo:         strings.TrimSpace(addr.Moo),
			Road:        strings.TrimSpace(addr.Road),
			Subdistrict: strings.TrimSpace(addr.Subdistrict),
			District:    strings.TrimSpace(addr.District),
			Province:    strings.TrimSpace(addr.Province),
			PostalCode:  strings.TrimSpace(addr.PostalCode),
		},
	}
}
