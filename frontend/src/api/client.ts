import type { APIError } from '@/types/api'

// Default to empty string for relative paths (Nginx proxy)
const API_BASE = import.meta.env.VITE_API ?? ''
const API_KEY = import.meta.env.VITE_API_KEY ?? ''

export function apiFetch<T = unknown>(path: string, init?: RequestInit): Promise<T> {
  const url = path.startsWith('http') ? path : `${API_BASE}${path}`
  const headers = new Headers(init?.headers)
  if (API_KEY) {
    headers.set('X-API-Key', API_KEY)
  }
  headers.set('Content-Type', 'application/json')
  return fetch(url, { ...init, headers })
    .then(async (res) => {
      if (!res.ok) {
        const err: APIError = await res.json().catch(() => ({ error: `HTTP ${res.status}` }))
        throw new Error(err.error || err.message || `HTTP ${res.status}`)
      }
      return res.json()
    })
}

export function sseUrl(path: string, params?: Record<string, string>): string {
  const fullPath = path.startsWith('http') ? path : `${API_BASE}${path}`
  // Use window.location.origin as base for relative URLs to prevent Invalid URL error
  const url = new URL(fullPath, window.location.origin)
  if (API_KEY) {
    url.searchParams.set('apiKey', API_KEY)
  }
  if (params) {
    Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v))
  }
  return url.toString()
}
