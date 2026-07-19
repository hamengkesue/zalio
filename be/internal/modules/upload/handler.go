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
	maxUploadSize = 5 * 1024 * 1024 // 5 MB
)

// Folder tujuan yang diizinkan untuk endpoint upload gambar generik.
var allowedFolders = map[string]bool{
	"brand_logo":      true,
	"category_banner": true,
	"product_image":   true,
}

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

// UploadProfileImage menyimpan foto profil user internal ke folder tetap.
func (h *Handler) UploadProfileImage(c *gin.Context) {
	h.uploadTo(c, profileFolder)
}

// UploadImage endpoint generik: folder ditentukan lewat query ?folder=...,
// dibatasi whitelist (brand_logo / category_banner / product_image).
func (h *Handler) UploadImage(c *gin.Context) {
	folder := c.Query("folder")
	if !allowedFolders[folder] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing folder"})
		return
	}
	h.uploadTo(c, folder)
}

// uploadTo menerima 1 file gambar (form field "file"), memvalidasi tipe &
// ukuran (maks 2 MB), menyimpannya ke MinIO di folder yang diberikan, lalu
// mengembalikan path-nya.
func (h *Handler) uploadTo(c *gin.Context, folder string) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if header.Size > maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 5 MB)"})
		return
	}

	contentType := header.Header.Get("Content-Type")
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only image files are allowed (jpg, png, webp, gif)"})
		return
	}

	key := fmt.Sprintf("%s/%s%s", folder, uuid.NewString(), ext)
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
