export interface ManagedUom {
  id: string
  name: string
  description: string
  is_active: boolean
  created_at: string
}

const items = ref<ManagedUom[]>([])
const total = ref(0)

export interface UomFetchOpts {
  offset: number
  limit: number
  search?: string
  sort?: string
  desc?: boolean
  status?: string
  append?: boolean
}

export function useUoms() {
  const toast = useToast()

  const fetchPage = async (opts: UomFetchOpts): Promise<number> => {
    const q = new URLSearchParams()
    q.set('limit', String(opts.limit))
    q.set('offset', String(opts.offset))
    if (opts.search) q.set('search', opts.search)
    if (opts.sort) q.set('sort', opts.sort)
    if (opts.desc) q.set('desc', 'true')
    if (opts.status) q.set('status', opts.status)
    try {
      const res = await apiFetch<{ data: ManagedUom[]; total: number }>(`/api/v1/uoms?${q.toString()}`)
      total.value = res.total ?? 0
      const rows = res.data ?? []
      if (opts.append) items.value.push(...rows)
      else items.value = rows
      return rows.length
    } catch (e) {
      console.error('Failed to fetch uoms:', e)
      toast.add({ title: 'Error', description: 'Failed to load units', color: 'error' })
      return 0
    }
  }

  const createUom = async (body: { name: string; description: string }) => {
    await apiFetch('/api/v1/uoms', { method: 'POST', body })
    toast.add({ title: 'Saved', description: 'Unit created', color: 'success' })
  }

  const updateUom = async (id: string, body: { name: string; description: string }) => {
    await apiFetch(`/api/v1/uoms/${id}`, { method: 'PUT', body })
    toast.add({ title: 'Saved', description: 'Unit updated', color: 'success' })
  }

  const toggleActive = async (u: ManagedUom) => {
    try {
      await apiFetch(`/api/v1/uoms/${u.id}/toggle-active`, { method: 'PATCH', body: { is_active: !u.is_active } })
      u.is_active = !u.is_active
      toast.add({ title: 'Status updated', description: `${u.name} is now ${u.is_active ? 'active' : 'inactive'}`, color: 'success' })
    } catch (e) {
      console.error('Failed to toggle uom:', e)
      toast.add({ title: 'Error', description: 'Failed to update status', color: 'error' })
    }
  }

  return { items, total, fetchPage, createUom, updateUom, toggleActive }
}
