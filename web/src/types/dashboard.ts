export interface DashboardSummary {
  total: number
  normal: number
  warning: number
  critical: number
  offline: number
}

export interface TopItem {
  server_id: number
  name: string
  host: string
  value: number
}

export interface RecentAlert {
  id: number
  server_name: string
  metric: string
  value: number
  level: 'info' | 'warning' | 'critical'
  status: 'firing' | 'recovered'
  triggered_at: string
}
