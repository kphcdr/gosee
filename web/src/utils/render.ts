import { h, type Component } from 'vue'
import { NIcon } from 'naive-ui'

// 渲染图标：NMenu 的 icon 字段需要一个 () => VNodeChild 的函数
export function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}
