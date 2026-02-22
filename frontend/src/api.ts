import type { DataResponse, LoginRequest, User } from './types'

async function apiFetch<T>(url: string, method: string, params?: object): Promise<T | null> {
  let response: Response

  if (method === 'GET') {
    const query = params ? `?${params}` : ''
    response = await fetch(`${url}${query}`)
  } else {
    const options: RequestInit = { method }
    if (params) {
      options.headers = { 'Content-Type': 'application/json' }
      options.body = JSON.stringify(params)
    }
    response = await fetch(url, options)
  }

  if (response.status === 401) return null
  if (!response.ok) throw new Error(`${method} ${url} failed: ${response.status}`)
  if (response.status === 204) return null

  return response.json() as Promise<T>
}

export function fetchData(params: URLSearchParams): Promise<DataResponse | null> {
  return apiFetch('/api/data', 'GET', params)
}

export function syncData(): Promise<null> {
  return apiFetch('/api/sync', 'POST')
}

export function login(params: LoginRequest): Promise<null> {
  return apiFetch('/api/auth/login', 'POST', params)
}

export function logout(): Promise<null> {
  return apiFetch('/api/auth/logout', 'POST')
}

export function checkAuth(): Promise<User | null> {
  return apiFetch('/api/auth/me', 'GET')
}
