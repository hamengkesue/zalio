package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"zalio-erp-be/internal/platform/token"
)

// Kunci untuk menaruh data user di context request.
const (
	CtxUserID = "userID"
	CtxRole   = "role"
)

// RequireAuth memblokir request yang tidak membawa token valid.
// Kalau lolos, ID & role user disimpan di context untuk dipakai handler.
func RequireAuth(tm *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token tidak ada"})
			return
		}
		claims, err := tm.Parse(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid atau kedaluwarsa"})
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// RequireRole memastikan user punya role tertentu (mis. "admin").
// Dipakai setelah RequireAuth.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(CtxRole) != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "akses ditolak: butuh role " + role})
			return
		}
		c.Next()
	}
}
