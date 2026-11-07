import type { TextareaHTMLAttributes } from 'react'
import { clsx } from 'clsx'
import styles from './Textarea.module.css'

type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  label?: string
  error?: string
}

export function Textarea({ label, error, maxLength, value, className, ...rest }: TextareaProps) {
  const length = typeof value === 'string' ? value.length : 0

  return (
    <div className={styles.field}>
      {label || maxLength ? (
        <div className={styles.labelRow}>
          {label ? <span className={styles.label}>{label}</span> : <span />}
          {maxLength ? (
            <span className={styles.counter}>
              {length} / {maxLength}
            </span>
          ) : null}
        </div>
      ) : null}
      <textarea
        aria-label={label}
        maxLength={maxLength}
        value={value}
        className={clsx(styles.textarea, error ? styles.invalid : '', className)}
        {...rest}
      />
      {error ? <span className={styles.error}>{error}</span> : null}
    </div>
  )
}
