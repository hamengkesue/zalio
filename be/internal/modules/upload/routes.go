package upload

import (
	"github.com/gin-gonic/gin"

	"zalio-erp-be/internal/platform/storage"
)

// Register memasang:
//   - GET  /files/*filepath             (publik, di root — agar <img src> bisa akses)
//   - POST /api/v1/upload/profile-image (butuh login)
//   - POST /api/v1/upload/image?folder= (butuh login; folder: brand_logo/category_banner/product_image)
func Register(root *gin.Engine, protected *gin.RouterGroup, store *storage.Storage) {
	h := NewHandler(store)
	root.GET("/files/*filepath", h.ServeFile)
	protected.POST("/upload/profile-image", h.UploadProfileImage)
	protected.POST("/upload/image", h.UploadImage)
}
