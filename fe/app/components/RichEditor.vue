<script setup lang="ts">
  // Editor teks berformat ringan (tanpa library) — menyimpan konten sebagai HTML.
  const model = defineModel<string>({ default: '' })
  defineProps<{ placeholder?: string; invalid?: boolean }>()

  const editor = ref<HTMLElement>()

  // Sinkron nilai luar → editor (hanya bila beda, agar kursor tidak lompat saat mengetik).
  watch(model, (v) => {
    if (editor.value && editor.value.innerHTML !== (v || '')) editor.value.innerHTML = v || ''
  })
  onMounted(() => { if (editor.value) editor.value.innerHTML = model.value || '' })

  function sync() {
    if (editor.value) model.value = editor.value.innerHTML
  }
  // Saat tempel: pertahankan bold/italic/underline/list/link, tapi buang
  // font/ukuran/warna sumber agar seragam dengan font app.
  const ALLOWED = new Set(['B', 'STRONG', 'I', 'EM', 'U', 'UL', 'OL', 'LI', 'BR', 'P', 'DIV', 'A'])
  function cleanNode(node: Node) {
    for (const child of [...node.childNodes]) {
      if (child.nodeType === 1) {
        const el = child as HTMLElement
        for (const attr of [...el.attributes]) {
          if (!(el.tagName === 'A' && attr.name === 'href')) el.removeAttribute(attr.name)
        }
        cleanNode(el)
        if (!ALLOWED.has(el.tagName)) el.replaceWith(...el.childNodes)
      } else if (child.nodeType === 8) {
        child.remove()
      }
    }
  }
  function onPaste(e: ClipboardEvent) {
    e.preventDefault()
    const html = e.clipboardData?.getData('text/html')
    if (html) {
      const holder = document.createElement('div')
      holder.innerHTML = html
      cleanNode(holder)
      document.execCommand('insertHTML', false, holder.innerHTML)
    } else {
      document.execCommand('insertText', false, e.clipboardData?.getData('text/plain') ?? '')
    }
    sync()
  }
  function cmd(command: string, value?: string) {
    editor.value?.focus()
    document.execCommand(command, false, value)
    sync()
  }
  function addLink() {
    const url = window.prompt('Enter URL (https://…):')
    if (url) cmd('createLink', url)
  }

  // Editor dianggap kosong bila tidak ada teks/gambar (mengabaikan tag & &nbsp;).
  const isEmpty = computed(() => {
    const v = (model.value || '').replace(/<[^>]*>/g, '').replace(/&nbsp;/g, '').trim()
    return v === ''
  })

  const tools = [
    { cmd: 'bold', icon: 'i-lucide-bold', title: 'Bold' },
    { cmd: 'italic', icon: 'i-lucide-italic', title: 'Italic' },
    { cmd: 'underline', icon: 'i-lucide-underline', title: 'Underline' },
    { sep: true },
    { cmd: 'insertUnorderedList', icon: 'i-lucide-list', title: 'Bullet list' },
    { cmd: 'insertOrderedList', icon: 'i-lucide-list-ordered', title: 'Numbered list' },
    { sep: true },
    { cmd: 'justifyLeft', icon: 'i-lucide-align-left', title: 'Align left' },
    { cmd: 'justifyCenter', icon: 'i-lucide-align-center', title: 'Align center' },
    { cmd: 'justifyRight', icon: 'i-lucide-align-right', title: 'Align right' },
    { sep: true },
    { link: true, icon: 'i-lucide-link', title: 'Insert link' },
  ] as const
</script>

<template>
  <div class="rte" :class="{ 'rte-invalid': invalid }">
    <div class="rte-toolbar">
      <template v-for="(t, i) in tools" :key="i">
        <span v-if="'sep' in t" class="rte-sep" />
        <button
          v-else
          type="button"
          class="rte-btn"
          :title="t.title"
          @mousedown.prevent="'link' in t ? addLink() : cmd(t.cmd)"
        ><UIcon :name="t.icon" /></button>
      </template>
    </div>
    <div class="rte-body">
      <div
        ref="editor"
        class="rte-content"
        contenteditable="true"
        @input="sync"
        @blur="sync"
        @paste="onPaste"
      />
      <span v-if="isEmpty" class="rte-placeholder">{{ placeholder }}</span>
    </div>
  </div>
</template>

<style scoped>
  .rte {
    border: 1px solid var(--border-color);
    border-radius: 12px;
    overflow: hidden;
    background: var(--bg-surface);
  }
  .rte-invalid { border-color: var(--danger); }
  .rte-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 2px;
    padding: 6px 8px;
    background: var(--bg-muted);
    border-bottom: 1px solid var(--border-color);
  }
  .rte-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    border: none;
    border-radius: 7px;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 16px;
  }
  .rte-btn:hover { background: var(--bg-hover); color: var(--text-primary); }
  .rte-sep { width: 1px; height: 20px; background: var(--border-color); margin: 0 4px; }
  .rte-body { position: relative; }
  .rte-content {
    height: 90px;            /* ~3 baris, seperti textarea sebelumnya */
    overflow-y: auto;        /* teks panjang → scroll di dalam kotak */
    padding: 12px 14px;
    font-family: var(--font-family);
    font-size: 14px;
    line-height: 1.5;
    color: var(--text-primary);
    outline: none;
  }
  /* paksa semua konten (termasuk hasil paste) memakai font & ukuran app */
  .rte-content :deep(*) { font-family: var(--font-family) !important; font-size: inherit; }
  .rte-content :deep(ul) { padding-left: 22px; list-style: disc; }
  .rte-content :deep(ol) { padding-left: 22px; list-style: decimal; }
  .rte-content :deep(a) { color: var(--accent); text-decoration: underline; }
  .rte-placeholder {
    position: absolute;
    top: 12px;
    left: 14px;
    font-size: 14px;
    color: var(--text-muted);
    pointer-events: none;
  }
</style>
