package product

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"zalio-erp-be/internal/platform/middleware"
)

// Register memasang rute master produk (brand/category/subcategory/uom).
// Semua admin-only, di bawah grup terproteksi (butuh login + role admin).
func Register(protected *gin.RouterGroup, pool *pgxpool.Pool) {
	// Referensi negara — cukup butuh login (dipakai dropdown Country of Origin).
	country := &CountryHandler{repo: &CountryRepo{pool: pool}}
	protected.GET("/countries", country.List)

	admin := protected.Group("")
	admin.Use(middleware.RequireRole("admin"))

	brand := &BrandHandler{repo: &BrandRepo{pool: pool}}
	admin.GET("/brands", brand.List)
	admin.POST("/brands", brand.Create)
	admin.PUT("/brands/:id", brand.Update)
	admin.PATCH("/brands/:id/toggle-active", brand.ToggleActive)

	category := &CategoryHandler{repo: &CategoryRepo{pool: pool}}
	admin.GET("/categories", category.List)
	admin.POST("/categories", category.Create)
	admin.PUT("/categories/:id", category.Update)
	admin.PATCH("/categories/:id/toggle-active", category.ToggleActive)

	subcategory := &SubcategoryHandler{repo: &SubcategoryRepo{pool: pool}}
	admin.GET("/subcategories", subcategory.List)
	admin.POST("/subcategories", subcategory.Create)
	admin.PUT("/subcategories/:id", subcategory.Update)
	admin.PATCH("/subcategories/:id/toggle-active", subcategory.ToggleActive)

	uom := &UomHandler{repo: &UomRepo{pool: pool}}
	admin.GET("/uoms", uom.List)
	admin.POST("/uoms", uom.Create)
	admin.PUT("/uoms/:id", uom.Update)
	admin.PATCH("/uoms/:id/toggle-active", uom.ToggleActive)

	prod := &ProductHandler{repo: &ProductRepo{pool: pool}}
	admin.GET("/products", prod.List)
	admin.GET("/product-next-sku", prod.NextSku)
	admin.GET("/products/:id", prod.Get)
	admin.POST("/products", prod.Create)
	admin.PUT("/products/:id", prod.Update)
	admin.PATCH("/products/:id/toggle-active", prod.ToggleActive)
	admin.PATCH("/product-variants/:id/toggle-active", prod.ToggleVariantActive)
	admin.GET("/product-import/template", prod.ImportTemplate)
	admin.GET("/product-import/existing", prod.ImportExisting)
	admin.POST("/product-import/validate", prod.ImportValidate)
	admin.POST("/product-import/commit", prod.ImportCommit)
	admin.POST("/product-import/failed", prod.ImportFailed)
}
