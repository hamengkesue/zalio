<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Chart of Account' })

  const { items, total, fetchPage, fetchClassifications, fetchTypes, createAccount, updateAccount, toggleActive } = useCoa()
  const toast = useToast()

  // ── search / sort / infinite-scroll pagination ──
  const search = ref('')
  const sortField = ref('code')
  const sortDesc = ref(false)
  const pageSize = 8
  const currentPage = ref(1)
  const loading = ref(false)
  const scrollEl = ref<HTMLElement>()
  const sortOptions = [
    { label: 'Account Code', value: 'code' },
    { label: 'Account Name', value: 'name' },
    { label: 'Type', value: 'type' },
    { label: 'Classification', value: 'classification' },
    { label: 'Balance', value: 'balance' },
    { label: 'Status', value: 'status' },
  ]

  // ── filter (Classification, Type, Status) ──
  const DEFAULT_STATUS = 'active'
  const filterClassification = ref('')
  const filterType = ref('')
  const filterStatus = ref(DEFAULT_STATUS)
  const statusFilterOptions = [{ value: '', label: 'All status' }, { value: 'active', label: 'Active' }, { value: 'inactive', label: 'Inactive' }]
  const filterCount = computed(() => (filterClassification.value ? 1 : 0) + (filterType.value ? 1 : 0) + (filterStatus.value !== DEFAULT_STATUS ? 1 : 0))
  function resetFilter() { filterClassification.value = ''; filterType.value = ''; filterStatus.value = DEFAULT_STATUS }

  // ── reference data (dropdowns) ──
  const classifications = ref<CoaClassification[]>([])
  const types = ref<CoaType[]>([])
  const classificationFilterOptions = computed(() => [{ value: '', label: 'All classifications' }, ...classifications.value.map(c => ({ value: String(c.classification_code), label: c.classification_name }))])
  const typeSelectOptions = computed(() => types.value.map(t => ({ value: t.account_type_code, label: t.account_type_name })))
  const typeFilterOptions = computed(() => {
    const list = filterClassification.value ? types.value.filter(t => String(t.classification_code) === filterClassification.value) : types.value
    return [{ value: '', label: 'All types' }, ...list.map(t => ({ value: t.account_type_code, label: t.account_type_name }))]
  })
  watch(filterClassification, () => {
    if (filterType.value && !typeFilterOptions.value.some(o => o.value === filterType.value)) filterType.value = ''
  })

  function baseQuery() {
    return {
      search: search.value.trim(), sort: sortField.value, desc: sortDesc.value,
      status: filterStatus.value, account_type: filterType.value, classification: filterClassification.value,
    }
  }

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
  const ROW_H = 64

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
  watch([filterClassification, filterType, filterStatus], reload)

  // ── form ──
  const showForm = ref(false)
  const editingId = ref<string | null>(null)
  const editingCode = ref('')
  const editLocked = ref(false) // true saat edit akun yang sudah dipakai referensi/transaksi
  const saving = ref(false)
  const form = reactive({ account_name: '', account_type_code: '', is_contra: false, opening_balance: 0, opening_date: '', notes: '' })
  const errors = reactive({ account_name: '', account_type_code: '', opening_date: '' })
  const obText = ref('')

  const formatPrice = (n: number) => 'Rp ' + Math.round(n || 0).toLocaleString('id-ID')
  function todayStr() { const d = new Date(); return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}` }

  function onObInput(e: Event) {
    const el = e.target as HTMLInputElement
    const digits = el.value.replace(/\D/g, '')
    form.opening_balance = digits ? Number(digits) : 0
    obText.value = digits ? Number(digits).toLocaleString('id-ID') : ''
    el.value = obText.value
  }

  function resetForm() {
    form.account_name = ''
    form.account_type_code = ''
    form.is_contra = false
    form.opening_balance = 0
    form.opening_date = todayStr()
    form.notes = ''
    obText.value = ''
    editingId.value = null
    editingCode.value = ''
    editLocked.value = false
    errors.account_name = errors.account_type_code = errors.opening_date = ''
  }

  function openForm() { resetForm(); showForm.value = true }

  function openEdit(a: CoaItem) {
    resetForm()
    editingId.value = a.id
    editingCode.value = a.account_code
    editLocked.value = a.in_use
    form.account_name = a.account_name
    form.account_type_code = a.account_type_code
    form.is_contra = a.is_contra
    form.opening_balance = a.opening_balance
    obText.value = a.opening_balance ? a.opening_balance.toLocaleString('id-ID') : ''
    form.opening_date = a.opening_date || todayStr()
    form.notes = a.notes
    showForm.value = true
  }

  function validate() {
    errors.account_name = form.account_name.trim() ? '' : 'required'
    errors.account_type_code = form.account_type_code ? '' : 'required'
    errors.opening_date = form.opening_date ? '' : 'required'
    return !errors.account_name && !errors.account_type_code && !errors.opening_date
  }

  async function submit() {
    if (!validate()) return
    saving.value = true
    try {
      const body = {
        account_name: form.account_name.trim(),
        account_type_code: form.account_type_code,
        is_contra: form.is_contra,
        opening_balance: form.opening_balance,
        opening_date: form.opening_date,
        notes: form.notes,
      }
      if (editingId.value) await updateAccount(editingId.value, body)
      else await createAccount(body)
      resetForm()
      showForm.value = false
      await reload()
    } catch (e: any) {
      const field = e?.data?.field
      const msg = e?.data?.error
      if (field && field in errors) {
        ;(errors as Record<string, string>)[field] = msg || 'Invalid'
      } else {
        toast.add({ title: 'Failed', description: msg || 'Could not save account', color: 'error' })
      }
    } finally {
      saving.value = false
    }
  }

  onMounted(async () => {
    await reload()
    classifications.value = await fetchClassifications()
    types.value = await fetchTypes()
  })
</script>

<template>
  <div class="page">
    <div class="page-body">
      <div class="page-header">
        <h1 class="page-title">Chart of Account</h1>
        <p class="breadcrumbs">
          <span>Finance &amp; Accounting</span>
          <span class="crumb-sep">›</span>
          <span>Chart of Account</span>
        </p>
      </div>

      <div class="toolbar">
        <div class="toolbar-left">
          <SearchSort
            v-model="search"
            v-model:sort="sortField"
            v-model:desc="sortDesc"
            :sort-options="sortOptions"
            placeholder="Search account code or name..."
          />
          <AppFilter :active-count="filterCount" width="min(680px, 92vw)" @reset="resetFilter">
            <div>
              <label class="form-label">Classification</label>
              <SelectSearch v-model="filterClassification" :options="classificationFilterOptions" placeholder="All classifications" />
            </div>
            <div>
              <label class="form-label">Type</label>
              <SelectSearch v-model="filterType" :options="typeFilterOptions" placeholder="All types" />
            </div>
            <div>
              <label class="form-label">Status</label>
              <SelectSearch v-model="filterStatus" :options="statusFilterOptions" placeholder="All status" />
            </div>
          </AppFilter>
        </div>
        <button class="btn-primary" @click="openForm">+ Add New</button>
      </div>

      <AppModal v-model="showForm" :title="editingId ? 'Edit Account' : 'Create New Account'" :hide-close="true">
        <form class="modal-form coa-form" @submit.prevent="submit">
          <div>
            <label class="form-label">Account Code</label>
            <div class="account-code-box">{{ editingId ? editingCode : '[auto generate]' }}</div>
          </div>

          <div>
            <label class="form-label">
              Account Name <span class="req">*</span>
              <span v-if="errors.account_name === 'required'" class="label-required">Required</span>
            </label>
            <input
              v-model="form.account_name"
              class="text-input"
              :class="{ 'input-error': errors.account_name }"
              placeholder="Enter Account Name..."
              :disabled="editLocked"
              @input="errors.account_name = ''"
            >
          </div>

          <div>
            <label class="form-label">
              Account Type <span class="req">*</span>
              <span v-if="errors.account_type_code === 'required'" class="label-required">Required</span>
            </label>
            <SelectSearch
              v-model="form.account_type_code"
              :options="typeSelectOptions"
              placeholder="--Select account type--"
              :invalid="!!errors.account_type_code"
              :disabled="!!editingId"
              @change="errors.account_type_code = ''"
            />
            <div v-if="editingId" class="lock-hint"><UIcon name="i-lucide-lock" /> Account type can't be changed after creation</div>
          </div>

          <div class="contra-block">
            <div class="contra-head">
              <div>
                <div class="contra-title">Contra Account</div>
                <div class="contra-sub">If enabled, the normal balance of this account will be reversed.</div>
              </div>
              <button
                type="button"
                class="toggle"
                :class="{ on: form.is_contra }"
                :disabled="!!editingId"
                :title="editingId ? 'Locked after creation' : (form.is_contra ? 'Contra on' : 'Contra off')"
                @click="editingId ? null : (form.is_contra = !form.is_contra)"
              >
                <span class="toggle-knob" />
              </button>
            </div>
            <div class="contra-example">example: Sales Account (Credit), contra account: Sales Return Account (Debit).</div>
          </div>

          <div class="opening-row">
            <div>
              <label class="form-label">Opening Balance <span class="req">*</span></label>
              <div class="rp-input" :class="{ disabled: editLocked }">
                <span class="rp-prefix">Rp</span>
                <input :value="obText" inputmode="numeric" placeholder="0" :disabled="editLocked" @input="onObInput">
              </div>
            </div>
            <div>
              <label class="form-label">
                Opening Date <span class="req">*</span>
                <span v-if="errors.opening_date === 'required'" class="label-required">Required</span>
              </label>
              <input
                v-model="form.opening_date"
                type="date"
                class="text-input"
                :class="{ 'input-error': errors.opening_date }"
                :disabled="editLocked"
                @input="errors.opening_date = ''"
              >
            </div>
          </div>

          <div v-if="editLocked" class="lock-banner">
            <UIcon name="i-lucide-lock" />
            <span>This account is already in use, so its name, opening balance, and opening date are locked. You can still edit notes and toggle its status.</span>
          </div>

          <div>
            <label class="form-label">Notes</label>
            <textarea v-model="form.notes" class="text-input" rows="2" placeholder="Optional note..." />
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
                <th style="width:120px">Account Code</th>
                <th style="min-width:220px">Account Name</th>
                <th style="min-width:160px">Type</th>
                <th style="min-width:150px">Classification</th>
                <th class="text-right" style="width:140px">Balance</th>
                <th class="text-center" style="width:110px">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="a in items" :key="a.id" class="clickable" @click="openEdit(a)">
                <td class="acc-code">{{ a.account_code }}</td>
                <td>{{ a.account_name }}</td>
                <td>{{ a.account_type_name }}</td>
                <td>{{ a.classification_name }}</td>
                <td class="text-right">{{ formatPrice(a.opening_balance) }}</td>
                <td class="text-center">
                  <button
                    class="toggle"
                    :class="{ on: a.is_active }"
                    :title="a.is_active ? 'Active — click to deactivate' : 'Inactive — click to activate'"
                    @click.stop="toggleActive(a)"
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
          <EmptyState v-if="!items.length && !loading" text="No accounts found" icon="i-lucide-book-open" />
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

  .coa-form { display: flex; flex-direction: column; gap: 16px; }
  textarea.text-input { resize: vertical; }

  .account-code-box {
    padding: 12px 15px;
    border-radius: 12px;
    background: var(--accent-light);
    color: var(--accent);
    font-weight: 700;
    font-size: 14px;
  }

  .lock-hint {
    display: flex;
    align-items: center;
    gap: 5px;
    margin-top: 6px;
    font-size: 12px;
    color: var(--text-muted);
  }
  .lock-banner {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 11px 13px;
    border-radius: 10px;
    background: var(--bg-muted);
    color: var(--text-secondary);
    font-size: 13px;
    line-height: 1.4;
  }
  .lock-banner .iconify { flex-shrink: 0; margin-top: 2px; }

  .input-error { border-color: var(--danger) !important; }
  .label-required { color: var(--danger); font-size: 12px; font-weight: 700; margin-left: 8px; }

  /* Contra block */
  .contra-block {
    border-top: 1px solid var(--border-color);
    border-bottom: 1px solid var(--border-color);
    padding: 16px 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .contra-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
  .contra-title { font-size: 15px; font-weight: 800; color: var(--text-primary); }
  .contra-sub { font-size: 13px; color: var(--text-muted); margin-top: 2px; }
  .contra-example {
    background: var(--warning-light);
    color: var(--warning);
    font-size: 13px;
    font-weight: 600;
    padding: 10px 12px;
    border-radius: 8px;
  }

  /* Opening balance + date */
  .opening-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
  .rp-input {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 15px;
    border: 1px solid var(--border-color);
    border-radius: 12px;
    background: var(--bg-surface);
  }
  .rp-input:focus-within { border-color: var(--border-focus); box-shadow: 0 0 0 3px rgba(0, 112, 242, 0.1); }
  .rp-input.disabled { background: var(--bg-muted); }
  .rp-input.disabled .rp-prefix, .rp-input.disabled input { color: var(--text-muted); cursor: not-allowed; }
  .rp-prefix { color: var(--text-muted); font-size: 14px; font-weight: 600; }
  .rp-input input {
    flex: 1;
    width: 100%;
    border: none;
    outline: none;
    background: transparent;
    padding: 12px 0;
    font-size: 14px;
    color: var(--text-primary);
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 6px;
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

  .acc-code { font-weight: 700; color: var(--text-primary); }

  .table-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 14px;
    font-size: 13px;
    color: var(--text-muted);
  }
  .spin { animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* Status / contra toggle */
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
    flex-shrink: 0;
    transition: background 0.15s ease;
  }
  .toggle.on { background: var(--success); }
  .toggle:disabled { opacity: 0.55; cursor: not-allowed; }
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

  @media (max-width: 560px) {
    .opening-row { grid-template-columns: 1fr; }
  }
</style>
