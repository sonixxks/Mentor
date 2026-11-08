import { emptySchedule } from '../../components/ui/HourGrid'
import type { Schedule } from '../../components/ui/HourGrid'
import { detectTimezone } from '../../lib/timezones'

export type SkillLevel = 'beginner' | 'confident' | 'pro'

export type StudentSkill = { name: string; level: SkillLevel }

export type StudentAnswers = {
  // Шаг 1 · О себе
  lastName: string
  firstName: string
  middleName: string
  age: string
  status: string
  university: string
  direction: string
  course: string
  position: string
  company: string
  experience: string
  // Шаг 2 · Навыки и интересы
  targetSkills: string[]
  currentSkills: StudentSkill[]
  // Шаг 3 · Цели
  goals: string[]
  durationMonths: number
  success: string
  // Шаг 4 · Время и стиль
  hoursPerWeek: number
  timezone: string
  schedule: Schedule
  style: string
}

export const initialStudent: StudentAnswers = {
  lastName: '',
  firstName: '',
  middleName: '',
  age: '',
  status: '',
  university: '',
  direction: '',
  course: '',
  position: '',
  company: '',
  experience: '',
  targetSkills: [],
  currentSkills: [],
  goals: [],
  durationMonths: 3,
  success: '',
  hoursPerWeek: 2,
  timezone: detectTimezone(),
  schedule: emptySchedule(),
  style: '',
}

export type MentorSkill = { name: string; level: 'confident' | 'expert' }

export type MentorAnswers = {
  // Шаг 1 · Опыт
  lastName: string
  firstName: string
  middleName: string
  position: string
  company: string
  experience: string
  about: string
  links: string
  // Шаг 2 · Экспертиза
  skills: MentorSkill[]
  topics: string[]
  // Шаг 3 · Кого веду (длительность в месяцах)
  studentLevels: string[]
  goals: string[]
  durationFrom: number
  durationTo: number
  maxStudents: number
  paid: boolean
  price: string
  // Шаг 4 · Время и стиль
  hoursPerWeek: number
  timezone: string
  schedule: Schedule
  style: string
  visible: boolean
}

export const initialMentor: MentorAnswers = {
  lastName: '',
  firstName: '',
  middleName: '',
  position: '',
  company: '',
  experience: '',
  about: '',
  links: '',
  skills: [],
  topics: [],
  studentLevels: [],
  goals: [],
  durationFrom: 1,
  durationTo: 6,
  maxStudents: 3,
  paid: false,
  price: '',
  hoursPerWeek: 4,
  timezone: detectTimezone(),
  schedule: emptySchedule(),
  style: '',
  visible: true,
}
