import type { SelectHTMLAttributes } from 'react'
import { ChevronDown } from 'lucide-react'
import { clsx } from 'clsx'
import styles from './Select.module.css'

export type SelectOption = { value: string; label: string }

type SelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  label?: string
  error?: string
  options: SelectOption[]
  placeholder?: string
}

export function Select({ label, error, options, placeholder, value, className, ...rest }: SelectProps) {
  return (
    <div className={styles.field}>
      {label ? <span className={styles.label}>{label}</span> : null}
      <div className={styles.control}>
        <select
          aria-label={label}
          value={value}
          className={clsx(
            styles.select,
            !value ? styles.placeholder : '',
            error ? styles.invalid : '',
            className,
          )}
          {...rest}
        >
          {placeholder ? (
            <option value="" disabled>
              {placeholder}
            </option>
          ) : null}
          {options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        <ChevronDown size={18} className={styles.arrow} aria-hidden="true" />
      </div>
      {error ? <span className={styles.error}>{error}</span> : null}
    </div>
  )
}