package service

import (
	"context"
	"strings"
	"time"

	"github.com/chumkosoft/backend/internal/crypto"
	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

const academicWorkGroupCode = "academic"

// WorkGroupChecker ตรวจการสังกัดกลุ่มงาน (สำหรับสิทธิ์กลุ่มวิชาการ)
type WorkGroupChecker interface {
	IsUserInWorkGroup(ctx context.Context, schoolID, userID, code string) (bool, error)
}

// academicGuard: school admin หรือสมาชิกกลุ่มวิชาการเท่านั้น (ข้อมูลนักเรียน/ผู้ปกครอง — PDPA)
type academicGuard struct {
	checker WorkGroupChecker
}

func (a academicGuard) authorize(ctx context.Context) error {
	if tenant.IsSchoolAdminFromContext(ctx) {
		return nil
	}
	ok, err := a.checker.IsUserInWorkGroup(ctx, tenant.SchoolIDFromContext(ctx), tenant.UserIDFromContext(ctx), academicWorkGroupCode)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

// maskNID ถอดรหัสแล้ว mask เลขบัตร (ไม่ส่งเลขเต็มออก)
func maskNID(cipher Cipher, enc []byte) string {
	if len(enc) == 0 {
		return ""
	}
	plain, err := cipher.Decrypt(enc)
	if err != nil {
		return crypto.Mask("")
	}
	return crypto.Mask(plain)
}

func toPersonProfile(prefix, first, last string, birth *time.Time, phone string, addr AddressDTO) domain.PersonProfile {
	return domain.PersonProfile{
		Prefix:    strings.TrimSpace(prefix),
		FirstName: strings.TrimSpace(first),
		LastName:  strings.TrimSpace(last),
		BirthDate: birth,
		Phone:     strings.TrimSpace(phone),
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

func addressDTO(a domain.Address) AddressDTO {
	return AddressDTO{
		HouseNo: a.HouseNo, Moo: a.Moo, Road: a.Road, Subdistrict: a.Subdistrict,
		District: a.District, Province: a.Province, PostalCode: a.PostalCode,
	}
}

// ================= Students =================

type StudentRepository interface {
	List(ctx context.Context, schoolID string, limit, offset int, search string) ([]domain.Student, int, error)
	GetByID(ctx context.Context, schoolID, id string) (*domain.Student, error)
	CurrentClass(ctx context.Context, schoolID, semesterID, studentID string) (enrollmentID, classID, label string, err error)
	Create(ctx context.Context, schoolID string, ns domain.NewStudent, audit domain.AuditEntry) (string, error)
	Update(ctx context.Context, schoolID, id string, us domain.UpdateStudent, audit domain.AuditEntry) (bool, error)
	SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error)
}

type StudentListItem struct {
	ID               string `json:"id"`
	StudentCode      string `json:"student_code"`
	Status           string `json:"status"`
	Prefix           string `json:"prefix"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	NationalIDMasked string `json:"national_id_masked"`
	Phone            string `json:"phone"`
	CreatedAt        string `json:"created_at"`
}

type StudentDetail struct {
	ID               string     `json:"id"`
	StudentCode      string     `json:"student_code"`
	Status           string     `json:"status"`
	Prefix           string     `json:"prefix"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	NationalIDMasked string     `json:"national_id_masked"`
	BirthDate        string     `json:"birth_date"`
	Phone            string     `json:"phone"`
	Address          AddressDTO `json:"address"`
	PhotoPath        string     `json:"photo_path"`
	// ห้องเรียนของเทอมปัจจุบัน (ว่างถ้ายังไม่จัดห้อง)
	CurrentClassID     string `json:"current_class_id"`
	CurrentClassLabel  string `json:"current_class_label"`
	CurrentEnrollmentID string `json:"current_enrollment_id"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type CreateStudentInput struct {
	NationalID  string
	StudentCode string
	Status      string
	Prefix      string
	FirstName   string
	LastName    string
	BirthDate   *time.Time
	Phone       string
	Address     AddressDTO
}

type UpdateStudentInput struct {
	NationalID  string
	StudentCode string
	Status      string
	Prefix      string
	FirstName   string
	LastName    string
	BirthDate   *time.Time
	Phone       string
	Address     AddressDTO
}

// normalizeStatus คืนสถานะที่ถูกต้อง (ว่าง = กำลังศึกษา); "" ถ้าไม่ถูกต้อง
func normalizeStatus(status string) string {
	if status == "" {
		return domain.StudentStatusStudying
	}
	if !domain.ValidStudentStatus(status) {
		return ""
	}
	return status
}

type StudentService struct {
	guard  academicGuard
	repo   StudentRepository
	cipher Cipher
}

func NewStudentService(repo StudentRepository, checker WorkGroupChecker, cipher Cipher) *StudentService {
	return &StudentService{guard: academicGuard{checker: checker}, repo: repo, cipher: cipher}
}

func (s *StudentService) List(ctx context.Context, page, pageSize int, search string) ([]StudentListItem, int, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	rows, total, err := s.repo.List(ctx, tenant.SchoolIDFromContext(ctx), pageSize, (page-1)*pageSize, search)
	if err != nil {
		return nil, 0, err
	}
	items := make([]StudentListItem, 0, len(rows))
	for i := range rows {
		st := &rows[i]
		items = append(items, StudentListItem{
			ID: st.ID, StudentCode: st.StudentCode, Status: st.Status, Prefix: st.Profile.Prefix,
			FirstName: st.Profile.FirstName, LastName: st.Profile.LastName,
			NationalIDMasked: maskNID(s.cipher, st.NationalIDEnc), Phone: st.Profile.Phone,
			CreatedAt: st.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, total, nil
}

func (s *StudentService) Get(ctx context.Context, id string) (*StudentDetail, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	st, err := s.repo.GetByID(ctx, tenant.SchoolIDFromContext(ctx), id)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return nil, domain.ErrStudentNotFound
	}
	birth := ""
	if st.Profile.BirthDate != nil {
		birth = st.Profile.BirthDate.Format("2006-01-02")
	}
	// ห้องของเทอมปัจจุบัน (ถ้ามี semester ใน context)
	var enrollmentID, classID, classLabel string
	if sem := tenant.SemesterIDFromContext(ctx); sem != "" {
		enrollmentID, classID, classLabel, err = s.repo.CurrentClass(ctx, tenant.SchoolIDFromContext(ctx), sem, id)
		if err != nil {
			return nil, err
		}
	}
	return &StudentDetail{
		ID: st.ID, StudentCode: st.StudentCode, Status: st.Status, Prefix: st.Profile.Prefix,
		FirstName: st.Profile.FirstName, LastName: st.Profile.LastName,
		NationalIDMasked: maskNID(s.cipher, st.NationalIDEnc), BirthDate: birth,
		Phone: st.Profile.Phone, Address: addressDTO(st.Profile.Address), PhotoPath: st.PhotoPath,
		CurrentClassID: classID, CurrentClassLabel: classLabel, CurrentEnrollmentID: enrollmentID,
		CreatedAt: st.CreatedAt.Format(time.RFC3339), UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *StudentService) Create(ctx context.Context, in CreateStudentInput) (string, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return "", err
	}
	if !nationalIDPattern.MatchString(in.NationalID) {
		return "", &domain.Error{Status: 400, Code: "INVALID_NATIONAL_ID", Message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"}
	}
	if strings.TrimSpace(in.StudentCode) == "" || strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return "", domain.ErrValidation
	}
	status := normalizeStatus(in.Status)
	if status == "" {
		return "", domain.ErrValidation
	}
	enc, err := s.cipher.Encrypt(in.NationalID)
	if err != nil {
		return "", err
	}
	audit := auditFor(ctx, domain.AuditCreate, "student", "", map[string]any{"fields": []string{"national_id", "student_code", "name"}})
	return s.repo.Create(ctx, tenant.SchoolIDFromContext(ctx), domain.NewStudent{
		NationalIDEnc: enc, NationalIDHash: s.cipher.Hash(in.NationalID),
		StudentCode: strings.TrimSpace(in.StudentCode), Status: status,
		Profile: toPersonProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Address),
	}, audit)
}

func (s *StudentService) Update(ctx context.Context, id string, in UpdateStudentInput) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(in.StudentCode) == "" || strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return domain.ErrValidation
	}
	status := normalizeStatus(in.Status)
	if status == "" {
		return domain.ErrValidation
	}
	up := domain.UpdateStudent{
		StudentCode: strings.TrimSpace(in.StudentCode), Status: status,
		Profile: toPersonProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Address),
	}
	touched := []string{"student_code", "name", "address", "contact"}
	if in.NationalID != "" {
		if !nationalIDPattern.MatchString(in.NationalID) {
			return &domain.Error{Status: 400, Code: "INVALID_NATIONAL_ID", Message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"}
		}
		enc, err := s.cipher.Encrypt(in.NationalID)
		if err != nil {
			return err
		}
		up.ChangeNationalID = true
		up.NationalIDEnc = enc
		up.NationalIDHash = s.cipher.Hash(in.NationalID)
		touched = append(touched, "national_id")
	}
	audit := auditFor(ctx, domain.AuditUpdate, "student", id, map[string]any{"fields": touched})
	found, err := s.repo.Update(ctx, tenant.SchoolIDFromContext(ctx), id, up, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrStudentNotFound
	}
	return nil
}

func (s *StudentService) Delete(ctx context.Context, id string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "student", id, map[string]any{"action": "soft_delete"})
	found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrStudentNotFound
	}
	return nil
}

// ================= Student ↔ Guardian links =================

type StudentGuardianRepository interface {
	ListByStudent(ctx context.Context, schoolID, studentID string) ([]domain.StudentGuardian, error)
	Link(ctx context.Context, schoolID, studentID string, nsg domain.NewStudentGuardian, audit domain.AuditEntry) error
	Unlink(ctx context.Context, schoolID, studentID, linkID string, audit domain.AuditEntry) (bool, error)
}

type StudentGuardianDTO struct {
	ID               string `json:"id"`
	GuardianID       string `json:"guardian_id"`
	Relationship     string `json:"relationship"`
	IsPrimary        bool   `json:"is_primary"`
	Prefix           string `json:"prefix"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Phone            string `json:"phone"`
	NationalIDMasked string `json:"national_id_masked"`
}

// LinkGuardianInput — ข้อมูลผู้ปกครอง (สร้าง inline) + ความสัมพันธ์ (ไม่ต้องเลือกจากรายการแยก)
type LinkGuardianInput struct {
	NationalID   string
	Prefix       string
	FirstName    string
	LastName     string
	BirthDate    *time.Time
	Phone        string
	Address      AddressDTO
	Relationship string
	IsPrimary    bool
}

// StudentGuardianService จัดการการเชื่อมผู้ปกครองเข้านักเรียน (ต้องมีนักเรียน+ผู้ปกครองจริงในโรงเรียน)
type StudentGuardianService struct {
	guard        academicGuard
	repo         StudentGuardianRepository
	students     StudentRepository
	guardians    GuardianRepository
	cipher       Cipher
}

func NewStudentGuardianService(repo StudentGuardianRepository, students StudentRepository, guardians GuardianRepository, checker WorkGroupChecker, cipher Cipher) *StudentGuardianService {
	return &StudentGuardianService{guard: academicGuard{checker: checker}, repo: repo, students: students, guardians: guardians, cipher: cipher}
}

func (s *StudentGuardianService) ensureStudent(ctx context.Context, studentID string) error {
	st, err := s.students.GetByID(ctx, tenant.SchoolIDFromContext(ctx), studentID)
	if err != nil {
		return err
	}
	if st == nil {
		return domain.ErrStudentNotFound
	}
	return nil
}

func (s *StudentGuardianService) List(ctx context.Context, studentID string) ([]StudentGuardianDTO, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	if err := s.ensureStudent(ctx, studentID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByStudent(ctx, tenant.SchoolIDFromContext(ctx), studentID)
	if err != nil {
		return nil, err
	}
	out := make([]StudentGuardianDTO, 0, len(rows))
	for i := range rows {
		g := &rows[i]
		out = append(out, StudentGuardianDTO{
			ID: g.ID, GuardianID: g.GuardianID, Relationship: g.Relationship, IsPrimary: g.IsPrimary,
			Prefix: g.Prefix, FirstName: g.FirstName, LastName: g.LastName, Phone: g.Phone,
			NationalIDMasked: maskNID(s.cipher, g.NationalIDEnc),
		})
	}
	return out, nil
}

func (s *StudentGuardianService) Link(ctx context.Context, studentID string, in LinkGuardianInput) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	if in.Relationship != domain.RelationshipFather && in.Relationship != domain.RelationshipMother && in.Relationship != domain.RelationshipOther {
		return domain.ErrInvalidRelationship
	}
	if !nationalIDPattern.MatchString(in.NationalID) {
		return &domain.Error{Status: 400, Code: "INVALID_NATIONAL_ID", Message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"}
	}
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return domain.ErrValidation
	}
	if err := s.ensureStudent(ctx, studentID); err != nil {
		return err
	}

	schoolID := tenant.SchoolIDFromContext(ctx)
	enc, err := s.cipher.Encrypt(in.NationalID)
	if err != nil {
		return err
	}
	// สร้างผู้ปกครอง หรือใช้คนเดิมถ้าเลขบัตรซ้ำ (พี่น้อง)
	guardianID, err := s.guardians.Upsert(ctx, schoolID, domain.NewGuardian{
		NationalIDEnc:  enc,
		NationalIDHash: s.cipher.Hash(in.NationalID),
		Profile:        toPersonProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Address),
	})
	if err != nil {
		return err
	}

	audit := auditFor(ctx, domain.AuditCreate, "student_guardian", studentID, map[string]any{"guardian_id": guardianID, "is_primary": in.IsPrimary})
	return s.repo.Link(ctx, schoolID, studentID, domain.NewStudentGuardian{
		GuardianID: guardianID, Relationship: in.Relationship, IsPrimary: in.IsPrimary,
	}, audit)
}

func (s *StudentGuardianService) Unlink(ctx context.Context, studentID, linkID string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	if err := s.ensureStudent(ctx, studentID); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "student_guardian", studentID, map[string]any{"link_id": linkID})
	found, err := s.repo.Unlink(ctx, tenant.SchoolIDFromContext(ctx), studentID, linkID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrGuardianLinkNotFound
	}
	return nil
}
