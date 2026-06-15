import { request } from './http'
import type { DashboardSummary, TopItem, RecentAlert } from '@/types/dashboard'

// 后端已实现（阶段七），直接走真实接口
export const dashboardApi = {
  summary: () => request<DashboardSummary>({ url: '/dashboard/summary', method: 'get' }),
  topCpu: () => request<TopItem[]>({ url: '/dashboard/top-cpu', method: 'get' }),
  topMemory: () => request<TopItem[]>({ url: '/dashboard/top-memory', method: 'get' }),
  topDisk: () => request<TopItem[]>({ url: '/dashboard/top-disk', method: 'get' }),
  recentAlerts: () => request<RecentAlert[]>({ url: '/dashboard/recent-alerts', method: 'get' }),
}
