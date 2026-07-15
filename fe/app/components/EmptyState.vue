<script setup lang="ts">
  withDefaults(defineProps<{ text?: string; icon?: string }>(), {
    text: "Seems like you don't have any data",
    icon: 'i-lucide-package-open',
  })
</script>

<template>
  <div class="empty-state">
    <div class="empty-icon-wrap">
      <UIcon :name="icon" class="empty-icon" />
    </div>
    <p class="empty-text">{{ text }}</p>
  </div>
</template>

<style scoped>
  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 14px;
    padding: 56px 20px;
  }
  .empty-icon-wrap {
    width: 72px;
    height: 72px;
    border-radius: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--bg-muted);
    color: var(--text-muted);
  }
  .empty-icon {
    font-size: 34px;
  }
  .empty-text {
    font-size: 14px;
    color: var(--text-muted);
  }
</style>
