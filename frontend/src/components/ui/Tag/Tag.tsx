import type { ReactNode } from 'react'
import { X } from 'lucide-react'
import { clsx } from 'clsx'
import styles from './Tag.module.css'

type TagProps = {
  children: ReactNode
  selected?: boolean
  onClick?: () => void
  onRemove?: () => void
}

export function Tag({ children, selected = false, onClick, onRemove }: TagProps) {
  const classes = clsx(
    styles.tag,
    selected ? styles.selected : '',
    onClick || onRemove ? styles.clickable : '',
  )

  if (onRemove) {
    return (
      <button type="button" className={classes} onClick={onRemove}>
        {children}
        <X size={14} className={styles.removeIcon} />
      </button>
    )
  }

  if (onClick) {
    return (
      <button type="button" className={classes} onClick={onClick}>
        {children}
      </button>
    )
  }

  return <span className={classes}>{children}</span>
}