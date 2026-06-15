import axios, { type AxiosRequestConfig, AxiosError } from 'axios'
import { message } from '@/utils/discrete'
import { getToken, clearToken } from '@/utils/storage'
import type { ApiResponse } from '@/types/api'

export const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 30000,
})

// 请求拦截：自动加 Authorization
http.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：剥壳 + 统一报错 + 401 跳登录
http.interceptors.response.use(
  (resp) => {
    const body = resp.data as ApiResponse<unknown>
    // 非标准响应（如静态文件）直接放行
    if (body == null || typeof body.code !== 'number') {
      return resp.data
    }
    if (body.code === 0) {
      return body.data
    }
    message.error(body.message || '请求失败')
    return Promise.reject(new Error(body.message || 'biz error'))
  },
  (err: AxiosError<ApiResponse<unknown>>) => {
    const status = err.response?.status
    if (status === 401) {
      clearToken()
      message.warning('登录已失效，请重新登录')
      // 用 window.location 跳转，避免 http ↔ router ↔ store ↔ api 循环依赖
      const redirect = window.location.pathname + window.location.search
      window.location.href = `/login?redirect=${encodeURIComponent(redirect)}`
      return Promise.reject(err)
    }
    const msg = err.response?.data?.message || err.message || '网络错误'
    // 400 通常是参数校验，交给调用方自行提示
    if (status !== 400) {
      message.error(msg)
    }
    return Promise.reject(err)
  },
)

// 业务侧统一用 request<T>，返回值即业务最终类型（拦截器已剥壳）
export function request<T = unknown>(config: AxiosRequestConfig): Promise<T> {
  return http.request<unknown, T>(config)
}
