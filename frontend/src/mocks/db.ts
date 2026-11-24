import { HttpResponse } from 'msw'
import type { Role } from '../types'

export const API = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export type MockUser = {
  id: string
  email: string
  password: string
  role: Role
  profile: Record<string, unknown>
}

const SEED: MockUser[] = [
  {
    id: 'u-student',
    email: 'student@test.ru',
    password: 'password',
    role: 'student',
    profile: { first_name: 'Екатерина', last_name: 'Смирнова', onboarding_completed: true },
  },
  {
    id: 'u-mentor',
    email: 'mentor@test.ru',
    password: 'password',
    role: 'mentor',
    profile: { first_name: 'Данил', last_name: 'Королёв', onboarding_completed: true },
  },
  {
    id: 'u-curator',
    email: 'curator@test.ru',
    password: 'password',
    role: 'curator',
    profile: { first_name: 'Анна', last_name: 'Белова', onboarding_completed: true },
  },
]

const KEY = 'mock-users'

function load(): MockUser[] {
  try {
    const saved = localStorage.getItem(KEY)
    return saved ? JSON.parse(saved) : SEED
  } catch {
    return SEED
  }
}

export const users: MockUser[] = load()

export function save() {
  localStorage.setItem(KEY, JSON.stringify(users))
}

export function tokensFor(user: MockUser) {
  return { access_token: `access-${user.id}`, refresh_token: `refresh-${user.id}` }
}

export function currentUser(request: Request) {
  const token = request.headers.get('Authorization')?.replace('Bearer ', '')
  return users.find((user) => `access-${user.id}` === token)
}

export function error(status: number, error_code: string, message: string) {
  return HttpResponse.json({ error_code, message }, { status })
}

export const unauthorized = () => error(401, 'UNAUTHORIZED', 'Требуется вход')
