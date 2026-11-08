import { useNavigate } from 'react-router-dom'
import { homeByRole, saveSession } from '../../../lib/auth'
import { OnboardingShell } from '../OnboardingShell'
import { STEPS } from '../steps'
import { initialStudent } from '../types'
import { useDraft } from '../useDraft'

export function StudentOnboarding({ step }: { step: number }) {
  const navigate = useNavigate()
  const draft = useDraft('onboarding-student', initialStudent)
  const steps = STEPS.student
  const isLast = step === steps.length

  function next() {
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
      onBack={step > 1 ? () => navigate(`/onboarding/student/${step - 1}`) : undefined}
      onNext={next}
      nextLabel={isLast ? 'Подобрать наставников' : 'Далее'}
      onExit={() => navigate('/login')}
    >
      <p className="t-body text-secondary">Поля шага {step} появятся в следующих коммитах.</p>
    </OnboardingShell>
  )
}
