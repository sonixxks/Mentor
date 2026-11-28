import type { RequestStatus } from '../api/requests'
import type { TaskStatus } from '../api/workspaces'
import { plural } from './plural'

export const EXPERIENCE_LABEL: Record<string, string> = {
  '1_to_3': '1–3 года',
  '3_to_5': '3–5 лет',
  '5_to_10': '5–10 лет',
  over_10: '10+ лет',
}

export const STYLE_LABEL: Record<string, string> = {
  friendly: 'Дружелюбный',
  structured: 'Структурный',
  free: 'Свободный',
}

export const HELP_TOPIC_LABEL: Record<string, string> = {
  interview_prep: 'Подготовка к собеседованию',
  code_review: 'Разбор кода',
  portfolio: 'Портфолио и проекты',
  career_growth: 'Развитие карьеры',
  change_direction: 'Смена направления',
  soft_skills: 'Развитие soft skills',
  first_job_in_it: 'Первая работа в IT',
}

export const STUDENT_LEVEL_LABEL: Record<string, string> = {
  beginner: 'Начинающие',
  has_basics: 'Есть база',
  advanced: 'Продвинутые',
}

type BadgeTone = 'pending' | 'approved' | 'rejected' | 'review' | 'neutral'

export const REQUEST_STATUS: Record<RequestStatus, { label: string; variant: BadgeTone }> = {
  pending: { label: 'Ждёт ответа', variant: 'pending' },
  approved_by_mentor: { label: 'Наставник согласен', variant: 'approved' },
  accepted: { label: 'Пара создана', variant: 'approved' },
  rejected_by_mentor: { label: 'Отклонена наставником', variant: 'rejected' },
  rejected_by_curator: { label: 'Отклонена куратором', variant: 'rejected' },
  cancelled: { label: 'Отменена', variant: 'neutral' },
}

export const TASK_STATUS: Record<TaskStatus, { label: string; variant: BadgeTone }> = {
  todo: { label: 'Запланировано', variant: 'neutral' },
  in_progress: { label: 'В работе', variant: 'neutral' },
  on_review: { label: 'На проверке', variant: 'review' },
  needs_revision: { label: 'На доработке', variant: 'pending' },
  completed: { label: 'Выполнено', variant: 'approved' },
}

export function formatPrice(price?: number, suffix = '/занятие') {
  if (!price) return 'Бесплатно'
  return `${price.toLocaleString('ru-RU')} ₽${suffix}`
}

export function formatRating(rating: number | null) {
  return rating === null ? 'нет оценок' : rating.toFixed(1).replace('.', ',')
}

export function reviewsLabel(count: number) {
  return plural(count, ['отзыв', 'отзыва', 'отзывов'])
}

export function freeSlotsLabel(count: number) {
  if (count === 0) return 'Мест нет'
  return `Есть ${plural(count, ['свободное место', 'свободных места', 'свободных мест'])}`
}
