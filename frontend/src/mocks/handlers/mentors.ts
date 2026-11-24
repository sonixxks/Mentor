import { http, HttpResponse } from 'msw'
import type { MentorDetail } from '../../api/mentors'
import { MENTORS, REVIEWS } from '../data/mentors'
import { API, currentUser, error, unauthorized } from '../db'

const EXPERIENCE_ORDER = ['1_to_3', '3_to_5', '5_to_10', 'over_10']

/** Полный профиль → карточка каталога. match показываем только ученику */
function toSummary(mentor: MentorDetail, withMatch: boolean) {
  const {
    help_topics,
    target_student_levels,
    availability_slots,
    max_students,
    service_packages,
    match,
    ...summary
  } = mentor
  return withMatch ? { ...summary, match } : summary
}

export const mentorHandlers = [
  // Этот адрес раньше /mentors/:id, иначе «recommendations» примется за id
  http.get(`${API}/mentors/recommendations`, ({ request }) => {
    const user = currentUser(request)
    if (!user) return unauthorized()
    const result = MENTORS.filter((m) => (m.match?.percent ?? 0) >= 60 && m.free_slots > 0)
      .sort((a, b) => (b.match?.percent ?? 0) - (a.match?.percent ?? 0))
      .slice(0, 10)
      .map((m) => toSummary(m, true))
    return HttpResponse.json(result)
  }),

  http.get(`${API}/mentors`, ({ request }) => {
    const isStudent = currentUser(request)?.role === 'student'
    const params = new URL(request.url).searchParams
    const num = (key: string) => (params.has(key) ? Number(params.get(key)) : undefined)

    const q = params.get('q')?.toLowerCase() ?? ''
    const skills = params.get('skills')?.split(',').filter(Boolean) ?? []
    const priceMin = num('price_min')
    const priceMax = num('price_max')
    const hoursMin = num('hours_min')
    const hoursMax = num('hours_max')
    const minMatch = num('min_match')
    const minRating = num('min_rating')
    const onlyFree = params.get('has_free_slots') === 'true'

    let list = MENTORS.filter((m) => {
      const names = m.current_skills.map((s) => s.skill)
      const price = m.hourly_rate ?? 0
      const hours = m.time_limit ?? 0
      const text = [m.first_name, m.last_name, m.company ?? '', ...names].join(' ').toLowerCase()

      if (q && !text.includes(q)) return false
      if (!skills.every((skill) => names.includes(skill))) return false
      if (priceMin !== undefined && price < priceMin) return false
      if (priceMax !== undefined && price > priceMax) return false
      if (hoursMin !== undefined && hours < hoursMin) return false
      if (hoursMax !== undefined && hours > hoursMax) return false
      if (isStudent && minMatch !== undefined && (m.match?.percent ?? 0) < minMatch) return false
      if (minRating !== undefined && (m.rating_avg ?? 0) < minRating) return false
      if (onlyFree && m.free_slots === 0) return false
      return true
    })

    const sort = params.get('sort') ?? (isStudent ? 'match' : 'rating')
    list = [...list].sort((a, b) => {
      if (sort === 'match' && isStudent) return (b.match?.percent ?? 0) - (a.match?.percent ?? 0)
      if (sort === 'price_asc') return (a.hourly_rate ?? 0) - (b.hourly_rate ?? 0)
      if (sort === 'price_desc') return (b.hourly_rate ?? 0) - (a.hourly_rate ?? 0)
      if (sort === 'experience')
        return (
          EXPERIENCE_ORDER.indexOf(b.experience ?? '') -
          EXPERIENCE_ORDER.indexOf(a.experience ?? '')
        )
      return (b.rating_avg ?? 0) - (a.rating_avg ?? 0)
    })

    const page = num('page') ?? 1
    const limit = num('limit') ?? 10
    const items = list.slice((page - 1) * limit, page * limit).map((m) => toSummary(m, isStudent))
    return HttpResponse.json({ total: list.length, page, items })
  }),

  http.get(`${API}/mentors/:id`, ({ params, request }) => {
    const mentor = MENTORS.find((m) => m.id === params.id)
    if (!mentor) return error(404, 'NOT_FOUND', 'Наставник не найден')
    const isStudent = currentUser(request)?.role === 'student'
    return HttpResponse.json(isStudent ? mentor : { ...mentor, match: undefined })
  }),

  http.get(`${API}/mentors/:id/reviews`, ({ params }) => {
    const items = REVIEWS[String(params.id)] ?? []
    return HttpResponse.json({ total: items.length, page: 1, items })
  }),
]
