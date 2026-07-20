export interface ProductListItem {
  id: string          // product id (untuk buka edit)
  variant_id: string  // id varian (key baris + toggle status)
  product_name: string
  product_type: string
  variant_value_1: string
  variant_value_2: string
  brand_name: string
  subcategory_name: string
  category_name: string
  country_of_origin: string
  sku: string
  barcode: string
  price: number
  image: string
  selling_uom_name: string
  cogs_selling: number
  is_active: boolean
  created_at: string
}

export interface Opt { id: string; name: string }
export interface CountryOpt { code: string; name: string; phone_code: string }
export interface CoaAccountOpt {
  id: string
  account_code: string
  account_name: string
  account_type_name: string
  classification_name: string
  is_contra: boolean
}

const items = ref<ProductListItem[]>([])
const total = ref(0)

export interface ProductFetchOpts {
  offset: number
  limit: number
  search?: string
  sort?: string
  desc?: boolean
  product_type?: string
  country?: string
  brand_id?: string
  category_id?: string
  subcategory_id?: string
  status?: string
  append?: boolean
}

export function useProducts() {
  const toast = useToast()

  const fetchPage = async (opts: ProductFetchOpts): Promise<number> => {
    const q = new URLSearchParams()
    q.set('limit', String(opts.limit))
    q.set('offset', String(opts.offset))
    if (opts.search) q.set('search', opts.search)
    if (opts.sort) q.set('sort', opts.sort)
    if (opts.desc) q.set('desc', 'true')
    if (opts.product_type) q.set('product_type', opts.product_type)
    if (opts.country) q.set('country', opts.country)
    if (opts.brand_id) q.set('brand_id', opts.brand_id)
    if (opts.category_id) q.set('category_id', opts.category_id)
    if (opts.subcategory_id) q.set('subcategory_id', opts.subcategory_id)
    if (opts.status) q.set('status', opts.status)
    try {
      const res = await apiFetch<{ data: ProductListItem[]; total: number }>(`/api/v1/products?${q.toString()}`)
      total.value = res.total ?? 0
      const rows = res.data ?? []
      if (opts.append) items.value.push(...rows)
      else items.value = rows
      return rows.length
    } catch (e) {
      console.error('Failed to fetch products:', e)
      toast.add({ title: 'Error', description: 'Failed to load products', color: 'error' })
      return 0
    }
  }

  const getProduct = async (id: string): Promise<any> => {
    const res = await apiFetch<{ data: any }>(`/api/v1/products/${id}`)
    return res.data
  }

  // Pratinjau SKU otomatis berikutnya (PSKU-YYYYMM-######) dari server.
  const getNextSku = async (): Promise<string> => {
    const res = await apiFetch<{ data: { sku: string } }>(`/api/v1/product-next-sku`)
    return res.data?.sku ?? ''
  }

  const createProduct = async (body: any) => {
    await apiFetch('/api/v1/products', { method: 'POST', body })
    toast.add({ title: 'Saved', description: 'Product created', color: 'success' })
  }

  const updateProduct = async (id: string, body: any) => {
    await apiFetch(`/api/v1/products/${id}`, { method: 'PUT', body })
    toast.add({ title: 'Saved', description: 'Product updated', color: 'success' })
  }

  const toggleActive = async (p: ProductListItem) => {
    try {
      await apiFetch(`/api/v1/products/${p.id}/toggle-active`, { method: 'PATCH', body: { is_active: !p.is_active } })
      p.is_active = !p.is_active
      toast.add({ title: 'Status updated', description: `${p.product_name} is now ${p.is_active ? 'active' : 'inactive'}`, color: 'success' })
    } catch (e) {
      console.error('Failed to toggle product:', e)
      toast.add({ title: 'Error', description: 'Failed to update status', color: 'error' })
    }
  }

  // Toggle status satu varian (list flat per varian).
  const toggleVariantActive = async (p: ProductListItem) => {
    try {
      await apiFetch(`/api/v1/product-variants/${p.variant_id}/toggle-active`, { method: 'PATCH', body: { is_active: !p.is_active } })
      p.is_active = !p.is_active
      toast.add({ title: 'Status updated', description: `${p.sku || p.product_name} is now ${p.is_active ? 'active' : 'inactive'}`, color: 'success' })
    } catch (e) {
      console.error('Failed to toggle variant:', e)
      toast.add({ title: 'Error', description: 'Failed to update status', color: 'error' })
    }
  }

  const uploadImage = async (file: File, folder = 'product_image'): Promise<string> => {
    const fd = new FormData()
    fd.append('file', file)
    const res = await apiFetch<{ path: string }>(`/api/v1/upload/image?folder=${folder}`, { method: 'POST', body: fd })
    return res.path
  }

  // ── dropdown option fetchers (hanya yang aktif) ──
  const fetchBrandOptions = async (): Promise<Opt[]> => {
    const res = await apiFetch<{ data: { id: string; name: string; is_active: boolean }[] }>(`/api/v1/brands?limit=1000&sort=name`)
    return (res.data ?? []).filter(b => b.is_active).map(b => ({ id: b.id, name: b.name }))
  }
  const fetchCategoryOptions = async (): Promise<Opt[]> => {
    const res = await apiFetch<{ data: { id: string; name: string; is_active: boolean }[] }>(`/api/v1/categories?limit=1000&sort=name`)
    return (res.data ?? []).filter(c => c.is_active).map(c => ({ id: c.id, name: c.name }))
  }
  const fetchSubcategoryOptions = async (categoryId: string): Promise<Opt[]> => {
    if (!categoryId) return []
    const res = await apiFetch<{ data: { id: string; name: string; is_active: boolean }[] }>(`/api/v1/subcategories?limit=1000&sort=name&category_id=${categoryId}`)
    return (res.data ?? []).filter(s => s.is_active).map(s => ({ id: s.id, name: s.name }))
  }
  const fetchUomOptions = async (): Promise<Opt[]> => {
    const res = await apiFetch<{ data: { id: string; name: string; is_active: boolean }[] }>(`/api/v1/uoms?limit=1000&sort=name`)
    return (res.data ?? []).filter(u => u.is_active).map(u => ({ id: u.id, name: u.name }))
  }
  const fetchCountryOptions = async (): Promise<CountryOpt[]> => {
    const res = await apiFetch<{ data: CountryOpt[] }>(`/api/v1/countries`)
    return res.data ?? []
  }
  // Akun COA aktif (untuk dropdown Chart of Accounts di form produk).
  // Endpoint khusus: mengembalikan SEMUA akun aktif tanpa paginasi.
  const fetchCoaOptions = async (): Promise<CoaAccountOpt[]> => {
    const res = await apiFetch<{ data: CoaAccountOpt[] }>(`/api/v1/coa-options`)
    return res.data ?? []
  }

  // ── quick-add master (dari form produk) → kembalikan opsi baru ──
  const createBrand = async (body: { name: string; description?: string; logo?: string }): Promise<Opt> => {
    const res = await apiFetch<{ data: { id: string; name: string } }>(`/api/v1/brands`, {
      method: 'POST', body: { name: body.name, description: body.description ?? '', logo: body.logo ?? '' },
    })
    return { id: res.data.id, name: res.data.name }
  }
  const createCategory = async (body: { name: string; banner_image?: string }): Promise<Opt> => {
    const res = await apiFetch<{ data: { id: string; name: string } }>(`/api/v1/categories`, {
      method: 'POST', body: { name: body.name, banner_image: body.banner_image ?? '' },
    })
    return { id: res.data.id, name: res.data.name }
  }
  const createSubcategory = async (name: string, categoryId: string): Promise<Opt> => {
    const res = await apiFetch<{ data: { id: string; name: string } }>(`/api/v1/subcategories`, { method: 'POST', body: { name, category_id: categoryId } })
    return { id: res.data.id, name: res.data.name }
  }

  return {
    items, total, fetchPage, getProduct, getNextSku, createProduct, updateProduct, toggleActive, toggleVariantActive, uploadImage,
    fetchBrandOptions, fetchCategoryOptions, fetchSubcategoryOptions, fetchUomOptions, fetchCountryOptions, fetchCoaOptions,
    createBrand, createCategory, createSubcategory,
  }
}

// Daftar negara PLACEHOLDER (menyusul: diambil dari country list resmi).
export const COUNTRY_PLACEHOLDER: string[] = [
  'Indonesia', 'Malaysia', 'Singapore', 'Thailand', 'Vietnam',
  'China', 'Japan', 'South Korea', 'India', 'United States',
]
