package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
)

// --- fakes ---

type fakePhotoStore struct {
	photos       []domain.StudentPhoto
	addedPrimary bool
	addedPath    string
	notFound     bool   // ให้ SetPrimary/SoftDelete คืน found=false
	deletedPath  string // path ที่ SoftDelete คืน
	datasetRows  []domain.StudentPhotoRow
}

func (s *fakePhotoStore) CountByStudent(_ context.Context, _, _ string) (int, error) {
	return len(s.photos), nil
}
func (s *fakePhotoStore) ListByStudent(_ context.Context, _, _ string) ([]domain.StudentPhoto, error) {
	return s.photos, nil
}
func (s *fakePhotoStore) Get(_ context.Context, _, photoID string) (*domain.StudentPhoto, error) {
	for i := range s.photos {
		if s.photos[i].ID == photoID {
			cp := s.photos[i]
			return &cp, nil
		}
	}
	return nil, nil
}
func (s *fakePhotoStore) Add(_ context.Context, _, _ string, np domain.NewStudentPhoto, makePrimary bool, _ domain.AuditEntry) (string, error) {
	s.addedPrimary = makePrimary
	s.addedPath = np.StoragePath
	s.photos = append(s.photos, domain.StudentPhoto{ID: "new", StoragePath: np.StoragePath, IsPrimary: makePrimary})
	return "new", nil
}
func (s *fakePhotoStore) SetPrimary(_ context.Context, _, _, _ string, _ domain.AuditEntry) (bool, error) {
	return !s.notFound, nil
}
func (s *fakePhotoStore) SoftDelete(_ context.Context, _, _, _ string, _ domain.AuditEntry) (string, bool, error) {
	if s.notFound {
		return "", false, nil
	}
	return s.deletedPath, true, nil
}
func (s *fakePhotoStore) Dataset(_ context.Context, _, _, _ string) ([]domain.StudentPhotoRow, error) {
	return s.datasetRows, nil
}

type fakeStudentExists struct {
	schoolID string
	id       string
}

func (r *fakeStudentExists) GetByID(_ context.Context, schoolID, id string) (*domain.Student, error) {
	if r.schoolID == schoolID && r.id == id {
		return &domain.Student{ID: id, SchoolID: schoolID}, nil
	}
	return nil, nil
}

func existsRepo() *fakeStudentExists { return &fakeStudentExists{schoolID: "school-A", id: "s1"} }

func jpegInput(size int64) PhotoInput {
	return PhotoInput{OriginalName: "face.jpg", ContentType: "image/jpeg", Size: size, Reader: strings.NewReader("img-bytes")}
}

// --- tests ---

func TestStudentPhoto_UploadForbidden(t *testing.T) {
	svc := NewStudentPhotoService(&fakePhotoStore{}, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	if _, err := svc.Upload(memberCtx("school-A", "u9"), "s1", jpegInput(100)); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestStudentPhoto_UploadInvalidType(t *testing.T) {
	svc := NewStudentPhotoService(&fakePhotoStore{}, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	in := jpegInput(100)
	in.ContentType = "application/pdf"
	if _, err := svc.Upload(adminCtx("school-A"), "s1", in); !errors.Is(err, domain.ErrInvalidImageType) {
		t.Errorf("err = %v, want ErrInvalidImageType", err)
	}
}

func TestStudentPhoto_UploadTooLarge(t *testing.T) {
	svc := NewStudentPhotoService(&fakePhotoStore{}, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	if _, err := svc.Upload(adminCtx("school-A"), "s1", jpegInput(maxStudentPhotoSize+1)); !errors.Is(err, domain.ErrFileTooLarge) {
		t.Errorf("err = %v, want ErrFileTooLarge", err)
	}
}

func TestStudentPhoto_UploadStorageUnavailable(t *testing.T) {
	svc := NewStudentPhotoService(&fakePhotoStore{}, existsRepo(), &fakeWGChecker{}, nil)
	if _, err := svc.Upload(adminCtx("school-A"), "s1", jpegInput(100)); !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Errorf("err = %v, want ErrStorageUnavailable", err)
	}
}

func TestStudentPhoto_UploadFirstBecomesPrimary(t *testing.T) {
	store := &fakePhotoStore{}
	svc := NewStudentPhotoService(store, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	dto, err := svc.Upload(adminCtx("school-A"), "s1", jpegInput(100))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !store.addedPrimary || !dto.IsPrimary {
		t.Error("รูปแรกต้องเป็นรูปโปรไฟล์อัตโนมัติ")
	}
	if !strings.HasPrefix(dto.URL, "https://signed.example/") {
		t.Errorf("url = %q ควรเป็น signed URL", dto.URL)
	}
}

func TestStudentPhoto_UploadSecondNotPrimary(t *testing.T) {
	store := &fakePhotoStore{photos: []domain.StudentPhoto{{ID: "p1", IsPrimary: true}}}
	svc := NewStudentPhotoService(store, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	dto, err := svc.Upload(adminCtx("school-A"), "s1", jpegInput(100))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if store.addedPrimary || dto.IsPrimary {
		t.Error("รูปที่ 2 ต้องไม่ใช่รูปโปรไฟล์โดยอัตโนมัติ")
	}
}

func TestStudentPhoto_UploadLimitReached(t *testing.T) {
	full := make([]domain.StudentPhoto, maxStudentPhotos)
	svc := NewStudentPhotoService(&fakePhotoStore{photos: full}, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	if _, err := svc.Upload(adminCtx("school-A"), "s1", jpegInput(100)); !errors.Is(err, domain.ErrPhotoLimitReached) {
		t.Errorf("err = %v, want ErrPhotoLimitReached", err)
	}
}

func TestStudentPhoto_UploadStudentNotFound(t *testing.T) {
	svc := NewStudentPhotoService(&fakePhotoStore{}, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	if _, err := svc.Upload(adminCtx("school-A"), "ghost", jpegInput(100)); !errors.Is(err, domain.ErrStudentNotFound) {
		t.Errorf("err = %v, want ErrStudentNotFound", err)
	}
}

func TestStudentPhoto_ListSignedURLs(t *testing.T) {
	store := &fakePhotoStore{photos: []domain.StudentPhoto{
		{ID: "p1", StoragePath: "schools/school-A/students/s1/photos/a.jpg", IsPrimary: true},
		{ID: "p2", StoragePath: "schools/school-A/students/s1/photos/b.jpg"},
	}}
	svc := NewStudentPhotoService(store, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	out, err := svc.List(adminCtx("school-A"), "s1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(out) != 2 || !out[0].IsPrimary || out[0].URL == "" {
		t.Errorf("list = %+v", out)
	}
}

func TestStudentPhoto_SetPrimaryNotFound(t *testing.T) {
	svc := NewStudentPhotoService(&fakePhotoStore{notFound: true}, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	if err := svc.SetPrimary(adminCtx("school-A"), "s1", "ghost"); !errors.Is(err, domain.ErrStudentPhotoNotFound) {
		t.Errorf("err = %v, want ErrStudentPhotoNotFound", err)
	}
}

func TestStudentPhoto_DeleteNotFound(t *testing.T) {
	svc := NewStudentPhotoService(&fakePhotoStore{notFound: true}, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	if err := svc.Delete(adminCtx("school-A"), "s1", "ghost"); !errors.Is(err, domain.ErrStudentPhotoNotFound) {
		t.Errorf("err = %v, want ErrStudentPhotoNotFound", err)
	}
}

func TestStudentPhoto_DatasetForbidden(t *testing.T) {
	svc := NewStudentPhotoService(&fakePhotoStore{}, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	if _, err := svc.Dataset(memberCtx("school-A", "u9"), ""); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestStudentPhoto_DatasetGroupsAndPresigns(t *testing.T) {
	store := &fakePhotoStore{datasetRows: []domain.StudentPhotoRow{
		{StudentID: "s1", StudentCode: "S001", FirstName: "ก", GradeLevel: "ม.1", RoomName: "1", PhotoID: "p1", StoragePath: "a.jpg", IsPrimary: true},
		{StudentID: "s1", StudentCode: "S001", FirstName: "ก", GradeLevel: "ม.1", RoomName: "1", PhotoID: "p2", StoragePath: "b.jpg"},
		{StudentID: "s2", StudentCode: "S002", FirstName: "ข", PhotoID: "p3", StoragePath: "c.jpg", IsPrimary: true},
	}}
	svc := NewStudentPhotoService(store, existsRepo(), &fakeWGChecker{}, newFakeStorage())
	out, err := svc.Dataset(adminCtx("school-A"), "")
	if err != nil {
		t.Fatalf("dataset: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("students = %d, want 2", len(out))
	}
	if len(out[0].Photos) != 2 || out[0].ClassLabel != "ม.1 1" {
		t.Errorf("s1 = %+v", out[0])
	}
	if out[0].Photos[0].URL == "" {
		t.Error("ต้องมี signed URL")
	}
	if len(out[1].Photos) != 1 {
		t.Errorf("s2 photos = %d, want 1", len(out[1].Photos))
	}
}

func TestStudentPhoto_DeleteRemovesObject(t *testing.T) {
	store := &fakePhotoStore{deletedPath: "schools/school-A/students/s1/photos/x.jpg"}
	st := newFakeStorage()
	st.objects[store.deletedPath] = []byte("x")
	svc := NewStudentPhotoService(store, existsRepo(), &fakeWGChecker{}, st)
	if err := svc.Delete(adminCtx("school-A"), "s1", "p1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, exists := st.objects[store.deletedPath]; exists {
		t.Error("object ควรถูกลบจาก storage")
	}
}
