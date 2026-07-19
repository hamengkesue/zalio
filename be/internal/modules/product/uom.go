package product

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Model ───

type Uom struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── Repository ───

const uomCols = `id::text, name, COALESCE(description, '') AS description, is_active, created_at`

type UomRepo struct{ pool *pgxpool.Pool }

func scanUom(row interface{ Scan(...any) error }) (*Uom, error) {
	var u Uom
	if err := row.Scan(&u.ID, &u.Name, &u.Description, &u.IsActive, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UomRepo) ListPaged(ctx context.Context, limit, offset int, search, sort string, desc bool, status string) ([]Uom, int, error) {
	filter := ` WHERE ($1 = '' OR name ILIKE '%'||$1||'%')` + statusCondCol(status, "is_active")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM m_uom`+filter, search).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderCol := "created_at"
	dir := "DESC"
	if sort == "name" {
		orderCol = "name"
		dir = ascOrDesc(desc)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+uomCols+` FROM m_uom`+filter+` ORDER BY `+orderCol+` `+dir+`, id LIMIT $2 OFFSET $3`,
		search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Uom{}
	for rows.Next() {
		u, err := scanUom(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *u)
	}
	return list, total, rows.Err()
}

func (r *UomRepo) Create(ctx context.Context, name, description string) (*Uom, error) {
	return scanUom(r.pool.QueryRow(ctx,
		`INSERT INTO m_uom (name, description) VALUES ($1, NULLIF($2, '')) RETURNING `+uomCols,
		name, description))
}

func (r *UomRepo) Update(ctx context.Context, id, name, description string) (*Uom, error) {
	return scanUom(r.pool.QueryRow(ctx,
		`UPDATE m_uom SET name = $1, description = NULLIF($2, '') WHERE id = $3::uuid RETURNING `+uomCols,
		name, description, id))
}

func (r *UomRepo) ToggleActive(ctx context.Context, id string, active bool) (*Uom, error) {
	return scanUom(r.pool.QueryRow(ctx,
		`UPDATE m_uom SET is_active = $1 WHERE id = $2::uuid RETURNING `+uomCols,
		active, id))
}

// ─── Handler ───

type UomHandler struct{ repo *UomRepo }

func (h *UomHandler) List(c *gin.Context) {
	limit, offset, search, sort, desc := listParams(c)
	items, total, err := h.repo.ListPaged(c.Request.Context(), limit, offset, search, sort, desc, c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func (h *UomHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.repo.Create(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		if handleDuplicate(c, err, "Unit name already exists") {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *UomHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.repo.Update(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		if handleDuplicate(c, err, "Unit name already exists") {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *UomHandler) ToggleActive(c *gin.Context) {
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
