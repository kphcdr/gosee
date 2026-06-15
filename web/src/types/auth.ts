export interface User {
  id: number
  username: string
  nickname: string
  email: string
}

export interface LoginInput {
  username: string
  password: string
}

// 后端 LoginResult：token + expire_in(秒) + user
export interface LoginResult {
  token: string
  expire_in: number
  user: User
}

export interface ChangePasswordInput {
  old_password: string
  new_password: string
}
