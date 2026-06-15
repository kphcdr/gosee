import { createDiscreteApi } from 'naive-ui'

// 离散 API：供 axios 拦截器等组件树外环境调用 message/dialog/notification/loadingBar
// 页面内组件仍可用 useMessage 等（App.vue 包了 Provider）
export const { message, dialog, notification, loadingBar } = createDiscreteApi(
  ['message', 'dialog', 'notification', 'loadingBar'],
)
