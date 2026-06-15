export interface ServerMetric {
  id: number
  server_id: number
  hostname: string
  os: string
  cpu_usage: number
  cpu_cores: number
  memory_total_mb: number
  memory_used_mb: number
  memory_available_mb: number
  memory_usage: number
  load_1: number
  load_5: number
  load_15: number
  disk_max_usage: number
  uptime_seconds: number
  collected_at: string
  created_at: string
}

export interface ServerDisk {
  id: number
  metric_id: number
  server_id: number
  filesystem: string
  mount_point: string
  size_bytes: number
  used_bytes: number
  available_bytes: number
  usage_percent: number
  created_at: string
}

export interface CollectResult {
  server_id: number
  success: boolean
  metric?: ServerMetric
  error?: string
}

export interface TrendResult {
  list: ServerMetric[]
  total: number
}

export interface TrendQuery {
  hours?: number
  limit?: number
}
