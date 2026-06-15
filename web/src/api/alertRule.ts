import { request } from './http'
import type { AlertRule, AlertRuleInput } from '@/types/alert'

// 后端已实现（阶段五），直接走真实接口
export const alertRuleApi = {
  list: () => request<AlertRule[]>({ url: '/alert-rules', method: 'get' }),
  create: (body: AlertRuleInput) => request<AlertRule>({ url: '/alert-rules', method: 'post', data: body }),
  update: (id: number, body: AlertRuleInput) =>
    request<AlertRule>({ url: `/alert-rules/${id}`, method: 'put', data: body }),
  remove: (id: number) => request<void>({ url: `/alert-rules/${id}`, method: 'delete' }),
  enable: (id: number) => request<void>({ url: `/alert-rules/${id}/enable`, method: 'post' }),
  disable: (id: number) => request<void>({ url: `/alert-rules/${id}/disable`, method: 'post' }),
}
