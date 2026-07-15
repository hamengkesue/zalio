export interface Ping {
  id: number
  message: string
  created_at: string
}

// API_BASE didefinisikan sekali di useApi.ts (auto-import Nuxt).

const pings = ref<Ping[]>([])

export function usePing() {
  const toast = useToast()

  const fetchPings = async () => {
    try {
      const res = await $fetch<{ data: Ping[] }>(`${API_BASE}/api/v1/ping`)
      pings.value = res.data ?? []
    } catch (e) {
      console.error('Failed to fetch pings:', e)
      toast.add({ title: 'Error', description: 'Failed to load data from backend', color: 'error' })
    }
  }

  const createPing = async (message: string) => {
    const res = await $fetch<{ data: Ping }>(`${API_BASE}/api/v1/ping`, {
      method: 'POST',
      body: { message },
    })
    await fetchPings()
    toast.add({ title: 'Saved', description: 'New message added to the database', color: 'success' })
    return res.data
  }

  return { pings, fetchPings, createPing }
}
