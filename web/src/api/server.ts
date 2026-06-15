import { request } from './http'
import type { PagedResult } from '@/types/api'
import type { Server, ServerSaveInput, ServerListQuery } from '@/types/server'
import type { ServerDisk, CollectResult, TrendResult, TrendQuery } from '@/types/metric'

export const serverApi = {
  list: (q: ServerListQuery = {}) =>
    request<PagedResult<Server>>({ url: '/servers', method: 'get', params: q }),
  get: (id: number) => request<Server>({ url: `/servers/${id}`, method: 'get' }),
  create: (body: ServerSaveInput) => request<Server>({ url: '/servers', method: 'post', data: body }),
  update: (id: number, body: ServerSaveInput) =>
    request<Server>({ url: `/servers/${id}`, method: 'put', data: body }),
  remove: (id: number) => request<void>({ url: `/servers/${id}`, method: 'delete' }),
  testSSH: (id: number) => request<void>({ url: `/servers/${id}/test-ssh`, method: 'post' }),
  collect: (id: number) => request<CollectResult>({ url: `/servers/${id}/collect`, method: 'post' }),
  metrics: (id: number, q: TrendQuery = {}) =>
    request<TrendResult>({ url: `/servers/${id}/metrics`, method: 'get', params: q }),
  // 注意：disks 直接返回数组（非分页）
  disks: (id: number) => request<ServerDisk[]>({ url: `/servers/${id}/disks`, method: 'get' }),
}
