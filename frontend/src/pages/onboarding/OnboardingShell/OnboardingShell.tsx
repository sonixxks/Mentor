import { useEffect } from 'react'
import type { ReactNode } from 'react'
import { Logo } from '../../../components/Logo'
import { Button } from '../../../components/ui/Button'
import { Steps } from '../../../components/ui/Steps'
import type { PublicRole } from '../../../types'
import type { StepInfo } from '../steps'
import styles from './OnboardingShell.module.css'

const CHIP: Record<PublicRole, string> = {
  student: 'Анкета ученика',
  mentor: 'Анкета наставника',
}

type OnboardingShellProps = {
  role: PublicRole
  steps: StepInfo[]
  current: number
  onBack?: () => void
  onNext: () => void
  nextLabel: string
  onExit: () => void
  children: ReactNode
}

export function OnboardingShell({
  role,
  steps,
  current,
  onBack,
  onNext,
  nextLabel,
  onExit,
  children,
}: OnboardingShellProps) {
  const info = steps[current - 1]

  useEffect(() => {
    window.scrollTo(0, 0)
  }, [current])

  return (
    <div className={styles.page}>
      <div className={styles.glow} aria-hidden="true" />

      <header className={styles.header}>
        <div className={styles.brand}>
          <Logo />
          <span className={styles.chip}>{CHIP[role]}</span>
        </div>
        <button type="button" className={styles.exit} onClick={onExit}>
          Сохранить и выйти
        </button>
      </header>

      <div className={styles.steps}>
        <Steps steps={steps.map((step) => step.name)} current={current} />
      </div>

      <main className={styles.main}>
        <div className={styles.intro}>
          <h1 className="t-h1">{info.title}</h1>
          <p className={styles.text}>{info.text}</p>
        </div>
        <div className={styles.fields}>{children}</div>
      </main>

      <footer className={styles.footer}>
        {onBack ? (
          <Button variant="secondary" size="lg" onClick={onBack}>
            Назад
          </Button>
        ) : (
          <span />
        )}
        <Button size="lg" onClick={onNext}>
          {nextLabel}
        </Button>
      </footer>
    </div>
  )
}
