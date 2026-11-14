import { clsx } from 'clsx'
import styles from './Segmented.module.css'

export type SegmentedOption = { value: string; label: string }

type SegmentedProps = {
  options: SegmentedOption[]
  value: string
  onChange: (value: string) => void
}

export function Segmented({ options, value, onChange }: SegmentedProps) {
  return (
    <div role="radiogroup" className={styles.segmented}>
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
  )
}
