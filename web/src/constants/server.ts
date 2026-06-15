// 服务器状态枚举（与后端 model.Server.Status 对齐）
export type ServerStatus = 'normal' | 'warning' | 'critical' | 'offline' | 'disabled' | 'unknown'

export type AuthType = 'private_key' | 'password'

// 状态 → 标签 + Naive Tag 类型（颜色）
export const STATUS_META: Record<ServerStatus, { label: string; type: TagType }> = {
  normal: { label: '正常', type: 'success' },
  warning: { label: '预警', type: 'warning' },
  critical: { label: '严重', type: 'error' },
  offline: { label: '离线', type: 'default' },
  disabled: { label: '已禁用', type: 'default' },
  unknown: { label: '未知', type: 'default' },
}

type TagType = 'success' | 'warning' | 'error' | 'default'

export function statusLabel(status: string | null | undefined): string {
  if (!status) return '-'
  return STATUS_META[status as ServerStatus]?.label || status
}

export function statusType(status: string | null | undefined): TagType {
  if (!status) return 'default'
  return STATUS_META[status as ServerStatus]?.type || 'default'
}

export const AUTH_TYPE_OPTIONS: { label: string; value: AuthType }[] = [
  { label: '私钥认证', value: 'private_key' },
  { label: '密码认证', value: 'password' },
]

export const ENABLED_OPTIONS: { label: string; value: 0 | 1 }[] = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 },
]
