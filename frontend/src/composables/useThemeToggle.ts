import { onMounted, ref } from 'vue'

/**
 * 明暗主题切换。与 main.ts 的 initThemeClass 共用 `theme` 这个 localStorage 键
 * 和 documentElement 上的 `dark` 类,保证刷新后状态一致。
 *
 * 上游视图(AppSidebar / KeyUsageView 等)各自持有同样逻辑的副本,按 D9 不在本
 * change 里重构;TokenPort 自有页面统一走这里,避免继续增加副本。
 */
export function useThemeToggle() {
  const isDark = ref(false)

  const syncFromDom = () => {
    isDark.value = document.documentElement.classList.contains('dark')
  }

  const toggleTheme = () => {
    isDark.value = !isDark.value
    document.documentElement.classList.toggle('dark', isDark.value)
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  }

  // SSR/测试环境下 document 可能尚未就绪,挂载后再取一次真实状态。
  syncFromDom()
  onMounted(syncFromDom)

  return { isDark, toggleTheme }
}
