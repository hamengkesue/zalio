const collapsed = ref(false)

export function useSidebar() {
  function toggle() {
    collapsed.value = !collapsed.value
  }

  function setCollapsed(value: boolean) {
    collapsed.value = value
  }

  return { collapsed, toggle, setCollapsed }
}
