package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/face"
	"github.com/chumkosoft/backend/internal/tenant"
)

// --- fakes ---

type fakeEmbedder struct {
	vec      []float32
	yaws     []float32 // yaw ต่อการเรียกแต่ละครั้ง (วน) — ใช้ทดสอบ liveness
	i        int
	noFace   bool
	disabled bool
}

func (f *fakeEmbedder) Enabled() bool { return !f.disabled }
func (f *fakeEmbedder) Embed(_ context.Context, _ []byte) (face.EmbedResult, error) {
	if f.noFace {
		return face.EmbedResult{}, domain.ErrNoFaceDetected
	}
	var yaw float32
	if len(f.yaws) > 0 {
		yaw = f.yaws[f.i%len(f.yaws)]
		f.i++
	}
	return face.EmbedResult{Embedding: f.vec, Yaw: yaw}, nil
}

type fakeFacePhotos struct{ rows []domain.StudentPhotoRow }

func (f *fakeFacePhotos) Dataset(_ context.Context, _, _, _ string) ([]domain.StudentPhotoRow, error) {
	return f.rows, nil
}

type fakeFaceStore struct {
	upserts  int
	list     []domain.FaceEmbedding
	keptCall []string
}

func (s *fakeFaceStore) Upsert(_ context.Context, _, _, _ string, _ []float32) error {
	s.upserts++
	return nil
}
func (s *fakeFaceStore) ListBySchool(_ context.Context, _ string) ([]domain.FaceEmbedding, error) {
	return s.list, nil
}
func (s *fakeFaceStore) DeleteOrphans(_ context.Context, _ string, keep []string) error {
	s.keptCall = keep
	return nil
}

type fakeObjGetter struct{}

func (fakeObjGetter) Get(_ context.Context, _ string) ([]byte, error) { return []byte("img"), nil }

type fakeFaceStudents struct {
	byID    map[string]*domain.Student
	classID string
}

func (r *fakeFaceStudents) GetByID(_ context.Context, _, id string) (*domain.Student, error) {
	return r.byID[id], nil
}
func (r *fakeFaceStudents) CurrentClass(_ context.Context, _, _, _ string) (string, string, string, error) {
	if r.classID == "" {
		return "", "", "", nil
	}
	return "enr1", r.classID, "ม.1 1", nil
}

type fakeFaceAttendance struct {
	marked        []domain.AttendanceMark
	existing      string // สถานะที่มีอยู่แล้ว (จำลองสแกนซ้ำ)
	existingFound bool
}

func (a *fakeFaceAttendance) BulkUpsert(_ context.Context, _, _, _ string, _ time.Time, marks []domain.AttendanceMark, _ string, _ domain.AuditEntry) error {
	a.marked = append(a.marked, marks...)
	return nil
}

func (a *fakeFaceAttendance) StatusForStudentDate(_ context.Context, _, _ string, _ time.Time) (string, bool, error) {
	return a.existing, a.existingFound, nil
}

// fakeFaceSchool คืนค่าตั้งโรงเรียน (เวลาตัดสาย + คะแนนหัก) สำหรับทดสอบ
type fakeFaceSchool struct {
	lateAfter string
	penalty   int
}

func (s *fakeFaceSchool) Get(_ context.Context, _ string) (*domain.School, error) {
	return &domain.School{AttendanceLateAfter: s.lateAfter, AttendanceLatePenalty: s.penalty}, nil
}

// fakeFaceBehavior บันทึกการหักคะแนนความประพฤติ
type fakeFaceBehavior struct{ records []domain.NewBehaviorRecord }

func (b *fakeFaceBehavior) Create(_ context.Context, _, _, _ string, nr domain.NewBehaviorRecord, _ string, _ domain.AuditEntry) (string, error) {
	b.records = append(b.records, nr)
	return "b1", nil
}

func newFaceSvc(emb *fakeEmbedder, photos *fakeFacePhotos, store *fakeFaceStore, students *fakeFaceStudents, att *fakeFaceAttendance) *FaceService {
	// liveness=true เพื่อทดสอบฟีเจอร์ (production ตั้งค่าผ่าน config; ดีฟอลต์ปิด)
	return NewFaceService(emb, photos, store, fakeObjGetter{}, students, att, &fakeWGChecker{}, "08:00", true, nil, nil)
}

// liveFrames = ชุดเฟรมจำลอง 2 เฟรม (เนื้อหาไม่สำคัญ — fake embedder คุม yaw เอง)
var liveFrames = [][]byte{[]byte("f1"), []byte("f2")}

// movingYaws = yaw ที่ขยับพอผ่าน liveness
var movingYaws = []float32{-0.2, 0.2}

func kioskCtx() context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{UserID: "k1", SchoolID: "school-A", Role: "kiosk", SemesterID: "sem-1"})
}

// --- reindex tests ---

func TestFace_ReindexForbidden(t *testing.T) {
	svc := newFaceSvc(&fakeEmbedder{}, &fakeFacePhotos{}, &fakeFaceStore{}, &fakeFaceStudents{}, &fakeFaceAttendance{})
	if _, err := svc.Reindex(memberCtx("school-A", "u9")); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestFace_ReindexForbiddenForKioskRole(t *testing.T) {
	svc := newFaceSvc(&fakeEmbedder{}, &fakeFacePhotos{}, &fakeFaceStore{}, &fakeFaceStudents{}, &fakeFaceAttendance{})
	if _, err := svc.Reindex(kioskCtx()); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("kiosk ต้อง reindex ไม่ได้: %v", err)
	}
}

func TestFace_ReindexServiceUnavailable(t *testing.T) {
	svc := newFaceSvc(&fakeEmbedder{disabled: true}, &fakeFacePhotos{}, &fakeFaceStore{}, &fakeFaceStudents{}, &fakeFaceAttendance{})
	if _, err := svc.Reindex(wAdmin("school-A", "sem-1")); !errors.Is(err, domain.ErrFaceServiceUnavailable) {
		t.Errorf("err = %v, want ErrFaceServiceUnavailable", err)
	}
}

func TestFace_ReindexEnrollsAndSkipsNoFace(t *testing.T) {
	photos := &fakeFacePhotos{rows: []domain.StudentPhotoRow{
		{StudentID: "s1", PhotoID: "p1", StoragePath: "a.jpg"},
		{StudentID: "s1", PhotoID: "p2", StoragePath: "b.jpg"},
	}}
	store := &fakeFaceStore{}
	svc := newFaceSvc(&fakeEmbedder{vec: []float32{1, 0, 0}}, photos, store, &fakeFaceStudents{}, &fakeFaceAttendance{})
	res, err := svc.Reindex(wAdmin("school-A", "sem-1"))
	if err != nil {
		t.Fatalf("reindex: %v", err)
	}
	if res.Enrolled != 2 || res.Total != 2 || store.upserts != 2 {
		t.Errorf("res = %+v upserts=%d, want enrolled 2", res, store.upserts)
	}
	if len(store.keptCall) != 2 {
		t.Errorf("DeleteOrphans keep = %v, want 2", store.keptCall)
	}
}

// --- recognize tests ---

func TestFace_RecognizeForbidden(t *testing.T) {
	svc := newFaceSvc(&fakeEmbedder{}, &fakeFacePhotos{}, &fakeFaceStore{}, &fakeFaceStudents{}, &fakeFaceAttendance{})
	if _, err := svc.RecognizeAndMark(memberCtx("school-A", "u9"), liveFrames); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestFace_RecognizeAllowedForKioskRole(t *testing.T) {
	// kiosk role ผ่าน guard → ไปถึง embed (noFace → ErrNoFaceDetected)
	svc := newFaceSvc(&fakeEmbedder{noFace: true}, &fakeFacePhotos{}, &fakeFaceStore{}, &fakeFaceStudents{}, &fakeFaceAttendance{})
	if _, err := svc.RecognizeAndMark(kioskCtx(), liveFrames); !errors.Is(err, domain.ErrNoFaceDetected) {
		t.Errorf("err = %v, want ErrNoFaceDetected (ผ่าน guard)", err)
	}
}

func TestFace_RecognizeNoFace(t *testing.T) {
	svc := newFaceSvc(&fakeEmbedder{noFace: true}, &fakeFacePhotos{}, &fakeFaceStore{}, &fakeFaceStudents{}, &fakeFaceAttendance{})
	if _, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames); !errors.Is(err, domain.ErrNoFaceDetected) {
		t.Errorf("err = %v, want ErrNoFaceDetected", err)
	}
}

func TestFace_RecognizeLivenessFailWhenStill(t *testing.T) {
	// yaw ไม่ขยับ (รูปนิ่ง) → liveness ไม่ผ่าน, ไม่ match/ไม่ mark
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{1, 0, 0}}}}
	att := &fakeFaceAttendance{}
	svc := newFaceSvc(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: []float32{0.0, 0.0}}, &fakeFacePhotos{}, store, &fakeFaceStudents{}, att)
	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames)
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if res.Matched || res.LivenessPassed || len(att.marked) != 0 {
		t.Errorf("รูปนิ่งต้องไม่ผ่าน liveness: %+v", res)
	}
}

func TestFace_RecognizeLivenessOffMarksStillFrame(t *testing.T) {
	// liveness ปิด → รูปนิ่ง/เฟรมเดียวก็ match + mark ได้ (ระบบสแกนแบบเดิม)
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{1, 0, 0}}}}
	students := &fakeFaceStudents{byID: map[string]*domain.Student{"s1": {ID: "s1", StudentCode: "S001"}}, classID: "c1"}
	att := &fakeFaceAttendance{}
	svc := NewFaceService(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: []float32{0.0}}, &fakeFacePhotos{}, store, fakeObjGetter{}, students, att, &fakeWGChecker{}, "08:00", false, nil, nil)
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 30, 0, 0, time.UTC) }
	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), [][]byte{[]byte("one")})
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if !res.Matched || !res.Marked {
		t.Errorf("liveness ปิด: เฟรมเดียวต้อง match+mark ได้: %+v", res)
	}
}

func TestFace_RecognizeMatchAndMark(t *testing.T) {
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{1, 0, 0}}}}
	students := &fakeFaceStudents{
		byID:    map[string]*domain.Student{"s1": {ID: "s1", StudentCode: "S001", Profile: domain.PersonProfile{FirstName: "ก", LastName: "ข"}}},
		classID: "c1",
	}
	att := &fakeFaceAttendance{}
	svc := newFaceSvc(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: movingYaws}, &fakeFacePhotos{}, store, students, att)
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 30, 0, 0, time.UTC) } // 07:30 ICT → present

	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames)
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if !res.Matched || !res.LivenessPassed || res.StudentID != "s1" || !res.Marked || res.Status != domain.AttendancePresent {
		t.Errorf("res = %+v, want matched+live+marked s1 present", res)
	}
	if len(att.marked) != 1 || att.marked[0].StudentID != "s1" {
		t.Errorf("marks = %+v", att.marked)
	}
}

func TestFace_RecognizeAlreadyMarkedDoesNotReMark(t *testing.T) {
	// เคยเช็คชื่อวันนี้แล้ว → แจ้ง already_marked, ไม่บันทึกทับ
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{1, 0, 0}}}}
	students := &fakeFaceStudents{byID: map[string]*domain.Student{"s1": {ID: "s1", StudentCode: "S001"}}, classID: "c1"}
	att := &fakeFaceAttendance{existing: domain.AttendancePresent, existingFound: true}
	svc := newFaceSvc(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: movingYaws}, &fakeFacePhotos{}, store, students, att)
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 30, 0, 0, time.UTC) }
	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames)
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if !res.Matched || !res.AlreadyMarked || res.Marked || res.Status != domain.AttendancePresent {
		t.Errorf("res = %+v, want matched + already_marked + not re-marked", res)
	}
	if len(att.marked) != 0 {
		t.Errorf("ต้องไม่บันทึกทับ: marks = %+v", att.marked)
	}
}

func TestFace_LateDeductsBehaviorPoints(t *testing.T) {
	// มาสาย + โรงเรียนตั้งหักคะแนน 5 → บันทึกหัก -5 ในคะแนนความประพฤติ + res.PenaltyApplied
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{1, 0, 0}}}}
	students := &fakeFaceStudents{byID: map[string]*domain.Student{"s1": {ID: "s1", StudentCode: "S001"}}, classID: "c1"}
	att := &fakeFaceAttendance{}
	beh := &fakeFaceBehavior{}
	svc := NewFaceService(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: movingYaws}, &fakeFacePhotos{}, store, fakeObjGetter{},
		students, att, &fakeWGChecker{}, "08:00", true, &fakeFaceSchool{lateAfter: "08:00", penalty: 5}, beh)
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 2, 30, 0, 0, time.UTC) } // 09:30 ICT → late
	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames)
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if res.Status != domain.AttendanceLate || res.PenaltyApplied != 5 {
		t.Errorf("res = %+v, want late + penalty 5", res)
	}
	if len(beh.records) != 1 || beh.records[0].Points != -5 {
		t.Errorf("behavior = %+v, want one record of -5", beh.records)
	}
}

func TestFace_PresentDoesNotDeduct(t *testing.T) {
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{1, 0, 0}}}}
	students := &fakeFaceStudents{byID: map[string]*domain.Student{"s1": {ID: "s1", StudentCode: "S001"}}, classID: "c1"}
	beh := &fakeFaceBehavior{}
	svc := NewFaceService(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: movingYaws}, &fakeFacePhotos{}, store, fakeObjGetter{},
		students, &fakeFaceAttendance{}, &fakeWGChecker{}, "08:00", true, &fakeFaceSchool{lateAfter: "08:00", penalty: 5}, beh)
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 30, 0, 0, time.UTC) } // 07:30 ICT → present
	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames)
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if res.PenaltyApplied != 0 || len(beh.records) != 0 {
		t.Errorf("มาเรียนตรงเวลาต้องไม่หักคะแนน: penalty=%d records=%+v", res.PenaltyApplied, beh.records)
	}
}

func TestFace_RecognizeMarksLateAfterCutoff(t *testing.T) {
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{1, 0, 0}}}}
	students := &fakeFaceStudents{byID: map[string]*domain.Student{"s1": {ID: "s1", StudentCode: "S001"}}, classID: "c1"}
	svc := newFaceSvc(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: movingYaws}, &fakeFacePhotos{}, store, students, &fakeFaceAttendance{})
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 2, 30, 0, 0, time.UTC) } // 09:30 ICT → late
	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames)
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if res.Status != domain.AttendanceLate {
		t.Errorf("status = %q, want late", res.Status)
	}
}

func TestFace_RecognizeNoMatch(t *testing.T) {
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{0, 1, 0}}}}
	svc := newFaceSvc(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: movingYaws}, &fakeFacePhotos{}, store, &fakeFaceStudents{}, &fakeFaceAttendance{})
	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames)
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if res.Matched {
		t.Errorf("ไม่ควร match: %+v", res)
	}
}

func TestFace_RecognizeMatchButNoClass(t *testing.T) {
	store := &fakeFaceStore{list: []domain.FaceEmbedding{{StudentID: "s1", Vector: []float32{1, 0, 0}}}}
	students := &fakeFaceStudents{byID: map[string]*domain.Student{"s1": {ID: "s1", StudentCode: "S001"}}, classID: ""}
	att := &fakeFaceAttendance{}
	svc := newFaceSvc(&fakeEmbedder{vec: []float32{1, 0, 0}, yaws: movingYaws}, &fakeFacePhotos{}, store, students, att)
	res, err := svc.RecognizeAndMark(wAdmin("school-A", "sem-1"), liveFrames)
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}
	if !res.Matched || res.Marked || res.Reason == "" {
		t.Errorf("ควร match แต่ไม่ mark (ไม่มีห้อง): %+v", res)
	}
	if len(att.marked) != 0 {
		t.Error("ต้องไม่บันทึกเมื่อไม่มีห้อง")
	}
}
