<script setup lang="ts">
  // Import Produk — wizard: Upload → Preview (2 tab: Product List + Variant List) → Done.
  const open = defineModel<boolean>({ default: false })
  const emit = defineEmits<{ imported: [] }>()
  const toast = useToast()

  interface RowRes {
    row: number
    cells: Record<string, string>
    exists: boolean
    status: string // 'valid' | 'error' | 'skip_exists'
    bad: string[]
    cellErr: Record<string, string>
    editCols: string[]
  }
  interface SrvRow { row: number; status: string; cells: Record<string, string> }
  interface ValidateRes { product_columns: string[]; variant_columns: string[]; product_rows: SrvRow[]; variant_rows: SrvRow[] }
  interface CommitRes { imported: number; failed: number; skipped: number; errors: number; failures: { product_name: string; error: string }[] }

  const step = ref<'upload' | 'preview' | 'done'>('upload')
  const file = ref<File | null>(null)
  const busy = ref(false)
  const productColumns = ref<string[]>([])
  const variantColumns = ref<string[]>([])
  const productRows = ref<RowRes[]>([])
  const variantRows = ref<RowRes[]>([])
  const activeTab = ref<'product' | 'variant'>('product')
  const commitRes = ref<CommitRes | null>(null)
  const fileInput = ref<HTMLInputElement>()

  watch(open, (v) => { if (v) { reset(); loadMasters() } })
  function reset() {
    step.value = 'upload'; file.value = null; productColumns.value = []; variantColumns.value = []
    productRows.value = []; variantRows.value = []; activeTab.value = 'product'; commitRes.value = null; busy.value = false
  }

  const stepIdx = computed(() => (step.value === 'upload' ? 0 : step.value === 'preview' ? 1 : 2))
  const steps = [
    { key: 'upload', title: 'Upload file', label: 'Upload file', sub: 'Choose the Excel file to import. Nothing is saved until you review it.' },
    { key: 'preview', title: 'Preview', label: 'Preview', sub: 'Red cells are invalid — click a red cell to fix it. Product List = products, Variant List = each variant.' },
    { key: 'done', title: 'Import result', label: 'Done', sub: 'Here is the result of your import.' },
  ]
  function stepState(i: number) { return stepIdx.value > i ? 'done' : stepIdx.value === i ? 'active' : 'todo' }

  const low = (s: string) => (s || '').trim().toLowerCase()
  const tr = (s: string) => (s || '').trim()
  function parseNum(s: string): number | null {
    const t = tr(s); if (t === '') return 0
    const v = Number(t.replace(/\s/g, '').replace(',', '.')); return isNaN(v) ? null : v
  }
  function leadingCode(s: string): string { const t = tr(s); const i = t.indexOf(' - '); return i >= 0 ? t.slice(0, i).trim() : t }

  // ─── Master data ───
  const { fetchBrandOptions, fetchUomOptions, fetchCountryOptions, fetchCoaOptions } = useProducts()
  const brands = ref<{ id: string; name: string }[]>([])
  const subcats = ref<string[]>([])
  const uoms = ref<{ id: string; name: string }[]>([])
  const countries = ref<{ code: string; name: string }[]>([])
  const coaAccounts = ref<CoaAccountOpt[]>([])
  const existBarcode = ref<Set<string>>(new Set())
  const existProduct = ref<Set<string>>(new Set())
  const mastersLoaded = ref(false)

  const brandSet = computed(() => new Set(brands.value.map(b => b.name)))
  const subcatSet = computed(() => new Set(subcats.value))
  const uomSet = computed(() => new Set(uoms.value.map(u => u.name)))
  const countrySet = computed(() => { const s = new Set<string>(); for (const c of countries.value) { s.add(c.code.toLowerCase()); s.add(c.name.toLowerCase()) } return s })
  const coaByCode = computed(() => new Map(coaAccounts.value.map(a => [a.account_code, a])))

  async function loadMasters() {
    if (mastersLoaded.value) return
    try {
      const [b, u, co, coa, sub, ex] = await Promise.all([
        fetchBrandOptions(), fetchUomOptions(), fetchCountryOptions(), fetchCoaOptions(),
        apiFetch<{ data: { name: string; is_active: boolean }[] }>('/api/v1/subcategories?limit=1000&sort=name'),
        apiFetch<{ barcodes: string[]; products: string[] }>('/api/v1/product-import/existing'),
      ])
      brands.value = b; uoms.value = u; countries.value = co.map(c => ({ code: c.code, name: c.name })); coaAccounts.value = coa
      subcats.value = (sub.data ?? []).filter(s => s.is_active).map(s => s.name)
      existBarcode.value = new Set(ex.barcodes ?? [])
      existProduct.value = new Set(ex.products ?? [])
      mastersLoaded.value = true
    } catch { toast.add({ title: 'Failed', description: 'Could not load master data', color: 'error' }) }
  }

  // ─── Scanner ───
  const scanOpen = ref(false)
  let scanSetter: ((code: string) => void) | null = null
  function openScanner(setter: (code: string) => void) { scanSetter = setter; scanOpen.value = true }
  function onScanDetected(code: string) { scanSetter?.(code.trim()); scanSetter = null; revalidate() }

  // ─── Konfigurasi widget per kolom ───
  const COA_KEYS = ['coa_inventory', 'coa_sales', 'coa_sales_return', 'coa_sales_discount', 'coa_good_in_transit', 'coa_cogs', 'coa_purchase_return', 'coa_unbilled_goods']
  const P_SS = new Set(['brand', 'subcategory', 'country', 'uom_1', 'uom_2', 'uom_3', 'selling_uom', 'stocking_uom'])
  function kind(sheet: string, col: string) {
    if (col === 'barcode') return 'barcode'
    if (sheet === 'product') {
      if (col === 'product_type') return 'ptype'
      if (col === 'is_perishable') return 'yesno'
      if (P_SS.has(col)) return 'ss'
      if (col.startsWith('coa_')) return 'coa'
      return 'text'
    }
    if (col === 'product_name') return 'vname'
    return 'text'
  }
  const isSelectKind = (k: string) => k === 'ss' || k === 'coa' || k === 'vname'
  const isInput = (sheet: string, r: RowRes, c: string) => r.status !== 'skip_exists' && (c === 'barcode' || r.editCols.includes(c))
  const isEditing = (r: RowRes) => r.editCols.length > 0

  // ─── Opsi dropdown ───
  const brandOpts = computed(() => brands.value.map(b => ({ value: b.name, label: b.name })))
  const subcatOpts = computed(() => subcats.value.map(n => ({ value: n, label: n })))
  const uomOpts = computed(() => uoms.value.map(u => ({ value: u.name, label: u.name })))
  const countryOpts = computed(() => [{ value: '', label: '— none —' }, ...countries.value.map(c => ({ value: c.code, label: `${c.code} - ${c.name}` }))])
  const variantNameOpts = computed(() => {
    const seen = new Set<string>(); const out: { value: string; label: string }[] = []
    for (const r of productRows.value) { if (low(r.cells.product_type) === 'variant') { const n = tr(r.cells.product_name); if (n && !seen.has(n)) { seen.add(n); out.push({ value: n, label: n }) } } }
    return out
  })
  function uomRowOpts(r: RowRes) { const uniq = [...new Set([r.cells.uom_1, r.cells.uom_2, r.cells.uom_3].map(x => tr(x)).filter(Boolean))]; return uniq.map(n => ({ value: n, label: n })) }
  const byType = (t: string) => (a: CoaAccountOpt) => a.account_type_name === t
  const byClass = (c: string, contra?: boolean) => (a: CoaAccountOpt) => a.classification_name === c && (contra === undefined || a.is_contra === contra)
  const COA_FILTERS: Record<string, (a: CoaAccountOpt) => boolean> = {
    coa_inventory: byType('Persediaan Barang'), coa_sales: byClass('Pendapatan', false),
    coa_sales_return: byClass('Pendapatan', true), coa_sales_discount: byClass('Pendapatan', true),
    coa_good_in_transit: byType('Persediaan Barang'), coa_cogs: byClass('Beban Pokok Pendapatan'),
    coa_purchase_return: byType('Persediaan Barang'), coa_unbilled_goods: byClass('Liabilitas'),
  }
  function coaOpts(c: string) { return coaAccounts.value.filter(COA_FILTERS[c]).map(a => ({ value: a.account_code, label: `${a.account_code} · ${a.account_name}` })) }
  function optionsFor(sheet: string, c: string, r: RowRes) {
    const k = kind(sheet, c)
    if (k === 'vname') return variantNameOpts.value
    if (k === 'coa') return coaOpts(c)
    if (c === 'brand') return brandOpts.value
    if (c === 'subcategory') return subcatOpts.value
    if (c === 'country') return countryOpts.value
    if (c === 'uom_1' || c === 'uom_2' || c === 'uom_3') return uomOpts.value
    if (c === 'selling_uom' || c === 'stocking_uom') return uomRowOpts(r)
    return []
  }
  function displayText(sheet: string, c: string, r: RowRes) {
    if (kind(sheet, c) === 'coa') { const a = coaByCode.value.get(leadingCode(r.cells[c])); return a ? `${a.account_code} · ${a.account_name}` : r.cells[c] }
    if (c === 'length_cm' || c === 'width_cm' || c === 'height_cm') return r.cells[c] || '0'
    return r.cells[c]
  }

  // ─── Validasi live (mirror backend) ───
  function validateProductRow(c: Record<string, string>, ctx: { nameCount: Record<string, number>; bcCount: Record<string, number>; hasVariants: (n: string) => boolean }) {
    const errs: Record<string, string> = {}
    const set = (col: string, msg: string) => { if (!errs[col]) errs[col] = msg }
    const type = low(c.product_type), isVar = type === 'variant'
    if (type !== 'single' && type !== 'variant') set('product_type', 'product_type must be "single" or "variant"')
    const name = tr(c.product_name)
    if (!name) set('product_name', 'product_name is required')
    else if (ctx.nameCount[low(name)] > 1) set('product_name', 'product_name is duplicated')
    if (!tr(c.is_perishable)) set('is_perishable', 'is_perishable is required (yes/no)')
    if (!tr(c.description)) set('description', 'description is required')
    const brand = tr(c.brand)
    if (!brand) set('brand', 'brand is required'); else if (!brandSet.value.has(brand)) set('brand', `brand "${c.brand}" not found`)
    const sub = tr(c.subcategory)
    if (!sub) set('subcategory', 'subcategory is required'); else if (!subcatSet.value.has(sub)) set('subcategory', `subcategory "${c.subcategory}" not found`)
    const country = tr(c.country)
    if (country && !countrySet.value.has(leadingCode(country).toLowerCase())) set('country', `country "${country}" not found`)
    const u1 = tr(c.uom_1), u2 = tr(c.uom_2), u3 = tr(c.uom_3)
    if (!u1) set('uom_1', 'uom_1 is required'); else if (!uomSet.value.has(u1)) set('uom_1', `uom_1 "${c.uom_1}" not found`)
    if (u2 && !uomSet.value.has(u2)) set('uom_2', `uom_2 "${c.uom_2}" not found`)
    if (u3 && !uomSet.value.has(u3)) set('uom_3', `uom_3 "${c.uom_3}" not found`)
    const u2ok = !!u2 && uomSet.value.has(u2), u3ok = !!u3 && uomSet.value.has(u3)
    const r2 = parseNum(c.ratio_2), r3 = parseNum(c.ratio_3)
    if (u2ok) { if (!(r2 !== null && r2 > 1)) set('ratio_2', 'ratio_2 must be greater than 1 when uom_2 is set') } else if (tr(c.ratio_2)) set('ratio_2', 'ratio_2 must be empty when uom_2 is empty')
    if (u3ok) { if (!(r3 !== null && r2 !== null && r3 > r2)) set('ratio_3', 'ratio_3 must be greater than ratio_2 when uom_3 is set') } else if (tr(c.ratio_3)) set('ratio_3', 'ratio_3 must be empty when uom_3 is empty')
    const rowUoms = new Set([u1, u2, u3].filter(u => u && uomSet.value.has(u)))
    const su = tr(c.selling_uom), stu = tr(c.stocking_uom)
    if (!su) set('selling_uom', 'selling_uom is required'); else if (!uomSet.value.has(su)) set('selling_uom', 'selling_uom not found'); else if (!rowUoms.has(su)) set('selling_uom', 'selling_uom must be one of uom_1/uom_2/uom_3')
    if (!stu) set('stocking_uom', 'stocking_uom is required'); else if (!uomSet.value.has(stu)) set('stocking_uom', 'stocking_uom not found'); else if (!rowUoms.has(stu)) set('stocking_uom', 'stocking_uom must be one of uom_1/uom_2/uom_3')
    for (const cf of COA_KEYS) { const code = leadingCode(c[cf]); if (code && !coaByCode.value.has(code)) set(cf, `${cf} account "${code}" not found`) }
    if (isVar) {
      if (!tr(c.variant_name_1)) set('variant_name_1', 'variant_name_1 is required for variant products')
      for (const f of ['barcode', 'def_selling_price', 'def_purchase_price', 'weight_gr']) if (tr(c[f])) set(f, `${f} must be empty for a variant (fill it in Variant List)`)
      if (name && !ctx.hasVariants(low(name))) set('product_name', 'variant product has no rows in Variant List')
    } else {
      if (tr(c.variant_name_1)) set('variant_name_1', 'variant_name_1 must be empty for a single product')
      if (tr(c.variant_name_2)) set('variant_name_2', 'variant_name_2 must be empty for a single product')
      const bc = tr(c.barcode); if (bc && (existBarcode.value.has(bc) || ctx.bcCount[bc] > 1)) set('barcode', `barcode "${bc}" already exists`)
      const sp = parseNum(c.def_selling_price); if (!(sp !== null && sp > 0)) set('def_selling_price', 'def_selling_price must be greater than 0')
      const ppRaw = tr(c.def_purchase_price)
      if (!ppRaw) set('def_purchase_price', 'def_purchase_price is required'); else { const pp = parseNum(c.def_purchase_price); if (!(pp !== null && pp >= 0)) set('def_purchase_price', 'def_purchase_price must be a number (0 or more)') }
      const wtRaw = tr(c.weight_gr)
      if (!wtRaw) set('weight_gr', 'weight_gr is required'); else { const wt = parseNum(c.weight_gr); if (!(wt !== null && wt > 0)) set('weight_gr', 'weight_gr must be greater than 0') }
    }
    return errs
  }
  function validateVariantRow(c: Record<string, string>, ctx: { prodType: Record<string, string>; prodVN2: Record<string, boolean>; bcCount: Record<string, number>; comboCount: Record<string, number> }) {
    const errs: Record<string, string> = {}
    const set = (col: string, msg: string) => { if (!errs[col]) errs[col] = msg }
    const name = low(c.product_name)
    if (!tr(c.product_name)) set('product_name', 'product_name is required')
    else if (!(name in ctx.prodType)) set('product_name', 'no matching product in Product List')
    else if (ctx.prodType[name] !== 'variant') set('product_name', 'must reference a variant-type product')
    if (!tr(c.variant_value_1)) set('variant_value_1', 'variant_value_1 is required')
    if (ctx.prodVN2[name] && !tr(c.variant_value_2)) set('variant_value_2', 'variant_value_2 is required (this product has variant_name_2)')
    const combo = `${name}||${low(c.variant_value_1)}||${low(c.variant_value_2)}`
    if (ctx.comboCount[combo] > 1) set('variant_value_1', 'duplicate variant combination')
    const bc = tr(c.barcode); if (bc && (existBarcode.value.has(bc) || ctx.bcCount[bc] > 1)) set('barcode', `barcode "${bc}" already exists`)
    for (const f of ['def_selling_price', 'def_purchase_price']) { const raw = tr(c[f]); if (raw) { const p = parseNum(c[f]); if (!(p !== null && p >= 0)) set(f, `${f} must be a number (0 or more)`) } }
    const wtRaw = tr(c.weight_gr)
    if (!wtRaw) set('weight_gr', 'weight_gr is required'); else { const wt = parseNum(c.weight_gr); if (!(wt !== null && wt > 0)) set('weight_gr', 'weight_gr must be greater than 0') }
    return errs
  }
  function revalidate() {
    const nameCount: Record<string, number> = {}, prodType: Record<string, string> = {}, prodVN2: Record<string, boolean> = {}
    const bcCount: Record<string, number> = {}, comboCount: Record<string, number> = {}
    const varByName: Record<string, RowRes[]> = {}
    for (const r of productRows.value) {
      const n = low(r.cells.product_name); if (n) { nameCount[n] = (nameCount[n] || 0) + 1; prodType[n] = low(r.cells.product_type); prodVN2[n] = !!tr(r.cells.variant_name_2) }
      const bc = tr(r.cells.barcode); if (bc) bcCount[bc] = (bcCount[bc] || 0) + 1
    }
    for (const r of variantRows.value) {
      const bc = tr(r.cells.barcode); if (bc) bcCount[bc] = (bcCount[bc] || 0) + 1
      const n = low(r.cells.product_name); (varByName[n] = varByName[n] || []).push(r)
      const combo = `${n}||${low(r.cells.variant_value_1)}||${low(r.cells.variant_value_2)}`; comboCount[combo] = (comboCount[combo] || 0) + 1
    }
    for (const r of variantRows.value) {
      if (existProduct.value.has(low(r.cells.product_name))) { r.status = 'skip_exists'; r.bad = []; r.cellErr = {}; continue }
      const errs = validateVariantRow(r.cells, { prodType, prodVN2, bcCount, comboCount })
      r.cellErr = errs; r.bad = Object.keys(errs); r.status = r.bad.length ? 'error' : 'valid'
    }
    for (const r of productRows.value) {
      if (existProduct.value.has(low(r.cells.product_name))) { r.status = 'skip_exists'; r.bad = []; r.cellErr = {}; continue }
      const errs = validateProductRow(r.cells, { nameCount, bcCount, hasVariants: (n) => (varByName[n] || []).length > 0 })
      r.cellErr = errs; r.bad = Object.keys(errs); r.status = r.bad.length ? 'error' : 'valid'
    }
  }
  const validProducts = computed(() => {
    const varByName: Record<string, RowRes[]> = {}
    for (const r of variantRows.value) { const n = low(r.cells.product_name); (varByName[n] = varByName[n] || []).push(r) }
    let n = 0
    for (const p of productRows.value) {
      if (p.status !== 'valid') continue
      if (low(p.cells.product_type) === 'single') n++
      else { const vs = varByName[low(p.cells.product_name)] || []; if (vs.length > 0 && vs.every(v => v.status === 'valid')) n++ }
    }
    return n
  })
  const allRows = computed(() => [...productRows.value, ...variantRows.value])
  const validCount = computed(() => allRows.value.filter(r => r.status === 'valid').length)
  const errorCount = computed(() => allRows.value.filter(r => r.status === 'error').length)
  const skipCount = computed(() => allRows.value.filter(r => r.status === 'skip_exists').length)
  const productErr = computed(() => productRows.value.filter(r => r.status === 'error').length)
  const variantErr = computed(() => variantRows.value.filter(r => r.status === 'error').length)

  // ─── Aksi sel ───
  function onCellClick(sheet: string, r: RowRes, c: string, e: MouseEvent) {
    if (r.status === 'skip_exists' || isInput(sheet, r, c)) return
    if (c !== 'barcode' && !r.bad.includes(c)) return
    r.editCols.push(c)
    const td = e.currentTarget as HTMLElement
    nextTick(() => { (td.querySelector('input, select') as HTMLElement | null)?.focus() })
  }
  function confirmCell(r: RowRes, c: string) { revalidate(); if (!r.bad.includes(c)) { const i = r.editCols.indexOf(c); if (i >= 0) r.editCols.splice(i, 1) } }

  function rowClass(s: string) { return s === 'valid' ? 'r-ok' : s === 'skip_exists' ? 'r-skip' : 'r-err' }
  function sortRows(list: RowRes[]) { const rank = (s: string) => (s === 'error' ? 0 : s === 'skip_exists' ? 1 : 2); return list.slice().sort((a, b) => rank(a.status) - rank(b.status)) }
  const activeColumns = computed(() => (activeTab.value === 'product' ? productColumns.value : variantColumns.value))
  const activeRows = computed(() => (activeTab.value === 'product' ? productRows.value : variantRows.value))

  // ─── I/O ───
  function download(blob: Blob, name: string) { const url = URL.createObjectURL(blob); const a = document.createElement('a'); a.href = url; a.download = name; a.click(); URL.revokeObjectURL(url) }
  async function downloadTemplate() {
    try { download(await apiFetch<Blob>('/api/v1/product-import/template', { responseType: 'blob' }), 'product_import_template.xlsx') }
    catch { toast.add({ title: 'Failed', description: 'Could not download template', color: 'error' }) }
  }
  function onFile(e: Event) { const input = e.target as HTMLInputElement; const f = input.files?.[0]; input.value = ''; if (f) file.value = f }
  function jsonPayload() {
    return {
      product_rows: productRows.value.map(r => ({ row: r.row, cells: r.cells })),
      variant_rows: variantRows.value.map(r => ({ row: r.row, cells: r.cells })),
    }
  }
  function mkRow(r: SrvRow): RowRes {
    const exists = r.status === 'skip_exists'
    return { row: r.row, cells: { ...r.cells }, exists, status: exists ? 'skip_exists' : 'error', bad: [], cellErr: {}, editCols: [] }
  }

  async function validateFile() {
    if (!file.value) return
    busy.value = true
    try {
      await loadMasters()
      const fd = new FormData(); fd.append('file', file.value)
      const res = await apiFetch<ValidateRes>('/api/v1/product-import/validate', { method: 'POST', body: fd })
      productColumns.value = res.product_columns || []
      variantColumns.value = res.variant_columns || []
      productRows.value = (res.product_rows || []).map(mkRow)
      variantRows.value = (res.variant_rows || []).map(mkRow)
      revalidate()
      productRows.value = sortRows(productRows.value)
      variantRows.value = sortRows(variantRows.value)
      activeTab.value = 'product'
      step.value = 'preview'
    } catch (e: any) {
      toast.add({ title: 'Cannot read file', description: e?.data?.error || 'Check the sheets/columns and try again', color: 'error' })
    } finally { busy.value = false }
  }
  async function commit() {
    if (validProducts.value === 0) return
    busy.value = true
    try {
      commitRes.value = await apiFetch<CommitRes>('/api/v1/product-import/commit', { method: 'POST', body: jsonPayload() })
      step.value = 'done'; emit('imported')
    } catch (e: any) { toast.add({ title: 'Import failed', description: e?.data?.error || 'Could not import', color: 'error' }) }
    finally { busy.value = false }
  }
  async function downloadFailed() {
    try { download(await apiFetch<Blob>('/api/v1/product-import/failed', { method: 'POST', body: jsonPayload(), responseType: 'blob' }), 'failed_import.xlsx') }
    catch { toast.add({ title: 'Failed', description: 'Could not download failed data', color: 'error' }) }
  }
  const hasFailed = computed(() => (commitRes.value?.errors || 0) + (commitRes.value?.failed || 0) > 0)
</script>

<template>
  <AppModal v-model="open" :title="steps[stepIdx].title" max-width="min(1200px, 96vw)">
    <p class="imp-subtitle">{{ steps[stepIdx].sub }}</p>

    <div class="imp-stepper">
      <template v-for="(s, i) in steps" :key="s.key">
        <div class="imp-stp" :class="stepState(i)">
          <span class="imp-stp-c"><UIcon v-if="stepState(i) === 'done'" name="i-lucide-check" /><template v-else>{{ i + 1 }}</template></span>
          <span class="imp-stp-l">{{ s.label }}</span>
        </div>
        <div v-if="i < steps.length - 1" class="imp-stp-line" :class="{ done: stepIdx > i }" />
      </template>
    </div>

    <div class="imp-body">
      <!-- Step 1 -->
      <div v-if="step === 'upload'" class="imp-step imp-center">
        <div class="imp-tpl-row"><button class="btn-ghost" @click="downloadTemplate"><UIcon name="i-lucide-download" /> Download template</button></div>
        <input ref="fileInput" type="file" accept=".xlsx" hidden @change="onFile">
        <div class="imp-drop" @click="fileInput?.click()">
          <UIcon :name="file ? 'i-lucide-file-spreadsheet' : 'i-lucide-upload'" class="imp-drop-ic" />
          <span class="imp-drop-title">{{ file ? file.name : 'Choose an Excel file (.xlsx)' }}</span>
          <span class="imp-drop-hint">{{ file ? 'Click Next to check the file.' : 'Two sheets: Product List + Variant List. Nothing is saved yet.' }}</span>
        </div>
      </div>

      <!-- Step 2 -->
      <div v-else-if="step === 'preview'" class="imp-step">
        <div class="imp-counts">
          <span class="c-ok"><strong>{{ validCount }}</strong> ready</span>
          <span class="c-err"><strong>{{ errorCount }}</strong> with errors</span>
          <span class="c-skip"><strong>{{ skipCount }}</strong> already exist</span>
        </div>
        <div class="imp-tabs">
          <button class="imp-tab" :class="{ active: activeTab === 'product' }" @click="activeTab = 'product'">
            Product List <span v-if="productErr" class="imp-tab-badge">{{ productErr }}</span>
          </button>
          <button class="imp-tab" :class="{ active: activeTab === 'variant' }" @click="activeTab = 'variant'">
            Variant List <span v-if="variantErr" class="imp-tab-badge">{{ variantErr }}</span>
          </button>
        </div>
        <div class="imp-table-wrap">
          <table class="imp-table">
            <thead>
              <tr>
                <th class="col-row">Row</th>
                <th v-for="c in activeColumns" :key="c">{{ c }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in activeRows" :key="r.row" :class="rowClass(r.status)">
                <td class="col-row">{{ r.row }}</td>
                <td v-for="c in activeColumns" :key="c"
                    :class="{ 'imp-cellpad': isInput(activeTab, r, c), 'imp-cell-bad': !isInput(activeTab, r, c) && r.bad.includes(c), 'imp-clickable': c === 'barcode' && !isInput(activeTab, r, c) && r.status !== 'skip_exists' }"
                    :title="r.cellErr[c] || ''" @click="onCellClick(activeTab, r, c, $event)">
                  <template v-if="isInput(activeTab, r, c)">
                    <div v-if="c === 'barcode'" class="imp-barcode-cell">
                      <input v-model="r.cells[c]" class="imp-cell-input" :class="{ bad: r.bad.includes(c) }" placeholder="Barcode" @keydown.enter.prevent="confirmCell(r, c)">
                      <button type="button" class="imp-scan" title="Scan barcode" @click="openScanner(code => { r.cells[c] = code; confirmCell(r, c) })"><UIcon name="i-lucide-scan-barcode" /></button>
                    </div>
                    <select v-else-if="kind(activeTab, c) === 'ptype'" v-model="r.cells[c]" class="imp-cell-select" :class="{ bad: r.bad.includes(c) }" @change="confirmCell(r, c)">
                      <option value="">—</option><option value="single">single</option><option value="variant">variant</option>
                    </select>
                    <select v-else-if="kind(activeTab, c) === 'yesno'" v-model="r.cells[c]" class="imp-cell-select" :class="{ bad: r.bad.includes(c) }" @change="confirmCell(r, c)">
                      <option value="">—</option><option value="yes">yes</option><option value="no">no</option>
                    </select>
                    <SelectSearch v-else-if="isSelectKind(kind(activeTab, c))" v-model="r.cells[c]" :options="optionsFor(activeTab, c, r)" :invalid="r.bad.includes(c)" placeholder="Select" class="imp-cell-ss" @change="confirmCell(r, c)" />
                    <input v-else v-model="r.cells[c]" class="imp-cell-input" :class="{ bad: r.bad.includes(c) }" @keyup.enter="confirmCell(r, c)">
                  </template>
                  <template v-else>
                    <span v-if="c.startsWith('coa_') && !r.cells[c]" class="imp-auto">[auto]</span>
                    <span v-else-if="c === 'barcode' && !r.cells[c]" class="imp-auto">[empty]</span>
                    <span v-else :class="{ 'imp-badtext': r.bad.includes(c) }">{{ displayText(activeTab, c, r) }}</span>
                  </template>
                </td>
              </tr>
              <tr v-if="!activeRows.length"><td :colspan="activeColumns.length + 1" class="imp-empty">No rows in this sheet.</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Step 3 -->
      <div v-else class="imp-step imp-center">
        <div class="imp-resultbox">
          <div class="imp-result-title">Import result</div>
          <div class="imp-pill ok">{{ commitRes?.imported || 0 }} Imported</div>
          <div class="imp-pill skip">{{ commitRes?.skipped || 0 }} Skipped</div>
          <div class="imp-pill fail">{{ (commitRes?.failed || 0) + (commitRes?.errors || 0) }} Failed</div>
        </div>
      </div>
    </div>

    <div class="modal-actions">
      <template v-if="step === 'upload'">
        <button class="btn-ghost" @click="open = false">Cancel</button>
        <button class="btn-primary" :disabled="!file || busy" @click="validateFile">{{ busy ? 'Checking…' : 'Next' }}</button>
      </template>
      <template v-else-if="step === 'preview'">
        <button class="btn-ghost" :disabled="busy" @click="step = 'upload'">Back</button>
        <button class="btn-primary" :disabled="busy || validProducts === 0" @click="commit">{{ busy ? 'Importing…' : 'Import' }}</button>
      </template>
      <template v-else>
        <button v-if="hasFailed" class="btn-ghost" @click="downloadFailed"><UIcon name="i-lucide-download" /> Download failed data</button>
        <button class="btn-primary" @click="open = false">Close</button>
      </template>
    </div>

    <AppBarcodeScanner v-model="scanOpen" :confirm="2" @detected="onScanDetected" />
  </AppModal>
</template>

<style scoped>
  .imp-subtitle { font-size: 13px; color: var(--text-secondary); margin: -4px 0 18px; line-height: 1.5; }
  .imp-stepper { display: flex; align-items: center; justify-content: center; margin: 6px 0 0; padding-bottom: 22px; border-bottom: 1px solid var(--border-color); }
  .imp-stp { display: flex; align-items: center; gap: 8px; }
  .imp-stp-c { width: 26px; height: 26px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 700; flex-shrink: 0; }
  .imp-stp-l { font-size: 13px; white-space: nowrap; }
  .imp-stp.active .imp-stp-c, .imp-stp.done .imp-stp-c { background: var(--accent); color: #fff; }
  .imp-stp.active .imp-stp-l { color: var(--text-primary); font-weight: 700; }
  .imp-stp.done .imp-stp-l { color: var(--text-secondary); }
  .imp-stp.todo .imp-stp-c { background: var(--bg-muted); color: var(--text-muted); }
  .imp-stp.todo .imp-stp-l { color: var(--text-muted); }
  .imp-stp-line { width: 64px; max-width: 12vw; height: 2px; background: var(--border-color); margin: 0 12px; }
  .imp-stp-line.done { background: var(--accent); }

  .imp-body { min-height: 440px; padding-top: 24px; display: flex; flex-direction: column; }
  .imp-step { flex: 1; display: flex; flex-direction: column; }
  .imp-center { justify-content: center; }

  .imp-tpl-row { display: flex; justify-content: center; margin-bottom: 16px; }
  .imp-drop { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; min-height: 200px; padding: 30px 20px; border: 1.5px dashed var(--border-color); border-radius: 14px; background: var(--bg-muted); cursor: pointer; text-align: center; transition: border-color 0.15s; }
  .imp-drop:hover { border-color: var(--accent); }
  .imp-drop-ic { font-size: 32px; color: var(--text-muted); }
  .imp-drop-title { font-size: 15px; font-weight: 700; color: var(--text-primary); }
  .imp-drop-hint { font-size: 12px; color: var(--text-muted); }

  .imp-counts { display: flex; gap: 18px; flex-wrap: wrap; align-items: center; margin-bottom: 12px; font-size: 13px; }
  .imp-counts strong { font-size: 15px; }
  .imp-counts .c-ok { color: var(--success); }
  .imp-counts .c-err { color: var(--danger); }
  .imp-counts .c-skip { color: var(--text-muted); }

  .imp-tabs { display: flex; gap: 6px; margin-bottom: 10px; }
  .imp-tab { display: inline-flex; align-items: center; gap: 7px; padding: 8px 16px; border: 1px solid var(--border-color); border-bottom: none; border-radius: 9px 9px 0 0; background: var(--bg-muted); color: var(--text-secondary); font-size: 13px; font-weight: 700; cursor: pointer; }
  .imp-tab.active { background: var(--bg-surface); color: var(--text-primary); border-color: var(--accent); }
  .imp-tab-badge { display: inline-flex; align-items: center; justify-content: center; min-width: 18px; height: 18px; padding: 0 5px; border-radius: 9px; background: var(--danger); color: #fff; font-size: 11px; }

  .imp-table-wrap { overflow: auto; max-height: 46vh; border: 1px solid var(--border-color); border-radius: 0 10px 10px 10px; }
  .imp-table { border-collapse: collapse; font-size: 12px; white-space: nowrap; }
  .imp-table th, .imp-table td { padding: 6px 10px; border-bottom: 1px solid var(--border-color); border-right: 1px solid var(--border-color); text-align: left; height: 38px; }
  .imp-table thead th { position: sticky; top: 0; z-index: 2; background: var(--bg-muted); color: var(--text-secondary); font-weight: 700; }
  .imp-table tbody tr.r-ok td { background: var(--success-light); }
  .imp-table tbody tr.r-err td { background: var(--danger-light); color: var(--danger); }
  .imp-table tbody tr.r-skip td { background: var(--warning-light); }
  .imp-empty { text-align: center; color: var(--text-muted); padding: 24px; }

  .col-row { position: sticky; left: 0; z-index: 1; width: 46px; min-width: 46px; font-weight: 700; color: var(--text-secondary); }
  .imp-table thead th.col-row { z-index: 3; }
  .imp-table tbody tr.r-ok td.col-row { background: var(--success-light); }
  .imp-table tbody tr.r-err td.col-row { background: var(--danger-light); }
  .imp-table tbody tr.r-skip td.col-row { background: var(--warning-light); }

  .imp-cellpad { padding: 3px 5px; }
  .imp-auto { color: var(--text-muted); font-style: italic; }
  .imp-badtext { color: var(--danger); font-weight: 600; }
  .imp-table tbody tr.r-err td.imp-cell-bad { background: var(--bg-surface); box-shadow: inset 0 0 0 1.6px var(--danger); border-radius: 5px; cursor: pointer; }
  .imp-clickable { cursor: pointer; }
  .imp-cell-input { width: 120px; min-width: 84px; padding: 5px 7px; border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-surface); font-size: 12px; color: var(--text-primary); outline: none; }
  .imp-cell-input:focus { border-color: var(--accent); box-shadow: 0 0 0 2px var(--danger-light); }
  .imp-cell-input.bad { border-color: var(--danger); }
  .imp-cell-select { width: 100px; padding: 5px 7px; border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-surface); font-size: 12px; color: var(--text-primary); outline: none; cursor: pointer; }
  .imp-cell-select.bad { border-color: var(--danger); }
  .imp-cell-ss { min-width: 150px; }
  .imp-barcode-cell { display: flex; align-items: center; gap: 4px; }
  .imp-barcode-cell .imp-cell-input { width: 110px; }
  .imp-scan { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; flex-shrink: 0; border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-muted); color: var(--text-secondary); cursor: pointer; font-size: 14px; }
  .imp-scan:hover { background: var(--bg-hover); color: var(--accent); border-color: var(--accent); }

  .imp-resultbox { align-self: center; display: flex; flex-direction: column; align-items: center; gap: 10px; }
  .imp-result-title { font-size: 19px; font-weight: 800; color: var(--accent); margin-bottom: 6px; }
  .imp-pill { min-width: 150px; text-align: center; padding: 8px 22px; border-radius: 999px; font-size: 14px; font-weight: 700; color: #fff; }
  .imp-pill.ok { background: var(--success); }
  .imp-pill.skip { background: var(--warning); }
  .imp-pill.fail { background: var(--danger); }

  .modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 22px; padding-top: 16px; border-top: 1px solid var(--border-color); }
  .btn-ghost { display: inline-flex; align-items: center; gap: 6px; padding: 10px 18px; border-radius: 10px; background: var(--bg-muted); color: var(--text-secondary); font-size: 14px; font-weight: 700; border: none; cursor: pointer; }
  .btn-ghost:hover { background: var(--bg-hover); color: var(--text-primary); }
  .btn-ghost:disabled { opacity: 0.5; cursor: default; }
</style>
