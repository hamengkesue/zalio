<script setup lang="ts">
  const route = useRoute()
  const { collapsed, toggle } = useSidebar()
  const { isAdmin } = useAuth()

  // Menu utama (atas).
  const topGroups = [
    {
      label: '',
      items: [
        { label: 'Dashboard', icon: 'i-lucide-layout-dashboard', to: '/' },
      ],
    },
    {
      label: 'Example',
      items: [
        { label: 'Ping (sample slice)', icon: 'i-lucide-radio', to: '/ping' },
      ],
    },
  ]

  // Isi submenu Settings (tampil di popover di atas tombol Settings).
  const settingsItems = [
    { label: 'Internal Users', icon: 'i-lucide-users', to: '/settings/internal-users' },
  ]
  const isSettingsActive = computed(() => settingsItems.some(i => i.to === route.path))

  // ── Flyout submenu Settings (muncul saat hover, melayang ke kanan) ──
  const btnRef = ref<HTMLElement>()
  const settingsOpen = ref(false)
  const popStyle = ref<Record<string, string>>({})
  let hideTimer: ReturnType<typeof setTimeout> | null = null

  function showSettings() {
    if (hideTimer) { clearTimeout(hideTimer); hideTimer = null }
    const el = btnRef.value
    if (el) {
      const r = el.getBoundingClientRect()
      popStyle.value = {
        left: `${r.right + 10}px`,
        bottom: `${window.innerHeight - r.bottom}px`,
        minWidth: '200px',
      }
    }
    settingsOpen.value = true
  }
  function hideSettingsSoon() {
    if (hideTimer) clearTimeout(hideTimer)
    hideTimer = setTimeout(() => { settingsOpen.value = false }, 120)
  }
</script>

<template>
  <aside class="app-sidebar" :class="{ collapsed }">
    <!-- ── Menu utama ── -->
    <nav class="sidebar-nav">
      <div v-for="(group, gi) in topGroups" :key="gi" class="nav-group">
        <span v-if="!collapsed && group.label" class="nav-section">{{ group.label }}</span>
        <div v-else-if="collapsed && gi > 0" class="nav-divider" />
        <NuxtLink
          v-for="item in group.items"
          :key="item.to"
          :to="item.to"
          class="sidebar-item"
          :class="{ active: route.path === item.to }"
          :title="collapsed ? item.label : undefined"
        >
          <UIcon :name="item.icon" class="sidebar-icon" />
          <span v-if="!collapsed" class="sidebar-label">{{ item.label }}</span>
        </NuxtLink>
      </div>
    </nav>

    <!-- ── Settings (submenu muncul saat hover) ── -->
    <div v-if="isAdmin" class="sidebar-settings" @mouseenter="showSettings" @mouseleave="hideSettingsSoon">
      <button
        ref="btnRef"
        type="button"
        class="sidebar-item settings-toggle"
        :class="{ active: isSettingsActive || settingsOpen }"
        :title="collapsed ? 'Settings' : undefined"
      >
        <UIcon name="i-lucide-settings" class="sidebar-icon" />
        <template v-if="!collapsed">
          <span class="sidebar-label">Settings</span>
          <UIcon name="i-lucide-chevron-right" class="settings-caret" />
        </template>
      </button>
    </div>

    <!-- ── Collapse toggle ── -->
    <button class="collapse-btn" :title="collapsed ? 'Expand' : 'Collapse'" @click="toggle">
      <UIcon :name="collapsed ? 'i-lucide-chevrons-right' : 'i-lucide-chevrons-left'" class="text-[18px]" />
      <span v-if="!collapsed" class="sidebar-label">Collapse</span>
    </button>

    <!-- ── Flyout submenu (melayang ke kanan tombol Settings) ── -->
    <Teleport to="body">
      <Transition name="pop">
        <div
          v-if="settingsOpen"
          class="settings-popover"
          :style="popStyle"
          @mouseenter="showSettings"
          @mouseleave="hideSettingsSoon"
        >
          <span class="settings-popover-title">Settings</span>
          <NuxtLink
            v-for="it in settingsItems"
            :key="it.to"
            :to="it.to"
            class="settings-pop-item"
            :class="{ active: route.path === it.to }"
            @click="settingsOpen = false"
          >
            <UIcon :name="it.icon" class="settings-pop-icon" />
            <span>{{ it.label }}</span>
          </NuxtLink>
        </div>
      </Transition>
    </Teleport>
  </aside>
</template>

<style scoped>
  .app-sidebar {
    width: 240px;
    flex-shrink: 0;
    background-color: var(--bg-surface);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    transition: width 0.2s ease;
    overflow: hidden;
  }
  .app-sidebar.collapsed {
    width: 72px;
  }

  .sidebar-nav {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 12px 8px;
    flex: 1;
    overflow-y: auto;
  }
  .nav-group {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .nav-group + .nav-group {
    margin-top: 10px;
  }
  .nav-section {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
    padding: 8px 12px 4px;
  }
  .nav-divider {
    height: 1px;
    background-color: var(--border-color);
    margin: 8px 12px;
  }

  .sidebar-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    text-decoration: none;
    transition: background-color 0.15s ease, color 0.15s ease;
    white-space: nowrap;
  }
  .app-sidebar.collapsed .sidebar-item {
    justify-content: center;
    padding: 10px 0;
  }
  .sidebar-item:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }
  .sidebar-item.active {
    background-color: var(--accent-light);
    color: var(--accent);
  }

  .sidebar-icon {
    font-size: 18px;
    flex-shrink: 0;
  }
  .sidebar-label {
    line-height: 1;
  }

  /* ── Settings section (bottom) ── */
  .sidebar-settings {
    padding: 8px;
    border-top: 1px solid var(--border-color);
    flex-shrink: 0;
  }
  .settings-toggle {
    width: 100%;
    background: transparent;
    border: none;
    cursor: pointer;
    font-family: inherit;
  }
  .settings-caret {
    margin-left: auto;
    font-size: 15px;
    flex-shrink: 0;
    transition: transform 0.15s ease;
  }
  .settings-caret.flip {
    transform: rotate(180deg);
  }

  /* ── Collapse toggle ── */
  .collapse-btn {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px 20px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    cursor: pointer;
    transition: background-color 0.15s ease, color 0.15s ease;
    white-space: nowrap;
    border-top: 1px solid var(--border-color);
    flex-shrink: 0;
  }
  .app-sidebar.collapsed .collapse-btn {
    justify-content: center;
    padding: 14px 0;
  }
  .collapse-btn:hover {
    color: var(--accent);
    background-color: var(--bg-hover);
  }

  /* ── Settings popover (floating above the button) ── */
  .settings-popover {
    position: fixed;
    z-index: 200;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.16);
    padding: 6px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  /* caret menunjuk ke kiri (ke tombol Settings) */
  .settings-popover::after {
    content: '';
    position: absolute;
    left: -6px;
    bottom: 16px;
    width: 11px;
    height: 11px;
    background: var(--bg-surface);
    border-left: 1px solid var(--border-color);
    border-bottom: 1px solid var(--border-color);
    transform: rotate(45deg);
  }
  .settings-popover-title {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
    padding: 6px 10px 4px;
  }
  .settings-pop-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    text-decoration: none;
    white-space: nowrap;
    transition: background-color 0.15s ease, color 0.15s ease;
  }
  .settings-pop-item:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }
  .settings-pop-item.active {
    background-color: var(--accent-light);
    color: var(--accent);
  }
  .settings-pop-icon {
    font-size: 17px;
    flex-shrink: 0;
  }

  .pop-enter-active,
  .pop-leave-active {
    transition: opacity 0.15s ease, transform 0.15s ease;
  }
  .pop-enter-from,
  .pop-leave-to {
    opacity: 0;
    transform: translateX(-6px);
  }
</style>
