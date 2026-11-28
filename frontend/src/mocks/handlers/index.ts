import { authHandlers } from './auth'
import { mentorHandlers } from './mentors'
import { profileHandlers } from './profiles'
import { requestHandlers } from './requests'
import { workspaceHandlers } from './workspaces'

export const handlers = [
  ...authHandlers,
  ...profileHandlers,
  ...workspaceHandlers,
  ...mentorHandlers,
  ...requestHandlers,
]
