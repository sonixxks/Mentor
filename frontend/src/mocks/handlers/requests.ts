import { http, HttpResponse } from 'msw'
import type { RequestStatus } from '../../api/requests'
import { MENTORS } from '../data/mentors'
import { ACTIVE, MENTOR_ID_BY_USER, requests, saveRequests } from '../data/requests'
import type { MockRequest } from '../data/requests'
import { API, currentUser, error, save, unauthorized, users } from '../db'

const NEXT: Record<string, Partial<Record<RequestStatus, RequestStatus[]>>> = {
  student: { pending: ['cancelled'], approved_by_mentor: ['accepted'] },
  mentor: { pending: ['approved_by_mentor', 'rejected_by_mentor'] },
  curator: { pending: ['rejected_by_curator'], approved_by_mentor: ['rejected_by_curator'] },
  admin: { pending: ['rejected_by_curator'], approved_by_mentor: ['rejected_by_curator'] },
}

function toResponse(item: MockRequest, forMentor: boolean) {
  const mentor = MENTORS.find((m) => m.id === item.mentor_id)
  const profile = users.find((u) => u.id === item.student_id)?.profile ?? {}
  const program = mentor?.service_packages.find((p) => p.id === item.service_package_id)

  return {
    ...item,
    mentor: mentor
      ? {
          id: mentor.id,
          first_name: mentor.first_name,
          last_name: mentor.last_name,
          avatar_url: null,
        }
      : undefined,
    student: {
      id: item.student_id,
      first_name: profile.first_name,
      target_skills: profile.target_skills,
      main_goals: profile.main_goals,
      time_limit: profile.time_limit,
      min_duration_months: profile.min_duration_months,
      max_duration_months: profile.max_duration_months,
      communication_style: profile.communication_style,
    },
    service_package_title: program?.title,
    match: forMentor ? mentor?.match : undefined,
  }
}

export const requestHandlers = [
  http.post(`${API}/requests`, async ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()
    if (user.role !== 'student') return error(403, 'FORBIDDEN', 'Only students can send requests')

    const body = (await request.json()) as {
      mentor_id?: string
      service_package_id?: string
      cover_letter?: string
    }
    const mentor = MENTORS.find((m) => m.id === body.mentor_id)
    if (!mentor) return error(404, 'NOT_FOUND', 'Mentor not found')

    if ((body.cover_letter ?? '').length > 1000) {
      return error(400, 'VALIDATION_ERROR', 'cover_letter is too long')
    }
    if (
      body.service_package_id &&
      !mentor.service_packages.some((p) => p.id === body.service_package_id)
    ) {
      return error(400, 'VALIDATION_ERROR', 'Package does not belong to this mentor')
    }

    const active = requests.filter((r) => r.student_id === user.id && ACTIVE.includes(r.status))
    if (active.some((r) => r.mentor_id === mentor.id)) {
      return error(409, 'DUPLICATE_REQUEST', 'Active request to this mentor already exists')
    }
    if (active.length >= 3) return error(409, 'REQUEST_LIMIT_REACHED', 'Request limit reached')
    if (mentor.free_slots === 0)
      return error(409, 'MENTOR_LIMIT_REACHED', 'Mentor has no free slots')

    const now = new Date().toISOString()
    const created: MockRequest = {
      id: `req-${Date.now()}`,
      student_id: user.id,
      mentor_id: mentor.id,
      service_package_id: body.service_package_id,
      cover_letter: body.cover_letter?.trim() || undefined,
      status: 'pending',
      created_at: now,
      updated_at: now,
    }
    requests.push(created)
    saveRequests()
    return HttpResponse.json(toResponse(created, false), { status: 201 })
  }),

  http.get(`${API}/requests/my`, ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()

    const params = new URL(request.url).searchParams
    const group = params.get('group')
    const forMentor = user.role === 'mentor'

    let list = requests.filter((r) =>
      forMentor ? r.mentor_id === MENTOR_ID_BY_USER[user.id] : r.student_id === user.id,
    )
    if (group === 'active') list = list.filter((r) => ACTIVE.includes(r.status))
    if (group === 'archive') list = list.filter((r) => !ACTIVE.includes(r.status))
    list = [...list].sort((a, b) => b.created_at.localeCompare(a.created_at))

    const page = Number(params.get('page') ?? 1)
    const limit = Number(params.get('limit') ?? 10)
    const items = list.slice((page - 1) * limit, page * limit).map((r) => toResponse(r, forMentor))
    return HttpResponse.json({ total: list.length, page, items })
  }),

  http.patch(`${API}/requests/:id/status`, async ({ params, request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()

    const target = requests.find((r) => r.id === params.id)
    if (!target) return error(404, 'NOT_FOUND', 'Request not found')

    const isOwner =
      user.role === 'student'
        ? target.student_id === user.id
        : user.role === 'mentor'
          ? target.mentor_id === MENTOR_ID_BY_USER[user.id]
          : true
    if (!isOwner) return error(403, 'FORBIDDEN', 'Not your request')

    const body = (await request.json()) as { status: RequestStatus; rejection_reason?: string }
    const allowed = NEXT[user.role]?.[target.status] ?? []
    if (!allowed.includes(body.status)) {
      return error(409, 'INVALID_TRANSITION', `Cannot change ${target.status} to ${body.status}`)
    }
    if (body.status === 'rejected_by_curator' && !body.rejection_reason?.trim()) {
      return error(400, 'VALIDATION_ERROR', 'rejection_reason is required')
    }

    const now = new Date().toISOString()
    target.status = body.status
    target.rejection_reason = body.rejection_reason?.trim() || undefined
    target.updated_at = now

    if (body.status === 'accepted') {
      target.workspace_id = `ws-${target.id}`
      requests.forEach((r) => {
        if (r !== target && r.student_id === target.student_id && ACTIVE.includes(r.status)) {
          r.status = 'cancelled'
          r.updated_at = now
        }
      })
      const student = users.find((u) => u.id === target.student_id)
      if (student) {
        student.profile.needs_curator_help = false
        save()
      }
    }

    saveRequests()
    return HttpResponse.json(toResponse(target, user.role === 'mentor'))
  }),
]
