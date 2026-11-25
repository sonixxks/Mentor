import { http, HttpResponse } from 'msw'
import type { TaskStatus, TaskStatusUpdate } from '../../api/workspaces'
import { MENTORS } from '../data/mentors'
import { MENTOR_ID_BY_USER } from '../data/requests'
import { saveStore, store } from '../data/workspaces'
import type { MockMeeting, MockTask, MockWorkspace } from '../data/workspaces'
import { API, currentUser, error, unauthorized, users } from '../db'
import type { MockUser } from '../db'

const DAY = 24 * 60 * 60 * 1000

const NEXT_TASK: Record<string, Partial<Record<TaskStatus, TaskStatus[]>>> = {
  student: { todo: ['in_progress'], in_progress: ['on_review'], needs_revision: ['on_review'] },
  mentor: { on_review: ['completed', 'needs_revision'] },
}

function isMember(user: MockUser, ws: MockWorkspace) {
  if (user.role === 'student') return ws.student_id === user.id
  if (user.role === 'mentor') return ws.mentor_id === MENTOR_ID_BY_USER[user.id]
  return user.role === 'curator' || user.role === 'admin'
}

function mentorRef(mentorId: string) {
  const mentor = MENTORS.find((m) => m.id === mentorId)
  return {
    id: mentorId,
    first_name: mentor?.first_name ?? 'Наставник',
    last_name: mentor?.last_name ?? '',
    avatar_url: null,
  }
}

function studentRef(studentId: string) {
  const profile = users.find((u) => u.id === studentId)?.profile ?? {}
  return {
    id: studentId,
    first_name: String(profile.first_name ?? 'Ученик'),
    last_name: String(profile.last_name ?? ''),
    avatar_url: null,
  }
}

function personOf(user: MockUser) {
  return user.role === 'mentor' ? mentorRef(MENTOR_ID_BY_USER[user.id]) : studentRef(user.id)
}

const jitsiLink = (ws: MockWorkspace) => `https://meet.jit.si/kurator-${ws.id}`

function meetingResponse(meeting: MockMeeting, ws: MockWorkspace, user: MockUser) {
  const { reviews, ...rest } = meeting
  return { ...rest, jitsi_link: jitsiLink(ws), my_review_rating: reviews[user.id] ?? null }
}

function taskResponse(task: MockTask) {
  const { workspace_id, ...rest } = task
  return {
    ...rest,
    comments_count: store.comments.filter((c) => c.task_id === task.id).length,
  }
}

function summary(ws: MockWorkspace, user: MockUser) {
  const tasks = store.tasks.filter((t) => t.workspace_id === ws.id)
  const done = tasks.filter((t) => t.status === 'completed').length
  return {
    id: ws.id,
    status: ws.status,
    partner: user.role === 'student' ? mentorRef(ws.mentor_id) : studentRef(ws.student_id),
    progress_percent: tasks.length ? Math.round((done / tasks.length) * 100) : 0,
    started_at: ws.started_at,
    unread_messages: 0,
    tasks_on_review: tasks.filter((t) => t.status === 'on_review').length,
    inactive_days: null,
  }
}

function detail(ws: MockWorkspace, user: MockUser) {
  const now = Date.now()
  const tasks = store.tasks.filter((t) => t.workspace_id === ws.id)
  const meetings = store.meetings.filter((m) => m.workspace_id === ws.id)
  const next = meetings
    .filter((m) => Date.parse(m.start_time) > now)
    .sort((a, b) => a.start_time.localeCompare(b.start_time))[0]
  const program = MENTORS.find((m) => m.id === ws.mentor_id)?.service_packages.find(
    (p) => p.id === ws.service_package_id,
  )
  const profile = users.find((u) => u.id === ws.student_id)?.profile ?? {}

  return {
    ...summary(ws, user),
    goal: profile.success_metric as string | undefined,
    service_package_title: program?.title,
    jitsi_link: jitsiLink(ws),
    sessions_done: meetings.filter((m) => Date.parse(m.start_time) <= now).length,
    sessions_total: program?.session_count ?? null,
    tasks_total: tasks.length,
    tasks_completed: tasks.filter((t) => t.status === 'completed').length,
    next_meeting: next ? meetingResponse(next, ws, user) : undefined,
  }
}

function access(request: Request, workspaceId: string) {
  const user = currentUser(request)
  if (!user) return { fail: unauthorized() }
  const ws = store.workspaces.find((w) => w.id === workspaceId)
  if (!ws) return { fail: error(404, 'NOT_FOUND', 'Workspace not found') }
  if (!isMember(user, ws))
    return { fail: error(403, 'FORBIDDEN', 'Not a member of this workspace') }
  return { user, ws }
}

function isUrl(value: string) {
  try {
    new URL(value)
    return true
  } catch {
    return false
  }
}

const isDate = (value?: string) => Boolean(value) && !Number.isNaN(Date.parse(value as string))

export const workspaceHandlers = [
  http.get(`${API}/mentors/me/stats`, ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()
    if (user.role !== 'mentor') return error(403, 'FORBIDDEN', 'Mentors only')

    const mentorId = MENTOR_ID_BY_USER[user.id]
    const mentor = MENTORS.find((m) => m.id === mentorId)
    const mine = store.workspaces.filter((ws) => ws.mentor_id === mentorId)
    const active = mine.filter((ws) => ws.status !== 'completed').length
    const max = mentor?.max_students ?? 0
    const monthAgo = Date.now() - 30 * DAY
    const meetings = store.meetings.filter((m) => {
      const start = Date.parse(m.start_time)
      return mine.some((ws) => ws.id === m.workspace_id) && start <= Date.now() && start >= monthAgo
    })

    return HttpResponse.json({
      rating_avg: mentor?.rating_avg ?? null,
      reviews_count: mentor?.reviews_count ?? 0,
      active_students: active,
      max_students: max,
      free_slots: Math.max(max - active, 0),
      completed_pairs: mine.length - active,
      meetings_this_month: meetings.length,
    })
  }),

  http.get(`${API}/workspaces`, ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()
    const items = store.workspaces.filter((ws) => isMember(user, ws)).map((ws) => summary(ws, user))
    return HttpResponse.json({ total: items.length, page: 1, items })
  }),

  http.get(`${API}/workspaces/:id`, ({ params, request }) => {
    const a = access(request, String(params.id))
    if ('fail' in a) return a.fail
    return HttpResponse.json(detail(a.ws, a.user))
  }),

  http.get(`${API}/workspaces/:id/tasks`, ({ params, request }) => {
    const a = access(request, String(params.id))
    if ('fail' in a) return a.fail
    const tasks = store.tasks
      .filter((t) => t.workspace_id === a.ws.id)
      .sort((x, y) => x.deadline_at.localeCompare(y.deadline_at))
      .map(taskResponse)
    return HttpResponse.json(tasks)
  }),

  http.post(`${API}/workspaces/:id/tasks`, async ({ params, request }) => {
    const a = access(request, String(params.id))
    if ('fail' in a) return a.fail
    if (a.user.role !== 'mentor') return error(403, 'FORBIDDEN', 'Only the mentor can create tasks')

    const body = (await request.json()) as {
      title?: string
      description?: string
      deadline_at?: string
    }
    const title = body.title?.trim() ?? ''
    if (!title) return error(400, 'VALIDATION_ERROR', 'title is required')
    if (!isDate(body.deadline_at)) return error(400, 'VALIDATION_ERROR', 'deadline_at is required')

    const now = new Date().toISOString()
    const task: MockTask = {
      id: `t-${Date.now()}`,
      workspace_id: a.ws.id,
      title,
      description: body.description?.trim() || undefined,
      status: 'todo',
      deadline_at: body.deadline_at as string,
      created_at: now,
      updated_at: now,
      completed_at: null,
    }
    store.tasks.push(task)
    saveStore()
    return HttpResponse.json(taskResponse(task), { status: 201 })
  }),

  http.patch(`${API}/tasks/:taskId`, async ({ params, request }) => {
    const task = store.tasks.find((t) => t.id === params.taskId)
    if (!task) return error(404, 'NOT_FOUND', 'Task not found')
    const a = access(request, task.workspace_id)
    if ('fail' in a) return a.fail
    if (a.user.role !== 'mentor') return error(403, 'FORBIDDEN', 'Only the mentor can edit tasks')

    const body = (await request.json()) as {
      title?: string
      description?: string
      deadline_at?: string
    }
    if (body.title !== undefined && !body.title.trim()) {
      return error(400, 'VALIDATION_ERROR', 'title cannot be empty')
    }
    if (body.deadline_at !== undefined && !isDate(body.deadline_at)) {
      return error(400, 'VALIDATION_ERROR', 'deadline_at is invalid')
    }

    if (body.title !== undefined) task.title = body.title.trim()
    if (body.description !== undefined) task.description = body.description.trim() || undefined
    if (body.deadline_at !== undefined) task.deadline_at = body.deadline_at
    task.updated_at = new Date().toISOString()
    saveStore()
    return HttpResponse.json(taskResponse(task))
  }),

  http.delete(`${API}/tasks/:taskId`, ({ params, request }) => {
    const task = store.tasks.find((t) => t.id === params.taskId)
    if (!task) return error(404, 'NOT_FOUND', 'Task not found')
    const a = access(request, task.workspace_id)
    if ('fail' in a) return a.fail
    if (a.user.role !== 'mentor') return error(403, 'FORBIDDEN', 'Only the mentor can delete tasks')

    store.tasks = store.tasks.filter((t) => t.id !== task.id)
    store.comments = store.comments.filter((c) => c.task_id !== task.id)
    saveStore()
    return new HttpResponse(null, { status: 204 })
  }),

  http.patch(`${API}/tasks/:taskId/status`, async ({ params, request }) => {
    const task = store.tasks.find((t) => t.id === params.taskId)
    if (!task) return error(404, 'NOT_FOUND', 'Task not found')
    const a = access(request, task.workspace_id)
    if ('fail' in a) return a.fail

    const body = (await request.json()) as TaskStatusUpdate
    const allowed = NEXT_TASK[a.user.role]?.[task.status] ?? []
    if (!allowed.includes(body.status)) {
      return error(409, 'INVALID_TRANSITION', `Cannot change ${task.status} to ${body.status}`)
    }
    if (body.artifacts_url && !isUrl(body.artifacts_url)) {
      return error(400, 'VALIDATION_ERROR', 'artifacts_url must be a valid URL')
    }

    const now = new Date().toISOString()
    task.status = body.status
    task.updated_at = now
    if (body.artifacts_url) task.artifacts_url = body.artifacts_url
    if (body.mentor_comment !== undefined) task.mentor_comment = body.mentor_comment
    task.completed_at = body.status === 'completed' ? now : null
    saveStore()
    return HttpResponse.json(taskResponse(task))
  }),

  http.get(`${API}/tasks/:taskId/comments`, ({ params, request }) => {
    const task = store.tasks.find((t) => t.id === params.taskId)
    if (!task) return error(404, 'NOT_FOUND', 'Task not found')
    const a = access(request, task.workspace_id)
    if ('fail' in a) return a.fail
    const comments = store.comments
      .filter((c) => c.task_id === task.id)
      .sort((x, y) => x.created_at.localeCompare(y.created_at))
    return HttpResponse.json(comments)
  }),

  http.post(`${API}/tasks/:taskId/comments`, async ({ params, request }) => {
    const task = store.tasks.find((t) => t.id === params.taskId)
    if (!task) return error(404, 'NOT_FOUND', 'Task not found')
    const a = access(request, task.workspace_id)
    if ('fail' in a) return a.fail

    const { text } = (await request.json()) as { text?: string }
    const clean = text?.trim() ?? ''
    if (clean.length === 0 || clean.length > 2000) {
      return error(400, 'VALIDATION_ERROR', 'text must be 1-2000 characters')
    }

    const comment = {
      id: `c-${Date.now()}`,
      task_id: task.id,
      author: personOf(a.user),
      text: clean,
      created_at: new Date().toISOString(),
    }
    store.comments.push(comment)
    saveStore()
    return HttpResponse.json(comment, { status: 201 })
  }),

  http.get(`${API}/workspaces/:id/meetings`, ({ params, request }) => {
    const a = access(request, String(params.id))
    if ('fail' in a) return a.fail
    const meetings = store.meetings
      .filter((m) => m.workspace_id === a.ws.id)
      .sort((x, y) => x.start_time.localeCompare(y.start_time))
      .map((m) => meetingResponse(m, a.ws, a.user))
    return HttpResponse.json(meetings)
  }),

  http.post(`${API}/workspaces/:id/meetings`, async ({ params, request }) => {
    const a = access(request, String(params.id))
    if ('fail' in a) return a.fail

    const body = (await request.json()) as {
      title?: string
      start_time?: string
      end_time?: string
    }
    if (!isDate(body.start_time)) return error(400, 'VALIDATION_ERROR', 'start_time is required')
    if (body.end_time && (!isDate(body.end_time) || body.end_time <= (body.start_time as string))) {
      return error(400, 'VALIDATION_ERROR', 'end_time must be after start_time')
    }

    const meeting: MockMeeting = {
      id: `m-${Date.now()}`,
      workspace_id: a.ws.id,
      title: body.title?.trim() || 'Встреча',
      start_time: body.start_time as string,
      end_time: body.end_time,
      reviews: {},
    }
    store.meetings.push(meeting)
    saveStore()
    return HttpResponse.json(meetingResponse(meeting, a.ws, a.user), { status: 201 })
  }),

  http.patch(`${API}/meetings/:meetingId`, async ({ params, request }) => {
    const meeting = store.meetings.find((m) => m.id === params.meetingId)
    if (!meeting) return error(404, 'NOT_FOUND', 'Meeting not found')
    const a = access(request, meeting.workspace_id)
    if ('fail' in a) return a.fail

    const body = (await request.json()) as {
      title?: string
      start_time?: string
      end_time?: string
    }
    if (body.start_time !== undefined && !isDate(body.start_time)) {
      return error(400, 'VALIDATION_ERROR', 'start_time is invalid')
    }

    if (body.title !== undefined) meeting.title = body.title.trim() || 'Встреча'
    if (body.start_time !== undefined) meeting.start_time = body.start_time
    if (body.end_time !== undefined) meeting.end_time = body.end_time
    saveStore()
    return HttpResponse.json(meetingResponse(meeting, a.ws, a.user))
  }),

  http.post(`${API}/meetings/:meetingId/reviews`, async ({ params, request }) => {
    const meeting = store.meetings.find((m) => m.id === params.meetingId)
    if (!meeting) return error(404, 'NOT_FOUND', 'Meeting not found')
    const a = access(request, meeting.workspace_id)
    if ('fail' in a) return a.fail

    const { rating, comment } = (await request.json()) as { rating?: number; comment?: string }
    if (!rating || rating < 1 || rating > 5) {
      return error(400, 'VALIDATION_ERROR', 'rating must be 1-5')
    }
    if (Date.parse(meeting.start_time) > Date.now()) {
      return error(400, 'VALIDATION_ERROR', 'Meeting has not started yet')
    }
    if (meeting.reviews[a.user.id]) {
      return error(409, 'REVIEW_EXISTS', 'Review already exists')
    }

    meeting.reviews[a.user.id] = rating
    saveStore()
    return HttpResponse.json(
      {
        id: `rv-${Date.now()}`,
        rating,
        comment,
        author_first_name: personOf(a.user).first_name,
        meeting_date: meeting.start_time,
        created_at: new Date().toISOString(),
      },
      { status: 201 },
    )
  }),

  http.get(`${API}/workspaces/:id/files`, ({ params, request }) => {
    const a = access(request, String(params.id))
    if ('fail' in a) return a.fail
    const files = store.files
      .filter((f) => f.workspace_id === a.ws.id)
      .map(({ workspace_id, ...file }) => file)
    return HttpResponse.json(files)
  }),
]
