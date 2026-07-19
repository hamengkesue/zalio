package finance

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"zalio-erp-be/internal/platform/middleware"
)

// ─── Model ───

type Coa struct {
	ID                 string    `json:"id"`
	AccountCode        string    `json:"account_code"`
	AccountTypeCode    string    `json:"account_type_code"`
	AccountTypeName    string    `json:"account_type_name"`
	ClassificationName string    `json:"classification_name"`
	ReportName         string    `json:"report_name"`
	AccountName        string    `json:"account_name"`
	IsContra           bool      `json:"is_contra"`
	IsCreditAccount    bool      `json:"is_credit_account"`
	OpeningBalance     float64   `json:"opening_balance"`
	OpeningDate        *string   `json:"opening_date"`
	Notes              string    `json:"notes"`
	IsActive           bool      `json:"is_active"`
	InUse              bool      `json:"in_use"` // true = dipakai referensi/transaksi -> field inti dikunci
	CreatedAt          time.Time `json:"created_at"`
}

// ─── Repository ───

// coaInUseExpr: apakah akun sudah dipakai sebagai referensi/transaksi di modul lain.
// Saat ini yang bisa mereferensikan akun hanyalah kolom coa_* di m_product.
// (Nanti bertambah, mis. journal entry, seiring modul baru dibangun.)
const coaInUseExpr = `EXISTS(SELECT 1 FROM m_product p WHERE a.id IN (
	p.coa_inventory, p.coa_sales, p.coa_sales_return, p.coa_sales_discount,
	p.coa_good_in_transit, p.coa_cogs, p.coa_purchase_return, p.coa_unbilled_goods))`

const coaSelect = `SELECT a.id::text, a.account_code, a.account_type_code, t.account_type_name,
	c.classification_name, c.report_name, a.account_name, a.is_contra, a.is_credit_account,
	a.opening_balance::float8, to_char(a.opening_date, 'YYYY-MM-DD'), COALESCE(a.notes, ''),
	a.is_active, a.created_at, ` + coaInUseExpr + ` AS in_use`
const coaFrom = ` FROM m_coa a
	JOIN m_coa_type t ON t.account_type_code = a.account_type_code
	JOIN m_coa_classification c ON c.classification_code = t.classification_code`

type CoaRepo struct{ pool *pgxpool.Pool }

func scanCoa(row pgx.Row) (*Coa, error) {
	var x Coa
	if err := row.Scan(&x.ID, &x.AccountCode, &x.AccountTypeCode, &x.AccountTypeName,
		&x.ClassificationName, &x.ReportName, &x.AccountName, &x.IsContra, &x.IsCreditAccount,
		&x.OpeningBalance, &x.OpeningDate, &x.Notes, &x.IsActive, &x.CreatedAt, &x.InUse); err != nil {
		return nil, err
	}
	return &x, nil
}

var coaSortCols = map[string]string{
	"code": "a.account_code", "account_code": "a.account_code",
	"name": "a.account_name", "account_name": "a.account_name",
	"type": "t.account_type_name", "account_type": "t.account_type_name",
	"classification": "c.classification_name",
	"balance":        "a.opening_balance", "opening_balance": "a.opening_balance",
	"status": "a.is_active",
}

// ListPaged: search di kode + nama; filter account_type, classification, status.
func (r *CoaRepo) ListPaged(ctx context.Context, limit, offset int, search, sort string, desc bool, accountType, classification, status string) ([]Coa, int, error) {
	// $1 search, $2 account_type_code, $3 classification_code
	filter := ` WHERE ($1 = '' OR a.account_code ILIKE '%'||$1||'%' OR a.account_name ILIKE '%'||$1||'%')` +
		` AND ($2 = '' OR a.account_type_code = $2)` +
		` AND ($3 = '' OR t.classification_code::text = $3)` +
		statusCondCol(status, "a.is_active")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*)`+coaFrom+filter, search, accountType, classification).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderCol := "a.account_code"
	dir := "ASC"
	if col, ok := coaSortCols[sort]; ok {
		orderCol = col
		dir = ascOrDesc(desc)
	}

	rows, err := r.pool.Query(ctx,
		coaSelect+coaFrom+filter+` ORDER BY `+orderCol+` `+dir+`, a.account_code LIMIT $4 OFFSET $5`,
		search, accountType, classification, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Coa{}
	for rows.Next() {
		x, err := scanCoa(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *x)
	}
	return list, total, rows.Err()
}

func (r *CoaRepo) getOne(ctx context.Context, id string) (*Coa, error) {
	return scanCoa(r.pool.QueryRow(ctx, coaSelect+coaFrom+` WHERE a.id = $1::uuid`, id))
}

// typeIsCredit: saldo normal default tipe. errNoType kalau tipe tidak ada.
var errNoType = errors.New("invalid account type")

func (r *CoaRepo) typeIsCredit(ctx context.Context, typeCode string) (bool, error) {
	var isCredit bool
	err := r.pool.QueryRow(ctx, `SELECT is_credit FROM m_coa_type WHERE account_type_code = $1`, typeCode).Scan(&isCredit)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, errNoType
	}
	return isCredit, err
}

type CoaInput struct {
	AccountName     string  `json:"account_name"`
	AccountTypeCode string  `json:"account_type_code"`
	IsContra        bool    `json:"is_contra"`
	OpeningBalance  float64 `json:"opening_balance"`
	OpeningDate     string  `json:"opening_date"`
	Notes           string  `json:"notes"`
}

// Create: kode akun dibuat otomatis (kode tipe + urutan 3 digit),
// saldo normal = tipe.is_credit XOR is_contra.
func (r *CoaRepo) Create(ctx context.Context, in CoaInput, userID string) (*Coa, error) {
	isCredit, err := r.typeIsCredit(ctx, in.AccountTypeCode)
	if err != nil {
		return nil, err
	}
	isCreditAccount := isCredit != in.IsContra

	var id string
	err = r.pool.QueryRow(ctx, `
		INSERT INTO m_coa (account_code, account_type_code, account_name, is_contra, is_credit_account,
			opening_balance, opening_date, notes, created_by, modified_by)
		SELECT $1 || lpad((COALESCE(MAX(substring(account_code from 3)::int), 0) + 1)::text, 3, '0'),
			$1, $2, $3, $4, $5, NULLIF($6, '')::date, NULLIF($7, ''), NULLIF($8, '')::uuid, NULLIF($8, '')::uuid
		FROM m_coa WHERE account_code LIKE $1 || '%'
		RETURNING id::text`,
		in.AccountTypeCode, in.AccountName, in.IsContra, isCreditAccount,
		in.OpeningBalance, in.OpeningDate, in.Notes, userID).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.getOne(ctx, id)
}

// Update: account_type, contra, dan kode akun TIDAK pernah diubah saat edit
// (dikunci). account_name / opening_balance / opening_date hanya boleh diubah
// selama akun belum dipakai referensi/transaksi (in_use). Notes selalu boleh.
func (r *CoaRepo) Update(ctx context.Context, id string, in CoaInput, userID string) (*Coa, error) {
	var inUse bool
	err := r.pool.QueryRow(ctx, `SELECT `+coaInUseExpr+` FROM m_coa a WHERE a.id = $1::uuid`, id).Scan(&inUse)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	if inUse {
		// Terkunci: hanya notes yang boleh berubah.
		_, err = r.pool.Exec(ctx,
			`UPDATE m_coa SET notes = NULLIF($1, ''), modified_by = NULLIF($2, '')::uuid WHERE id = $3::uuid`,
			in.Notes, userID, id)
	} else {
		_, err = r.pool.Exec(ctx, `
			UPDATE m_coa SET account_name = $1, opening_balance = $2, opening_date = NULLIF($3, '')::date,
				notes = NULLIF($4, ''), modified_by = NULLIF($5, '')::uuid
			WHERE id = $6::uuid`,
			in.AccountName, in.OpeningBalance, in.OpeningDate, in.Notes, userID, id)
	}
	if err != nil {
		return nil, err
	}
	return r.getOne(ctx, id)
}

func (r *CoaRepo) ToggleActive(ctx context.Context, id string, active bool, userID string) (*Coa, error) {
	ct, err := r.pool.Exec(ctx,
		`UPDATE m_coa SET is_active = $1, modified_by = NULLIF($2, '')::uuid WHERE id = $3::uuid`,
		active, userID, id)
	if err != nil {
		return nil, err
	}
	if ct.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.getOne(ctx, id)
}

// ─── Handler ───

type CoaHandler struct{ repo *CoaRepo }

func (h *CoaHandler) List(c *gin.Context) {
	limit, offset, search, sort, desc := listParams(c)
	items, total, err := h.repo.ListPaged(c.Request.Context(), limit, offset, search, sort, desc,
		c.Query("account_type"), c.Query("classification"), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func bindCoa(c *gin.Context) (CoaInput, bool) {
	var in CoaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return in, false
	}
	if in.AccountName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account name is required", "field": "account_name"})
		return in, false
	}
	if in.AccountTypeCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account type is required", "field": "account_type_code"})
		return in, false
	}
	return in, true
}

func (h *CoaHandler) Create(c *gin.Context) {
	in, ok := bindCoa(c)
	if !ok {
		return
	}
	item, err := h.repo.Create(c.Request.Context(), in, c.GetString(middleware.CtxUserID))
	if err != nil {
		if errors.Is(err, errNoType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account type", "field": "account_type_code"})
			return
		}
		if handleDuplicate(c, err, "Account code already exists", "account_code") {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *CoaHandler) Update(c *gin.Context) {
	in, ok := bindCoa(c)
	if !ok {
		return
	}
	item, err := h.repo.Update(c.Request.Context(), c.Param("id"), in, c.GetString(middleware.CtxUserID))
	if err != nil {
		if errors.Is(err, errNoType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account type", "field": "account_type_code"})
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *CoaHandler) ToggleActive(c *gin.Context) {
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.repo.ToggleActive(c.Request.Context(), c.Param("id"), req.IsActive, c.GetString(middleware.CtxUserID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}
