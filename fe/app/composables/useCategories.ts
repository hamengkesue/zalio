export interface ManagedCategory {
  id: string
  name: string
  banner_image: string
  in_use: boolean
  is_active: boolean
  created_at: string
}

const items = ref<ManagedCategory[]>([])
const total = ref(0)

export interface CategoryFetchOpts {
  offset: number
  limit: number
  search?: string
  sort?: string
  desc?: boolean
  status?: string
  append?: boolean
}

export function useCategories() {
  const toast = useToast()

  const fetchPage = async (opts: CategoryFetchOpts): Promise<number> => {
    const q = new URLSearchParams()
    q.set('limit', String(opts.limit))
    q.set('offset', String(opts.offset))
    if (opts.search) q.set('search', opts.search)
    if (opts.sort) q.set('sort', opts.sort)
    if (opts.desc) q.set('desc', 'true')
    if (opts.status) q.set('status', opts.status)
    try {
      const res = await apiFetch<{ data: ManagedCategory[]; total: number }>(`/api/v1/categories?${q.toString()}`)
      total.value = res.total ?? 0
      const rows = res.data ?? []
      if (opts.append) items.value.push(...rows)
      else items.value = rows
      return rows.length
    } catch (e) {
      console.error('Failed to fetch categories:', e)
      toast.add({ title: 'Error', description: 'Failed to load categories', color: 'error' })
      return 0
    }
  }

  const createCategory = async (body: { name: string; banner_image: string }) => {
    await apiFetch('/api/v1/categories', { method: 'POST', body })
    toast.add({ title: 'Saved', description: 'Category created', color: 'success' })
  }

  const updateCategory = async (id: string, body: { name: string; banner_image: string }) => {
    await apiFetch(`/api/v1/categories/${id}`, { method: 'PUT', body })
    toast.add({ title: 'Saved', description: 'Category updated', color: 'success' })
  }

  const uploadImage = async (file: File): Promise<string> => {
    const fd = new FormData()
    fd.append('file', file)
    const res = await apiFetch<{ path: string }>('/api/v1/upload/image?folder=category_banner', { method: 'POST', body: fd })
    return res.path
  }

  const toggleActive = async (cat: ManagedCategory) => {
    try {
      await apiFetch(`/api/v1/categories/${cat.id}/toggle-active`, { method: 'PATCH', body: { is_active: !cat.is_active } })
      cat.is_active = !cat.is_active
      toast.add({ title: 'Status updated', description: `${cat.name} is now ${cat.is_active ? 'active' : 'inactive'}`, color: 'success' })
    } catch (e) {
      console.error('Failed to toggle category:', e)
      toast.add({ title: 'Error', description: 'Failed to update status', color: 'error' })
    }
  }

  return { items, total, fetchPage, createCategory, updateCategory, uploadImage, toggleActive }
}
