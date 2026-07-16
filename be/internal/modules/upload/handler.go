package upload

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"zalio-erp-be/internal/platform/storage"
)

const (
	profileFolder = "internal_user_profile_image"
	maxUploadSize = 2 * 1024 * 1024 // 2 MB
)

// Tipe gambar yang diizinkan → ekstensi file.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type Handler struct {
	store *storage.Storage
}

func NewHandler(store *storage.Storage) *Handler {
	return &Handler{store: store}
}

// UploadProfileImage menerima 1 file gambar (form field "file"), memvalidasi
// tipe & ukuran (maks 2 MB), menyimpannya ke MinIO di folder
// internal_user_profile_image/, lalu mengembalikan path-nya.
func (h *Handler) UploadProfileImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if header.Size > maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 2 MB)"})
		return
	}

	contentType := header.Header.Get("Content-Type")
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only image files are allowed (jpg, png, webp, gif)"})
		return
	}

	key := fmt.Sprintf("%s/%s%s", profileFolder, uuid.NewString(), ext)
	_, err = h.store.Client.PutObject(context.Background(), h.store.Bucket, key, file, header.Size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to upload: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"path": key})
}

// ServeFile menstreaming file dari MinIO (publik, agar bisa dipakai <img src>).
func (h *Handler) ServeFile(c *gin.Context) {
	filePath := strings.TrimPrefix(c.Param("filepath"), "/")
	obj, err := h.store.Client.GetObject(context.Background(), h.store.Bucket, filePath, minio.GetObjectOptions{})
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.DataFromReader(http.StatusOK, info.Size, info.ContentType, obj, nil)
}
