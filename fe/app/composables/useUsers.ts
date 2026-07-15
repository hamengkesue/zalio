export interface ManagedUser {
  id: number
  name: string
  email: string
  role: string
  is_active: boolean
  created_at: string
}

const users = ref<ManagedUser[]>([])

export function useUsers() {
  const toast = useToast()

  const fetchUsers = async () => {
    try {
      const res = await apiFetch<{ data: ManagedUser[] }>('/api/v1/users')
      users.value = res.data ?? []
    } catch (e) {
      console.error('Failed to fetch users:', e)
      toast.add({ title: 'Error', description: 'Gagal memuat daftar user', color: 'error' })
    }
  }

  const createUser = async (body: { name: string; email: string; password: string; role: string }) => {
    try {
      await apiFetch('/api/v1/users', { method: 'POST', body })
      await fetchUsers()
      toast.add({ title: 'Tersimpan', description: 'User baru dibuat', color: 'success' })
    } catch (e: any) {
      toast.add({ title: 'Gagal', description: e?.data?.error || 'Tidak bisa membuat user', color: 'error' })
      throw e
    }
  }

  const toggleActive = async (u: ManagedUser) => {
    try {
      await apiFetch(`/api/v1/users/${u.id}/toggle-active`, {
        method: 'PATCH',
        body: { is_active: !u.is_active },
      })
      u.is_active = !u.is_active
      toast.add({
        title: 'Status diperbarui',
        description: `${u.name} kini ${u.is_active ? 'aktif' : 'nonaktif'}`,
        color: 'success',
      })
    } catch (e) {
      console.error('Failed to toggle user:', e)
      toast.add({ title: 'Error', description: 'Gagal mengubah status', color: 'error' })
    }
  }

  return { users, fetchUsers, createUser, toggleActive }
}
