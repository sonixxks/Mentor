import { useState } from 'react'
import type { InputHTMLAttributes, ReactNode } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import { clsx } from 'clsx'
import styles from './Input.module.css'

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string
  error?: string
  labelAction?: ReactNode
}

export function Input({ label, error, labelAction, type, className, ...rest }: InputProps) {
  const [visible, setVisible] = useState(false)
  const isPassword = type === 'password'

  return (
    <div className={styles.field}>
      {label || labelAction ? (
        <div className={styles.labelRow}>
          {label ? <span className={styles.label}>{label}</span> : <span />}
          {labelAction}
        </div>
      ) : null}
      <div className={styles.control}>
        <input
          type={isPassword && visible ? 'text' : type}
          aria-label={label}
          className={clsx(
            styles.input,
            isPassword ? styles.withToggle : '',
            error ? styles.invalid : '',
            className,
          )}
          {...rest}
        />
        {isPassword ? (
          <button
            type="button"
            className={styles.toggle}
            aria-label={visible ? 'Скрыть пароль' : 'Показать пароль'}
            onClick={() => setVisible(!visible)}
          >
            {visible ? <EyeOff size={20} strokeWidth={1.75} /> : <Eye size={20} strokeWidth={1.75} />}
          </button>
        ) : null}
      </div>
      {error ? <span className={styles.error}>{error}</span> : null}
    </div>
  )
}