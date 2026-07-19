<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Brands' })

  const { items, total, fetchPage, createBrand, updateBrand, uploadImage, toggleActive } = useBrands()
  const toast = useToast()

  // ── search / sort / infinite-scroll pagination (server-side) ──
  const search = ref('')
  const sortField = ref('name')
  const sortDesc = ref(false)
  const pageSize = 8
  const currentPage = ref(1)
  const loading = ref(false)
  const scrollEl = ref<HTMLElement>()
  const sortOptions = [{ label: 'Name', value: 'name' }]

  // ── filter (Status) ──
  const DEFAULT_STATUS = 'active' // default tabel: hanya brand Active
  const filterStatus = ref(DEFAULT_STATUS) // ''=All, 'active', 'inactive'
  const statusFilterOptions = [{ value: '', label: 'All status' }, { value: 'active', label: 'Active' }, { value: 'inactive', label: 'Inactive' }]
  const filterCount = computed(() => (filterStatus.value !== DEFAULT_STATUS ? 1 : 0))
  function resetFilter() { filterStatus.value = DEFAULT_STATUS }

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
  const ROW_H = 64

  function baseQuery() {
    return { search: search.value.trim(), sort: sortField.value, desc: sortDesc.value, status: filterStatus.value }
  }

  async function loadMore() {
    if (loading.value || items.value.length >= total.value) return
    loading.value = true
    try {
      await fetchPage({ offset: items.value.length, limit: pageSize, ...baseQuery(), append: true })
    } finally {
      loading.value = false
    }
    await fillViewport()
  }

  async function reload() {
    loading.value = true
    items.value = []
    currentPage.value = 1
    try {
      await fetchPage({ offset: 0, limit: pageSize, ...baseQuery(), append: false })
    } finally {
      loading.value = false
    }
    if (scrollEl.value) scrollEl.value.scrollTop = 0
    await fillViewport()
  }

  async function fillViewport() {
    await nextTick()
    const el = scrollEl.value
    if (el && el.scrollHeight <= el.clientHeight + 4 && items.value.length < total.value && !loading.value) {
      await loadMore()
    }
  }

  function onScroll() {
    const el = scrollEl.value
    if (!el) return
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 240) loadMore()
    updateCurrentPage(el)
  }

  function updateCurrentPage(el: HTMLElement) {
    const tp = totalPages.value
    if (el.scrollTop <= 2) { currentPage.value = 1; return }
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 2) { currentPage.value = tp; return }
    const headerH = el.querySelector('thead')?.clientHeight ?? 0
    const centerRow = Math.floor((el.scrollTop + el.clientHeight / 2 - headerH) / ROW_H)
    currentPage.value = Math.min(tp, Math.max(1, Math.floor(Math.max(0, centerRow) / pageSize) + 1))
  }

  let searchTimer: ReturnType<typeof setTimeout> | undefined
  watch(search, () => { clearTimeout(searchTimer); searchTimer = setTimeout(reload, 300) })
  watch([sortField, sortDesc], reload)
  watch(filterStatus, reload)

  // ── form ──
  const showForm = ref(false)
  const editingId = ref<string | null>(null)
  const saving = ref(false)
  const form = reactive({ name: '', description: '', logo: '' })
  const errors = reactive({ name: '' })
  const imgError = reactive<Record<string, boolean>>({})

  const fileInput = ref<HTMLInputElement>()
  const uploading = ref(false)
  const previewUrl = ref('')

  function onAvatarClick() {
    if (!form.logo) fileInput.value?.click()
    else previewUrl.value = API_BASE + '/files/' + form.logo
  }

  async function onFileChange(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
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
      form.logo = await uploadImage(file)
    } catch (err: any) {
      toast.add({ title: 'Upload failed', description: err?.data?.error || 'Could not upload image', color: 'error' })
    } finally {
      uploading.value = false
    }
  }

  onMounted(reload)

  function resetForm() {
    form.name = ''
    form.description = ''
    form.logo = ''
    errors.name = ''
    editingId.value = null
  }

  function openForm() {
    resetForm()
    showForm.value = true
  }

  function openEdit(b: ManagedBrand) {
    resetForm()
    editingId.value = b.id
    form.name = b.name
    form.description = b.description
    form.logo = b.logo
    showForm.value = true
  }

  function validate() {
    errors.name = form.name.trim() ? '' : 'required'
    return !errors.name
  }

  async function submit() {
    if (!validate()) return
    saving.value = true
    try {
      const body = { ...form }
      if (editingId.value) await updateBrand(editingId.value, body)
      else await createBrand(body)
      resetForm()
      showForm.value = false
      await reload()
    } catch (e: any) {
      const field = e?.data?.field
      const msg = e?.data?.error
      if (field && field in errors) {
        ;(errors as Record<string, string>)[field] = msg || 'Already exists'
      } else {
        toast.add({ title: 'Failed', description: msg || 'Could not save brand', color: 'error' })
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
        <h1 class="page-title">Brands</h1>
        <p class="breadcrumbs">
          <span>Products</span>
          <span class="crumb-sep">›</span>
          <span>Brands</span>
        </p>
      </div>

      <div class="toolbar">
        <div class="toolbar-left">
          <SearchSort
            v-model="search"
            v-model:sort="sortField"
            v-model:desc="sortDesc"
            :sort-options="sortOptions"
            placeholder="Search brand name..."
          />
          <AppFilter :active-count="filterCount" width="min(360px, 90vw)" @reset="resetFilter">
            <div>
              <label class="form-label">Status</label>
              <SelectSearch v-model="filterStatus" :options="statusFilterOptions" placeholder="All status" />
            </div>
          </AppFilter>
        </div>
        <button class="btn-primary" @click="openForm">+ Add New</button>
      </div>

      <AppModal v-model="showForm" :title="editingId ? 'Edit Brand' : 'New Brand'" :hide-close="true">
        <form class="modal-form" @submit.prevent="submit">
          <div class="brand-form-grid">
            <div class="span-2">
              <label class="form-label">
                Name <span class="req">*</span>
                <span v-if="errors.name === 'required'" class="label-required">Required</span>
              </label>
              <input
                v-model="form.name"
                class="text-input"
                :class="{ 'input-error': errors.name }"
                placeholder="Enter brand name"
                @input="errors.name = ''"
              >
              <div v-if="errors.name && errors.name !== 'required'" class="field-tip">{{ errors.name }}</div>
            </div>

            <div class="span-2">
              <label class="form-label">Description</label>
              <textarea v-model="form.description" class="text-input" rows="2" placeholder="Optional" />
            </div>

            <div class="span-2">
              <label class="form-label">Logo</label>
              <input ref="fileInput" type="file" accept="image/*" hidden @change="onFileChange">
              <div class="avatar-row">
                <div class="avatar-wrap">
                  <div class="avatar-upload" :class="{ empty: !form.logo }" @click="onAvatarClick">
                    <img v-if="form.logo" :src="`${API_BASE}/files/${form.logo}`" alt="">
                    <UIcon v-else-if="uploading" name="i-lucide-loader-circle" class="avatar-cam spin" />
                    <UIcon v-else name="i-lucide-image" class="avatar-cam" />
                  </div>
                  <button v-if="form.logo" type="button" class="avatar-x" title="Remove image" @click="form.logo = ''">
                    <UIcon name="i-lucide-x" />
                  </button>
                </div>
                <p class="upload-hint">JPG, PNG, WEBP or GIF · <span class="hint-max">max 2 MB</span></p>
              </div>
            </div>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn-ghost" @click="showForm = false">Cancel</button>
            <button class="btn-primary" :disabled="saving" type="submit">{{ saving ? 'Saving...' : 'Save' }}</button>
          </div>
        </form>
      </AppModal>

      <Teleport to="body">
        <Transition name="modal">
          <div v-if="previewUrl" class="img-lightbox" @click="previewUrl = ''">
            <img :src="previewUrl" alt="Logo preview">
          </div>
        </Transition>
      </Teleport>

      <div class="table-card">
        <div ref="scrollEl" class="table-scroll" @scroll="onScroll">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width:56px">Logo</th>
                <th style="min-width:200px">Name</th>
                <th style="min-width:240px">Description</th>
                <th class="text-center" style="width:110px">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="b in items" :key="b.id" class="clickable" @click="openEdit(b)">
                <td class="text-center">
                  <img
                    v-if="b.logo && !imgError[b.id]"
                    :src="`${API_BASE}/files/${b.logo}`"
                    class="avatar-img"
                    alt=""
                    @error="imgError[b.id] = true"
                  >
                  <div v-else class="avatar-fallback"><UIcon name="i-lucide-image" /></div>
                </td>
                <td>{{ b.name }}</td>
                <td>{{ b.description || '—' }}</td>
                <td class="text-center">
                  <button
                    class="toggle"
                    :class="{ on: b.is_active }"
                    :title="b.is_active ? 'Active — click to deactivate' : 'Inactive — click to activate'"
                    @click.stop="toggleActive(b)"
                  >
                    <span class="toggle-knob" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-if="loading && items.length" class="table-loading">
            <UIcon name="i-lucide-loader-circle" class="spin" /> Loading…
          </div>
          <EmptyState v-if="!items.length && !loading" text="No brands found" icon="i-lucide-tag" />
        </div>
        <TablePager :page="currentPage" :total="total" :page-size="pageSize" readonly />
      </div>
    </div>
  </div>
</template>

<style scoped>
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
  .crumb-sep { color: var(--text-muted); opacity: 0.7; }

  .modal-form { display: flex; flex-direction: column; }
  .brand-form-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px;
  }
  .brand-form-grid .span-2 { grid-column: 1 / -1; }
  .brand-form-grid > div { position: relative; }
  textarea.text-input { resize: vertical; }

  .input-error { border-color: var(--danger) !important; }
  .label-required { color: var(--danger); font-size: 12px; font-weight: 700; margin-left: 8px; }
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
  }
  .btn-ghost:hover { background: var(--bg-hover); color: var(--text-primary); }

  /* ── Avatar / logo ── */
  .avatar-img,
  .avatar-fallback {
    width: 34px;
    min-width: 34px;
    height: 34px;
    border-radius: 8px;
  }
  .avatar-img { display: block; margin: 0 auto; object-fit: cover; }
  .avatar-fallback {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-muted);
    color: var(--text-muted);
    font-size: 18px;
  }
  .avatar-wrap { position: relative; flex-shrink: 0; }
  .avatar-upload {
    width: 84px;
    height: 84px;
    border-radius: 12px;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-muted);
    border: 1px solid var(--border-color);
    cursor: pointer;
  }
  .avatar-upload.empty:hover { border-color: var(--accent); }
  .avatar-upload.empty:hover .avatar-cam { color: var(--accent); }
  .avatar-upload img { width: 100%; height: 100%; object-fit: cover; }
  .avatar-cam { font-size: 26px; color: var(--text-muted); }
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
  .spin { animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
  .avatar-row { display: flex; align-items: center; gap: 14px; }
  .upload-hint { font-size: 11px; color: var(--text-muted); line-height: 1.4; }
  .hint-max { color: var(--danger); font-weight: 600; }

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
    border-radius: 16px;
    object-fit: cover;
    box-shadow: 0 30px 60px rgba(0, 0, 0, 0.4);
  }

  .table-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 14px;
    font-size: 13px;
    color: var(--text-muted);
  }

  /* ── Status toggle ── */
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
  .toggle.on { background: var(--success); }
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
  .toggle.on .toggle-knob { left: 21px; }
</style>
