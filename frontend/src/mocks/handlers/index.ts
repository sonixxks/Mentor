import { authHandlers } from './auth'
import { profileHandlers } from './profiles'

export const handlers = [...authHandlers, ...profileHandlers]