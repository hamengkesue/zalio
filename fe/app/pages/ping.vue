<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Ping' })

  const { pings, fetchPings, createPing } = usePing()
  const newMessage = ref('')
  const saving = ref(false)

  onMounted(fetchPings)

  async function submit() {
    if (!newMessage.value.trim()) return
    saving.value = true
    try {
      await createPing(newMessage.value.trim())
      newMessage.value = ''
    } finally {
      saving.value = false
    }
  }

  const fmt = (s: string) => new Date(s).toLocaleString('id-ID')
</script>

<template>
  <div class="page">
    <div class="page-body">
      <div class="page-header">
        <h1 class="page-title">Ping — Contoh Vertical Slice</h1>
        <p class="page-subtitle">
          Data di tabel bawah datang dari database lewat backend Go. Tambah pesan untuk
          membuktikan alur tulis → baca berjalan penuh.
        </p>
      </div>

      <form class="form-row" @submit.prevent="submit">
        <input
          v-model="newMessage"
          class="text-input"
          placeholder="Tulis pesan baru lalu klik Tambah..."
        >
        <button class="btn-primary" :disabled="saving" type="submit">
          {{ saving ? 'Menyimpan...' : 'Tambah' }}
        </button>
      </form>

      <div class="table-card">
        <div class="table-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width: 80px">ID</th>
                <th>Pesan</th>
                <th style="width: 200px">Dibuat</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in pings" :key="p.id">
                <td class="font-semibold">{{ p.id }}</td>
                <td>{{ p.message }}</td>
                <td>{{ fmt(p.created_at) }}</td>
              </tr>
              <tr v-if="!pings.length">
                <td colspan="3" style="text-align: center; color: var(--text-muted); padding: 28px">
                  Belum ada data
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
