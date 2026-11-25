import { apiFetch } from './client'
import type { AvailabilitySlots, SkillObject } from './profiles'

export type Experience = '1_to_3' | '3_to_5' | '5_to_10' | 'over_10'
export type CommunicationStyle = 'friendly' | 'structured' | 'free'

export type MatchInfo = {
  percent: number
  breakdown: { skills: number; goals: number; time: number; style: number }
}

export type MentorSummary = {
  id: string
  first_name: string
  last_name: string
  avatar_url: string | null
  job_title?: string
  company: string | null
  experience?: Experience
  bio?: string
  hourly_rate?: number
  verification_status: string
  current_skills: SkillObject[]
  rating_avg: number | null
  reviews_count: number
  free_slots: number
  time_limit?: number
  communication_style?: CommunicationStyle
  match?: MatchInfo
}

export type ServicePackage = {
  id: string
  title: string
  description?: string
  result?: string
  price?: number
  session_count?: number
  skills: string[]
  status: string
  pairs_count: number
}

export type MentorDetail = MentorSummary & {
  help_topics: string[]
  target_student_levels: string[]
  availability_slots?: AvailabilitySlots
  max_students?: number
  service_packages: ServicePackage[]
}

export type Review = {
  id: string
  rating: number
  comment?: string
  author_first_name: string
  meeting_date?: string
  created_at: string
}

export type MentorStats = {
  rating_avg: number | null
  reviews_count: number
  active_students: number
  max_students: number
  free_slots: number
  completed_pairs: number
  meetings_this_month: number
}

export type Page<T> = { total: number; page: number; items: T[] }

export type MentorSort = 'match' | 'rating' | 'price_asc' | 'price_desc' | 'experience'

export type MentorFilters = {
  page?: number
  limit?: number
  q?: string
  skills?: string[]
  sort?: MentorSort
  price_min?: number
  price_max?: number
  hours_min?: number
  hours_max?: number
  min_match?: number
  min_rating?: number
  has_free_slots?: boolean
}

function toQuery(filters: MentorFilters) {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([key, value]) => {
    if (value === undefined || value === '' || value === false) return
    if (Array.isArray(value)) {
      if (value.length > 0) params.set(key, value.join(','))
      return
    }
    params.set(key, String(value))
  })
  return params.toString()
}

export function getMentors(filters: MentorFilters) {
  return apiFetch<Page<MentorSummary>>(`/mentors?${toQuery(filters)}`)
}

export function getRecommendations() {
  return apiFetch<MentorSummary[]>('/mentors/recommendations')
}

export function getMentor(id: string) {
  return apiFetch<MentorDetail>(`/mentors/${id}`)
}

export function getMentorReviews(id: string, page = 1) {
  return apiFetch<Page<Review>>(`/mentors/${id}/reviews?page=${page}&limit=10`)
}

export function getMyMentorStats() {
  return apiFetch<MentorStats>('/mentors/me/stats')
}
