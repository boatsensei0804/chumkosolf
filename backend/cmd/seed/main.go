// Command seed สร้าง super admin user ตั้งต้นสำหรับโรงเรียน dev
// (แยกจาก migration เพราะ password ต้อง bcrypt hash จริง ไม่ hardcode ใน SQL)
//
// ใช้งาน: DATABASE_URL=... SEED_ADMIN_PASSWORD=... go run ./cmd/seed
package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/chumko-platform/backend/internal/database"
)

// โรงเรียน dev (ตรงกับ migration 000099_seed_initial)
const devSchoolID = "00000000-0000-0000-0000-000000000001"

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("ต้องตั้ง DATABASE_URL")
	}
	username := getenv("SEED_ADMIN_USERNAME", "superadmin")
	password := getenv("SEED_ADMIN_PASSWORD", "admin1234")

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("เชื่อมต่อ DB ไม่สำเร็จ: %v", err)
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash password ไม่สำเร็จ: %v", err)
	}

	const q = `
		INSERT INTO users (school_id, username, password_hash, role, is_school_admin, is_active)
		VALUES ($1, $2, $3, 'super_admin', TRUE, TRUE)
		ON CONFLICT (school_id, username) DO UPDATE
		    SET password_hash = EXCLUDED.password_hash
		RETURNING id`

	var id string
	err = pool.QueryRow(ctx, q, devSchoolID, username, string(hash)).Scan(&id)
	if err != nil && err != pgx.ErrNoRows {
		log.Fatalf("seed super admin ไม่สำเร็จ: %v", err)
	}

	log.Printf("seed super admin สำเร็จ (username=%s, id=%s)", username, id)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
