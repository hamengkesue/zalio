package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"zalio-erp-be/internal/platform/middleware"
	"zalio-erp-be/internal/platform/token"
)

type Handler struct {
	repo *Repo
	tm   *token.Manager
}

// ─── Login ───

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cr, err := h.repo.GetCredentialsByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username", "field": "username"})
		return
	}
	if !cr.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled", "field": "username"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(cr.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password", "field": "password"})
		return
	}

	tok, err := h.tm.Generate(cr.ID, cr.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u, _ := h.repo.GetByID(c.Request.Context(), cr.ID)
	c.JSON(http.StatusOK, gin.H{"token": tok, "user": u})
}

// ─── Current user ───

func (h *Handler) Me(c *gin.Context) {
	uid := c.GetString(middleware.CtxUserID)
	u, err := h.repo.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": u})
}

// ─── User management (admin only) ───

func (h *Handler) ListUsers(c *gin.Context) {
	limit := atoiDefault(c.Query("limit"), 8)
	offset := atoiDefault(c.Query("offset"), 0)
	if limit < 1 || limit > 100 {
		limit = 8
	}
	if offset < 0 {
		offset = 0
	}
	// Filter role & status (di-whitelist; nilai lain diabaikan → tanpa filter).
	role := c.Query("role")
	if role != "admin" && role != "staff" {
		role = ""
	}
	status := c.Query("status")
	if status != "active" && status != "inactive" {
		status = ""
	}

	items, total, err := h.repo.ListPaged(c.Request.Context(), ListParams{
		Limit:  limit,
		Offset: offset,
		Search: c.Query("search"),
		Sort:   c.Query("sort"),
		Desc:   c.Query("desc") == "true",
		Role:   role,
		Status: status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

// atoiDefault mengubah string query jadi int; kalau kosong/invalid pakai def.
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Username string `json:"username" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		Whatsapp     string `json:"whatsapp"`
		ProfileImage string `json:"profile_image"`
		GroupAccess  string `json:"group_access"`
		Password     string `json:"password" binding:"required,min=6"`
		Role         string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != "admin" && req.Role != "staff" {
		req.Role = "staff"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	u, err := h.repo.Create(c.Request.Context(), req.Name, req.Username, req.Email, req.Whatsapp, req.ProfileImage, req.GroupAccess, string(hash), req.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			switch {
			case strings.Contains(pgErr.ConstraintName, "username"):
				c.JSON(http.StatusConflict, gin.H{"error": "Username already taken", "field": "username"})
			case strings.Contains(pgErr.ConstraintName, "email"):
				c.JSON(http.StatusConflict, gin.H{"error": "Email already taken", "field": "email"})
			default:
				c.JSON(http.StatusConflict, gin.H{"error": "Username or email already taken"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": u})
}

// UpdateUser mengubah data user. Password opsional (kosong = tidak diubah).
// Username tidak bisa diubah.
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id") // UUID
	var req struct {
		Name         string `json:"name" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		Whatsapp     string `json:"whatsapp"`
		ProfileImage string `json:"profile_image"`
		GroupAccess  string `json:"group_access"`
		Password     string `json:"password"`
		Role         string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != "admin" && req.Role != "staff" {
		req.Role = "staff"
	}
	if req.Password != "" && len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}

	u, err := h.repo.Update(c.Request.Context(), id, req.Name, req.Email, req.Whatsapp, req.ProfileImage, req.GroupAccess, req.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // email unik
			c.JSON(http.StatusConflict, gin.H{"error": "Email already taken", "field": "email"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := h.repo.UpdatePassword(c.Request.Context(), id, string(hash)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": u})
}

func (h *Handler) ToggleUserActive(c *gin.Context) {
	id := c.Param("id") // UUID
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.repo.ToggleActive(c.Request.Context(), id, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": u})
}
