import { useIntervalFn } from '@vueuse/core'
import { onBeforeUnmount } from 'vue'

// 轮询：基于 useIntervalFn，组件卸载自动暂停
export function usePolling(fn: () => void | Promise<void>, interval = 30_000, immediate = false) {
  if (immediate) fn()
  const { pause, resume } = useIntervalFn(fn, interval)
  onBeforeUnmount(pause)
  return { pause, resume }
}
