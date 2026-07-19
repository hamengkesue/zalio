<script setup lang="ts">
  useHead({ title: 'Zalio ERP — UoM' })

  const { items, total, fetchPage, createUom, updateUom, toggleActive } = useUoms()
  const toast = useToast()

  // ── search / sort / infinite-scroll pagination ──
  const search = ref('')
  const sortField = ref('name')
  const sortDesc = ref(false)
  const pageSize = 8
  const currentPage = ref(1)
  const loading = ref(false)
  const scrollEl = ref<HTMLElement>()
  const sortOptions = [{ label: 'Name', value: 'name' }]

  // ── filter (Status) ──
  const DEFAULT_STATUS = 'active' // default tabel: hanya unit Active
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
  const form = reactive({ name: '', description: '' })
  const errors = reactive({ name: '' })

  onMounted(reload)

  function resetForm() {
    form.name = ''
    form.description = ''
    errors.name = ''
    editingId.value = null
  }

  function openForm() {
    resetForm()
    showForm.value = true
  }

  function openEdit(u: ManagedUom) {
    resetForm()
    editingId.value = u.id
    form.name = u.name
    form.description = u.description
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
      if (editingId.value) await updateUom(editingId.value, body)
      else await createUom(body)
      resetForm()
      showForm.value = false
      await reload()
    } catch (e: any) {
      const field = e?.data?.field
      const msg = e?.data?.error
      if (field && field in errors) {
        ;(errors as Record<string, string>)[field] = msg || 'Already exists'
      } else {
        toast.add({ title: 'Failed', description: msg || 'Could not save unit', color: 'error' })
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
        <h1 class="page-title">UoM</h1>
        <p class="breadcrumbs">
          <span>Products</span>
          <span class="crumb-sep">›</span>
          <span>UoM</span>
        </p>
      </div>

      <div class="toolbar">
        <div class="toolbar-left">
          <SearchSort
            v-model="search"
            v-model:sort="sortField"
            v-model:desc="sortDesc"
            :sort-options="sortOptions"
            placeholder="Search unit name..."
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

      <AppModal v-model="showForm" :title="editingId ? 'Edit Unit' : 'New Unit'" :hide-close="true">
        <form class="modal-form" @submit.prevent="submit">
          <div class="uom-form-grid">
            <div>
              <label class="form-label">
                Name <span class="req">*</span>
                <span v-if="errors.name === 'required'" class="label-required">Required</span>
              </label>
              <input
                v-model="form.name"
                class="text-input"
                :class="{ 'input-error': errors.name }"
                placeholder="e.g. Kilogram"
                @input="errors.name = ''"
              >
              <div v-if="errors.name && errors.name !== 'required'" class="field-tip">{{ errors.name }}</div>
            </div>

            <div>
              <label class="form-label">Description</label>
              <input v-model="form.description" class="text-input" placeholder="e.g. kg">
            </div>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn-ghost" @click="showForm = false">Cancel</button>
            <button class="btn-primary" :disabled="saving" type="submit">{{ saving ? 'Saving...' : 'Save' }}</button>
          </div>
        </form>
      </AppModal>

      <div class="table-card">
        <div ref="scrollEl" class="table-scroll" @scroll="onScroll">
          <table class="data-table">
            <thead>
              <tr>
                <th style="min-width:220px">Name</th>
                <th style="min-width:240px">Description</th>
                <th class="text-center" style="width:110px">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in items" :key="u.id" class="clickable" @click="openEdit(u)">
                <td>{{ u.name }}</td>
                <td>{{ u.description || '—' }}</td>
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
          <div v-if="loading && items.length" class="table-loading">
            <UIcon name="i-lucide-loader-circle" class="spin" /> Loading…
          </div>
          <EmptyState v-if="!items.length && !loading" text="No units found" icon="i-lucide-ruler" />
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
  .uom-form-grid { display: flex; flex-direction: column; gap: 16px; }
  .uom-form-grid > div { position: relative; }

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

  .table-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 14px;
    font-size: 13px;
    color: var(--text-muted);
  }

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
