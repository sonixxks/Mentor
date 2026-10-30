import type { ButtonHTMLAttributes } from 'react'
import { Link } from 'react-router-dom'
import { clsx } from 'clsx'
import styles from './Button.module.css'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'accent' | 'secondary' | 'danger'
  size?: 'lg' | 'md' | 'sm'
  fullWidth?: boolean
  to?: string
}

export function Button({
  variant = 'primary',
  size = 'md',
  fullWidth = false,
  to,
  className,
  type = 'button',
  children,
  ...rest
}: ButtonProps) {
  const classes = clsx(
    styles.button,
    styles[variant],
    size !== 'lg' ? styles[size] : '',
    fullWidth ? styles.full : '',
    className,
  )

  if (to) {
    return (
      <Link to={to} className={classes}>
        {children}
      </Link>
    )
  }

  return (
    <button type={type} className={classes} {...rest}>
      {children}
    </button>
  )
}
