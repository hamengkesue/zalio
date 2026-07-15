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

  const fmt = (s: string) => new Date(s).toLocaleString('en-GB')
</script>

<template>
  <div class="page">
    <div class="page-body">
      <div class="page-header">
        <h1 class="page-title">Ping — Sample Vertical Slice</h1>
        <p class="page-subtitle">
          The rows below come from the database via the Go backend. Add a message to
          prove the write → read flow works end to end.
        </p>
      </div>

      <form class="form-row" @submit.prevent="submit">
        <input
          v-model="newMessage"
          class="text-input"
          placeholder="Type a new message, then click Add..."
        >
        <button class="btn-primary" :disabled="saving" type="submit">
          {{ saving ? 'Saving...' : 'Add' }}
        </button>
      </form>

      <div class="table-card">
        <div class="table-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width: 80px">ID</th>
                <th>Message</th>
                <th style="width: 200px">Created</th>
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
                  No data yet
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
