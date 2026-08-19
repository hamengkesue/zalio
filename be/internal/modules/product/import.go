package product

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"

	"zalio-erp-be/internal/platform/middleware"
)

// ─── Struktur template (2 sheet) ───
//
// Sheet "Product List": 1 baris = 1 produk. Single = lengkap; Variant = hanya
// level-produk (barcode/harga/berat dikosongkan, itu per-varian).
// Sheet "Variant List": 1 baris = 1 varian, dihubungkan ke produk lewat product_name.

const (
	sheetProduct = "Product List"
	sheetVariant = "Variant List"
	sheetRef     = "Reference"
)

var productCols = []string{
	"product_type", "barcode", "product_name", "variant_name_1", "variant_name_2",
	"brand", "subcategory", "country", "is_perishable", "description", "ingredients",
	"uom_1", "uom_2", "uom_3", "ratio_2", "ratio_3", "selling_uom", "stocking_uom",
	"def_selling_price", "def_purchase_price", "weight_gr", "length_cm", "width_cm", "height_cm",
	"coa_inventory", "coa_sales", "coa_sales_return", "coa_sales_discount",
	"coa_good_in_transit", "coa_cogs", "coa_purchase_return", "coa_unbilled_goods",
}

var variantCols = []string{
	"product_name", "barcode", "variant_value_1", "variant_value_2",
	"def_selling_price", "def_purchase_price", "weight_gr",
}

var coaFields = []string{
	"coa_inventory", "coa_sales", "coa_sales_return", "coa_sales_discount",
	"coa_good_in_transit", "coa_cogs", "coa_purchase_return", "coa_unbilled_goods",
}

// Default kode COA (dipakai bila kolom dikosongkan).
var coaDefaults = map[string]string{
	"coa_inventory": "13001", "coa_sales": "40001", "coa_sales_return": "40003",
	"coa_sales_discount": "40004", "coa_good_in_transit": "13002", "coa_cogs": "51001",
	"coa_purchase_return": "13001", "coa_unbilled_goods": "22003",
}

// Filter SQL per kolom COA (untuk daftar dropdown di sheet Reference).
var coaRefFilter = map[string]string{
	"coa_inventory":       `t.account_type_name = 'Persediaan Barang'`,
	"coa_sales":           `c.classification_name = 'Pendapatan' AND NOT a.is_contra`,
	"coa_sales_return":    `c.classification_name = 'Pendapatan' AND a.is_contra`,
	"coa_sales_discount":  `c.classification_name = 'Pendapatan' AND a.is_contra`,
	"coa_good_in_transit": `t.account_type_name = 'Persediaan Barang'`,
	"coa_cogs":            `c.classification_name = 'Beban Pokok Pendapatan'`,
	"coa_purchase_return": `t.account_type_name = 'Persediaan Barang'`,
	"coa_unbilled_goods":  `c.classification_name = 'Liabilitas'`,
}

// Label header + catatan (baris 1) untuk template.
var productRequired = map[string]bool{
	"product_type": true, "product_name": true, "variant_name_1": true, "brand": true,
	"subcategory": true, "is_perishable": true, "description": true, "uom_1": true,
	"selling_uom": true, "stocking_uom": true, "weight_gr": true,
}
var productOptional = map[string]bool{
	"barcode": true, "variant_name_2": true, "country": true, "ingredients": true,
	"uom_2": true, "uom_3": true,
}
var variantRequired = map[string]bool{"product_name": true, "variant_value_1": true, "weight_gr": true}
var variantOptional = map[string]bool{"barcode": true, "variant_value_2": true}

var productNotes = map[string]string{
	"product_type":       "wajib: 'single' atau 'variant'",
	"barcode":            "single: opsional, jika diisi harus unik.\nvariant: WAJIB dikosongkan (isi di Variant List)",
	"product_name":       "wajib & unik (tidak boleh duplikat)",
	"variant_name_1":     "single: WAJIB dikosongkan.\nvariant: wajib diisi",
	"variant_name_2":     "single: WAJIB dikosongkan.\nvariant: opsional",
	"brand":              "wajib, harus ada di sheet Reference",
	"subcategory":        "wajib, harus ada di sheet Reference",
	"country":            "opsional, jika diisi harus ada di sheet Reference",
	"is_perishable":      "wajib: 'yes' atau 'no'",
	"description":        "wajib diisi",
	"ingredients":        "opsional",
	"uom_1":              "wajib, harus ada di sheet Reference",
	"uom_2":              "opsional",
	"uom_3":              "opsional",
	"ratio_2":            "wajib > 1 jika uom_2 diisi; kosongkan jika uom_2 kosong",
	"ratio_3":            "wajib > ratio_2 jika uom_3 diisi; kosongkan jika uom_3 kosong",
	"selling_uom":        "wajib, salah satu dari uom_1/uom_2/uom_3",
	"stocking_uom":       "wajib, salah satu dari uom_1/uom_2/uom_3",
	"def_selling_price":  "single: wajib > 0.\nvariant: WAJIB dikosongkan (isi di Variant List)",
	"def_purchase_price": "single: wajib >= 0.\nvariant: WAJIB dikosongkan (isi di Variant List)",
	"weight_gr":          "single: wajib > 0.\nvariant: WAJIB dikosongkan (isi di Variant List)",
	"length_cm":          "opsional (cm)",
	"width_cm":           "opsional (cm)",
	"height_cm":          "opsional (cm)",
	"coa_inventory":      "opsional. Jika diisi cek Reference. Kosong = auto (default)",
	"coa_sales":          "opsional. Jika diisi cek Reference. Kosong = auto (default)",
	"coa_sales_return":   "opsional. Jika diisi cek Reference. Kosong = auto (default)",
	"coa_sales_discount": "opsional. Jika diisi cek Reference. Kosong = auto (default)",
	"coa_good_in_transit": "opsional. Jika diisi cek Reference. Kosong = auto (default)",
	"coa_cogs":           "opsional. Jika diisi cek Reference. Kosong = auto (default)",
	"coa_purchase_return": "opsional. Jika diisi cek Reference. Kosong = auto (default)",
	"coa_unbilled_goods": "opsional. Jika diisi cek Reference. Kosong = auto (default)",
}
var variantNotes = map[string]string{
	"product_name":       "copy dari product_name di 'Product List' yang bertipe 'variant'",
	"barcode":            "opsional, jika diisi harus unik",
	"variant_value_1":    "wajib diisi. Kombinasi value_1 + value_2 harus unik per produk",
	"variant_value_2":    "opsional",
	"def_selling_price":  "opsional (kosong = 0)",
	"def_purchase_price": "opsional (kosong = 0)",
	"weight_gr":          "wajib > 0",
}

// ─── Master lookup ───

type coaRef struct{ id string }

type importMasters struct {
	brands       map[string]string // name -> id (case-sensitive)
	subcats      map[string]string // name -> id (case-sensitive; nama unik)
	uoms         map[string]string // name -> id (case-sensitive)
	countries    map[string]string // lower(name|code) -> code
	coa          map[string]coaRef // account_code -> ref
	existProduct map[string]bool   // lower(product_name)
	existBarcode map[string]bool
}

func loadImportMasters(ctx context.Context, pool *pgxpool.Pool) (*importMasters, error) {
	m := &importMasters{
		brands: map[string]string{}, subcats: map[string]string{}, uoms: map[string]string{},
		countries: map[string]string{}, coa: map[string]coaRef{},
		existProduct: map[string]bool{}, existBarcode: map[string]bool{},
	}
	load := func(q string, fn func(scan func(...any) error) error) error {
		rows, err := pool.Query(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			if err := fn(rows.Scan); err != nil {
				return err
			}
		}
		return rows.Err()
	}
	scanKV := func(dst map[string]string) func(func(...any) error) error {
		return func(s func(...any) error) error {
			var id, n string
			if err := s(&id, &n); err != nil {
				return err
			}
			dst[n] = id
			return nil
		}
	}
	if err := load(`SELECT id::text, name FROM m_brand WHERE is_active`, scanKV(m.brands)); err != nil {
		return nil, err
	}
	if err := load(`SELECT id::text, name FROM m_subcategory WHERE is_active`, scanKV(m.subcats)); err != nil {
		return nil, err
	}
	if err := load(`SELECT id::text, name FROM m_uom WHERE is_active`, scanKV(m.uoms)); err != nil {
		return nil, err
	}
	if err := load(`SELECT country_code, lower(country_name) FROM m_country`, func(s func(...any) error) error {
		var code, n string
		if err := s(&code, &n); err != nil {
			return err
		}
		m.countries[n] = code
		m.countries[strings.ToLower(code)] = code
		return nil
	}); err != nil {
		return nil, err
	}
	if err := load(`SELECT id::text, account_code FROM m_coa WHERE is_active`, func(s func(...any) error) error {
		var id, code string
		if err := s(&id, &code); err != nil {
			return err
		}
		m.coa[code] = coaRef{id: id}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := load(`SELECT lower(product_name) FROM m_product`, func(s func(...any) error) error {
		var n string
		if err := s(&n); err != nil {
			return err
		}
		m.existProduct[n] = true
		return nil
	}); err != nil {
		return nil, err
	}
	if err := load(`SELECT COALESCE(barcode,'') FROM m_product_variant`, func(s func(...any) error) error {
		var bc string
		if err := s(&bc); err != nil {
			return err
		}
		if bc != "" {
			m.existBarcode[bc] = true
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return m, nil
}

// ─── Parsing ───

type importRow struct {
	rowNum int
	cells  map[string]string
}

func low(s string) string  { return strings.ToLower(strings.TrimSpace(s)) }
func trim(s string) string { return strings.TrimSpace(s) }

// normHeader: "barcode (optional)" / "product_type *" -> "barcode" / "product_type".
func normHeader(h string) string {
	h = strings.TrimSpace(h)
	if i := strings.Index(h, "("); i >= 0 {
		h = h[:i]
	}
	h = strings.ReplaceAll(h, "*", "")
	return strings.ToLower(strings.TrimSpace(h))
}

func parseFloatLoose(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, true
	}
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func leadingCode(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " - "); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func truthy(s string) bool {
	switch low(s) {
	case "yes", "y", "ya", "true", "1":
		return true
	}
	return false
}

// parseSheet membaca sheet berdasarkan nama; header dicari otomatis (baris dgn
// paling banyak nama kolom cocok), data mulai baris setelahnya.
func parseSheet(f *excelize.File, sheetName string, cols []string) ([]importRow, error) {
	if idx, _ := f.GetSheetIndex(sheetName); idx < 0 {
		return nil, fmt.Errorf("sheet %q tidak ditemukan", sheetName)
	}
	raw, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	colSet := map[string]bool{}
	for _, c := range cols {
		colSet[c] = true
	}
	headerIdx, best := -1, 0
	for i := 0; i < len(raw) && i < 12; i++ {
		n := 0
		for _, cell := range raw[i] {
			if colSet[normHeader(cell)] {
				n++
			}
		}
		if n > best {
			best = n
			headerIdx = i
		}
	}
	if headerIdx < 0 {
		return nil, fmt.Errorf("baris header tidak ditemukan di sheet %q", sheetName)
	}
	idxToCol := map[int]string{}
	for j, cell := range raw[headerIdx] {
		if nc := normHeader(cell); colSet[nc] {
			idxToCol[j] = nc
		}
	}
	// Data mulai 2 baris setelah header (lewati 2 baris contoh). Header di baris 2 → data mulai baris 5.
	rows := []importRow{}
	for i := headerIdx + 3; i < len(raw); i++ {
		cells := map[string]string{}
		for _, c := range cols {
			cells[c] = ""
		}
		blank := true
		for j, v := range raw[i] {
			if c, ok := idxToCol[j]; ok {
				v = strings.TrimSpace(v)
				cells[c] = v
				if v != "" {
					blank = false
				}
			}
		}
		if blank {
			continue
		}
		rows = append(rows, importRow{rowNum: i + 1, cells: cells})
	}
	return rows, nil
}

// ─── Validasi ───

type importRowResult struct {
	Row     int               `json:"row"`
	Status  string            `json:"status"`
	Cells   map[string]string `json:"cells"`
	Errors  []string          `json:"errors"`
	BadCols []string          `json:"bad_cols"`
}

type importResult struct {
	Totals         map[string]int    `json:"totals"`
	ProductColumns []string          `json:"product_columns"`
	VariantColumns []string          `json:"variant_columns"`
	ProductRows    []importRowResult `json:"product_rows"`
	VariantRows    []importRowResult `json:"variant_rows"`
	builds         []*ProductInput   // produk siap-impor (internal)
}

// validateProductRow memvalidasi satu baris Product List.
func validateProductRow(cells map[string]string, m *importMasters, nameCount, bcCount map[string]int) (errs, bad []string, in *ProductInput, isVariant bool) {
	badSet := map[string]bool{}
	add := func(col, msg string) {
		errs = append(errs, msg)
		if col != "" {
			badSet[col] = true
		}
	}
	ptype := low(cells["product_type"])
	isVariant = ptype == "variant"
	in = &ProductInput{
		ProductName: trim(cells["product_name"]), ProductType: ptype,
		Description: trim(cells["description"]), Ingredients: trim(cells["ingredients"]),
		IsPerishable: truthy(cells["is_perishable"]),
	}
	if ptype != "single" && ptype != "variant" {
		add("product_type", `product_type must be "single" or "variant"`)
	}
	if name := trim(cells["product_name"]); name == "" {
		add("product_name", "product_name is required")
	} else if nameCount[low(name)] > 1 {
		add("product_name", "product_name is duplicated in Product List")
	}
	if trim(cells["is_perishable"]) == "" {
		add("is_perishable", "is_perishable is required (yes/no)")
	}
	if trim(cells["description"]) == "" {
		add("description", "description is required")
	}
	if v := trim(cells["brand"]); v == "" {
		add("brand", "brand is required")
	} else if id, ok := m.brands[v]; ok {
		in.BrandID = id
	} else {
		add("brand", fmt.Sprintf("brand %q not found", cells["brand"]))
	}
	if v := trim(cells["subcategory"]); v == "" {
		add("subcategory", "subcategory is required")
	} else if id, ok := m.subcats[v]; ok {
		in.SubcategoryID = id
	} else {
		add("subcategory", fmt.Sprintf("subcategory %q not found", cells["subcategory"]))
	}
	if raw := trim(cells["country"]); raw != "" {
		if code, ok := m.countries[strings.ToLower(leadingCode(raw))]; ok {
			in.CountryOfOrigin = code
		} else {
			add("country", fmt.Sprintf("country %q not found", raw))
		}
	}
	resolveUom := func(field string, required bool) string {
		v := trim(cells[field])
		if v == "" {
			if required {
				add(field, field+" is required")
			}
			return ""
		}
		if id, ok := m.uoms[v]; ok {
			return id
		}
		add(field, fmt.Sprintf("%s %q not found", field, cells[field]))
		return ""
	}
	in.Uom1 = resolveUom("uom_1", true)
	in.Uom2 = resolveUom("uom_2", false)
	in.Uom3 = resolveUom("uom_3", false)
	if in.Uom2 != "" {
		if r, ok := parseFloatLoose(cells["ratio_2"]); !ok || r <= 1 {
			add("ratio_2", "ratio_2 must be a number greater than 1 when uom_2 is set")
		} else {
			in.Ratio2 = r
		}
	} else if trim(cells["ratio_2"]) != "" {
		add("ratio_2", "ratio_2 must be empty when uom_2 is empty")
	}
	if in.Uom3 != "" {
		if r, ok := parseFloatLoose(cells["ratio_3"]); !ok || r <= in.Ratio2 {
			add("ratio_3", "ratio_3 must be greater than ratio_2 when uom_3 is set")
		} else {
			in.Ratio3 = r
		}
	} else if trim(cells["ratio_3"]) != "" {
		add("ratio_3", "ratio_3 must be empty when uom_3 is empty")
	}
	uomSet := map[string]bool{}
	for _, id := range []string{in.Uom1, in.Uom2, in.Uom3} {
		if id != "" {
			uomSet[id] = true
		}
	}
	in.SellingUom = resolveUom("selling_uom", true)
	in.StockingUom = resolveUom("stocking_uom", true)
	if in.SellingUom != "" && !uomSet[in.SellingUom] {
		add("selling_uom", "selling_uom must be one of uom_1/uom_2/uom_3")
	}
	if in.StockingUom != "" && !uomSet[in.StockingUom] {
		add("stocking_uom", "stocking_uom must be one of uom_1/uom_2/uom_3")
	}
	// COA
	coaPtr := map[string]*string{
		"coa_inventory": &in.CoaInventory, "coa_sales": &in.CoaSales, "coa_sales_return": &in.CoaSalesReturn,
		"coa_sales_discount": &in.CoaSalesDiscount, "coa_good_in_transit": &in.CoaGoodInTransit,
		"coa_cogs": &in.CoaCogs, "coa_purchase_return": &in.CoaPurchaseReturn, "coa_unbilled_goods": &in.CoaUnbilledGoods,
	}
	for _, field := range coaFields {
		code := leadingCode(cells[field])
		if code == "" {
			code = coaDefaults[field]
		}
		if ref, ok := m.coa[code]; ok {
			*coaPtr[field] = ref.id
		} else {
			add(field, fmt.Sprintf("%s account %q not found", field, code))
		}
	}
	if isVariant {
		in.VariantName1 = trim(cells["variant_name_1"])
		in.VariantName2 = trim(cells["variant_name_2"])
		if in.VariantName1 == "" {
			add("variant_name_1", "variant_name_1 is required for variant products")
		}
		// Barcode/harga/berat harus di Variant List → kosong di sini.
		for _, f := range []string{"barcode", "def_selling_price", "def_purchase_price", "weight_gr"} {
			if trim(cells[f]) != "" {
				add(f, f+" must be empty for a variant (fill it in Variant List)")
			}
		}
	} else {
		if trim(cells["variant_name_1"]) != "" {
			add("variant_name_1", "variant_name_1 must be empty for a single product")
		}
		if trim(cells["variant_name_2"]) != "" {
			add("variant_name_2", "variant_name_2 must be empty for a single product")
		}
		if bc := trim(cells["barcode"]); bc != "" && (m.existBarcode[bc] || bcCount[bc] > 1) {
			add("barcode", fmt.Sprintf("barcode %q already exists", bc))
		}
		if p, ok := parseFloatLoose(cells["def_selling_price"]); !ok || p <= 0 {
			add("def_selling_price", "def_selling_price must be greater than 0")
		}
		if raw := trim(cells["def_purchase_price"]); raw == "" {
			add("def_purchase_price", "def_purchase_price is required")
		} else if p, ok := parseFloatLoose(raw); !ok || p < 0 {
			add("def_purchase_price", "def_purchase_price must be a number (0 or more)")
		}
		if raw := trim(cells["weight_gr"]); raw == "" {
			add("weight_gr", "weight_gr is required")
		} else if w, ok := parseFloatLoose(raw); !ok || w <= 0 {
			add("weight_gr", "weight_gr must be greater than 0")
		}
	}
	for c := range badSet {
		bad = append(bad, c)
	}
	return
}

// validateVariantRow memvalidasi satu baris Variant List.
func validateVariantRow(cells map[string]string, m *importMasters, prodType map[string]string, prodVN2 map[string]bool, bcCount, comboCount map[string]int) (errs, bad []string) {
	badSet := map[string]bool{}
	add := func(col, msg string) {
		errs = append(errs, msg)
		if col != "" {
			badSet[col] = true
		}
	}
	name := low(cells["product_name"])
	if trim(cells["product_name"]) == "" {
		add("product_name", "product_name is required")
	} else if t, ok := prodType[name]; !ok {
		add("product_name", "no matching product in Product List")
	} else if t != "variant" {
		add("product_name", "product_name must reference a variant-type product")
	}
	if trim(cells["variant_value_1"]) == "" {
		add("variant_value_1", "variant_value_1 is required")
	}
	// variant_value_2 wajib bila produknya (di Product List) punya variant_name_2.
	if prodVN2[name] && trim(cells["variant_value_2"]) == "" {
		add("variant_value_2", "variant_value_2 is required (this product has variant_name_2)")
	}
	combo := name + "||" + low(cells["variant_value_1"]) + "||" + low(cells["variant_value_2"])
	if comboCount[combo] > 1 {
		add("variant_value_1", "duplicate variant combination")
	}
	if bc := trim(cells["barcode"]); bc != "" && (m.existBarcode[bc] || bcCount[bc] > 1) {
		add("barcode", fmt.Sprintf("barcode %q already exists", bc))
	}
	if raw := trim(cells["def_selling_price"]); raw != "" {
		if p, ok := parseFloatLoose(raw); !ok || p < 0 {
			add("def_selling_price", "def_selling_price must be a number (0 or more)")
		}
	}
	if raw := trim(cells["def_purchase_price"]); raw != "" {
		if p, ok := parseFloatLoose(raw); !ok || p < 0 {
			add("def_purchase_price", "def_purchase_price must be a number (0 or more)")
		}
	}
	if raw := trim(cells["weight_gr"]); raw == "" {
		add("weight_gr", "weight_gr is required")
	} else if w, ok := parseFloatLoose(raw); !ok || w <= 0 {
		add("weight_gr", "weight_gr must be greater than 0")
	}
	for c := range badSet {
		bad = append(bad, c)
	}
	return
}

func variantFromCells(cells, prodCells map[string]string) VariantInput {
	sp, _ := parseFloatLoose(cells["def_selling_price"])
	pp, _ := parseFloatLoose(cells["def_purchase_price"])
	wt, _ := parseFloatLoose(cells["weight_gr"])
	length, _ := parseFloatLoose(prodCells["length_cm"])
	width, _ := parseFloatLoose(prodCells["width_cm"])
	height, _ := parseFloatLoose(prodCells["height_cm"])
	return VariantInput{
		SkuAuto: true, IsActive: true, Barcode: trim(cells["barcode"]),
		VariantValue1: trim(cells["variant_value_1"]), VariantValue2: trim(cells["variant_value_2"]),
		DefSellingPrice: sp, DefPurchasePrice: pp, WeightGr: wt,
		LengthCm: length, WidthCm: width, HeightCm: height,
	}
}

func singleVariantFromCells(cells map[string]string) VariantInput {
	sp, _ := parseFloatLoose(cells["def_selling_price"])
	pp, _ := parseFloatLoose(cells["def_purchase_price"])
	wt, _ := parseFloatLoose(cells["weight_gr"])
	length, _ := parseFloatLoose(cells["length_cm"])
	width, _ := parseFloatLoose(cells["width_cm"])
	height, _ := parseFloatLoose(cells["height_cm"])
	return VariantInput{
		SkuAuto: true, IsActive: true, Barcode: trim(cells["barcode"]),
		DefSellingPrice: sp, DefPurchasePrice: pp, WeightGr: wt,
		LengthCm: length, WidthCm: width, HeightCm: height,
	}
}

func validateImport(prodRows, varRows []importRow, m *importMasters) importResult {
	res := importResult{
		ProductColumns: productCols, VariantColumns: variantCols,
		ProductRows: []importRowResult{}, VariantRows: []importRowResult{},
	}
	// Precompute.
	nameCount := map[string]int{}
	prodType := map[string]string{}
	prodVN2 := map[string]bool{}
	prodCellsByName := map[string]map[string]string{}
	for _, r := range prodRows {
		if n := low(r.cells["product_name"]); n != "" {
			nameCount[n]++
			prodType[n] = low(r.cells["product_type"])
			prodVN2[n] = trim(r.cells["variant_name_2"]) != ""
			prodCellsByName[n] = r.cells
		}
	}
	bcCount := map[string]int{}
	for _, r := range prodRows {
		if bc := trim(r.cells["barcode"]); bc != "" {
			bcCount[bc]++
		}
	}
	for _, r := range varRows {
		if bc := trim(r.cells["barcode"]); bc != "" {
			bcCount[bc]++
		}
	}
	comboCount := map[string]int{}
	varsByName := map[string][]importRow{}
	for _, r := range varRows {
		name := low(r.cells["product_name"])
		combo := name + "||" + low(r.cells["variant_value_1"]) + "||" + low(r.cells["variant_value_2"])
		comboCount[combo]++
		varsByName[name] = append(varsByName[name], r)
	}

	// Validasi baris varian dulu (untuk tahu validitas per produk).
	varAllValid := map[string]bool{}
	for name := range varsByName {
		varAllValid[name] = true
	}
	for _, r := range varRows {
		name := low(r.cells["product_name"])
		exists := m.existProduct[name]
		errs, bad := validateVariantRow(r.cells, m, prodType, prodVN2, bcCount, comboCount)
		status := "valid"
		if exists {
			status = "skip_exists"
		} else if len(errs) > 0 {
			status = "error"
			varAllValid[name] = false
		}
		res.VariantRows = append(res.VariantRows, importRowResult{Row: r.rowNum, Status: status, Cells: r.cells, Errors: errs, BadCols: bad})
	}

	// Validasi baris produk (+ cek silang variant harus punya baris varian).
	pinput := map[string]*ProductInput{}
	pvalid := map[string]bool{}
	for _, r := range prodRows {
		name := low(r.cells["product_name"])
		exists := m.existProduct[name]
		errs, bad, in, isVariant := validateProductRow(r.cells, m, nameCount, bcCount)
		if isVariant && !exists && len(varsByName[name]) == 0 {
			errs = append(errs, "variant product has no rows in Variant List")
			bad = append(bad, "product_name")
		}
		status := "valid"
		if exists {
			status = "skip_exists"
		} else if len(errs) > 0 {
			status = "error"
		}
		res.ProductRows = append(res.ProductRows, importRowResult{Row: r.rowNum, Status: status, Cells: r.cells, Errors: errs, BadCols: bad})
		if status == "valid" {
			pinput[name] = in
			pvalid[name] = true
		}
	}

	// Bangun produk siap-impor.
	for _, r := range prodRows {
		name := low(r.cells["product_name"])
		if m.existProduct[name] || !pvalid[name] {
			continue
		}
		in := pinput[name]
		if prodType[name] == "single" {
			in.Variants = []VariantInput{singleVariantFromCells(r.cells)}
			res.builds = append(res.builds, in)
		} else {
			if len(varsByName[name]) == 0 || !varAllValid[name] {
				continue
			}
			vs := []VariantInput{}
			for _, vr := range varsByName[name] {
				vs = append(vs, variantFromCells(vr.cells, r.cells))
			}
			in.Variants = vs
			res.builds = append(res.builds, in)
		}
	}

	// Totals per baris (gabungan kedua sheet).
	valid, errCnt, skip := 0, 0, 0
	for _, rr := range append(append([]importRowResult{}, res.ProductRows...), res.VariantRows...) {
		switch rr.Status {
		case "valid":
			valid++
		case "skip_exists":
			skip++
		default:
			errCnt++
		}
	}
	res.Totals = map[string]int{"valid": valid, "error": errCnt, "skip": skip, "valid_products": len(res.builds)}
	return res
}

// ─── Handlers ───

func label(col string, req, opt map[string]bool) string {
	if req[col] {
		return col + " *"
	}
	if opt[col] {
		return col + " (optional)"
	}
	return col
}

// writeSheet menulis satu sheet import: baris1 catatan, baris2 header, baris3-4 contoh
// (teks merah). Baris 1-4 dikunci; area input (baris 5+) tidak dikunci.
func writeSheet(f *excelize.File, sheet string, cols []string, req, opt map[string]bool, notesTxt map[string]string, examples []map[string]string) {
	noteStyle, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"}, Protection: &excelize.Protection{Locked: true}})
	headStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Protection: &excelize.Protection{Locked: true}})
	redStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "FF0000"}, Protection: &excelize.Protection{Locked: true}})
	unlocked, _ := f.NewStyle(&excelize.Style{Protection: &excelize.Protection{Locked: false}})
	colOf := map[string]int{}
	for i, c := range cols {
		colOf[c] = i + 1
	}
	for _, c := range cols {
		nCell, _ := excelize.CoordinatesToCellName(colOf[c], 1)
		f.SetCellValue(sheet, nCell, notesTxt[c])
		hCell, _ := excelize.CoordinatesToCellName(colOf[c], 2)
		f.SetCellValue(sheet, hCell, label(c, req, opt))
	}
	lastCol, _ := excelize.ColumnNumberToName(len(cols))
	f.SetColWidth(sheet, "A", lastCol, 22)
	f.SetRowHeight(sheet, 1, 70)
	f.SetCellStyle(sheet, "A1", lastCol+"1", noteStyle)
	f.SetCellStyle(sheet, "A2", lastCol+"2", headStyle)
	for r, ex := range examples {
		rowN := r + 3
		for _, c := range cols {
			if v := ex[c]; v != "" {
				cell, _ := excelize.CoordinatesToCellName(colOf[c], rowN)
				f.SetCellValue(sheet, cell, v)
			}
		}
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowN), fmt.Sprintf("%s%d", lastCol, rowN), redStyle)
	}
	f.SetCellStyle(sheet, "A5", fmt.Sprintf("%s1000", lastCol), unlocked)
}

// ImportTemplate: unduh template Excel (Reference + Product List + Variant List, dgn dropdown).
func (h *ProductHandler) ImportTemplate(c *gin.Context) {
	ctx := c.Request.Context()
	f := excelize.NewFile()
	defer f.Close()

	// ── Sheet Reference (data asli DB; semua sel dikunci) ──
	f.SetSheetName(f.GetSheetName(0), sheetRef)
	redRef, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "FF0000"}, Protection: &excelize.Protection{Locked: true}})
	headRef, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Protection: &excelize.Protection{Locked: true}})
	lockRef, _ := f.NewStyle(&excelize.Style{Protection: &excelize.Protection{Locked: true}})
	type refRng struct{ col, start, end int }
	refRange := map[string]refRng{}
	// Brands/Subcategories/UoM: baris 2-3 contoh (merah), data mulai baris 4.
	// Country/COA: tanpa contoh, data mulai baris 2.
	writeRef := func(key string, col int, title string, examples []string, startRow int, q string) {
		cL, _ := excelize.ColumnNumberToName(col)
		f.SetCellValue(sheetRef, cL+"1", title)
		f.SetCellStyle(sheetRef, cL+"1", cL+"1", headRef)
		for i, ex := range examples {
			cell := fmt.Sprintf("%s%d", cL, 2+i)
			f.SetCellValue(sheetRef, cell, ex)
			f.SetCellStyle(sheetRef, cell, cell, redRef)
		}
		r := startRow
		if rows, err := h.repo.pool.Query(ctx, q); err == nil {
			defer rows.Close()
			for rows.Next() {
				var v string
				if rows.Scan(&v) == nil {
					cell := fmt.Sprintf("%s%d", cL, r)
					f.SetCellValue(sheetRef, cell, v)
					f.SetCellStyle(sheetRef, cell, cell, lockRef)
					r++
				}
			}
		}
		end := r - 1
		if end < startRow {
			end = startRow
		}
		refRange[key] = refRng{col: col, start: startRow, end: end}
	}
	writeRef("brand", 1, "Brands", []string{"Contoh Brand 1", "Contoh Brand 2"}, 4, `SELECT name FROM m_brand WHERE is_active ORDER BY name`)
	writeRef("subcategory", 2, "Subcategories", []string{"Contoh Subkategori 1", "Contoh Subkategori 2"}, 4, `SELECT name FROM m_subcategory WHERE is_active ORDER BY name`)
	writeRef("uom", 3, "UoM", []string{"Contoh UoM 1", "Contoh UoM 2"}, 4, `SELECT name FROM m_uom WHERE is_active ORDER BY name`)
	writeRef("country", 4, "Country", nil, 2, `SELECT country_code||' - '||country_name FROM m_country ORDER BY country_name`)
	for i, field := range coaFields {
		q := `SELECT a.account_code||' - '||a.account_name FROM m_coa a
			JOIN m_coa_type t ON t.account_type_code=a.account_type_code
			JOIN m_coa_classification c ON c.classification_code=t.classification_code
			WHERE a.is_active AND (` + coaRefFilter[field] + `) ORDER BY a.account_code`
		writeRef(field, 5+i, field, nil, 2, q)
	}
	f.SetColWidth(sheetRef, "A", "L", 26)

	// ── Sheet Product List ──
	f.NewSheet(sheetProduct)
	prodExamples := []map[string]string{
		{
			"product_type": "single", "barcode": "123456", "product_name": "Contoh Produk Single",
			"brand": "Brand A", "subcategory": "Subkategori A", "country": "ID", "is_perishable": "yes",
			"description": "Deskripsi produk single", "ingredients": "komposisi", "uom_1": "Pcs",
			"selling_uom": "Pcs", "stocking_uom": "Pcs", "def_selling_price": "5000", "def_purchase_price": "3000",
			"weight_gr": "80", "length_cm": "16", "width_cm": "3", "height_cm": "11",
		},
		{
			"product_type": "variant", "product_name": "Contoh Produk Variant", "variant_name_1": "Warna",
			"variant_name_2": "Ukuran", "brand": "Brand B", "subcategory": "Subkategori B", "country": "ID",
			"is_perishable": "no", "description": "Deskripsi produk variant", "uom_1": "Pcs", "uom_2": "Lusin",
			"ratio_2": "12", "selling_uom": "Pcs", "stocking_uom": "Lusin", "length_cm": "25", "width_cm": "15", "height_cm": "2",
		},
	}
	writeSheet(f, sheetProduct, productCols, productRequired, productOptional, productNotes, prodExamples)

	// ── Sheet Variant List ──
	f.NewSheet(sheetVariant)
	varExamples := []map[string]string{
		{"product_name": "Contoh Produk Variant", "barcode": "98765", "variant_value_1": "Hitam", "variant_value_2": "S", "def_selling_price": "35000", "def_purchase_price": "25000", "weight_gr": "100"},
		{"product_name": "Contoh Produk Variant", "barcode": "98761", "variant_value_1": "Navy", "variant_value_2": "L", "def_selling_price": "38000", "def_purchase_price": "25000", "weight_gr": "98"},
	}
	writeSheet(f, sheetVariant, variantCols, variantRequired, variantOptional, variantNotes, varExamples)

	// ── Data validation (dropdown) di Product List, baris input (5+) ──
	prodColL := func(colName string) string {
		for i, c := range productCols {
			if c == colName {
				l, _ := excelize.ColumnNumberToName(i + 1)
				return l
			}
		}
		return "A"
	}
	addDV := func(colName, refKey string) {
		ri, ok := refRange[refKey]
		if !ok {
			return
		}
		colL := prodColL(colName)
		refL, _ := excelize.ColumnNumberToName(ri.col)
		dv := excelize.NewDataValidation(true)
		dv.Sqref = fmt.Sprintf("%s5:%s1000", colL, colL)
		dv.SetSqrefDropList(fmt.Sprintf("%s!$%s$%d:$%s$%d", sheetRef, refL, ri.start, refL, ri.end))
		_ = f.AddDataValidation(sheetProduct, dv)
	}
	addInlineDV := func(colName string, opts []string) {
		colL := prodColL(colName)
		dv := excelize.NewDataValidation(true)
		dv.Sqref = fmt.Sprintf("%s5:%s1000", colL, colL)
		_ = dv.SetDropList(opts)
		_ = f.AddDataValidation(sheetProduct, dv)
	}
	addInlineDV("product_type", []string{"single", "variant"})
	addInlineDV("is_perishable", []string{"yes", "no"})
	addDV("brand", "brand")
	addDV("subcategory", "subcategory")
	addDV("country", "country")
	for _, u := range []string{"uom_1", "uom_2", "uom_3", "selling_uom", "stocking_uom"} {
		addDV(u, "uom")
	}
	for _, cf := range coaFields {
		addDV(cf, cf)
	}

	// ── Proteksi: sel struktur/contoh terkunci, area input bebas ──
	prot := &excelize.SheetProtectionOptions{SelectLockedCells: true, SelectUnlockedCells: true}
	_ = f.ProtectSheet(sheetProduct, prot)
	_ = f.ProtectSheet(sheetVariant, prot)
	_ = f.ProtectSheet(sheetRef, prot)

	if idx, err := f.GetSheetIndex(sheetProduct); err == nil {
		f.SetActiveSheet(idx)
	}
	c.Header("Content-Disposition", "attachment; filename=product_import_template.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// getRows membaca dua sheet dari file Excel (multipart) ATAU dari JSON hasil edit.
func (h *ProductHandler) getRows(c *gin.Context) (prod, variant []importRow, ok bool) {
	if strings.Contains(c.ContentType(), "application/json") {
		var body struct {
			ProductRows []struct {
				Row   int               `json:"row"`
				Cells map[string]string `json:"cells"`
			} `json:"product_rows"`
			VariantRows []struct {
				Row   int               `json:"row"`
				Cells map[string]string `json:"cells"`
			} `json:"variant_rows"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
			return nil, nil, false
		}
		for _, r := range body.ProductRows {
			cells := map[string]string{}
			for _, col := range productCols {
				cells[col] = trim(r.Cells[col])
			}
			prod = append(prod, importRow{rowNum: r.Row, cells: cells})
		}
		for _, r := range body.VariantRows {
			cells := map[string]string{}
			for _, col := range variantCols {
				cells[col] = trim(r.Cells[col])
			}
			variant = append(variant, importRow{rowNum: r.Row, cells: cells})
		}
		return prod, variant, true
	}
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return nil, nil, false
	}
	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot open file"})
		return nil, nil, false
	}
	defer src.Close()
	xf, err := excelize.OpenReader(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Excel file"})
		return nil, nil, false
	}
	defer xf.Close()
	prod, err = parseSheet(xf, sheetProduct, productCols)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, nil, false
	}
	variant, err = parseSheet(xf, sheetVariant, variantCols)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, nil, false
	}
	if len(prod) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No product rows found in 'Product List'"})
		return nil, nil, false
	}
	return prod, variant, true
}

// ImportExisting: SKU & barcode yang sudah ada (untuk validasi keunikan live di frontend).
func (h *ProductHandler) ImportExisting(c *gin.Context) {
	m, err := loadImportMasters(c.Request.Context(), h.repo.pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	barcodes := make([]string, 0, len(m.existBarcode))
	for k := range m.existBarcode {
		barcodes = append(barcodes, k)
	}
	products := make([]string, 0, len(m.existProduct))
	for k := range m.existProduct {
		products = append(products, k)
	}
	c.JSON(http.StatusOK, gin.H{"barcodes": barcodes, "products": products})
}

// ImportValidate: dry-run.
func (h *ProductHandler) ImportValidate(c *gin.Context) {
	prod, variant, ok := h.getRows(c)
	if !ok {
		return
	}
	m, err := loadImportMasters(c.Request.Context(), h.repo.pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, validateImport(prod, variant, m))
}

// ImportCommit: impor produk yang valid.
func (h *ProductHandler) ImportCommit(c *gin.Context) {
	prod, variant, ok := h.getRows(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	m, err := loadImportMasters(ctx, h.repo.pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := validateImport(prod, variant, m)
	userID := c.GetString(middleware.CtxUserID)
	imported, failed := 0, 0
	failures := []gin.H{}
	for _, in := range res.builds {
		if _, err := h.repo.Create(ctx, *in, userID); err != nil {
			failed++
			failures = append(failures, gin.H{"product_name": in.ProductName, "error": err.Error()})
		} else {
			imported++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"imported": imported, "failed": failed,
		"skipped": res.Totals["skip"], "errors": res.Totals["error"], "failures": failures,
	})
}

// ImportFailed: unduh Excel (format sama template) berisi baris yang gagal (font merah pada sel error).
func (h *ProductHandler) ImportFailed(c *gin.Context) {
	prod, variant, ok := h.getRows(c)
	if !ok {
		return
	}
	m, err := loadImportMasters(c.Request.Context(), h.repo.pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := validateImport(prod, variant, m)

	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName(f.GetSheetName(0), sheetProduct)
	f.NewSheet(sheetVariant)
	redStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "FF0000"}})

	writeFailed := func(sheet string, cols []string, req, opt map[string]bool, notesTxt map[string]string, rows []importRowResult) {
		colOf := map[string]int{}
		for i, cc := range cols {
			colOf[cc] = i + 1
		}
		for _, cc := range cols {
			nCell, _ := excelize.CoordinatesToCellName(colOf[cc], 1)
			f.SetCellValue(sheet, nCell, notesTxt[cc])
			hCell, _ := excelize.CoordinatesToCellName(colOf[cc], 2)
			f.SetCellValue(sheet, hCell, label(cc, req, opt))
		}
		lastCol, _ := excelize.ColumnNumberToName(len(cols) + 1)
		hs, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheet, "A2", lastCol+"2", hs)
		ec, _ := excelize.CoordinatesToCellName(len(cols)+1, 2)
		f.SetCellValue(sheet, ec, "Errors")
		out := 3
		for _, rr := range rows {
			if rr.Status != "error" {
				continue
			}
			bad := map[string]bool{}
			for _, b := range rr.BadCols {
				bad[b] = true
			}
			for i, cc := range cols {
				cell, _ := excelize.CoordinatesToCellName(i+1, out)
				f.SetCellValue(sheet, cell, rr.Cells[cc])
				if bad[cc] {
					f.SetCellStyle(sheet, cell, cell, redStyle)
				}
			}
			ecc, _ := excelize.CoordinatesToCellName(len(cols)+1, out)
			f.SetCellValue(sheet, ecc, strings.Join(rr.Errors, "; "))
			f.SetCellStyle(sheet, ecc, ecc, redStyle)
			out++
		}
	}
	writeFailed(sheetProduct, productCols, productRequired, productOptional, productNotes, res.ProductRows)
	writeFailed(sheetVariant, variantCols, variantRequired, variantOptional, variantNotes, res.VariantRows)
	if idx, err := f.GetSheetIndex(sheetProduct); err == nil {
		f.SetActiveSheet(idx)
	}
	c.Header("Content-Disposition", "attachment; filename=failed_import.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
