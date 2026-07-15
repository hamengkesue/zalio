<script setup lang="ts">
  defineProps<{ sortOptions?: { label: string; value: string }[]; placeholder?: string }>()
  const search = defineModel<string>({ default: '' })
  const sort = defineModel<string>('sort', { default: '' })
  const desc = defineModel<boolean>('desc', { default: false })

  const open = ref(false)
  const rootEl = ref<HTMLElement>()

  const select = (v: string) => {
    sort.value = v
    open.value = false
  }
  const onDocClick = (e: MouseEvent) => {
    if (rootEl.value && !rootEl.value.contains(e.target as Node)) open.value = false
  }
  onMounted(() => document.addEventListener('click', onDocClick))
  onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<template>
  <div ref="rootEl" class="search-sort">
    <div class="ss-search">
      <UIcon name="i-lucide-search" class="ss-icon" />
      <input v-model="search" :placeholder="placeholder ?? 'Search...'" class="ss-input">
    </div>

    <template v-if="sortOptions?.length">
      <div class="ss-divider" />
      <div class="ss-sort">
        <div class="ss-dropdown">
          <button class="ss-icon-btn" title="Sort by" @click="open = !open">
            <UIcon name="i-lucide-chevron-down" class="text-[16px]" />
          </button>
          <div v-if="open" class="ss-menu">
            <button
              v-for="o in sortOptions"
              :key="o.value"
              class="ss-menu-item"
              :class="{ active: sort === o.value }"
              @click="select(o.value)"
            >
              {{ o.label }}
              <UIcon v-if="sort === o.value" name="i-lucide-check" class="text-[14px]" />
            </button>
          </div>
        </div>
        <button class="ss-icon-btn" :title="desc ? 'Descending' : 'Ascending'" @click="desc = !desc">
          <UIcon :name="desc ? 'i-lucide-arrow-down-wide-narrow' : 'i-lucide-arrow-up-narrow-wide'" class="text-[16px]" />
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
  .search-sort {
    position: relative;
    display: flex;
    align-items: center;
    height: 40px;
    flex: 1 1 320px;
    max-width: 560px;
    border: 1px solid var(--border-color);
    border-radius: 10px;
    background-color: var(--bg-surface);
  }
  .ss-search {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 12px;
    flex: 1;
    min-width: 0;
  }
  .ss-icon {
    font-size: 16px;
    color: var(--text-muted);
    flex-shrink: 0;
  }
  .ss-input {
    border: none;
    outline: none;
    background: transparent;
    font-size: 14px;
    color: var(--text-primary);
    width: 100%;
  }
  .ss-input::placeholder {
    color: var(--text-muted);
  }
  .ss-divider {
    width: 1px;
    align-self: stretch;
    background-color: var(--border-color);
  }
  .ss-sort {
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 0 6px;
    flex-shrink: 0;
  }
  .ss-dropdown {
    position: relative;
    display: flex;
  }
  .ss-icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    color: var(--text-secondary);
    cursor: pointer;
    flex-shrink: 0;
  }
  .ss-icon-btn:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }
  .ss-menu {
    position: absolute;
    top: calc(100% + 6px);
    right: 0;
    min-width: 150px;
    background-color: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 10px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
    padding: 4px;
    z-index: 50;
  }
  .ss-menu-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    width: 100%;
    padding: 8px 10px;
    border-radius: 7px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    background: transparent;
    cursor: pointer;
    white-space: nowrap;
  }
  .ss-menu-item:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }
  .ss-menu-item.active {
    color: var(--accent);
  }
</style>
