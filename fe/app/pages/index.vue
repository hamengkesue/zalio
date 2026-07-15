<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Dashboard' })

  const { pings, fetchPings } = usePing()
  const backendOk = ref(false)

  onMounted(async () => {
    try {
      await $fetch(`${API_BASE}/api/v1/health`)
      backendOk.value = true
    } catch {
      backendOk.value = false
    }
    await fetchPings()
  })
</script>

<template>
  <div class="page">
    <div class="page-body">
      <div class="page-header">
        <h1 class="page-title">Dashboard</h1>
        <p class="page-subtitle">Kerangka awal Zalio ERP — Fase 0 (Fondasi &amp; Scaffold).</p>
      </div>

      <div class="stat-grid">
        <div class="stat-card">
          <div class="stat-icon" :class="backendOk ? 'success' : 'accent'">
            <UIcon name="i-lucide-server" />
          </div>
          <div>
            <div class="stat-label">Status Backend (Go)</div>
            <div class="stat-value">{{ backendOk ? 'Terhubung' : 'Terputus' }}</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon purple">
            <UIcon name="i-lucide-database" />
          </div>
          <div>
            <div class="stat-label">Baris di tabel tb_ping</div>
            <div class="stat-value">{{ pings.length }}</div>
          </div>
        </div>
      </div>

      <div class="info-card">
        <p>
          🎉 <strong>Fase 0 selesai.</strong> Kerangka proyek sudah tembus dari
          <strong>database → backend → frontend</strong>. Angka di kartu atas diambil langsung dari backend Go
          (<code>{{ API_BASE }}</code>) yang membaca database PostgreSQL.
        </p>
        <p style="margin-top: 10px">
          Buka menu <strong>Ping (contoh slice)</strong> di kiri untuk mencoba pola tambah-data
          (tulis → baca) yang akan ditiru semua modul ERP berikutnya (Produk, Inventory, Purchasing, dst.).
        </p>
      </div>
    </div>
  </div>
</template>
