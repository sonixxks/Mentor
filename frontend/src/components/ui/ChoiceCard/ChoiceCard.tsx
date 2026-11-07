import { Check } from 'lucide-react'
import { clsx } from 'clsx'
import styles from './ChoiceCard.module.css'

type ChoiceCardProps = {
  title: string
  description?: string
  selected: boolean
  onClick: () => void
}

export function ChoiceCard({ title, description, selected, onClick }: ChoiceCardProps) {
  return (
    <button
      type="button"
      aria-pressed={selected}
      className={clsx(styles.card, selected ? styles.selected : '')}
      onClick={onClick}
    >
      <span className={styles.check}>{selected ? <Check size={12} strokeWidth={3} /> : null}</span>
      <span className={styles.title}>{title}</span>
      {description ? <span className={styles.description}>{description}</span> : null}
    </button>
  )
}