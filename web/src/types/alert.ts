import type { MetricType, ScopeType, AlertLevel } from '@/constants/alert'

export type { MetricType, ScopeType, AlertLevel }

export interface AlertRule {
  id: number
  name: string
  metric_type: MetricType
  operator: string
  threshold: number
  duration_count: number
  level: AlertLevel
  scope_type: ScopeType
  scope_id: number | null
  enabled: 0 | 1
  created_at: string
  updated_at: string
}

// 创建/编辑入参（不含 id/时间戳）
export type AlertRuleInput = Omit<AlertRule, 'id' | 'created_at' | 'updated_at'>

export interface AlertEvent {
  id: number
  server_id: number
  server_name: string
  rule_name: string
  metric: string
  current_value: number
  threshold: number
  level: AlertLevel
  status: 'firing' | 'recovered' | 'closed'
  first_triggered_at: string
  last_triggered_at: string
  recovered_at: string | null
  acked_at: string | null
  acked_by: number | null
}
