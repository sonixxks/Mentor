import { clsx } from 'clsx'
import styles from './Avatar.module.css'

type AvatarProps = {
  name: string
  size?: 's' | 'm' | 'l'
}

function initials(name: string) {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('')
}

export function Avatar({ name, size = 'm' }: AvatarProps) {
  return (
    <span className={clsx(styles.avatar, styles[size])}>
      {initials(name)}
    </span>
  )
}