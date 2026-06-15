import dayjs from 'dayjs'

// 时间格式化
export function formatTime(t: string | number | null | undefined, fmt = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!t) return '-'
  const d = dayjs(t)
  return d.isValid() ? d.format(fmt) : '-'
}

// 字节 → 人类可读
export function formatBytes(bytes: number | null | undefined, decimals = 1): string {
  if (!bytes || bytes <= 0) return '0 B'
  const k = 1024
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  const idx = Math.min(i, units.length - 1)
  return `${(bytes / Math.pow(k, idx)).toFixed(decimals)} ${units[idx]}`
}

// 百分比
export function formatPercent(v: number | null | undefined, decimals = 1): string {
  if (v == null || Number.isNaN(v)) return '-'
  return `${v.toFixed(decimals)}%`
}

// 运行时长（秒 → 天/小时/分钟）
export function formatUptime(seconds: number | null | undefined): string {
  if (!seconds || seconds <= 0) return '-'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}天${h}小时`
  if (h > 0) return `${h}小时${m}分钟`
  return `${m}分钟`
}
