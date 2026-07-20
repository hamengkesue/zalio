package finance

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"zalio-erp-be/internal/platform/middleware"
)

// Register memasang rute Finance & Accounting — Chart of Account (admin-only).
func Register(protected *gin.RouterGroup, pool *pgxpool.Pool) {
	admin := protected.Group("")
	admin.Use(middleware.RequireRole("admin"))

	ref := &RefHandler{repo: &RefRepo{pool: pool}}
	admin.GET("/coa-classifications", ref.ListClassifications)
	admin.GET("/coa-types", ref.ListTypes)

	coa := &CoaHandler{repo: &CoaRepo{pool: pool}}
	admin.GET("/coa", coa.List)
	admin.GET("/coa-options", coa.Options)
	admin.POST("/coa", coa.Create)
	admin.PUT("/coa/:id", coa.Update)
	admin.PATCH("/coa/:id/toggle-active", coa.ToggleActive)
}
