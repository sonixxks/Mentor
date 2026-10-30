import type { InputHTMLAttributes } from 'react'
import { clsx } from 'clsx'
import styles from './Input.module.css'

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string
  error?: string
}

export function Input({ label, error, className, ...rest }: InputProps) {
  return (
    <label className={styles.field}>
      {label ? <span className={styles.label}>{label}</span> : null}
      <input
        className={clsx(styles.input, error ? styles.invalid : '', className)}
        {...rest}
      />
      {error ? <span className={styles.error}>{error}</span> : null}
    </label>
  )
}
