import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { homeByRole, saveSession } from '../../../lib/auth'
import { OnboardingShell } from '../OnboardingShell'
import { STEPS } from '../steps'
import { initialMentor } from '../types'
import type { MentorAnswers } from '../types'
import { useDraft } from '../useDraft'
import { AboutStep } from './AboutStep'
import styles from '../Onboarding.module.css'
import { isUrl } from '../../../lib/validation'

function validate(step: number, a: MentorAnswers, documents: File[]) {
  if (step === 1) {
    if (!a.lastName.trim() || !a.firstName.trim()) return 'Укажите фамилию и имя'
    if (!a.position.trim()) return 'Укажите должность'
    if (!a.experience) return 'Выберите опыт работы'
    if (!a.about.trim()) return 'Расскажите коротко о себе'
    const links = a.links.map((link) => link.trim()).filter(Boolean)
    if (links.some((link) => !isUrl(link)))
      return 'Проверьте ссылки — похоже, в одной из них ошибка'
    if (documents.length === 0 && links.length === 0)
      return 'Прикрепите документ или ссылку, подтверждающие опыт'
  }
  return ''
}

export function MentorOnboarding({ step }: { step: number }) {
  const navigate = useNavigate()
  const draft = useDraft('onboarding-mentor', initialMentor)
  const [documents, setDocuments] = useState<File[]>([])
  const [error, setError] = useState('')
  const steps = STEPS.mentor
  const isLast = step === steps.length

  function back() {
    setError('')
    navigate(`/onboarding/mentor/${step - 1}`)
  }

  function next() {
    const message = validate(step, draft.value, documents)
    setError(message)
    if (message) return

    if (!isLast) {
      navigate(`/onboarding/mentor/${step + 1}`)
      return
    }
    // TODO: POST /api/onboarding/mentor — отправить draft.value и documents
    draft.clear()
    saveSession('dev-token', 'mentor')
    navigate(homeByRole.mentor)
  }

  return (
    <OnboardingShell
      role="mentor"
      steps={steps}
      current={step}
      onBack={step > 1 ? back : undefined}
      onNext={next}
      nextLabel={isLast ? 'Отправить на проверку' : 'Далее'}
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

      {step > 1 ? (
        <p className="t-body text-secondary">Поля шага {step} появятся в следующих коммитах.</p>
      ) : null}

      {error ? <p className={styles.error}>{error}</p> : null}
    </OnboardingShell>
  )
}
