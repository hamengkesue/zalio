package product

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// Helper bersama untuk semua master produk (brand/category/subcategory/uom).

// handleDuplicate: kalau err adalah unique violation (23505), balas 409 dengan
// pesan di field "name" dan kembalikan true. Selain itu kembalikan false.
func handleDuplicate(c *gin.Context, err error, message string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		c.JSON(http.StatusConflict, gin.H{"error": message, "field": "name"})
		return true
	}
	return false
}

// atoiDefault mengubah string query jadi int; kosong/invalid pakai def.
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

// ascOrDesc: true -> "DESC", false -> "ASC".
func ascOrDesc(desc bool) string {
	if desc {
		return "DESC"
	}
	return "ASC"
}

// statusCondCol menghasilkan potongan SQL " AND <col> = true/false" dari filter
// status ("active"/"inactive"). Aman dari injeksi (col & nilai literal, bukan input user).
func statusCondCol(status, col string) string {
	switch status {
	case "active":
		return " AND " + col + " = true"
	case "inactive":
		return " AND " + col + " = false"
	}
	return ""
}

// listParams membaca parameter pagination + search + sort dari query string.
// Pola sama dengan endpoint Internal Users.
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
