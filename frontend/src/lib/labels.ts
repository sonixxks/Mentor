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

/** 0 или нет цены → «Бесплатно», иначе «1 500 ₽/занятие» */
export function formatPrice(price?: number, suffix = '/занятие') {
  if (!price) return 'Бесплатно'
  return `${price.toLocaleString('ru-RU')} ₽${suffix}`
}

/** 4.9 → «4,9». Пока нет отзывов → «нет оценок» */
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
