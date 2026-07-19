package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"zalio-erp-be/internal/modules/auth"
	"zalio-erp-be/internal/modules/finance"
	"zalio-erp-be/internal/modules/ping"
	"zalio-erp-be/internal/modules/product"
	"zalio-erp-be/internal/modules/upload"
	"zalio-erp-be/internal/platform/config"
	"zalio-erp-be/internal/platform/database"
	"zalio-erp-be/internal/platform/middleware"
	"zalio-erp-be/internal/platform/storage"
	"zalio-erp-be/internal/platform/token"
)

func main() {
	cfg := config.Load()

	pool, err := database.NewPool(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Buat admin default kalau belum ada user (biar bisa login pertama kali).
	if err := auth.EnsureSeed(context.Background(), pool); err != nil {
		log.Fatalf("Failed to seed default admin: %v", err)
	}

	tm := token.NewManager(cfg.JWTSecret)

	store, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
	}))

	api := r.Group("/api/v1")

	// Health check — dipakai untuk memastikan server hidup.
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "zalio-erp"})
	})

	// Grup terproteksi: semua rute di bawahnya butuh token login valid.
	protected := api.Group("")
	protected.Use(middleware.RequireAuth(tm))

	// ─── Modules ───
	// Rute publik + terproteksi. Modul baru (Fase 2 dst.) daftar di `protected`.
	auth.Register(api, protected, pool, tm)
	ping.Register(api, pool)               // demo Fase 0 — dibiarkan publik
	upload.Register(r, protected, store)   // /files/* (publik) + /upload/profile-image (login)
	product.Register(protected, pool)      // master produk (brand/category/subcategory/uom) — admin-only
	finance.Register(protected, pool)      // Finance & Accounting — Chart of Account (admin-only)

	log.Printf("Server starting on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
