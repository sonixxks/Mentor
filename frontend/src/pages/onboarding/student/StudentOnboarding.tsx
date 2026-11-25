import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { updateMyProfile } from '../../../api/profiles'
import type { UpdateProfileRequest } from '../../../api/profiles'
import { homeByRole } from '../../../lib/auth'
import { getSubmitErrorMessage } from '../errorMessage'
import { OnboardingShell } from '../OnboardingShell'
import { STEPS } from '../steps'
import { MAX_MONTHS, initialStudent } from '../types'
import type { StudentAnswers } from '../types'
import { useDraft } from '../useDraft'
import { AboutStep } from './AboutStep'
import { GoalsStep } from './GoalsStep'
import { SkillsStep } from './SkillsStep'
import { TimeStep } from './TimeStep'
import styles from '../Onboarding.module.css'
import { useSession } from '../../../hooks/useSession'

function validate(step: number, a: StudentAnswers) {
  if (step === 1) {
    const age = Number(a.age)
    if (!age || age < 14 || age > 99) return 'Укажи возраст от 14 до 99'

    if (!a.status) return 'Выбери, кто ты сейчас'

    if (
      (a.status === 'student' || a.status === 'graduate') &&
      (!a.university.trim() || !a.direction.trim())
    )
      return 'Укажи вуз и направление'

    if (a.status === 'student' && !a.course) return 'Выбери курс'

    if ((a.status === 'junior' || a.status === 'switcher') && !a.position.trim())
      return 'Укажи должность'

    if (!a.experience) return 'Выбери опыт работы'
  }

  if (step === 2) {
    if (a.targetSkills.length === 0) return 'Выбери хотя бы один навык, который хочешь развить'
  }

  if (step === 3) {
    if (a.goals.length === 0) return 'Выбери хотя бы одну цель'
    if (a.durationFrom < 1 || a.durationTo > MAX_MONTHS || a.durationFrom > a.durationTo)
      return `Проверь срок: от 1 до ${MAX_MONTHS} месяцев, «от» не больше «до»`
  }

  if (step === 4) {
    const hasTime = Object.values(a.schedule).some((hours) => hours.length > 0)
    if (!hasTime) return 'Отметь хотя бы один удобный час'
    if (!a.style) return 'Выбери стиль общения'
  }

  return ''
}

function toProfile(a: StudentAnswers): UpdateProfileRequest {
  const studies = a.status === 'student' || a.status === 'graduate'
  const works = a.status === 'junior' || a.status === 'switcher'

  return {
    age: Number(a.age),
    current_status: a.status as UpdateProfileRequest['current_status'],
    university: studies ? a.university.trim() : undefined,
    major: studies ? a.direction.trim() : undefined,
    course: a.status === 'student' ? Number(a.course) : undefined,
    job_title: works ? a.position.trim() : undefined,
    company: works ? a.company.trim() || null : undefined,
    experience: a.experience,
    target_skills: a.targetSkills,
    current_skills: a.currentSkills.map((skill) => ({ skill: skill.name, level: skill.level })),
    main_goals: a.goals,
    min_duration_months: a.durationFrom,
    max_duration_months: a.durationTo,
    success_metric: a.success.trim() || undefined,
    time_limit: a.hoursPerWeek,
    availability_slots: a.schedule,
    communication_style: a.style as UpdateProfileRequest['communication_style'],
    onboarding_completed: true,
  }
}

export function StudentOnboarding({ step }: { step: number }) {
  const navigate = useNavigate()
  const { logout } = useSession()
  const draft = useDraft('onboarding-student', initialStudent)
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const steps = STEPS.student
  const isLast = step === steps.length

  function back() {
    setError('')
    navigate(`/onboarding/student/${step - 1}`)
  }

  async function next() {
    if (isSubmitting) return

    const message = validate(step, draft.value)
    setError(message)
    if (message) return

    if (!isLast) {
      navigate(`/onboarding/student/${step + 1}`)
      return
    }

    setIsSubmitting(true)
    try {
      await updateMyProfile(toProfile(draft.value))
      draft.clear()
      navigate(homeByRole.student)
    } catch (submitError) {
      setError(getSubmitErrorMessage(submitError))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <OnboardingShell
      role="student"
      steps={steps}
      current={step}
      onBack={step > 1 ? back : undefined}
      onNext={next}
      nextLabel={isSubmitting ? 'Отправляем...' : isLast ? 'Подобрать наставников' : 'Далее'}
      onExit={logout}
    >
      {step === 1 ? <AboutStep value={draft.value} update={draft.update} /> : null}
      {step === 2 ? <SkillsStep value={draft.value} update={draft.update} /> : null}
      {step === 3 ? <GoalsStep value={draft.value} update={draft.update} /> : null}
      {step === 4 ? <TimeStep value={draft.value} update={draft.update} /> : null}

      {error ? <p className={styles.error}>{error}</p> : null}
    </OnboardingShell>
  )
}
