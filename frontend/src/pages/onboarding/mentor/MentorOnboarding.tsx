import { useNavigate } from 'react-router-dom'
import { homeByRole, saveSession } from '../../../lib/auth'
import { OnboardingShell } from '../OnboardingShell'
import { STEPS } from '../steps'
import { initialMentor } from '../types'
import { useDraft } from '../useDraft'

export function MentorOnboarding({ step }: { step: number }) {
  const navigate = useNavigate()
  const draft = useDraft('onboarding-mentor', initialMentor)
  const steps = STEPS.mentor
  const isLast = step === steps.length

  function next() {
    if (!isLast) {
      navigate(`/onboarding/mentor/${step + 1}`)
      return
    }
    // TODO: POST /api/onboarding/mentor — отправить draft.value и документы
    draft.clear()
    saveSession('dev-token', 'mentor')
    navigate(homeByRole.mentor)
  }

  return (
    <OnboardingShell
      role="mentor"
      steps={steps}
      current={step}
      onBack={step > 1 ? () => navigate(`/onboarding/mentor/${step - 1}`) : undefined}
      onNext={next}
      nextLabel={isLast ? 'Отправить на проверку' : 'Далее'}
      onExit={() => navigate('/login')}
    >
      <p className="t-body text-secondary">Поля шага {step} появятся в следующих коммитах.</p>
    </OnboardingShell>
  )
}
