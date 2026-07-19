package product

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Model ───

type Subcategory struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CategoryID   string    `json:"category_id"`
	CategoryName string    `json:"category_name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// ─── Repository ───

// Kolom termasuk nama kategori (hasil join) untuk ditampilkan di tabel.
const subcategoryCols = `s.id::text, s.name, s.category_id::text, c.name AS category_name, s.is_active, s.created_at`
const subcategoryFrom = ` FROM m_subcategory s JOIN m_category c ON c.id = s.category_id`

// Kolom RETURNING untuk INSERT/UPDATE — pakai subquery (bukan join) supaya
// baris yang baru diubah tetap terbaca (join ke tabel yang sama di CTE tidak
// melihat perubahan).
const subcategoryReturn = `id::text, name, category_id::text, (SELECT name FROM m_category WHERE id = category_id) AS category_name, is_active, created_at`

type SubcategoryRepo struct{ pool *pgxpool.Pool }

func scanSubcategory(row interface{ Scan(...any) error }) (*Subcategory, error) {
	var s Subcategory
	if err := row.Scan(&s.ID, &s.Name, &s.CategoryID, &s.CategoryName, &s.IsActive, &s.CreatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

// ListPaged: search di nama subkategori; filter opsional categoryID.
func (r *SubcategoryRepo) ListPaged(ctx context.Context, limit, offset int, search, sort string, desc bool, categoryID, status string) ([]Subcategory, int, error) {
	// $1 search, $2 categoryID, $3 limit, $4 offset
	filter := ` WHERE ($1 = '' OR s.name ILIKE '%'||$1||'%') AND ($2 = '' OR s.category_id = $2::uuid)` + statusCondCol(status, "s.is_active")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM m_subcategory s`+filter, search, categoryID).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderCol := "s.created_at"
	dir := "DESC"
	switch sort {
	case "name":
		orderCol = "s.name"
		dir = ascOrDesc(desc)
	case "category":
		orderCol = "c.name"
		dir = ascOrDesc(desc)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+subcategoryCols+subcategoryFrom+filter+` ORDER BY `+orderCol+` `+dir+`, s.name, s.id LIMIT $3 OFFSET $4`,
		search, categoryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Subcategory{}
	for rows.Next() {
		s, err := scanSubcategory(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *s)
	}
	return list, total, rows.Err()
}

func (r *SubcategoryRepo) Create(ctx context.Context, name, categoryID string) (*Subcategory, error) {
	return scanSubcategory(r.pool.QueryRow(ctx,
		`INSERT INTO m_subcategory (name, category_id) VALUES ($1, $2::uuid) RETURNING `+subcategoryReturn,
		name, categoryID))
}

func (r *SubcategoryRepo) Update(ctx context.Context, id, name, categoryID string) (*Subcategory, error) {
	return scanSubcategory(r.pool.QueryRow(ctx,
		`UPDATE m_subcategory SET name = $1, category_id = $2::uuid WHERE id = $3::uuid RETURNING `+subcategoryReturn,
		name, categoryID, id))
}

// ParentCategoryActive: apakah category induk dari subcategory ini aktif?
func (r *SubcategoryRepo) ParentCategoryActive(ctx context.Context, subID string) (bool, error) {
	var active bool
	err := r.pool.QueryRow(ctx,
		`SELECT c.is_active FROM m_subcategory s JOIN m_category c ON c.id = s.category_id WHERE s.id = $1::uuid`,
		subID).Scan(&active)
	return active, err
}

func (r *SubcategoryRepo) ToggleActive(ctx context.Context, id string, active bool) (*Subcategory, error) {
	return scanSubcategory(r.pool.QueryRow(ctx,
		`UPDATE m_subcategory SET is_active = $1 WHERE id = $2::uuid RETURNING `+subcategoryReturn,
		active, id))
}

// ─── Handler ───

type SubcategoryHandler struct{ repo *SubcategoryRepo }

// badCategory mengembalikan true + kirim 400 kalau error karena category_id
// tidak valid / tidak ada (FK violation atau uuid salah).
func badCategory(c *gin.Context, err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "23503" || pgErr.Code == "22P02") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category", "field": "category_id"})
		return true
	}
	return false
}

func (h *SubcategoryHandler) List(c *gin.Context) {
	limit, offset, search, sort, desc := listParams(c)
	items, total, err := h.repo.ListPaged(c.Request.Context(), limit, offset, search, sort, desc, c.Query("category_id"), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func (h *SubcategoryHandler) Create(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		CategoryID string `json:"category_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.repo.Create(c.Request.Context(), req.Name, req.CategoryID)
	if err != nil {
		if handleDuplicate(c, err, "Subcategory name already exists") {
			return
		}
		if badCategory(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *SubcategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name       string `json:"name" binding:"required"`
		CategoryID string `json:"category_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.repo.Update(c.Request.Context(), id, req.Name, req.CategoryID)
	if err != nil {
		if handleDuplicate(c, err, "Subcategory name already exists") {
			return
		}
		if badCategory(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *SubcategoryHandler) ToggleActive(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Blokir aktivasi subcategory kalau category induknya nonaktif.
	if req.IsActive {
		active, err := h.repo.ParentCategoryActive(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !active {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot activate: its category is inactive"})
			return
		}
	}
	item, err := h.repo.ToggleActive(c.Request.Context(), id, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}
