import { clsx } from 'clsx'
import styles from './Segmented.module.css'

export type SegmentedOption = { value: string; label: string }

type SegmentedProps = {
  label?: string
  options: SegmentedOption[]
  value: string
  onChange: (value: string) => void
}

export function Segmented({ label, options, value, onChange }: SegmentedProps) {
  return (
    <div className={styles.field}>
      {label ? <span className={styles.label}>{label}</span> : null}
      <div className={styles.group} role="radiogroup" aria-label={label}>
        {options.map((option) => (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={option.value === value}
            className={clsx(styles.option, option.value === value ? styles.active : '')}
            onClick={() => onChange(option.value)}
          >
            {option.label}
          </button>
        ))}
      </div>
    </div>
  )
}
