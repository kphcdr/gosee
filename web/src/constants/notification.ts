import type { ChannelType } from '@/types/notification'

export const CHANNEL_TYPE_OPTIONS: { label: string; value: ChannelType }[] = [
  { label: '飞书机器人', value: 'feishu' },
  { label: 'Telegram Bot', value: 'telegram' },
  { label: '邮件 SMTP', value: 'email' },
]

export interface ChannelField {
  key: string
  label: string
  placeholder?: string
  secret?: boolean
}

// 各通道类型的配置字段（用于动态渲染表单）
export const CHANNEL_FIELDS: Record<ChannelType, ChannelField[]> = {
  feishu: [
    { key: 'webhook', label: 'Webhook URL', placeholder: 'https://open.feishu.cn/open-apis/bot/v2/hook/xxx' },
    { key: 'secret', label: '签名密钥', placeholder: '可选，用于校验', secret: true },
  ],
  telegram: [
    { key: 'bot_token', label: 'Bot Token', secret: true },
    { key: 'chat_id', label: 'Chat ID' },
  ],
  email: [
    { key: 'smtp_host', label: 'SMTP 主机' },
    { key: 'smtp_port', label: 'SMTP 端口' },
    { key: 'username', label: '用户名' },
    { key: 'password', label: '密码', secret: true },
    { key: 'from', label: '发件人邮箱' },
  ],
}
