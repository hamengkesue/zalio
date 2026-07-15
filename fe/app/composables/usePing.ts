export interface Ping {
  id: number
  message: string
  created_at: string
}

// Alamat backend Go (Fase 0 memakai port 8082).
export const API_BASE = 'http://localhost:8082'

const pings = ref<Ping[]>([])

export function usePing() {
  const toast = useToast()

  const fetchPings = async () => {
    try {
      const res = await $fetch<{ data: Ping[] }>(`${API_BASE}/api/v1/ping`)
      pings.value = res.data ?? []
    } catch (e) {
      console.error('Failed to fetch pings:', e)
      toast.add({ title: 'Error', description: 'Gagal memuat data dari backend', color: 'error' })
    }
  }

  const createPing = async (message: string) => {
    const res = await $fetch<{ data: Ping }>(`${API_BASE}/api/v1/ping`, {
      method: 'POST',
      body: { message },
    })
    await fetchPings()
    toast.add({ title: 'Tersimpan', description: 'Pesan baru ditambahkan ke database', color: 'success' })
    return res.data
  }

  return { pings, fetchPings, createPing }
}
