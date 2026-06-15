import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api/auth'
import { getToken, setToken, clearToken } from '@/utils/storage'
import type { LoginInput, User } from '@/types/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(getToken())
  const user = ref<User | null>(null)

  async function login(input: LoginInput) {
    const res = await authApi.login(input)
    token.value = res.token
    setToken(res.token)
    user.value = res.user
  }

  async function fetchProfile() {
    if (!token.value) return
    try {
      user.value = await authApi.profile()
    } catch {
      // 401 已由 http 拦截器统一处理
    }
  }

  async function changePassword(old_password: string, new_password: string) {
    await authApi.changePassword({ old_password, new_password })
  }

  function logout() {
    token.value = ''
    user.value = null
    clearToken()
  }

  return { token, user, login, fetchProfile, changePassword, logout }
})
