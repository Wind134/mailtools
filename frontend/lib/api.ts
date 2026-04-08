const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export interface LoginResponse {
  token: string
  expires_at: number
  username: string
  is_admin: boolean
}

export interface SendMailResponse {
  success: boolean
  message: string
}

export interface User {
  id: number
  username: string
  is_admin: boolean
  created_at: number
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  const res = await fetch(`${API_BASE}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })

  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || '登录失败')
  }

  return res.json()
}

export async function register(username: string, password: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })

  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || '注册失败')
  }
}

export async function sendMail(
  token: string,
  to: string,
  subject: string,
  body: string,
  attachments?: File[]
): Promise<SendMailResponse> {
  const formData = new FormData()
  formData.append('to', to)
  formData.append('subject', subject)
  formData.append('body', body)
  
  if (attachments) {
    for (const file of attachments) {
      formData.append('attachments', file)
    }
  }

  const res = await fetch(`${API_BASE}/api/send`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    body: formData,
  })

  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || '发送失败')
  }

  return res.json()
}

export async function getUsers(token: string): Promise<User[]> {
  const res = await fetch(`${API_BASE}/api/admin/users`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  })

  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || '获取用户列表失败')
  }

  const data = await res.json()
  return data.users
}

export async function createUser(
  token: string,
  username: string,
  password: string,
  isAdmin: boolean
): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/users`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({ username, password, is_admin: isAdmin }),
  })

  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || '创建用户失败')
  }
}

export async function updateUser(
  token: string,
  id: number,
  username: string,
  isAdmin: boolean
): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/users/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({ username, is_admin: isAdmin }),
  })

  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || '更新用户失败')
  }
}

export async function deleteUser(token: string, id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/users/${id}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  })

  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || '删除用户失败')
  }
}

export function setToken(token: string) {
  localStorage.setItem('mailtools_token', token)
}

export function getToken(): string | null {
  return localStorage.getItem('mailtools_token')
}

export function removeToken() {
  localStorage.removeItem('mailtools_token')
}

export function setUsername(username: string) {
  localStorage.setItem('mailtools_username', username)
}

export function getUsername(): string | null {
  return localStorage.getItem('mailtools_username')
}

export function removeUsername() {
  localStorage.removeItem('mailtools_username')
}

export function setIsAdmin(isAdmin: boolean) {
  localStorage.setItem('mailtools_is_admin', isAdmin ? 'true' : 'false')
}

export function getIsAdmin(): boolean {
  return localStorage.getItem('mailtools_is_admin') === 'true'
}

export function removeIsAdmin() {
  localStorage.removeItem('mailtools_is_admin')
}
