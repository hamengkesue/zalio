<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Internal Users' })

  const { users, total, fetchPage, createUser, updateUser, uploadProfileImage, toggleActive } = useUsers()
  const { isAdmin } = useAuth()
  const toast = useToast()

  // ── search / sort / infinite-scroll pagination (server-side) ──
  const search = ref('')
  const sortField = ref('')
  const sortDesc = ref(false)
  const pageSize = 8
  const currentPage = ref(1)          // halaman yang sedang tampak (indikator, dari posisi scroll)
  const loading = ref(false)          // sedang fetch batch
  const scrollEl = ref<HTMLElement>() // ref ke .table-scroll
  const sortOptions = [
    { label: 'Name', value: 'name' },
    { label: 'Username', value: 'username' },
    { label: 'Role', value: 'role' },
  ]

  // ── filter (Role & Status) ──
  const DEFAULT_STATUS = 'active' // default tabel: hanya user Active
  const filterRole = ref('')                // ''=All, 'admin', 'staff'
  const filterStatus = ref(DEFAULT_STATUS)  // ''=All, 'active', 'inactive'
  const showFilter = ref(false)
  // true kalau filter menyimpang dari default → tampilkan titik indikator di tombol.
  const filterActive = computed(() => filterRole.value !== '' || filterStatus.value !== DEFAULT_STATUS)

  function resetFilter() {
    filterRole.value = ''
    filterStatus.value = DEFAULT_STATUS
  }

  // Param query yang dipakai bersama loadMore & reload (search + sort + filter).
  function baseQuery() {
    return {
      search: search.value.trim(),
      sort: sortField.value,
      desc: sortDesc.value,
      role: filterRole.value,
      status: filterStatus.value,
    }
  }

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
  const ROW_H = 64 // tinggi baris; samakan dengan .data-table tbody tr di main.css

  // Ambil batch berikutnya lalu tambahkan ke bawah (inti infinite scroll).
  async function loadMore() {
    if (loading.value || users.value.length >= total.value) return
    loading.value = true
    try {
      await fetchPage({
        offset: users.value.length,
        limit: pageSize,
        ...baseQuery(),
        append: true,
      })
    } finally {
      loading.value = false
    }
    await fillViewport()
  }

  // Reset daftar & ambil batch pertama (saat mount, ganti search/sort, atau sesudah simpan).
  async function reload() {
    loading.value = true
    users.value = []
    currentPage.value = 1
    try {
      await fetchPage({
        offset: 0,
        limit: pageSize,
        ...baseQuery(),
        append: false,
      })
    } finally {
      loading.value = false
    }
    if (scrollEl.value) scrollEl.value.scrollTop = 0
    await fillViewport()
  }

  // Kalau konten belum memenuhi tinggi container (belum ada scrollbar) padahal
  // data masih ada, ambil lagi supaya user bisa mulai men-scroll.
  async function fillViewport() {
    await nextTick()
    const el = scrollEl.value
    if (el && el.scrollHeight <= el.clientHeight + 4 && users.value.length < total.value && !loading.value) {
      await loadMore()
    }
  }

  // Scroll: (1) fetch saat mendekati bawah, (2) update badge halaman dari posisi scroll.
  function onScroll() {
    const el = scrollEl.value
    if (!el) return
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 240) loadMore()
    updateCurrentPage(el)
  }

  // Badge halaman dari posisi scroll. Dihitung dari baris di TENGAH viewport supaya
  // row-aware; lalu dijamin: paling atas = halaman 1, paling bawah = halaman terakhir
  // (kalau pakai baris teratas, halaman terakhir tak pernah tercapai saat viewport
  // lebih tinggi dari satu halaman).
  function updateCurrentPage(el: HTMLElement) {
    const tp = totalPages.value
    if (el.scrollTop <= 2) { currentPage.value = 1; return }
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 2) { currentPage.value = tp; return }
    const headerH = el.querySelector('thead')?.clientHeight ?? 0
    const centerRow = Math.floor((el.scrollTop + el.clientHeight / 2 - headerH) / ROW_H)
    currentPage.value = Math.min(tp, Math.max(1, Math.floor(Math.max(0, centerRow) / pageSize) + 1))
  }

  // Ganti search/sort → reset dari batch pertama. Search di-debounce 300ms.
  let searchTimer: ReturnType<typeof setTimeout> | undefined
  watch(search, () => {
    clearTimeout(searchTimer)
    searchTimer = setTimeout(reload, 300)
  })
  watch([sortField, sortDesc], reload)
  // Filter live: begitu Role/Status berubah, reset & muat ulang dari server.
  watch([filterRole, filterStatus], reload)

  // ── create form ──
  const showForm = ref(false)
  const editingId = ref<string | null>(null)
  const saving = ref(false)
  const showPassword = ref(false)
  const passwordEditing = ref(false) // hanya relevan di mode edit
  const form = reactive({ name: '', username: '', email: '', whatsapp: '', group_access: '', profile_image: '', password: '', role: 'staff' })

  // Label role untuk tampilan.
  const roleLabel = (r: string) => (r === 'admin' ? 'Administrator' : 'Operator')

  // Group Access = multi-select. Nilai disimpan di form.group_access sebagai
  // teks dipisah koma (mis. "Finance,Warehouse"). Opsi masih placeholder.
  const GROUP_OPTIONS = ['Management', 'Finance', 'Warehouse', 'Sales', 'HR']
  const groupList = () => form.group_access.split(',').map(s => s.trim()).filter(Boolean)
  const hasGroup = (g: string) => groupList().includes(g)
  function toggleGroup(g: string) {
    const set = new Set(groupList())
    set.has(g) ? set.delete(g) : set.add(g)
    form.group_access = [...set].join(',')
  }
  const errors = reactive({ name: '', username: '', email: '', whatsapp: '', password: '' })
  const imgError = reactive<Record<string, boolean>>({})

  // ── upload gambar profil ──
  const fileInput = ref<HTMLInputElement>()
  const uploading = ref(false)
  const previewUrl = ref('') // URL gambar yang sedang di-preview (lightbox)

  // Klik lingkaran avatar: buka filepicker HANYA kalau masih kosong.
  function onAvatarClick() {
    if (!form.profile_image) fileInput.value?.click()
    else previewUrl.value = API_BASE + '/files/' + form.profile_image
  }

  async function onFileChange(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    input.value = '' // reset supaya bisa memilih file yang sama lagi
    if (!file) return
    if (!file.type.startsWith('image/')) {
      toast.add({ title: 'Invalid file', description: 'Please choose an image file', color: 'error' })
      return
    }
    if (file.size > 2 * 1024 * 1024) {
      toast.add({ title: 'File too large', description: 'Maximum size is 2 MB', color: 'error' })
      return
    }
    uploading.value = true
    try {
      form.profile_image = await uploadProfileImage(file)
    } catch (err: any) {
      toast.add({ title: 'Upload failed', description: err?.data?.error || 'Could not upload image', color: 'error' })
    } finally {
      uploading.value = false
    }
  }

  onMounted(reload)

  function resetForm() {
    form.name = ''
    form.username = ''
    form.email = ''
    form.whatsapp = ''
    form.group_access = ''
    form.profile_image = ''
    form.password = ''
    form.role = 'staff'
    errors.name = ''
    errors.username = ''
    errors.email = ''
    errors.whatsapp = ''
    errors.password = ''
    showPassword.value = false
    passwordEditing.value = false
    editingId.value = null
  }

  // Toggle "edit"/"cancel" ganti password (mode edit).
  function togglePasswordEdit() {
    passwordEditing.value = !passwordEditing.value
    form.password = ''
    errors.password = ''
    showPassword.value = false
  }

  function openForm() {
    resetForm()
    showForm.value = true
  }

  // Klik baris → buka modal edit, form terisi data user.
  function openEdit(u: ManagedUser) {
    resetForm()
    editingId.value = u.id
    form.name = u.name
    form.username = u.username
    form.email = u.email
    form.whatsapp = u.whatsapp
    form.group_access = u.group_access
    form.profile_image = u.profile_image
    form.role = u.role
    form.password = ''
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
    // WhatsApp opsional: kosong = valid; kalau diisi harus format E.164.
    errors.whatsapp = !wa
      ? ''
      : (phoneRe.test(wa) ? '' : 'Use a phone number with country code, e.g. +6281234567890')
    // Password wajib saat create, atau saat sedang mengganti password di mode edit.
    // Kalau edit tanpa klik "edit", password lama dipertahankan (tak divalidasi).
    if (!editingId.value || passwordEditing.value) {
      errors.password = !form.password
        ? 'required'
        : (form.password.length >= 6 ? '' : 'Use at least 6 characters')
    } else {
      errors.password = ''
    }
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
      const body = { ...form, whatsapp: form.whatsapp.replace(/[\s-]/g, '') }
      if (editingId.value) await updateUser(editingId.value, body)
      else await createUser(body)
      resetForm()
      showForm.value = false
      await reload() // muat ulang daftar dari batch pertama (server-side)
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
        <!-- Breadcrumbs: teks saja (tidak bisa diklik) -->
        <p class="breadcrumbs">
          <span>Settings</span>
          <span class="crumb-sep">›</span>
          <span>Internal Users</span>
        </p>
      </div>

      <div class="toolbar">
        <div class="toolbar-left">
          <SearchSort
            v-model="search"
            v-model:sort="sortField"
            v-model:desc="sortDesc"
            :sort-options="sortOptions"
            placeholder="Search name, username, email..."
          />
          <button class="filter-btn" :class="{ 'filter-btn--active': filterActive }" title="Filter" @click="showFilter = true">
            <UIcon name="i-lucide-list-filter" />
            <span v-if="filterActive" class="filter-dot" />
          </button>
        </div>
        <button class="btn-primary" @click="openForm">
          + Add New
        </button>
      </div>

      <!-- Modal filter: Role & Status (live, langsung terapkan) -->
      <AppModal v-model="showFilter" title="Filter">
        <div class="filter-modal">
          <div class="filter-group">
            <div class="filter-group-label">Role</div>
            <div class="filter-options">
              <label class="filter-opt"><input v-model="filterRole" type="radio" value=""><span>All</span></label>
              <label class="filter-opt"><input v-model="filterRole" type="radio" value="admin"><span>Administrator</span></label>
              <label class="filter-opt"><input v-model="filterRole" type="radio" value="staff"><span>Operator</span></label>
            </div>
          </div>
          <div class="filter-group">
            <div class="filter-group-label">Status</div>
            <div class="filter-options">
              <label class="filter-opt"><input v-model="filterStatus" type="radio" value=""><span>All</span></label>
              <label class="filter-opt"><input v-model="filterStatus" type="radio" value="active"><span>Active</span></label>
              <label class="filter-opt"><input v-model="filterStatus" type="radio" value="inactive"><span>Inactive</span></label>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button type="button" class="btn-ghost" @click="resetFilter">Reset to default</button>
          <button type="button" class="btn-primary" @click="showFilter = false">Done</button>
        </div>
      </AppModal>

      <AppModal v-model="showForm" :title="editingId ? 'Edit Internal User' : 'New Internal User'" :hide-close="true">
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
                :disabled="!!editingId"
                @input="stripSpaces(); errors.username = ''"
              >
              <div v-if="errors.username && errors.username !== 'required'" class="field-tip">{{ errors.username }}</div>
            </div>

            <div>
              <div class="pw-head">
                <label class="form-label pw-head-label">
                  Password <span v-if="!editingId" class="req">*</span>
                  <span v-if="errors.password === 'required'" class="label-required">Required</span>
                </label>
                <button v-if="editingId" type="button" class="pw-edit-btn" @click="togglePasswordEdit">
                  {{ passwordEditing ? 'cancel' : 'edit' }}
                </button>
              </div>
              <!-- Edit mode & belum klik "edit": password lama disamarkan & terkunci. -->
              <input
                v-if="editingId && !passwordEditing"
                class="text-input"
                type="password"
                value="passwordmask1"
                disabled
              >
              <!-- Create, atau sedang mengganti password: field bisa diisi. -->
              <div v-else class="input-with-icon">
                <input
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  class="text-input"
                  :class="{ 'input-error': errors.password }"
                  :placeholder="editingId ? 'Enter new password' : 'min. 6 characters'"
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
              <label class="form-label">WhatsApp</label>
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
              <select v-model="form.role" class="text-input" :disabled="!isAdmin">
                <option value="admin">Administrator</option>
                <option value="staff">Operator</option>
              </select>
            </div>

            <div>
              <label class="form-label">Group Access</label>
              <div class="group-chips">
                <button
                  v-for="g in GROUP_OPTIONS"
                  :key="g"
                  type="button"
                  class="group-chip"
                  :class="{ 'group-chip--on': hasGroup(g) }"
                  @click="toggleGroup(g)"
                >{{ g }}</button>
              </div>
            </div>

            <div>
              <label class="form-label">Profile Image</label>
              <input ref="fileInput" type="file" accept="image/*" hidden @change="onFileChange">
              <div class="avatar-row">
                <div class="avatar-wrap">
                  <div class="avatar-upload" :class="{ empty: !form.profile_image }" @click="onAvatarClick">
                    <img v-if="form.profile_image" :src="`${API_BASE}/files/${form.profile_image}`" alt="">
                    <UIcon v-else-if="uploading" name="i-lucide-loader-circle" class="avatar-cam spin" />
                    <UIcon v-else name="i-lucide-camera" class="avatar-cam" />
                  </div>
                  <button v-if="form.profile_image" type="button" class="avatar-x" title="Remove image" @click="form.profile_image = ''">
                    <UIcon name="i-lucide-x" />
                  </button>
                </div>
                <p class="upload-hint">JPG, PNG, WEBP or GIF · <span class="hint-max">max 2 MB</span></p>
              </div>
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

      <!-- Lightbox preview gambar profil -->
      <Teleport to="body">
        <Transition name="modal">
          <div v-if="previewUrl" class="img-lightbox" @click="previewUrl = ''">
            <img :src="previewUrl" alt="Profile image preview">
          </div>
        </Transition>
      </Teleport>

      <div class="table-card">
        <div ref="scrollEl" class="table-scroll" @scroll="onScroll">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width:56px">Profile</th>
                <th style="min-width:190px">Full Name</th>
                <th style="min-width:150px">Username</th>
                <th style="min-width:200px">Email</th>
                <th class="shrink-col" style="min-width:180px">WhatsApp</th>
                <th class="shrink-col">Role</th>
                <th class="text-center" style="width:110px">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id" class="clickable" @click="openEdit(u)">
                <td class="text-center">
                  <img
                    v-if="u.profile_image && !imgError[u.id]"
                    :src="`${API_BASE}/files/${u.profile_image}`"
                    class="avatar-img"
                    alt=""
                    @error="imgError[u.id] = true"
                  >
                  <div v-else class="avatar-fallback"><UIcon name="i-lucide-user" /></div>
                </td>
                <td>{{ u.name }}</td>
                <td class="font-semibold">{{ u.username }}</td>
                <td>{{ u.email }}</td>
                <td>{{ u.whatsapp || '—' }}</td>
                <td>{{ roleLabel(u.role) }}</td>
                <td class="text-center">
                  <button
                    class="toggle"
                    :class="{ on: u.is_active }"
                    :title="u.is_active ? 'Active — click to deactivate' : 'Inactive — click to activate'"
                    @click.stop="toggleActive(u)"
                  >
                    <span class="toggle-knob" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-if="loading && users.length" class="table-loading">
            <UIcon name="i-lucide-loader-circle" class="spin" /> Loading…
          </div>
          <EmptyState v-if="!users.length && !loading" text="No users found" icon="i-lucide-users" />
        </div>
        <TablePager :page="currentPage" :total="total" :page-size="pageSize" readonly />
      </div>
    </div>
  </div>
</template>

<style scoped>
  /* Header halaman: breadcrumbs teks + garis batas bawah. */
  .page-header {
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 20px;
  }
  .breadcrumbs {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
    font-size: 13px;
    color: var(--text-muted);
  }
  .crumb-sep {
    color: var(--text-muted);
    opacity: 0.7;
  }

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
  .pw-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }
  .pw-head-label {
    margin-bottom: 0;
  }
  .pw-edit-btn {
    font-size: 12px;
    font-weight: 700;
    color: var(--accent);
    background: none;
    border: none;
    cursor: pointer;
  }
  .pw-edit-btn:hover {
    text-decoration: underline;
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

  /* Kolom yang menciut ke lebar kontennya; sisa ruang tabel
     dialirkan ke kolom lain (Full Name/Email/Username). */
  .shrink-col {
    width: 1%;
    white-space: nowrap;
  }

  /* ── Toolbar: grup kiri (search+sort+filter) & tombol Filter ── */
  .toolbar-left {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: nowrap;
    min-width: 0;
  }
  .filter-btn {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    font-size: 18px;
    cursor: pointer;
    flex-shrink: 0;
    transition: border-color 0.15s ease, color 0.15s ease;
  }
  .filter-btn:hover {
    border-color: var(--accent);
    color: var(--text-primary);
  }
  .filter-btn--active {
    border-color: var(--accent);
    color: var(--accent);
  }
  .filter-dot {
    position: absolute;
    top: 6px;
    right: 6px;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--accent);
  }

  /* ── Modal filter ── */
  .filter-modal {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }
  .filter-group-label {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-secondary);
    margin-bottom: 10px;
  }
  .filter-options {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }
  .filter-opt {
    display: inline-flex;
    align-items: center;
    padding: 8px 16px;
    border: 1px solid var(--border-color);
    border-radius: 999px;
    cursor: pointer;
    font-size: 14px;
    color: var(--text-secondary);
    user-select: none;
    transition: all 0.15s ease;
  }
  .filter-opt:hover {
    border-color: var(--accent);
  }
  .filter-opt input {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
  }
  /* Chip aktif saat radio-nya terpilih. */
  .filter-opt:has(input:checked) {
    border-color: var(--accent);
    background: var(--accent);
    color: #fff;
    font-weight: 600;
  }

  /* ── Group Access multi-select (chip toggle) ── */
  .group-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .group-chip {
    padding: 7px 12px;
    border: 1px solid var(--border-color);
    border-radius: 999px;
    background: var(--bg-surface);
    color: var(--text-secondary);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .group-chip:hover {
    border-color: var(--accent);
  }
  .group-chip--on {
    border-color: var(--accent);
    background: var(--accent);
    color: #fff;
  }

  /* Indikator "sedang memuat batch berikutnya" (infinite scroll). */
  .table-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 14px;
    font-size: 13px;
    color: var(--text-muted);
  }

  /* ── Avatar ── */
  .avatar-img,
  .avatar-fallback {
    width: 34px;
    min-width: 34px;
    height: 34px;
    border-radius: 50%;
  }
  .avatar-img {
    display: block;
    margin: 0 auto;
    object-fit: cover;
  }
  .avatar-fallback {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-muted);
    color: var(--text-muted);
    font-size: 18px;
  }
  /* Avatar + nama dalam satu kolom */
  .user-cell {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .user-cell .avatar-img,
  .user-cell .avatar-fallback {
    flex-shrink: 0;
  }

  /* ── Upload profile image (avatar) ── */
  .avatar-wrap {
    position: relative;
    flex-shrink: 0;
  }
  .avatar-upload {
    width: 84px;
    height: 84px;
    border-radius: 50%;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-muted);
    border: 1px solid var(--border-color);
    cursor: pointer;
  }
  .avatar-upload.empty:hover {
    border-color: var(--accent);
  }
  .avatar-upload.empty:hover .avatar-cam {
    color: var(--accent);
  }
  .avatar-upload img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .avatar-cam {
    font-size: 26px;
    color: var(--text-muted);
  }
  /* Tombol X di pojok kanan atas untuk hapus gambar. */
  .avatar-x {
    position: absolute;
    top: 0;
    right: 0;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--danger);
    color: #fff;
    font-size: 11px;
    border: 2px solid var(--bg-surface);
    cursor: pointer;
  }
  .spin {
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  .avatar-row {
    display: flex;
    align-items: center;
    gap: 14px;
  }
  .upload-hint {
    font-size: 11px;
    color: var(--text-muted);
    line-height: 1.4;
  }
  .hint-max {
    color: var(--danger);
    font-weight: 600;
  }

  /* ── Image preview lightbox ── */
  .img-lightbox {
    position: fixed;
    inset: 0;
    z-index: 200;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 32px;
    background: rgba(15, 23, 42, 0.55);
    backdrop-filter: blur(2px);
    cursor: zoom-out;
  }
  .img-lightbox img {
    width: min(300px, 80vw);
    height: min(300px, 80vw);
    border-radius: 50%;
    object-fit: cover;
    box-shadow: 0 30px 60px rgba(0, 0, 0, 0.4);
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
