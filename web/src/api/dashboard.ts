import { request } from './http'
import type { DashboardSummary, TopItem, RecentAlert } from '@/types/dashboard'

// 后端已实现（阶段七），直接走真实接口
export const dashboardApi = {
  summary: (groupId?: number) => request<DashboardSummary>({ url: '/dashboard/summary', method: 'get', params: { group_id: groupId } }),
  topCpu: (groupId?: number) => request<TopItem[]>({ url: '/dashboard/top-cpu', method: 'get', params: { group_id: groupId } }),
  topMemory: (groupId?: number) => request<TopItem[]>({ url: '/dashboard/top-memory', method: 'get', params: { group_id: groupId } }),
  topDisk: (groupId?: number) => request<TopItem[]>({ url: '/dashboard/top-disk', method: 'get', params: { group_id: groupId } }),
  recentAlerts: (groupId?: number) => request<RecentAlert[]>({ url: '/dashboard/recent-alerts', method: 'get', params: { group_id: groupId } }),
}
