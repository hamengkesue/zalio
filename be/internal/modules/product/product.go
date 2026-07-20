package product

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"zalio-erp-be/internal/platform/middleware"
)

// errSkuRequired = SKU kosong padahal mode manual (bukan auto-generate).
var errSkuRequired = errors.New("SKU is required")

// ─── Model ───

// Variant = satu unit produk yang dijual (m_product_variant).
type Variant struct {
	ID               string  `json:"id"`
	Sku              string  `json:"sku"`
	Barcode          string  `json:"barcode"`
	VariantValue1    string  `json:"variant_value_1"`
	VariantValue2    string  `json:"variant_value_2"`
	DefSellingPrice  float64 `json:"def_selling_price"`
	DefPurchasePrice float64 `json:"def_purchase_price"`
	CogsUnit         float64 `json:"cogs_unit"`
	LengthCm         float64 `json:"length_cm"`
	WidthCm          float64 `json:"width_cm"`
	HeightCm         float64 `json:"height_cm"`
	WeightGr         float64 `json:"weight_gr"`
	MainImage        string  `json:"main_image"`
	VariantImage     string  `json:"variant_image"` // main image per-varian (default = main image induk)
	Image1           string  `json:"image_1"`
	Image2           string  `json:"image_2"`
	Image3           string  `json:"image_3"`
	IsActive         bool    `json:"is_active"`
}

// Product = induk (m_product) + varian.
type Product struct {
	ID              string    `json:"id"`
	ProductName     string    `json:"product_name"`
	ProductType     string    `json:"product_type"`
	BrandID         string    `json:"brand_id"`
	BrandName       string    `json:"brand_name"`
	SubcategoryID   string    `json:"subcategory_id"`
	SubcategoryName string    `json:"subcategory_name"`
	CategoryID      string    `json:"category_id"`
	CategoryName    string    `json:"category_name"`
	CountryOfOrigin string    `json:"country_of_origin"`
	Description     string    `json:"description"`
	Ingredients     string    `json:"ingredients"`
	IsPerishable    bool      `json:"is_perishable"`
	Uom1            string    `json:"uom_1"`
	Uom2            string    `json:"uom_2"`
	Ratio2          float64   `json:"ratio_2"`
	Uom3            string    `json:"uom_3"`
	Ratio3          float64   `json:"ratio_3"`
	SellingUom      string    `json:"selling_uom"`
	StockingUom     string    `json:"stocking_uom"`
	VariantName1    string    `json:"variant_name_1"`
	VariantName2    string    `json:"variant_name_2"`
	CoaInventory      string  `json:"coa_inventory"`
	CoaSales          string  `json:"coa_sales"`
	CoaSalesReturn    string  `json:"coa_sales_return"`
	CoaSalesDiscount  string  `json:"coa_sales_discount"`
	CoaGoodInTransit  string  `json:"coa_good_in_transit"`
	CoaCogs           string  `json:"coa_cogs"`
	CoaPurchaseReturn string  `json:"coa_purchase_return"`
	CoaUnbilledGoods  string  `json:"coa_unbilled_goods"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	Variants        []Variant `json:"variants,omitempty"`
}

// ProductListItem = satu baris per VARIAN (flat). Produk single = 1 baris.
type ProductListItem struct {
	ID              string    `json:"id"`         // product id (untuk buka edit)
	VariantID       string    `json:"variant_id"` // id varian (untuk key baris + toggle status)
	ProductName     string    `json:"product_name"`
	ProductType     string    `json:"product_type"`
	VariantValue1   string    `json:"variant_value_1"`
	VariantValue2   string    `json:"variant_value_2"`
	BrandName       string    `json:"brand_name"`
	SubcategoryName string    `json:"subcategory_name"`
	CategoryName    string    `json:"category_name"`
	CountryOfOrigin string    `json:"country_of_origin"`
	Sku             string    `json:"sku"`
	Barcode         string    `json:"barcode"`
	Price           float64   `json:"price"`
	Image           string    `json:"image"`            // main_image (single) / variant_image (variant)
	SellingUomName  string    `json:"selling_uom_name"` // satuan jual
	CogsSelling     float64   `json:"cogs_selling"`     // COGS/unit disesuaikan ke selling uom
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

// ─── Input (request body) ───

type VariantInput struct {
	ID               string  `json:"id"`       // ada = update varian existing; kosong = varian baru
	Sku              string  `json:"sku"`      // boleh kosong bila SkuAuto=true (di-generate sistem)
	SkuAuto          bool    `json:"sku_auto"` // true = sistem generate SKU otomatis (PSKU-YYYYMM-######)
	Barcode          string  `json:"barcode"`
	VariantValue1    string  `json:"variant_value_1"`
	VariantValue2    string  `json:"variant_value_2"`
	DefSellingPrice  float64 `json:"def_selling_price"`
	DefPurchasePrice float64 `json:"def_purchase_price"`
	CogsUnit         float64 `json:"cogs_unit"`
	LengthCm         float64 `json:"length_cm"`
	WidthCm          float64 `json:"width_cm"`
	HeightCm         float64 `json:"height_cm"`
	WeightGr         float64 `json:"weight_gr"`
	MainImage        string  `json:"main_image"`
	VariantImage     string  `json:"variant_image"`
	Image1           string  `json:"image_1"`
	Image2           string  `json:"image_2"`
	Image3           string  `json:"image_3"`
	IsActive         bool    `json:"is_active"`
}

type ProductInput struct {
	ProductName       string         `json:"product_name" binding:"required"`
	ProductType       string         `json:"product_type"`
	BrandID           string         `json:"brand_id"`
	SubcategoryID     string         `json:"subcategory_id" binding:"required"`
	CountryOfOrigin   string         `json:"country_of_origin"`
	Description       string         `json:"description"`
	Ingredients       string         `json:"ingredients"`
	IsPerishable      bool           `json:"is_perishable"`
	Uom1              string         `json:"uom_1" binding:"required"`
	Uom2              string         `json:"uom_2"`
	Ratio2            float64        `json:"ratio_2"`
	Uom3              string         `json:"uom_3"`
	Ratio3            float64        `json:"ratio_3"`
	SellingUom        string         `json:"selling_uom"`
	StockingUom       string         `json:"stocking_uom"`
	VariantName1      string         `json:"variant_name_1"`
	VariantName2      string         `json:"variant_name_2"`
	CoaInventory      string         `json:"coa_inventory"`
	CoaSales          string         `json:"coa_sales"`
	CoaSalesReturn    string         `json:"coa_sales_return"`
	CoaSalesDiscount  string         `json:"coa_sales_discount"`
	CoaGoodInTransit  string         `json:"coa_good_in_transit"`
	CoaCogs           string         `json:"coa_cogs"`
	CoaPurchaseReturn string         `json:"coa_purchase_return"`
	CoaUnbilledGoods  string         `json:"coa_unbilled_goods"`
	Variants          []VariantInput `json:"variants" binding:"required,min=1,dive"`
}

// ─── SKU otomatis (PSKU-YYYYMM-######) ───

// rowQuerier dipenuhi oleh *pgxpool.Pool maupun pgx.Tx.
type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// skuPrefix = "PSKU-<YYYYMM>-" untuk bulan pada now.
func skuPrefix(now time.Time) string { return "PSKU-" + now.Format("200601") + "-" }

// formatSku merangkai SKU final: PSKU-YYYYMM-000001.
func formatSku(now time.Time, n int) string { return fmt.Sprintf("%s%06d", skuPrefix(now), n) }

// currentMaxSku mengembalikan nomor urut terbesar yang sudah dipakai pada bulan now (0 bila belum ada).
func currentMaxSku(ctx context.Context, q rowQuerier, now time.Time) (int, error) {
	prefix := skuPrefix(now)
	var last string
	if err := q.QueryRow(ctx,
		`SELECT COALESCE(MAX(sku),'') FROM m_product_variant WHERE sku LIKE $1`, prefix+"%").Scan(&last); err != nil {
		return 0, err
	}
	if len(last) <= len(prefix) {
		return 0, nil
	}
	n, err := strconv.Atoi(last[len(prefix):])
	if err != nil {
		return 0, nil // format tak terduga → mulai dari 0
	}
	return n, nil
}

// assignAutoSkus mengisi SKU untuk varian yang minta auto-generate, dan
// memvalidasi setiap varian akhirnya punya SKU (mode manual wajib isi).
func assignAutoSkus(ctx context.Context, q rowQuerier, variants []VariantInput, now time.Time) error {
	base := -1
	for i := range variants {
		if variants[i].SkuAuto && strings.TrimSpace(variants[i].Sku) == "" {
			if base < 0 {
				n, err := currentMaxSku(ctx, q, now)
				if err != nil {
					return err
				}
				base = n
			}
			base++
			variants[i].Sku = formatSku(now, base)
		}
	}
	for i := range variants {
		if strings.TrimSpace(variants[i].Sku) == "" {
			return errSkuRequired
		}
	}
	return nil
}

// ─── Repository ───

type ProductRepo struct{ pool *pgxpool.Pool }

// PreviewNextSku mengintip SKU berikutnya (untuk ditampilkan di form mode Auto).
func (r *ProductRepo) PreviewNextSku(ctx context.Context, now time.Time) (string, error) {
	n, err := currentMaxSku(ctx, r.pool, now)
	if err != nil {
		return "", err
	}
	return formatSku(now, n+1), nil
}

const productJoin = ` FROM m_product p
	JOIN m_subcategory s ON s.id = p.subcategory_id
	JOIN m_category c ON c.id = s.category_id
	LEFT JOIN m_brand b ON b.id = p.brand_id`

// variantJoin = list flat: satu baris per varian.
const variantJoin = ` FROM m_product_variant v
	JOIN m_product p ON p.id = v.product_id
	JOIN m_subcategory s ON s.id = p.subcategory_id
	JOIN m_category c ON c.id = s.category_id
	LEFT JOIN m_brand b ON b.id = p.brand_id
	LEFT JOIN m_uom su ON su.id = p.selling_uom`

// ProductListFilter = parameter search/sort/filter list produk (flat per varian).
type ProductListFilter struct {
	Search        string
	Sort          string
	Desc          bool
	ProductType   string
	Country       string
	BrandID       string
	CategoryID    string
	SubcategoryID string
	Status        string // "active" | "inactive" | ""
}

// Kolom sort yang diizinkan (nama dari query → ekspresi SQL).
var productSortCols = map[string]string{
	"sku": "v.sku", "name": "p.product_name", "product_name": "p.product_name",
	"type": "p.product_type", "product_type": "p.product_type",
	"brand": "b.name", "category": "c.name", "subcategory": "s.name",
	"price": "v.def_selling_price", "selling_price": "v.def_selling_price",
	"cogs": "cogs_selling", "status": "v.is_active",
}

const productListCols = `p.id::text, v.id::text, p.product_name, p.product_type,
	COALESCE(v.variant_value_1,''), COALESCE(v.variant_value_2,''),
	COALESCE(b.name,'') AS brand_name, s.name AS subcategory_name, c.name AS category_name,
	COALESCE(p.country_of_origin,'') AS country_of_origin,
	COALESCE(v.sku,'') AS sku, COALESCE(v.barcode,'') AS barcode,
	COALESCE(v.def_selling_price,0) AS price,
	CASE WHEN p.product_type='variant' THEN COALESCE(NULLIF(v.variant_image,''), v.main_image, '') ELSE COALESCE(v.main_image,'') END AS image,
	COALESCE(su.name,'') AS selling_uom_name,
	COALESCE(v.cogs_unit,0) * CASE
		WHEN p.selling_uom = p.uom_2 THEN COALESCE(p.ratio_2,1)
		WHEN p.selling_uom = p.uom_3 THEN COALESCE(p.ratio_3,1)
		ELSE 1 END AS cogs_selling,
	v.is_active, v.created_at`

func (r *ProductRepo) ListPaged(ctx context.Context, limit, offset int, f ProductListFilter) ([]ProductListItem, int, error) {
	var conds []string
	var args []any
	if f.Search != "" {
		args = append(args, f.Search)
		p := len(args)
		conds = append(conds, fmt.Sprintf("(p.product_name ILIKE '%%'||$%d||'%%' OR v.sku ILIKE '%%'||$%d||'%%' OR v.barcode ILIKE '%%'||$%d||'%%')", p, p, p))
	}
	if f.ProductType != "" {
		args = append(args, f.ProductType)
		conds = append(conds, fmt.Sprintf("p.product_type = $%d", len(args)))
	}
	if f.Country != "" {
		args = append(args, f.Country)
		conds = append(conds, fmt.Sprintf("p.country_of_origin = $%d", len(args)))
	}
	if f.BrandID != "" {
		args = append(args, f.BrandID)
		conds = append(conds, fmt.Sprintf("p.brand_id = $%d::uuid", len(args)))
	}
	if f.CategoryID != "" {
		args = append(args, f.CategoryID)
		conds = append(conds, fmt.Sprintf("c.id = $%d::uuid", len(args)))
	}
	if f.SubcategoryID != "" {
		args = append(args, f.SubcategoryID)
		conds = append(conds, fmt.Sprintf("s.id = $%d::uuid", len(args)))
	}
	if f.Status == "active" {
		conds = append(conds, "v.is_active = true")
	} else if f.Status == "inactive" {
		conds = append(conds, "v.is_active = false")
	}
	whereClause := ""
	if len(conds) > 0 {
		whereClause = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*)`+variantJoin+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderCol := "p.created_at"
	dir := "DESC"
	if col, ok := productSortCols[f.Sort]; ok {
		orderCol = col
		dir = ascOrDesc(f.Desc)
	}

	args = append(args, limit)
	limitPh := len(args)
	args = append(args, offset)
	offsetPh := len(args)

	rows, err := r.pool.Query(ctx,
		`SELECT `+productListCols+variantJoin+whereClause+
			` ORDER BY `+orderCol+` `+dir+`, p.product_name, p.id, v.created_at LIMIT $`+strconv.Itoa(limitPh)+` OFFSET $`+strconv.Itoa(offsetPh),
		args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []ProductListItem{}
	for rows.Next() {
		var it ProductListItem
		if err := rows.Scan(&it.ID, &it.VariantID, &it.ProductName, &it.ProductType,
			&it.VariantValue1, &it.VariantValue2, &it.BrandName, &it.SubcategoryName, &it.CategoryName,
			&it.CountryOfOrigin, &it.Sku, &it.Barcode, &it.Price, &it.Image, &it.SellingUomName, &it.CogsSelling,
			&it.IsActive, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, it)
	}
	return list, total, rows.Err()
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*Product, error) {
	var p Product
	err := r.pool.QueryRow(ctx,
		`SELECT p.id::text, p.product_name, p.product_type,
			COALESCE(p.brand_id::text,''), COALESCE(b.name,''),
			p.subcategory_id::text, s.name, c.id::text, c.name,
			COALESCE(p.country_of_origin,''), COALESCE(p.description,''), COALESCE(p.ingredients,''), p.is_perishable,
			COALESCE(p.uom_1::text,''), COALESCE(p.uom_2::text,''), COALESCE(p.ratio_2,0),
			COALESCE(p.uom_3::text,''), COALESCE(p.ratio_3,0), COALESCE(p.selling_uom::text,''), COALESCE(p.stocking_uom::text,''),
			COALESCE(p.variant_name_1,''), COALESCE(p.variant_name_2,''),
			COALESCE(p.coa_inventory::text,''), COALESCE(p.coa_sales::text,''), COALESCE(p.coa_sales_return::text,''),
			COALESCE(p.coa_sales_discount::text,''), COALESCE(p.coa_good_in_transit::text,''), COALESCE(p.coa_cogs::text,''),
			COALESCE(p.coa_purchase_return::text,''), COALESCE(p.coa_unbilled_goods::text,''),
			p.is_active, p.created_at`+productJoin+` WHERE p.id = $1::uuid`, id).
		Scan(&p.ID, &p.ProductName, &p.ProductType, &p.BrandID, &p.BrandName, &p.SubcategoryID, &p.SubcategoryName,
			&p.CategoryID, &p.CategoryName, &p.CountryOfOrigin, &p.Description, &p.Ingredients, &p.IsPerishable,
			&p.Uom1, &p.Uom2, &p.Ratio2, &p.Uom3, &p.Ratio3, &p.SellingUom, &p.StockingUom, &p.VariantName1, &p.VariantName2,
			&p.CoaInventory, &p.CoaSales, &p.CoaSalesReturn, &p.CoaSalesDiscount, &p.CoaGoodInTransit, &p.CoaCogs,
			&p.CoaPurchaseReturn, &p.CoaUnbilledGoods, &p.IsActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	vrows, err := r.pool.Query(ctx,
		`SELECT id::text, sku, COALESCE(barcode,''), COALESCE(variant_value_1,''), COALESCE(variant_value_2,''),
			def_selling_price, def_purchase_price, cogs_unit,
			COALESCE(length_cm,0), COALESCE(width_cm,0), COALESCE(height_cm,0), COALESCE(weight_gr,0),
			COALESCE(main_image,''), COALESCE(variant_image,''), COALESCE(image_1,''), COALESCE(image_2,''), COALESCE(image_3,''), is_active
		 FROM m_product_variant WHERE product_id = $1::uuid ORDER BY created_at`, id)
	if err != nil {
		return nil, err
	}
	defer vrows.Close()
	p.Variants = []Variant{}
	for vrows.Next() {
		var v Variant
		if err := vrows.Scan(&v.ID, &v.Sku, &v.Barcode, &v.VariantValue1, &v.VariantValue2,
			&v.DefSellingPrice, &v.DefPurchasePrice, &v.CogsUnit, &v.LengthCm, &v.WidthCm, &v.HeightCm, &v.WeightGr,
			&v.MainImage, &v.VariantImage, &v.Image1, &v.Image2, &v.Image3, &v.IsActive); err != nil {
			return nil, err
		}
		p.Variants = append(p.Variants, v)
	}
	return &p, vrows.Err()
}

// insertProductRow memasukkan baris m_product dan mengembalikan id-nya.
func insertProductRow(ctx context.Context, tx pgx.Tx, in ProductInput, userID string) (string, error) {
	ptype := in.ProductType
	if ptype != "single" && ptype != "variant" {
		ptype = "single"
	}
	var id string
	err := tx.QueryRow(ctx,
		`INSERT INTO m_product (product_name, product_type, brand_id, subcategory_id, country_of_origin,
			description, ingredients, is_perishable, uom_1, uom_2, ratio_2, uom_3, ratio_3, selling_uom,
			variant_name_1, variant_name_2, coa_inventory, coa_sales, coa_sales_return, coa_sales_discount,
			coa_good_in_transit, coa_cogs, coa_purchase_return, coa_unbilled_goods, stocking_uom, created_by, modified_by)
		 VALUES ($1, $2, NULLIF($3,'')::uuid, $4::uuid, NULLIF($5,''), NULLIF($6,''), NULLIF($7,''), $8,
			$9::uuid, NULLIF($10,'')::uuid, NULLIF($11::numeric,0), NULLIF($12,'')::uuid, NULLIF($13::numeric,0), NULLIF($14,'')::uuid,
			NULLIF($15,''), NULLIF($16,''), NULLIF($17,'')::uuid, NULLIF($18,'')::uuid, NULLIF($19,'')::uuid,
			NULLIF($20,'')::uuid, NULLIF($21,'')::uuid, NULLIF($22,'')::uuid, NULLIF($23,'')::uuid, NULLIF($24,'')::uuid, NULLIF($25,'')::uuid, NULLIF($26,'')::uuid, NULLIF($26,'')::uuid)
		 RETURNING id::text`,
		in.ProductName, ptype, in.BrandID, in.SubcategoryID, in.CountryOfOrigin, in.Description, in.Ingredients,
		in.IsPerishable, in.Uom1, in.Uom2, in.Ratio2, in.Uom3, in.Ratio3, in.SellingUom, in.VariantName1, in.VariantName2,
		in.CoaInventory, in.CoaSales, in.CoaSalesReturn, in.CoaSalesDiscount, in.CoaGoodInTransit, in.CoaCogs,
		in.CoaPurchaseReturn, in.CoaUnbilledGoods, in.StockingUom, userID).Scan(&id)
	return id, err
}

func insertVariantRow(ctx context.Context, tx pgx.Tx, productID string, v VariantInput, userID string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO m_product_variant (product_id, sku, barcode, variant_value_1, variant_value_2,
			def_selling_price, def_purchase_price, cogs_unit, length_cm, width_cm, height_cm, weight_gr,
			main_image, variant_image, image_1, image_2, image_3, is_active, created_by, modified_by)
		 VALUES ($1::uuid, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), $6, $7, $8,
			$9::numeric, $10::numeric, $11::numeric, $12::numeric, NULLIF($13,''), NULLIF($14,''), NULLIF($15,''), NULLIF($16,''), NULLIF($17,''), $18, NULLIF($19,'')::uuid, NULLIF($19,'')::uuid)`,
		productID, v.Sku, v.Barcode, v.VariantValue1, v.VariantValue2, v.DefSellingPrice, v.DefPurchasePrice, v.CogsUnit,
		v.LengthCm, v.WidthCm, v.HeightCm, v.WeightGr, v.MainImage, v.VariantImage, v.Image1, v.Image2, v.Image3, v.IsActive, userID)
	return err
}

func updateVariantRow(ctx context.Context, tx pgx.Tx, v VariantInput, userID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE m_product_variant SET sku=$2, barcode=NULLIF($3,''), variant_value_1=NULLIF($4,''), variant_value_2=NULLIF($5,''),
			def_selling_price=$6, def_purchase_price=$7, cogs_unit=$8, length_cm=$9::numeric, width_cm=$10::numeric,
			height_cm=$11::numeric, weight_gr=$12::numeric, main_image=NULLIF($13,''), variant_image=NULLIF($14,''), image_1=NULLIF($15,''),
			image_2=NULLIF($16,''), image_3=NULLIF($17,''), is_active=$18, modified_by=NULLIF($19,'')::uuid, modified_at=now() WHERE id=$1::uuid`,
		v.ID, v.Sku, v.Barcode, v.VariantValue1, v.VariantValue2, v.DefSellingPrice, v.DefPurchasePrice, v.CogsUnit,
		v.LengthCm, v.WidthCm, v.HeightCm, v.WeightGr, v.MainImage, v.VariantImage, v.Image1, v.Image2, v.Image3, v.IsActive, userID)
	return err
}

func (r *ProductRepo) Create(ctx context.Context, in ProductInput, userID string) (*Product, error) {
	if err := assignAutoSkus(ctx, r.pool, in.Variants, time.Now()); err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	id, err := insertProductRow(ctx, tx, in, userID)
	if err != nil {
		return nil, err
	}
	for _, v := range in.Variants {
		if err := insertVariantRow(ctx, tx, id, v, userID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *ProductRepo) Update(ctx context.Context, id string, in ProductInput, userID string) (*Product, error) {
	if err := assignAutoSkus(ctx, r.pool, in.Variants, time.Now()); err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	ptype := in.ProductType
	if ptype != "single" && ptype != "variant" {
		ptype = "single"
	}
	_, err = tx.Exec(ctx,
		`UPDATE m_product SET product_name=$1, product_type=$2, brand_id=NULLIF($3,'')::uuid, subcategory_id=$4::uuid,
			country_of_origin=NULLIF($5,''), description=NULLIF($6,''), ingredients=NULLIF($7,''), is_perishable=$8,
			uom_1=$9::uuid, uom_2=NULLIF($10,'')::uuid, ratio_2=NULLIF($11::numeric,0), uom_3=NULLIF($12,'')::uuid, ratio_3=NULLIF($13::numeric,0),
			selling_uom=NULLIF($14,'')::uuid, variant_name_1=NULLIF($15,''), variant_name_2=NULLIF($16,''),
			coa_inventory=NULLIF($17,'')::uuid, coa_sales=NULLIF($18,'')::uuid, coa_sales_return=NULLIF($19,'')::uuid,
			coa_sales_discount=NULLIF($20,'')::uuid, coa_good_in_transit=NULLIF($21,'')::uuid, coa_cogs=NULLIF($22,'')::uuid,
			coa_purchase_return=NULLIF($23,'')::uuid, coa_unbilled_goods=NULLIF($24,'')::uuid, stocking_uom=NULLIF($25,'')::uuid,
			modified_by=NULLIF($27,'')::uuid, modified_at=now()
		 WHERE id=$26::uuid`,
		in.ProductName, ptype, in.BrandID, in.SubcategoryID, in.CountryOfOrigin, in.Description, in.Ingredients,
		in.IsPerishable, in.Uom1, in.Uom2, in.Ratio2, in.Uom3, in.Ratio3, in.SellingUom, in.VariantName1, in.VariantName2,
		in.CoaInventory, in.CoaSales, in.CoaSalesReturn, in.CoaSalesDiscount, in.CoaGoodInTransit, in.CoaCogs,
		in.CoaPurchaseReturn, in.CoaUnbilledGoods, in.StockingUom, id, userID)
	if err != nil {
		return nil, err
	}

	// Varian: yang punya id → update; yang tidak → insert baru. (Hapus varian = Tahap 3.)
	for _, v := range in.Variants {
		if v.ID != "" {
			if err := updateVariantRow(ctx, tx, v, userID); err != nil {
				return nil, err
			}
		} else {
			if err := insertVariantRow(ctx, tx, id, v, userID); err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *ProductRepo) ToggleActive(ctx context.Context, id string, active bool) error {
	if _, err := r.pool.Exec(ctx, `UPDATE m_product SET is_active=$1 WHERE id=$2::uuid`, active, id); err != nil {
		return err
	}
	// Produk induk di-nonaktifkan → semua varian ikut nonaktif.
	if !active {
		if _, err := r.pool.Exec(ctx, `UPDATE m_product_variant SET is_active=false WHERE product_id=$1::uuid`, id); err != nil {
			return err
		}
	}
	return nil
}

// ToggleVariantActive mengubah status satu varian, lalu menyelaraskan status
// produk induk (aktif bila masih ada minimal satu varian aktif).
func (r *ProductRepo) ToggleVariantActive(ctx context.Context, variantID string, active bool) error {
	if _, err := r.pool.Exec(ctx, `UPDATE m_product_variant SET is_active=$1 WHERE id=$2::uuid`, active, variantID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE m_product p SET is_active = EXISTS(SELECT 1 FROM m_product_variant v WHERE v.product_id = p.id AND v.is_active)
		 WHERE p.id = (SELECT product_id FROM m_product_variant WHERE id=$1::uuid)`, variantID)
	return err
}

// ─── Handler ───

type ProductHandler struct{ repo *ProductRepo }

func (h *ProductHandler) List(c *gin.Context) {
	limit, offset, search, sort, desc := listParams(c)
	f := ProductListFilter{
		Search: search, Sort: sort, Desc: desc,
		ProductType:   c.Query("product_type"),
		Country:       c.Query("country"),
		BrandID:       c.Query("brand_id"),
		CategoryID:    c.Query("category_id"),
		SubcategoryID: c.Query("subcategory_id"),
		Status:        c.Query("status"),
	}
	items, total, err := h.repo.ListPaged(c.Request.Context(), limit, offset, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

// NextSku menampilkan pratinjau SKU otomatis berikutnya (mode Auto di form).
func (h *ProductHandler) NextSku(c *gin.Context) {
	sku, err := h.repo.PreviewNextSku(c.Request.Context(), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"sku": sku}})
}

func (h *ProductHandler) Get(c *gin.Context) {
	p, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

// productSaveError memetakan error umum ke pesan yang ramah.
func productSaveError(c *gin.Context, err error) {
	if errors.Is(err, errSkuRequired) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU is required", "field": "sku"})
		return
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique violation (sku/barcode)
			field := "sku"
			msg := "SKU already exists"
			value := ""
			if pgErr.ConstraintName == "uq_variant_barcode" {
				field, msg = "barcode", "Barcode already exists"
				// pgErr.Detail: "Key (barcode)=(<value>) already exists."
				if i := strings.Index(pgErr.Detail, "=("); i >= 0 {
					if j := strings.Index(pgErr.Detail[i+2:], ")"); j >= 0 {
						value = pgErr.Detail[i+2 : i+2+j]
					}
				}
			}
			c.JSON(http.StatusConflict, gin.H{"error": msg, "field": field, "value": value})
			return
		case "23503", "22P02": // FK / uuid tidak valid
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reference (brand/subcategory/uom)"})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var in ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.repo.Create(c.Request.Context(), in, c.GetString(middleware.CtxUserID))
	if err != nil {
		productSaveError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p})
}

func (h *ProductHandler) Update(c *gin.Context) {
	var in ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.repo.Update(c.Request.Context(), c.Param("id"), in, c.GetString(middleware.CtxUserID))
	if err != nil {
		productSaveError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *ProductHandler) ToggleActive(c *gin.Context) {
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.ToggleActive(c.Request.Context(), c.Param("id"), req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	p, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

// ToggleVariantActive: nonaktif/aktifkan satu varian (list flat per varian).
func (h *ProductHandler) ToggleVariantActive(c *gin.Context) {
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.ToggleVariantActive(c.Request.Context(), c.Param("id"), req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"is_active": req.IsActive}})
}
