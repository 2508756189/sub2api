import { onMounted, onUnmounted, type Ref } from 'vue'

// 排除 disabled 与 tabindex="-1",否则 Tab 会停在无法交互的元素上。
const FOCUSABLE_SELECTOR =
  'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

/** 取容器内当前真正可聚焦的元素(已过滤不可见项)。 */
export function getFocusableWithin(root: HTMLElement | null | undefined): HTMLElement[] {
  if (!root) return []
  return Array.from(root.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
    (el) => el.getClientRects().length > 0
  )
}

/**
 * 把 Tab / Shift+Tab 圈定在 `root` 内(frontend-a11y R1)。
 *
 * 只在焦点确实位于 `root` 内时介入——嵌套弹窗同时挂载时,下层若也去抢焦点,
 * 上层弹窗就永远拿不到,所以这里必须让位。
 *
 * @param root     圈定范围的容器(通常是弹窗面板,不含遮罩层)
 * @param isActive 弹窗当前是否可见;不可见时不介入
 */
export function useFocusTrap(root: Ref<HTMLElement | null>, isActive: () => boolean = () => true) {
  const handleKeydown = (event: KeyboardEvent) => {
    if (event.key !== 'Tab' || !isActive()) return

    const container = root.value
    const active = document.activeElement as HTMLElement | null
    if (!container || !active || !container.contains(active)) return

    const focusable = getFocusableWithin(container)
    if (!focusable.length) return

    const first = focusable[0]
    const last = focusable[focusable.length - 1]

    if (event.shiftKey && active === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && active === last) {
      event.preventDefault()
      first.focus()
    }
  }

  onMounted(() => document.addEventListener('keydown', handleKeydown))
  onUnmounted(() => document.removeEventListener('keydown', handleKeydown))
}
