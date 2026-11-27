import { http, HttpResponse } from 'msw'
import { SKILLS } from '../data/skills'
import { API, currentUser, save, unauthorized } from '../db'
import type { MockUser } from '../db'

function profileOf(user: MockUser) {
  return {
    id: user.id,
    user_id: user.id,
    role: user.role,
    verification_status: 'none',
    ...user.profile,
  }
}

export const profileHandlers = [
  http.get(`${API}/profiles/me`, ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()
    return HttpResponse.json(profileOf(user))
  }),

  http.patch(`${API}/profiles/me`, async ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()
    const patch = (await request.json()) as Record<string, unknown>

    user.profile = { ...user.profile, ...patch }
    // Наставник при публикации уходит на проверку куратору
    if (user.role === 'mentor' && patch.is_published === true) {
      user.profile.verification_status = 'pending'
    }
    save()
    return HttpResponse.json(profileOf(user))
  }),

  http.post(`${API}/profiles/me/verification-assets`, async ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()
    const form = await request.formData()
    const assets = [
      ...form.getAll('files').map((file) => ({
        id: crypto.randomUUID(),
        type: 'file',
        name: (file as File).name,
        url: 'https://example.com/file.pdf',
      })),
      ...form.getAll('links').map((link) => ({
        id: crypto.randomUUID(),
        type: 'link',
        name: String(link),
        url: String(link),
      })),
    ]
    return HttpResponse.json({ verification_status: 'none', verification_assets: assets })
  }),

  http.get(`${API}/skills`, ({ request }) => {
    const url = new URL(request.url)
    const q = url.searchParams.get('q')?.toLowerCase() ?? ''
    const limit = Number(url.searchParams.get('limit') ?? 20)
    const result = SKILLS.filter((name) => name.toLowerCase().includes(q)).slice(0, limit)
    return HttpResponse.json(result)
  }),
]