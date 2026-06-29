package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"path"
	"strings"
	"time"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/storage"
	"github.com/chumko-platform/backend/internal/tenant"
)

// signedURLExpiry คืออายุของ signed URL ดาวน์โหลดไฟล์แนบ (สั้น ๆ ตาม PDPA)
const signedURLExpiry = 5 * time.Minute

// MaxWorkFileSize คือขนาดไฟล์แนบสูงสุด (10 MB)
const MaxWorkFileSize = 10 << 20

// PersonnelWorkRepository contract ของชั้น DB (ผลงานครู + ไฟล์แนบ)
type PersonnelWorkRepository interface {
	ListByPersonnel(ctx context.Context, schoolID, semesterID, personnelID string) ([]domain.PersonnelWork, error)
	GetByID(ctx context.Context, schoolID, personnelID, id string) (*domain.PersonnelWork, error)
	Create(ctx context.Context, schoolID, semesterID, personnelID string, nw domain.NewPersonnelWork, audit domain.AuditEntry) (string, error)
	Update(ctx context.Context, schoolID, personnelID, id string, uw domain.UpdatePersonnelWork, audit domain.AuditEntry) (bool, error)
	SoftDelete(ctx context.Context, schoolID, personnelID, id string, audit domain.AuditEntry) ([]string, bool, error)

	ListFiles(ctx context.Context, schoolID, workID string) ([]domain.PersonnelWorkFile, error)
	AddFile(ctx context.Context, schoolID, workID string, nf domain.NewPersonnelWorkFile, audit domain.AuditEntry) (string, error)
	GetFile(ctx context.Context, schoolID, workID, id string) (*domain.PersonnelWorkFile, error)
	SoftDeleteFile(ctx context.Context, schoolID, workID, id string, audit domain.AuditEntry) (string, bool, error)
}

// PersonnelWorkDTO ข้อมูลผลงานสำหรับ response
type PersonnelWorkDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	WorkDate    string `json:"work_date"`
	FileCount   int    `json:"file_count"`
	CreatedAt   string `json:"created_at"`
}

// WorkInput ข้อมูลสร้าง/แก้ผลงาน
type WorkInput struct {
	Title       string
	Description string
	WorkDate    *time.Time
}

// WorkFileDTO ข้อมูลไฟล์แนบสำหรับ response (URL = signed URL หมดอายุ)
type WorkFileDTO struct {
	ID           string `json:"id"`
	FileType     string `json:"file_type"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	SizeBytes    int64  `json:"size_bytes"`
	URL          string `json:"url"`
	CreatedAt    string `json:"created_at"`
}

// UploadFileInput ข้อมูลไฟล์ที่จะอัปโหลด (Reader = เนื้อไฟล์)
type UploadFileInput struct {
	FileType     string
	OriginalName string
	ContentType  string
	Size         int64
	Reader       io.Reader
}

// PersonnelWorkService จัดการผลงานครู (รายเทอม) + ไฟล์แนบ (signed URL)
type PersonnelWorkService struct {
	access  personnelAccess
	repo    PersonnelWorkRepository
	storage storage.Storage // อาจเป็น nil ถ้าไม่ได้ตั้งค่า storage
}

func NewPersonnelWorkService(repo PersonnelWorkRepository, guard personnelGuard, store storage.Storage) *PersonnelWorkService {
	return &PersonnelWorkService{access: personnelAccess{guard: guard}, repo: repo, storage: store}
}

func workToDTO(w domain.PersonnelWork) PersonnelWorkDTO {
	return PersonnelWorkDTO{
		ID:          w.ID,
		Title:       w.Title,
		Description: w.Description,
		WorkDate:    dateStr(w.WorkDate),
		FileCount:   w.FileCount,
		CreatedAt:   w.CreatedAt.Format(time.RFC3339),
	}
}

// List คืนผลงานของบุคลากรในเทอมปัจจุบัน
func (s *PersonnelWorkService) List(ctx context.Context, personnelID string) ([]PersonnelWorkDTO, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return nil, err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByPersonnel(ctx, tenant.SchoolIDFromContext(ctx), sem, personnelID)
	if err != nil {
		return nil, err
	}
	out := make([]PersonnelWorkDTO, 0, len(rows))
	for i := range rows {
		out = append(out, workToDTO(rows[i]))
	}
	return out, nil
}

// Create เพิ่มผลงานในเทอมปัจจุบัน
func (s *PersonnelWorkService) Create(ctx context.Context, personnelID string, in WorkInput) (string, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return "", err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return "", err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return "", domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditCreate, "personnel_work", "", map[string]any{"personnel_id": personnelID})
	return s.repo.Create(ctx, tenant.SchoolIDFromContext(ctx), sem, personnelID, domain.NewPersonnelWork{
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		WorkDate:    in.WorkDate,
	}, audit)
}

// Update แก้ไขผลงาน
func (s *PersonnelWorkService) Update(ctx context.Context, personnelID, workID string, in WorkInput) error {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditUpdate, "personnel_work", workID, map[string]any{"personnel_id": personnelID})
	found, err := s.repo.Update(ctx, tenant.SchoolIDFromContext(ctx), personnelID, workID, domain.UpdatePersonnelWork{
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		WorkDate:    in.WorkDate,
	}, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrWorkNotFound
	}
	return nil
}

// Delete ลบผลงาน + ไฟล์แนบทั้งหมด (ลบ object ออกจาก storage แบบ best-effort)
func (s *PersonnelWorkService) Delete(ctx context.Context, personnelID, workID string) error {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "personnel_work", workID, map[string]any{"personnel_id": personnelID})
	paths, found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), personnelID, workID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrWorkNotFound
	}
	s.removeObjects(ctx, paths)
	return nil
}

// ensureWork ยืนยันว่าผลงานมีอยู่จริง (scope school_id + personnel_id) ก่อนแตะไฟล์
func (s *PersonnelWorkService) ensureWork(ctx context.Context, personnelID, workID string) error {
	w, err := s.repo.GetByID(ctx, tenant.SchoolIDFromContext(ctx), personnelID, workID)
	if err != nil {
		return err
	}
	if w == nil {
		return domain.ErrWorkNotFound
	}
	return nil
}

// ListFiles คืนไฟล์แนบของผลงาน พร้อม signed URL ดาวน์โหลด
func (s *PersonnelWorkService) ListFiles(ctx context.Context, personnelID, workID string) ([]WorkFileDTO, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return nil, err
	}
	if s.storage == nil {
		return nil, domain.ErrStorageUnavailable
	}
	if err := s.ensureWork(ctx, personnelID, workID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListFiles(ctx, tenant.SchoolIDFromContext(ctx), workID)
	if err != nil {
		return nil, err
	}
	out := make([]WorkFileDTO, 0, len(rows))
	for i := range rows {
		url, err := s.storage.PresignGet(ctx, rows[i].StoragePath, rows[i].OriginalName, signedURLExpiry)
		if err != nil {
			return nil, err
		}
		out = append(out, WorkFileDTO{
			ID:           rows[i].ID,
			FileType:     rows[i].FileType,
			OriginalName: rows[i].OriginalName,
			ContentType:  rows[i].ContentType,
			SizeBytes:    rows[i].SizeBytes,
			URL:          url,
			CreatedAt:    rows[i].CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

// UploadFile อัปโหลดไฟล์แนบขึ้น storage แล้วบันทึก metadata
func (s *PersonnelWorkService) UploadFile(ctx context.Context, personnelID, workID string, in UploadFileInput) (string, error) {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return "", err
	}
	if s.storage == nil {
		return "", domain.ErrStorageUnavailable
	}
	if !domain.ValidWorkFileType(in.FileType) {
		return "", domain.ErrInvalidFileType
	}
	if in.Size <= 0 {
		return "", domain.ErrFileRequired
	}
	if in.Size > MaxWorkFileSize {
		return "", domain.ErrFileTooLarge
	}
	if err := s.ensureWork(ctx, personnelID, workID); err != nil {
		return "", err
	}

	schoolID := tenant.SchoolIDFromContext(ctx)
	objectPath := buildObjectPath(schoolID, personnelID, workID, in.OriginalName)

	if err := s.storage.Put(ctx, objectPath, in.Reader, in.Size, in.ContentType); err != nil {
		return "", err
	}

	audit := auditFor(ctx, domain.AuditCreate, "personnel_work_file", "", map[string]any{"work_id": workID, "file_type": in.FileType})
	id, err := s.repo.AddFile(ctx, schoolID, workID, domain.NewPersonnelWorkFile{
		FileType:     in.FileType,
		StoragePath:  objectPath,
		OriginalName: in.OriginalName,
		ContentType:  in.ContentType,
		SizeBytes:    in.Size,
	}, audit)
	if err != nil {
		// metadata ล้มเหลว → ลบ object ที่เพิ่งอัปโหลดทิ้ง (กัน orphan)
		_ = s.storage.Remove(ctx, objectPath)
		return "", err
	}
	return id, nil
}

// DeleteFile ลบไฟล์แนบ (DB เป็น truth, ลบ object จาก storage แบบ best-effort)
func (s *PersonnelWorkService) DeleteFile(ctx context.Context, personnelID, workID, fileID string) error {
	if err := s.access.authorize(ctx, personnelID); err != nil {
		return err
	}
	if err := s.ensureWork(ctx, personnelID, workID); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "personnel_work_file", fileID, map[string]any{"work_id": workID})
	path, found, err := s.repo.SoftDeleteFile(ctx, tenant.SchoolIDFromContext(ctx), workID, fileID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrWorkFileNotFound
	}
	s.removeObjects(ctx, []string{path})
	return nil
}

// removeObjects ลบ object ออกจาก storage แบบ best-effort (DB ลบไปแล้วถือเป็น truth)
func (s *PersonnelWorkService) removeObjects(ctx context.Context, paths []string) {
	if s.storage == nil {
		return
	}
	for _, p := range paths {
		if p == "" {
			continue
		}
		_ = s.storage.Remove(ctx, p)
	}
}

// buildObjectPath สร้าง path ของ object ใน storage แบบ scope ตามโรงเรียน/บุคลากร/ผลงาน
// มี token สุ่มกันชื่อชนกัน + sanitize ชื่อไฟล์
func buildObjectPath(schoolID, personnelID, workID, originalName string) string {
	token := randomToken()
	name := sanitizeFileName(originalName)
	if name == "" {
		name = "file"
	}
	return path.Join("schools", schoolID, "personnel", personnelID, "works", workID, token+"-"+name)
}

func sanitizeFileName(name string) string {
	name = path.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	// เอาเฉพาะ base name; กัน path traversal
	if name == "." || name == ".." || name == "/" {
		return ""
	}
	return name
}

func randomToken() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// fallback ที่ยังใช้งานได้ (ความน่าจะชนต่ำมากในทางปฏิบัติ)
		return "t"
	}
	return hex.EncodeToString(b)
}
