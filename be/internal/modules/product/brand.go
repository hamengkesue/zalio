package product

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Model ───

type Brand struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Logo        string    `json:"logo"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── Repository ───

const brandCols = `id::text, name, COALESCE(description, '') AS description, COALESCE(logo, '') AS logo, is_active, created_at`

type BrandRepo struct{ pool *pgxpool.Pool }

func scanBrand(row interface{ Scan(...any) error }) (*Brand, error) {
	var b Brand
	if err := row.Scan(&b.ID, &b.Name, &b.Description, &b.Logo, &b.IsActive, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BrandRepo) ListPaged(ctx context.Context, limit, offset int, search, sort string, desc bool, status string) ([]Brand, int, error) {
	filter := ` WHERE ($1 = '' OR name ILIKE '%'||$1||'%')` + statusCondCol(status, "is_active")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM m_brand`+filter, search).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderCol := "created_at"
	dir := "DESC"
	if sort == "name" {
		orderCol = "name"
		dir = ascOrDesc(desc)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+brandCols+` FROM m_brand`+filter+` ORDER BY `+orderCol+` `+dir+`, id LIMIT $2 OFFSET $3`,
		search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Brand{}
	for rows.Next() {
		b, err := scanBrand(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *b)
	}
	return list, total, rows.Err()
}

func (r *BrandRepo) Create(ctx context.Context, name, description, logo string) (*Brand, error) {
	return scanBrand(r.pool.QueryRow(ctx,
		`INSERT INTO m_brand (name, description, logo)
		 VALUES ($1, NULLIF($2, ''), NULLIF($3, ''))
		 RETURNING `+brandCols,
		name, description, logo))
}

func (r *BrandRepo) Update(ctx context.Context, id, name, description, logo string) (*Brand, error) {
	return scanBrand(r.pool.QueryRow(ctx,
		`UPDATE m_brand SET name = $1, description = NULLIF($2, ''), logo = NULLIF($3, '')
		 WHERE id = $4::uuid
		 RETURNING `+brandCols,
		name, description, logo, id))
}

func (r *BrandRepo) ToggleActive(ctx context.Context, id string, active bool) (*Brand, error) {
	return scanBrand(r.pool.QueryRow(ctx,
		`UPDATE m_brand SET is_active = $1 WHERE id = $2::uuid RETURNING `+brandCols,
		active, id))
}

// ─── Handler ───

type BrandHandler struct{ repo *BrandRepo }

func (h *BrandHandler) List(c *gin.Context) {
	limit, offset, search, sort, desc := listParams(c)
	items, total, err := h.repo.ListPaged(c.Request.Context(), limit, offset, search, sort, desc, c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func (h *BrandHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Logo        string `json:"logo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.repo.Create(c.Request.Context(), req.Name, req.Description, req.Logo)
	if err != nil {
		if handleDuplicate(c, err, "Brand name already exists") {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": b})
}

func (h *BrandHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Logo        string `json:"logo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.repo.Update(c.Request.Context(), id, req.Name, req.Description, req.Logo)
	if err != nil {
		if handleDuplicate(c, err, "Brand name already exists") {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": b})
}

func (h *BrandHandler) ToggleActive(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.repo.ToggleActive(c.Request.Context(), id, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": b})
}
