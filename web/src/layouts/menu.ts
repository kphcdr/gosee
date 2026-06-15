import type { MenuOption } from 'naive-ui'
import { renderIcon } from '@/utils/render'
import {
  GridOutline,
  ServerOutline,
  NotificationsOutline,
  AlertCircleOutline,
} from '@vicons/ionicons5'

// 菜单 key 直接用路由 path，NMenu update:value 即可跳转
export const menuOptions: MenuOption[] = [
  { label: '仪表盘', key: '/dashboard', icon: renderIcon(GridOutline) },
  {
    label: '服务器管理',
    key: 'server-mgmt',
    icon: renderIcon(ServerOutline),
    children: [
      { label: '服务器列表', key: '/servers' },
      { label: '服务器分组', key: '/server-groups' },
    ],
  },
  {
    label: '告警中心',
    key: 'alert-center',
    icon: renderIcon(AlertCircleOutline),
    children: [
      { label: '告警规则', key: '/alert-rules' },
      { label: '告警事件', key: '/alert-events' },
    ],
  },
  { label: '通知通道', key: '/notification-channels', icon: renderIcon(NotificationsOutline) },
]
