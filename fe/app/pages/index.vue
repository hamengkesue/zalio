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
        <p class="page-subtitle">Zalio ERP starter shell — Phase 0 (Foundation &amp; Scaffold).</p>
      </div>

      <div class="stat-grid">
        <div class="stat-card">
          <div class="stat-icon" :class="backendOk ? 'success' : 'accent'">
            <UIcon name="i-lucide-server" />
          </div>
          <div>
            <div class="stat-label">Backend status (Go)</div>
            <div class="stat-value">{{ backendOk ? 'Connected' : 'Disconnected' }}</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon purple">
            <UIcon name="i-lucide-database" />
          </div>
          <div>
            <div class="stat-label">Rows in tb_ping table</div>
            <div class="stat-value">{{ pings.length }}</div>
          </div>
        </div>
      </div>

      <div class="info-card">
        <p>
          🎉 <strong>Phase 0 complete.</strong> The project shell is wired end to end from
          <strong>database → backend → frontend</strong>. The numbers in the cards above come straight from the Go backend
          (<code>{{ API_BASE }}</code>) reading the PostgreSQL database.
        </p>
        <p style="margin-top: 10px">
          Open the <strong>Ping (sample slice)</strong> menu on the left to try the write → read pattern
          that every next ERP module will follow (Products, Inventory, Purchasing, etc.).
        </p>
      </div>
    </div>
  </div>
</template>
