export interface ManagedSubcategory {
  id: string
  name: string
  category_id: string
  category_name: string
  is_active: boolean
  created_at: string
}

export interface CategoryOption {
  id: string
  name: string
}

const items = ref<ManagedSubcategory[]>([])
const total = ref(0)

export interface SubcategoryFetchOpts {
  offset: number
  limit: number
  search?: string
  sort?: string
  desc?: boolean
  status?: string
  category_id?: string
  append?: boolean
}

export function useSubcategories() {
  const toast = useToast()

  const fetchPage = async (opts: SubcategoryFetchOpts): Promise<number> => {
    const q = new URLSearchParams()
    q.set('limit', String(opts.limit))
    q.set('offset', String(opts.offset))
    if (opts.search) q.set('search', opts.search)
    if (opts.sort) q.set('sort', opts.sort)
    if (opts.desc) q.set('desc', 'true')
    if (opts.status) q.set('status', opts.status)
    if (opts.category_id) q.set('category_id', opts.category_id)
    try {
      const res = await apiFetch<{ data: ManagedSubcategory[]; total: number }>(`/api/v1/subcategories?${q.toString()}`)
      total.value = res.total ?? 0
      const rows = res.data ?? []
      if (opts.append) items.value.push(...rows)
      else items.value = rows
      return rows.length
    } catch (e) {
      console.error('Failed to fetch subcategories:', e)
      toast.add({ title: 'Error', description: 'Failed to load subcategories', color: 'error' })
      return 0
    }
  }

  // Ambil daftar kategori aktif untuk dropdown di form.
  const fetchCategoryOptions = async (): Promise<CategoryOption[]> => {
    try {
      const res = await apiFetch<{ data: { id: string; name: string; is_active: boolean }[] }>(`/api/v1/categories?limit=1000&sort=name`)
      return (res.data ?? []).filter(c => c.is_active).map(c => ({ id: c.id, name: c.name }))
    } catch (e) {
      console.error('Failed to fetch category options:', e)
      return []
    }
  }

  const createSubcategory = async (body: { name: string; category_id: string }) => {
    await apiFetch('/api/v1/subcategories', { method: 'POST', body })
    toast.add({ title: 'Saved', description: 'Subcategory created', color: 'success' })
  }

  const updateSubcategory = async (id: string, body: { name: string; category_id: string }) => {
    await apiFetch(`/api/v1/subcategories/${id}`, { method: 'PUT', body })
    toast.add({ title: 'Saved', description: 'Subcategory updated', color: 'success' })
  }

  // Sukses → toast. Error DILEMPAR ke pemanggil (halaman) supaya bisa
  // ditampilkan sebagai tooltip di dekat toggle, bukan toast.
  const toggleActive = async (s: ManagedSubcategory) => {
    await apiFetch(`/api/v1/subcategories/${s.id}/toggle-active`, { method: 'PATCH', body: { is_active: !s.is_active } })
    s.is_active = !s.is_active
    toast.add({ title: 'Status updated', description: `${s.name} is now ${s.is_active ? 'active' : 'inactive'}`, color: 'success' })
  }

  return { items, total, fetchPage, fetchCategoryOptions, createSubcategory, updateSubcategory, toggleActive }
}
