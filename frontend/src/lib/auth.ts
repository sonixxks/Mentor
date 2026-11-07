/** Пока нет бэкенда, токен и роль хранятся в браузере (localStorage) */

import type { Role } from '../types'

const TOKEN_KEY = 'token'
const ROLE_KEY = 'role'
const ROLES: Role[] = ['student', 'mentor', 'curator', 'admin']

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function getRole(): Role | null {
  const value = localStorage.getItem(ROLE_KEY)
  return ROLES.find((role) => role === value) ?? null
}

export function saveSession(token: string, role: Role) {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(ROLE_KEY, role)
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY)
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