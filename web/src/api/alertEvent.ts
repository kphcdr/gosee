import { request } from './http'
import type { AlertEvent } from '@/types/alert'

// 后端已实现（阶段五），直接走真实接口
export const alertEventApi = {
  list: (params?: { status?: string; server_id?: number }) =>
    request<AlertEvent[]>({ url: '/alert-events', method: 'get', params }),
  ack: (id: number) => request<void>({ url: `/alert-events/${id}/ack`, method: 'post' }),
  close: (id: number) => request<void>({ url: `/alert-events/${id}/close`, method: 'post' }),
}
