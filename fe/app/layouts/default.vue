<script setup lang="ts">
  const colorMode = useColorMode()

  function toggleTheme() {
    colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'
  }
</script>

<template>
  <div class="app-shell">
    <header class="app-topbar">
      <NuxtLink to="/" class="brand">
        <img src="/logo.svg" alt="Zalio ERP" class="brand-logo" width="34" height="34">
        <span class="brand-name">Zalio ERP</span>
      </NuxtLink>

      <div class="flex items-center gap-1">
        <button class="icon-btn" title="Toggle theme" @click="toggleTheme">
          <UIcon :name="colorMode.value === 'dark' ? 'i-lucide-sun' : 'i-lucide-moon'" class="text-[18px]" />
        </button>
        <div class="topbar-avatar">
          <UIcon name="i-lucide-user" class="text-[18px]" />
        </div>
      </div>
    </header>

    <div class="app-body">
      <AppSidebar />
      <main class="app-content">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
  .app-shell {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
    background-color: var(--bg-page);
  }

  .app-topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 20px;
    height: 60px;
    background-color: var(--bg-surface);
    border-bottom: 1px solid var(--border-color);
    flex-shrink: 0;
    z-index: 20;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    text-decoration: none;
    flex-shrink: 0;
  }
  .brand-logo {
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    display: block;
    object-fit: contain;
  }
  .brand-name {
    font-size: 17px;
    font-weight: 800;
    color: var(--text-primary);
    white-space: nowrap;
  }

  .icon-btn {
    width: 38px;
    height: 38px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .icon-btn:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }

  .topbar-avatar {
    width: 38px;
    height: 38px;
    border-radius: 999px;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--accent-light);
    color: var(--accent);
    margin-left: 6px;
    flex-shrink: 0;
  }

  .app-body {
    display: flex;
    flex: 1;
    overflow: hidden;
    min-height: 0;
  }

  .app-content {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
</style>
