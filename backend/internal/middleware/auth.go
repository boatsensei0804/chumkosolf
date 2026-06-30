package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/chumkosoft/backend/internal/auth"
	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/tenant"
)

// localIsSchoolAdmin คือ key ใน fiber locals สำหรับสถานะ school admin
const localIsSchoolAdmin = "is_school_admin"

// semesterHeader คือ header ที่ผู้ใช้ส่งมาเพื่อ "สลับเทอมทำงาน" (override)
const semesterHeader = "X-Semester-Id"

// SemesterVerifier ตรวจว่า semester อยู่ในโรงเรียนที่ระบุ (กันสลับไปเทอมของโรงเรียนอื่น)
type SemesterVerifier interface {
	SemesterInSchool(ctx context.Context, schoolID, semesterID string) (bool, error)
}

// RequireAuth ตรวจ Bearer access token แล้วฝัง tenant.Identity ลง context
// school_id/role/user_id มาจาก token เท่านั้น (ห้ามรับจาก client)
// semester_id: ปกติมาจาก token; ถ้ามี header X-Semester-Id จะ override ได้ก็ต่อเมื่อ
// เทอมนั้นอยู่ในโรงเรียนเดียวกับ token (validate กับ school_id จาก token เสมอ)
func RequireAuth(tm *auth.TokenManager, sem SemesterVerifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := bearerToken(c.Get(fiber.HeaderAuthorization))
		if !ok {
			return httputil.Error(c, domain.ErrUnauthorized.Status, domain.ErrUnauthorized.Code, domain.ErrUnauthorized.Message)
		}

		claims, err := tm.Parse(token)
		if err != nil {
			return httputil.Error(c, domain.ErrInvalidToken.Status, domain.ErrInvalidToken.Code, domain.ErrInvalidToken.Message)
		}

		semesterID := claims.SemesterID
		if override := strings.TrimSpace(c.Get(semesterHeader)); override != "" {
			if _, perr := uuid.Parse(override); perr != nil {
				return httputil.Error(c, fiber.StatusBadRequest, "INVALID_SEMESTER", "รหัสเทอมไม่ถูกต้อง")
			}
			inSchool, verr := sem.SemesterInSchool(c.UserContext(), claims.SchoolID, override)
			if verr != nil {
				return httputil.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "เกิดข้อผิดพลาดภายในระบบ")
			}
			if !inSchool {
				return httputil.Error(c, fiber.StatusForbidden, "INVALID_SEMESTER", "ไม่พบเทอมนี้ในโรงเรียนของคุณ")
			}
			semesterID = override
		}

		ctx := tenant.WithIdentity(c.UserContext(), tenant.Identity{
			UserID:        claims.Subject,
			SchoolID:      claims.SchoolID,
			SemesterID:    semesterID,
			Role:          claims.Role,
			IsSchoolAdmin: claims.IsSchoolAdmin,
			IPAddress:     c.IP(),
		})
		c.SetUserContext(ctx)
		c.Locals(localIsSchoolAdmin, claims.IsSchoolAdmin)

		return c.Next()
	}
}

// bearerToken ดึง token จาก header "Authorization: Bearer <token>"
func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}
