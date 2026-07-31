import { onBeforeUnmount, onMounted, readonly, ref } from 'vue'

function readDarkMode(): boolean {
  return typeof document !== 'undefined' && document.documentElement.classList.contains('dark')
}

/**
 * Keeps non-CSS UI such as Chart.js options in sync with the root theme class.
 */
export function useDarkMode() {
  const isDarkMode = ref(readDarkMode())
  let observer: MutationObserver | null = null

  const syncFromDom = () => {
    isDarkMode.value = readDarkMode()
  }

  onMounted(() => {
    syncFromDom()
    if (typeof MutationObserver === 'undefined') return

    observer = new MutationObserver(syncFromDom)
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    })
  })

  onBeforeUnmount(() => {
    observer?.disconnect()
    observer = null
  })

  return readonly(isDarkMode)
}
