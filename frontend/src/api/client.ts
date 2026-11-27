import { clearSession, getToken, getRefreshToken, updateTokens } from '../lib/auth'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export class ApiError extends Error {
  status: number
  code?: string

  constructor(status: number, message: string, code?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

type ApiErrorBody = {
  message?: string
  error_code?: string
  details?: string[]
}

type TokenResponse = {
  access_token: string
  refresh_token: string
}

let refreshPromise: Promise<string | null> | null = null

const PUBLIC_AUTH_PATHS = [
  '/auth/register',
  '/auth/verify-email',
  '/auth/resend-code',
  '/auth/login',
  '/auth/refresh',
  '/auth/password/forgot',
  '/auth/password/verify-code',
  '/auth/password/reset',
]

function isPublicAuthPath(path: string): boolean {
  return PUBLIC_AUTH_PATHS.includes(path)
}

function getErrorMessage(body: ApiErrorBody | null): string {
  return body?.message || 'Не удалось выполнить запрос'
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken()

  if (!refreshToken) {
    clearSession()
    return null
  }

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        refresh_token: refreshToken,
      }),
    })

    const data = (await response.json().catch(() => null)) as TokenResponse | ApiErrorBody | null

    if (!response.ok) {
      clearSession()
      return null
    }

    const tokens = data as TokenResponse | null

    if (!tokens?.access_token || !tokens?.refresh_token) {
      clearSession()
      return null
    }

    updateTokens(tokens.access_token, tokens.refresh_token)

    return tokens.access_token
  } catch {
    throw new ApiError(0, 'Не удалось обновить сессию. Проверьте подключение к интернету.')
  }
}

async function getNewAccessToken(): Promise<string | null> {
  if (!refreshPromise) {
    refreshPromise = refreshAccessToken().finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

async function sendRequest(
  url: string,
  init: RequestInit,
  token: string | null,
): Promise<Response> {
  const headers = new Headers(init.headers)

  const isFormData = typeof FormData !== 'undefined' && init.body instanceof FormData

  if (init.body && !isFormData && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  headers.set('Accept', 'application/json')

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  } else {
    headers.delete('Authorization')
  }

  return fetch(url, {
    ...init,
    headers,
  })
}

async function parseResponse<T>(response: Response): Promise<T> {
  if (response.status === 204) {
    return undefined as T
  }

  const responseBody = await response.json().catch(() => null)

  if (!response.ok) {
    const body = responseBody as ApiErrorBody | null

    throw new ApiError(response.status, getErrorMessage(body), body?.error_code)
  }

  return responseBody as T
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const url = `${API_BASE_URL}${normalizedPath}`

  const token = getToken()

  let response = await sendRequest(url, init, token)

  const canRefresh = response.status === 401 && Boolean(token) && !isPublicAuthPath(normalizedPath)

  if (canRefresh) {
    const newToken = await getNewAccessToken()

    if (newToken) {
      response = await sendRequest(url, init, newToken)

      if (response.status === 401) {
        clearSession()
      }
    } else {
      clearSession()
    }
  }

  return parseResponse<T>(response)
}
