export interface ManagedUser {
  id: string
  name: string
  username: string
  email: string
  whatsapp: string
  group_access: string
  profile_image: string
  role: string
  is_active: boolean
  created_at: string
}

const users = ref<ManagedUser[]>([])
const total = ref(0) // total baris yang cocok di server (untuk pager & stop infinite scroll)

export interface FetchPageOpts {
  offset: number
  limit: number
  search?: string
  sort?: string
  desc?: boolean
  role?: string   // ''=semua, 'admin', 'staff'
  status?: string // ''=semua, 'active', 'inactive'
  append?: boolean // true = tambahkan ke bawah (infinite scroll), false = ganti (reset)
}

export function useUsers() {
  const toast = useToast()

  // fetchPage: ambil satu batch dari server. Kembalikan jumlah baris yang didapat.
  const fetchPage = async (opts: FetchPageOpts): Promise<number> => {
    const q = new URLSearchParams()
    q.set('limit', String(opts.limit))
    q.set('offset', String(opts.offset))
    if (opts.search) q.set('search', opts.search)
    if (opts.sort) q.set('sort', opts.sort)
    if (opts.desc) q.set('desc', 'true')
    if (opts.role) q.set('role', opts.role)
    if (opts.status) q.set('status', opts.status)
    try {
      const res = await apiFetch<{ data: ManagedUser[]; total: number }>(`/api/v1/users?${q.toString()}`)
      total.value = res.total ?? 0
      const rows = res.data ?? []
      if (opts.append) users.value.push(...rows)
      else users.value = rows
      return rows.length
    } catch (e) {
      console.error('Failed to fetch users:', e)
      toast.add({ title: 'Error', description: 'Failed to load user list', color: 'error' })
      return 0
    }
  }

  const createUser = async (body: { name: string; username: string; email: string; whatsapp: string; group_access: string; profile_image: string; password: string; role: string }) => {
    // Error dilempar ke pemanggil (halaman) supaya bisa ditampilkan inline per-field.
    // Pemanggil bertanggung jawab reload daftar (reset ke batch pertama) setelah ini.
    await apiFetch('/api/v1/users', { method: 'POST', body })
    toast.add({ title: 'Saved', description: 'New user created', color: 'success' })
  }

  const updateUser = async (id: string, body: { name: string; email: string; whatsapp: string; group_access: string; profile_image: string; password: string; role: string }) => {
    await apiFetch(`/api/v1/users/${id}`, { method: 'PUT', body })
    toast.add({ title: 'Saved', description: 'User updated', color: 'success' })
  }

  // Upload gambar profil → MinIO; kembalikan path yang disimpan.
  const uploadProfileImage = async (file: File): Promise<string> => {
    const fd = new FormData()
    fd.append('file', file)
    const res = await apiFetch<{ path: string }>('/api/v1/upload/profile-image', { method: 'POST', body: fd })
    return res.path
  }

  const toggleActive = async (u: ManagedUser) => {
    try {
      await apiFetch(`/api/v1/users/${u.id}/toggle-active`, {
        method: 'PATCH',
        body: { is_active: !u.is_active },
      })
      u.is_active = !u.is_active
      toast.add({
        title: 'Status updated',
        description: `${u.name} is now ${u.is_active ? 'active' : 'inactive'}`,
        color: 'success',
      })
    } catch (e) {
      console.error('Failed to toggle user:', e)
      toast.add({ title: 'Error', description: 'Failed to update status', color: 'error' })
    }
  }

  return { users, total, fetchPage, createUser, updateUser, uploadProfileImage, toggleActive }
}
