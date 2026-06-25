// Package server ประกอบ Fiber app, middleware กลาง, และ route ทั้งหมด
package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/chumko-platform/backend/internal/auth"
	"github.com/chumko-platform/backend/internal/config"
	"github.com/chumko-platform/backend/internal/crypto"
	"github.com/chumko-platform/backend/internal/handler"
	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/middleware"
	"github.com/chumko-platform/backend/internal/repository"
	"github.com/chumko-platform/backend/internal/service"
)

// Deps คือ dependency ที่ server ต้องใช้ (inject จาก main)
type Deps struct {
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Config *config.Config
	// Cipher ใช้เข้ารหัสข้อมูลอ่อนไหว (เลขบัตรประชาชน) — สร้างจาก main
	Cipher *crypto.Cipher
}

// New สร้าง Fiber app พร้อม middleware กลางและ route พื้นฐาน
func New(deps Deps) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "chumko-platform",
		DisableStartupMessage: true,
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

	return app
}

// registerStudents mount route นักเรียน/ผู้ปกครอง (กลุ่มวิชาการ) — RequireAuth ทุก route
func registerStudents(router fiber.Router, deps Deps, tokens *auth.TokenManager) {
	wgRepo := repository.NewWorkGroupRepository(deps.DB) // ใช้ตรวจสิทธิ์กลุ่มวิชาการ
	studentRepo := repository.NewStudentRepository(deps.DB)
	guardianRepo := repository.NewGuardianRepository(deps.DB)
	linkRepo := repository.NewStudentGuardianRepository(deps.DB)

	sH := handler.NewStudentHandler(service.NewStudentService(studentRepo, wgRepo, deps.Cipher))
	gH := handler.NewGuardianHandler(service.NewGuardianService(guardianRepo, wgRepo, deps.Cipher))
	sgH := handler.NewStudentGuardianHandler(
		service.NewStudentGuardianService(linkRepo, studentRepo, guardianRepo, wgRepo, deps.Cipher))

	authMW := middleware.RequireAuth(tokens)

	sg := router.Group("/students", authMW)
	sg.Get("/", sH.List)
	sg.Post("/", sH.Create)
	sg.Get("/:id", sH.Get)
	sg.Put("/:id", sH.Update)
	sg.Delete("/:id", sH.Delete)
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
	grp.Get("/me", middleware.RequireAuth(tokens), authHandler.Me)
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

	auth := middleware.RequireAuth(tokens)

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
