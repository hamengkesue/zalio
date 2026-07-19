package product

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Country = data referensi negara (m_country) untuk dropdown Country of Origin.
type Country struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"` // ISO 3166-1 alpha-2 (mis. "ID") — dipakai untuk bendera
	PhoneCode string `json:"phone_code"`
}

type CountryRepo struct{ pool *pgxpool.Pool }

// List mengembalikan semua negara aktif, urut nama.
func (r *CountryRepo) List(ctx context.Context) ([]Country, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::text, country_name, country_code, COALESCE(phone_code,'')
		 FROM m_country WHERE is_active ORDER BY country_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Country{}
	for rows.Next() {
		var c Country
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.PhoneCode); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

type CountryHandler struct{ repo *CountryRepo }

func (h *CountryHandler) List(c *gin.Context) {
	items, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
