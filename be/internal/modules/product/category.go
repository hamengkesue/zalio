package product

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Model ───

type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	BannerImage string    `json:"banner_image"`
	InUse       bool      `json:"in_use"` // true kalau sudah dipakai subcategory (nama dikunci)
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── Repository ───

const categoryCols = `id::text, name, COALESCE(banner_image, '') AS banner_image, ` +
	`EXISTS(SELECT 1 FROM m_subcategory s WHERE s.category_id = m_category.id) AS in_use, ` +
	`is_active, created_at`

type CategoryRepo struct{ pool *pgxpool.Pool }

func scanCategory(row interface{ Scan(...any) error }) (*Category, error) {
	var c Category
	if err := row.Scan(&c.ID, &c.Name, &c.BannerImage, &c.InUse, &c.IsActive, &c.CreatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepo) ListPaged(ctx context.Context, limit, offset int, search, sort string, desc bool, status string) ([]Category, int, error) {
	filter := ` WHERE ($1 = '' OR name ILIKE '%'||$1||'%')` + statusCondCol(status, "is_active")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM m_category`+filter, search).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderCol := "created_at"
	dir := "DESC"
	if sort == "name" {
		orderCol = "name"
		dir = ascOrDesc(desc)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+categoryCols+` FROM m_category`+filter+` ORDER BY `+orderCol+` `+dir+`, id LIMIT $2 OFFSET $3`,
		search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *c)
	}
	return list, total, rows.Err()
}

func (r *CategoryRepo) Create(ctx context.Context, name, bannerImage string) (*Category, error) {
	return scanCategory(r.pool.QueryRow(ctx,
		`INSERT INTO m_category (name, banner_image)
		 VALUES ($1, NULLIF($2, ''))
		 RETURNING `+categoryCols,
		name, bannerImage))
}

func (r *CategoryRepo) Update(ctx context.Context, id, name, bannerImage string) (*Category, error) {
	// Nama dikunci kalau category sudah dipakai subcategory: pertahankan nama lama.
	return scanCategory(r.pool.QueryRow(ctx,
		`UPDATE m_category SET
		   name = CASE WHEN EXISTS(SELECT 1 FROM m_subcategory s WHERE s.category_id = m_category.id)
		               THEN name ELSE $1 END,
		   banner_image = NULLIF($2, '')
		 WHERE id = $3::uuid
		 RETURNING `+categoryCols,
		name, bannerImage, id))
}

func (r *CategoryRepo) ToggleActive(ctx context.Context, id string, active bool) (*Category, error) {
	return scanCategory(r.pool.QueryRow(ctx,
		`UPDATE m_category SET is_active = $1 WHERE id = $2::uuid RETURNING `+categoryCols,
		active, id))
}

// ─── Handler ───

type CategoryHandler struct{ repo *CategoryRepo }

func (h *CategoryHandler) List(c *gin.Context) {
	limit, offset, search, sort, desc := listParams(c)
	items, total, err := h.repo.ListPaged(c.Request.Context(), limit, offset, search, sort, desc, c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		BannerImage string `json:"banner_image"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.repo.Create(c.Request.Context(), req.Name, req.BannerImage)
	if err != nil {
		if handleDuplicate(c, err, "Category name already exists") {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name" binding:"required"`
		BannerImage string `json:"banner_image"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.repo.Update(c.Request.Context(), id, req.Name, req.BannerImage)
	if err != nil {
		if handleDuplicate(c, err, "Category name already exists") {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *CategoryHandler) ToggleActive(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.repo.ToggleActive(c.Request.Context(), id, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}
