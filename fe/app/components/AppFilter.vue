<script setup lang="ts">
  // Tombol filter (ikon saja) + panel dropdown. Isi field lewat <slot>.
  // Dipakai lintas submenu (Products, Internal Users, dst) — desain seragam.
  const props = defineProps<{ activeCount?: number; width?: string }>()
  const emit = defineEmits<{ reset: [] }>()

  const open = ref(false)
  const wrapEl = ref<HTMLElement>()

  // Tutup saat klik di luar — abaikan popup SelectSearch yang di-teleport ke body.
  function onDocClick(e: MouseEvent) {
    if (!open.value) return
    const t = e.target as HTMLElement
    if (wrapEl.value?.contains(t) || t?.closest?.('.sls-pop')) return
    open.value = false
  }
  onMounted(() => document.addEventListener('mousedown', onDocClick))
  onUnmounted(() => document.removeEventListener('mousedown', onDocClick))
</script>

<template>
  <div ref="wrapEl" class="filter-wrap">
    <button type="button" class="filter-btn" :class="{ active: !!activeCount }" title="Filter" @click="open = !open">
      <UIcon name="i-lucide-sliders-horizontal" />
      <span v-if="activeCount" class="filter-count">{{ activeCount }}</span>
    </button>
    <div v-if="open" class="filter-panel" :style="width ? { width } : undefined">
      <div class="filter-panel-head">
        <span>Filter</span>
        <button v-if="activeCount" type="button" class="filter-reset" @click="emit('reset')">Reset</button>
      </div>
      <div class="filter-grid">
        <slot />
      </div>
    </div>
  </div>
</template>

<style scoped>
  .filter-wrap { position: relative; }
  .filter-btn {
    position: relative; display: inline-flex; align-items: center; justify-content: center;
    width: 40px; height: 40px; border: 1px solid var(--border-color); border-radius: 10px;
    background: var(--bg-surface); color: var(--text-secondary); cursor: pointer; font-size: 18px;
    transition: border-color 0.12s, color 0.12s, background 0.12s;
  }
  .filter-btn:hover { color: var(--text-primary); border-color: var(--text-muted); }
  .filter-btn.active { border-color: var(--accent); color: var(--accent); background: var(--accent-light); }
  .filter-count {
    position: absolute; top: -6px; right: -6px;
    display: inline-flex; align-items: center; justify-content: center;
    min-width: 17px; height: 17px; padding: 0 4px; border-radius: 999px;
    background: var(--accent); color: #fff; font-size: 10px; font-weight: 700;
  }
  .filter-panel {
    position: absolute; top: calc(100% + 8px); right: 0; z-index: 60; width: min(560px, 90vw);
    background: var(--bg-surface); border: 1px solid var(--border-color); border-radius: 14px;
    box-shadow: 0 14px 40px rgba(0, 0, 0, 0.16); padding: 16px;
  }
  .filter-panel-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; font-weight: 800; color: var(--text-primary); }
  .filter-reset { border: none; background: transparent; color: var(--accent); font-size: 13px; font-weight: 700; cursor: pointer; }
  .filter-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px 16px; }
  .filter-grid :deep(.form-label) { display: block; margin-bottom: 6px; font-size: 13px; font-weight: 600; color: var(--text-secondary); }
  @media (max-width: 620px) { .filter-grid { grid-template-columns: 1fr; } }
</style>
