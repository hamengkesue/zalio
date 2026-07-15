package auth

import (
	"errors"
	"net/http"
	"strconv"

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
		// username tidak ditemukan — pesan disamakan agar tidak membocorkan info
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	if !cr.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(cr.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
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
	uid := c.GetInt(middleware.CtxUserID)
	u, err := h.repo.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": u})
}

// ─── User management (admin only) ───

func (h *Handler) ListUsers(c *gin.Context) {
	items, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role"`
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

	u, err := h.repo.Create(c.Request.Context(), req.Name, req.Username, req.Email, string(hash), req.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			c.JSON(http.StatusConflict, gin.H{"error": "username or email already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": u})
}

func (h *Handler) ToggleUserActive(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
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
