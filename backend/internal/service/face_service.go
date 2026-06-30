package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/face"
	"github.com/chumkosoft/backend/internal/tenant"
)

// --- contracts (narrow interfaces เพื่อ test ง่าย) ---

type faceEmbedder interface {
	Embed(ctx context.Context, image []byte) (face.EmbedResult, error)
	Enabled() bool
}

// livenessYawDelta = พิสัย yaw ขั้นต่ำระหว่างเฟรมในชุด ที่ถือว่า "ขยับหน้าจริง" (กันรูปนิ่ง)
const livenessYawDelta = 0.12

type facePhotoSource interface {
	Dataset(ctx context.Context, schoolID, semesterID, classID string) ([]domain.StudentPhotoRow, error)
}

type faceEmbeddingStore interface {
	Upsert(ctx context.Context, schoolID, studentID, photoID string, vector []float32) error
	ListBySchool(ctx context.Context, schoolID string) ([]domain.FaceEmbedding, error)
	DeleteOrphans(ctx context.Context, schoolID string, keepPhotoIDs []string) error
}

type faceObjectGetter interface {
	Get(ctx context.Context, objectPath string) ([]byte, error)
}

type faceStudentRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Student, error)
	CurrentClass(ctx context.Context, schoolID, semesterID, studentID string) (enrollmentID, classID, label string, err error)
}

type faceAttendanceRepo interface {
	BulkUpsert(ctx context.Context, schoolID, semesterID, classID string, date time.Time, marks []domain.AttendanceMark, checkedBy string, audit domain.AuditEntry) error
	StatusForStudentDate(ctx context.Context, schoolID, studentID string, date time.Time) (string, bool, error)
}

// faceSchoolRepo อ่านค่าตั้งของโรงเรียน (เวลาตัดสาย + คะแนนหัก) — อาจเป็น nil (ใช้ค่า default)
type faceSchoolRepo interface {
	Get(ctx context.Context, schoolID string) (*domain.School, error)
}

// faceBehaviorRepo บันทึกหักคะแนนความประพฤติเมื่อมาสาย — อาจเป็น nil (ปิดการหักคะแนน)
type faceBehaviorRepo interface {
	Create(ctx context.Context, schoolID, semesterID, studentID string, nr domain.NewBehaviorRecord, recordedBy string, audit domain.AuditEntry) (string, error)
}

// --- DTOs ---

type ReindexResult struct {
	Enrolled int `json:"enrolled"` // จำนวนรูปที่คำนวณ embedding สำเร็จ
	Skipped  int `json:"skipped"`  // รูปที่ตรวจไม่พบใบหน้า
	Total    int `json:"total"`    // รูปทั้งหมดที่ลองทำ
}

type RecognizeResult struct {
	Matched     bool    `json:"matched"`
	StudentID   string  `json:"student_id"`
	StudentCode string  `json:"student_code"`
	FullName    string  `json:"full_name"`
	ClassLabel  string  `json:"class_label"`
	Score          float32 `json:"score"`
	Marked         bool    `json:"marked"`          // บันทึกเช็คชื่อแล้วหรือยัง
	AlreadyMarked  bool    `json:"already_marked"`  // เคยเช็คชื่อวันนี้แล้ว (สแกนซ้ำ — ไม่บันทึกทับ)
	Status         string  `json:"status"`          // สถานะที่บันทึก/ที่มีอยู่ (present/late)
	PenaltyApplied int     `json:"penalty_applied"` // คะแนนความประพฤติที่หักไป (เมื่อมาสาย)
	LivenessPassed bool    `json:"liveness_passed"` // ผ่านการตรวจว่าเป็นคนจริง (ขยับหน้า)
	Reason         string  `json:"reason"`          // เหตุผลถ้าไม่ได้บันทึก
}

// FaceService — ระบบสแกนหน้าเข้าเรียน: enroll (จากรูปนักเรียน) + จดจำ + บันทึกเช็คชื่อ
// สิทธิ์: กลุ่มวิชาการ/แอดมิน (เจ้าของข้อมูลนักเรียน + รูป + ใบหน้า)
type FaceService struct {
	guard      academicGuard
	embedder   faceEmbedder
	photos     facePhotoSource
	store      faceEmbeddingStore
	storage    faceObjectGetter
	students   faceStudentRepo
	attendance faceAttendanceRepo
	schoolRepo faceSchoolRepo
	behavior   faceBehaviorRepo
	threshold  float32
	lateAfter  string // "HH:MM" เวลาตัดสาย default (ถ้าโรงเรียนไม่ได้ตั้ง)
	liveness   bool   // เปิดการตรวจ liveness (ต้องขยับหน้า)
	now        func() time.Time
}

// thaiZone = เขตเวลาไทย (UTC+7, ไม่มี DST) — ใช้ fixed offset เลี่ยงพึ่ง tzdata ใน image
var thaiZone = time.FixedZone("ICT", 7*3600)

func abs32(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}

func NewFaceService(
	embedder faceEmbedder,
	photos facePhotoSource,
	store faceEmbeddingStore,
	storage faceObjectGetter,
	students faceStudentRepo,
	attendance faceAttendanceRepo,
	checker WorkGroupChecker,
	lateAfter string,
	liveness bool,
	schoolRepo faceSchoolRepo,
	behavior faceBehaviorRepo,
) *FaceService {
	if lateAfter == "" {
		lateAfter = "08:00"
	}
	return &FaceService{
		guard:      academicGuard{checker: checker},
		embedder:   embedder,
		photos:     photos,
		store:      store,
		storage:    storage,
		students:   students,
		attendance: attendance,
		schoolRepo: schoolRepo,
		behavior:   behavior,
		threshold:  face.DefaultThreshold,
		lateAfter:  lateAfter,
		liveness:   liveness,
		now:        time.Now,
	}
}

// attendanceStatusNow คืน present/late ตามเวลาปัจจุบัน (โซนไทย) เทียบเวลาตัดของโรงเรียน, วันที่ (โซนไทย), และคะแนนหักเมื่อสาย
func (s *FaceService) attendanceStatusNow(ctx context.Context) (status string, date time.Time, latePenalty int) {
	cutoff := s.lateAfter
	if s.schoolRepo != nil {
		if sc, err := s.schoolRepo.Get(ctx, tenant.SchoolIDFromContext(ctx)); err == nil && sc != nil {
			if sc.AttendanceLateAfter != "" {
				cutoff = sc.AttendanceLateAfter
			}
			latePenalty = sc.AttendanceLatePenalty
		}
	}
	now := s.now().In(thaiZone)
	status = domain.AttendancePresent
	var h, m int
	if _, err := fmt.Sscanf(cutoff, "%d:%d", &h, &m); err == nil {
		cutoff := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, thaiZone)
		if now.After(cutoff) {
			status = domain.AttendanceLate
		}
	}
	return status, now, latePenalty
}

// Reindex สร้างฐานใบหน้าใหม่ทั้งโรงเรียนจากรูปนักเรียนทั้งหมด (idempotent)
func (s *FaceService) Reindex(ctx context.Context) (ReindexResult, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return ReindexResult{}, err
	}
	if !s.embedder.Enabled() {
		return ReindexResult{}, domain.ErrFaceServiceUnavailable
	}
	if s.storage == nil {
		return ReindexResult{}, domain.ErrStorageUnavailable
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	rows, err := s.photos.Dataset(ctx, schoolID, tenant.SemesterIDFromContext(ctx), "")
	if err != nil {
		return ReindexResult{}, err
	}

	var res ReindexResult
	keep := make([]string, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		res.Total++
		img, err := s.storage.Get(ctx, r.StoragePath)
		if err != nil {
			return res, err
		}
		emb, err := s.embedder.Embed(ctx, img)
		if errors.Is(err, domain.ErrNoFaceDetected) {
			res.Skipped++
			continue
		}
		if err != nil {
			return res, err
		}
		if err := s.store.Upsert(ctx, schoolID, r.StudentID, r.PhotoID, emb.Embedding); err != nil {
			return res, err
		}
		keep = append(keep, r.PhotoID)
		res.Enrolled++
	}
	// ล้าง embedding ของรูปที่ถูกลบไปแล้ว
	if err := s.store.DeleteOrphans(ctx, schoolID, keep); err != nil {
		return res, err
	}
	return res, nil
}

// authorizeRecognize: บัญชี kiosk หรือกลุ่มวิชาการ/แอดมิน เรียกสแกนได้ (reindex ยังจำกัดวิชาการ/แอดมิน)
func (s *FaceService) authorizeRecognize(ctx context.Context) error {
	if tenant.RoleFromContext(ctx) == domain.RoleKiosk {
		return nil
	}
	return s.guard.authorize(ctx)
}

// RecognizeAndMark จดจำใบหน้าจากชุดเฟรม (burst) + ตรวจ liveness (ขยับหน้า) แล้วบันทึกเช็คชื่อ
// frames = หลายเฟรมที่ถ่ายต่อเนื่องระหว่างให้ผู้ใช้หันหน้าซ้าย-ขวา (กันการถือรูปนิ่ง)
func (s *FaceService) RecognizeAndMark(ctx context.Context, frames [][]byte) (RecognizeResult, error) {
	if err := s.authorizeRecognize(ctx); err != nil {
		return RecognizeResult{}, err
	}
	if !s.embedder.Enabled() {
		return RecognizeResult{}, domain.ErrFaceServiceUnavailable
	}

	// คำนวณ embedding + yaw ของทุกเฟรมที่มีใบหน้า
	usable := make([]face.EmbedResult, 0, len(frames))
	for i := range frames {
		r, err := s.embedder.Embed(ctx, frames[i])
		if errors.Is(err, domain.ErrNoFaceDetected) {
			continue
		}
		if err != nil {
			return RecognizeResult{}, err
		}
		usable = append(usable, r)
	}
	if len(usable) == 0 {
		return RecognizeResult{}, domain.ErrNoFaceDetected
	}

	// liveness: ต้องมีการขยับหน้าจริง (พิสัย yaw ระหว่างเฟรม) — รูปนิ่งบนมือถือจะไม่ผ่าน
	minYaw, maxYaw := usable[0].Yaw, usable[0].Yaw
	frontal := usable[0] // เฟรมที่หน้าตรงที่สุด (|yaw| น้อยสุด) ใช้ match
	for i := range usable {
		y := usable[i].Yaw
		if y < minYaw {
			minYaw = y
		}
		if y > maxYaw {
			maxYaw = y
		}
		if abs32(y) < abs32(frontal.Yaw) {
			frontal = usable[i]
		}
	}
	if s.liveness && (len(usable) < 2 || (maxYaw-minYaw) < livenessYawDelta) {
		return RecognizeResult{Matched: false, LivenessPassed: false,
			Reason: "ตรวจไม่พบการขยับใบหน้า — โปรดหันหน้าซ้าย-ขวาช้า ๆ ระหว่างสแกน"}, nil
	}
	query := frontal.Embedding

	schoolID := tenant.SchoolIDFromContext(ctx)
	embs, err := s.store.ListBySchool(ctx, schoolID)
	if err != nil {
		return RecognizeResult{}, err
	}
	samples := make([]face.Sample, len(embs))
	for i := range embs {
		samples[i] = face.Sample{StudentID: embs[i].StudentID, Vector: embs[i].Vector}
	}

	id, score, ok := face.BestMatch(query, samples, s.threshold)
	if !ok {
		return RecognizeResult{Matched: false, Score: score, Reason: "ไม่พบนักเรียนที่ตรงกับใบหน้า"}, nil
	}

	st, err := s.students.GetByID(ctx, schoolID, id)
	if err != nil {
		return RecognizeResult{}, err
	}
	if st == nil {
		return RecognizeResult{Matched: false, Score: score, Reason: "ข้อมูลใบหน้าไม่ตรงกับนักเรียนปัจจุบัน"}, nil
	}

	res := RecognizeResult{
		Matched: true, LivenessPassed: s.liveness, StudentID: id, StudentCode: st.StudentCode, Score: score,
		FullName: strings.TrimSpace(st.Profile.Prefix + st.Profile.FirstName + " " + st.Profile.LastName),
	}

	sem := tenant.SemesterIDFromContext(ctx)
	if sem == "" {
		res.Reason = "ยังไม่กำหนดเทอมทำงาน จึงยังไม่บันทึกเช็คชื่อ"
		return res, nil
	}
	_, classID, label, err := s.students.CurrentClass(ctx, schoolID, sem, id)
	if err != nil {
		return RecognizeResult{}, err
	}
	if classID == "" {
		res.Reason = "นักเรียนยังไม่ถูกจัดห้องในเทอมนี้ จึงยังไม่บันทึกเช็คชื่อ"
		return res, nil
	}
	res.ClassLabel = label

	status, date, latePenalty := s.attendanceStatusNow(ctx)

	// กันสแกนซ้ำ: ถ้าวันนี้เช็คชื่อไปแล้ว ไม่บันทึกทับ แต่แจ้งผู้สแกนว่า "สแกนไปแล้ว"
	if existing, found, err := s.attendance.StatusForStudentDate(ctx, schoolID, id, date); err != nil {
		return RecognizeResult{}, err
	} else if found {
		res.AlreadyMarked = true
		res.Status = existing
		res.Reason = "นักเรียนคนนี้เช็คชื่อเข้าเรียนไปแล้ววันนี้"
		return res, nil
	}

	userID := tenant.UserIDFromContext(ctx)
	audit := auditFor(ctx, domain.AuditUpdate, "attendance", classID, map[string]any{"student_id": id, "source": "face", "status": status})
	if err := s.attendance.BulkUpsert(ctx, schoolID, sem, classID, date,
		[]domain.AttendanceMark{{StudentID: id, Status: status}}, userID, audit); err != nil {
		return RecognizeResult{}, err
	}
	res.Marked = true
	res.Status = status

	// มาสาย → หักคะแนนความประพฤติ + บันทึกประวัติ (ครั้งแรกของวันเท่านั้น)
	if status == domain.AttendanceLate && latePenalty > 0 && s.behavior != nil {
		occurred := date
		bAudit := auditFor(ctx, domain.AuditCreate, "behavior_record", id, map[string]any{"points": -latePenalty, "source": "face", "reason": "late"})
		if _, err := s.behavior.Create(ctx, schoolID, sem, id,
			domain.NewBehaviorRecord{Points: -latePenalty, Reason: "มาสาย (สแกนหน้าเข้าเรียน)", OccurredAt: &occurred},
			userID, bAudit); err != nil {
			return RecognizeResult{}, err
		}
		res.PenaltyApplied = latePenalty
	}
	return res, nil
}
