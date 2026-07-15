export interface ManagedUser {
  id: number
  name: string
  username: string
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
      toast.add({ title: 'Error', description: 'Failed to load user list', color: 'error' })
    }
  }

  const createUser = async (body: { name: string; username: string; email: string; password: string; role: string }) => {
    try {
      await apiFetch('/api/v1/users', { method: 'POST', body })
      await fetchUsers()
      toast.add({ title: 'Saved', description: 'New user created', color: 'success' })
    } catch (e: any) {
      toast.add({ title: 'Failed', description: e?.data?.error || 'Could not create user', color: 'error' })
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
        title: 'Status updated',
        description: `${u.name} is now ${u.is_active ? 'active' : 'inactive'}`,
        color: 'success',
      })
    } catch (e) {
      console.error('Failed to toggle user:', e)
      toast.add({ title: 'Error', description: 'Failed to update status', color: 'error' })
    }
  }

  return { users, fetchUsers, createUser, toggleActive }
}
