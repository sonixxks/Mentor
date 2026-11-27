
import { apiFetch } from './client'
import type { PublicRole, Role } from '../types'

export type RegisterRequest = {
  role: PublicRole
  email: string
  password: string
  first_name: string
  last_name: string
  middle_name?: string
  terms_accepted: true
}

export type LoginRequest = {
  email: string
  password: string
  remember_me: boolean
}

export type TokenResponse = {
  access_token: string
  refresh_token: string
}

export type CodeSentResponse = {
  expires_in_seconds: number
  resend_after_seconds: number
}

export type UserMeResponse = {
  id: string
  email: string
  role: Role
  has_password: boolean
  password_changed_at?: string | null
}

export function register(data: RegisterRequest) {
  return apiFetch<CodeSentResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function verifyEmail(email: string, code: string) {
  return apiFetch<TokenResponse>('/auth/verify-email', {
    method: 'POST',
    body: JSON.stringify({ email, code }),
  })
}

export function resendEmailCode(email: string) {
  return apiFetch<CodeSentResponse>('/auth/resend-code', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

export function login(data: LoginRequest) {
  return apiFetch<TokenResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function getMyAccount() {
  return apiFetch<UserMeResponse>('/users/me')
}

export function refreshTokens(refreshToken: string) {
  return apiFetch<TokenResponse>('/auth/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export function requestPasswordReset(email: string) {
  return apiFetch<CodeSentResponse>('/auth/password/forgot', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

export function verifyPasswordResetCode(email: string, code: string) {
  return apiFetch<void>('/auth/password/verify-code', {
    method: 'POST',
    body: JSON.stringify({ email, code }),
  })
}

export function resetPassword(email: string, code: string, newPassword: string) {
  return apiFetch<TokenResponse>('/auth/password/reset', {
    method: 'POST',
    body: JSON.stringify({
      email,
      code,
      new_password: newPassword,
    }),
  })
}
