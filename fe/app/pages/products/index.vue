<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Products' })

  const {
    items, total, fetchPage, getProduct, getNextSku, createProduct, updateProduct, toggleVariantActive, uploadImage,
    fetchBrandOptions, fetchCategoryOptions, fetchSubcategoryOptions, fetchUomOptions, fetchCountryOptions, fetchCoaOptions,
    createBrand, createCategory, createSubcategory,
  } = useProducts()
  const toast = useToast()

  // ── search / sort / infinite scroll ──
  const search = ref('')
  const sortField = ref('brand')
  const sortDesc = ref(false)
  const pageSize = 8
  const currentPage = ref(1)
  const loading = ref(false)
  const scrollEl = ref<HTMLElement>()
  const sortOptions = [
    { label: 'SKU', value: 'sku' },
    { label: 'Product Name', value: 'name' },
    { label: 'Type', value: 'type' },
    { label: 'Brand', value: 'brand' },
    { label: 'Category', value: 'category' },
    { label: 'Subcategory', value: 'subcategory' },
    { label: 'Selling Price', value: 'price' },
    { label: 'COGS', value: 'cogs' },
    { label: 'Status', value: 'status' },
  ]
  // ── Filter (server-side, lintas semua page) ──
  const filters = reactive({ product_type: '', country: '', brand_id: '', category_id: '', subcategory_id: '', status: 'active' })
  const filterSubcatOpts = ref<{ id: string; name: string }[]>([])
  const productTypeFilterOptions = [{ value: '', label: 'All types' }, { value: 'single', label: 'Single' }, { value: 'variant', label: 'Variant' }]
  const statusFilterOptions = [{ value: '', label: 'All status' }, { value: 'active', label: 'Active' }, { value: 'inactive', label: 'Inactive' }]
  // Opsi filter (dropdown) dengan opsi "All" untuk menghapus filter — terpisah dari opsi form.
  const fBrandOptions = computed(() => [{ value: '', label: 'All brands' }, ...brandOptions.value])
  const fCategoryOptions = computed(() => [{ value: '', label: 'All categories' }, ...categoryOptions.value])
  const fCountryOptions = computed(() => [{ value: '', label: 'All countries' }, ...countryOptions.value])
  const filterSubcatOptions = computed(() => [{ value: '', label: 'All subcategories' }, ...filterSubcatOpts.value.map(s => ({ value: s.id, label: s.name }))])
  // status default 'active' tidak dihitung sebagai filter aktif (badge).
  const activeFilterCount = computed(() =>
    Object.entries(filters).filter(([k, v]) => (k === 'status' ? v !== 'active' : !!v)).length)
  function resetFilters() {
    filters.product_type = filters.country = filters.brand_id = filters.category_id = filters.subcategory_id = ''
    filters.status = 'active'
  }
  // Subcategory filter ikut kategori yang dipilih di filter.
  watch(() => filters.category_id, async (cid) => {
    filters.subcategory_id = ''
    filterSubcatOpts.value = cid ? await fetchSubcategoryOptions(cid) : []
  })
  watch(filters, () => reload(), { deep: true })
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
  const ROW_H = 64

  function baseQuery() {
    return { search: search.value.trim(), sort: sortField.value, desc: sortDesc.value, ...filters }
  }
  async function loadMore() {
    if (loading.value || items.value.length >= total.value) return
    loading.value = true
    try { await fetchPage({ offset: items.value.length, limit: pageSize, ...baseQuery(), append: true }) }
    finally { loading.value = false }
    await fillViewport()
  }
  // Nama baris flat: produk single = nama produk; variant = "nama induk - value1 - value2".
  function variantName(p: any) {
    if (p.product_type !== 'variant') return p.product_name
    const combo = [p.variant_value_1, p.variant_value_2].filter(Boolean).join(' - ')
    return combo ? `${p.product_name} - ${combo}` : p.product_name
  }

  async function reload() {
    loading.value = true
    items.value = []
    currentPage.value = 1
    try { await fetchPage({ offset: 0, limit: pageSize, ...baseQuery(), append: false }) }
    finally { loading.value = false }
    if (scrollEl.value) scrollEl.value.scrollTop = 0
    await fillViewport()
  }
  async function fillViewport() {
    await nextTick()
    const el = scrollEl.value
    if (el && el.scrollHeight <= el.clientHeight + 4 && items.value.length < total.value && !loading.value) await loadMore()
  }
  function onScroll() {
    const el = scrollEl.value
    if (!el) return
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 240) loadMore()
    const tp = totalPages.value
    if (el.scrollTop <= 2) { currentPage.value = 1; return }
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 2) { currentPage.value = tp; return }
    const headerH = el.querySelector('thead')?.clientHeight ?? 0
    const centerRow = Math.floor((el.scrollTop + el.clientHeight / 2 - headerH) / ROW_H)
    currentPage.value = Math.min(tp, Math.max(1, Math.floor(Math.max(0, centerRow) / pageSize) + 1))
  }
  let searchTimer: ReturnType<typeof setTimeout> | undefined
  watch(search, () => { clearTimeout(searchTimer); searchTimer = setTimeout(reload, 300) })
  watch([sortField, sortDesc], reload)

  const formatPrice = (n: number) => 'Rp ' + Math.round(n || 0).toLocaleString('id-ID')

  // Input harga: tampil sebagai Rupiah lengkap (mis. "Rp 15.000"); hanya menerima
  // angka (karakter non-digit dibuang) → nilai tersimpan tetap number.
  const priceDisplay = (v: number) => (v ? 'Rp ' + Number(v).toLocaleString('id-ID') : '')
  function onPriceInput(e: Event, field: 'def_selling_price' | 'def_purchase_price') {
    const el = e.target as HTMLInputElement
    const digits = el.value.replace(/\D/g, '')
    const num = digits ? parseInt(digits, 10) : 0
    ;(form as any)[field] = num
    el.value = num ? 'Rp ' + num.toLocaleString('id-ID') : ''
  }
  // Versi "wajib": tak pernah kosong, minimal "Rp 0" (dipakai harga varian saat "same for all" aktif).
  const priceDisplayMin0 = (v: number) => 'Rp ' + Math.round(Number(v) || 0).toLocaleString('id-ID')
  function onPriceInputMin0(e: Event, field: 'def_selling_price' | 'def_purchase_price') {
    const el = e.target as HTMLInputElement
    const digits = el.value.replace(/\D/g, '')
    const num = digits ? parseInt(digits, 10) : 0
    ;(form as any)[field] = num
    el.value = 'Rp ' + num.toLocaleString('id-ID')
  }

  // ── options ──
  const brandOpts = ref<Opt[]>([])
  const categoryOpts = ref<Opt[]>([])
  const subcategoryOpts = ref<Opt[]>([])
  const uomOpts = ref<Opt[]>([])
  const countryOpts = ref<CountryOpt[]>([])

  // Bendera emoji dari kode ISO 2 huruf (mis. "ID" → 🇮🇩); nama → Title Case.
  const flagEmoji = (code: string) =>
    code && code.length === 2
      ? String.fromCodePoint(...[...code.toUpperCase()].map(ch => 0x1f1e6 + ch.charCodeAt(0) - 65))
      : ''
  const titleCase = (s: string) => (s || '').toLowerCase().replace(/\b\w/g, m => m.toUpperCase())
  const countryMap = computed(() => new Map(countryOpts.value.map(c => [c.code, c])))
  function countryLabel(code: string) {
    const c = countryMap.value.get(code)
    return c ? `${flagEmoji(c.code)} ${titleCase(c.name)}` : (code || '—')
  }

  // ── form ──
  const showForm = ref(false)
  const editingId = ref<string | null>(null)
  const saving = ref(false)
  const emptyForm = () => ({
    product_name: '', product_type: 'single', brand_id: '', category_id: '', subcategory_id: '',
    country_of_origin: '', description: '', ingredients: '', is_perishable: true,
    uom_1: '', uom_2: '', ratio_2: 0, uom_3: '', ratio_3: 0, selling_uom: '', stocking_uom: '',
    coa_inventory: '', coa_sales: '', coa_sales_return: '', coa_sales_discount: '',
    coa_good_in_transit: '', coa_cogs: '', coa_purchase_return: '', coa_unbilled_goods: '',
    variant_name_1: '', variant_name_2: '',
    variant_id: '', sku: '', barcode: '', def_selling_price: 0, def_purchase_price: 0, cogs_unit: 0,
    length_cm: 0, width_cm: 0, height_cm: 0, weight_gr: 0,
    main_image: '', image_1: '', image_2: '', image_3: '',
  })
  const form = reactive(emptyForm())

  // ── Variant produk (product_type === 'variant') ──
  // Alur: toggle Variant → modal "Variant Options" (tentukan sumbu + centang
  // field apply-to-all) → kembali ke form → tombol "to Variant List" →
  // tabel Variant List (isi SKU/detail per baris) → Save.
  interface VariantRow {
    id: string // id varian existing (mode edit) — kosong = varian baru
    variant_value_1: string; variant_value_2: string
    sku: string; sku_auto: boolean; barcode: string
    def_selling_price: number; def_purchase_price: number; cogs_unit: number
    length_cm: number; width_cm: number; height_cm: number; weight_gr: number
    main_image: string; variant_image: string; image_1: string; image_2: string; image_3: string
    stock_qty: number // stok per varian (dari modul inventory, sementara 0)
    is_active: boolean
  }
  const variantAxes = reactive({ values1: [] as string[], values2: [] as string[] })
  const variantRows = ref<VariantRow[]>([])
  const axisInput1 = ref('')
  const axisInput2 = ref('')
  const vKey = (a: string, b: string) => `${a}||${b}`

  // Field yang bisa "diterapkan ke semua varian" (nilai default diambil dari form utama).
  const APPLY_FIELDS = [
    { key: 'main_image', label: 'Main Image', requiredInSingle: true },
    { key: 'def_selling_price', label: 'Def. Selling Price', requiredInSingle: false },
    { key: 'def_purchase_price', label: 'Def. Purchase Price', requiredInSingle: false },
    { key: 'length_cm', label: 'Length', requiredInSingle: false },
    { key: 'width_cm', label: 'Width', requiredInSingle: false },
    { key: 'height_cm', label: 'Height', requiredInSingle: false },
    { key: 'weight_gr', label: 'Weight', requiredInSingle: true },
  ] as const
  const applyToAll = reactive<Record<string, boolean>>({
    main_image: true, def_selling_price: true, def_purchase_price: true,
    length_cm: true, width_cm: true, height_cm: true, weight_gr: true,
  })
  const variantStep = ref<'form' | 'list'>('form') // 'list' = Variant Options + tabel Variant List
  const productActive = ref(true) // status aktif produk induk (mode edit) — dari DB
  // Sumbu/value yang sudah tersimpan di DB (mode edit) → dikunci, tak boleh diubah/hapus.
  const variantLocked = reactive({ values1: [] as string[], values2: [] as string[] })
  const nameLocked = (axis: 1 | 2) => !!editingId.value && !!(axis === 1 ? form.variant_name_1 : form.variant_name_2).trim()
  const isLockedValue = (axis: 1 | 2, val: string) => !!editingId.value && (axis === 1 ? variantLocked.values1 : variantLocked.values2).includes(val)

  // true bila checkbox apply-to-all di-UNCHECK → field di form utama disable, dan
  // kolomnya di Variant List jadi editable per baris. (main_image & dimensi selalu
  // true di applyToAll → applyOff-nya selalu false.)
  function applyOff(key: string) { return form.product_type === 'variant' && !applyToAll[key] }
  // Harga wajib (variant + checkbox "same for all" aktif) → tak boleh kosong, minimal Rp 0.
  function priceRequired(field: string) { return form.product_type === 'variant' && applyToAll[field] }
  // Toggle "same for all": saat di-uncheck, kosongkan nilai form (diisi per varian di Variant List).
  function onApplyToggle(field: string) {
    if (applyToAll[field]) return
    if (field === 'weight_gr') { form.weight_gr = 0; dimInput.weight_gr = '' }
    else (form as any)[field] = 0
    // Uncheck → kolom jadi per-baris; semua baris di Variant List di-reset ke 0.
    variantRows.value.forEach(r => { (r as any)[field] = 0 })
    if (field === 'weight_gr') for (const k in vNum) if (k.endsWith(':weight_gr')) delete vNum[k]
  }

  function blankVariant(v1: string, v2: string): VariantRow {
    return {
      id: '', variant_value_1: v1, variant_value_2: v2, sku: '', sku_auto: skuMode.value === 'auto', barcode: '',
      // Harga/berat per-baris hanya dipakai saat checkbox di-uncheck → mulai 0.
      def_selling_price: 0, def_purchase_price: 0, cogs_unit: 0,
      // Dimensi selalu ikut form (tak ada kolom di tabel Variant List).
      length_cm: Number(form.length_cm) || 0, width_cm: Number(form.width_cm) || 0, height_cm: Number(form.height_cm) || 0,
      weight_gr: 0,
      main_image: form.main_image,    // main image induk (sama untuk semua varian)
      variant_image: form.main_image, // default gambar varian = main image induk, bisa diganti per baris
      image_1: '', image_2: '', image_3: '', stock_qty: 0, is_active: true,
    }
  }
  // Regenerasi matrix kombinasi; pertahankan data baris yang sudah diisi.
  function regenVariants() {
    const existing = new Map(variantRows.value.map(r => [vKey(r.variant_value_1, r.variant_value_2), r]))
    const v1 = variantAxes.values1.filter(Boolean)
    const v2 = variantAxes.values2.filter(Boolean)
    const rows: VariantRow[] = []
    for (const a of v1) {
      if (v2.length) for (const b of v2) rows.push(existing.get(vKey(a, b)) ?? blankVariant(a, b))
      else rows.push(existing.get(vKey(a, '')) ?? blankVariant(a, ''))
    }
    variantRows.value = rows
  }
  // Sumbu diedit langsung di step Variant List → tabel ikut ter-generate otomatis.
  watch(() => [variantAxes.values1.slice(), variantAxes.values2.slice()], regenVariants, { deep: true })

  function addAxisValue(axis: 1 | 2) {
    const inp = axis === 1 ? axisInput1 : axisInput2
    const arr = axis === 1 ? variantAxes.values1 : variantAxes.values2
    const val = inp.value.trim()
    if (val && !arr.some(x => x.toLowerCase() === val.toLowerCase())) arr.push(val)
    inp.value = ''
  }
  function removeAxisValue(axis: 1 | 2, i: number) {
    const arr = axis === 1 ? variantAxes.values1 : variantAxes.values2
    if (isLockedValue(axis, arr[i])) return // value existing (DB) tak boleh dihapus
    arr.splice(i, 1)
  }

  // Toggle tipe (radio). Variant → tetap di form (TIDAK ada modal). Sumbu didefinisikan
  // nanti di step Variant List.
  function onSelectType() { variantStep.value = 'form' }

  function goToVariantList() {
    if (!variantFormValid.value) return
    regenVariants()
    variantStep.value = 'list'
  }
  function backToVariantForm() { variantStep.value = 'form' }

  const variantSummary = computed(() => {
    const p1 = form.variant_name_1 ? `${form.variant_name_1} (${variantAxes.values1.join(', ')})` : ''
    const p2 = form.variant_name_2 && variantAxes.values2.length ? `${form.variant_name_2} (${variantAxes.values2.join(', ')})` : ''
    return [p1, p2].filter(Boolean).join('  ×  ')
  })
  const variantCombinations = computed(() => {
    const a = variantAxes.values1.filter(Boolean).length
    const b = variantAxes.values2.filter(Boolean).length
    return a * (b || 1)
  })
  // "to Variant List" aktif bila field WAJIB di form terisi (sumbu belum perlu di sini).
  const variantFormValid = computed(() => {
    if (form.product_type !== 'variant') return false
    const sharedOk = form.product_name.trim() && form.brand_id && form.category_id && form.subcategory_id
      && form.uom_1 && form.selling_uom && form.stocking_uom && htmlHasText(form.description)
      && !ratio2Err.value && !ratio3Err.value
    const imgOk = !!form.main_image                                     // main image selalu wajib
    const weightOk = !applyToAll.weight_gr || Number(form.weight_gr) > 0 // weight wajib hanya bila dicentang
    // Harga "same for all" (dicentang) tidak boleh 0.
    const priceOk = (!applyToAll.def_selling_price || Number(form.def_selling_price) > 0)
      && (!applyToAll.def_purchase_price || Number(form.def_purchase_price) > 0)
    return !!(sharedOk && imgOk && weightOk && priceOk)
  })
  // Sumbu valid: Variant Name 1 + ≥1 value; bila Name 2 diisi → wajib ada Values 2.
  const variantAxesValid = computed(() => {
    const a1 = form.variant_name_1.trim() && variantAxes.values1.filter(Boolean).length > 0
    const a2 = !form.variant_name_2.trim() || variantAxes.values2.filter(Boolean).length > 0
    return !!(a1 && a2)
  })
  // Save aktif: sumbu valid + ada baris + tiap baris punya SKU.
  const variantListValid = computed(() => {
    if (!variantAxesValid.value || !variantRows.value.length) return false
    return variantRows.value.every(r => r.sku_auto || !!r.sku.trim())
  })

  function resetVariants() {
    variantAxes.values1 = []
    variantAxes.values2 = []
    variantRows.value = []
    axisInput1.value = ''
    axisInput2.value = ''
    variantStep.value = 'form'
    Object.keys(applyToAll).forEach(k => { applyToAll[k] = true })
    for (const k in vNum) delete vNum[k]
  }

  // Per-baris angka desimal-koma (weight/L/W/H). Buffer teks saat aktif diedit,
  // lalu jatuh ke angka terformat setelah blur — pola sama seperti dimInput.
  const vNum = reactive<Record<string, string>>({})
  const vk = (i: number, f: string) => `${i}:${f}`
  function vNumVal(i: number, f: 'weight_gr' | 'length_cm' | 'width_cm' | 'height_cm') {
    const k = vk(i, f)
    if (k in vNum) return vNum[k]
    const n = (variantRows.value[i] as any)?.[f]
    return n ? formatDecimalComma(n) : ''
  }
  function onVarNumInput(e: Event, i: number, f: 'weight_gr' | 'length_cm' | 'width_cm' | 'height_cm') {
    const raw = (e.target as HTMLInputElement).value
    vNum[vk(i, f)] = raw
    ;(variantRows.value[i] as any)[f] = parseDecimalComma(raw)
  }
  function onVarNumBlur(i: number, f: 'weight_gr' | 'length_cm' | 'width_cm' | 'height_cm') {
    const k = vk(i, f)
    ;(variantRows.value[i] as any)[f] = parseDecimalComma(vNum[k] ?? '')
    delete vNum[k]
  }
  function onVarPriceInput(e: Event, i: number, f: 'def_selling_price' | 'def_purchase_price') {
    const el = e.target as HTMLInputElement
    const digits = el.value.replace(/\D/g, '')
    const num = digits ? parseInt(digits, 10) : 0
    ;(variantRows.value[i] as any)[f] = num
    el.value = num ? 'Rp ' + num.toLocaleString('id-ID') : ''
  }

  // Upload gambar utama per baris varian.
  const vImgInput = ref<HTMLInputElement>()
  const currentVariantIdx = ref(-1)
  const uploadingVariantIdx = ref(-1)
  function pickVariantImage(idx: number) { currentVariantIdx.value = idx; vImgInput.value?.click() }
  async function onVariantImageChange(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
    if (!file) return
    if (!file.type.startsWith('image/')) { toast.add({ title: 'Invalid file', description: 'Please choose an image', color: 'error' }); return }
    if (file.size > 5 * 1024 * 1024) { toast.add({ title: 'File too large', description: 'Max 5 MB', color: 'error' }); return }
    const idx = currentVariantIdx.value
    if (idx < 0 || !variantRows.value[idx]) return
    uploadingVariantIdx.value = idx
    // Gambar di Variant List = gambar utama per-varian → disimpan ke kolom variant_image.
    try { variantRows.value[idx].variant_image = await uploadImage(file) }
    catch (err: any) { toast.add({ title: 'Upload failed', description: err?.data?.error || 'Could not upload', color: 'error' }) }
    finally { uploadingVariantIdx.value = -1 }
  }
  const errors = reactive({ product_name: '', brand_id: '', category_id: '', subcategory_id: '', uom_1: '', selling_uom: '', stocking_uom: '', ratio_2: '', ratio_3: '', sku: '', weight_gr: '', description: '', variant_name_1: '', variant_values_1: '', variant_values_2: '' })

  // ── Other Information (read-only, dari DB/inventory) ──
  const totalQtyStock = ref(0) // menyusul: diambil dari modul inventory
  const totalCogs = computed(() => totalQtyStock.value * (Number(form.cogs_unit) || 0))
  const stockingUomName = computed(() => uomOpts.value.find(u => u.id === form.stocking_uom)?.name ?? '')

  // Agregasi stok untuk produk variant (dijumlahkan dari semua baris varian).
  // stock_qty per varian masih 0 sampai modul inventory tersambung; rumus sudah siap.
  const variantTotalQty = computed(() => variantRows.value.reduce((s, r) => s + (Number(r.stock_qty) || 0), 0))
  const variantTotalCogs = computed(() => variantRows.value.reduce((s, r) => s + (Number(r.stock_qty) || 0) * (Number(r.cogs_unit) || 0), 0))
  // COGS/unit = rata-rata tertimbang = Total COGS ÷ Total Qty (0 bila belum ada stok).
  const variantCogsUnit = computed(() => variantTotalQty.value > 0 ? variantTotalCogs.value / variantTotalQty.value : 0)

  // Rasio stocking uom terhadap UoM 1 (basis penyimpanan cogs_unit):
  // 1 bila stocking = uom_1, ratio_2 bila = uom_2, ratio_3 bila = uom_3.
  const stockingRatio = computed(() => {
    if (form.stocking_uom && form.stocking_uom === form.uom_2) return Number(form.ratio_2) || 1
    if (form.stocking_uom && form.stocking_uom === form.uom_3) return Number(form.ratio_3) || 1
    return 1
  })

  // Nilai yang ditampilkan di kartu Stock — ikut mode (variant = agregasi baris).
  // COGS/unit & Total COGS disesuaikan ke stocking uom (× stockingRatio), sejalan
  // dengan kolom COGS/unit di tabel produk yang per selling uom.
  const isVariantMode = computed(() => form.product_type === 'variant')
  const displayTotalQty = computed(() => isVariantMode.value ? variantTotalQty.value : totalQtyStock.value)
  const displayCogsUnit = computed(() => (isVariantMode.value ? variantCogsUnit.value : (Number(form.cogs_unit) || 0)) * stockingRatio.value)
  const displayTotalCogs = computed(() => (isVariantMode.value ? variantTotalCogs.value : totalCogs.value) * stockingRatio.value)

  // Nilai untuk ilustrasi dimensi (koma desimal, "–" bila kosong)
  const dimText = computed(() => {
    const n = (v: number) => (v ? String(v).replace('.', ',') : '')
    return {
      l: n(form.length_cm), w: n(form.width_cm), h: n(form.height_cm),
      wt: form.weight_gr ? String(form.weight_gr).replace('.', ',') : '',
    }
  })

  // ── SKU: mode Auto (di-generate sistem) vs Manual (diketik user) ──
  const skuMode = ref<'auto' | 'manual'>('auto')
  const skuPreview = ref('')
  async function loadSkuPreview() {
    skuPreview.value = ''
    try { skuPreview.value = await getNextSku() } catch { skuPreview.value = '' }
  }
  function setSkuMode(m: 'auto' | 'manual') {
    if (skuMode.value === m) return
    skuMode.value = m
    errors.sku = ''
    if (m === 'auto') { form.sku = ''; loadSkuPreview() }
    else form.sku = ''
  }
  const ratio2Text = ref('')
  const ratio3Text = ref('')

  const uom1Selected = computed(() => !!form.uom_1)
  const uom2Selected = computed(() => !!form.uom_2)
  const uom3Selected = computed(() => !!form.uom_3)
  const ratio2Valid = computed(() => Number(form.ratio_2) > 0)
  const ratio3Valid = computed(() => Number(form.ratio_3) > 0)
  const uom1Disabled = computed(() => !!form.uom_2)
  const uom2Disabled = computed(() => !uom1Selected.value || !!form.ratio_2)
  const ratio2Disabled = computed(() => !uom2Selected.value || !!form.uom_3)
  const uom3Disabled = computed(() => !uom1Selected.value || !uom2Selected.value || !ratio2Valid.value || !!form.ratio_3)
  const ratio3Disabled = computed(() => !uom3Selected.value)
  const sellingUomDisabled = computed(() => !uom1Selected.value)

  // nama UoM 1 (teks statis di samping ratio) & validasi ratio LIVE (tanpa klik Save)
  const uom1Name = computed(() => uomOpts.value.find(u => u.id === form.uom_1)?.name ?? '')
  const ratio2Err = computed(() => (uom2Selected.value && !ratio2Valid.value) ? 'required' : '')
  const ratio3Err = computed(() => {
    if (!uom3Selected.value) return ''
    if (!ratio3Valid.value) return 'required'
    if (Number(form.ratio_3) <= Number(form.ratio_2)) return 'must be greater than ratio 2'
    return ''
  })

  function formatRatioText(value: number | string | null | undefined) {
    if (value === null || value === undefined || value === '') return ''
    const num = Number(value)
    if (!Number.isFinite(num)) return ''
    if (Number.isInteger(num)) return String(num)
    return num.toFixed(1).replace('.', ',')
  }

  function parseRatioInput(raw: string) {
    const normalized = raw.replace(/\s/g, '').trim()
    if (!normalized) return 0

    const safe = normalized.replace(/\./g, '').replace(/[^0-9,]/g, '')
    if (!safe) return 0

    if (safe.includes(',')) {
      const [whole, frac] = safe.split(',')
      const safeWhole = (whole || '').replace(/^0+(?=\d)/, '') || '0'
      const safeFrac = (frac || '').slice(0, 1)
      if (!safeFrac) return Number(safeWhole) || 0
      return Number(`${safeWhole}.${safeFrac}`)
    }

    return Number(safe) || 0
  }

  function onRatioInput(e: Event, field: 'ratio_2' | 'ratio_3') {
    const el = e.target as HTMLInputElement
    const raw = el.value.replace(/\./g, '').replace(/[^0-9,]/g, '')
    if (field === 'ratio_2') ratio2Text.value = raw
    else ratio3Text.value = raw
    el.value = raw
    ;(form as any)[field] = parseRatioInput(raw) // update nilai numerik live → validasi reaktif
  }

  function onRatioFocus(e: Event, field: 'ratio_2' | 'ratio_3') {
    const el = e.target as HTMLInputElement
    const value = el.value
    if (!value || value === '0' || value === '0,0') {
      el.value = ''
      if (field === 'ratio_2') ratio2Text.value = ''
      else ratio3Text.value = ''
    }
    el.select()
  }

  function onRatioBlur(e: Event, field: 'ratio_2' | 'ratio_3') {
    const el = e.target as HTMLInputElement
    const value = parseRatioInput(el.value)
    ;(form as any)[field] = value
    if (field === 'ratio_2') ratio2Text.value = formatRatioText(value)
    else ratio3Text.value = formatRatioText(value)
    el.value = field === 'ratio_2' ? ratio2Text.value : ratio3Text.value
  }

  // Isi teks tampilan ratio dari nilai numerik form (dipakai saat openEdit).
  function syncRatioText() {
    ratio2Text.value = formatRatioText(form.ratio_2)
    ratio3Text.value = formatRatioText(form.ratio_3)
    dimInput.length_cm = formatDecimalComma(form.length_cm)
    dimInput.width_cm = formatDecimalComma(form.width_cm)
    dimInput.height_cm = formatDecimalComma(form.height_cm)
    dimInput.weight_gr = formatDecimalComma(form.weight_gr)
  }

  // ── Dimensi & Weight: input teks dengan desimal koma (mis. "16,5") ──
  type DimField = 'length_cm' | 'width_cm' | 'height_cm' | 'weight_gr'
  const dimInput = reactive({ length_cm: '', width_cm: '', height_cm: '', weight_gr: '' })
  function parseDecimalComma(raw: string | number) {
    const s = String(raw).replace(/\./g, '').replace(/[^0-9,]/g, '')
    if (!s) return 0
    if (s.includes(',')) {
      const [w, f] = s.split(',')
      return Number(`${w || '0'}.${(f || '').replace(/[^0-9]/g, '')}`) || 0
    }
    return Number(s) || 0
  }
  function formatDecimalComma(v: number | string | null | undefined) {
    if (v === null || v === undefined || v === '' || Number(v) === 0) return ''
    const n = Number(v)
    return Number.isFinite(n) ? String(n).replace('.', ',') : ''
  }
  function onDimInput(e: Event, field: DimField) {
    const el = e.target as HTMLInputElement
    const raw = el.value.replace(/\./g, '').replace(/[^0-9,]/g, '')
    dimInput[field] = raw
    el.value = raw
    form[field] = parseDecimalComma(raw)
    if (field === 'weight_gr') errors.weight_gr = ''
  }
  function onDimFocus(e: Event, field: DimField) {
    const el = e.target as HTMLInputElement
    if (!el.value || el.value === '0') { el.value = ''; dimInput[field] = '' }
    el.select()
  }
  function onDimBlur(e: Event, field: DimField) {
    const el = e.target as HTMLInputElement
    const val = parseDecimalComma(el.value)
    form[field] = val
    dimInput[field] = formatDecimalComma(val)
    el.value = dimInput[field]
  }

  function getUomOptionsFor(field: 'uom_1' | 'uom_2' | 'uom_3') {
    const selectedValue = (form as Record<string, string>)[field]
    const excluded = new Set<string>()
    if (field !== 'uom_1' && form.uom_1) excluded.add(form.uom_1)
    if (field !== 'uom_2' && form.uom_2) excluded.add(form.uom_2)
    if (field !== 'uom_3' && form.uom_3) excluded.add(form.uom_3)

    return uomOpts.value.filter((option) => !excluded.has(option.id) || option.id === selectedValue)
  }

  const uom1Opts = computed(() => getUomOptionsFor('uom_1'))
  const uom2Opts = computed(() => getUomOptionsFor('uom_2'))
  const uom3Opts = computed(() => getUomOptionsFor('uom_3'))

  function syncUomState() {
    if (!form.uom_1) {
      form.selling_uom = ''
      form.stocking_uom = ''
      return
    }
    if (!form.selling_uom || !sellingUomOpts.value.some((option) => option.id === form.selling_uom)) {
      form.selling_uom = form.uom_1
    }
    if (!form.stocking_uom || !sellingUomOpts.value.some((option) => option.id === form.stocking_uom)) {
      form.stocking_uom = form.uom_1
    }
  }

  function syncUomSelections(changedField: 'uom_1' | 'uom_2' | 'uom_3') {
    if (changedField === 'uom_1') {
      errors.uom_1 = ''
      if (form.uom_1) {
        form.selling_uom = form.uom_1
        form.stocking_uom = form.uom_1
      } else {
        form.selling_uom = ''
        form.stocking_uom = ''
      }
    }

    if (form.uom_2 && (form.uom_2 === form.uom_1 || form.uom_2 === form.uom_3)) {
      form.uom_2 = ''
      form.ratio_2 = 0
    }
    if (form.uom_3 && (form.uom_3 === form.uom_1 || form.uom_3 === form.uom_2)) {
      form.uom_3 = ''
      form.ratio_3 = 0
    }

    syncUomState()
  }

  // Tiap field COA disaring ke akun yang relevan (per klasifikasi/tipe + flag kontra).
  const byType = (t: string) => (a: CoaAccountOpt) => a.account_type_name === t
  const byClass = (c: string, contra?: boolean) => (a: CoaAccountOpt) => a.classification_name === c && (contra === undefined || a.is_contra === contra)
  const COA_FIELDS: { key: string; label: string; filter: (a: CoaAccountOpt) => boolean }[] = [
    { key: 'coa_inventory', label: 'Inventory', filter: byType('Persediaan Barang') },
    { key: 'coa_sales', label: 'Sales', filter: byClass('Pendapatan', false) },
    { key: 'coa_sales_return', label: 'Sales Return', filter: byClass('Pendapatan', true) },
    { key: 'coa_sales_discount', label: 'Sales Discount', filter: byClass('Pendapatan', true) },
    { key: 'coa_good_in_transit', label: 'Good in Transit', filter: byType('Persediaan Barang') },
    { key: 'coa_cogs', label: 'COGS', filter: byClass('Beban Pokok Pendapatan') },
    { key: 'coa_purchase_return', label: 'Purchase Return', filter: byType('Persediaan Barang') },
    { key: 'coa_unbilled_goods', label: 'Unbilled Goods', filter: byClass('Liabilitas') },
  ]

  // Akun default untuk produk BARU (per kode akun). Diterapkan saat openForm.
  const COA_DEFAULT_CODES: Record<string, string> = {
    coa_inventory: '13001',       // Persediaan Produk
    coa_sales: '40001',           // Penjualan Produk
    coa_sales_return: '40003',    // Retur Penjualan
    coa_sales_discount: '40004',  // Diskon Penjualan
    coa_good_in_transit: '13002', // Persediaan Terkirim
    coa_cogs: '51001',            // Harga Pokok Penjualan
    coa_purchase_return: '13001', // Persediaan Produk
    coa_unbilled_goods: '22003',  // Hutang Pembelian Belum Ditagih
  }
  function applyCoaDefaults() {
    for (const [key, code] of Object.entries(COA_DEFAULT_CODES)) {
      (form as any)[key] = coaAccounts.value.find(a => a.account_code === code)?.id || ''
    }
  }

  const coaAccounts = ref<CoaAccountOpt[]>([])
  const coaErrors = reactive<Record<string, boolean>>({})
  // Opsi dropdown untuk sebuah field COA: akun ter-saring, label "kode · nama".
  function coaOptionsFor(f: { filter: (a: CoaAccountOpt) => boolean }) {
    return coaAccounts.value.filter(f.filter).map(a => ({ value: a.id, label: `${a.account_code} · ${a.account_name}` }))
  }

  // selling_uom hanya dari uom_1/2/3 yang dipilih
  const sellingUomOpts = computed(() => {
    const ids = [form.uom_1, form.uom_2, form.uom_3].filter(Boolean)
    return uomOpts.value.filter(o => ids.includes(o.id))
  })

  // ── daftar opsi {value,label} untuk komponen SelectSearch ──
  const asOpts = (list: Opt[]) => list.map(o => ({ value: o.id, label: o.name }))
  const brandOptions = computed(() => asOpts(brandOpts.value))
  const categoryOptions = computed(() => asOpts(categoryOpts.value))
  const subcategoryOptions = computed(() => asOpts(subcategoryOpts.value))
  const uom1Options = computed(() => asOpts(uom1Opts.value))
  const uom2Options = computed(() => asOpts(uom2Opts.value))
  const uom3Options = computed(() => asOpts(uom3Opts.value))
  const sellingUomOptions = computed(() => asOpts(sellingUomOpts.value))
  const countryOptions = computed(() => countryOpts.value.map(c => ({ value: c.code, label: `${flagEmoji(c.code)} ${titleCase(c.name)}` })))

  async function onCategoryChange() {
    form.subcategory_id = ''
    errors.category_id = ''
    errors.subcategory_id = ''
    subcategoryOpts.value = form.category_id ? await fetchSubcategoryOptions(form.category_id) : []
  }

  // ── quick-add master (Brand / Category / Subcategory) via modal ──
  const quickAdd = reactive({
    open: false, field: '' as '' | 'brand' | 'category' | 'subcategory',
    name: '', description: '', logo: '', banner_image: '',
    saving: false, uploading: false, error: '',
  })
  const quickAddTitle = computed(() => ({ brand: 'New Brand', category: 'New Category', subcategory: 'New Sub Category' }[quickAdd.field as 'brand'] ?? ''))
  const currentCategoryName = computed(() => categoryOpts.value.find(c => c.id === form.category_id)?.name ?? '')
  const qaImageValue = computed(() => (quickAdd.field === 'brand' ? quickAdd.logo : quickAdd.banner_image))

  function openQuickAdd(field: 'brand' | 'category' | 'subcategory') {
    Object.assign(quickAdd, { field, name: '', description: '', logo: '', banner_image: '', error: '', saving: false, uploading: false, open: true })
  }

  const qaFileInput = ref<HTMLInputElement>()
  function pickQuickAddImage() { qaFileInput.value?.click() }
  async function onQuickAddFile(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
    if (!file) return
    if (!file.type.startsWith('image/')) { toast.add({ title: 'Invalid file', description: 'Please choose an image', color: 'error' }); return }
    if (file.size > 5 * 1024 * 1024) { toast.add({ title: 'File too large', description: 'Max 5 MB', color: 'error' }); return }
    quickAdd.uploading = true
    try {
      const path = await uploadImage(file, quickAdd.field === 'brand' ? 'brand_logo' : 'category_banner')
      if (quickAdd.field === 'brand') quickAdd.logo = path
      else quickAdd.banner_image = path
    } catch (err: any) { toast.add({ title: 'Upload failed', description: err?.data?.error || 'Could not upload', color: 'error' }) }
    finally { quickAdd.uploading = false }
  }

  const byName = (a: Opt, b: Opt) => a.name.localeCompare(b.name)
  async function submitQuickAdd() {
    const name = quickAdd.name.trim()
    quickAdd.error = ''
    if (!name) { quickAdd.error = 'Name is required'; return }
    if (quickAdd.saving) return
    quickAdd.saving = true
    try {
      if (quickAdd.field === 'brand') {
        const opt = await createBrand({ name, description: quickAdd.description, logo: quickAdd.logo })
        brandOpts.value = [...brandOpts.value, opt].sort(byName)
        form.brand_id = opt.id
        errors.brand_id = ''
      } else if (quickAdd.field === 'category') {
        const opt = await createCategory({ name, banner_image: quickAdd.banner_image })
        categoryOpts.value = [...categoryOpts.value, opt].sort(byName)
        form.category_id = opt.id
        form.subcategory_id = ''
        subcategoryOpts.value = []
        errors.category_id = ''
      } else if (quickAdd.field === 'subcategory') {
        const opt = await createSubcategory(name, form.category_id)
        subcategoryOpts.value = [...subcategoryOpts.value, opt].sort(byName)
        form.subcategory_id = opt.id
        errors.subcategory_id = ''
      }
      quickAdd.open = false
    } catch (e: any) {
      quickAdd.error = e?.data?.error || 'Failed to add'
    } finally {
      quickAdd.saving = false
    }
  }

  function resetForm() {
    Object.assign(form, emptyForm())
    ratio2Text.value = ''
    ratio3Text.value = ''
    dimInput.length_cm = dimInput.width_cm = dimInput.height_cm = dimInput.weight_gr = ''
    errors.product_name = errors.brand_id = errors.category_id = errors.subcategory_id = errors.uom_1 = errors.selling_uom = errors.stocking_uom = errors.ratio_2 = errors.ratio_3 = errors.sku = errors.weight_gr = errors.description = ''
    errors.variant_name_1 = errors.variant_values_1 = errors.variant_values_2 = ''
    for (const f of COA_FIELDS) coaErrors[f.key] = false
    productActive.value = true
    variantLocked.values1 = []
    variantLocked.values2 = []
    resetVariants()
    editingId.value = null
    subcategoryOpts.value = []
  }

  watch(() => form.uom_1, () => syncUomState())
  watch(() => form.uom_2, () => syncUomState())
  watch(() => form.uom_3, () => syncUomState())

  async function loadOptions() {
    ;[brandOpts.value, categoryOpts.value, uomOpts.value, coaAccounts.value] = await Promise.all([
      fetchBrandOptions(), fetchCategoryOptions(), fetchUomOptions(), fetchCoaOptions(),
    ])
  }

  async function openForm() {
    resetForm()
    skuMode.value = 'auto'
    loadSkuPreview()
    await loadOptions()
    applyCoaDefaults() // produk baru: akun COA terisi default
    showForm.value = true
  }

  async function openEdit(row: ProductListItem) {
    resetForm()
    skuMode.value = 'manual'
    await loadOptions()
    const p = await getProduct(row.id)
    const v = (p.variants && p.variants[0]) || {}
    Object.assign(form, {
      product_name: p.product_name, product_type: p.product_type, brand_id: p.brand_id, category_id: p.category_id,
      subcategory_id: p.subcategory_id, country_of_origin: p.country_of_origin, description: p.description,
      ingredients: p.ingredients, is_perishable: p.is_perishable,
      uom_1: p.uom_1, uom_2: p.uom_2, ratio_2: p.ratio_2, uom_3: p.uom_3, ratio_3: p.ratio_3, selling_uom: p.selling_uom, stocking_uom: p.stocking_uom,
      coa_inventory: p.coa_inventory, coa_sales: p.coa_sales, coa_sales_return: p.coa_sales_return,
      coa_sales_discount: p.coa_sales_discount, coa_good_in_transit: p.coa_good_in_transit, coa_cogs: p.coa_cogs,
      coa_purchase_return: p.coa_purchase_return, coa_unbilled_goods: p.coa_unbilled_goods,
      variant_id: v.id || '', sku: v.sku || '', barcode: v.barcode || '',
      def_selling_price: v.def_selling_price || 0, def_purchase_price: v.def_purchase_price || 0, cogs_unit: v.cogs_unit || 0,
      length_cm: v.length_cm || 0, width_cm: v.width_cm || 0, height_cm: v.height_cm || 0, weight_gr: v.weight_gr || 0,
      main_image: v.main_image || '', image_1: v.image_1 || '', image_2: v.image_2 || '', image_3: v.image_3 || '',
    })
    // Rekonstruksi matrix varian dari daftar variants (mode variant).
    if (p.product_type === 'variant' && p.variants && p.variants.length) {
      form.variant_name_1 = p.variant_name_1 || ''
      form.variant_name_2 = p.variant_name_2 || ''
      const v1: string[] = []
      const v2: string[] = []
      for (const v of p.variants) {
        if (v.variant_value_1 && !v1.includes(v.variant_value_1)) v1.push(v.variant_value_1)
        if (v.variant_value_2 && !v2.includes(v.variant_value_2)) v2.push(v.variant_value_2)
      }
      variantAxes.values1 = v1
      variantAxes.values2 = v2
      // Sumbu/value dari DB dikunci (tak bisa diubah/hapus, hanya boleh tambah value baru).
      variantLocked.values1 = [...v1]
      variantLocked.values2 = [...v2]
      const parentActive = p.is_active !== false
      await nextTick() // biarkan watcher regen dulu, lalu timpa dengan data DB
      variantRows.value = p.variants.map((v: any) => ({
        id: v.id || '', variant_value_1: v.variant_value_1 || '', variant_value_2: v.variant_value_2 || '',
        sku: v.sku || '', sku_auto: false, barcode: v.barcode || '',
        def_selling_price: v.def_selling_price || 0, def_purchase_price: v.def_purchase_price || 0, cogs_unit: v.cogs_unit || 0,
        length_cm: v.length_cm || 0, width_cm: v.width_cm || 0, height_cm: v.height_cm || 0, weight_gr: v.weight_gr || 0,
        main_image: v.main_image || '', variant_image: v.variant_image || v.main_image || '',
        image_1: v.image_1 || '', image_2: v.image_2 || '', image_3: v.image_3 || '',
        // Induk inactive → semua varian dipaksa inactive.
        stock_qty: v.stock_qty || 0, is_active: parentActive ? (v.is_active !== false) : false,
      }))
    }
    productActive.value = p.is_active !== false
    subcategoryOpts.value = p.category_id ? await fetchSubcategoryOptions(p.category_id) : []
    syncUomState()
    syncRatioText()
    editingId.value = row.id
    showForm.value = true
  }

  // ── image upload (4 slot) ──
  const imgInput = ref<HTMLInputElement>()
  const currentImageField = ref<'main_image' | 'image_1' | 'image_2' | 'image_3'>('main_image')
  const uploadingField = ref('')
  function pickImage(field: 'main_image' | 'image_1' | 'image_2' | 'image_3') {
    currentImageField.value = field
    imgInput.value?.click()
  }
  async function onImageChange(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
    if (!file) return
    if (!file.type.startsWith('image/')) { toast.add({ title: 'Invalid file', description: 'Please choose an image', color: 'error' }); return }
    if (file.size > 5 * 1024 * 1024) { toast.add({ title: 'File too large', description: 'Max 5 MB', color: 'error' }); return }
    const field = currentImageField.value
    uploadingField.value = field
    try { (form as any)[field] = await uploadImage(file) }
    catch (err: any) { toast.add({ title: 'Upload failed', description: err?.data?.error || 'Could not upload', color: 'error' }) }
    finally { uploadingField.value = '' }
  }

  // true bila HTML punya teks nyata (abaikan tag & &nbsp;) — untuk validasi editor.
  function htmlHasText(html: string) {
    return !!(html || '').replace(/<[^>]*>/g, '').replace(/&nbsp;/g, '').trim()
  }

  function validate() {
    const isVariant = form.product_type === 'variant'
    errors.product_name = form.product_name.trim() ? '' : 'required'
    errors.brand_id = form.brand_id ? '' : 'required'
    errors.category_id = form.category_id ? '' : 'required'
    errors.subcategory_id = form.subcategory_id ? '' : 'required'
    errors.uom_1 = form.uom_1 ? '' : 'required'
    errors.selling_uom = form.uom_1 ? (form.selling_uom ? '' : 'required') : ''
    errors.stocking_uom = form.uom_1 ? (form.stocking_uom ? '' : 'required') : ''
    errors.description = htmlHasText(form.description) ? '' : 'required'
    // Field khusus single (SKU/weight produk) hanya berlaku di mode single.
    errors.sku = !isVariant && skuMode.value === 'manual' ? (form.sku.trim() ? '' : 'required') : ''
    errors.weight_gr = !isVariant ? (Number(form.weight_gr) > 0 ? '' : 'required') : ''
    // Field khusus variant.
    errors.variant_name_1 = isVariant ? (form.variant_name_1.trim() ? '' : 'required') : ''
    errors.variant_values_1 = isVariant ? (variantAxes.values1.filter(Boolean).length ? '' : 'required') : ''
    // Chart of Accounts — ke-8 akun wajib diisi.
    let coaOk = true
    for (const f of COA_FIELDS) {
      const missing = !(form as any)[f.key]
      coaErrors[f.key] = missing
      if (missing) coaOk = false
    }
    if (!coaOk) {
      toast.add({ title: 'Chart of Accounts required', description: 'Please select all 8 accounts in the Chart of Accounts section.', color: 'error' })
    }
    const sharedOk = !errors.product_name && !errors.brand_id && !errors.category_id && !errors.subcategory_id && !errors.uom_1 && !errors.selling_uom && !errors.stocking_uom && !ratio2Err.value && !ratio3Err.value && !errors.description && coaOk
    if (isVariant) {
      const rowsOk = variantListValid.value
      if (!rowsOk && sharedOk && !errors.variant_name_1 && !errors.variant_values_1) {
        toast.add({ title: 'Check variants', description: 'Every variant row needs a SKU' + (applyToAll.weight_gr ? ' and a weight greater than 0' : '') + '.', color: 'error' })
      }
      return sharedOk && !errors.variant_name_1 && !errors.variant_values_1 && rowsOk
    }
    return sharedOk && !errors.sku && !errors.weight_gr
  }

  async function submit() {
    if (!validate()) return
    saving.value = true
    try {
      const isVariant = form.product_type === 'variant'
      // Dicentang → 1 nilai dari form untuk semua varian; uncheck → nilai per baris.
      const sell = (r: any) => applyToAll.def_selling_price ? Number(form.def_selling_price) : Number(r.def_selling_price)
      const buy = (r: any) => applyToAll.def_purchase_price ? Number(form.def_purchase_price) : Number(r.def_purchase_price)
      const wt = (r: any) => applyToAll.weight_gr ? Number(form.weight_gr) : Number(r.weight_gr)
      const variants = isVariant
        ? variantRows.value.map(r => ({
            id: r.id, sku: r.sku_auto ? '' : r.sku, sku_auto: r.sku_auto, barcode: r.barcode,
            variant_value_1: r.variant_value_1, variant_value_2: r.variant_value_2,
            def_selling_price: sell(r), def_purchase_price: buy(r),
            // Dimensi selalu dari form (dipakai semua varian); weight ikut aturan checkbox.
            cogs_unit: 0, length_cm: Number(form.length_cm), width_cm: Number(form.width_cm),
            height_cm: Number(form.height_cm), weight_gr: wt(r),
            main_image: form.main_image, variant_image: r.variant_image || form.main_image,
            image_1: form.image_1, image_2: form.image_2, image_3: form.image_3,
            is_active: r.is_active,
          }))
        : [{
            id: form.variant_id || '', sku: skuMode.value === 'auto' ? '' : form.sku, sku_auto: skuMode.value === 'auto', barcode: form.barcode,
            variant_value_1: '', variant_value_2: '',
            def_selling_price: Number(form.def_selling_price), def_purchase_price: Number(form.def_purchase_price),
            cogs_unit: Number(form.cogs_unit), length_cm: Number(form.length_cm), width_cm: Number(form.width_cm),
            height_cm: Number(form.height_cm), weight_gr: Number(form.weight_gr),
            main_image: form.main_image, variant_image: '', image_1: form.image_1, image_2: form.image_2, image_3: form.image_3,
            is_active: true,
          }]
      const body = {
        product_name: form.product_name, product_type: form.product_type, brand_id: form.brand_id,
        subcategory_id: form.subcategory_id, country_of_origin: form.country_of_origin,
        description: form.description, ingredients: form.ingredients, is_perishable: form.is_perishable,
        uom_1: form.uom_1, uom_2: form.uom_2, ratio_2: Number(form.ratio_2), uom_3: form.uom_3, ratio_3: Number(form.ratio_3),
        selling_uom: form.selling_uom, stocking_uom: form.stocking_uom,
        variant_name_1: isVariant ? form.variant_name_1 : '', variant_name_2: isVariant ? form.variant_name_2 : '',
        coa_inventory: form.coa_inventory, coa_sales: form.coa_sales, coa_sales_return: form.coa_sales_return,
        coa_sales_discount: form.coa_sales_discount, coa_good_in_transit: form.coa_good_in_transit, coa_cogs: form.coa_cogs,
        coa_purchase_return: form.coa_purchase_return, coa_unbilled_goods: form.coa_unbilled_goods,
        variants,
      }
      if (editingId.value) await updateProduct(editingId.value, body)
      else await createProduct(body)
      resetForm()
      showForm.value = false
      await reload()
    } catch (e: any) {
      const field = e?.data?.field
      const msg = e?.data?.error
      if (field && field in errors) (errors as Record<string, string>)[field] = msg || 'Invalid'
      else toast.add({ title: 'Failed', description: msg || 'Could not save product', color: 'error' })
    } finally {
      saving.value = false
    }
  }

  onMounted(async () => {
    fetchCountryOptions().then(list => { countryOpts.value = list }).catch(() => {})
    // Opsi untuk dropdown filter (brand & category).
    fetchBrandOptions().then(list => { brandOpts.value = list }).catch(() => {})
    fetchCategoryOptions().then(list => { categoryOpts.value = list }).catch(() => {})
    await reload()
  })
</script>

<template>
  <div class="page">
    <div class="page-body">
      <div class="page-header">
        <h1 class="page-title">Products</h1>
        <p class="breadcrumbs"><span>Product Management</span> <span class="crumb-sep">›</span> <span>Products</span></p>
      </div>

      <div class="toolbar">
        <div class="toolbar-left">
          <SearchSort
            v-model="search"
            v-model:sort="sortField"
            v-model:desc="sortDesc"
            :sort-options="sortOptions"
            placeholder="Search product name or SKU..."
          />
          <AppFilter :active-count="activeFilterCount" @reset="resetFilters">
            <div>
              <label class="form-label">Product Type</label>
              <SelectSearch v-model="filters.product_type" :options="productTypeFilterOptions" placeholder="All types" />
            </div>
            <div>
              <label class="form-label">Status</label>
              <SelectSearch v-model="filters.status" :options="statusFilterOptions" placeholder="All status" />
            </div>
            <div>
              <label class="form-label">Brand</label>
              <SelectSearch v-model="filters.brand_id" :options="fBrandOptions" placeholder="All brands" />
            </div>
            <div>
              <label class="form-label">Country</label>
              <SelectSearch v-model="filters.country" :options="fCountryOptions" placeholder="All countries" />
            </div>
            <div>
              <label class="form-label">Category</label>
              <SelectSearch v-model="filters.category_id" :options="fCategoryOptions" placeholder="All categories" />
            </div>
            <div>
              <label class="form-label">Subcategory</label>
              <SelectSearch v-model="filters.subcategory_id" :options="filterSubcatOptions" :disabled="!filters.category_id" :placeholder="filters.category_id ? 'All subcategories' : 'Pick a category first'" />
            </div>
          </AppFilter>
        </div>
        <button class="btn-primary" @click="openForm">+ Add New</button>
      </div>

      <!-- ── Form ── -->
      <AppModal v-model="showForm" :title="editingId ? 'Edit Product' : 'New Product'" :hide-close="true" max-width="min(1150px, 96vw)">
        <form class="pform" @submit.prevent="submit">
          <input ref="imgInput" type="file" accept="image/*" hidden @change="onImageChange">
          <input ref="vImgInput" type="file" accept="image/*" hidden @change="onVariantImageChange">
          <div class="pform-body">
            <!-- ========== STEP: Form (Single & Variant setup) ========== -->
            <template v-if="!(form.product_type === 'variant' && variantStep === 'list')">
            <!-- Single / Variant -->
            <div class="type-toggle" :class="{ 'type-locked': !!editingId }">
              <label class="type-opt" :class="{ active: form.product_type === 'single', disabled: !!editingId }">
                <input v-model="form.product_type" type="radio" value="single" :disabled="!!editingId" @change="onSelectType"><span>Single</span>
              </label>
              <label class="type-opt" :class="{ active: form.product_type === 'variant', disabled: !!editingId }">
                <input v-model="form.product_type" type="radio" value="variant" :disabled="!!editingId" @change="onSelectType"><span>Variant</span>
              </label>
            </div>

            <div class="pform-cols">
              <!-- LEFT -->
              <div class="pform-col">
                <div class="gi-img-row">
                <div class="pcard">
                  <div class="pcard-title">General Information</div>
                  <div class="pcard-body gen-grid">
                    <div class="c3">
                      <div class="sku-head">
                        <label class="form-label">SKU <span v-if="form.product_type === 'single'" class="req">*</span>
                          <span v-if="errors.sku === 'required'" class="label-required">Required</span></label>
                        <div v-if="!editingId" class="sku-toggle">
                          <button type="button" class="tip" :class="{ active: skuMode === 'auto' }" data-tip="The system creates the SKU for you (e.g. PSKU-202607-000001). It is always unique and numbered per month." @click="setSkuMode('auto')">Auto</button>
                          <button type="button" class="tip" :class="{ active: skuMode === 'manual' }" data-tip="Type your own SKU code (e.g. MILO-500). It must be unique — no two products can share the same SKU." @click="setSkuMode('manual')">Manual</button>
                        </div>
                      </div>
                      <input v-if="form.product_type === 'variant'" class="text-input sku-locked" value="" placeholder="[autoset per variant]" disabled>
                      <input v-else-if="skuMode === 'manual'" v-model="form.sku" class="text-input" :class="{ 'input-error': errors.sku }" placeholder="e.g. MILO-500" :disabled="!!editingId" @input="errors.sku = ''">
                      <input v-else class="text-input sku-locked" :value="skuPreview || 'Generating…'" disabled title="Dibuat otomatis oleh sistem">
                      <div v-if="errors.sku && errors.sku !== 'required'" class="field-tip">{{ errors.sku }}</div>
                    </div>
                    <div class="c3">
                      <label class="form-label">Barcode</label>
                      <input v-model="form.barcode" class="text-input" :placeholder="form.product_type === 'variant' ? '[set per variant]' : 'Optional'" :disabled="form.product_type === 'variant'">
                    </div>

                    <div class="c6">
                      <div class="pn-head">
                        <label class="form-label">Product Name <span class="req">*</span>
                          <span v-if="errors.product_name === 'required'" class="label-required">Required</span></label>
                        <label class="checkbox-row tip tip--right" :class="{ 'row-locked': !!editingId }" data-tip="Turn this on if the product has an expiry date (e.g. food or drinks). Stock will then be tracked by expiry date."><input v-model="form.is_perishable" type="checkbox" :disabled="!!editingId"> Perishable (has expiry)</label>
                      </div>
                      <input v-model="form.product_name" class="text-input" :class="{ 'input-error': errors.product_name }" placeholder="Example: Saus Tomat 600ml" :disabled="!!editingId" @input="errors.product_name = ''">
                    </div>

                    <div class="c3">
                      <div class="field-head">
                        <label class="form-label">Brand <span class="req">*</span>
                          <span v-if="errors.brand_id === 'required'" class="label-required">Required</span></label>
                        <button type="button" class="label-add" title="Add new brand" @click="openQuickAdd('brand')"><UIcon name="i-lucide-plus" /></button>
                      </div>
                      <SelectSearch v-model="form.brand_id" :options="brandOptions" placeholder="Select Brand" :invalid="!!errors.brand_id" @change="errors.brand_id = ''" />
                    </div>
                    <div class="c3">
                      <label class="form-label">Country of Origin</label>
                      <SelectSearch v-model="form.country_of_origin" :options="countryOptions" placeholder="Select Country" :disabled="!!editingId && !!form.country_of_origin" />
                    </div>

                    <div class="c3">
                      <div class="field-head">
                        <label class="form-label">Category <span class="req">*</span>
                          <span v-if="errors.category_id === 'required'" class="label-required">Required</span></label>
                        <button type="button" class="label-add" title="Add new category" @click="openQuickAdd('category')"><UIcon name="i-lucide-plus" /></button>
                      </div>
                      <SelectSearch v-model="form.category_id" :options="categoryOptions" placeholder="Select Category" :invalid="!!errors.category_id" @change="onCategoryChange" />
                    </div>
                    <div class="c3">
                      <div class="field-head">
                        <label class="form-label">Sub Category <span class="req">*</span>
                          <span v-if="errors.subcategory_id === 'required'" class="label-required">Required</span></label>
                        <button v-if="form.category_id" type="button" class="label-add" title="Add new sub category" @click="openQuickAdd('subcategory')"><UIcon name="i-lucide-plus" /></button>
                      </div>
                      <SelectSearch v-model="form.subcategory_id" :options="subcategoryOptions" :placeholder="form.category_id ? 'Select Sub Category' : 'Pick a category first'" :disabled="!form.category_id" :invalid="!!errors.subcategory_id" @change="errors.subcategory_id = ''" />
                    </div>

                    <div class="c6">
                      <label class="form-label">Description <span class="req">*</span>
                        <span v-if="errors.description === 'required'" class="label-required">Required</span></label>
                      <RichEditor v-model="form.description" :invalid="!!errors.description" placeholder="Write a short description of the product…" @update:model-value="errors.description = ''" />
                    </div>

                    <div class="c6">
                      <label class="form-label">Ingredients</label>
                      <RichEditor v-model="form.ingredients" placeholder="List the product ingredients…" />
                    </div>
                  </div>
                </div>

                <div class="gi-right">
                <div class="pcard img-card">
                  <div class="pcard-title">Images <span class="section-note">PNG, JPEG, JPG, WEBP</span> <span class="img-max">max 5 MB</span>
                    <span v-if="applyOff('main_image')" class="apply-off-note">disabled for variants</span></div>
                  <div class="pcard-body">
                    <!-- Main Image: 1 kotak besar -->
                    <div class="img-slot-wrap img-main" :class="{ 'apply-disabled': applyOff('main_image') }">
                      <span class="img-cap">Main Image<span v-if="!applyOff('main_image')" class="req"> *</span></span>
                      <div class="img-slot" @click="!applyOff('main_image') && pickImage('main_image')">
                        <img v-if="form.main_image" :src="`${API_BASE}/files/${form.main_image}`" alt="">
                        <UIcon v-else-if="uploadingField === 'main_image'" name="i-lucide-loader-circle" class="img-ic spin" />
                        <template v-else><UIcon name="i-lucide-upload" class="img-ic" /><span class="img-hint">Upload Image</span></template>
                      </div>
                      <button v-if="form.main_image && !applyOff('main_image')" type="button" class="img-x" @click="form.main_image = ''"><UIcon name="i-lucide-x" /></button>
                    </div>
                    <!-- Image 1-3: hanya di mode single (bukan bagian apply-to-all) -->
                    <div class="img-thumbs" :class="{ 'apply-disabled': form.product_type === 'variant' }">
                      <div v-for="f in (['image_1','image_2','image_3'] as const)" :key="f" class="img-slot-wrap">
                        <span class="img-cap">{{ f.replace('image_', 'Image ') }}</span>
                        <div class="img-slot" @click="form.product_type === 'single' && pickImage(f)">
                          <img v-if="(form as any)[f]" :src="`${API_BASE}/files/${(form as any)[f]}`" alt="">
                          <UIcon v-else-if="uploadingField === f" name="i-lucide-loader-circle" class="img-ic spin" />
                          <template v-else><UIcon name="i-lucide-upload" class="img-ic" /><span class="img-hint">Upload Image</span></template>
                        </div>
                        <button v-if="(form as any)[f] && form.product_type === 'single'" type="button" class="img-x" @click="(form as any)[f] = ''"><UIcon name="i-lucide-x" /></button>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="pcard oi-card">
                  <div class="pcard-title">Stock <span class="section-note">(All Warehouse)</span></div>
                  <div class="pcard-body oi-body">
                    <div class="oi-field">
                      <label class="form-label">Stocking UoM <span class="req">*</span>
                        <span v-if="errors.stocking_uom === 'required'" class="label-required">Required</span></label>
                      <SelectSearch v-model="form.stocking_uom" :options="sellingUomOptions" placeholder="Select Stocking UoM" :disabled="sellingUomDisabled || (!!editingId && !!form.stocking_uom)" :invalid="!!errors.stocking_uom" @change="errors.stocking_uom = ''" />
                      <span class="field-hint">displayed on the inventory report</span>
                    </div>
                    <div class="oi-line"><span>Total Qty:</span><strong>{{ displayTotalQty.toLocaleString('id-ID') }}<span v-if="stockingUomName" class="oi-unit">{{ stockingUomName }}</span></strong></div>
                    <div class="oi-line"><span>COGS/unit:</span><strong>{{ formatPrice(displayCogsUnit) }}<span v-if="stockingUomName" class="oi-unit">/ {{ stockingUomName }}</span></strong></div>
                    <div class="oi-divider"></div>
                    <div class="oi-line"><span>Total COGS:</span><strong>{{ formatPrice(displayTotalCogs) }}</strong></div>
                  </div>
                </div>
                </div>
                </div>

                <div class="uom-price-row">
                  <div class="pcard">
                    <div class="pcard-title">Unit of Measurement</div>
                    <div class="pcard-body uom-stack">
                      <div class="uom-ratio">
                        <div>
                          <label class="form-label">UoM 1 <span class="req">*</span>
                            <span v-if="errors.uom_1 === 'required'" class="label-required">Required</span></label>
                          <SelectSearch v-model="form.uom_1" :options="uom1Options" placeholder="Select UoM" :disabled="uom1Disabled || (!!editingId && !!form.uom_1)" :invalid="!!errors.uom_1" @change="syncUomSelections('uom_1')" />
                        </div>
                        <div class="uom1-info-cell">
                          <div class="uom1-info">The basic unit used for the product (e.g. pcs, kg, box, pack, or liter).</div>
                        </div>
                      </div>
                      <div class="uom-ratio">
                        <div>
                          <label class="form-label">UoM 2</label>
                          <SelectSearch v-model="form.uom_2" :options="uom2Options" placeholder="Select UoM" :disabled="uom2Disabled || (!!editingId && !!form.uom_2)" @change="syncUomSelections('uom_2')" />
                        </div>
                        <div>
                          <label class="form-label">Ratio 2 <span v-if="form.uom_2" class="req">*</span>
                            <span v-if="ratio2Err === 'required'" class="label-required">Required</span></label>
                          <div class="ratio-row">
                            <input :value="ratio2Text" type="text" inputmode="decimal" autocomplete="off" placeholder="0" class="text-input" :class="{ 'input-error': ratio2Err }" :disabled="ratio2Disabled || (!!editingId && Number(form.ratio_2) > 0)" @focus="onRatioFocus($event, 'ratio_2')" @input="onRatioInput($event, 'ratio_2')" @blur="onRatioBlur($event, 'ratio_2')">
                            <span class="ratio-unit">{{ uom1Name }}</span>
                          </div>
                        </div>
                      </div>
                      <div class="uom-ratio">
                        <div>
                          <label class="form-label">UoM 3</label>
                          <SelectSearch v-model="form.uom_3" :options="uom3Options" placeholder="Select UoM" :disabled="uom3Disabled || (!!editingId && !!form.uom_3)" @change="syncUomSelections('uom_3')" />
                        </div>
                        <div>
                          <label class="form-label">Ratio 3 <span v-if="form.uom_3" class="req">*</span>
                            <span v-if="ratio3Err === 'required'" class="label-required">Required</span></label>
                          <div class="ratio-row">
                            <input :value="ratio3Text" type="text" inputmode="decimal" autocomplete="off" placeholder="0" class="text-input" :class="{ 'input-error': ratio3Err }" :disabled="ratio3Disabled || (!!editingId && Number(form.ratio_3) > 0)" @focus="onRatioFocus($event, 'ratio_3')" @input="onRatioInput($event, 'ratio_3')" @blur="onRatioBlur($event, 'ratio_3')">
                            <span class="ratio-unit">{{ uom1Name }}</span>
                          </div>
                          <div v-if="ratio3Err === 'must be greater than ratio 2'" class="field-tip">Must be greater than Ratio 2</div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="pcard">
                    <div class="pcard-title">Price</div>
                    <div class="pcard-body">
                      <div>
                        <label class="form-label">Selling Uom <span class="req">*</span>
                          <span v-if="errors.selling_uom === 'required'" class="label-required">Required</span></label>
                        <SelectSearch v-model="form.selling_uom" :options="sellingUomOptions" placeholder="Select Selling UoM" :disabled="sellingUomDisabled || (!!editingId && !!form.selling_uom)" :invalid="!!errors.selling_uom" @change="errors.selling_uom = ''" />
                      </div>
                      <div>
                        <div class="field-head-check">
                          <label class="form-label">Def. Selling Price <span v-if="priceRequired('def_selling_price')" class="req">*</span></label>
                          <label v-if="form.product_type === 'variant'" class="apply-check" :class="{ on: applyToAll.def_selling_price }" title="Checked: satu nilai untuk semua varian (dari form ini). Unchecked: isi per varian di Variant List."><input v-model="applyToAll.def_selling_price" type="checkbox" @change="onApplyToggle('def_selling_price')"> same for all</label>
                        </div>
                        <input :value="priceRequired('def_selling_price') ? priceDisplayMin0(form.def_selling_price) : priceDisplay(form.def_selling_price)" class="text-input" :class="{ 'input-error': priceRequired('def_selling_price') && Number(form.def_selling_price) <= 0 }" inputmode="numeric" placeholder="Rp 0" :disabled="applyOff('def_selling_price')" @input="priceRequired('def_selling_price') ? onPriceInputMin0($event, 'def_selling_price') : onPriceInput($event, 'def_selling_price')">
                      </div>
                      <div>
                        <div class="field-head-check">
                          <label class="form-label">Def. Purchase Price <span v-if="priceRequired('def_purchase_price')" class="req">*</span></label>
                          <label v-if="form.product_type === 'variant'" class="apply-check" :class="{ on: applyToAll.def_purchase_price }" title="Checked: satu nilai untuk semua varian (dari form ini). Unchecked: isi per varian di Variant List."><input v-model="applyToAll.def_purchase_price" type="checkbox" @change="onApplyToggle('def_purchase_price')"> same for all</label>
                        </div>
                        <input :value="priceRequired('def_purchase_price') ? priceDisplayMin0(form.def_purchase_price) : priceDisplay(form.def_purchase_price)" class="text-input" :class="{ 'input-error': priceRequired('def_purchase_price') && Number(form.def_purchase_price) <= 0 }" inputmode="numeric" placeholder="Rp 0" :disabled="applyOff('def_purchase_price')" @input="priceRequired('def_purchase_price') ? onPriceInputMin0($event, 'def_purchase_price') : onPriceInput($event, 'def_purchase_price')">
                      </div>
                    </div>
                  </div>

                  <div class="pcard">
                    <div class="pcard-title">Dimension &amp; Weight</div>
                    <div class="pcard-body dim-grid">
                      <div class="dim-col">
                        <div><label class="form-label">Length (cm) <span v-if="applyOff('length_cm')" class="apply-off-note">off</span></label><input :value="dimInput.length_cm" type="text" inputmode="decimal" autocomplete="off" placeholder="0" class="text-input" :disabled="applyOff('length_cm')" @focus="onDimFocus($event, 'length_cm')" @input="onDimInput($event, 'length_cm')" @blur="onDimBlur($event, 'length_cm')"></div>
                        <div><label class="form-label">Width (cm) <span v-if="applyOff('width_cm')" class="apply-off-note">off</span></label><input :value="dimInput.width_cm" type="text" inputmode="decimal" autocomplete="off" placeholder="0" class="text-input" :disabled="applyOff('width_cm')" @focus="onDimFocus($event, 'width_cm')" @input="onDimInput($event, 'width_cm')" @blur="onDimBlur($event, 'width_cm')"></div>
                        <div><label class="form-label">Height (cm) <span v-if="applyOff('height_cm')" class="apply-off-note">off</span></label><input :value="dimInput.height_cm" type="text" inputmode="decimal" autocomplete="off" placeholder="0" class="text-input" :disabled="applyOff('height_cm')" @focus="onDimFocus($event, 'height_cm')" @input="onDimInput($event, 'height_cm')" @blur="onDimBlur($event, 'height_cm')"></div>
                      </div>
                      <div class="dim-col">
                        <div>
                          <div class="field-head-check">
                            <label class="form-label">Weight (gr) <span v-if="!applyOff('weight_gr')" class="req">*</span>
                              <span v-if="errors.weight_gr === 'required'" class="label-required">Required</span></label>
                            <label v-if="form.product_type === 'variant'" class="apply-check" :class="{ on: applyToAll.weight_gr }" title="Checked: satu nilai untuk semua varian (dari form ini). Unchecked: isi per varian di Variant List (tidak wajib)."><input v-model="applyToAll.weight_gr" type="checkbox" @change="onApplyToggle('weight_gr')"> same for all</label>
                          </div>
                          <input :value="dimInput.weight_gr" type="text" inputmode="decimal" autocomplete="off" placeholder="0" class="text-input" :class="{ 'input-error': errors.weight_gr }" :disabled="applyOff('weight_gr')" @focus="onDimFocus($event, 'weight_gr')" @input="onDimInput($event, 'weight_gr')" @blur="onDimBlur($event, 'weight_gr')">
                          <span class="field-hint">Weight (gr) per Selling UoM</span>
                        </div>

                        <!-- ilustrasi dimensi (nilai live L/W/H/Weight) -->
                        <div class="dim-illus">
                          <svg class="dim-svg" viewBox="0 0 196 145" xmlns="http://www.w3.org/2000/svg">
                            <defs>
                              <marker id="dimHead" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse">
                                <path d="M0,0 L10,5 L0,10 z" class="dim-head" />
                              </marker>
                            </defs>
                            <text v-if="dimText.wt" x="93" y="14" text-anchor="middle" class="dim-cap">Weight {{ dimText.wt }} g</text>
                            <!-- box -->
                            <polygon class="box-face" points="50,55 112,55 137,37 75,37" />
                            <polygon class="box-face" points="112,55 137,37 137,85 112,103" />
                            <rect class="box-face" x="50" y="55" width="62" height="48" />
                            <!-- Height -->
                            <line class="dim-arrow" x1="42" y1="55" x2="42" y2="103" marker-start="url(#dimHead)" marker-end="url(#dimHead)" />
                            <text x="37" y="79" text-anchor="end"><tspan class="dim-lbl">H</tspan> <tspan class="dim-val">{{ dimText.h || '–' }}</tspan></text>
                            <text v-if="dimText.h" x="37" y="89" text-anchor="end" class="dim-unit">cm</text>
                            <!-- Width -->
                            <line class="dim-arrow" x1="50" y1="115" x2="112" y2="115" marker-start="url(#dimHead)" marker-end="url(#dimHead)" />
                            <text x="81" y="127" text-anchor="middle"><tspan class="dim-lbl">W</tspan> <tspan class="dim-val">{{ dimText.w || '–' }}</tspan></text>
                            <text v-if="dimText.w" x="81" y="137" text-anchor="middle" class="dim-unit">cm</text>
                            <!-- Length -->
                            <line class="dim-arrow" x1="118" y1="108" x2="143" y2="90" marker-start="url(#dimHead)" marker-end="url(#dimHead)" />
                            <text x="146" y="103"><tspan class="dim-lbl">L</tspan> <tspan class="dim-val">{{ dimText.l || '–' }}</tspan></text>
                            <text v-if="dimText.l" x="146" y="113" class="dim-unit">cm</text>
                          </svg>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="pcard">
              <div class="pcard-title">Chart of Accounts <span class="section-note">Accounts that can be selected based on pre-defined accounts</span></div>
              <div class="pcard-body coa-grid">
                <div v-for="f in COA_FIELDS" :key="f.key">
                  <label class="form-label">
                    {{ f.label }} <span class="req">*</span>
                    <span v-if="coaErrors[f.key]" class="label-required">Required</span>
                  </label>
                  <SelectSearch
                    v-model="(form as any)[f.key]"
                    :options="coaOptionsFor(f)"
                    placeholder="Select account"
                    :invalid="coaErrors[f.key]"
                    @change="coaErrors[f.key] = false"
                  />
                </div>
              </div>
            </div>
            </template>

            <!-- ========== STEP: Variant Options + Variant List ========== -->
            <template v-else>
              <!-- Variant Options (sumbu) di atas Variant List -->
              <div class="pcard">
                <div class="pcard-title">Variant Options <span class="section-note">Tentukan variant apa saja — kombinasinya otomatis jadi baris di bawah</span></div>
                <div class="pcard-body var-axes">
                  <div class="var-axis">
                    <div class="var-axis-head">
                      <label class="form-label">Variant Name 1 <span class="req">*</span></label>
                    </div>
                    <input v-model="form.variant_name_1" class="text-input" :disabled="nameLocked(1)" placeholder="e.g. Color, Size">
                    <label class="form-label var-vals-lbl">Variant Values 1 <span class="req">*</span></label>
                    <div class="chip-input">
                      <span v-for="(val, i) in variantAxes.values1" :key="`v1-${i}`" class="chip" :class="{ locked: isLockedValue(1, val) }">{{ val }}<button v-if="!isLockedValue(1, val)" type="button" class="chip-x" @click="removeAxisValue(1, i)"><UIcon name="i-lucide-x" /></button></span>
                      <input v-model="axisInput1" class="chip-field" :placeholder="variantAxes.values1.length ? 'Add more…' : 'Type a value, press Enter'" @keydown.enter.prevent="addAxisValue(1)">
                    </div>
                    <span class="field-hint">Press Enter to add each value</span>
                  </div>
                  <div class="var-axis">
                    <div class="var-axis-head">
                      <label class="form-label">Variant Name 2 <span class="section-note">optional</span></label>
                    </div>
                    <input v-model="form.variant_name_2" class="text-input" :disabled="nameLocked(2)" placeholder="e.g. Size (leave empty for 1 axis)">
                    <label class="form-label var-vals-lbl">Variant Values 2 <span v-if="form.variant_name_2.trim()" class="req">*</span></label>
                    <div class="chip-input" :class="{ disabled: !form.variant_name_2.trim() }">
                      <span v-for="(val, i) in variantAxes.values2" :key="`v2-${i}`" class="chip" :class="{ locked: isLockedValue(2, val) }">{{ val }}<button v-if="!isLockedValue(2, val)" type="button" class="chip-x" @click="removeAxisValue(2, i)"><UIcon name="i-lucide-x" /></button></span>
                      <input v-model="axisInput2" class="chip-field" :disabled="!form.variant_name_2.trim()" :placeholder="variantAxes.values2.length ? 'Add more…' : 'Type a value, press Enter'" @keydown.enter.prevent="addAxisValue(2)">
                    </div>
                    <span class="field-hint">Leave empty to keep a single-axis variant</span>
                  </div>
                </div>
              </div>

              <!-- Variant List -->
              <div class="vlist-head">
                <div class="pcard-title vlist-title">Variant List <span class="section-note">{{ variantRows.length }} variant{{ variantRows.length === 1 ? '' : 's' }}</span></div>
                <span class="field-hint">Isi SKU tiap varian. Kolom "same for all" terkunci (nilai dari form); yang tidak, isi per baris.</span>
              </div>
              <div v-if="!variantRows.length" class="var-empty">Tambahkan minimal satu nilai ke <strong>Variant Values 1</strong> untuk membuat baris varian.</div>
              <div v-else class="var-table-wrap">
                <table class="var-table">
                  <thead>
                    <tr>
                      <th class="vt-img">Image</th>
                      <th class="vt-variant">{{ form.variant_name_1 || 'Variant 1' }}<template v-if="form.variant_name_2 && variantAxes.values2.length"> - {{ form.variant_name_2 }}</template></th>
                      <th class="vt-sku">SKU</th>
                      <th class="vt-barcode">Barcode</th>
                      <th class="vt-price">Selling Price</th>
                      <th class="vt-price">Purchase Price</th>
                      <th class="vt-weight">Weight (gr)</th>
                      <th v-if="editingId" class="vt-status">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(row, i) in variantRows" :key="`${row.variant_value_1}||${row.variant_value_2}`">
                      <td class="vt-img">
                        <div class="vt-img-slot" @click="pickVariantImage(i)">
                          <img v-if="row.variant_image" :src="`${API_BASE}/files/${row.variant_image}`" alt="">
                          <UIcon v-else-if="uploadingVariantIdx === i" name="i-lucide-loader-circle" class="img-ic spin" />
                          <UIcon v-else name="i-lucide-upload" class="img-ic" />
                        </div>
                      </td>
                      <td class="vt-variant">
                        <span class="vt-combo">{{ row.variant_value_1 }}<template v-if="row.variant_value_2"> - {{ row.variant_value_2 }}</template></span>
                      </td>
                      <td class="vt-sku">
                        <input v-if="row.sku_auto && !row.id" class="cell-input" value="(auto)" disabled>
                        <input v-else v-model="row.sku" class="cell-input" :class="{ 'input-error': !row.id && !row.sku.trim() }" :disabled="!!row.id" placeholder="SKU">
                      </td>
                      <td class="vt-barcode"><input v-model="row.barcode" class="cell-input" placeholder="Barcode"></td>
                      <td class="vt-price"><input :value="applyToAll.def_selling_price ? priceDisplayMin0(form.def_selling_price) : priceDisplay(row.def_selling_price)" class="cell-input" inputmode="numeric" placeholder="Rp 0" :disabled="applyToAll.def_selling_price" @input="onVarPriceInput($event, i, 'def_selling_price')"></td>
                      <td class="vt-price"><input :value="applyToAll.def_purchase_price ? priceDisplayMin0(form.def_purchase_price) : priceDisplay(row.def_purchase_price)" class="cell-input" inputmode="numeric" placeholder="Rp 0" :disabled="applyToAll.def_purchase_price" @input="onVarPriceInput($event, i, 'def_purchase_price')"></td>
                      <td class="vt-weight"><input :value="applyToAll.weight_gr ? formatDecimalComma(Number(form.weight_gr)) : vNumVal(i, 'weight_gr')" type="text" inputmode="decimal" class="cell-input" :disabled="applyToAll.weight_gr" placeholder="0" @input="onVarNumInput($event, i, 'weight_gr')" @blur="onVarNumBlur(i, 'weight_gr')"></td>
                      <td v-if="editingId" class="vt-status">
                        <button type="button" class="row-switch" :class="{ on: row.is_active }" :disabled="!productActive" :title="productActive ? (row.is_active ? 'Active' : 'Inactive') : 'Produk induk nonaktif — aktifkan dulu produknya'" @click="row.is_active = !row.is_active"><span class="row-switch-knob"></span></button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </template>
          </div>

          <div class="modal-actions">
            <template v-if="form.product_type === 'single'">
              <button type="button" class="btn-ghost" @click="showForm = false">Cancel</button>
              <button class="btn-primary" :disabled="saving" type="submit">{{ saving ? 'Saving...' : 'Save' }}</button>
            </template>
            <template v-else-if="variantStep === 'form'">
              <button type="button" class="btn-ghost" @click="showForm = false">Cancel</button>
              <button type="button" class="btn-primary" :disabled="!variantFormValid" @click="goToVariantList">to Variant List →</button>
            </template>
            <template v-else>
              <button type="button" class="btn-ghost" @click="backToVariantForm">← Back</button>
              <button class="btn-primary" :disabled="saving || !variantListValid" type="submit">{{ saving ? 'Saving...' : 'Save' }}</button>
            </template>
          </div>
        </form>
      </AppModal>

      <!-- ── Quick-add master (Brand / Category / Sub Category) ── -->
      <AppModal v-model="quickAdd.open" :title="quickAddTitle" :hide-close="true">
        <form class="qa-form" @submit.prevent="submitQuickAdd">
          <input ref="qaFileInput" type="file" accept="image/*" hidden @change="onQuickAddFile">

          <div v-if="quickAdd.field === 'subcategory'" class="qa-field">
            <label class="form-label">Category</label>
            <div class="readonly-value">{{ currentCategoryName }}</div>
          </div>

          <div class="qa-field">
            <label class="form-label">Name <span class="req">*</span></label>
            <input v-model="quickAdd.name" class="text-input" :class="{ 'input-error': quickAdd.error }" :placeholder="`Enter ${quickAddTitle.replace('New ', '')} name`" @input="quickAdd.error = ''">
            <div v-if="quickAdd.error" class="field-tip">{{ quickAdd.error }}</div>
          </div>

          <div v-if="quickAdd.field === 'brand'" class="qa-field">
            <label class="form-label">Description</label>
            <textarea v-model="quickAdd.description" class="text-input" rows="2" placeholder="Optional" />
          </div>

          <div v-if="quickAdd.field === 'brand' || quickAdd.field === 'category'" class="qa-field">
            <label class="form-label">{{ quickAdd.field === 'brand' ? 'Logo' : 'Banner image' }}</label>
            <div class="img-slot-wrap">
              <div class="img-slot qa-img-slot" @click="pickQuickAddImage">
                <img v-if="qaImageValue" :src="`${API_BASE}/files/${qaImageValue}`" alt="">
                <UIcon v-else-if="quickAdd.uploading" name="i-lucide-loader-circle" class="img-ic spin" />
                <template v-else><UIcon name="i-lucide-upload" class="img-ic" /><span class="img-hint">Upload Image</span></template>
              </div>
              <button v-if="qaImageValue" type="button" class="img-x" @click="quickAdd.field === 'brand' ? (quickAdd.logo = '') : (quickAdd.banner_image = '')"><UIcon name="i-lucide-x" /></button>
            </div>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn-ghost" @click="quickAdd.open = false">Cancel</button>
            <button class="btn-primary" :disabled="quickAdd.saving || quickAdd.uploading" type="submit">{{ quickAdd.saving ? 'Saving...' : 'Save' }}</button>
          </div>
        </form>
      </AppModal>

      <!-- ── Table ── -->
      <div class="table-card">
        <div ref="scrollEl" class="table-scroll" @scroll="onScroll">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width:56px"></th>
                <th style="min-width:240px">Product</th>
                <th>Brand</th>
                <th style="min-width:160px">Category - Subcategory</th>
                <th style="min-width:130px">Selling Price</th>
                <th style="min-width:130px">COGS Value</th>
                <th class="text-center" style="width:110px">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in items" :key="p.variant_id" class="clickable" @click="openEdit(p)">
                <td class="col-img">
                  <div class="list-img">
                    <img v-if="p.image" :src="`${API_BASE}/files/${p.image}`" alt="">
                    <UIcon v-else name="i-lucide-image" class="list-img-ph" />
                  </div>
                </td>
                <td>
                  <div class="prod-info">
                    <div class="prod-line">
                      <span class="prod-sku">{{ p.sku || '—' }}</span>
                      <span class="type-badge" :class="p.product_type">{{ titleCase(p.product_type) }}</span>
                    </div>
                    <div class="prod-line">
                      <span class="prod-name">{{ variantName(p) }}</span>
                      <span v-if="p.country_of_origin" class="prod-country">{{ flagEmoji(p.country_of_origin) }} {{ p.country_of_origin.toUpperCase() }}</span>
                    </div>
                  </div>
                </td>
                <td>{{ p.brand_name || '—' }}</td>
                <td>{{ p.category_name }}<template v-if="p.subcategory_name"> - {{ p.subcategory_name }}</template></td>
                <td>{{ formatPrice(p.price) }}<span v-if="p.selling_uom_name" class="unit-suffix"> / {{ p.selling_uom_name }}</span></td>
                <td>{{ formatPrice(p.cogs_selling) }}<span v-if="p.selling_uom_name" class="unit-suffix"> / {{ p.selling_uom_name }}</span></td>
                <td class="text-center">
                  <button class="toggle" :class="{ on: p.is_active }" :title="p.is_active ? 'Active' : 'Inactive'" @click.stop="toggleVariantActive(p)">
                    <span class="toggle-knob" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-if="loading && items.length" class="table-loading"><UIcon name="i-lucide-loader-circle" class="spin" /> Loading…</div>
          <EmptyState v-if="!items.length && !loading" text="No products yet" icon="i-lucide-box" />
        </div>
        <TablePager :page="currentPage" :total="total" :page-size="pageSize" readonly />
      </div>
    </div>
  </div>
</template>

<style scoped>
  .page-header { padding-bottom: 16px; border-bottom: 1px solid var(--border-color); margin-bottom: 20px; }
  .breadcrumbs { margin-top: 6px; font-size: 13px; color: var(--text-muted); }

  .pform { display: flex; flex-direction: column; }
  .pform-body { height: 70vh; overflow-y: auto; padding-right: 4px; }

  /* Single / Variant toggle */
  .type-toggle { display: inline-flex; gap: 8px; margin-bottom: 18px; }
  .type-opt {
    display: inline-flex; align-items: center; gap: 8px; padding: 8px 18px; border-radius: 12px;
    border: 1px solid var(--border-color); background: var(--bg-surface); font-size: 14px; font-weight: 700;
    color: var(--text-secondary); cursor: pointer; transition: all 0.15s ease;
  }
  .type-opt input { accent-color: var(--accent); }
  .type-opt.active { border-color: var(--accent); color: var(--accent); background: var(--accent-light); }

  /* 2-column card layout */
  .pform-cols { display: grid; grid-template-columns: 1fr; gap: 18px; align-items: start; }
  .pform-col { display: flex; flex-direction: column; gap: 18px; }
  .pcard { border: 1px solid var(--border-color); border-radius: 14px; padding: 18px; }
  .pform-cols + .pcard, .pcard + .pcard { margin-top: 18px; }
  .pcard-title { font-size: 15px; font-weight: 800; color: var(--text-primary); margin-bottom: 14px; }
  .section-note { font-weight: 600; color: var(--text-muted); font-size: 11px; margin-left: 6px; }
  .pcard-body { display: flex; flex-direction: column; gap: 20px; }
  .pcard-body > div { position: relative; }
  /* beri napas antara label dan input di form produk */
  .pcard .form-label { margin-bottom: 11px; }

  /* UoM & Price: 2 sub-columns */
  .uom-price { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
  .uom-col { display: flex; flex-direction: column; gap: 14px; }
  .uom-stack { display: flex; flex-direction: column; gap: 20px; }
  .uom-stack > div { width: 100%; }
  .uom-ratio { display: grid; grid-template-columns: 1fr 1.2fr; gap: 10px; }
  .uom-ratio > div { position: relative; }
  /* ratio input (lebar tetap kecil) + unit UoM 1 (wrap maks 2 baris lalu ellipsis) */
  .ratio-row { display: flex; align-items: center; gap: 8px; }
  .ratio-row > .text-input { flex: 0 0 92px; width: 92px; }
  .ratio-unit {
    flex: 1 1 auto; min-width: 0;
    color: var(--text-secondary); font-weight: 700; font-size: 13px; line-height: 1.25;
    display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
  }
  /* kotak info kuning soft di samping kanan UoM 1 */
  .uom1-info-cell { display: flex; align-items: flex-end; }
  .uom1-info {
    background: #fef9c3; border: 1px solid #fde68a; color: #854d0e;
    border-radius: 8px; padding: 8px 10px; font-size: 11px; font-weight: 600; line-height: 1.4;
  }
  /* hint statis abu-abu di bawah field */
  .field-hint { display: block; margin-top: 6px; font-size: 12px; color: var(--text-muted); }

  .grid-2 { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 18px 20px; }
  .grid-3 { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 18px 20px; }
  .grid-4 { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 18px 20px; }
  /* COA: 2 kolom x 4 baris, urut per-kolom (kolom 1 penuh dulu, lalu kolom 2) */
  .coa-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); grid-template-rows: repeat(4, auto); grid-auto-flow: column; gap: 18px 24px; }
  .coa-grid > div { position: relative; }

  /* ── Variant options + table ── */
  .price-variant-note { font-size: 12px; color: var(--text-muted); font-style: italic; }
  .apply-off-note { font-size: 11px; font-weight: 700; color: var(--text-muted); background: var(--bg-muted); padding: 1px 6px; border-radius: 6px; margin-left: 4px; text-transform: uppercase; letter-spacing: 0.03em; }
  /* Label + checkbox "same for all" sejajar di ujung kanan */
  .field-head-check { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
  .apply-check { display: inline-flex; align-items: center; gap: 5px; font-size: 11px; font-weight: 700; color: var(--text-muted); cursor: pointer; user-select: none; white-space: nowrap; }
  .apply-check input { width: 14px; height: 14px; accent-color: var(--accent); }
  .apply-check.on { color: var(--accent); }
  .apply-disabled { opacity: 0.5; pointer-events: none; }

  /* Baris toggle Single/Variant + banner (sejajar horizontal) */
  .type-row { display: flex; align-items: stretch; gap: 16px; margin-bottom: 18px; }
  .type-row .type-toggle { margin-bottom: 0; align-items: center; }

  /* Toggle terkunci saat mode edit */
  .type-opt.disabled { cursor: not-allowed; opacity: 0.6; }
  .type-locked .type-opt input { cursor: not-allowed; }

  /* Banner ringkasan varian di form utama (lebar mengikuti isi, didorong ke kanan) */
  .var-banner {
    flex: 0 1 auto; min-width: 0; margin-left: auto;
    display: flex; align-items: center; gap: 14px;
    padding: 8px 12px 8px 14px; border-radius: 12px;
    background: var(--accent-light); border: 1px solid var(--accent);
  }
  .var-banner-info { display: flex; align-items: center; gap: 12px; min-width: 0; }
  .var-banner-ic { color: var(--accent); width: 22px; height: 22px; flex: 0 0 auto; }
  .var-banner-title { font-weight: 700; color: var(--text-primary); font-size: 14px; }
  .var-banner-sub { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
  .var-banner-btn { flex: 0 0 auto; }

  /* Modal Variant Options */
  .vmodal-body { display: flex; flex-direction: column; gap: 16px; }
  .vmodal-pname { margin-bottom: 16px; }
  .row-locked { opacity: 0.6; cursor: not-allowed; }
  .vmodal-body .pcard + .pcard { margin-top: 0; }
  .apply-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px 16px; }
  .apply-item {
    display: flex; align-items: center; gap: 10px; padding: 10px 12px;
    border: 1px solid var(--border-color); border-radius: 10px; cursor: pointer; user-select: none;
    font-size: 14px; color: var(--text-primary); transition: background 0.12s, border-color 0.12s;
  }
  .apply-item.checked { background: var(--accent-light); border-color: var(--accent); font-weight: 600; }
  .apply-item input { width: 16px; height: 16px; flex: 0 0 auto; }
  .apply-req { color: var(--danger, #dc2626); font-weight: 700; }

  /* Header Variant List step */
  .vlist-head { margin-bottom: 14px; }
  .vlist-title { margin: 0 0 2px; }
  .var-axes { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; align-items: start; }
  .var-axis-head { display: flex; align-items: center; justify-content: space-between; }
  .var-vals-lbl { margin-top: 12px; }
  .chip-input {
    display: flex; flex-wrap: wrap; gap: 6px; align-items: center;
    min-height: 40px; padding: 6px 8px; margin-top: 4px;
    border: 1px solid var(--border-color); border-radius: 10px; background: var(--bg-surface);
  }
  .chip-input.input-error { border-color: var(--danger, #dc2626); }
  .chip-input.disabled { background: var(--bg-muted); opacity: 0.7; }
  .chip {
    display: inline-flex; align-items: center; gap: 4px;
    padding: 3px 6px 3px 10px; border-radius: 999px;
    background: var(--accent-light); color: var(--accent); font-size: 13px; font-weight: 600;
  }
  .chip-x { display: inline-flex; padding: 2px; border: none; background: transparent; color: inherit; cursor: pointer; border-radius: 999px; }
  .chip-x:hover { background: rgba(0, 0, 0, 0.1); }
  .chip-field { flex: 1 1 90px; min-width: 90px; border: none; outline: none; background: transparent; font-size: 14px; color: var(--text-primary); font-family: var(--font-family); }
  .var-empty { padding: 24px; text-align: center; color: var(--text-muted); font-size: 14px; }
  .var-table-wrap { overflow-x: auto; }
  .var-table { width: 100%; border-collapse: collapse; font-size: 13px; }
  .var-table th, .var-table td { padding: 7px 8px; border-bottom: 1px solid var(--border-color); text-align: left; vertical-align: middle; }
  .var-table th { font-weight: 700; color: var(--text-muted); font-size: 11px; text-transform: uppercase; letter-spacing: 0.02em; white-space: nowrap; }
  /* Kolom Variant (color-size) = paling lebar */
  .var-table .vt-variant { width: 40%; min-width: 175px; }
  .vt-combo { display: inline-block; padding: 4px 12px; border-radius: 8px; background: var(--accent-light); color: var(--accent); font-weight: 700; font-size: 13px; white-space: nowrap; }
  .vt-img { width: 52px; }
  .vt-img-slot {
    width: 40px; height: 40px; border-radius: 8px; border: 1px dashed var(--border-color);
    display: flex; align-items: center; justify-content: center; cursor: pointer; overflow: hidden; background: var(--bg-muted);
  }
  .vt-img-slot img { width: 100%; height: 100%; object-fit: cover; }
  .vt-img-slot .img-ic { color: var(--text-muted); width: 16px; height: 16px; }
  .cell-input {
    width: 100%; min-width: 0; padding: 6px 8px; border: 1px solid var(--border-color); border-radius: 8px;
    background: var(--bg-surface); font-size: 13px; color: var(--text-primary); font-family: var(--font-family); box-sizing: border-box;
  }
  .cell-input:disabled { background: var(--bg-muted); color: var(--text-muted); }
  .cell-input.input-error { border-color: var(--danger, #dc2626); }
  /* Sel bertumpuk (SKU/Barcode & Sell/Buy Price atas-bawah) */
  .cell-stack { display: flex; flex-direction: column; gap: 4px; }
  /* Chip value yang terkunci (sudah tersimpan di DB, mode edit) */
  .chip.locked { background: var(--bg-muted); color: var(--text-secondary); padding-right: 10px; }
  /* Toggle status aktif/inactive per baris varian */
  .var-table .vt-status { width: 1%; white-space: nowrap; text-align: center; }
  .row-switch {
    position: relative; width: 38px; height: 22px; border-radius: 999px; border: none;
    background: var(--border-color); cursor: pointer; padding: 0; vertical-align: middle; transition: background 0.15s;
  }
  .row-switch.on { background: var(--success, #16a34a); }
  .row-switch:disabled { opacity: 0.5; cursor: not-allowed; }
  .row-switch-knob { position: absolute; top: 2px; left: 2px; width: 18px; height: 18px; border-radius: 50%; background: #fff; transition: transform 0.15s; }
  .row-switch.on .row-switch-knob { transform: translateX(16px); }
  .row-switch-label { margin-left: 8px; font-size: 12px; font-weight: 600; color: var(--text-muted); }
  .var-table .vt-sku { width: 18%; }
  .var-table .vt-barcode { width: 1%; }
  .var-table .vt-sku .cell-input { min-width: 150px; }
  .var-table .vt-barcode .cell-input { min-width: 119px; width: 119px; }
  .var-table .vt-price { width: 1%; }
  .var-table th.vt-price { white-space: normal; line-height: 1.2; } /* header boleh wrap → kolom bisa sempit */
  .var-table .vt-price .cell-input { min-width: 100px; width: 108px; }
  .var-table .vt-weight { width: 1%; }
  .var-table .vt-weight .cell-input { min-width: 60px; width: 72px; }
  .grid-2 > div, .grid-3 > div, .grid-4 > div { position: relative; }
  /* General Information: grid 6 kolom + span (c2/c3/c4) */
  .gen-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 18px 20px; align-items: start; }
  .gen-grid > div { position: relative; }
  .gen-grid .c2 { grid-column: span 2; }
  .gen-grid .c3 { grid-column: span 3; }
  .gen-grid .c4 { grid-column: span 4; }
  .gen-grid .c6 { grid-column: span 6; }
  /* UoM (lebar) di sebelah Price (sempit) */
  .uom-price-row { display: grid; grid-template-columns: 1.55fr 1fr 1.2fr; gap: 18px; align-items: stretch; }
  .uom-price-row > .pcard { height: 100%; margin-top: 0; display: flex; flex-direction: column; }
  .uom-price-row > .pcard > .pcard-body { flex: 1; }
  /* General Information (kiri) sejajar Images (kanan, sempit ~ lebar Dimensions) */
  .gi-img-row { display: grid; grid-template-columns: 1fr 343px; gap: 18px; align-items: stretch; }
  .gi-img-row > .pcard { margin-top: 0; }
  /* kolom kanan: Images di atas, Other Information mengisi ruang di bawahnya
     hingga sisi bawahnya sejajar dengan General Information */
  .gi-right { display: flex; flex-direction: column; gap: 18px; }
  .gi-right > .pcard { margin-top: 0; }
  .gi-right .oi-card { flex: 1; }
  /* Other Information: ringkasan read-only (dari DB/inventory) */
  .oi-body { gap: 12px; }
  .oi-field { position: relative; padding-bottom: 14px; border-bottom: 1px solid var(--border-color); }
  .oi-line { font-size: 14px; color: var(--text-secondary); font-weight: 600; display: flex; justify-content: space-between; align-items: baseline; gap: 12px; }
  .oi-line strong { color: var(--text-primary); font-weight: 800; text-align: right; }
  .oi-unit { color: var(--text-secondary); font-weight: 700; margin-left: 6px; }
  /* garis batas di atas Total COGS, dengan tanda x di ujung kanannya */
  .oi-divider { position: relative; border-top: 1px solid var(--border-color); height: 0; }
  .oi-mult {
    position: absolute; right: 0; top: 0; transform: translateY(-50%);
    padding-left: 6px; background: var(--bg-surface);
    color: var(--accent); font-weight: 800; font-size: 16px; line-height: 1;
  }

  /* Dimensions: kiri (Length/Width/Height vertikal) + kanan (Weight/COGS) */
  .dim-grid { display: grid; grid-template-columns: 108px 1fr; gap: 18px 20px; }
  .dim-col { display: flex; flex-direction: column; gap: 20px; }
  .dim-col > div { position: relative; }
  /* ilustrasi dimensi (SVG box dengan nilai live) */
  .dim-illus { display: flex; justify-content: center; margin-top: 4px; }
  .dim-svg { width: 100%; max-width: 195px; height: auto; }
  .dim-svg text, .dim-svg tspan { font-family: var(--font-family); }
  .box-face { fill: none; stroke: var(--text-secondary); stroke-width: 1.4; stroke-linejoin: round; }
  .dim-arrow { stroke: var(--accent); stroke-width: 1; }
  .dim-head { fill: var(--accent); }
  .dim-lbl { fill: var(--text-secondary); font-weight: 800; font-size: 9px; }
  .dim-val { fill: var(--text-primary); font-weight: 700; font-size: 9px; }
  .dim-unit { fill: var(--text-muted); font-weight: 600; font-size: 7.5px; }
  .dim-cap { fill: var(--text-secondary); font-weight: 700; font-size: 9px; }

  /* SKU: header (label + toggle Auto/Manual) */
  .sku-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 6px; }
  .sku-head .form-label { margin-bottom: 0; }
  /* Product Name: label kiri + checkbox Perishable di ujung kanan (satu baris) */
  .pn-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 11px; }
  .pn-head .form-label { margin-bottom: 0; }

  /* header field dengan tombol "+" quick-add (Brand/Category/Subcategory) */
  .field-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 11px; }
  .field-head .form-label { margin-bottom: 0; }
  .label-add {
    display: inline-flex; align-items: center; justify-content: center; flex: 0 0 auto;
    width: 22px; height: 22px; border-radius: 6px; border: 1px solid var(--border-color);
    background: var(--bg-surface); color: var(--accent); cursor: pointer; font-size: 14px;
  }
  .label-add:hover { background: var(--accent); color: #fff; border-color: var(--accent); }
  /* form quick-add di dalam modal */
  .qa-form { display: flex; flex-direction: column; gap: 16px; }
  .qa-field { position: relative; }
  .qa-field > .form-label { display: block; }
  .sku-toggle { display: inline-flex; border: 1px solid var(--border-color); border-radius: 8px; }
  .sku-toggle button { border: none; background: var(--bg-surface); color: var(--text-secondary); font-size: 11px; font-weight: 700; padding: 3px 10px; cursor: pointer; border-radius: 0; }
  .sku-toggle button:first-child { border-radius: 7px 0 0 7px; }
  .sku-toggle button:last-child { border-radius: 0 7px 7px 0; }
  .sku-toggle button.active { background: var(--accent); color: #fff; }

  /* Tooltip kustom (hover) — dipakai di SKU Auto/Manual & Perishable */
  .tip { position: relative; }
  .tip::before, .tip::after { opacity: 0; visibility: hidden; transition: opacity 0.15s ease; pointer-events: none; }
  .tip::after {
    content: attr(data-tip);
    position: absolute; z-index: 60; top: calc(100% + 9px); left: 0;
    width: 250px; white-space: normal; text-align: left; font-weight: 500;
    background: var(--text-primary); color: var(--bg-surface);
    font-size: 12px; line-height: 1.5; padding: 9px 11px; border-radius: 9px;
    box-shadow: 0 10px 26px rgba(0, 0, 0, 0.3);
  }
  .tip::before {
    content: ''; position: absolute; z-index: 61; top: calc(100% + 3px); left: 12px;
    border: 6px solid transparent; border-bottom-color: var(--text-primary);
  }
  .tip:hover::after, .tip:hover::before { opacity: 1; visibility: visible; }
  .tip--right::after { left: auto; right: 0; }
  .tip--right::before { left: auto; right: 12px; }
  .sku-locked { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: 0.3px; color: var(--text-secondary); }

  .checkbox-cell { display: flex; align-items: flex-end; padding-bottom: 8px; }
  .checkbox-row { display: inline-flex; align-items: center; gap: 8px; font-size: 14px; color: var(--text-secondary); font-weight: 600; cursor: pointer; }

  textarea.text-input { resize: vertical; }
  .text-input:disabled { opacity: 0.7; cursor: not-allowed; }
  .input-error { border-color: var(--danger) !important; }
  .label-required { color: var(--danger); font-size: 12px; font-weight: 700; margin-left: 8px; }
  .field-tip {
    position: absolute; top: calc(100% + 8px); left: 0; z-index: 30; max-width: 260px;
    background: var(--danger); color: #fff; font-size: 12px; font-weight: 600; line-height: 1.35;
    padding: 7px 10px; border-radius: 8px; box-shadow: 0 8px 20px rgba(0, 0, 0, 0.2);
  }
  .field-tip::before { content: ''; position: absolute; bottom: 100%; left: 16px; border: 5px solid transparent; border-bottom-color: var(--danger); }

  /* image slots */
  .img-slots { display: flex; gap: 14px; flex-wrap: wrap; }
  /* Main Image (kotak besar) + Image 1-3 (sejajar horizontal, kotak sama) */
  .img-main { width: 100%; }
  .img-main .img-slot { width: 100%; height: auto; aspect-ratio: 1 / 1; }
  .img-thumbs { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; }
  .img-thumbs .img-slot-wrap { width: 100%; }
  .img-thumbs .img-slot { width: 100%; height: auto; aspect-ratio: 1 / 1; }
  .img-main .img-x, .img-thumbs .img-x { right: 6px; }
  .img-slot-wrap { position: relative; display: flex; flex-direction: column; gap: 6px; }
  .img-slot {
    width: 110px; height: 110px; border-radius: 12px; overflow: hidden; display: flex; flex-direction: column;
    align-items: center; justify-content: center; gap: 6px; background: var(--bg-surface);
    border: 1.5px dashed var(--border-color); cursor: pointer;
  }
  .img-slot:hover { border-color: var(--accent); }
  .img-slot img { width: 100%; height: 100%; object-fit: cover; }
  .img-ic { font-size: 22px; color: var(--text-muted); }
  .img-hint { font-size: 11px; color: var(--text-muted); font-weight: 600; }
  .img-note { display: flex; flex-direction: column; gap: 2px; margin-top: 4px; }
  .img-formats { font-size: 11px; font-weight: 600; color: var(--text-muted); }
  .img-max { font-size: 11px; font-weight: 700; color: var(--danger); }
  .img-x {
    position: absolute; top: 24px; right: -6px; width: 20px; height: 20px; border-radius: 50%;
    display: flex; align-items: center; justify-content: center; background: var(--danger); color: #fff;
    font-size: 12px; border: 2px solid var(--bg-surface); cursor: pointer;
  }
  .img-cap { font-size: 12px; color: var(--text-secondary); font-weight: 600; }

  @media (max-width: 820px) {
    .pform-cols, .grid-2, .uom-price, .gen-grid, .uom-price-row, .gi-img-row, .dim-grid, .var-axes, .apply-grid { grid-template-columns: 1fr; }
    .var-banner { flex-direction: column; align-items: stretch; }
    .grid-3, .grid-4 { grid-template-columns: 1fr 1fr; }
    .coa-grid { grid-template-columns: 1fr 1fr; grid-template-rows: none; grid-auto-flow: row; }
    .gen-grid .c2, .gen-grid .c3, .gen-grid .c4, .gen-grid .c6 { grid-column: span 1; }
  }
  .spin { animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  .modal-actions {
    display: flex; justify-content: flex-end; gap: 10px; margin-top: 18px; padding-top: 16px;
    border-top: 1px solid var(--border-color);
  }
  .btn-ghost {
    padding: 8px 18px; border-radius: 10px; background: var(--bg-muted); color: var(--text-secondary);
    font-size: 14px; font-weight: 700; border: none; cursor: pointer;
  }
  .btn-ghost:hover { background: var(--bg-hover); color: var(--text-primary); }

  .table-loading { display: flex; align-items: center; justify-content: center; gap: 8px; padding: 14px; font-size: 13px; color: var(--text-muted); }
  .toggle { display: inline-flex; align-items: center; width: 42px; height: 24px; border-radius: 999px; background: var(--bg-muted); border: none; cursor: pointer; padding: 0; position: relative; transition: background 0.15s ease; }
  .toggle.on { background: var(--success); }
  .toggle-knob { position: absolute; top: 3px; left: 3px; width: 18px; height: 18px; border-radius: 50%; background: #fff; transition: left 0.15s ease; box-shadow: 0 1px 2px rgba(0, 0, 0, 0.25); }
  .toggle.on .toggle-knob { left: 21px; }

  /* Kolom Image */
  .col-img { width: 56px; }
  .list-img {
    width: 42px; height: 42px; border-radius: 8px; overflow: hidden; background: var(--bg-muted);
    display: flex; align-items: center; justify-content: center; border: 1px solid var(--border-color);
  }
  .list-img img { width: 100%; height: 100%; object-fit: cover; }
  .list-img-ph { color: var(--text-muted); width: 18px; height: 18px; }
  /* Kolom Product (sku kecil / nama / type + country) */
  .prod-info { display: flex; flex-direction: column; gap: 3px; }
  .prod-line { display: flex; align-items: center; gap: 8px; }
  .prod-sku { font-size: 11px; color: var(--text-muted); font-weight: 600; }
  .prod-name { font-weight: 700; color: var(--text-primary); }
  .type-badge {
    font-size: 9px; font-weight: 700; letter-spacing: 0.02em;
    padding: 1px 6px; border-radius: 999px; background: var(--bg-muted); color: var(--text-secondary);
  }
  .type-badge.variant { background: var(--accent-light); color: var(--accent); }
  .prod-country { font-size: 12px; color: var(--text-muted); font-weight: 600; }
  .unit-suffix { color: var(--text-muted); font-size: 12px; font-weight: 400; }
</style>
