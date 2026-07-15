<script setup lang="ts">
  const route = useRoute()
  const { collapsed, toggle } = useSidebar()
  const { isAdmin } = useAuth()

  const menuGroups = computed(() => [
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
    // Administration menu only shows for admins.
    ...(isAdmin.value
      ? [{
          label: 'Administration',
          items: [
            { label: 'Users', icon: 'i-lucide-users', to: '/users' },
          ],
        }]
      : []),
  ])
</script>

<template>
  <aside class="app-sidebar" :class="{ collapsed }">
    <nav class="sidebar-nav">
      <div v-for="(group, gi) in menuGroups" :key="gi" class="nav-group">
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

    <button class="collapse-btn" :title="collapsed ? 'Expand' : 'Collapse'" @click="toggle">
      <UIcon :name="collapsed ? 'i-lucide-chevrons-right' : 'i-lucide-chevrons-left'" class="text-[18px]" />
      <span v-if="!collapsed" class="sidebar-label">Collapse</span>
    </button>
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
</style>
