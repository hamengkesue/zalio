<script setup lang="ts">
  const open = defineModel<boolean>({ default: false })
  defineProps<{ title?: string; hideClose?: boolean }>()

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') open.value = false
  }
  onMounted(() => window.addEventListener('keydown', onKey))
  onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="open" class="modal-overlay" @click.self="open = false">
        <div class="modal-panel">
          <div class="modal-head">
            <h2 class="modal-title">{{ title }}</h2>
            <button v-if="!hideClose" class="modal-close" title="Close" @click="open = false">
              <UIcon name="i-lucide-x" />
            </button>
          </div>
          <div class="modal-body">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
  .modal-overlay {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    background: rgba(15, 23, 42, 0.55);
    backdrop-filter: blur(2px);
  }
  .modal-panel {
    width: 100%;
    max-width: 520px;
    max-height: calc(100vh - 40px);
    overflow-y: auto;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 18px;
    padding: 24px;
    box-shadow: 0 30px 60px rgba(0, 0, 0, 0.3);
  }
  .modal-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 16px;
    margin-bottom: 18px;
    border-bottom: 1px solid var(--border-color);
  }
  .modal-title {
    font-size: 20px;
    font-weight: 800;
    color: var(--text-primary);
  }
  .modal-close {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    border-radius: 9px;
    color: var(--text-secondary);
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 18px;
  }
  .modal-close:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .modal-enter-active,
  .modal-leave-active {
    transition: opacity 0.18s ease;
  }
  .modal-enter-from,
  .modal-leave-to {
    opacity: 0;
  }
  .modal-enter-active .modal-panel,
  .modal-leave-active .modal-panel {
    transition: transform 0.18s ease;
  }
  .modal-enter-from .modal-panel,
  .modal-leave-to .modal-panel {
    transform: translateY(8px) scale(0.98);
  }
</style>
