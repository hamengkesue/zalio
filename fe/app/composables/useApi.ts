// Alamat backend Go (Fase 0 memakai port 8082).
export const API_BASE = 'http://localhost:8082'

// apiFetch = $fetch yang otomatis menyertakan token login (dari cookie)
// di header Authorization. Dipakai semua request yang butuh login.
export function apiFetch<T>(path: string, opts: Record<string, any> = {}) {
  const token = useCookie<string | null>('zalio_token')
  return $fetch<T>(`${API_BASE}${path}`, {
    ...opts,
    headers: {
      ...(opts.headers || {}),
      ...(token.value ? { Authorization: `Bearer ${token.value}` } : {}),
    },
  })
}
