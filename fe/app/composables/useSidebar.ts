const collapsed = ref(true) // default: collapsed

export function useSidebar() {
  function toggle() {
    collapsed.value = !collapsed.value
  }

  function setCollapsed(value: boolean) {
    collapsed.value = value
  }

  return { collapsed, toggle, setCollapsed }
}
