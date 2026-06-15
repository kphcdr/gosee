export type AlertLevel = 'info' | 'warning' | 'critical'

export const LEVEL_OPTIONS: { label: string; value: AlertLevel }[] = [
  { label: '提醒', value: 'info' },
  { label: '警告', value: 'warning' },
  { label: '严重', value: 'critical' },
]

const LEVEL_META: Record<AlertLevel, { label: string; type: 'default' | 'info' | 'warning' | 'error' }> = {
  info: { label: '提醒', type: 'info' },
  warning: { label: '警告', type: 'warning' },
  critical: { label: '严重', type: 'error' },
}

export function levelLabel(level: string | null | undefined): string {
  if (!level) return '-'
  return LEVEL_META[level as AlertLevel]?.label || level
}

export function levelType(level: string | null | undefined): 'default' | 'info' | 'warning' | 'error' {
  if (!level) return 'default'
  return LEVEL_META[level as AlertLevel]?.type || 'default'
}

// 告警规则相关枚举（PRD 6.2）
export type MetricType =
  | 'cpu_usage'
  | 'memory_usage'
  | 'disk_usage'
  | 'load_1'
  | 'load_5'
  | 'load_15'
  | 'ssh_fail'

export const METRIC_TYPE_OPTIONS: { label: string; value: MetricType; unit: string }[] = [
  { label: 'CPU 使用率', value: 'cpu_usage', unit: '%' },
  { label: '内存使用率', value: 'memory_usage', unit: '%' },
  { label: '磁盘使用率', value: 'disk_usage', unit: '%' },
  { label: '1 分钟负载', value: 'load_1', unit: '' },
  { label: '5 分钟负载', value: 'load_5', unit: '' },
  { label: '15 分钟负载', value: 'load_15', unit: '' },
  { label: 'SSH 连接失败', value: 'ssh_fail', unit: '次' },
]

export const OPERATOR_OPTIONS: { label: string; value: string }[] = [
  { label: '> 大于', value: '>' },
  { label: '>= 大于等于', value: '>=' },
  { label: '< 小于', value: '<' },
  { label: '<= 小于等于', value: '<=' },
  { label: '== 等于', value: '==' },
  { label: '!= 不等于', value: '!=' },
]

export type ScopeType = 'global' | 'group' | 'server'

export const SCOPE_TYPE_OPTIONS: { label: string; value: ScopeType }[] = [
  { label: '全局（所有服务器）', value: 'global' },
  { label: '指定分组', value: 'group' },
  { label: '指定服务器', value: 'server' },
]
