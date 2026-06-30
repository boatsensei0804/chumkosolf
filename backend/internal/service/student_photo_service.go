package service

import (
	"context"
	"io"
	"path"
	"strings"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/storage"
	"github.com/chumkosoft/backend/internal/tenant"
)

// studentPhotoURLExpiry คืออายุของ signed URL รูปนักเรียน (ใช้แสดง/ดึงไปทำ dataset สแกนหน้า)
const studentPhotoURLExpiry = 15 * time.Minute

// maxStudentPhotoSize คือขนาดรูปสูงสุด (5 MB)
const maxStudentPhotoSize = 5 << 20

// maxStudentPhotos คือจำนวนรูปสูงสุดต่อนักเรียน (หลายรูปเพื่อความแม่นยำของการสแกนหน้า)
const maxStudentPhotos = 10

// allowedImageTypes map content-type → นามสกุลไฟล์ที่อนุญาต
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// StudentPhotoStore — เข้าถึงตาราง student_photos
type StudentPhotoStore interface {
	CountByStudent(ctx context.Context, schoolID, studentID string) (int, error)
	ListByStudent(ctx context.Context, schoolID, studentID string) ([]domain.StudentPhoto, error)
	Get(ctx context.Context, schoolID, photoID string) (*domain.StudentPhoto, error)
	Add(ctx context.Context, schoolID, studentID string, np domain.NewStudentPhoto, makePrimary bool, audit domain.AuditEntry) (string, error)
	SetPrimary(ctx context.Context, schoolID, studentID, photoID string, audit domain.AuditEntry) (bool, error)
	SoftDelete(ctx context.Context, schoolID, studentID, photoID string, audit domain.AuditEntry) (string, bool, error)
	Dataset(ctx context.Context, schoolID, semesterID, classID string) ([]domain.StudentPhotoRow, error)
}

// StudentExistsRepo ใช้ตรวจว่านักเรียนมีอยู่จริงในโรงเรียน
type StudentExistsRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Student, error)
}

// PhotoInput ข้อมูลรูปที่อัปโหลด (Reader = เนื้อไฟล์)
type PhotoInput struct {
	OriginalName string
	ContentType  string
	Size         int64
	Reader       io.Reader
}

// PhotoDTO รูปนักเรียน 1 รายการ (URL = signed URL หมดอายุ)
type PhotoDTO struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	CreatedAt string `json:"created_at"`
}

// FacePhotoDTO รูป 1 ใบใน dataset (signed URL หมดอายุ)
type FacePhotoDTO struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
}

// FaceDatasetStudentDTO นักเรียน 1 คน + รูปทั้งหมด สำหรับ enroll เข้าระบบสแกนหน้า
type FaceDatasetStudentDTO struct {
	StudentID   string         `json:"student_id"`
	StudentCode string         `json:"student_code"`
	Prefix      string         `json:"prefix"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	ClassLabel  string         `json:"class_label"`
	Photos      []FacePhotoDTO `json:"photos"`
}

// StudentPhotoService จัดการรูปนักเรียน (หลายรูป + เลือกรูปโปรไฟล์) สำหรับระบบสแกนหน้าเข้าเรียน
// PDPA: รูปเด็กเข้าถึงผ่าน signed URL หมดอายุเท่านั้น; จำกัดสิทธิ์กลุ่มวิชาการ (academicGuard)
type StudentPhotoService struct {
	guard    academicGuard
	photos   StudentPhotoStore
	students StudentExistsRepo
	storage  storage.Storage // อาจเป็น nil ถ้าไม่ได้ตั้งค่า storage
}

func NewStudentPhotoService(photos StudentPhotoStore, students StudentExistsRepo, checker WorkGroupChecker, store storage.Storage) *StudentPhotoService {
	return &StudentPhotoService{guard: academicGuard{checker: checker}, photos: photos, students: students, storage: store}
}

func (s *StudentPhotoService) ensureStudent(ctx context.Context, studentID string) error {
	st, err := s.students.GetByID(ctx, tenant.SchoolIDFromContext(ctx), studentID)
	if err != nil {
		return err
	}
	if st == nil {
		return domain.ErrStudentNotFound
	}
	return nil
}

func (s *StudentPhotoService) toDTO(ctx context.Context, p *domain.StudentPhoto) (PhotoDTO, error) {
	url, err := s.storage.PresignGet(ctx, p.StoragePath, "", studentPhotoURLExpiry)
	if err != nil {
		return PhotoDTO{}, err
	}
	return PhotoDTO{ID: p.ID, URL: url, IsPrimary: p.IsPrimary, CreatedAt: p.CreatedAt.Format(time.RFC3339)}, nil
}

// List คืนรูปทั้งหมดของนักเรียน พร้อม signed URL (รูปโปรไฟล์มาก่อน)
func (s *StudentPhotoService) List(ctx context.Context, studentID string) ([]PhotoDTO, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	if err := s.ensureStudent(ctx, studentID); err != nil {
		return nil, err
	}
	if s.storage == nil {
		return nil, domain.ErrStorageUnavailable
	}
	rows, err := s.photos.ListByStudent(ctx, tenant.SchoolIDFromContext(ctx), studentID)
	if err != nil {
		return nil, err
	}
	out := make([]PhotoDTO, 0, len(rows))
	for i := range rows {
		dto, err := s.toDTO(ctx, &rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, nil
}

// Upload เพิ่มรูปนักเรียน 1 รูป (รูปแรกจะถูกตั้งเป็นรูปโปรไฟล์อัตโนมัติ)
func (s *StudentPhotoService) Upload(ctx context.Context, studentID string, in PhotoInput) (PhotoDTO, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return PhotoDTO{}, err
	}
	if s.storage == nil {
		return PhotoDTO{}, domain.ErrStorageUnavailable
	}
	ext, ok := allowedImageTypes[in.ContentType]
	if !ok {
		return PhotoDTO{}, domain.ErrInvalidImageType
	}
	if in.Size <= 0 {
		return PhotoDTO{}, domain.ErrFileRequired
	}
	if in.Size > maxStudentPhotoSize {
		return PhotoDTO{}, domain.ErrFileTooLarge
	}
	if err := s.ensureStudent(ctx, studentID); err != nil {
		return PhotoDTO{}, err
	}

	schoolID := tenant.SchoolIDFromContext(ctx)
	count, err := s.photos.CountByStudent(ctx, schoolID, studentID)
	if err != nil {
		return PhotoDTO{}, err
	}
	if count >= maxStudentPhotos {
		return PhotoDTO{}, domain.ErrPhotoLimitReached
	}

	objectPath := path.Join("schools", schoolID, "students", studentID, "photos", randomToken()+ext)
	if err := s.storage.Put(ctx, objectPath, in.Reader, in.Size, in.ContentType); err != nil {
		return PhotoDTO{}, err
	}

	makePrimary := count == 0 // รูปแรก = รูปโปรไฟล์
	audit := auditFor(ctx, domain.AuditUpdate, "student_photo", studentID, map[string]any{"action": "upload", "primary": makePrimary})
	id, err := s.photos.Add(ctx, schoolID, studentID, domain.NewStudentPhoto{
		StoragePath: objectPath, ContentType: in.ContentType, SizeBytes: in.Size,
	}, makePrimary, audit)
	if err != nil {
		_ = s.storage.Remove(ctx, objectPath) // กัน orphan
		return PhotoDTO{}, err
	}

	url, err := s.storage.PresignGet(ctx, objectPath, "", studentPhotoURLExpiry)
	if err != nil {
		return PhotoDTO{}, err
	}
	return PhotoDTO{ID: id, URL: url, IsPrimary: makePrimary, CreatedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

// Dataset คืนนักเรียน + รูปทั้งหมด (signed URL) สำหรับ build dataset สแกนหน้า
// classID ว่าง = ทั้งโรงเรียน (เฉพาะคนที่มีรูป); ระบุ = เฉพาะห้องนั้น (ต้องมีเทอม)
func (s *StudentPhotoService) Dataset(ctx context.Context, classID string) ([]FaceDatasetStudentDTO, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	if s.storage == nil {
		return nil, domain.ErrStorageUnavailable
	}
	sem := tenant.SemesterIDFromContext(ctx)
	if classID != "" && sem == "" {
		return nil, domain.ErrNoActiveSemester
	}

	rows, err := s.photos.Dataset(ctx, tenant.SchoolIDFromContext(ctx), sem, classID)
	if err != nil {
		return nil, err
	}

	out := make([]FaceDatasetStudentDTO, 0)
	idx := make(map[string]int) // student_id → ตำแหน่งใน out
	for i := range rows {
		r := &rows[i]
		pos, ok := idx[r.StudentID]
		if !ok {
			pos = len(out)
			idx[r.StudentID] = pos
			out = append(out, FaceDatasetStudentDTO{
				StudentID: r.StudentID, StudentCode: r.StudentCode, Prefix: r.Prefix,
				FirstName: r.FirstName, LastName: r.LastName,
				ClassLabel: strings.TrimSpace(r.GradeLevel + " " + r.RoomName),
				Photos:     []FacePhotoDTO{},
			})
		}
		url, err := s.storage.PresignGet(ctx, r.StoragePath, "", studentPhotoURLExpiry)
		if err != nil {
			return nil, err
		}
		out[pos].Photos = append(out[pos].Photos, FacePhotoDTO{ID: r.PhotoID, URL: url, IsPrimary: r.IsPrimary})
	}
	return out, nil
}

// SetPrimary ตั้งรูปที่เลือกเป็นรูปโปรไฟล์
func (s *StudentPhotoService) SetPrimary(ctx context.Context, studentID, photoID string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditUpdate, "student_photo", studentID, map[string]any{"action": "set_primary", "photo_id": photoID})
	found, err := s.photos.SetPrimary(ctx, tenant.SchoolIDFromContext(ctx), studentID, photoID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrStudentPhotoNotFound
	}
	return nil
}

// Delete ลบรูป 1 รูป (เคลียร์ object ใน storage แบบ best-effort; เลื่อนรูปโปรไฟล์ใหม่ถ้าจำเป็น)
func (s *StudentPhotoService) Delete(ctx context.Context, studentID, photoID string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditUpdate, "student_photo", studentID, map[string]any{"action": "delete", "photo_id": photoID})
	storagePath, found, err := s.photos.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), studentID, photoID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrStudentPhotoNotFound
	}
	if s.storage != nil && storagePath != "" {
		_ = s.storage.Remove(ctx, storagePath)
	}
	return nil
}
