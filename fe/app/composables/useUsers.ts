export interface ManagedUser {
  id: string
  name: string
  username: string
  email: string
  whatsapp: string
  profile_image: string
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
      toast.add({ title: 'Error', description: 'Failed to load user list', color: 'error' })
    }
  }

  const createUser = async (body: { name: string; username: string; email: string; whatsapp: string; profile_image: string; password: string; role: string }) => {
    // Error dilempar ke pemanggil (users.vue) supaya bisa ditampilkan inline per-field.
    await apiFetch('/api/v1/users', { method: 'POST', body })
    await fetchUsers()
    toast.add({ title: 'Saved', description: 'New user created', color: 'success' })
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

  return { users, fetchUsers, createUser, uploadProfileImage, toggleActive }
}
