import { apiFetch } from './client'

export type SkillObject = {
  skill: string
  level: 1 | 2 | 3
}

export type AvailabilitySlots = {
  mon: number[]
  tue: number[]
  wed: number[]
  thu: number[]
  fri: number[]
  sat: number[]
  sun: number[]
}

export type UpdateProfileRequest = {
  first_name?: string
  last_name?: string
  middle_name?: string
  age?: number
  current_status?: 'student' | 'graduate' | 'junior' | 'switcher'
  university?: string
  major?: string
  course?: number
  experience?: string
  job_title?: string
  company?: string | null
  bio?: string
  hourly_rate?: number
  max_students?: number
  is_published?: boolean
  communication_style?: 'friendly' | 'structured' | 'free'
  availability_slots?: AvailabilitySlots
  main_goals?: string[]
  success_metric?: string
  time_limit?: number
  needs_curator_help?: boolean
  min_duration_months?: number
  max_duration_months?: number
  current_skills?: SkillObject[]
  target_skills?: string[]
  help_topics?: string[]
  target_student_levels?: string[]
  onboarding_completed?: boolean
}

export type ProfileResponse = UpdateProfileRequest & {
  id: string
  user_id: string
  role: 'student' | 'mentor' | 'curator' | 'admin'
  verification_status: string
}

export type VerificationResponse = {
  verification_status: string
}

export function getSkills() {
  return apiFetch<string[]>('/skills')
}

export function updateMyProfile(data: UpdateProfileRequest) {
  return apiFetch<ProfileResponse>('/profiles/me', {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export function getMyProfile() {
  return apiFetch<ProfileResponse>('/profiles/me')
}

export function addVerificationAssets(files: File[], links: string[]) {
  const formData = new FormData()

  files.forEach((file) => {
    formData.append('files', file)
  })

  links.forEach((link) => {
    formData.append('links', link)
  })

  return apiFetch<VerificationResponse>('/profiles/me/verification-assets', {
    method: 'POST',
    body: formData,
  })
}
