package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// MePersonnelRepository — เข้าถึง/แก้โปรไฟล์บุคลากรของผู้ใช้เอง (ไม่มี work-group guard)
type MePersonnelRepository interface {
	GetByUserID(ctx context.Context, schoolID, userID string) (*domain.Personnel, error)
	Update(ctx context.Context, schoolID, id string, up domain.UpdatePersonnel, audit domain.AuditEntry) (bool, error)
	InsertAudit(ctx context.Context, audit domain.AuditEntry) error
}

// MeAdviseeRepository — รายชื่อ + การตรวจความเป็นเจ้าของห้องที่ปรึกษาของผู้ใช้เอง
type MeAdviseeRepository interface {
	Advisees(ctx context.Context, schoolID, semesterID, userID string, date time.Time) ([]domain.Advisee, error)
	IsAdvisee(ctx context.Context, schoolID, semesterID, userID, studentID string) (bool, error)
}

// MeStudentRepository — อ่าน/แก้ข้อมูลนักเรียน (ครูที่ปรึกษาแก้ของห้องตัวเองได้)
type MeStudentRepository interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Student, error)
	Update(ctx context.Context, schoolID, id string, us domain.UpdateStudent, audit domain.AuditEntry) (bool, error)
}

// MeService รวม use case แบบ self-service: ผู้ใช้ดู/แก้ข้อมูลของตัวเองได้โดยไม่ต้องสังกัดกลุ่มงาน
// สิทธิ์มาจาก "ความเป็นเจ้าของ" (user_id จาก token) ไม่ใช่ role หรือกลุ่มงาน
type MeService struct {
	personnel MePersonnelRepository
	advisees  MeAdviseeRepository
	students  MeStudentRepository
	cipher    Cipher
	now       func() time.Time
}

func NewMeService(personnel MePersonnelRepository, advisees MeAdviseeRepository, students MeStudentRepository, cipher Cipher) *MeService {
	return &MeService{personnel: personnel, advisees: advisees, students: students, cipher: cipher, now: time.Now}
}

// UpdateMyProfileInput — ฟิลด์ที่ผู้ใช้แก้เองได้ (ไม่รวมเลขบัตร/เลขราชการ/role/username ซึ่งเป็นงานกลุ่มบุคคล)
type UpdateMyProfileInput struct {
	Prefix    string
	FirstName string
	LastName  string
	BirthDate *time.Time
	Phone     string
	Email     string
	Address   AddressDTO
}

// AdviseeDTO นักเรียนในห้องที่ปรึกษา (เลขบัตร mask เสมอ — PDPA)
type AdviseeDTO struct {
	StudentID        string `json:"student_id"`
	StudentCode      string `json:"student_code"`
	Prefix           string `json:"prefix"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Phone            string `json:"phone"`
	NationalIDMasked string `json:"national_id_masked"`
	ClassLabel       string `json:"class_label"`
	TodayStatus      string `json:"today_status"` // "" = ยังไม่เช็คชื่อวันนี้
}

var errProfileNotFound = &domain.Error{Status: 404, Code: "PROFILE_NOT_FOUND", Message: "ไม่พบข้อมูลบุคลากรของคุณ"}

// Profile คืนโปรไฟล์ของผู้ใช้เอง
func (s *MeService) Profile(ctx context.Context) (*PersonnelDetail, error) {
	schoolID := tenant.SchoolIDFromContext(ctx)
	userID := tenant.UserIDFromContext(ctx)

	p, err := s.personnel.GetByUserID(ctx, schoolID, userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errProfileNotFound
	}
	s.auditBestEffort(ctx, domain.AuditView, "personnel", p.ID, map[string]any{"action": "view", "self": true})
	detail := meDetail(s.cipher, p)
	return &detail, nil
}

// UpdateProfile แก้ข้อมูลของผู้ใช้เอง
func (s *MeService) UpdateProfile(ctx context.Context, in UpdateMyProfileInput) error {
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return domain.ErrValidation
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	userID := tenant.UserIDFromContext(ctx)

	p, err := s.personnel.GetByUserID(ctx, schoolID, userID)
	if err != nil {
		return err
	}
	if p == nil {
		return errProfileNotFound
	}

	up := domain.UpdatePersonnel{
		Profile: toProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Email, in.Address),
	}
	audit := auditFor(ctx, domain.AuditUpdate, "personnel", p.ID, map[string]any{
		"fields": []string{"first_name", "last_name", "address", "contact"}, "self": true,
	})
	found, err := s.personnel.Update(ctx, schoolID, p.ID, up, audit)
	if err != nil {
		return err
	}
	if !found {
		return errProfileNotFound
	}
	return nil
}

// Advisees คืนรายชื่อนักเรียนในห้องที่ปรึกษาของผู้ใช้ในเทอมปัจจุบัน (ว่างถ้าไม่ได้เป็นที่ปรึกษา)
func (s *MeService) Advisees(ctx context.Context) ([]AdviseeDTO, error) {
	sem := tenant.SemesterIDFromContext(ctx)
	if sem == "" {
		return []AdviseeDTO{}, nil // ยังไม่กำหนดเทอม → ไม่มีรายชื่อ
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	userID := tenant.UserIDFromContext(ctx)

	rows, err := s.advisees.Advisees(ctx, schoolID, sem, userID, s.now().UTC())
	if err != nil {
		return nil, err
	}
	out := make([]AdviseeDTO, 0, len(rows))
	for i := range rows {
		a := &rows[i]
		out = append(out, AdviseeDTO{
			StudentID: a.StudentID, StudentCode: a.StudentCode, Prefix: a.Prefix,
			FirstName: a.FirstName, LastName: a.LastName, Phone: a.Phone,
			NationalIDMasked: maskNID(s.cipher, a.NationalIDEnc),
			ClassLabel:       strings.TrimSpace(a.GradeLevel + " " + a.RoomName),
			TodayStatus:      a.TodayStatus,
		})
	}
	s.auditBestEffort(ctx, domain.AuditView, "advisees", "", map[string]any{"action": "list", "count": len(out)})
	return out, nil
}

// UpdateAdviseeInput — ฟิลด์ที่ครูที่ปรึกษาแก้ของนักเรียนในห้องตัวเองได้
// (ไม่รวมเลขบัตร/รหัสนักเรียน/สถานะ ซึ่งเป็นงานทะเบียนของกลุ่มวิชาการ)
type UpdateAdviseeInput struct {
	Prefix    string
	FirstName string
	LastName  string
	BirthDate *time.Time
	Phone     string
	Address   AddressDTO
}

// ensureAdvisee ตรวจว่านักเรียนอยู่ในห้องที่ปรึกษาของผู้ใช้ในเทอมปัจจุบัน (สิทธิ์จากความเป็นเจ้าของ)
func (s *MeService) ensureAdvisee(ctx context.Context, studentID string) error {
	sem := tenant.SemesterIDFromContext(ctx)
	if sem == "" {
		return domain.ErrForbidden
	}
	ok, err := s.advisees.IsAdvisee(ctx, tenant.SchoolIDFromContext(ctx), sem, tenant.UserIDFromContext(ctx), studentID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

// AdviseeDetail คืนข้อมูลนักเรียนในห้องที่ปรึกษาของผู้ใช้ (สำหรับฟอร์มแก้ไข)
func (s *MeService) AdviseeDetail(ctx context.Context, studentID string) (*StudentDetail, error) {
	if err := s.ensureAdvisee(ctx, studentID); err != nil {
		return nil, err
	}
	st, err := s.students.GetByID(ctx, tenant.SchoolIDFromContext(ctx), studentID)
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
	s.auditBestEffort(ctx, domain.AuditView, "student", st.ID, map[string]any{"action": "view", "by": "advisor"})
	return &StudentDetail{
		ID: st.ID, StudentCode: st.StudentCode, Status: st.Status, Prefix: st.Profile.Prefix,
		FirstName: st.Profile.FirstName, LastName: st.Profile.LastName,
		NationalIDMasked: maskNID(s.cipher, st.NationalIDEnc), BirthDate: birth,
		Phone: st.Profile.Phone, Address: addressDTO(st.Profile.Address), PhotoPath: st.PhotoPath,
		CreatedAt: st.CreatedAt.Format(time.RFC3339), UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateAdvisee แก้ข้อมูลนักเรียนในห้องที่ปรึกษาของผู้ใช้ (คงรหัสนักเรียน/สถานะ/เลขบัตรเดิมไว้)
func (s *MeService) UpdateAdvisee(ctx context.Context, studentID string, in UpdateAdviseeInput) error {
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return domain.ErrValidation
	}
	if err := s.ensureAdvisee(ctx, studentID); err != nil {
		return err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	st, err := s.students.GetByID(ctx, schoolID, studentID)
	if err != nil {
		return err
	}
	if st == nil {
		return domain.ErrStudentNotFound
	}
	up := domain.UpdateStudent{
		StudentCode: st.StudentCode, // คงค่าเดิม — ครูที่ปรึกษาแก้ไม่ได้
		Status:      st.Status,
		Profile:     toPersonProfile(in.Prefix, in.FirstName, in.LastName, in.BirthDate, in.Phone, in.Address),
	}
	audit := auditFor(ctx, domain.AuditUpdate, "student", studentID, map[string]any{
		"fields": []string{"name", "address", "contact"}, "by": "advisor",
	})
	found, err := s.students.Update(ctx, schoolID, studentID, up, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrStudentNotFound
	}
	return nil
}

func (s *MeService) auditBestEffort(ctx context.Context, action, targetType, targetID string, detail map[string]any) {
	if err := s.personnel.InsertAudit(ctx, auditFor(ctx, action, targetType, targetID, detail)); err != nil {
		log.Printf("me: บันทึก audit (%s %s) ล้มเหลว: %v", action, targetType, err)
	}
}

// meDetail สร้าง PersonnelDetail จาก domain (เลขบัตร mask เสมอ)
func meDetail(cipher Cipher, p *domain.Personnel) PersonnelDetail {
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
		NationalIDMasked:     maskNID(cipher, p.NationalIDEnc),
		CivilServantIDMasked: maskCivilTail(cipher, p.CivilServantIDEnc),
		BirthDate:            birth,
		Phone:                p.Profile.Phone,
		Email:                p.Profile.Email,
		Address:              addressDTO(p.Profile.Address),
		PhotoPath:            p.PhotoPath,
		CreatedAt:            p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            p.UpdatedAt.Format(time.RFC3339),
	}
}

// maskCivilTail เปิดเผยเฉพาะ 4 ตัวท้ายของเลขบัตรประจำตัวราชการ
func maskCivilTail(cipher Cipher, enc []byte) string {
	if len(enc) == 0 {
		return ""
	}
	plain, err := cipher.Decrypt(enc)
	if err != nil || len(plain) < 4 {
		return "xxxx"
	}
	return strings.Repeat("x", len(plain)-4) + plain[len(plain)-4:]
}
