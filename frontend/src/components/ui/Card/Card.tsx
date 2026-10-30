import type { HTMLAttributes } from 'react'
import { clsx } from 'clsx'
import styles from './Card.module.css'

type CardProps = HTMLAttributes<HTMLDivElement> & {
  variant?: 'default' | 'accent' | 'glass'
}

export function Card({ variant = 'default', className, children, ...rest }: CardProps) {
  return (
    <div
      className={clsx(styles.card, variant !== 'default' ? styles[variant] : '', className)}
      {...rest}
    >
      {children}
    </div>
  )
}