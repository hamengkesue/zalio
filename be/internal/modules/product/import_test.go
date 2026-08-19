package product

import "testing"

func TestImportSingleAndVariant(t *testing.T) {
	m := testMasters()
	prod := []importRow{
		prow(3, singleBase("Single A")),
		prow(4, variantProduct("Var B")),
	}
	varn := []importRow{
		vrow(3, map[string]string{"product_name": "Var B", "variant_value_1": "Red", "weight_gr": "50"}),
		vrow(4, map[string]string{"product_name": "Var B", "variant_value_1": "Blue", "weight_gr": "50"}),
	}
	res := validateImport(prod, varn, m)
	if res.Totals["valid_products"] != 2 {
		t.Fatalf("valid_products=%d want 2; prodRows=%+v", res.Totals["valid_products"], res.ProductRows)
	}
	if res.Totals["error"] != 0 {
		t.Fatalf("error=%d want 0", res.Totals["error"])
	}
	if len(res.builds) != 2 {
		t.Fatalf("builds=%d want 2", len(res.builds))
	}
}

func TestVariantWithoutVariantRowsErrors(t *testing.T) {
	m := testMasters()
	prod := []importRow{prow(3, variantProduct("Var X"))}
	res := validateImport(prod, nil, m)
	if res.Totals["valid_products"] != 0 || res.ProductRows[0].Status != "error" {
		t.Fatalf("expected variant product error (no variant rows), got %+v", res.ProductRows[0])
	}
}

func TestSingleMustNotHaveVariantFields(t *testing.T) {
	m := testMasters()
	kv := singleBase("Single Bad")
	kv["variant_name_1"] = "Color"
	res := validateImport([]importRow{prow(3, kv)}, nil, m)
	if res.ProductRows[0].Status != "error" {
		t.Fatalf("single with variant_name_1 should error, got %+v", res.ProductRows[0])
	}
}

func TestVariantValue2RequiredWhenName2Set(t *testing.T) {
	m := testMasters()
	pv := variantProduct("Var Y")
	pv["variant_name_2"] = "Size" // produk punya axis ke-2
	prod := []importRow{prow(3, pv)}
	varn := []importRow{
		vrow(3, map[string]string{"product_name": "Var Y", "variant_value_1": "Red", "variant_value_2": "", "weight_gr": "50"}),
	}
	res := validateImport(prod, varn, m)
	if res.VariantRows[0].Status != "error" {
		t.Fatalf("variant_value_2 should be required when product has variant_name_2, got %+v", res.VariantRows[0])
	}
	bad := false
	for _, b := range res.VariantRows[0].BadCols {
		if b == "variant_value_2" {
			bad = true
		}
	}
	if !bad {
		t.Fatalf("expected variant_value_2 in bad_cols, got %+v", res.VariantRows[0].BadCols)
	}
}

func TestVariantRowOrphan(t *testing.T) {
	m := testMasters()
	varn := []importRow{vrow(3, map[string]string{"product_name": "Nope", "variant_value_1": "Red", "weight_gr": "50"})}
	res := validateImport(nil, varn, m)
	if res.VariantRows[0].Status != "error" {
		t.Fatalf("orphan variant row should error, got %+v", res.VariantRows[0])
	}
}

func TestDuplicateBarcodeAcrossSheets(t *testing.T) {
	m := testMasters()
	single := singleBase("S1")
	single["barcode"] = "123"
	prod := []importRow{prow(3, single), prow(4, variantProduct("V1"))}
	varn := []importRow{vrow(3, map[string]string{"product_name": "V1", "variant_value_1": "Red", "weight_gr": "50", "barcode": "123"})}
	res := validateImport(prod, varn, m)
	// Barcode 123 muncul di single (Product List) + varian (Variant List) → keduanya error.
	if res.Totals["error"] < 2 {
		t.Fatalf("expected duplicate-barcode errors on both sheets, totals=%+v", res.Totals)
	}
}

func testMasters() *importMasters {
	m := &importMasters{
		brands:       map[string]string{"Brand A": "b1", "Brand B": "b2"},
		subcats:      map[string]string{"Sub A": "s1", "Sub B": "s2"},
		uoms:         map[string]string{"Pcs": "u1", "Lusin": "u2"},
		countries:    map[string]string{"indonesia": "ID", "id": "ID"},
		coa:          map[string]coaRef{},
		existProduct: map[string]bool{},
		existBarcode: map[string]bool{},
	}
	for _, code := range coaDefaults {
		m.coa[code] = coaRef{id: "coa-" + code}
	}
	return m
}

func prow(rowNum int, kv map[string]string) importRow {
	cells := map[string]string{}
	for _, c := range productCols {
		cells[c] = kv[c]
	}
	return importRow{rowNum: rowNum, cells: cells}
}
func vrow(rowNum int, kv map[string]string) importRow {
	cells := map[string]string{}
	for _, c := range variantCols {
		cells[c] = kv[c]
	}
	return importRow{rowNum: rowNum, cells: cells}
}

func singleBase(name string) map[string]string {
	return map[string]string{
		"product_type": "single", "product_name": name, "brand": "Brand A", "subcategory": "Sub A",
		"is_perishable": "no", "description": "d", "uom_1": "Pcs", "selling_uom": "Pcs", "stocking_uom": "Pcs",
		"def_selling_price": "1000", "def_purchase_price": "700", "weight_gr": "100",
	}
}
func variantProduct(name string) map[string]string {
	return map[string]string{
		"product_type": "variant", "product_name": name, "variant_name_1": "Color", "brand": "Brand B",
		"subcategory": "Sub B", "is_perishable": "no", "description": "d", "uom_1": "Pcs",
		"selling_uom": "Pcs", "stocking_uom": "Pcs",
	}
}
