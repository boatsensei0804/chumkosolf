package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// --- fake work repo (in-memory) ---

type fakeWorkRepo struct {
	works map[string]*domain.PersonnelWork
	files map[string]*domain.PersonnelWorkFile
	seq   int
}

func newFakeWorkRepo() *fakeWorkRepo {
	return &fakeWorkRepo{works: map[string]*domain.PersonnelWork{}, files: map[string]*domain.PersonnelWorkFile{}}
}

func (r *fakeWorkRepo) ListByPersonnel(_ context.Context, schoolID, semesterID, personnelID string) ([]domain.PersonnelWork, error) {
	var out []domain.PersonnelWork
	for _, w := range r.works {
		if w.SchoolID == schoolID && w.SemesterID == semesterID && w.PersonnelID == personnelID {
			cp := *w
			cp.FileCount = r.countFiles(w.ID)
			out = append(out, cp)
		}
	}
	return out, nil
}

func (r *fakeWorkRepo) countFiles(workID string) int {
	n := 0
	for _, f := range r.files {
		if f.PersonnelWorkID == workID {
			n++
		}
	}
	return n
}

func (r *fakeWorkRepo) GetByID(_ context.Context, schoolID, personnelID, id string) (*domain.PersonnelWork, error) {
	w, ok := r.works[id]
	if !ok || w.SchoolID != schoolID || w.PersonnelID != personnelID {
		return nil, nil
	}
	cp := *w
	return &cp, nil
}

func (r *fakeWorkRepo) Create(_ context.Context, schoolID, semesterID, personnelID string, nw domain.NewPersonnelWork, _ domain.AuditEntry) (string, error) {
	r.seq++
	id := "w" + string(rune('0'+r.seq))
	r.works[id] = &domain.PersonnelWork{
		ID: id, SchoolID: schoolID, SemesterID: semesterID, PersonnelID: personnelID,
		Title: nw.Title, Description: nw.Description, WorkDate: nw.WorkDate,
	}
	return id, nil
}

func (r *fakeWorkRepo) Update(_ context.Context, schoolID, personnelID, id string, uw domain.UpdatePersonnelWork, _ domain.AuditEntry) (bool, error) {
	w, ok := r.works[id]
	if !ok || w.SchoolID != schoolID || w.PersonnelID != personnelID {
		return false, nil
	}
	w.Title = uw.Title
	w.Description = uw.Description
	w.WorkDate = uw.WorkDate
	return true, nil
}

func (r *fakeWorkRepo) SoftDelete(_ context.Context, schoolID, personnelID, id string, _ domain.AuditEntry) ([]string, bool, error) {
	w, ok := r.works[id]
	if !ok || w.SchoolID != schoolID || w.PersonnelID != personnelID {
		return nil, false, nil
	}
	delete(r.works, id)
	var paths []string
	for fid, f := range r.files {
		if f.PersonnelWorkID == id && f.SchoolID == schoolID {
			paths = append(paths, f.StoragePath)
			delete(r.files, fid)
		}
	}
	return paths, true, nil
}

func (r *fakeWorkRepo) ListFiles(_ context.Context, schoolID, workID string) ([]domain.PersonnelWorkFile, error) {
	var out []domain.PersonnelWorkFile
	for _, f := range r.files {
		if f.SchoolID == schoolID && f.PersonnelWorkID == workID {
			out = append(out, *f)
		}
	}
	return out, nil
}

func (r *fakeWorkRepo) AddFile(_ context.Context, schoolID, workID string, nf domain.NewPersonnelWorkFile, _ domain.AuditEntry) (string, error) {
	r.seq++
	id := "f" + string(rune('0'+r.seq))
	r.files[id] = &domain.PersonnelWorkFile{
		ID: id, SchoolID: schoolID, PersonnelWorkID: workID,
		FileType: nf.FileType, StoragePath: nf.StoragePath, OriginalName: nf.OriginalName,
		ContentType: nf.ContentType, SizeBytes: nf.SizeBytes,
	}
	return id, nil
}

func (r *fakeWorkRepo) GetFile(_ context.Context, schoolID, workID, id string) (*domain.PersonnelWorkFile, error) {
	f, ok := r.files[id]
	if !ok || f.SchoolID != schoolID || f.PersonnelWorkID != workID {
		return nil, nil
	}
	cp := *f
	return &cp, nil
}

func (r *fakeWorkRepo) SoftDeleteFile(_ context.Context, schoolID, workID, id string, _ domain.AuditEntry) (string, bool, error) {
	f, ok := r.files[id]
	if !ok || f.SchoolID != schoolID || f.PersonnelWorkID != workID {
		return "", false, nil
	}
	delete(r.files, id)
	return f.StoragePath, true, nil
}

// --- fake storage ---

type fakeStorage struct {
	objects map[string][]byte
	removed []string
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{objects: map[string][]byte{}}
}

func (s *fakeStorage) Put(_ context.Context, objectPath string, r io.Reader, _ int64, _ string) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s.objects[objectPath] = b
	return nil
}

func (s *fakeStorage) Get(_ context.Context, objectPath string) ([]byte, error) {
	b, ok := s.objects[objectPath]
	if !ok {
		return nil, fmt.Errorf("not found: %s", objectPath)
	}
	return b, nil
}

func (s *fakeStorage) PresignGet(_ context.Context, objectPath, _ string, _ time.Duration) (string, error) {
	return "https://signed.example/" + objectPath, nil
}

func (s *fakeStorage) Remove(_ context.Context, objectPath string) error {
	delete(s.objects, objectPath)
	s.removed = append(s.removed, objectPath)
	return nil
}

// --- ctx helpers (ผลงานครูเป็นข้อมูลรายเทอม จึงต้องมี SemesterID) ---

const workSchool = "school-A"
const workPersonnel = "person-1"

func wAdmin(schoolID, sem string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: "admin", SchoolID: schoolID, Role: "super_admin", IsSchoolAdmin: true, SemesterID: sem,
	})
}

func wMember(schoolID, userID, sem string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: userID, SchoolID: schoolID, Role: "teacher", SemesterID: sem,
	})
}

func newWorkSvc() (*PersonnelWorkService, *fakeWorkRepo, *fakeStorage, *fakePersonnelRepo) {
	repo := newFakeWorkRepo()
	store := newFakeStorage()
	guard := guardWith(workSchool, workPersonnel)
	return NewPersonnelWorkService(repo, guard, store), repo, store, guard
}

func uploadInput(t io.Reader, size int64, fileType string) UploadFileInput {
	return UploadFileInput{FileType: fileType, OriginalName: "ผลงาน รางวัล.pdf", ContentType: "application/pdf", Size: size, Reader: t}
}

// --- work tests ---

func TestWork_CreateSuccess(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	id, err := svc.Create(wAdmin(workSchool, "sem-1"), workPersonnel, WorkInput{Title: "รางวัลครูดีเด่น"})
	if err != nil || id == "" {
		t.Fatalf("create: id=%q err=%v", id, err)
	}
}

func TestWork_CreateRequiresTitle(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	_, err := svc.Create(wAdmin(workSchool, "sem-1"), workPersonnel, WorkInput{Title: "   "})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestWork_CreateNoSemester(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	// ctx ของ admin แต่ไม่มี SemesterID
	_, err := svc.Create(adminCtx(workSchool), workPersonnel, WorkInput{Title: "ผลงาน"})
	if !errors.Is(err, domain.ErrNoActiveSemester) {
		t.Errorf("err = %v, want ErrNoActiveSemester", err)
	}
}

func TestWork_ForbiddenForNonMember(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	// teacher ที่ไม่ได้สังกัดกลุ่มบุคคล
	_, err := svc.Create(wMember(workSchool, "u9", "sem-1"), workPersonnel, WorkInput{Title: "ผลงาน"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestWork_MemberInPersonnelGroupCanCreate(t *testing.T) {
	svc, _, _, guard := newWorkSvc()
	guard.groups[workSchool+"|u9|personnel"] = true
	id, err := svc.Create(wMember(workSchool, "u9", "sem-1"), workPersonnel, WorkInput{Title: "ผลงาน"})
	if err != nil || id == "" {
		t.Fatalf("create by group member: id=%q err=%v", id, err)
	}
}

func TestWork_CrossSchoolPersonnelNotFound(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	// personnel อยู่ school-A แต่เรียกด้วย scope school-B → ไม่พบบุคลากร
	_, err := svc.Create(wAdmin("school-B", "sem-1"), workPersonnel, WorkInput{Title: "ผลงาน"})
	if !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("err = %v, want ErrPersonnelNotFound", err)
	}
}

func TestWork_ListIsolatedBySemester(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	if _, err := svc.Create(wAdmin(workSchool, "sem-1"), workPersonnel, WorkInput{Title: "ผลงานเทอม 1"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	// list เทอม 2 ต้องไม่เห็นผลงานของเทอม 1
	list, err := svc.List(wAdmin(workSchool, "sem-2"), workPersonnel)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("เทอม 2 เห็นผลงาน %d รายการ ควรเป็น 0 (ข้อมูลปนข้ามเทอม)", len(list))
	}
}

func TestWork_UpdateNotFound(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	err := svc.Update(wAdmin(workSchool, "sem-1"), workPersonnel, "missing", WorkInput{Title: "x"})
	if !errors.Is(err, domain.ErrWorkNotFound) {
		t.Errorf("err = %v, want ErrWorkNotFound", err)
	}
}

func TestWork_DeleteRemovesAttachedFiles(t *testing.T) {
	svc, repo, store, _ := newWorkSvc()
	ctx := wAdmin(workSchool, "sem-1")
	id, _ := svc.Create(ctx, workPersonnel, WorkInput{Title: "ผลงาน"})
	if _, err := svc.UploadFile(ctx, workPersonnel, id, uploadInput(bytes.NewReader([]byte("hello")), 5, domain.WorkFileDocument)); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if err := svc.Delete(ctx, workPersonnel, id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(repo.files) != 0 {
		t.Errorf("ไฟล์แนบเหลือ %d รายการ ควรถูกลบหมด", len(repo.files))
	}
	if len(store.removed) != 1 {
		t.Errorf("ลบ object จาก storage %d ครั้ง ควรเป็น 1", len(store.removed))
	}
}

// --- file tests ---

func TestWorkFile_UploadSuccess(t *testing.T) {
	svc, _, store, _ := newWorkSvc()
	ctx := wAdmin(workSchool, "sem-1")
	id, _ := svc.Create(ctx, workPersonnel, WorkInput{Title: "ผลงาน"})
	fid, err := svc.UploadFile(ctx, workPersonnel, id, uploadInput(bytes.NewReader([]byte("data")), 4, domain.WorkFileImage))
	if err != nil || fid == "" {
		t.Fatalf("upload: fid=%q err=%v", fid, err)
	}
	if len(store.objects) != 1 {
		t.Errorf("storage มี object %d ควรเป็น 1", len(store.objects))
	}
}

func TestWorkFile_UploadInvalidType(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	ctx := wAdmin(workSchool, "sem-1")
	id, _ := svc.Create(ctx, workPersonnel, WorkInput{Title: "ผลงาน"})
	_, err := svc.UploadFile(ctx, workPersonnel, id, uploadInput(bytes.NewReader([]byte("x")), 1, "video"))
	if !errors.Is(err, domain.ErrInvalidFileType) {
		t.Errorf("err = %v, want ErrInvalidFileType", err)
	}
}

func TestWorkFile_UploadTooLarge(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	ctx := wAdmin(workSchool, "sem-1")
	id, _ := svc.Create(ctx, workPersonnel, WorkInput{Title: "ผลงาน"})
	_, err := svc.UploadFile(ctx, workPersonnel, id, uploadInput(bytes.NewReader(nil), MaxWorkFileSize+1, domain.WorkFileImage))
	if !errors.Is(err, domain.ErrFileTooLarge) {
		t.Errorf("err = %v, want ErrFileTooLarge", err)
	}
}

func TestWorkFile_UploadWorkNotFound(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	_, err := svc.UploadFile(wAdmin(workSchool, "sem-1"), workPersonnel, "missing", uploadInput(bytes.NewReader([]byte("x")), 1, domain.WorkFileImage))
	if !errors.Is(err, domain.ErrWorkNotFound) {
		t.Errorf("err = %v, want ErrWorkNotFound", err)
	}
}

func TestWorkFile_UploadStorageUnavailable(t *testing.T) {
	repo := newFakeWorkRepo()
	svc := NewPersonnelWorkService(repo, guardWith(workSchool, workPersonnel), nil) // storage = nil
	_, err := svc.UploadFile(wAdmin(workSchool, "sem-1"), workPersonnel, "any", uploadInput(bytes.NewReader([]byte("x")), 1, domain.WorkFileImage))
	if !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Errorf("err = %v, want ErrStorageUnavailable", err)
	}
}

func TestWorkFile_ListReturnsSignedURL(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	ctx := wAdmin(workSchool, "sem-1")
	id, _ := svc.Create(ctx, workPersonnel, WorkInput{Title: "ผลงาน"})
	if _, err := svc.UploadFile(ctx, workPersonnel, id, uploadInput(bytes.NewReader([]byte("x")), 1, domain.WorkFileCertificate)); err != nil {
		t.Fatalf("upload: %v", err)
	}
	files, err := svc.ListFiles(ctx, workPersonnel, id)
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	if len(files) != 1 || files[0].URL == "" {
		t.Errorf("files=%+v ควรมี 1 รายการพร้อม signed URL", files)
	}
}

func TestWorkFile_DeleteNotFound(t *testing.T) {
	svc, _, _, _ := newWorkSvc()
	ctx := wAdmin(workSchool, "sem-1")
	id, _ := svc.Create(ctx, workPersonnel, WorkInput{Title: "ผลงาน"})
	if err := svc.DeleteFile(ctx, workPersonnel, id, "missing"); !errors.Is(err, domain.ErrWorkFileNotFound) {
		t.Errorf("err = %v, want ErrWorkFileNotFound", err)
	}
}

func TestWorkFile_DeleteRemovesFromStorage(t *testing.T) {
	svc, _, store, _ := newWorkSvc()
	ctx := wAdmin(workSchool, "sem-1")
	id, _ := svc.Create(ctx, workPersonnel, WorkInput{Title: "ผลงาน"})
	fid, _ := svc.UploadFile(ctx, workPersonnel, id, uploadInput(bytes.NewReader([]byte("x")), 1, domain.WorkFileImage))
	if err := svc.DeleteFile(ctx, workPersonnel, id, fid); err != nil {
		t.Fatalf("delete file: %v", err)
	}
	if len(store.objects) != 0 {
		t.Errorf("object ใน storage เหลือ %d ควรเป็น 0", len(store.objects))
	}
}
