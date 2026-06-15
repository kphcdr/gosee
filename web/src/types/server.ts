import type { AuthType } from '@/constants/server'

export type { AuthType }
export type { ServerStatus } from '@/constants/server'

export interface Server {
  id: number
  name: string
  group_id: number | null
  host: string
  port: number
  username: string
  auth_type: AuthType
  remark: string
  status: string
  enabled: 0 | 1
  last_checked_at: string | null
  last_error: string | null
  created_at: string
  updated_at: string
}

// 新建/编辑入参。私钥/密码留空 = 不改（后端 applyCredentials 已处理）
export interface ServerSaveInput {
  name: string
  group_id?: number | null
  host: string
  port?: number
  username: string
  auth_type?: AuthType
  private_key?: string
  passphrase?: string
  password?: string
  remark?: string
  enabled?: 0 | 1
}

export interface ServerListQuery {
  page?: number
  page_size?: number
  group_id?: number
  enabled?: 0 | 1
  keyword?: string
}
