import { request } from './http'
import type { NotificationChannel, NotificationChannelInput } from '@/types/notification'

// 后端已实现（阶段六，飞书），直接走真实接口
export const notificationChannelApi = {
  list: () => request<NotificationChannel[]>({ url: '/notification-channels', method: 'get' }),
  create: (body: NotificationChannelInput) =>
    request<NotificationChannel>({ url: '/notification-channels', method: 'post', data: body }),
  update: (id: number, body: NotificationChannelInput) =>
    request<NotificationChannel>({ url: `/notification-channels/${id}`, method: 'put', data: body }),
  remove: (id: number) => request<void>({ url: `/notification-channels/${id}`, method: 'delete' }),
  test: (id: number) => request<void>({ url: `/notification-channels/${id}/test`, method: 'post' }),
}
