
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { homeByRole } from '../../../lib/auth'
import { ApiError } from '../../../api/client'
import {
  addVerificationAssets,
  updateMyProfile,
} from '../../../api/profiles'
import type { AvailabilitySlots } from '../../../api/profiles'
import { OnboardingShell } from '../OnboardingShell'
import { STEPS } from '../steps'
import { MAX_MONTHS, initialMentor } from '../types'
import type { MentorAnswers } from '../types'
import { useDraft } from '../useDraft'
import { AboutStep } from './AboutStep'
import styles from '../Onboarding.module.css'
import { isUrl } from '../../../lib/validation'
import { SkillsStep } from './SkillsStep'
import { FormatStep } from './FormatStep'
import { TimeStep } from './TimeStep'

const EKATERINBURG_TIMEZONE = 'Asia/Yekaterinburg'

const DAYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const

function getTimezoneOffset(timezone: string): number {
  const date = new Date('2026-01-07T12:00:00Z')

  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(date)

  const values = Object.fromEntries(
    parts.map((part) => [part.type, part.value]),
  )

  const localTime = Date.UTC(
    Number(values.year),
    Number(values.month) - 1,
    Number(values.day),
    Number(values.hour),
    Number(values.minute),
    Number(values.second),
  )

  return (localTime - date.getTime()) / 60000
}

function convertSchedule(
  schedule: MentorAnswers['schedule'],
  timezone: string,
): AvailabilitySlots {
  const sourceOffset = getTimezoneOffset(timezone)
  const targetOffset = getTimezoneOffset(EKATERINBURG_TIMEZONE)
  const shift = (targetOffset - sourceOffset) / 60

  if (!Number.isInteger(shift)) {
    throw new Error('Не удалось преобразовать выбранный часовой пояс')
  }

  const result: AvailabilitySlots = {
    mon: [],
    tue: [],
    wed: [],
    thu: [],
    fri: [],
    sat: [],
    sun: [],
  }

  DAYS.forEach((day, dayIndex) => {
    schedule[day].forEach((hour) => {
      const totalHours = dayIndex * 24 + hour + shift
      const targetDay = ((Math.floor(totalHours / 24) % 7) + 7) % 7
      const targetHour = ((totalHours % 24) + 24) % 24
      const targetDayName = DAYS[targetDay]

      if (targetHour < 8 || targetHour > 21) {
        throw new Error(
          'После перевода в Екатеринбургское время некоторые выбранные часы выходят за допустимый диапазон 08:00–22:00. Измените расписание.',
        )
      }

      result[targetDayName].push(targetHour)
    })
  })

  DAYS.forEach((day) => {
    result[day] = [...new Set(result[day])].sort((a, b) => a - b)
  })

  return result
}

function mapStudentLevels(levels: string[]): string[] {
  return levels.map((level) => {
    if (level === 'base') return 'has_basics'
    return level
  })
}

function validate(step: number, a: MentorAnswers, documents: File[]) {
  if (step === 1) {
    if (!a.lastName.trim() || !a.firstName.trim()) return 'Укажите фамилию и имя'
    if (!a.position.trim()) return 'Укажите должность'
    if (!a.experience) return 'Выберите опыт работы'
    if (!a.about.trim()) return 'Расскажите коротко о себе'
    if (a.about.trim().length > 300) return 'Описание должно содержать не более 300 символов'

    const links = a.links.map((link) => link.trim()).filter(Boolean)

    if (links.some((link) => !isUrl(link))) {
      return 'Проверьте ссылки — похоже, в одной из них ошибка'
    }

    if (documents.length > 5) return 'Можно прикрепить не более 5 документов'
    if (links.length > 5) return 'Можно добавить не более 5 ссылок'

    if (documents.length === 0 && links.length === 0) {
      return 'Прикрепите документ или ссылку, подтверждающие опыт'
    }
  }

  if (step === 2) {
    if (a.skills.length === 0) return 'Добавьте хотя бы один навык'
    if (a.goals.length === 0) return 'Выберите, с чем помогаете'
  }

  if (step === 3) {
    if (a.studentLevels.length === 0) return 'Выберите уровень учеников'

    if (
      a.durationFrom < 1 ||
      a.durationTo > MAX_MONTHS ||
      a.durationFrom > a.durationTo
    ) {
      return `Проверьте длительность: от 1 до ${MAX_MONTHS} месяцев, «от» не больше «до»`
    }

    if (a.maxStudents < 1 || a.maxStudents > 10) {
      return 'Количество учеников должно быть от 1 до 10'
    }

    if (a.paid && !(Number(a.price) > 0)) {
      return 'Укажите цену за занятие'
    }
  }

  if (step === 4) {
    const hasTime = Object.values(a.schedule).some((hours) => hours.length > 0)

    if (!hasTime) return 'Отметьте хотя бы один удобный час'
    if (!a.style) return 'Выберите стиль наставничества'

    try {
      convertSchedule(a.schedule, a.timezone)
    } catch (error) {
      return error instanceof Error ? error.message : 'Проверьте расписание'
    }
  }

  return ''
}

function getApiErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 400) {
      return `Сервер не принял данные анкеты: ${error.message}`
    }

    if (error.status === 401) {
      return 'Сессия истекла. Войдите в аккаунт заново.'
    }

    if (error.status === 0) {
      return error.message
    }
  }

  return error instanceof Error
    ? error.message
    : 'Не удалось отправить анкету. Попробуйте ещё раз.'
}

export function MentorOnboarding({ step }: { step: number }) {
  const navigate = useNavigate()
  const draft = useDraft('onboarding-mentor', initialMentor)
  const [documents, setDocuments] = useState<File[]>([])
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const steps = STEPS.mentor
  const isLast = step === steps.length

  function back() {
    setError('')
    navigate(`/onboarding/mentor/${step - 1}`)
  }

  async function submit() {
    const a = draft.value

    try {
      const availabilitySlots = convertSchedule(a.schedule, a.timezone)

      const links = a.links
        .map((link) => link.trim())
        .filter(Boolean)

      await updateMyProfile({
        first_name: a.firstName.trim(),
        last_name: a.lastName.trim(),
        middle_name: a.middleName.trim() || undefined,
        job_title: a.position.trim(),
        company: a.company.trim() || null,
        experience: a.experience,
        bio: a.about.trim(),
        current_skills: a.skills.map((skill) => ({
          skill: skill.name,
          level: skill.level === 'expert' ? 3 : 2,
        })),
        help_topics: a.goals,
        target_student_levels: mapStudentLevels(a.studentLevels),
        min_duration_months: a.durationFrom,
        max_duration_months: a.durationTo,
        hourly_rate: a.paid ? Number(a.price) : 0,
        max_students: a.maxStudents,
        time_limit: a.hoursPerWeek,
        availability_slots: availabilitySlots,
        communication_style: a.style as 'friendly' | 'structured' | 'free',
        is_published: false,
        onboarding_completed: true,
      })

      await addVerificationAssets(documents, links)

      await updateMyProfile({
        is_published: a.visible,
      })

      draft.clear()
      navigate(homeByRole.mentor)
    } catch (submitError) {
      setError(getApiErrorMessage(submitError))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function next() {
    if (isSubmitting) return

    const message = validate(step, draft.value, documents)
    setError(message)

    if (message) return

    if (!isLast) {
      navigate(`/onboarding/mentor/${step + 1}`)
      return
    }

    setIsSubmitting(true)
    setError('')
    await submit()
  }

  return (
    <OnboardingShell
      role="mentor"
      steps={steps}
      current={step}
      onBack={step > 1 ? back : undefined}
      onNext={next}
      nextLabel={
        isSubmitting
          ? 'Отправляем...'
          : isLast
            ? 'Отправить на проверку'
            : 'Далее'
      }
      onExit={() => navigate('/login')}
    >
      {step === 1 ? (
        <AboutStep
          value={draft.value}
          update={draft.update}
          documents={documents}
          onDocumentsChange={setDocuments}
        />
      ) : null}

      {step === 2 ? (
        <SkillsStep value={draft.value} update={draft.update} />
      ) : null}

      {step === 3 ? (
        <FormatStep value={draft.value} update={draft.update} />
      ) : null}

      {step === 4 ? (
        <TimeStep value={draft.value} update={draft.update} />
      ) : null}

      {error ? <p className={styles.error}>{error}</p> : null}
    </OnboardingShell>
  )
}
