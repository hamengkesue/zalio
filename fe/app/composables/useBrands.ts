export interface ManagedBrand {
  id: string
  name: string
  description: string
  logo: string
  is_active: boolean
  created_at: string
}

const items = ref<ManagedBrand[]>([])
const total = ref(0)

export interface BrandFetchOpts {
  offset: number
  limit: number
  search?: string
  sort?: string
  desc?: boolean
  status?: string
  append?: boolean
}

export function useBrands() {
  const toast = useToast()

  const fetchPage = async (opts: BrandFetchOpts): Promise<number> => {
    const q = new URLSearchParams()
    q.set('limit', String(opts.limit))
    q.set('offset', String(opts.offset))
    if (opts.search) q.set('search', opts.search)
    if (opts.sort) q.set('sort', opts.sort)
    if (opts.desc) q.set('desc', 'true')
    if (opts.status) q.set('status', opts.status)
    try {
      const res = await apiFetch<{ data: ManagedBrand[]; total: number }>(`/api/v1/brands?${q.toString()}`)
      total.value = res.total ?? 0
      const rows = res.data ?? []
      if (opts.append) items.value.push(...rows)
      else items.value = rows
      return rows.length
    } catch (e) {
      console.error('Failed to fetch brands:', e)
      toast.add({ title: 'Error', description: 'Failed to load brands', color: 'error' })
      return 0
    }
  }

  const createBrand = async (body: { name: string; description: string; logo: string }) => {
    await apiFetch('/api/v1/brands', { method: 'POST', body })
    toast.add({ title: 'Saved', description: 'Brand created', color: 'success' })
  }

  const updateBrand = async (id: string, body: { name: string; description: string; logo: string }) => {
    await apiFetch(`/api/v1/brands/${id}`, { method: 'PUT', body })
    toast.add({ title: 'Saved', description: 'Brand updated', color: 'success' })
  }

  // Upload logo → MinIO folder brand_logo; kembalikan path.
  const uploadImage = async (file: File): Promise<string> => {
    const fd = new FormData()
    fd.append('file', file)
    const res = await apiFetch<{ path: string }>('/api/v1/upload/image?folder=brand_logo', { method: 'POST', body: fd })
    return res.path
  }

  const toggleActive = async (b: ManagedBrand) => {
    try {
      await apiFetch(`/api/v1/brands/${b.id}/toggle-active`, { method: 'PATCH', body: { is_active: !b.is_active } })
      b.is_active = !b.is_active
      toast.add({ title: 'Status updated', description: `${b.name} is now ${b.is_active ? 'active' : 'inactive'}`, color: 'success' })
    } catch (e) {
      console.error('Failed to toggle brand:', e)
      toast.add({ title: 'Error', description: 'Failed to update status', color: 'error' })
    }
  }

  return { items, total, fetchPage, createBrand, updateBrand, uploadImage, toggleActive }
}
