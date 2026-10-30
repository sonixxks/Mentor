import type { ReactNode } from 'react'
import { Check, Clock, Eye, X } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { clsx } from 'clsx'
import styles from './Badge.module.css'

export type BadgeVariant = 'pending' | 'approved' | 'rejected' | 'review' | 'match' | 'neutral'

const icons: Partial<Record<BadgeVariant, LucideIcon>> = {
  pending: Clock,
  review: Eye,
  approved: Check,
  rejected: X,
}

type BadgeProps = {
  variant?: BadgeVariant
  children: ReactNode
}

export function Badge({ variant = 'neutral', children }: BadgeProps) {
  const Icon = icons[variant]

  return (
    <span className={clsx(styles.badge, styles[variant])}>
      {Icon ? <Icon size={13} strokeWidth={2.25} aria-hidden="true" /> : null}
      {children}
    </span>
  )
}