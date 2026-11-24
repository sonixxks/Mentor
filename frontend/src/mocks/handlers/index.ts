import { authHandlers } from './auth'
import { mentorHandlers } from './mentors'
import { profileHandlers } from './profiles'

export const handlers = [...authHandlers, ...profileHandlers, ...mentorHandlers]