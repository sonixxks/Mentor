import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { homeByRole, saveSession } from '../../../lib/auth'
import { OnboardingShell } from '../OnboardingShell'
import { STEPS } from '../steps'
import { initialStudent } from '../types'
import type { StudentAnswers } from '../types'
import { useDraft } from '../useDraft'
import { AboutStep } from './AboutStep'
import styles from '../Onboarding.module.css'
import { SkillsStep } from './SkillsStep'
import { GoalsStep } from './GoalsStep'
import { TimeStep } from './TimeStep'

function validate(step: number, a: StudentAnswers) {
  if (step === 1) {
    if (!a.lastName.trim() || !a.firstName.trim()) return 'Укажи фамилию и имя'

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
  }

  if (step === 4) {
    const hasTime = Object.values(a.schedule).some((hours) => hours.length > 0)
    if (!hasTime) return 'Отметь хотя бы один удобный час'
    if (!a.style) return 'Выбери стиль общения'
  }

  return ''
}

export function StudentOnboarding({ step }: { step: number }) {
  const navigate = useNavigate()
  const draft = useDraft('onboarding-student', initialStudent)
  const [error, setError] = useState('')
  const steps = STEPS.student
  const isLast = step === steps.length

  function back() {
    setError('')
    navigate(`/onboarding/student/${step - 1}`)
  }

  function next() {
    const message = validate(step, draft.value)
    setError(message)
    if (message) return

    if (!isLast) {
      navigate(`/onboarding/student/${step + 1}`)
      return
    }
    // TODO: POST /api/onboarding/student — отправить draft.value
    draft.clear()
    saveSession('dev-token', 'student')
    navigate(homeByRole.student)
  }

  return (
    <OnboardingShell
      role="student"
      steps={steps}
      current={step}
      onBack={step > 1 ? back : undefined}
      onNext={next}
      nextLabel={isLast ? 'Подобрать наставников' : 'Далее'}
      onExit={() => navigate('/login')}
    >
      {step === 1 ? <AboutStep value={draft.value} update={draft.update} /> : null}
      {step === 2 ? <SkillsStep value={draft.value} update={draft.update} /> : null}
      {step === 3 ? <GoalsStep value={draft.value} update={draft.update} /> : null}
      {step === 4 ? <TimeStep value={draft.value} update={draft.update} /> : null}

      {error ? <p className={styles.error}>{error}</p> : null}
    </OnboardingShell>
  )
}
