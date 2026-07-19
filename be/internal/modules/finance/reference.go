package finance

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Data referensi CoA: klasifikasi (7) & tipe akun (17). Seed tetap, hanya dibaca
// (dipakai untuk mengisi dropdown di form Chart of Account).

type Classification struct {
	Code       int    `json:"classification_code"`
	Name       string `json:"classification_name"`
	ReportName string `json:"report_name"`
}

type AccountType struct {
	Code               string `json:"account_type_code"`
	ClassificationCode int    `json:"classification_code"`
	ClassificationName string `json:"classification_name"`
	Name               string `json:"account_type_name"`
	IsCredit           bool   `json:"is_credit"`
}

type RefRepo struct{ pool *pgxpool.Pool }

func (r *RefRepo) Classifications(ctx context.Context) ([]Classification, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT classification_code, classification_name, report_name
		 FROM m_coa_classification ORDER BY classification_code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Classification{}
	for rows.Next() {
		var x Classification
		if err := rows.Scan(&x.Code, &x.Name, &x.ReportName); err != nil {
			return nil, err
		}
		list = append(list, x)
	}
	return list, rows.Err()
}

func (r *RefRepo) Types(ctx context.Context) ([]AccountType, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.account_type_code, t.classification_code, c.classification_name, t.account_type_name, t.is_credit
		 FROM m_coa_type t
		 JOIN m_coa_classification c ON c.classification_code = t.classification_code
		 ORDER BY t.account_type_code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []AccountType{}
	for rows.Next() {
		var x AccountType
		if err := rows.Scan(&x.Code, &x.ClassificationCode, &x.ClassificationName, &x.Name, &x.IsCredit); err != nil {
			return nil, err
		}
		list = append(list, x)
	}
	return list, rows.Err()
}

type RefHandler struct{ repo *RefRepo }

func (h *RefHandler) ListClassifications(c *gin.Context) {
	items, err := h.repo.Classifications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *RefHandler) ListTypes(c *gin.Context) {
	items, err := h.repo.Types(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
