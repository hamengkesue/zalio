<script setup lang="ts">
  // Dropdown yang bisa dicari + tinggi terbatas. Popup di-teleport ke body
  // agar tidak terpotong container yang punya overflow (mis. modal).
  interface Option { value: string; label: string }
  const model = defineModel<string>({ default: '' })
  const props = defineProps<{
    options: Option[]
    placeholder?: string
    disabled?: boolean
    invalid?: boolean
  }>()
  const emit = defineEmits<{ change: [] }>()

  const open = ref(false)
  const search = ref('')
  const triggerEl = ref<HTMLElement>()
  const popEl = ref<HTMLElement>()
  const searchEl = ref<HTMLInputElement>()
  const activeIndex = ref(0)

  const selectedLabel = computed(() => props.options.find(o => o.value === model.value)?.label ?? '')
  const filtered = computed(() => {
    const q = search.value.trim().toLowerCase()
    return q ? props.options.filter(o => o.label.toLowerCase().includes(q)) : props.options
  })

  const pos = reactive({ left: 0, top: 0, bottom: 0, width: 0, up: false, maxH: 280 })
  function updatePos() {
    const el = triggerEl.value
    if (!el) return
    const r = el.getBoundingClientRect()
    const gap = 4
    const below = window.innerHeight - r.bottom - gap
    const above = r.top - gap
    pos.up = below < 220 && above > below
    pos.left = r.left
    pos.width = r.width
    pos.maxH = Math.min(300, (pos.up ? above : below) - 4)
    if (pos.up) pos.bottom = window.innerHeight - r.top + gap
    else pos.top = r.bottom + gap
  }
  const popStyle = computed(() => ({
    position: 'fixed' as const,
    left: pos.left + 'px',
    width: pos.width + 'px',
    zIndex: 200,
    ...(pos.up ? { bottom: pos.bottom + 'px' } : { top: pos.top + 'px' }),
    // Semua style popup ditaruh inline (bukan file CSS) supaya PASTI terpakai —
    // popup di-teleport & CSS-nya (scoped/global) sempat tidak konsisten antar-browser/HMR.
    boxSizing: 'border-box' as const,
    background: 'var(--bg-surface)',
    border: '1px solid var(--border-color)',
    borderRadius: '12px',
    boxShadow: '0 12px 30px rgba(0, 0, 0, 0.18)',
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column' as const,
  }))
  const searchStyle = {
    boxSizing: 'border-box' as const, minWidth: '0', display: 'flex', alignItems: 'center', gap: '9px',
    padding: '10px 18px', borderBottom: '1px solid var(--border-color)',
  }
  const searchInputStyle = {
    // min-width:0 → input bisa menyusut di dalam flex (default text input ~170px,
    // yang tanpa ini memaksa lebar popup melebar/menggeser isinya di popup sempit).
    flex: '1 1 0%', minWidth: '0', width: '100%', border: 'none', outline: 'none', background: 'transparent',
    fontSize: '14px', color: 'var(--text-primary)', fontFamily: 'var(--font-family)',
  }
  const optBaseStyle = {
    boxSizing: 'border-box' as const, display: 'block', width: '100%', margin: '0',
    textAlign: 'left' as const,
    padding: '9px 14px', border: 'none', borderRadius: '8px',
    fontSize: '14px', cursor: 'pointer', fontFamily: 'var(--font-family)',
    whiteSpace: 'nowrap' as const, overflow: 'hidden', textOverflow: 'ellipsis',
  }
  function optStyle(o: Option, i: number) {
    const selected = o.value === model.value
    return {
      ...optBaseStyle,
      background: selected ? 'var(--accent-light)' : (i === activeIndex.value ? 'var(--bg-hover)' : 'transparent'),
      color: selected ? 'var(--accent)' : 'var(--text-primary)',
      fontWeight: selected ? 700 : 400,
    }
  }

  function openMenu() {
    if (props.disabled) return
    search.value = ''
    // mulai tanpa highlight; keyboard mulai dari item terpilih bila ada
    activeIndex.value = props.options.findIndex(o => o.value === model.value)
    updatePos()
    open.value = true
    nextTick(() => searchEl.value?.focus())
  }
  function closeMenu() { open.value = false }
  function toggle() { open.value ? closeMenu() : openMenu() }
  function pick(o: Option) {
    if (model.value !== o.value) { model.value = o.value; emit('change') }
    closeMenu()
  }

  function onKeydown(e: KeyboardEvent) {
    if (!open.value) return
    if (e.key === 'ArrowDown') { e.preventDefault(); activeIndex.value = Math.min(filtered.value.length - 1, activeIndex.value + 1) }
    else if (e.key === 'ArrowUp') { e.preventDefault(); activeIndex.value = Math.max(0, activeIndex.value - 1) }
    else if (e.key === 'Enter') { e.preventDefault(); const o = filtered.value[activeIndex.value]; if (o) pick(o) }
    else if (e.key === 'Escape') { closeMenu() }
  }
  watch(search, () => { activeIndex.value = -1 })

  function onDocMouseDown(e: MouseEvent) {
    const t = e.target as Node
    if (triggerEl.value?.contains(t) || popEl.value?.contains(t)) return
    closeMenu()
  }
  // Scroll di luar popup (mis. body modal) → tutup, agar posisi tidak pernah "lepas".
  // Scroll di dalam daftar opsi tetap diabaikan.
  function onScroll(e: Event) {
    if (!open.value) return
    if (popEl.value && popEl.value.contains(e.target as Node)) return
    closeMenu()
  }
  function onResize() { if (open.value) closeMenu() }

  onMounted(() => {
    document.addEventListener('mousedown', onDocMouseDown)
    window.addEventListener('resize', onResize)
    window.addEventListener('scroll', onScroll, true)
  })
  onUnmounted(() => {
    document.removeEventListener('mousedown', onDocMouseDown)
    window.removeEventListener('resize', onResize)
    window.removeEventListener('scroll', onScroll, true)
  })
</script>

<template>
  <div class="sls">
    <button
      ref="triggerEl"
      type="button"
      class="text-input sls-trigger"
      :class="{ 'input-error': invalid, 'sls-open': open }"
      :disabled="disabled"
      @click="toggle"
    >
      <span :class="selectedLabel ? 'sls-value' : 'sls-placeholder'">{{ selectedLabel || placeholder }}</span>
      <UIcon name="i-lucide-chevron-down" class="sls-caret" :class="{ up: open }" />
    </button>

    <Teleport to="body">
      <div v-if="open" ref="popEl" class="sls-pop" :style="popStyle" @keydown="onKeydown">
        <div class="sls-search" :style="searchStyle">
          <UIcon name="i-lucide-search" class="sls-search-ic" :style="{ color: 'var(--text-muted)', flex: '0 0 auto' }" />
          <input ref="searchEl" v-model="search" type="text" placeholder="Search..." :style="searchInputStyle" @keydown="onKeydown">
        </div>
        <div class="sls-list" :style="{ boxSizing: 'border-box', maxHeight: pos.maxH + 'px', overflowY: 'auto', overflowX: 'hidden', padding: '6px' }">
          <button
            v-for="(o, i) in filtered"
            :key="o.value"
            type="button"
            class="sls-opt"
            :style="optStyle(o, i)"
            @click="pick(o)"
            @mousemove="activeIndex = i"
          >{{ o.label }}</button>
          <div v-if="!filtered.length" class="sls-empty" :style="{ padding: '12px 14px', fontSize: '13px', color: 'var(--text-muted)', textAlign: 'center' }">No results</div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
  .sls { position: relative; }
  .sls-trigger {
    display: flex; align-items: center; justify-content: space-between; gap: 8px;
    text-align: left; cursor: pointer;
  }
  .sls-trigger:disabled { opacity: 0.7; cursor: not-allowed; background: var(--bg-muted); }
  .sls-value { color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .sls-placeholder { color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .sls-caret { flex: 0 0 auto; color: var(--text-muted); transition: transform 0.15s ease; }
  .sls-caret.up { transform: rotate(180deg); }
  /* Style popup ada di main.css (global) — popup di-teleport ke body, jadi pakai
     CSS global (prefix sls-, tak bentrok) agar pasti termuat & konsisten. */
</style>
