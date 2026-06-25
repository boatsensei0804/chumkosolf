package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/auth"
	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/tenant"
)

// localIsSchoolAdmin คือ key ใน fiber locals สำหรับสถานะ school admin
const localIsSchoolAdmin = "is_school_admin"

// RequireAuth ตรวจ Bearer access token แล้วฝัง tenant.Identity ลง context
// (school_id/semester_id/role/user_id มาจาก token เท่านั้น — ห้ามรับจาก client)
func RequireAuth(tm *auth.TokenManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := bearerToken(c.Get(fiber.HeaderAuthorization))
		if !ok {
			return httputil.Error(c, domain.ErrUnauthorized.Status, domain.ErrUnauthorized.Code, domain.ErrUnauthorized.Message)
		}

		claims, err := tm.Parse(token)
		if err != nil {
			return httputil.Error(c, domain.ErrInvalidToken.Status, domain.ErrInvalidToken.Code, domain.ErrInvalidToken.Message)
		}

		ctx := tenant.WithIdentity(c.UserContext(), tenant.Identity{
			UserID:        claims.Subject,
			SchoolID:      claims.SchoolID,
			SemesterID:    claims.SemesterID,
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
