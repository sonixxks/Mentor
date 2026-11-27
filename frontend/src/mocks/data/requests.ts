import type { RequestStatus } from '../../api/requests'

export type MockRequest = {
  id: string
  student_id: string
  mentor_id: string
  service_package_id?: string
  cover_letter?: string
  status: RequestStatus
  rejection_reason?: string
  workspace_id?: string
  created_at: string
  updated_at: string
}

export const ACTIVE: RequestStatus[] = ['pending', 'approved_by_mentor']

export const MENTOR_ID_BY_USER: Record<string, string> = { 'u-mentor': 'm-2' }

const KEY = 'mock-requests'

function load(): MockRequest[] {
  try {
    return JSON.parse(localStorage.getItem(KEY) ?? '[]')
  } catch {
    return []
  }
}

export const requests: MockRequest[] = load()

export function saveRequests() {
  localStorage.setItem(KEY, JSON.stringify(requests))
}
