// Command api คือ entrypoint ของ HTTP server
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chumko-platform/backend/internal/config"
	"github.com/chumko-platform/backend/internal/crypto"
	"github.com/chumko-platform/backend/internal/database"
	"github.com/chumko-platform/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("โหลด config ไม่สำเร็จ: %v", err)
	}

	ctx := context.Background()

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("เชื่อมต่อ PostgreSQL ไม่สำเร็จ: %v", err)
	}
	defer db.Close()

	rdb, err := database.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("เชื่อมต่อ Redis ไม่สำเร็จ: %v", err)
	}
	if rdb != nil {
		defer func() { _ = rdb.Close() }()
	}

	cipher, err := crypto.NewCipher(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("สร้าง cipher ไม่สำเร็จ (ตรวจ ENCRYPTION_KEY): %v", err)
	}

	app := server.New(server.Deps{DB: db, Redis: rdb, Config: cfg, Cipher: cipher})

	// start server แบบ non-blocking เพื่อรองรับ graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.HTTPPort); err != nil {
			log.Fatalf("server หยุดทำงาน: %v", err)
		}
	}()
	log.Printf("chumko-platform backend ทำงานที่พอร์ต %s (env=%s)", cfg.HTTPPort, cfg.AppEnv)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("กำลังปิด server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("ปิด server ไม่เรียบร้อย: %v", err)
	}
}
