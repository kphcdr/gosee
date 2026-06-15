// token 持久化（仅 token 需持久化，user 信息在路由进入时拉取最新）
const TOKEN_KEY = 'gosee_token'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}
