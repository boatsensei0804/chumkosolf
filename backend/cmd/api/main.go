// Command api คือ entrypoint ของ HTTP server
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chumkosoft/backend/internal/config"
	"github.com/chumkosoft/backend/internal/crypto"
	"github.com/chumkosoft/backend/internal/database"
	"github.com/chumkosoft/backend/internal/server"
	"github.com/chumkosoft/backend/internal/storage"
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

	// object storage สำหรับไฟล์แนบ — optional (ถ้าไม่ตั้งค่าจะใช้งานไฟล์แนบไม่ได้แต่ระบบยังขึ้นได้)
	var fileStore storage.Storage
	if cfg.Storage.Enabled() {
		ms, err := storage.NewMinioStorage(ctx, cfg.Storage)
		if err != nil {
			log.Fatalf("เชื่อมต่อ object storage ไม่สำเร็จ: %v", err)
		}
		fileStore = ms
		log.Printf("object storage พร้อมใช้งาน (bucket=%s)", cfg.Storage.Bucket)
	} else {
		log.Println("object storage ไม่ได้ตั้งค่า — ฟีเจอร์ไฟล์แนบจะใช้งานไม่ได้")
	}

	app := server.New(server.Deps{DB: db, Redis: rdb, Config: cfg, Cipher: cipher, Storage: fileStore})

	// start server แบบ non-blocking เพื่อรองรับ graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.HTTPPort); err != nil {
			log.Fatalf("server หยุดทำงาน: %v", err)
		}
	}()
	log.Printf("chumkosoft backend ทำงานที่พอร์ต %s (env=%s)", cfg.HTTPPort, cfg.AppEnv)

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
