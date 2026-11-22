import type { Role } from '../types'

const ACCESS_TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const ROLE_KEY = 'role'

const ROLES: Role[] = ['student', 'mentor', 'curator', 'admin']

export function getToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function getRole(): Role | null {
  const value = localStorage.getItem(ROLE_KEY)
  return ROLES.find((role) => role === value) ?? null
}

// Новый вариант: saveSession(accessToken, refreshToken, role)
export function saveSession(accessToken: string, refreshToken: string, role: Role): void

// Старый вариант для совместимости: saveSession(accessToken, role)
export function saveSession(accessToken: string, role: Role): void

export function saveSession(accessToken: string, refreshTokenOrRole: string, role?: Role): void {
  const refreshToken = role ? refreshTokenOrRole : null
  const userRole = role ?? (refreshTokenOrRole as Role)

  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  localStorage.setItem(ROLE_KEY, userRole)

  if (refreshToken) {
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  } else {
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  }
}

export function updateTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

export function clearSession(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(ROLE_KEY)
}

export const homeByRole: Record<Role, string> = {
  student: '/student',
  mentor: '/mentor',
  curator: '/curator',
  admin: '/curator',
}

export const roleLabel: Record<Role, string> = {
  student: 'Ученик',
  mentor: 'Наставник',
  curator: 'Куратор',
  admin: 'Администратор',
}
