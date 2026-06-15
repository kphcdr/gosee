// 后端统一响应：{ code, message, data }，code===0 成功
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 分页响应 data 结构
export interface PagedResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// 分页请求参数
export interface PagedQuery {
  page?: number
  page_size?: number
}
