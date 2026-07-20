<script setup lang="ts">
  // Modal scan barcode via kamera (HP/webcam). Dibuka lewat v-model (boolean).
  // Saat terdeteksi → emit 'detected' dengan teks barcode, lalu tutup.
  // Library @zxing/browser di-import dinamis (client-only) supaya aman SSR.
  const open = defineModel<boolean>({ default: false })
  // confirm = berapa kali barcode yang sama harus terbaca berturut-turut sebelum diterima.
  const props = defineProps<{ confirm?: number }>()
  const emit = defineEmits<{ detected: [code: string] }>()

  const videoEl = ref<HTMLVideoElement>()
  const status = ref<'idle' | 'starting' | 'scanning' | 'error'>('idle')
  const errorMsg = ref('')
  let controls: { stop: () => void } | null = null
  // Konfirmasi: barcode harus terbaca sama beberapa kali berturut-turut → cegah salah-baca.
  let lastCode = ''
  let hits = 0

  async function start() {
    status.value = 'starting'
    errorMsg.value = ''
    try {
      const [{ BrowserMultiFormatReader }, { DecodeHintType, BarcodeFormat }] = await Promise.all([
        import('@zxing/browser'),
        import('@zxing/library'),
      ])
      // Hints: coba lebih keras + batasi ke format barcode produk yang umum → lebih peka & cepat.
      const hints = new Map<number, unknown>()
      hints.set(DecodeHintType.TRY_HARDER, true)
      hints.set(DecodeHintType.POSSIBLE_FORMATS, [
        BarcodeFormat.EAN_13, BarcodeFormat.EAN_8, BarcodeFormat.UPC_A, BarcodeFormat.UPC_E,
        BarcodeFormat.CODE_128, BarcodeFormat.CODE_39, BarcodeFormat.ITF, BarcodeFormat.CODABAR,
        BarcodeFormat.QR_CODE,
      ])
      const reader = new BrowserMultiFormatReader(hints, { delayBetweenScanAttempts: 200, delayBetweenScanSuccess: 400 })
      await nextTick()
      if (!videoEl.value) return
      lastCode = ''
      hits = 0
      const need = Math.max(1, props.confirm ?? 2)
      controls = await reader.decodeFromConstraints(
        { video: { facingMode: { ideal: 'environment' }, width: { ideal: 1280 }, height: { ideal: 720 } } },
        videoEl.value,
        (result) => {
          if (!result) return
          const text = result.getText()
          // Terima hanya setelah barcode yang sama terbaca `need` kali berturut-turut.
          if (text === lastCode) hits++
          else { lastCode = text; hits = 1 }
          if (hits >= need) {
            emit('detected', text)
            open.value = false
          }
        },
      )
      status.value = 'scanning'
    } catch (e: any) {
      status.value = 'error'
      errorMsg.value = e?.name === 'NotAllowedError'
        ? 'Camera permission denied. Please allow camera access and try again.'
        : e?.name === 'NotFoundError'
          ? 'No camera found on this device. Use a USB scanner instead.'
          : (e?.message || 'Could not start the camera.')
    }
  }

  function stop() {
    try { controls?.stop() } catch { /* ignore */ }
    controls = null
    status.value = 'idle'
  }

  watch(open, (v) => { if (v) start(); else stop() })
  onUnmounted(stop)
</script>

<template>
  <Teleport to="body">
    <Transition name="scan-fade">
      <div v-if="open" class="scan-overlay" @click.self="open = false">
        <div class="scan-box">
          <div class="scan-head">
            <span>Scan barcode</span>
            <button type="button" class="scan-x" title="Close" @click="open = false"><UIcon name="i-lucide-x" /></button>
          </div>
          <div class="scan-video-wrap">
            <video ref="videoEl" class="scan-video" autoplay playsinline muted />
            <div v-if="status === 'scanning'" class="scan-frame" />
            <div v-if="status === 'starting'" class="scan-loading"><UIcon name="i-lucide-loader-circle" class="spin" /> Starting camera…</div>
          </div>
          <div v-if="status === 'error'" class="scan-msg error">{{ errorMsg }}</div>
          <div v-else class="scan-msg">Point the camera at a barcode.</div>
          <button type="button" class="scan-cancel" @click="open = false">Cancel</button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
  .scan-overlay {
    position: fixed;
    inset: 0;
    z-index: 300;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgba(15, 23, 42, 0.6);
    backdrop-filter: blur(2px);
  }
  .scan-box {
    width: min(440px, 92vw);
    background: var(--bg-surface);
    border-radius: 16px;
    padding: 18px;
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.35);
  }
  .scan-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 16px;
    font-weight: 800;
    color: var(--text-primary);
    margin-bottom: 14px;
  }
  .scan-x {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    border-radius: 8px;
    border: none;
    background: var(--bg-muted);
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 16px;
  }
  .scan-x:hover { background: var(--bg-hover); color: var(--text-primary); }
  .scan-video-wrap {
    position: relative;
    width: 100%;
    aspect-ratio: 4 / 3;
    background: #000;
    border-radius: 12px;
    overflow: hidden;
  }
  .scan-video { width: 100%; height: 100%; object-fit: cover; display: block; }
  .scan-frame {
    position: absolute;
    inset: 18% 12%;
    border: 3px solid var(--accent);
    border-radius: 10px;
    box-shadow: 0 0 0 100vmax rgba(0, 0, 0, 0.25);
  }
  .scan-loading {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: #fff;
    font-size: 13px;
  }
  .spin { animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
  .scan-msg { margin: 12px 0; font-size: 13px; color: var(--text-muted); text-align: center; }
  .scan-msg.error { color: var(--danger); font-weight: 600; }
  .scan-cancel {
    width: 100%;
    padding: 11px;
    border-radius: 10px;
    border: none;
    background: var(--bg-muted);
    color: var(--text-secondary);
    font-size: 14px;
    font-weight: 700;
    cursor: pointer;
  }
  .scan-cancel:hover { background: var(--bg-hover); color: var(--text-primary); }

  .scan-fade-enter-active, .scan-fade-leave-active { transition: opacity 0.15s ease; }
  .scan-fade-enter-from, .scan-fade-leave-to { opacity: 0; }
</style>
