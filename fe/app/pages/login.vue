<script setup lang="ts">
  definePageMeta({ layout: false })
  useHead({ title: 'Zalio ERP — Sign in' })

  const { login } = useAuth()
  const username = ref('')
  const password = ref('')
  const showPassword = ref(false)
  const loading = ref(false)
  const errors = reactive({ username: '', password: '' })

  function validate() {
    errors.username = username.value.trim() ? '' : 'required'
    errors.password = password.value ? '' : 'required'
    return !errors.username && !errors.password
  }

  async function submit() {
    if (!validate()) return
    loading.value = true
    try {
      await login(username.value, password.value)
      await navigateTo('/')
    } catch (e: any) {
      const field = e?.data?.field
      const msg = e?.data?.error || 'Sign in failed'
      if (field === 'username') errors.username = msg
      else if (field === 'password') errors.password = msg
      else { errors.username = msg; errors.password = msg }
    } finally {
      loading.value = false
    }
  }
</script>

<template>
  <div class="login-page">
    <!-- ── Left: branding ── -->
    <div class="login-brand">
      <div class="brand-badge">
        <img src="/logo.svg" width="52" height="52" alt="Zalio ERP">
        <span class="brand-badge-name">Zalio ERP</span>
      </div>
      <h2 class="brand-tagline">Efficiency in Every Process</h2>
      <p class="brand-subtagline">Connecting Data, Empowering Teams</p>
      <div class="brand-dots">
        <span class="dot" /><span class="dot active" /><span class="dot" />
      </div>
    </div>

    <!-- ── Right: form ── -->
    <div class="login-form-col">
      <form class="login-card" @submit.prevent="submit">
        <h1 class="login-title">Welcome back</h1>
        <p class="login-sub">Sign into your Zalio Account</p>

        <label class="field-label">
          USERNAME
          <span v-if="errors.username === 'required'" class="login-required">Required</span>
        </label>
        <div class="input-group">
          <UIcon name="i-lucide-user" class="input-icon" />
          <input
            v-model="username"
            class="field-input"
            :class="{ 'field-input--err': errors.username }"
            placeholder="Enter your username"
            autocomplete="username"
            @input="errors.username = ''"
          >
          <div v-if="errors.username && errors.username !== 'required'" class="login-tip">{{ errors.username }}</div>
        </div>

        <label class="field-label">
          PASSWORD
          <span v-if="errors.password === 'required'" class="login-required">Required</span>
        </label>
        <div class="input-group">
          <UIcon name="i-lucide-lock" class="input-icon" />
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            class="field-input"
            :class="{ 'field-input--err': errors.password }"
            placeholder="Enter your password"
            autocomplete="current-password"
            @input="errors.password = ''"
          >
          <button
            type="button"
            class="input-eye"
            :title="showPassword ? 'Hide password' : 'Show password'"
            @click="showPassword = !showPassword"
          >
            <UIcon :name="showPassword ? 'i-lucide-eye' : 'i-lucide-eye-off'" />
          </button>
          <div v-if="errors.password && errors.password !== 'required'" class="login-tip">{{ errors.password }}</div>
        </div>

        <button class="signin-btn" :disabled="loading" type="submit">
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </button>

        <p class="login-footer">Zalio ERP — Driving Operational Excellence.</p>
      </form>
    </div>
  </div>
</template>

<style scoped>
  .login-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    color: #fff;
    background-color: #1e478f;
    background-image:
      radial-gradient(1000px 600px at 78% 8%, rgba(90, 160, 255, 0.35), transparent 60%),
      linear-gradient(135deg, #1c407f 0%, #2a5bab 55%, #214d95 100%);
  }

  /* ── Branding (left) ── */
  .login-brand {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    text-align: center;
  }
  .brand-badge {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 20px 28px;
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.045);
    border: 1px solid rgba(255, 255, 255, 0.09);
    backdrop-filter: blur(6px);
    margin-bottom: 30px;
  }
  .brand-badge-name {
    font-size: 28px;
    font-weight: 800;
    letter-spacing: -0.5px;
  }
  .brand-tagline {
    font-size: 26px;
    font-weight: 700;
  }
  .brand-subtagline {
    font-size: 15px;
    color: rgba(255, 255, 255, 0.6);
    margin-top: 6px;
  }
  .brand-dots {
    display: flex;
    gap: 8px;
    margin-top: 26px;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.25);
    transition: all 0.2s ease;
  }
  .dot.active {
    width: 22px;
    background: #2f8bff;
  }

  /* ── Form (right) ── */
  .login-form-col {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px;
  }
  .login-card {
    width: 100%;
    max-width: 420px;
    display: flex;
    flex-direction: column;
    background: #f6f7f9;
    border-radius: 22px;
    padding: 40px;
    color: #13141b;
    box-shadow: 0 30px 60px rgba(0, 0, 0, 0.35);
  }
  .login-title {
    font-size: 30px;
    font-weight: 800;
    color: #0f1830;
  }
  .login-sub {
    font-size: 14px;
    color: #64748b;
    margin-top: 6px;
    margin-bottom: 24px;
  }
  .field-label {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.06em;
    color: #64748b;
    margin-top: 16px;
    margin-bottom: 8px;
  }
  .input-group {
    position: relative;
    display: flex;
    align-items: center;
  }
  .input-icon {
    position: absolute;
    left: 14px;
    font-size: 18px;
    color: #94a3b8;
    pointer-events: none;
  }
  .field-input {
    width: 100%;
    padding: 13px 44px;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    background: #fff;
    font-size: 14px;
    color: #13141b;
    outline: none;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
  }
  .field-input:focus {
    border-color: #0070f2;
    box-shadow: 0 0 0 3px rgba(0, 112, 242, 0.12);
  }
  .field-input--err {
    border-color: #dc2626 !important;
  }
  .login-required {
    margin-left: 8px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0;
    color: #dc2626;
  }
  .input-eye {
    position: absolute;
    right: 12px;
    display: flex;
    align-items: center;
    background: none;
    border: none;
    cursor: pointer;
    color: #94a3b8;
    font-size: 18px;
  }
  .input-eye:hover {
    color: #64748b;
  }
  .signin-btn {
    margin-top: 26px;
    padding: 14px;
    border-radius: 12px;
    background: #1a7fff;
    color: #fff;
    font-size: 15px;
    font-weight: 700;
    border: none;
    cursor: pointer;
    transition: background 0.15s ease;
  }
  .signin-btn:hover:not(:disabled) {
    background: #0f6fef;
  }
  .signin-btn:disabled {
    opacity: 0.6;
    cursor: default;
  }
  /* Tooltip error autentikasi: melayang di bawah field password. */
  .login-tip {
    position: absolute;
    top: calc(100% + 8px);
    left: 0;
    z-index: 20;
    max-width: 100%;
    background: #dc2626;
    color: #fff;
    font-size: 12px;
    font-weight: 600;
    line-height: 1.35;
    padding: 8px 12px;
    border-radius: 8px;
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.22);
    pointer-events: none;
  }
  .login-tip::before {
    content: '';
    position: absolute;
    bottom: 100%;
    left: 18px;
    border: 5px solid transparent;
    border-bottom-color: #dc2626;
  }
  .login-footer {
    text-align: center;
    margin-top: 24px;
    font-size: 12px;
    color: #94a3b8;
  }

  /* ── Responsive ── */
  @media (max-width: 860px) {
    .login-page {
      flex-direction: column;
    }
    .login-brand {
      padding: 40px 20px 8px;
    }
    .brand-badge {
      margin-bottom: 16px;
    }
    .brand-tagline {
      font-size: 20px;
    }
    .brand-dots {
      margin-top: 14px;
    }
  }
</style>
