import { apiFetch } from './client'
import type { AvailabilitySlots } from './profiles'
import type { PersonRef } from './requests'
import type { Meeting, TaskStatus } from './workspaces'

export type TaskDeadline = {
  task_id: string
  workspace_id: string
  title: string
  deadline_at: string
  status: TaskStatus
}

export type PartnerAvailability = {
  workspace_id: string
  partner: PersonRef
  availability_slots: AvailabilitySlots
}

export type CalendarData = {
  meetings: Meeting[]
  availability_slots?: AvailabilitySlots
  task_deadlines: TaskDeadline[]
  partners_availability: PartnerAvailability[]
}

export function getCalendar(from: string, to: string) {
  return apiFetch<CalendarData>(`/calendar?from=${from}&to=${to}`)
}
