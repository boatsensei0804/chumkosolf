// Package middleware รวม middleware กลาง: auth, tenant context, logging, recovery, CORS
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Common คืน middleware พื้นฐานที่ทุก request ต้องผ่าน
// (recover กัน panic, requestid, logger, cors)
func Common() []fiber.Handler {
	return []fiber.Handler{
		recover.New(),
		requestid.New(),
		logger.New(logger.Config{
			// ไม่ log body เพื่อกันข้อมูลส่วนบุคคลรั่ว (PDPA)
			Format: "${time} ${locals:requestid} ${status} ${method} ${path} ${latency}\n",
		}),
		cors.New(),
	}
}
