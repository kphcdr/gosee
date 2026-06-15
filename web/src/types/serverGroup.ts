export interface ServerGroup {
  id: number
  name: string
  remark: string
  created_at: string
  updated_at: string
}

export interface ServerGroupInput {
  name: string
  remark?: string
}
