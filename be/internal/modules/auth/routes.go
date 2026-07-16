package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"zalio-erp-be/internal/platform/middleware"
	"zalio-erp-be/internal/platform/token"
)

// Register memasang rute auth:
//   - public   : POST /auth/login (tanpa token)
//   - protected: GET  /auth/me    (butuh login)
//   - admin    : /users ...       (butuh login + role admin)
func Register(public, protected *gin.RouterGroup, pool *pgxpool.Pool, tm *token.Manager) {
	h := &Handler{repo: NewRepo(pool), tm: tm}

	public.POST("/auth/login", h.Login)

	protected.GET("/auth/me", h.Me)

	admin := protected.Group("")
	admin.Use(middleware.RequireRole("admin"))
	admin.GET("/users", h.ListUsers)
	admin.POST("/users", h.CreateUser)
	admin.PUT("/users/:id", h.UpdateUser)
	admin.PATCH("/users/:id/toggle-active", h.ToggleUserActive)
}
