import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { homeByRole } from '../../../lib/auth'
import { getSubmitErrorMessage } from '../errorMessage'
import { addVerificationAssets, updateMyProfile } from '../../../api/profiles'
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
import { useSession } from '../../../hooks/useSession'

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

    if (a.durationFrom < 1 || a.durationTo > MAX_MONTHS || a.durationFrom > a.durationTo) {
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
  }

  return ''
}

export function MentorOnboarding({ step }: { step: number }) {
  const navigate = useNavigate()
  const { logout } = useSession()
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
      const links = a.links.map((link) => link.trim()).filter(Boolean)

      await updateMyProfile({
        first_name: a.firstName.trim(),
        last_name: a.lastName.trim(),
        middle_name: a.middleName.trim() || undefined,
        job_title: a.position.trim(),
        company: a.selfEmployed ? null : a.company.trim() || null,
        experience: a.experience,
        bio: a.about.trim(),
        current_skills: a.skills.map((skill) => ({
          skill: skill.name,
          level: skill.level === 'expert' ? 3 : 2,
        })),
        help_topics: a.goals,
        target_student_levels: a.studentLevels,
        min_duration_months: a.durationFrom,
        max_duration_months: a.durationTo,
        hourly_rate: a.paid ? Number(a.price) : 0,
        max_students: a.maxStudents,
        time_limit: a.hoursPerWeek,
        availability_slots: a.schedule,
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
      setError(getSubmitErrorMessage(submitError))
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
      nextLabel={isSubmitting ? 'Отправляем...' : isLast ? 'Отправить на проверку' : 'Далее'}
      onExit={logout}
    >
      {step === 1 ? (
        <AboutStep
          value={draft.value}
          update={draft.update}
          documents={documents}
          onDocumentsChange={setDocuments}
        />
      ) : null}

      {step === 2 ? <SkillsStep value={draft.value} update={draft.update} /> : null}

      {step === 3 ? <FormatStep value={draft.value} update={draft.update} /> : null}

      {step === 4 ? <TimeStep value={draft.value} update={draft.update} /> : null}

      {error ? <p className={styles.error}>{error}</p> : null}
    </OnboardingShell>
  )
}
