import { apiFetch } from './client'
import type { Page } from './mentors'
import type { PersonRef } from './requests'

export type WorkspaceStatus = 'active' | 'paused' | 'stagnation' | 'completed'

export type Meeting = {
  id: string
  workspace_id: string
  title: string
  start_time: string
  end_time?: string
  jitsi_link: string
  my_review_rating?: number | null
}

export type WorkspaceSummary = {
  id: string
  status: WorkspaceStatus
  partner: PersonRef
  progress_percent: number
  started_at: string
  unread_messages?: number
  tasks_on_review?: number
  inactive_days?: number | null
}

export type WorkspaceDetail = WorkspaceSummary & {
  goal?: string
  service_package_title?: string
  jitsi_link?: string
  sessions_done?: number
  sessions_total?: number | null
  tasks_total?: number
  tasks_completed?: number
  next_meeting?: Meeting
}

export type TaskStatus = 'todo' | 'in_progress' | 'on_review' | 'needs_revision' | 'completed'

export type Task = {
  id: string
  title: string
  description?: string
  status: TaskStatus
  deadline_at: string
  created_at: string
  updated_at?: string
  completed_at?: string | null
  artifacts_url?: string
  mentor_comment?: string
  comments_count?: number
}

export type TaskStatusUpdate = {
  status: Exclude<TaskStatus, 'todo'>
  artifacts_url?: string
  mentor_comment?: string
}

export type TaskComment = {
  id: string
  task_id: string
  author: PersonRef
  text: string
  created_at: string
}

export type WorkspaceFile = {
  id: string
  type: 'file' | 'link'
  file_name: string
  file_url: string
  file_size_kb?: number | null
  uploaded_by: string
  created_at: string
}

export function getMyWorkspaces() {
  return apiFetch<Page<WorkspaceSummary>>('/workspaces?limit=50')
}

export function getWorkspace(id: string) {
  return apiFetch<WorkspaceDetail>(`/workspaces/${id}`)
}

export function getTasks(workspaceId: string) {
  return apiFetch<Task[]>(`/workspaces/${workspaceId}/tasks`)
}

export function changeTaskStatus(taskId: string, update: TaskStatusUpdate) {
  return apiFetch<Task>(`/tasks/${taskId}/status`, {
    method: 'PATCH',
    body: JSON.stringify(update),
  })
}

export function getTaskComments(taskId: string) {
  return apiFetch<TaskComment[]>(`/tasks/${taskId}/comments`)
}

export function addTaskComment(taskId: string, text: string) {
  return apiFetch<TaskComment>(`/tasks/${taskId}/comments`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  })
}

export function getMeetings(workspaceId: string) {
  return apiFetch<Meeting[]>(`/workspaces/${workspaceId}/meetings`)
}

export function reviewMeeting(meetingId: string, rating: number, comment?: string) {
  return apiFetch<{ id: string; rating: number }>(`/meetings/${meetingId}/reviews`, {
    method: 'POST',
    body: JSON.stringify({ rating, comment }),
  })
}

export function getWorkspaceFiles(workspaceId: string) {
  return apiFetch<WorkspaceFile[]>(`/workspaces/${workspaceId}/files`)
}
