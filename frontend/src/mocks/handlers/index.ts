import { authHandlers } from './auth'
import { mentorHandlers } from './mentors'
import { profileHandlers } from './profiles'
import { requestHandlers } from './requests'

export const handlers = [...authHandlers, ...profileHandlers, ...mentorHandlers, ...requestHandlers]
