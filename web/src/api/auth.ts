import { request } from './http'
import type { LoginInput, LoginResult, ChangePasswordInput, User } from '@/types/auth'

export const authApi = {
  login: (body: LoginInput) =>
    request<LoginResult>({ url: '/auth/login', method: 'post', data: body }),
  profile: () => request<User>({ url: '/auth/profile', method: 'get' }),
  changePassword: (body: ChangePasswordInput) =>
    request<void>({ url: '/auth/password', method: 'put', data: body }),
}
