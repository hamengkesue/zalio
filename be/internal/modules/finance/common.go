package finance

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// Helper bersama modul finance — sepadan dengan helper di modul product.

// handleDuplicate: unique violation (23505) -> 409 di field yang diberikan.
func handleDuplicate(c *gin.Context, err error, message, field string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		c.JSON(http.StatusConflict, gin.H{"error": message, "field": field})
		return true
	}
	return false
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func ascOrDesc(desc bool) string {
	if desc {
		return "DESC"
	}
	return "ASC"
}

// statusCondCol -> " AND <col> = true/false" dari filter status. Kolom & nilai
// literal (bukan input user), jadi aman dari injeksi.
func statusCondCol(status, col string) string {
	switch status {
	case "active":
		return " AND " + col + " = true"
	case "inactive":
		return " AND " + col + " = false"
	}
	return ""
}

// listParams: pagination + search + sort dari query string (pola sama product).
func listParams(c *gin.Context) (limit, offset int, search, sort string, desc bool) {
	limit = atoiDefault(c.Query("limit"), 8)
	offset = atoiDefault(c.Query("offset"), 0)
	if limit < 1 {
		limit = 8
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset, c.Query("search"), c.Query("sort"), c.Query("desc") == "true"
}
