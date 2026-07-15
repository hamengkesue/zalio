package ping

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register merakit modul ping: repo -> handler -> daftar route.
// INILAH pola yang ditiru setiap modul baru. Cukup panggil
// ping.Register(api, pool) di main.go, dan modul ini "colok-pasang".
func Register(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	repo := NewRepo(pool)
	h := NewHandler(repo)

	rg.GET("/ping", h.List)
	rg.POST("/ping", h.Create)
}
