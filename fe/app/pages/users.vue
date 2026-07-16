<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Internal Users' })

  const { users, fetchUsers, createUser, toggleActive } = useUsers()
  const toast = useToast()

  // ── search / sort / pagination ──
  const search = ref('')
  const sortField = ref('')
  const sortDesc = ref(false)
  const page = ref(1)
  const pageSize = 8
  const sortOptions = [
    { label: 'Name', value: 'name' },
    { label: 'Username', value: 'username' },
    { label: 'Role', value: 'role' },
  ]

  const filtered = computed(() => {
    let list = users.value
    if (search.value.trim()) {
      const q = search.value.toLowerCase()
      list = list.filter(u =>
        u.name.toLowerCase().includes(q)
        || u.username.toLowerCase().includes(q)
        || u.email.toLowerCase().includes(q))
    }
    if (sortField.value) {
      const k = sortField.value
      list = [...list].sort((a, b) => {
        const cmp = String((a as any)[k]).localeCompare(String((b as any)[k]))
        return sortDesc.value ? -cmp : cmp
      })
    }
    return list
  })

  const paged = computed(() => {
    const start = (page.value - 1) * pageSize
    return filtered.value.slice(start, start + pageSize)
  })

  watch([search, sortField, sortDesc], () => { page.value = 1 })

  // ── create form ──
  const showForm = ref(false)
  const saving = ref(false)
  const showPassword = ref(false)
  const form = reactive({ name: '', username: '', email: '', whatsapp: '', password: '', role: 'staff' })
  const errors = reactive({ name: '', username: '', email: '', whatsapp: '', password: '' })

  onMounted(fetchUsers)

  function resetForm() {
    form.name = ''
    form.username = ''
    form.email = ''
    form.whatsapp = ''
    form.password = ''
    form.role = 'staff'
    errors.name = ''
    errors.username = ''
    errors.email = ''
    errors.whatsapp = ''
    errors.password = ''
    showPassword.value = false
  }

  function openForm() {
    resetForm()
    showForm.value = true
  }

  // Full Name -> Proper Case saat berpindah field
  function properCaseName() {
    const v = form.name.trim().replace(/\s+/g, ' ')
    if (v) form.name = v.replace(/\S+/g, w => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase())
  }
  // Username tidak boleh ada spasi
  function stripSpaces() {
    form.username = form.username.replace(/\s+/g, '')
  }

  const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  const phoneRe = /^\+[1-9]\d{7,14}$/ // + kode negara lalu angka (E.164)

  // '' = valid, 'required' = wajib tapi kosong (tampil di kanan label),
  // teks lain = error format (tampil di bawah field).
  function validate() {
    errors.name = form.name.trim() ? '' : 'required'
    errors.username = form.username.trim() ? '' : 'required'
    errors.email = !form.email.trim()
      ? 'required'
      : (emailRe.test(form.email.trim()) ? '' : 'Invalid email address')
    const wa = form.whatsapp.replace(/[\s-]/g, '')
    errors.whatsapp = !wa
      ? 'required'
      : (phoneRe.test(wa) ? '' : 'Use a phone number with country code, e.g. +6281234567890')
    errors.password = !form.password
      ? 'required'
      : (form.password.length >= 6 ? '' : 'Use at least 6 characters')
    return !errors.name && !errors.username && !errors.email && !errors.whatsapp && !errors.password
  }

  // Validasi langsung saat meninggalkan field (blur). Kalau kosong, error dibersihkan
  // (belum dinilai wajib sampai Save); kalau ada isi tapi tidak valid → tampilkan pesan.
  function validateEmailField() {
    const v = form.email.trim()
    errors.email = !v ? '' : (emailRe.test(v) ? '' : 'Invalid email address')
  }
  function validateWhatsappField() {
    const wa = form.whatsapp.replace(/[\s-]/g, '')
    errors.whatsapp = !wa ? '' : (phoneRe.test(wa) ? '' : 'Use a phone number with country code, e.g. +6281234567890')
  }
  function validatePasswordField() {
    errors.password = !form.password ? '' : (form.password.length >= 6 ? '' : 'Use at least 6 characters')
  }

  async function submit() {
    if (!validate()) return
    saving.value = true
    try {
      await createUser({ ...form, whatsapp: form.whatsapp.replace(/[\s-]/g, '') })
      resetForm()
      showForm.value = false
    } catch (e: any) {
      const field = e?.data?.field
      const msg = e?.data?.error
      if (field && field in errors) {
        // duplikat username/email → tampilkan inline di field-nya
        ;(errors as Record<string, string>)[field] = msg || 'Already taken'
      } else {
        toast.add({ title: 'Failed', description: msg || 'Could not create user', color: 'error' })
      }
    } finally {
      saving.value = false
    }
  }
</script>

<template>
  <div class="page">
    <div class="page-body">
      <div class="page-header">
        <h1 class="page-title">Internal Users</h1>
        <p class="page-subtitle">Manage internal back-office users &amp; their access.</p>
      </div>

      <div class="toolbar">
        <SearchSort
          v-model="search"
          v-model:sort="sortField"
          v-model:desc="sortDesc"
          :sort-options="sortOptions"
          placeholder="Search name, username, email..."
        />
        <button class="btn-primary" @click="openForm">
          + Add New
        </button>
      </div>

      <AppModal v-model="showForm" title="New Internal User" :hide-close="true">
        <form class="modal-form" @submit.prevent="submit">
          <div class="user-form-grid">
            <div class="span-2">
              <label class="form-label">
                Full Name <span class="req">*</span>
                <span v-if="errors.name === 'required'" class="label-required">Required</span>
              </label>
              <input
                v-model="form.name"
                class="text-input"
                :class="{ 'input-error': errors.name }"
                placeholder="Enter full name"
                @blur="properCaseName"
                @input="errors.name = ''"
              >
            </div>

            <div>
              <label class="form-label">
                Username <span class="req">*</span>
                <span v-if="errors.username === 'required'" class="label-required">Required</span>
              </label>
              <input
                v-model="form.username"
                class="text-input"
                :class="{ 'input-error': errors.username }"
                placeholder="e.g. johndoe"
                @input="stripSpaces(); errors.username = ''"
              >
              <div v-if="errors.username && errors.username !== 'required'" class="field-tip">{{ errors.username }}</div>
            </div>

            <div>
              <label class="form-label">
                Password <span class="req">*</span>
                <span v-if="errors.password === 'required'" class="label-required">Required</span>
              </label>
              <div class="input-with-icon">
                <input
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  class="text-input"
                  :class="{ 'input-error': errors.password }"
                  placeholder="min. 6 characters"
                  @input="errors.password = ''"
                  @blur="validatePasswordField"
                >
                <button type="button" class="input-eye" :title="showPassword ? 'Hide' : 'Show'" @click="showPassword = !showPassword">
                  <UIcon :name="showPassword ? 'i-lucide-eye' : 'i-lucide-eye-off'" />
                </button>
              </div>
              <div v-if="errors.password && errors.password !== 'required'" class="field-tip">{{ errors.password }}</div>
            </div>

            <div>
              <label class="form-label">
                Email <span class="req">*</span>
                <span v-if="errors.email === 'required'" class="label-required">Required</span>
              </label>
              <input
                v-model="form.email"
                type="text"
                inputmode="email"
                class="text-input"
                :class="{ 'input-error': errors.email }"
                placeholder="email@example.com"
                @input="errors.email = ''"
                @blur="validateEmailField"
              >
              <div v-if="errors.email && errors.email !== 'required'" class="field-tip">{{ errors.email }}</div>
            </div>

            <div>
              <label class="form-label">
                WhatsApp <span class="req">*</span>
                <span v-if="errors.whatsapp === 'required'" class="label-required">Required</span>
              </label>
              <input
                v-model="form.whatsapp"
                class="text-input"
                :class="{ 'input-error': errors.whatsapp }"
                placeholder="+6281234567890"
                @input="errors.whatsapp = ''"
                @blur="validateWhatsappField"
              >
              <div v-if="errors.whatsapp && errors.whatsapp !== 'required'" class="field-tip">{{ errors.whatsapp }}</div>
            </div>

            <div>
              <label class="form-label">Role</label>
              <select v-model="form.role" class="text-input">
                <option value="staff">staff</option>
                <option value="admin">admin</option>
              </select>
            </div>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn-ghost" @click="showForm = false">Cancel</button>
            <button class="btn-primary" :disabled="saving" type="submit">
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </form>
      </AppModal>

      <div class="table-card">
        <div class="table-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th>Full Name</th>
                <th>Username</th>
                <th>Email</th>
                <th>Role</th>
                <th class="text-center" style="width:110px">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in paged" :key="u.id">
                <td>{{ u.name }}</td>
                <td class="font-semibold">{{ u.username }}</td>
                <td>{{ u.email }}</td>
                <td><span class="badge" :class="u.role === 'admin' ? 'badge-accent' : 'badge-muted'">{{ u.role }}</span></td>
                <td class="text-center">
                  <button
                    class="toggle"
                    :class="{ on: u.is_active }"
                    :title="u.is_active ? 'Active — click to deactivate' : 'Inactive — click to activate'"
                    @click="toggleActive(u)"
                  >
                    <span class="toggle-knob" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <EmptyState v-if="!filtered.length" text="No users found" icon="i-lucide-users" />
        </div>
        <TablePager v-model:page="page" :total="filtered.length" :page-size="pageSize" />
      </div>
    </div>
  </div>
</template>

<style scoped>
  .modal-form {
    display: flex;
    flex-direction: column;
  }
  .user-form-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px 16px;
  }
  .user-form-grid .span-2 {
    grid-column: 1 / -1;
  }
  @media (max-width: 460px) {
    .user-form-grid {
      grid-template-columns: 1fr;
    }
  }

  .input-with-icon {
    position: relative;
    display: flex;
    align-items: center;
  }
  .input-with-icon .text-input {
    padding-right: 42px;
  }
  .input-eye {
    position: absolute;
    right: 10px;
    display: flex;
    align-items: center;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    font-size: 18px;
  }
  .input-eye:hover {
    color: var(--text-secondary);
  }

  .input-error {
    border-color: var(--danger) !important;
  }
  /* Tiap sel field jadi anchor untuk tooltip error. */
  .user-form-grid > div {
    position: relative;
  }
  /* Tooltip error: melayang (absolute) di bawah field, TIDAK menggeser field lain. */
  .field-tip {
    position: absolute;
    top: calc(100% + 8px);
    left: 0;
    z-index: 30;
    max-width: 260px;
    background: var(--danger);
    color: #fff;
    font-size: 12px;
    font-weight: 600;
    line-height: 1.35;
    padding: 7px 10px;
    border-radius: 8px;
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.2);
  }
  .field-tip::before {
    content: '';
    position: absolute;
    bottom: 100%;
    left: 16px;
    border: 5px solid transparent;
    border-bottom-color: var(--danger);
  }
  .label-required {
    color: var(--danger);
    font-size: 12px;
    font-weight: 700;
    margin-left: 8px;
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 22px;
    padding-top: 18px;
    border-top: 1px solid var(--border-color);
  }
  .btn-ghost {
    padding: 10px 18px;
    border-radius: 10px;
    background: var(--bg-muted);
    color: var(--text-secondary);
    font-size: 14px;
    font-weight: 700;
    border: none;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease;
  }
  .btn-ghost:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  /* ── Status on/off toggle ── */
  .toggle {
    display: inline-flex;
    align-items: center;
    width: 42px;
    height: 24px;
    border-radius: 999px;
    background: var(--bg-muted);
    border: none;
    cursor: pointer;
    padding: 0;
    position: relative;
    transition: background 0.15s ease;
  }
  .toggle.on {
    background: var(--success);
  }
  .toggle-knob {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #fff;
    transition: left 0.15s ease;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
  }
  .toggle.on .toggle-knob {
    left: 21px;
  }
</style>
