// Package server ประกอบ Fiber app, middleware กลาง, และ route ทั้งหมด
package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/chumkosoft/backend/internal/auth"
	"github.com/chumkosoft/backend/internal/config"
	"github.com/chumkosoft/backend/internal/crypto"
	"github.com/chumkosoft/backend/internal/face"
	"github.com/chumkosoft/backend/internal/handler"
	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/middleware"
	"github.com/chumkosoft/backend/internal/repository"
	"github.com/chumkosoft/backend/internal/service"
	"github.com/chumkosoft/backend/internal/storage"
)

// Deps คือ dependency ที่ server ต้องใช้ (inject จาก main)
type Deps struct {
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Config *config.Config
	// Cipher ใช้เข้ารหัสข้อมูลอ่อนไหว (เลขบัตรประชาชน) — สร้างจาก main
	Cipher *crypto.Cipher
	// Storage เก็บไฟล์แนบส่วนบุคคล (signed URL) — อาจเป็น nil ถ้าไม่ได้ตั้งค่า
	Storage storage.Storage
}

// New สร้าง Fiber app พร้อม middleware กลางและ route พื้นฐาน
func New(deps Deps) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "chumkosoft",
		DisableStartupMessage: true,
		// รองรับอัปโหลดไฟล์แนบผลงานครู (สูงสุด 10 MB + เผื่อ overhead ของ multipart)
		BodyLimit: 12 << 20,
		// ErrorHandler กลาง: ตอบ error format มาตรฐาน
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return httputil.Error(c, code, "INTERNAL_ERROR", "เกิดข้อผิดพลาดภายในระบบ")
		},
	})

	for _, m := range middleware.Common() {
		app.Use(m)
	}

	registerHealth(app, deps)

	// token manager ใช้ร่วมกันทั้ง auth (ออก token) และ middleware ตรวจสิทธิ์
	tokens := auth.NewTokenManager(deps.Config.JWTSecret, deps.Config.AccessTokenExpiry)

	// route เวอร์ชัน 1 — feature ของ Phase 1 mount ใต้นี้
	v1 := app.Group("/api/v1")
	registerAuth(v1, deps, tokens)
	registerPersonnel(v1, deps, tokens)
	registerStudents(v1, deps, tokens)
	registerClasses(v1, deps, tokens)
	registerAttendance(v1, deps, tokens)
	registerBehavior(v1, deps, tokens)
	registerTimetable(v1, deps, tokens)
	registerTerm(v1, deps, tokens)
	registerAcademic(v1, deps, tokens)
	registerSchool(v1, deps, tokens)
	registerDashboard(v1, deps, tokens)
	registerMe(v1, deps, tokens)
	registerDirectory(v1, deps, tokens)
	registerFace(v1, deps, tokens)
	registerKioskAccounts(v1, deps, tokens)

	return app
}

// registerFace mount route ระบบสแกนหน้าเข้าเรียน (กลุ่มวิชาการ/แอดมิน)
// reindex = สร้างฐานใบหน้าจากรูปนักเรียน; recognize = จดจำจากกล้อง + บันทึกเช็คชื่อ
func registerFace(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	svc := service.NewFaceService(
		face.NewClient(deps.Config.FaceServiceURL),
		repository.NewStudentPhotoRepository(deps.DB),
		repository.NewStudentFaceEmbeddingRepository(deps.DB),
		deps.Storage,
		repository.NewStudentRepository(deps.DB),
		repository.NewAttendanceRepository(deps.DB),
		repository.NewWorkGroupRepository(deps.DB),
		deps.Config.AttendanceLateAfter,
		deps.Config.FaceLiveness,
		repository.NewSchoolRepository(deps.DB),
		repository.NewBehaviorRepository(deps.DB),
	)
	h := handler.NewFaceHandler(svc)

	grp := router.Group("/face", requireAuth(deps, tokens))
	grp.Post("/reindex", h.Reindex)
	grp.Post("/recognize", h.Recognize)
}

// registerKioskAccounts mount route จัดการบัญชีเครื่องสแกนหน้า (school admin เท่านั้น — บังคับใน service)
func registerKioskAccounts(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	h := handler.NewKioskAccountHandler(service.NewKioskAccountService(repository.NewUserRepository(deps.DB)))
	grp := router.Group("/kiosk-accounts", requireAuth(deps, tokens))
	grp.Get("/", h.List)
	grp.Post("/", h.Create)
	grp.Delete("/:id", h.Delete)
}

// registerDirectory mount route ดู/ค้นหาห้องเรียน-นักเรียน (read-only, ข้อมูลพื้นฐาน — ครู/บุคลากร)
func registerDirectory(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	svc := service.NewDirectoryService(
		repository.NewClassRepository(deps.DB),
		repository.NewEnrollmentRepository(deps.DB),
	)
	h := handler.NewDirectoryHandler(svc)

	grp := router.Group("/directory", requireAuth(deps, tokens))
	grp.Get("/classes", h.Classes)
	grp.Get("/classes/:classId/students", h.ClassStudents)
	grp.Get("/students", h.SearchStudents)
}

// registerDashboard mount route ข้อมูลสรุปหน้าแรก (ผู้ล็อกอินทุกคน — สรุปของตัวเอง)
func registerDashboard(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	h := handler.NewDashboardHandler(service.NewDashboardService(repository.NewDashboardRepository(deps.DB)))
	router.Get("/dashboard", requireAuth(deps, tokens), h.Summary)
}

// registerMe mount route self-service ของผู้ใช้เอง (โปรไฟล์ + นักเรียนที่ปรึกษา)
// สิทธิ์มาจากความเป็นเจ้าของ (user_id จาก token) ไม่ต้องสังกัดกลุ่มงาน
func registerMe(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	svc := service.NewMeService(
		repository.NewPersonnelRepository(deps.DB),
		repository.NewDashboardRepository(deps.DB),
		repository.NewStudentRepository(deps.DB),
		deps.Cipher,
	)
	h := handler.NewMeHandler(svc)

	auth := requireAuth(deps, tokens)
	grp := router.Group("/me", auth)
	grp.Get("/profile", h.Profile)
	grp.Put("/profile", h.UpdateProfile)
	grp.Get("/advisees", h.Advisees)
	grp.Get("/advisees/:studentId", h.AdviseeDetail)
	grp.Put("/advisees/:studentId", h.UpdateAdvisee)
}

// registerSchool mount route ข้อมูลโรงเรียน (อ่านได้ทุกคน, แก้ไขเฉพาะ school admin — บังคับใน service)
func registerSchool(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	h := handler.NewSchoolHandler(service.NewSchoolService(repository.NewSchoolRepository(deps.DB)))
	g := router.Group("/school", requireAuth(deps, tokens))
	g.Get("/", h.Get)
	g.Put("/", h.Update)
}

// registerAcademic mount route จัดการปีการศึกษา/ภาคเรียน
// (อ่านได้ทุกคนที่ล็อกอิน, แก้ไขเฉพาะ school admin — บังคับใน service)
func registerAcademic(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	h := handler.NewAcademicHandler(service.NewAcademicService(repository.NewAcademicRepository(deps.DB)))
	g := router.Group("/academic", requireAuth(deps, tokens))

	g.Get("/years", h.ListYears)
	g.Post("/years", h.CreateYear)
	g.Post("/years/:id/current", h.SetCurrentYear)

	g.Get("/semesters", h.ListSemesters)
	g.Post("/semesters", h.CreateSemester)
	g.Post("/semesters/:id/current", h.SetCurrentSemester)
}

// requireAuth สร้าง auth middleware ที่รองรับการสลับเทอม (validate เทอมกับโรงเรียนใน token)
func requireAuth(deps Deps, tokens *auth.TokenManager) fiber.Handler {
	return middleware.RequireAuth(tokens, repository.NewTermRepository(deps.DB))
}

// registerTerm mount route ข้อมูลปี/เทอมปัจจุบัน (ผู้ล็อกอินทุกคนเรียกได้)
func registerTerm(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	h := handler.NewTermHandler(service.NewTermService(repository.NewTermRepository(deps.DB)))
	router.Get("/current-term", requireAuth(deps, tokens), h.Current)
}

// registerTimetable mount route วิชา/มอบหมายสอน/ตารางสอน (กลุ่มวิชาการ + admin)
func registerTimetable(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	wgRepo := repository.NewWorkGroupRepository(deps.DB)
	subjectRepo := repository.NewSubjectRepository(deps.DB)
	personnelRepo := repository.NewPersonnelRepository(deps.DB)
	classRepo := repository.NewClassRepository(deps.DB)

	taRepo := repository.NewTeachingAssignmentRepository(deps.DB)

	subjH := handler.NewSubjectHandler(service.NewSubjectService(subjectRepo, wgRepo))
	taH := handler.NewTeachingAssignmentHandler(service.NewTeachingAssignmentService(
		taRepo, personnelRepo, subjectRepo, classRepo, wgRepo))
	ttH := handler.NewTimetableHandler(service.NewTimetableService(
		repository.NewTimetableRepository(deps.DB), taRepo, classRepo, wgRepo))
	saH := handler.NewSubjectAttendanceHandler(service.NewSubjectAttendanceService(
		repository.NewSubjectAttendanceRepository(deps.DB), wgRepo))

	authMW := requireAuth(deps, tokens)

	sg := router.Group("/subjects", authMW)
	sg.Get("/", subjH.List)
	sg.Post("/", subjH.Create)
	sg.Put("/:id", subjH.Update)
	sg.Delete("/:id", subjH.Delete)

	tg := router.Group("/teaching-assignments", authMW)
	tg.Get("/", taH.List)
	tg.Post("/", taH.Create)
	tg.Delete("/:id", taH.Delete)

	ttg := router.Group("/timetable", authMW)
	ttg.Get("/config", ttH.GetConfig)
	ttg.Put("/config", ttH.SaveConfig)
	ttg.Get("/free-teachers", ttH.FreeTeachers)
	ttg.Get("/classes/:classId", ttH.ListSlots)
	ttg.Post("/classes/:classId/slots", ttH.SetSlot)
	ttg.Delete("/classes/:classId/slots/:slotId", ttH.ClearSlot)

	// เช็คชื่อรายวิชา (รายคาบ) — ครูประจำวิชา/วิชาการ/admin
	ttg.Get("/my-checkin", saH.Overview)
	ttg.Get("/slots/:slotId/attendance", saH.List)
	ttg.Post("/slots/:slotId/attendance", saH.Save)
}

// registerBehavior mount route คะแนนความประพฤติ (กลุ่มบริหารทั่วไป + admin, รายเทอม)
func registerBehavior(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	bH := handler.NewBehaviorHandler(service.NewBehaviorService(
		repository.NewBehaviorRepository(deps.DB),
		repository.NewStudentRepository(deps.DB),
		repository.NewWorkGroupRepository(deps.DB),
	))

	authMW := requireAuth(deps, tokens)
	g := router.Group("/students/:id/behavior", authMW)
	g.Get("/", bH.Summary)
	g.Post("/", bH.Create)
	g.Delete("/:recordId", bH.Delete)
}

// registerAttendance mount route เช็คชื่อเข้าเรียน (กลุ่มบริหารทั่วไป + ครูที่ปรึกษา, รายเทอม)
func registerAttendance(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	aH := handler.NewAttendanceHandler(service.NewAttendanceService(
		repository.NewAttendanceRepository(deps.DB),
		repository.NewClassRepository(deps.DB),
		repository.NewWorkGroupRepository(deps.DB),
	))

	authMW := requireAuth(deps, tokens)
	g := router.Group("/classes/:id/attendance", authMW)
	g.Get("/", aH.List)
	g.Post("/", aH.Save)
}

// registerClasses mount route ห้องเรียน/จัดห้อง (กลุ่มวิชาการ, รายเทอม) — RequireAuth ทุก route
func registerClasses(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	wgRepo := repository.NewWorkGroupRepository(deps.DB)
	classRepo := repository.NewClassRepository(deps.DB)
	personnelRepo := repository.NewPersonnelRepository(deps.DB)
	studentRepo := repository.NewStudentRepository(deps.DB)

	cH := handler.NewClassHandler(service.NewClassService(classRepo, wgRepo))
	aH := handler.NewClassAdvisorHandler(
		service.NewClassAdvisorService(repository.NewClassAdvisorRepository(deps.DB), classRepo, personnelRepo, wgRepo))
	eH := handler.NewEnrollmentHandler(
		service.NewEnrollmentService(repository.NewEnrollmentRepository(deps.DB), classRepo, studentRepo, wgRepo))

	authMW := requireAuth(deps, tokens)
	g := router.Group("/classes", authMW)
	g.Get("/", cH.List)
	g.Post("/", cH.Create)
	g.Get("/:id", cH.Get)
	g.Put("/:id", cH.Update)
	g.Delete("/:id", cH.Delete)

	// ครูที่ปรึกษา
	g.Get("/:id/advisors", aH.List)
	g.Post("/:id/advisors", aH.Add)
	g.Delete("/:id/advisors/:advisorId", aH.Remove)

	// นักเรียนในห้อง
	g.Get("/:id/students", eH.List)
	g.Post("/:id/students", eH.Enroll)
	g.Post("/:id/students/bulk", eH.EnrollMany)
	g.Delete("/:id/students/:enrollmentId", eH.Remove)
}

// registerStudents mount route นักเรียน/ผู้ปกครอง (กลุ่มวิชาการ) — RequireAuth ทุก route
func registerStudents(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	wgRepo := repository.NewWorkGroupRepository(deps.DB) // ใช้ตรวจสิทธิ์กลุ่มวิชาการ
	studentRepo := repository.NewStudentRepository(deps.DB)
	guardianRepo := repository.NewGuardianRepository(deps.DB)
	linkRepo := repository.NewStudentGuardianRepository(deps.DB)

	sH := handler.NewStudentHandler(service.NewStudentService(studentRepo, wgRepo, deps.Cipher))
	photoH := handler.NewStudentPhotoHandler(service.NewStudentPhotoService(
		repository.NewStudentPhotoRepository(deps.DB), studentRepo, wgRepo, deps.Storage))
	gH := handler.NewGuardianHandler(service.NewGuardianService(guardianRepo, wgRepo, deps.Cipher))
	sgH := handler.NewStudentGuardianHandler(
		service.NewStudentGuardianService(linkRepo, studentRepo, guardianRepo, wgRepo, deps.Cipher))

	authMW := requireAuth(deps, tokens)

	sg := router.Group("/students", authMW)
	sg.Get("/", sH.List)
	sg.Post("/", sH.Create)
	sg.Get("/:id", sH.Get)
	sg.Put("/:id", sH.Update)
	sg.Delete("/:id", sH.Delete)
	// dataset รูปนักเรียนเป็นชุด (สำหรับ enroll ระบบสแกนหน้า) — กลุ่มวิชาการ
	router.Get("/face-dataset", authMW, photoH.Dataset)

	// รูปนักเรียน (หลายรูป + เลือกรูปโปรไฟล์ สำหรับระบบสแกนหน้าเข้าเรียน) — signed URL, กลุ่มวิชาการ
	sg.Get("/:id/photos", photoH.List)
	sg.Post("/:id/photos", photoH.Upload)
	sg.Put("/:id/photos/:photoId/primary", photoH.SetPrimary)
	sg.Delete("/:id/photos/:photoId", photoH.Delete)
	// ผู้ปกครองของนักเรียน
	sg.Get("/:id/guardians", sgH.List)
	sg.Post("/:id/guardians", sgH.Link)
	sg.Delete("/:id/guardians/:linkId", sgH.Unlink)

	gg := router.Group("/guardians", authMW)
	gg.Get("/", gH.List)
	gg.Post("/", gH.Create)
	gg.Get("/:id", gH.Get)
	gg.Put("/:id", gH.Update)
	gg.Delete("/:id", gH.Delete)
}

// registerAuth ประกอบ dependency ของกลุ่มงาน auth (repo → service → handler) แล้ว mount route
func registerAuth(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	userRepo := repository.NewUserRepository(deps.DB)
	refreshStore := repository.NewRedisRefreshStore(deps.Redis)
	authSvc := service.NewAuthService(userRepo, tokens, refreshStore, deps.Config.RefreshTokenExpiry)
	authHandler := handler.NewAuthHandler(authSvc)

	grp := router.Group("/auth")
	grp.Post("/login", authHandler.Login)
	grp.Post("/refresh", authHandler.Refresh)
	grp.Post("/logout", authHandler.Logout)
	grp.Get("/me", requireAuth(deps, tokens), authHandler.Me)
}

// registerPersonnel mount route จัดการบุคลากร (กลุ่มงานบุคคล) — ทุก route ต้องผ่าน RequireAuth
func registerPersonnel(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	repo := repository.NewPersonnelRepository(deps.DB)
	svc := service.NewPersonnelService(repo, deps.Cipher)
	h := handler.NewPersonnelHandler(svc)

	// sub-resource: ใช้ personnel repo เป็น guard (ตรวจสิทธิ์ + ยืนยัน personnel มีอยู่)
	posH := handler.NewAdminPositionHandler(service.NewAdminPositionService(repository.NewAdminPositionRepository(deps.DB), repo))
	stH := handler.NewAcademicStandingHandler(service.NewAcademicStandingService(repository.NewAcademicStandingRepository(deps.DB), repo))
	wgH := handler.NewWorkGroupHandler(service.NewWorkGroupService(repository.NewWorkGroupRepository(deps.DB), repo))
	wkH := handler.NewPersonnelWorkHandler(service.NewPersonnelWorkService(repository.NewPersonnelWorkRepository(deps.DB), repo, deps.Storage))

	auth := requireAuth(deps, tokens)

	// list กลุ่มงานของโรงเรียน (ให้เลือกตอนมอบหมาย)
	router.Get("/work-groups", auth, wgH.ListGroups)

	grp := router.Group("/personnel", auth)
	grp.Get("/", h.List)
	grp.Post("/", h.Create)
	grp.Get("/:id", h.Get)
	grp.Put("/:id", h.Update)
	grp.Delete("/:id", h.Delete)

	// ตำแหน่งบริหาร
	grp.Get("/:id/positions", posH.List)
	grp.Post("/:id/positions", posH.Create)
	grp.Delete("/:id/positions/:posId", posH.Delete)

	// วิทยฐานะ (ประวัติ)
	grp.Get("/:id/standings", stH.List)
	grp.Post("/:id/standings", stH.Create)
	grp.Put("/:id/standings/:sid", stH.Update)
	grp.Delete("/:id/standings/:sid", stH.Delete)

	// การมอบหมายกลุ่มงาน
	grp.Get("/:id/work-groups", wgH.ListForPersonnel)
	grp.Post("/:id/work-groups", wgH.Assign)
	grp.Delete("/:id/work-groups/:wgId", wgH.Unassign)

	// ผลงานครู (รายเทอม) + ไฟล์แนบ (signed URL)
	grp.Get("/:id/works", wkH.List)
	grp.Post("/:id/works", wkH.Create)
	grp.Put("/:id/works/:workId", wkH.Update)
	grp.Delete("/:id/works/:workId", wkH.Delete)
	grp.Get("/:id/works/:workId/files", wkH.ListFiles)
	grp.Post("/:id/works/:workId/files", wkH.UploadFile)
	grp.Delete("/:id/works/:workId/files/:fileId", wkH.DeleteFile)
}

func registerHealth(app *fiber.App, deps Deps) {
	// liveness — แค่บอกว่า process ยังตอบได้
	app.Get("/health", func(c *fiber.Ctx) error {
		return httputil.OK(c, fiber.Map{"status": "ok"})
	})

	// readiness — ตรวจ DB/Redis ว่าพร้อมรับ traffic
	app.Get("/ready", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 3*time.Second)
		defer cancel()

		if err := deps.DB.Ping(ctx); err != nil {
			return httputil.Error(c, fiber.StatusServiceUnavailable, "DB_UNAVAILABLE", "ฐานข้อมูลไม่พร้อมใช้งาน")
		}
		if deps.Redis != nil {
			if err := deps.Redis.Ping(ctx).Err(); err != nil {
				return httputil.Error(c, fiber.StatusServiceUnavailable, "REDIS_UNAVAILABLE", "Redis ไม่พร้อมใช้งาน")
			}
		}
		return httputil.OK(c, fiber.Map{"status": "ready"})
	})
}
