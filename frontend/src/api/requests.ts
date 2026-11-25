import { apiFetch } from './client'
import type { MatchInfo, Page } from './mentors'

export type RequestStatus =
  | 'pending'
  | 'approved_by_mentor'
  | 'accepted'
  | 'rejected_by_mentor'
  | 'rejected_by_curator'
  | 'cancelled'

export type PersonRef = {
  id: string
  first_name: string
  last_name: string
  avatar_url?: string | null
}

export type StudentPublicSummary = {
  id: string
  first_name: string
  target_skills?: string[]
  main_goals?: string[]
  time_limit?: number
  min_duration_months?: number
  max_duration_months?: number
  communication_style?: 'friendly' | 'structured' | 'free'
}

export type MentorRequest = {
  id: string
  student_id: string
  mentor_id: string
  student?: StudentPublicSummary
  mentor?: PersonRef
  service_package_id?: string
  service_package_title?: string
  status: RequestStatus
  cover_letter?: string
  rejection_reason?: string
  workspace_id?: string
  match?: MatchInfo
  created_at: string
  updated_at?: string
}

export type CreateRequestPayload = {
  mentor_id: string
  service_package_id?: string
  cover_letter?: string
}

export function createRequest(payload: CreateRequestPayload) {
  return apiFetch<MentorRequest>('/requests', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function getMyRequests(group?: 'active' | 'archive') {
  const params = new URLSearchParams({ limit: '50' })
  if (group) params.set('group', group)
  return apiFetch<Page<MentorRequest>>(`/requests/my?${params}`)
}

export function changeRequestStatus(id: string, status: RequestStatus, rejection_reason?: string) {
  return apiFetch<MentorRequest>(`/requests/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status, rejection_reason }),
  })
}
