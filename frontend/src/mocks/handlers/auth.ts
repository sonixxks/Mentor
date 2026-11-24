import { http, HttpResponse } from 'msw'
import { API, currentUser, error, save, tokensFor, unauthorized, users } from '../db'
import type { MockUser } from '../db'

type RegisterBody = {
  role: 'student' | 'mentor'
  email: string
  password: string
  first_name: string
  last_name: string
  middle_name?: string
}

/** Регистрации, которые ждут код из письма */
const pending = new Map<string, RegisterBody>()

const codeSent = { expires_in_seconds: 600, resend_after_seconds: 60 }

/** Код 000000 считаем неверным, чтобы проверять ошибки. Любой другой подходит */
const isBadCode = (code: string) => code === '000000'

export const authHandlers = [
  http.post(`${API}/auth/register`, async ({ request }) => {
    const body = (await request.json()) as RegisterBody
    if (users.some((user) => user.email === body.email)) {
      return error(409, 'CONFLICT', 'Аккаунт с такой почтой уже существует')
    }
    pending.set(body.email, body)
    return HttpResponse.json(codeSent)
  }),

  http.post(`${API}/auth/verify-email`, async ({ request }) => {
    const { email, code } = (await request.json()) as { email: string; code: string }
    const data = pending.get(email)
    if (!data || isBadCode(code)) {
      return error(400, 'INVALID_CODE', 'Неверный или просроченный код')
    }
    const user: MockUser = {
      id: crypto.randomUUID(),
      email,
      password: data.password,
      role: data.role,
      profile: {
        first_name: data.first_name,
        last_name: data.last_name,
        middle_name: data.middle_name,
        onboarding_completed: false,
      },
    }
    users.push(user)
    save()
    pending.delete(email)
    return HttpResponse.json(tokensFor(user), { status: 201 })
  }),

  http.post(`${API}/auth/resend-code`, () => HttpResponse.json(codeSent)),

  http.post(`${API}/auth/login`, async ({ request }) => {
    const { email, password } = (await request.json()) as { email: string; password: string }
    const user = users.find((item) => item.email === email && item.password === password)
    if (!user) return error(401, 'UNAUTHORIZED', 'Неверная почта или пароль')
    return HttpResponse.json(tokensFor(user))
  }),

  http.post(`${API}/auth/refresh`, async ({ request }) => {
    const { refresh_token } = (await request.json()) as { refresh_token: string }
    const user = users.find((item) => `refresh-${item.id}` === refresh_token)
    if (!user) return unauthorized()
    return HttpResponse.json(tokensFor(user))
  }),

  http.get(`${API}/users/me`, ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()
    return HttpResponse.json({
      id: user.id,
      email: user.email,
      role: user.role,
      has_password: true,
      password_changed_at: null,
    })
  }),

  http.post(`${API}/auth/password/forgot`, () => HttpResponse.json(codeSent)),

  http.post(`${API}/auth/password/verify-code`, async ({ request }) => {
    const { code } = (await request.json()) as { code: string }
    if (isBadCode(code)) return error(400, 'INVALID_CODE', 'Неверный код')
    return new HttpResponse(null, { status: 204 })
  }),

  http.post(`${API}/auth/password/reset`, async ({ request }) => {
    const body = (await request.json()) as { email: string; code: string; new_password: string }
    const user = users.find((item) => item.email === body.email)
    if (!user || isBadCode(body.code)) return error(400, 'INVALID_CODE', 'Неверный код')
    user.password = body.new_password
    save()
    return HttpResponse.json(tokensFor(user))
  }),
]
