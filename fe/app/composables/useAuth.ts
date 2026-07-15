export interface AuthUser {
  id: number
  name: string
  username: string
  email: string
  role: string
  is_active: boolean
}

// State user dibagi ke seluruh app (module-level ref).
const user = ref<AuthUser | null>(null)

export function useAuth() {
  // Token disimpan di cookie (bertahan saat refresh, dibaca middleware).
  const token = useCookie<string | null>('zalio_token', {
    maxAge: 60 * 60 * 8,
    sameSite: 'lax',
  })
  const toast = useToast()

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(username: string, password: string) {
    const res = await $fetch<{ token: string; user: AuthUser }>(
      `${API_BASE}/api/v1/auth/login`,
      { method: 'POST', body: { username, password } },
    )
    token.value = res.token
    user.value = res.user
  }

  // Ambil profil user dari token yang tersimpan (dipanggil saat app dibuka).
  async function fetchMe() {
    if (!token.value) {
      user.value = null
      return
    }
    try {
      const res = await apiFetch<{ data: AuthUser }>('/api/v1/auth/me')
      user.value = res.data
    } catch {
      // token kedaluwarsa / tidak valid → paksa login ulang
      token.value = null
      user.value = null
      await navigateTo('/login')
    }
  }

  function logout() {
    token.value = null
    user.value = null
    toast.add({ title: 'Signed out', description: 'You have been logged out', color: 'info' })
    navigateTo('/login')
  }

  return { user, token, isAuthenticated, isAdmin, login, fetchMe, logout }
}
