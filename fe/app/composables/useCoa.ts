export interface CoaItem {
  id: string
  account_code: string
  account_type_code: string
  account_type_name: string
  classification_name: string
  report_name: string
  account_name: string
  is_contra: boolean
  is_credit_account: boolean
  opening_balance: number
  opening_date: string | null
  notes: string
  is_active: boolean
  in_use: boolean
  created_at: string
}

export interface CoaClassification {
  classification_code: number
  classification_name: string
  report_name: string
}

export interface CoaType {
  account_type_code: string
  classification_code: number
  classification_name: string
  account_type_name: string
  is_credit: boolean
}

export interface CoaBody {
  account_name: string
  account_type_code: string
  is_contra: boolean
  opening_balance: number
  opening_date: string
  notes: string
}

const items = ref<CoaItem[]>([])
const total = ref(0)

export interface CoaFetchOpts {
  offset: number
  limit: number
  search?: string
  sort?: string
  desc?: boolean
  status?: string
  account_type?: string
  classification?: string
  append?: boolean
}

export function useCoa() {
  const toast = useToast()

  const fetchPage = async (opts: CoaFetchOpts): Promise<number> => {
    const q = new URLSearchParams()
    q.set('limit', String(opts.limit))
    q.set('offset', String(opts.offset))
    if (opts.search) q.set('search', opts.search)
    if (opts.sort) q.set('sort', opts.sort)
    if (opts.desc) q.set('desc', 'true')
    if (opts.status) q.set('status', opts.status)
    if (opts.account_type) q.set('account_type', opts.account_type)
    if (opts.classification) q.set('classification', opts.classification)
    try {
      const res = await apiFetch<{ data: CoaItem[]; total: number }>(`/api/v1/coa?${q.toString()}`)
      total.value = res.total ?? 0
      const rows = res.data ?? []
      if (opts.append) items.value.push(...rows)
      else items.value = rows
      return rows.length
    } catch (e) {
      console.error('Failed to fetch accounts:', e)
      toast.add({ title: 'Error', description: 'Failed to load accounts', color: 'error' })
      return 0
    }
  }

  const fetchClassifications = async (): Promise<CoaClassification[]> => {
    try {
      const res = await apiFetch<{ data: CoaClassification[] }>('/api/v1/coa-classifications')
      return res.data ?? []
    } catch {
      return []
    }
  }

  const fetchTypes = async (): Promise<CoaType[]> => {
    try {
      const res = await apiFetch<{ data: CoaType[] }>('/api/v1/coa-types')
      return res.data ?? []
    } catch {
      return []
    }
  }

  const createAccount = async (body: CoaBody) => {
    await apiFetch('/api/v1/coa', { method: 'POST', body })
    toast.add({ title: 'Saved', description: 'Account created', color: 'success' })
  }

  const updateAccount = async (id: string, body: CoaBody) => {
    await apiFetch(`/api/v1/coa/${id}`, { method: 'PUT', body })
    toast.add({ title: 'Saved', description: 'Account updated', color: 'success' })
  }

  const toggleActive = async (a: CoaItem) => {
    try {
      await apiFetch(`/api/v1/coa/${a.id}/toggle-active`, { method: 'PATCH', body: { is_active: !a.is_active } })
      a.is_active = !a.is_active
      toast.add({ title: 'Status updated', description: `${a.account_name} is now ${a.is_active ? 'active' : 'inactive'}`, color: 'success' })
    } catch (e) {
      console.error('Failed to toggle account:', e)
      toast.add({ title: 'Error', description: 'Failed to update status', color: 'error' })
    }
  }

  return { items, total, fetchPage, fetchClassifications, fetchTypes, createAccount, updateAccount, toggleActive }
}
