<script setup lang="ts">
  definePageMeta({ layout: false })
  useHead({ title: 'Zalio ERP — Masuk' })

  const { login } = useAuth()
  const email = ref('admin@zalio.local')
  const password = ref('')
  const loading = ref(false)
  const error = ref('')

  async function submit() {
    error.value = ''
    loading.value = true
    try {
      await login(email.value, password.value)
      await navigateTo('/')
    } catch (e: any) {
      error.value = e?.data?.error || 'Gagal masuk. Periksa email & password.'
    } finally {
      loading.value = false
    }
  }
</script>

<template>
  <div class="login-wrap">
    <form class="login-card" @submit.prevent="submit">
      <img src="/logo.svg" class="login-logo" width="48" height="48" alt="Zalio ERP">
      <h1 class="login-title">Zalio ERP</h1>
      <p class="login-sub">Masuk ke back-office</p>

      <label class="login-label">Email</label>
      <input v-model="email" type="email" class="text-input" placeholder="email@contoh.com" autocomplete="username">

      <label class="login-label">Password</label>
      <input v-model="password" type="password" class="text-input" placeholder="••••••••" autocomplete="current-password">

      <p v-if="error" class="login-error">{{ error }}</p>

      <button class="btn-primary login-btn" :disabled="loading" type="submit">
        {{ loading ? 'Memproses...' : 'Masuk' }}
      </button>

      <p class="login-hint">Demo: <strong>admin@zalio.local</strong> / <strong>admin123</strong></p>
    </form>
  </div>
</template>

<style scoped>
  .login-wrap {
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--bg-page);
    padding: 20px;
  }
  .login-card {
    width: 100%;
    max-width: 380px;
    display: flex;
    flex-direction: column;
    background-color: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 16px;
    padding: 32px;
  }
  .login-logo {
    align-self: center;
  }
  .login-title {
    text-align: center;
    font-size: 22px;
    font-weight: 800;
    color: var(--text-primary);
    margin-top: 12px;
  }
  .login-sub {
    text-align: center;
    font-size: 14px;
    color: var(--text-secondary);
    margin-bottom: 20px;
  }
  .login-label {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-secondary);
    margin-bottom: 6px;
    margin-top: 12px;
  }
  .login-error {
    margin-top: 12px;
    padding: 10px 12px;
    border-radius: 8px;
    background-color: var(--danger-light);
    color: var(--danger);
    font-size: 13px;
    font-weight: 600;
  }
  .login-btn {
    margin-top: 20px;
    width: 100%;
  }
  .login-hint {
    text-align: center;
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 16px;
  }
</style>
