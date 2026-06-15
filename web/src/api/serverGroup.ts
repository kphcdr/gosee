import { request } from './http'
import type { ServerGroup, ServerGroupInput } from '@/types/serverGroup'

// 注意：后端 /server-groups 直接返回数组（非分页）
export const serverGroupApi = {
  list: (keyword?: string) =>
    request<ServerGroup[]>({
      url: '/server-groups',
      method: 'get',
      params: keyword ? { keyword } : undefined,
    }),
  create: (body: ServerGroupInput) =>
    request<ServerGroup>({ url: '/server-groups', method: 'post', data: body }),
  update: (id: number, body: ServerGroupInput) =>
    request<ServerGroup>({ url: `/server-groups/${id}`, method: 'put', data: body }),
  remove: (id: number) => request<void>({ url: `/server-groups/${id}`, method: 'delete' }),
}
