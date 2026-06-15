export type ChannelType = 'feishu' | 'telegram' | 'email'

export interface NotificationChannel {
  id: number
  name: string
  type: ChannelType
  config: Record<string, string>
  enabled: 0 | 1
  created_at: string
  updated_at: string
}

export type NotificationChannelInput = Omit<NotificationChannel, 'id' | 'created_at' | 'updated_at'>
