<script setup lang="ts">
  const props = withDefaults(defineProps<{ total: number; pageSize?: number }>(), { pageSize: 9999 })
  const page = defineModel<number>('page', { default: 1 })

  const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
  const from = computed(() => (props.total === 0 ? 0 : (page.value - 1) * props.pageSize + 1))
  const to = computed(() => Math.min(page.value * props.pageSize, props.total))

  const pages = computed(() => {
    const tp = totalPages.value
    const cur = page.value
    const win = 2
    let start = Math.max(1, cur - win)
    let end = Math.min(tp, cur + win)
    if (cur <= win) end = Math.min(tp, 1 + win * 2)
    if (cur > tp - win) start = Math.max(1, tp - win * 2)
    const arr: number[] = []
    for (let p = start; p <= end; p++) arr.push(p)
    return arr
  })

  function go(p: number) {
    if (p >= 1 && p <= totalPages.value) page.value = p
  }
</script>

<template>
  <div class="table-pager">
    <span class="pager-info">Showing {{ from }}-{{ to }} of {{ total }} item{{ total === 1 ? '' : 's' }}</span>
    <div class="pager-btns">
      <button class="pager-btn" :disabled="page <= 1" @click="go(1)"><UIcon name="i-lucide-chevrons-left" /></button>
      <button class="pager-btn" :disabled="page <= 1" @click="go(page - 1)"><UIcon name="i-lucide-chevron-left" /></button>
      <button v-for="p in pages" :key="p" class="pager-btn" :class="{ active: p === page }" @click="go(p)">{{ p }}</button>
      <button class="pager-btn" :disabled="page >= totalPages" @click="go(page + 1)"><UIcon name="i-lucide-chevron-right" /></button>
      <button class="pager-btn" :disabled="page >= totalPages" @click="go(totalPages)"><UIcon name="i-lucide-chevrons-right" /></button>
    </div>
  </div>
</template>
