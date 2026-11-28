import { apiFetch } from './client'
import type { WorkspaceFile } from './workspaces'

export type Message = {
  id: string
  type: 'text' | 'system'
  sender_id?: string | null
  text?: string | null
  file?: WorkspaceFile | null
  is_read: boolean
  created_at: string
}

export function getMessages(workspaceId: string, before?: string, limit = 50) {
  const params = new URLSearchParams({ limit: String(limit) })
  if (before) params.set('before', before)
  return apiFetch<Message[]>(`/workspaces/${workspaceId}/messages?${params}`)
}

export function sendMessage(workspaceId: string, text: string) {
  return apiFetch<Message>(`/workspaces/${workspaceId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  })
}

export function markMessagesRead(workspaceId: string) {
  return apiFetch<void>(`/workspaces/${workspaceId}/messages/read`, { method: 'POST' })
}
