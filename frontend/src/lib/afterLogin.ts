import { getMyProfile } from '../api/profiles'
import type { Role } from '../types'
import { homeByRole } from './auth'

/** Куда вести после входа: в анкету, если она не заполнена, иначе в кабинет */
export async function pathAfterLogin(role: Role) {
  if (role !== 'student' && role !== 'mentor') return homeByRole[role]

  const profile = await getMyProfile()
  return profile.onboarding_completed ? homeByRole[role] : `/onboarding/${role}/1`
}